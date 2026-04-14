package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joedaviesio/tirith/internal/config"
	"github.com/joedaviesio/tirith/internal/dashboard"
	"github.com/joedaviesio/tirith/internal/pricing"
	"github.com/joedaviesio/tirith/internal/proxy"
	"github.com/joedaviesio/tirith/internal/report"
	"github.com/joedaviesio/tirith/internal/storage"
	"github.com/spf13/cobra"
)

var version = "2.0.1"

func main() {
	var noOpen bool

	root := &cobra.Command{
		Use:   "tirith",
		Short: "AI API cost observability — one import, full transparency",
		Long: `Tirith is a transparent proxy for AI API cost observability.

Run with no subcommand to start the proxy, dashboard, and open your browser.`,
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(0, !noOpen)
		},
	}

	root.Flags().BoolVar(&noOpen, "no-open", false, "don't open the dashboard in your browser")

	root.AddCommand(startCmd())
	root.AddCommand(stopCmd())
	root.AddCommand(reportCmd())
	root.AddCommand(dashboardCmd())
	root.AddCommand(runCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func runStart(port int, openBrowser bool) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if port != 0 {
		cfg.Proxy.Port = port
	}

	if err := config.EnsureDir(); err != nil {
		return err
	}

	store, err := storage.New(cfg.Storage.Path)
	if err != nil {
		return fmt.Errorf("opening storage: %w", err)
	}
	defer store.Close()

	pricer, err := pricing.New()
	if err != nil {
		return fmt.Errorf("initializing pricing engine: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	srv := proxy.NewServer(cfg, store, pricer, logger)

	// Start dashboard alongside proxy.
	proxyURL := fmt.Sprintf("http://%s:%d", cfg.Proxy.Host, cfg.Proxy.Port)
	dash := dashboard.NewServer(cfg.Dashboard.Host, cfg.Dashboard.Port, proxyURL, store, logger)
	go func() {
		if err := dash.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "Dashboard failed to start: %s\n", err)
		}
	}()

	// Handle graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	shutdownDone := make(chan struct{})

	go func() {
		<-sigCh
		// Restore default signal behavior so a second Ctrl+C force-kills.
		signal.Stop(sigCh)
		fmt.Fprintln(os.Stderr, "\nStopping Tirith... bye!")
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); dash.Shutdown() }()
		go func() { defer wg.Done(); srv.Shutdown() }()
		wg.Wait()
		close(shutdownDone)
	}()

	dashURL := fmt.Sprintf("http://localhost:%d", cfg.Dashboard.Port)

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "  Kia ora from Tirith — %s\n", dashURL)
	fmt.Fprintln(os.Stderr, "  Start:     tirith start")
	fmt.Fprintln(os.Stderr, "  Stop:      Ctrl+C")
	fmt.Fprintln(os.Stderr, "  Reinstall: import tirith")
	fmt.Fprintln(os.Stderr)

	if openBrowser {
		go func() {
			// Wait for dashboard to be ready before opening browser.
			for i := 0; i < 20; i++ {
				resp, err := http.Get(dashURL)
				if err == nil {
					resp.Body.Close()
					openURL(dashURL)
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
			// Dashboard didn't come up in time — skip browser open.
			logger.Warn("dashboard did not respond in time, skipping browser open")
		}()
	}

	if err := srv.Start(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			<-shutdownDone
			return nil
		}
		return err
	}
	return nil
}

func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	cmd.Start()
}

func startCmd() *cobra.Command {
	var port int
	var noOpen bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Tirith proxy server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(port, !noOpen)
		},
	}

	cmd.Flags().IntVar(&port, "port", 0, "proxy port (default: from config or 5555)")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "don't open the dashboard in your browser")
	return cmd
}

func stopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the Tirith proxy server",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stderr, "To stop Tirith, press Ctrl+C in the terminal where it's running.")
			return nil
		},
	}
}

func dashboardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dashboard",
		Short: "Open the Tirith dashboard in your browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			url := fmt.Sprintf("http://localhost:%d", cfg.Dashboard.Port)
			fmt.Fprintln(os.Stderr, "Opening dashboard...")

			openURL(url)
			return nil
		},
	}
}

func reportCmd() *cobra.Command {
	var last string
	var tag string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Print cost summary to terminal",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if err := config.EnsureDir(); err != nil {
				return err
			}

			store, err := storage.New(cfg.Storage.Path)
			if err != nil {
				return fmt.Errorf("opening storage: %w", err)
			}
			defer store.Close()

			filter := storage.ReportFilter{
				Tag: tag,
			}

			period := "All Time"
			if last != "" {
				dur, err := parseDuration(last)
				if err != nil {
					return fmt.Errorf("invalid --last value: %w", err)
				}
				filter.Since = time.Now().Add(-dur)
				period = "Last " + last
			}

			summary, err := store.GetSummary(filter)
			if err != nil {
				return err
			}

			models, err := store.GetByModel(filter)
			if err != nil {
				return err
			}

			tags, err := store.GetByTag(filter)
			if err != nil {
				return err
			}

			w := os.Stdout
			report.PrintSummary(w, summary, period)
			report.PrintByModel(w, models)
			report.PrintByTag(w, tags)

			return nil
		},
	}

	cmd.Flags().StringVar(&last, "last", "", "time period (e.g., 1h, 7d, 30d)")
	cmd.Flags().StringVar(&tag, "tag", "", "filter by tag")
	return cmd
}

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run -- <command>",
		Short: "Run a command with Tirith proxy env vars injected",
		Long:  "Starts the proxy (if needed), sets ANTHROPIC_BASE_URL and OPENAI_BASE_URL, then runs your command.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if err := config.EnsureDir(); err != nil {
				return err
			}

			store, err := storage.New(cfg.Storage.Path)
			if err != nil {
				return fmt.Errorf("opening storage: %w", err)
			}

			pricer, err := pricing.New()
			if err != nil {
				return fmt.Errorf("initializing pricing engine: %w", err)
			}

			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

			srv := proxy.NewServer(cfg, store, pricer, logger)

			// Start proxy in background.
			go func() {
				if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					fmt.Fprintf(os.Stderr, "Proxy failed to start: %s\n", err)
				}
			}()

			// Give proxy a moment to bind.
			time.Sleep(100 * time.Millisecond)

			proxyBase := fmt.Sprintf("http://localhost:%d", cfg.Proxy.Port)

			// Build subprocess with injected env vars.
			child := exec.Command(args[0], args[1:]...)
			child.Stdout = os.Stdout
			child.Stderr = os.Stderr
			child.Stdin = os.Stdin
			child.Env = append(os.Environ(),
				fmt.Sprintf("ANTHROPIC_BASE_URL=%s/proxy/anthropic", proxyBase),
				fmt.Sprintf("OPENAI_BASE_URL=%s/proxy/openai", proxyBase),
			)

			fmt.Fprintf(os.Stderr, "Tirith tracking costs. Running: %s\n", strings.Join(args, " "))

			err = child.Run()

			// Shutdown proxy after child exits.
			srv.Shutdown()
			store.Close()

			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					os.Exit(exitErr.ExitCode())
				}
				return err
			}
			return nil
		},
	}
}

// parseDuration parses durations like "1h", "7d", "30d", "2w".
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return 0, fmt.Errorf("too short: %q", s)
	}

	unit := s[len(s)-1]
	val := s[:len(s)-1]

	var n int
	if _, err := fmt.Sscanf(val, "%d", &n); err != nil {
		return 0, fmt.Errorf("invalid number: %q", val)
	}

	switch unit {
	case 'h':
		return time.Duration(n) * time.Hour, nil
	case 'd':
		return time.Duration(n) * 24 * time.Hour, nil
	case 'w':
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	default:
		// Fall back to Go's standard duration parsing.
		return time.ParseDuration(s)
	}
}
