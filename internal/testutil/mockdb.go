package testutil

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// MockDB implements repository.DB and records expectations for queries executed
// by services and handlers during tests.
type MockDB struct {
	t             *testing.T
	now           time.Time
	queryCalls    []func(context.Context, string, []any) (pgx.Rows, error)
	queryRowCalls []func(context.Context, string, []any) (pgx.Row, error)
	execCalls     []func(context.Context, string, []any) (pgconn.CommandTag, error)

	queryIdx    int
	queryRowIdx int
	execIdx     int
}

// NewMockDB creates a MockDB bound to the provided testing.T.
func NewMockDB(t *testing.T) *MockDB {
	t.Helper()
	return &MockDB{t: t, now: time.Now()}
}

// Now returns the deterministic timestamp used when building mock rows.
func (m *MockDB) Now() time.Time { return m.now }

// ExpectQuery registers a callback that will be invoked for the next Query call.
func (m *MockDB) ExpectQuery(fn func(context.Context, string, []any) (pgx.Rows, error)) {
	m.queryCalls = append(m.queryCalls, fn)
}

// ExpectQueryRow registers a callback for the next QueryRow call.
func (m *MockDB) ExpectQueryRow(fn func(context.Context, string, []any) (pgx.Row, error)) {
	m.queryRowCalls = append(m.queryRowCalls, fn)
}

// ExpectExec registers a callback for the next Exec call.
func (m *MockDB) ExpectExec(fn func(context.Context, string, []any) (pgconn.CommandTag, error)) {
	m.execCalls = append(m.execCalls, fn)
}

// Verify ensures that all registered expectations were satisfied.
func (m *MockDB) Verify(t *testing.T) {
	t.Helper()
	if m.queryIdx != len(m.queryCalls) {
		t.Fatalf("expected %d Query calls, got %d", len(m.queryCalls), m.queryIdx)
	}
	if m.queryRowIdx != len(m.queryRowCalls) {
		t.Fatalf("expected %d QueryRow calls, got %d", len(m.queryRowCalls), m.queryRowIdx)
	}
	if m.execIdx != len(m.execCalls) {
		t.Fatalf("expected %d Exec calls, got %d", len(m.execCalls), m.execIdx)
	}
}

func (m *MockDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.queryIdx >= len(m.queryCalls) {
		m.t.Fatalf("unexpected Query call: %s", sql)
	}
	fn := m.queryCalls[m.queryIdx]
	m.queryIdx++
	return fn(ctx, sql, args)
}

func (m *MockDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRowIdx >= len(m.queryRowCalls) {
		m.t.Fatalf("unexpected QueryRow call: %s", sql)
	}
	fn := m.queryRowCalls[m.queryRowIdx]
	m.queryRowIdx++
	row, err := fn(ctx, sql, args)
	if err != nil {
		return &errorRow{err: err}
	}
	return row
}

func (m *MockDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.execIdx >= len(m.execCalls) {
		m.t.Fatalf("unexpected Exec call: %s", sql)
	}
	fn := m.execCalls[m.execIdx]
	m.execIdx++
	return fn(ctx, sql, args)
}

// Row is the minimal interface returned by ExpectQueryRow callbacks.
type Row interface{ Scan(dest ...any) error }

// Rows is the minimal interface returned by ExpectQuery callbacks.
type Rows interface {
	pgx.Rows
}

type errorRow struct{ err error }

func (r *errorRow) Scan(dest ...any) error { return r.err }

// SliceRow is a helper implementing pgx.Row backed by a slice of values.
type SliceRow struct {
	values []any
	err    error
}

// NewSliceRow creates a row that scans the provided values into destination pointers.
func NewSliceRow(values []any) *SliceRow { return &SliceRow{values: values} }

func (r *SliceRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	return assignValues(dest, r.values)
}

// MockRows implements pgx.Rows for deterministic result sets.
type MockRows struct {
	data [][]any
	idx  int
	err  error
}

// NewMockRows constructs a MockRows instance that iterates over the provided data.
func NewMockRows(data [][]any) *MockRows { return &MockRows{data: data} }

func (r *MockRows) Close()                                       {}
func (r *MockRows) Err() error                                   { return r.err }
func (r *MockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *MockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *MockRows) RawValues() [][]byte                          { return nil }
func (r *MockRows) Conn() *pgx.Conn                              { return nil }

func (r *MockRows) Next() bool {
	if r.err != nil {
		return false
	}
	if r.idx >= len(r.data) {
		return false
	}
	r.idx++
	return true
}

func (r *MockRows) Scan(dest ...any) error {
	if r.idx == 0 || r.idx > len(r.data) {
		return fmt.Errorf("scan called without Next")
	}
	return assignValues(dest, r.data[r.idx-1])
}

func (r *MockRows) Values() ([]any, error) {
	if r.idx == 0 || r.idx > len(r.data) {
		return nil, fmt.Errorf("values called without Next")
	}
	row := r.data[r.idx-1]
	out := make([]any, len(row))
	copy(out, row)
	return out, nil
}

func assignValues(dest []any, values []any) error {
	if len(dest) != len(values) {
		return fmt.Errorf("expected %d dest values, got %d", len(values), len(dest))
	}
	for i := range dest {
		if err := assignValue(dest[i], values[i]); err != nil {
			return fmt.Errorf("assign column %d: %w", i, err)
		}
	}
	return nil
}

func assignValue(dest any, value any) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("dest must be pointer")
	}
	target := rv.Elem()
	if value == nil {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}
	sv := reflect.ValueOf(value)
	if sv.Type().AssignableTo(target.Type()) {
		target.Set(sv)
		return nil
	}
	if sv.Type().ConvertibleTo(target.Type()) {
		target.Set(sv.Convert(target.Type()))
		return nil
	}
	return fmt.Errorf("cannot assign %T to %s", value, target.Type())
}
