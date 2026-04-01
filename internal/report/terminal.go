package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/joedaviesio/tirith/internal/storage"
)

func FormatCents(cents int) string {
	if cents < 100 {
		return fmt.Sprintf("$0.%02d", cents)
	}
	return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
}

func PrintSummary(w io.Writer, summary *storage.Summary, period string) {
	fmt.Fprintf(w, "\n  Tirith Report — %s\n", period)
	fmt.Fprintf(w, "  %s\n\n", strings.Repeat("─", 40))
	fmt.Fprintf(w, "  Total Cost:       %s\n", FormatCents(summary.TotalCostCents))
	fmt.Fprintf(w, "  Total Calls:      %d\n", summary.TotalCalls)
	fmt.Fprintf(w, "  Input Tokens:     %s\n", formatNumber(summary.TotalInput))
	fmt.Fprintf(w, "  Output Tokens:    %s\n", formatNumber(summary.TotalOutput))
	fmt.Fprintf(w, "  Avg Latency:      %dms\n", summary.AvgLatencyMs)
	fmt.Fprintln(w)
}

func PrintByModel(w io.Writer, models []storage.ModelSummary) {
	if len(models) == 0 {
		fmt.Fprintln(w, "  No data by model.")
		return
	}

	fmt.Fprintln(w, "  By Model")
	fmt.Fprintf(w, "  %s\n", strings.Repeat("─", 70))
	fmt.Fprintf(w, "  %-28s %8s %8s %10s %10s\n", "Model", "Cost", "Calls", "Input", "Output")
	fmt.Fprintf(w, "  %s\n", strings.Repeat("─", 70))

	for _, m := range models {
		fmt.Fprintf(w, "  %-28s %8s %8d %10s %10s\n",
			m.Model, FormatCents(m.TotalCostCents), m.TotalCalls,
			formatNumber(m.TotalInput), formatNumber(m.TotalOutput),
		)
	}
	fmt.Fprintln(w)
}

func PrintByTag(w io.Writer, tags []storage.TagSummary) {
	if len(tags) == 0 {
		fmt.Fprintln(w, "  No data by tag.")
		return
	}

	fmt.Fprintln(w, "  By Tag")
	fmt.Fprintf(w, "  %s\n", strings.Repeat("─", 60))
	fmt.Fprintf(w, "  %-24s %8s %8s %10s %8s\n", "Tag", "Cost", "Calls", "Avg Cost", "Avg Lat")
	fmt.Fprintf(w, "  %s\n", strings.Repeat("─", 60))

	for _, t := range tags {
		fmt.Fprintf(w, "  %-24s %8s %8d %10s %6dms\n",
			t.Tag, FormatCents(t.TotalCostCents), t.TotalCalls,
			FormatCents(t.AvgCostCents), t.AvgLatencyMs,
		)
	}
	fmt.Fprintln(w)
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1_000_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
}
