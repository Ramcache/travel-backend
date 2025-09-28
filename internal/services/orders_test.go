package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"

	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
	"github.com/Ramcache/travel-backend/internal/testutil"
)

func TestOrderService_Create(t *testing.T) {
	pool := testutil.NewMockDB(t)
	defer pool.Verify(t)

	createdAt := pool.Now()
	pool.ExpectQueryRow(func(_ context.Context, query string, args []any) (pgx.Row, error) {
		require.Contains(t, query, "INSERT INTO orders")
		require.Len(t, args, 4)
		tripArg, ok := args[0].(sql.NullInt32)
		require.True(t, ok)
		require.Equal(t, int32(7), tripArg.Int32)
		require.True(t, tripArg.Valid)
		require.Equal(t, "Ivan", args[1])
		require.Equal(t, "+7999", args[2])
		require.Equal(t, "new", args[3])
		return testutil.NewSliceRow([]any{42, createdAt}), nil
	})

	repo := repository.NewOrderRepo(pool)
	svc := NewOrderService(repo)

	order, err := svc.Create(context.Background(), 7, "Ivan", "+7999")
	require.NoError(t, err)
	require.Equal(t, 42, order.ID)
	require.Equal(t, "new", order.Status)
	require.Equal(t, "+7999", order.UserPhone)
}

func TestOrderService_List(t *testing.T) {
	pool := testutil.NewMockDB(t)
	defer pool.Verify(t)

	pool.ExpectQueryRow(func(_ context.Context, query string, _ []any) (pgx.Row, error) {
		require.Contains(t, query, "SELECT COUNT(*) FROM orders")
		return testutil.NewSliceRow([]any{2}), nil
	})

	now := pool.Now()
	pool.ExpectQuery(func(_ context.Context, query string, args []any) (pgx.Rows, error) {
		require.Contains(t, query, "LIMIT $1 OFFSET $2")
		require.Equal(t, []any{15, 30}, args)
		rows := testutil.NewMockRows([][]any{{
			1,
			models.NullInt32{NullInt32: sql.NullInt32{Valid: false}},
			"Maria",
			"+7888",
			"new",
			false,
			now,
		}})
		return rows, nil
	})

	repo := repository.NewOrderRepo(pool)
	svc := NewOrderService(repo)

	result, err := svc.List(context.Background(), 15, 30, "", "", nil)
	require.NoError(t, err)
	require.Equal(t, 2, result.Total)
	require.Len(t, result.Orders, 1)
	require.Equal(t, "Maria", result.Orders[0].UserName)
}

func TestOrderService_UpdateStatus_Error(t *testing.T) {
	pool := testutil.NewMockDB(t)
	defer pool.Verify(t)

	pool.ExpectExec(func(_ context.Context, query string, args []any) (pgconn.CommandTag, error) {
		require.Contains(t, query, "UPDATE orders SET status")
		require.Equal(t, []any{"done", 5}, args)
		return pgconn.CommandTag{}, errors.New("boom")
	})

	repo := repository.NewOrderRepo(pool)
	svc := NewOrderService(repo)

	err := svc.UpdateStatus(context.Background(), 5, "done")
	require.EqualError(t, err, "boom")
}
