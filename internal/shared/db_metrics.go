package shared

import (
	"context"
	"database/sql"
	"time"
)

type MetricsDB struct {
	db      *sql.DB
	metrics *AppMetrics
}

func NewMetricsDB(db *sql.DB, metrics *AppMetrics) *MetricsDB {
	return &MetricsDB{
		db:      db,
		metrics: metrics,
	}
}

func (m *MetricsDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := m.db.QueryRowContext(ctx, query, args...)

	m.metrics.RecordDatabaseOperation(ctx, "query_row", "unknown")

	// Log query duration if needed
	_ = time.Since(start)
	return row
}

func (m *MetricsDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := m.db.QueryContext(ctx, query, args...)

	m.metrics.RecordDatabaseOperation(ctx, "query", "unknown")

	// Log query duration if needed
	_ = time.Since(start)
	return rows, err
}

func (m *MetricsDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := m.db.ExecContext(ctx, query, args...)

	m.metrics.RecordDatabaseOperation(ctx, "exec", "unknown")

	// Log query duration if needed
	_ = time.Since(start)
	return result, err
}

func (m *MetricsDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return m.db.PrepareContext(ctx, query)
}

func (m *MetricsDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return m.db.BeginTx(ctx, opts)
}

func (m *MetricsDB) Close() error {
	return m.db.Close()
}

func (m *MetricsDB) PingContext(ctx context.Context) error {
	return m.db.PingContext(ctx)
}

func (m *MetricsDB) SetMaxOpenConns(n int) {
	m.db.SetMaxOpenConns(n)
}

func (m *MetricsDB) SetMaxIdleConns(n int) {
	m.db.SetMaxIdleConns(n)
}

func (m *MetricsDB) SetConnMaxLifetime(d time.Duration) {
	m.db.SetConnMaxLifetime(d)
}

func (m *MetricsDB) Stats() sql.DBStats {
	return m.db.Stats()
}
