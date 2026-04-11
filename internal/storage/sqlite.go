package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type APICall struct {
	ID               string
	Timestamp        time.Time
	Provider         string
	Model            string
	InputTokens      int
	OutputTokens     int
	CacheReadTokens  int
	CacheWriteTokens int
	CostCents        int
	LatencyMs        int
	StatusCode       int
	Streaming        bool
	Tag              string
	UserTag          string
	SessionTag       string
	Environment      string
	Endpoint         string
}

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("creating schema: %w", err)
	}

	// Enable WAL mode for better concurrent read/write performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting WAL mode: %w", err)
	}

	// Limit connection pool to prevent idle connections accumulating.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Insert(call *APICall) error {
	_, err := s.db.Exec(`
		INSERT INTO api_calls (
			id, timestamp, provider, model,
			input_tokens, output_tokens, cache_read_tokens, cache_write_tokens,
			cost_cents, latency_ms, status_code, streaming,
			tag, user_tag, session_tag, environment, endpoint
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		call.ID, call.Timestamp.UTC().Format(time.RFC3339), call.Provider, call.Model,
		call.InputTokens, call.OutputTokens, call.CacheReadTokens, call.CacheWriteTokens,
		call.CostCents, call.LatencyMs, call.StatusCode, call.Streaming,
		nullStr(call.Tag), nullStr(call.UserTag), nullStr(call.SessionTag),
		call.Environment, nullStr(call.Endpoint),
	)
	if err != nil {
		return fmt.Errorf("inserting api call: %w", err)
	}
	return nil
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// ReportFilter controls which rows are included in reports.
type ReportFilter struct {
	Since       time.Time
	Tag         string
	UserTag     string
	Environment string
	Provider    string
	Model       string
	TZOffset    string // SQLite offset for DATE(), e.g. "+12:00", "-05:00"
}

type Summary struct {
	TotalCostCents  int
	TotalCalls      int
	TotalInput      int
	TotalOutput     int
	AvgLatencyMs    int
}

type ModelSummary struct {
	Model          string
	Provider       string
	TotalCostCents int
	TotalCalls     int
	TotalInput     int
	TotalOutput    int
	AvgLatencyMs   int
}

type TagSummary struct {
	Tag            string
	TotalCostCents int
	TotalCalls     int
	AvgCostCents   int
	AvgLatencyMs   int
}

func (s *Store) GetSummary(f ReportFilter) (*Summary, error) {
	where, args := buildWhere(f)
	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(cost_cents), 0),
		       COUNT(*),
		       COALESCE(SUM(input_tokens), 0),
		       COALESCE(SUM(output_tokens), 0),
		       COALESCE(AVG(latency_ms), 0)
		FROM api_calls %s`, where)

	var sum Summary
	var avgLatency float64
	err := s.db.QueryRow(query, args...).Scan(
		&sum.TotalCostCents, &sum.TotalCalls,
		&sum.TotalInput, &sum.TotalOutput, &avgLatency,
	)
	if err != nil {
		return nil, fmt.Errorf("querying summary: %w", err)
	}
	sum.AvgLatencyMs = int(avgLatency)
	return &sum, nil
}

func (s *Store) GetByModel(f ReportFilter) ([]ModelSummary, error) {
	where, args := buildWhere(f)
	query := fmt.Sprintf(`
		SELECT model, provider,
		       SUM(cost_cents), COUNT(*),
		       SUM(input_tokens), SUM(output_tokens),
		       AVG(latency_ms)
		FROM api_calls %s
		GROUP BY model, provider
		ORDER BY SUM(cost_cents) DESC
		LIMIT 100`, where)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying by model: %w", err)
	}
	defer rows.Close()

	var results []ModelSummary
	for rows.Next() {
		var m ModelSummary
		var avgLatency float64
		if err := rows.Scan(&m.Model, &m.Provider, &m.TotalCostCents, &m.TotalCalls,
			&m.TotalInput, &m.TotalOutput, &avgLatency); err != nil {
			return nil, fmt.Errorf("scanning model row: %w", err)
		}
		m.AvgLatencyMs = int(avgLatency)
		results = append(results, m)
	}
	return results, rows.Err()
}

func (s *Store) GetByTag(f ReportFilter) ([]TagSummary, error) {
	where, args := buildWhere(f)
	query := fmt.Sprintf(`
		SELECT COALESCE(tag, '(untagged)'),
		       SUM(cost_cents), COUNT(*),
		       AVG(cost_cents), AVG(latency_ms)
		FROM api_calls %s
		GROUP BY tag
		ORDER BY SUM(cost_cents) DESC
		LIMIT 100`, where)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying by tag: %w", err)
	}
	defer rows.Close()

	var results []TagSummary
	for rows.Next() {
		var t TagSummary
		var avgCost, avgLatency float64
		if err := rows.Scan(&t.Tag, &t.TotalCostCents, &t.TotalCalls,
			&avgCost, &avgLatency); err != nil {
			return nil, fmt.Errorf("scanning tag row: %w", err)
		}
		t.AvgCostCents = int(avgCost)
		t.AvgLatencyMs = int(avgLatency)
		results = append(results, t)
	}
	return results, rows.Err()
}

type UserSummary struct {
	UserTag        string
	TotalCostCents int
	TotalCalls     int
	FirstSeen      string
	LastSeen       string
}

type DailySpend struct {
	Date           string
	TotalCostCents int
	TotalCalls     int
	TotalInput     int
	TotalOutput    int
}

