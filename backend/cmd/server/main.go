// @title           Maturin API
// @version         1.0
// @description     B2B platform REST API
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the access token

package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/sskotezhov/maturin/config"
	_ "github.com/sskotezhov/maturin/docs"
	"github.com/sskotezhov/maturin/internal/app"
	"github.com/sskotezhov/maturin/pkg/cache"
	"github.com/sskotezhov/maturin/pkg/database"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.Connect(&cfg.Maturin.Database)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	rdb, err := cache.Connect(&cfg.Maturin.Redis)
	if err != nil {
		log.Fatalf("cache: %v", err)
	}

	a, err := app.New(cfg, db, rdb)
	if err != nil {
		log.Fatalf("app: %v", err)
	}

	log.Fatal(a.Echo.Start(":" + cfg.Maturin.Server.Port))
}
