package repository

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestOrderRepo_Count_WithFilters(t *testing.T) {
	pool := newMockPool(t)
	defer pool.verify(t)

	pool.expectQueryRow(func(_ context.Context, sql string, args []any) (pgx.Row, error) {
		require.Contains(t, sql, "SELECT COUNT(*) FROM orders")
		require.True(t, strings.Contains(sql, "status = $1"))
		require.True(t, strings.Contains(sql, "user_phone ILIKE $2"))
		require.True(t, strings.Contains(sql, "is_read = $3"))
		require.Len(t, args, 3)
		require.Equal(t, "done", args[0])
		require.Equal(t, "%+7999%", args[1])
		require.Equal(t, true, args[2])
		return newSliceRow([]any{5}), nil
	})

	repo := NewOrderRepo(pool)
	total, err := repo.Count(context.Background(), "done", "+7999", ptrBool(true))
	require.NoError(t, err)
	require.Equal(t, 5, total)
}

func TestOrderRepo_List_ReturnsOrders(t *testing.T) {
	pool := newMockPool(t)
	defer pool.verify(t)

	now := time.Now()
	pool.expectQuery(func(_ context.Context, query string, args []any) (pgx.Rows, error) {
		require.Contains(t, query, "FROM orders")
		require.Contains(t, query, "LIMIT $1 OFFSET $2")
		require.Equal(t, []any{10, 5}, args)
		rows := newMockRows([][]any{{
			1,
			models.NullInt32{NullInt32: sql.NullInt32{Int32: 3, Valid: true}},
			"Ivan",
			"+7999",
			"new",
			true,
			now,
		}})
		return rows, nil
	})

	repo := NewOrderRepo(pool)
	orders, err := repo.List(context.Background(), 10, 5, "", "", nil)
	require.NoError(t, err)
	require.Len(t, orders, 1)
	require.Equal(t, "Ivan", orders[0].UserName)
	require.True(t, orders[0].TripID.Valid)
}

func TestOrderRepo_UpdateStatus_NotFound(t *testing.T) {
	pool := newMockPool(t)
	defer pool.verify(t)

	pool.expectExec(func(_ context.Context, query string, args []any) (pgconn.CommandTag, error) {
		require.Contains(t, query, "UPDATE orders SET status")
		require.Equal(t, []any{"canceled", 9}, args)
		return pgconn.NewCommandTag("UPDATE 0"), nil
	})

	repo := NewOrderRepo(pool)
	err := repo.UpdateStatus(context.Background(), 9, "canceled")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNotFound)
}

func ptrBool(v bool) *bool { return &v }
