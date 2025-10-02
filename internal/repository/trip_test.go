package repository

import (
	"context"
	"github.com/Ramcache/travel-backend/internal/models"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestTripRepository_List_WithFilters(t *testing.T) {
	pool := newMockPool(t)
	defer pool.verify(t)

	now := time.Now()
	deadline := now.Add(24 * time.Hour)

	pool.expectQuery(func(_ context.Context, sql string, args []any) (pgx.Rows, error) {
		require.Contains(t, sql, "FROM trips")
		require.Contains(t, sql, "WHERE")

		require.Equal(t, "Moscow", args[0])
		require.Equal(t, "пляжный", args[1])
		require.Equal(t, 10, args[2]) // Limit всегда идёт после фильтров

		rows := newMockRows([][]any{{
			1,
			"Trip",
			"Desc",
			"photo.jpg",
			"Moscow",
			"пляжный",
			"summer",
			1500.0,
			"RUB",
			now,
			now.Add(48 * time.Hour),
			&deadline,
			true, // main
			true, // active
			10,   // views_count
			2,    // buys_count
			now,
			now,
		}})
		return rows, nil
	})

	repo := NewTripRepository(pool)

	filter := models.TripFilter{
		DepartureCity: "Moscow",
		TripType:      "пляжный",
		Limit:         10,
		Offset:        0,
	}

	trips, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	require.Len(t, trips, 1)
	require.Equal(t, "Trip", trips[0].Title)
	require.True(t, trips[0].Main)
	require.True(t, trips[0].Active)
}

func TestTripRepository_GetByID_NotFound(t *testing.T) {
	pool := newMockPool(t)
	defer pool.verify(t)

	pool.expectQueryRow(func(_ context.Context, sql string, args []any) (pgx.Row, error) {
		require.True(t, strings.Contains(sql, "FROM trips"))
		require.Equal(t, []any{99}, args)
		return nil, pgx.ErrNoRows
	})

	repo := NewTripRepository(pool)
	trip, err := repo.GetByID(context.Background(), 99)
	require.Nil(t, trip)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestTripRepository_Delete_NotFound(t *testing.T) {
	pool := newMockPool(t)
	defer pool.verify(t)

	pool.expectExec(func(_ context.Context, sql string, args []any) (pgconn.CommandTag, error) {
		require.Contains(t, sql, "DELETE FROM trips")
		require.Equal(t, []any{5}, args)
		return pgconn.NewCommandTag("DELETE 0"), nil
	})

	repo := NewTripRepository(pool)
	err := repo.Delete(context.Background(), 5)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestTripRepository_ResetMain_WithExclude(t *testing.T) {
	pool := newMockPool(t)
	defer pool.verify(t)

	pool.expectExec(func(_ context.Context, sql string, args []any) (pgconn.CommandTag, error) {
		require.Contains(t, sql, "UPDATE trips SET main=false WHERE id <> $1")
		require.Equal(t, []any{7}, args)
		return pgconn.NewCommandTag("UPDATE 3"), nil
	})

	repo := NewTripRepository(pool)
	id := 7
	require.NoError(t, repo.ResetMain(context.Background(), &id))
}

func TestTripRepository_Popular(t *testing.T) {
	pool := newMockPool(t)
	defer pool.verify(t)

	now := time.Now()

	pool.expectQuery(func(_ context.Context, sql string, args []any) (pgx.Rows, error) {
		require.Contains(t, sql, "ORDER BY buys_count DESC")
		require.Equal(t, []any{3}, args)

		rows := newMockRows([][]any{{
			1,
			"Trip",
			"Desc",
			"photo.jpg",
			"Moscow",
			"пляжный",
			"summer",
			2500.0,
			"RUB",
			now,
			now.Add(72 * time.Hour),
			(*time.Time)(nil),
			false, // main
			true,  // active
			100,   // views_count
			20,    // buys_count
			now,
			now,
		}})
		return rows, nil
	})

	repo := NewTripRepository(pool)
	trips, err := repo.Popular(context.Background(), 3)
	require.NoError(t, err)
	require.Len(t, trips, 1)
	require.Equal(t, 100, trips[0].ViewsCount)
	require.Equal(t, 20, trips[0].BuysCount)
	require.True(t, trips[0].Active)
}
