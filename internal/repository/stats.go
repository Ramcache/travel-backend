package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StatsRepository struct {
	db *pgxpool.Pool
}

func NewStatsRepository(db *pgxpool.Pool) *StatsRepository {
	return &StatsRepository{db: db}
}

type KV struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

type Stats struct {
	TotalUsers     int64 `json:"total_users"`
	TotalNews      int64 `json:"total_news"`
	TotalTrips     int64 `json:"total_trips"`
	UsersByRole    []KV  `json:"users_by_role"`
	NewsByStatus   []KV  `json:"news_by_status"`
	NewsByCategory []KV  `json:"news_by_category"`
	TripsByType    []KV  `json:"trips_by_type"`
	TripsByCity    []KV  `json:"trips_by_city"`
}

func (r *StatsRepository) Get(ctx context.Context) (Stats, error) {
	var out Stats

	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&out.TotalUsers); err != nil {
		return out, err
	}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM news`).Scan(&out.TotalNews); err != nil {
		return out, err
	}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM trips`).Scan(&out.TotalTrips); err != nil {
		return out, err
	}

	qKV := func(sql string) ([]KV, error) {
		rows, err := r.db.Query(ctx, sql)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var res []KV
		for rows.Next() {
			var kv KV
			if err := rows.Scan(&kv.Key, &kv.Count); err != nil {
				return nil, err
			}
			res = append(res, kv)
		}
		return res, rows.Err()
	}

	var err error
	if out.UsersByRole, err = qKV(`SELECT role, cnt FROM v_users_by_role`); err != nil {
		return out, err
	}
	if out.NewsByStatus, err = qKV(`SELECT status, cnt FROM v_news_by_status`); err != nil {
		return out, err
	}
	if out.NewsByCategory, err = qKV(`SELECT category, cnt FROM v_news_by_category`); err != nil {
		return out, err
	}
	if out.TripsByType, err = qKV(`SELECT trip_type, cnt FROM v_trips_by_type`); err != nil {
		return out, err
	}
	if out.TripsByCity, err = qKV(`SELECT departure_city, cnt FROM v_trips_by_city`); err != nil {
		return out, err
	}

	return out, nil
}
