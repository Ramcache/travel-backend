package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/Ramcache/travel-backend/internal/handlers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
	"github.com/Ramcache/travel-backend/internal/services"
	"github.com/Ramcache/travel-backend/internal/testutil"
)

func TestOrderHandler_List_Defaults(t *testing.T) {
	pool := testutil.NewMockDB(t)
	defer pool.Verify(t)

	pool.ExpectQueryRow(func(_ context.Context, sql string, _ []any) (pgx.Row, error) {
		require.Contains(t, sql, "SELECT COUNT(*) FROM orders")
		return testutil.NewSliceRow([]any{1}), nil
	})

	now := pool.Now()
	pool.ExpectQuery(func(_ context.Context, sql string, args []any) (pgx.Rows, error) {
		require.Contains(t, sql, "LIMIT $1 OFFSET $2")
		require.Equal(t, []any{20, 0}, args)
		rows := testutil.NewMockRows([][]any{{
			1,
			models.NullInt32{},
			"Ivan",
			"+7999",
			"new",
			false,
			now,
		}})
		return rows, nil
	})

	repo := repository.NewOrderRepo(pool)
	svc := services.NewOrderService(repo)
	handler := handlers.NewOrderHandler(svc, zaptest.NewLogger(t).Sugar())

	req := httptest.NewRequest(http.MethodGet, "/admin/orders", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data services.OrdersWithTotal `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, 1, resp.Data.Total)
	require.Len(t, resp.Data.Orders, 1)
	require.Equal(t, "Ivan", resp.Data.Orders[0].UserName)
}

func TestOrderHandler_UpdateStatus_InvalidID(t *testing.T) {
	pool := testutil.NewMockDB(t)
	defer pool.Verify(t)
	svc := services.NewOrderService(repository.NewOrderRepo(pool))
	handler := handlers.NewOrderHandler(svc, zaptest.NewLogger(t).Sugar())

	req := httptest.NewRequest(http.MethodPost, "/admin/orders/abc/status", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.UpdateStatus(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_UpdateStatus_NotFound(t *testing.T) {
	pool := testutil.NewMockDB(t)
	defer pool.Verify(t)

	pool.ExpectExec(func(_ context.Context, sql string, args []any) (pgconn.CommandTag, error) {
		require.Contains(t, sql, "UPDATE orders SET status")
		require.Equal(t, []any{"done", 5}, args)
		return pgconn.NewCommandTag("UPDATE 0"), nil
	})

	repo := repository.NewOrderRepo(pool)
	svc := services.NewOrderService(repo)
	handler := handlers.NewOrderHandler(svc, zaptest.NewLogger(t).Sugar())

	req := httptest.NewRequest(http.MethodPost, "/admin/orders/5/status?status=done", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.UpdateStatus(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_MarkAsRead_Success(t *testing.T) {
	pool := testutil.NewMockDB(t)
	defer pool.Verify(t)

	pool.ExpectExec(func(_ context.Context, sql string, args []any) (pgconn.CommandTag, error) {
		require.Contains(t, sql, "UPDATE orders SET is_read = true")
		require.Equal(t, []any{3}, args)
		return pgconn.NewCommandTag("UPDATE 1"), nil
	})

	repo := repository.NewOrderRepo(pool)
	svc := services.NewOrderService(repo)
	handler := handlers.NewOrderHandler(svc, zaptest.NewLogger(t).Sugar())

	req := httptest.NewRequest(http.MethodPost, "/admin/orders/3/read", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "3")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.MarkAsRead(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestOrderHandler_Delete_NotFound(t *testing.T) {
	pool := testutil.NewMockDB(t)
	defer pool.Verify(t)

	pool.ExpectExec(func(_ context.Context, sql string, args []any) (pgconn.CommandTag, error) {
		require.Contains(t, sql, "DELETE FROM orders")
		require.Equal(t, []any{11}, args)
		return pgconn.NewCommandTag("DELETE 0"), nil
	})

	repo := repository.NewOrderRepo(pool)
	svc := services.NewOrderService(repo)
	handler := handlers.NewOrderHandler(svc, zaptest.NewLogger(t).Sugar())

	req := httptest.NewRequest(http.MethodDelete, "/admin/orders/11", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "11")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.Delete(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}
