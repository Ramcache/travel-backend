package repository

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type mockPool struct {
	t             *testing.T
	queryCalls    []func(context.Context, string, []any) (pgx.Rows, error)
	queryRowCalls []func(context.Context, string, []any) (pgx.Row, error)
	execCalls     []func(context.Context, string, []any) (pgconn.CommandTag, error)

	queryIdx    int
	queryRowIdx int
	execIdx     int
}

func newMockPool(t *testing.T) *mockPool {
	t.Helper()
	return &mockPool{t: t}
}

func (m *mockPool) expectQuery(fn func(context.Context, string, []any) (pgx.Rows, error)) {
	m.queryCalls = append(m.queryCalls, fn)
}

func (m *mockPool) expectQueryRow(fn func(context.Context, string, []any) (pgx.Row, error)) {
	m.queryRowCalls = append(m.queryRowCalls, fn)
}

func (m *mockPool) expectExec(fn func(context.Context, string, []any) (pgconn.CommandTag, error)) {
	m.execCalls = append(m.execCalls, fn)
}

func (m *mockPool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.queryIdx >= len(m.queryCalls) {
		m.t.Fatalf("unexpected Query call: %s", sql)
	}
	fn := m.queryCalls[m.queryIdx]
	m.queryIdx++
	return fn(ctx, sql, args)
}

func (m *mockPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
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

func (m *mockPool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.execIdx >= len(m.execCalls) {
		m.t.Fatalf("unexpected Exec call: %s", sql)
	}
	fn := m.execCalls[m.execIdx]
	m.execIdx++
	return fn(ctx, sql, args)
}

func (m *mockPool) verify(t *testing.T) {
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

type errorRow struct {
	err error
}

func (r *errorRow) Scan(dest ...any) error { return r.err }

type sliceRow struct {
	values []any
	err    error
}

func newSliceRow(values []any) *sliceRow {
	return &sliceRow{values: values}
}

func (r *sliceRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	return assignValues(dest, r.values)
}

type mockRows struct {
	data [][]any
	idx  int
	err  error
}

func newMockRows(data [][]any) *mockRows { return &mockRows{data: data} }

func (r *mockRows) Close()                                       {}
func (r *mockRows) Err() error                                   { return r.err }
func (r *mockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *mockRows) RawValues() [][]byte                          { return nil }
func (r *mockRows) Conn() *pgx.Conn                              { return nil }

func (r *mockRows) Next() bool {
	if r.err != nil {
		return false
	}
	if r.idx >= len(r.data) {
		return false
	}
	r.idx++
	return true
}

func (r *mockRows) Scan(dest ...any) error {
	if r.idx == 0 || r.idx > len(r.data) {
		return fmt.Errorf("scan called without Next")
	}
	row := r.data[r.idx-1]
	return assignValues(dest, row)
}

func (r *mockRows) Values() ([]any, error) {
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
		return fmt.Errorf("dest must be non-nil pointer, got %T", dest)
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