type DailyModelSpend struct {
	Date           string
	Model          string
	TotalCostCents int
	TotalCalls     int
	TotalInput     int
	TotalOutput    int
}

func (s *Store) GetByUser(f ReportFilter) ([]UserSummary, error) {
	where, args := buildWhere(f)
	de := dateExpr(f)
	query := fmt.Sprintf(`
		SELECT COALESCE(user_tag, '(anonymous)'),
		       SUM(cost_cents), COUNT(*),
		       MIN(%s), MAX(%s)
		FROM api_calls %s
		GROUP BY user_tag
		ORDER BY SUM(cost_cents) DESC
		LIMIT 100`, de, de, where)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying by user: %w", err)
	}
	defer rows.Close()

	var results []UserSummary
	for rows.Next() {
		var u UserSummary
		if err := rows.Scan(&u.UserTag, &u.TotalCostCents, &u.TotalCalls,
			&u.FirstSeen, &u.LastSeen); err != nil {
			return nil, fmt.Errorf("scanning user row: %w", err)
		}
		results = append(results, u)
	}
	return results, rows.Err()
}

func (s *Store) GetDailySpend(f ReportFilter) ([]DailySpend, error) {
	where, args := buildWhere(f)
	de := dateExpr(f)
	query := fmt.Sprintf(`
		SELECT %s as day,
		       SUM(cost_cents), COUNT(*),
		       SUM(input_tokens), SUM(output_tokens)
		FROM api_calls %s
		GROUP BY day
		ORDER BY day ASC
		LIMIT 90`, de, where)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying daily spend: %w", err)
	}
	defer rows.Close()

	var results []DailySpend
	for rows.Next() {
		var d DailySpend
		if err := rows.Scan(&d.Date, &d.TotalCostCents, &d.TotalCalls,
			&d.TotalInput, &d.TotalOutput); err != nil {
			return nil, fmt.Errorf("scanning daily row: %w", err)
		}
		results = append(results, d)
	}
	return results, rows.Err()
}

func (s *Store) GetDailyByModel(f ReportFilter) ([]DailyModelSpend, error) {
	where, args := buildWhere(f)
	de := dateExpr(f)
	query := fmt.Sprintf(`
		SELECT %s as day, model,
		       SUM(cost_cents), COUNT(*),
		       SUM(input_tokens), SUM(output_tokens)
		FROM api_calls %s
		GROUP BY day, model
		ORDER BY day ASC, SUM(cost_cents) DESC
		LIMIT 900`, de, where)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying daily by model: %w", err)
	}
	defer rows.Close()

	var results []DailyModelSpend
	for rows.Next() {
		var d DailyModelSpend
		if err := rows.Scan(&d.Date, &d.Model, &d.TotalCostCents, &d.TotalCalls,
			&d.TotalInput, &d.TotalOutput); err != nil {
			return nil, fmt.Errorf("scanning daily model row: %w", err)
		}
		results = append(results, d)
	}
	return results, rows.Err()
}

func (s *Store) GetRecentCalls(limit int) ([]APICall, error) {
	rows, err := s.db.Query(`
		SELECT id, timestamp, provider, model,
		       input_tokens, output_tokens, cache_read_tokens, cache_write_tokens,
		       cost_cents, latency_ms, status_code, streaming,
		       COALESCE(tag, ''), COALESCE(user_tag, ''), COALESCE(session_tag, ''),
		       COALESCE(environment, 'default'), COALESCE(endpoint, '')
		FROM api_calls
		ORDER BY timestamp DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("querying recent calls: %w", err)
	}
	defer rows.Close()

	var results []APICall
	for rows.Next() {
		var c APICall
		if err := rows.Scan(
			&c.ID, &c.Timestamp, &c.Provider, &c.Model,
			&c.InputTokens, &c.OutputTokens, &c.CacheReadTokens, &c.CacheWriteTokens,
			&c.CostCents, &c.LatencyMs, &c.StatusCode, &c.Streaming,
			&c.Tag, &c.UserTag, &c.SessionTag, &c.Environment, &c.Endpoint,
		); err != nil {
			return nil, fmt.Errorf("scanning call row: %w", err)
		}
		results = append(results, c)
	}
	return results, rows.Err()
}

// dateExpr returns a SQLite DATE expression adjusted by the filter's timezone offset.
func dateExpr(f ReportFilter) string {
	if f.TZOffset != "" {
		return fmt.Sprintf("DATE(timestamp, '%s')", f.TZOffset)
	}
	return "DATE(timestamp)"
}

func buildWhere(f ReportFilter) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	if !f.Since.IsZero() {
		clauses = append(clauses, "timestamp >= ?")
		args = append(args, f.Since.UTC().Format(time.RFC3339))
	}
	if f.Tag != "" {
		clauses = append(clauses, "tag = ?")
		args = append(args, f.Tag)
	}
	if f.UserTag != "" {
		clauses = append(clauses, "user_tag = ?")
		args = append(args, f.UserTag)
	}
	if f.Environment != "" {
		clauses = append(clauses, "environment = ?")
		args = append(args, f.Environment)
	}
	if f.Provider != "" {
		clauses = append(clauses, "provider = ?")
		args = append(args, f.Provider)
	}
	if f.Model != "" {
		clauses = append(clauses, "model = ?")
		args = append(args, f.Model)
	}

	if len(clauses) == 0 {
		return "", nil
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}
