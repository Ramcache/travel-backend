package main

import (
	"fmt"
	"github.com/Ramcache/travel-backend/internal/cli"
	"github.com/Ramcache/travel-backend/internal/validators"
	"os"
)

// @title        Travel API
// @version      1.0
// @description  API for Travel project
// @BasePath     https://api.web95.tech/api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() {
	rootCmd := cli.NewRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(" error:", err)
		os.Exit(1)
	}
	validators.Init()

}
