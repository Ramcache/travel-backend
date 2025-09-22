package repository

import (
	"context"

	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatsRepository struct {
	db *pgxpool.Pool
}

func NewStatsRepository(db *pgxpool.Pool) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) Get(ctx context.Context) (models.Stats, error) {
	var out models.Stats

	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&out.TotalUsers); err != nil {
		return out, err
	}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM news`).Scan(&out.TotalNews); err != nil {
		return out, err
	}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM trips`).Scan(&out.TotalTrips); err != nil {
		return out, err
	}

	qKV := func(sql string) ([]models.KV, error) {
		rows, err := r.db.Query(ctx, sql)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var res []models.KV
		for rows.Next() {
			var kv models.KV
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
