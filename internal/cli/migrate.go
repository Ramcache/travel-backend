package cli

import (
	"database/sql"
	"fmt"
	"github.com/pressly/goose"
	"log"

	"github.com/spf13/cobra"

	"github.com/Ramcache/travel-backend/internal/config"

	_ "github.com/lib/pq"
)

func NewMigrateCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:   "migrate [command]",
		Short: "Run database migrations",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()

			db, err := sql.Open("postgres", cfg.DBUrl)
			if err != nil {
				log.Fatalf("db open: %v", err)
			}
			defer db.Close()

			if err := goose.SetDialect("postgres"); err != nil {
				log.Fatalf("goose dialect: %v", err)
			}

			// путь к миграциям
			if dir == "" {
				dir = "migrations"
			}

			action := args[0]
			switch action {
			case "up":
				if err := goose.Up(db, dir); err != nil {
					log.Fatalf("migrate up: %v", err)
				}
			case "down":
				if err := goose.Down(db, dir); err != nil {
					log.Fatalf("migrate down: %v", err)
				}
			case "status":
				if err := goose.Status(db, dir); err != nil {
					log.Fatalf("migrate status: %v", err)
				}
			default:
				fmt.Println("unknown migrate command. use: up | down | status")
			}
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "migrations", "directory with migration files")

	return cmd
}
