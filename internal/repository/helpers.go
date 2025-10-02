package repository

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("record not found")

// mapNotFound мапит pgx.ErrNoRows в ErrNotFound
func mapNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
