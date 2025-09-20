package main

import (
	"fmt"
	"github.com/Ramcache/travel-backend/internal/cli"
	"os"
)

func main() {
	rootCmd := cli.NewRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(" error:", err)
		os.Exit(1)
	}
}
