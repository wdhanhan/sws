package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/starfall-warsong/sws/internal/engine"
	"github.com/starfall-warsong/sws/internal/repository"
	"github.com/starfall-warsong/sws/pkg/config"
	"github.com/starfall-warsong/sws/pkg/database"
	"github.com/starfall-warsong/sws/pkg/logger"
)

func main() {
	count := flag.Int("systems", 1000, "number of star systems to generate")
	seed := flag.Int64("seed", 20260428, "master seed for generation")
	flag.Parse()

	logger.Init("debug")
	cfg := config.Load()

	db, err := database.NewPostgres(&cfg.Database)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	slog.Info("generating starmap", "systems", *count, "seed", *seed)
	start := time.Now()

	gen := engine.NewStarmapGenerator(*seed)
	starmap := gen.Generate(*count)

	slog.Info("generation complete",
		"systems", len(starmap.Systems),
		"gates", len(starmap.Gates),
		"planets", len(starmap.Planets),
		"belts", len(starmap.Belts),
		"duration", time.Since(start),
	)

	ctx := context.Background()
	repo := repository.NewStarmapRepo(db)

	slog.Info("inserting systems...")
	if err := repo.BulkInsertSystems(ctx, starmap.Systems); err != nil {
		log.Fatalf("insert systems failed: %v", err)
	}

	slog.Info("inserting gates...")
	if err := repo.BulkInsertGates(ctx, starmap.Gates); err != nil {
		log.Fatalf("insert gates failed: %v", err)
	}

	slog.Info("inserting planets...")
	if err := repo.BulkInsertPlanets(ctx, starmap.Planets); err != nil {
		log.Fatalf("insert planets failed: %v", err)
	}

	slog.Info("inserting belts...")
	if err := repo.BulkInsertBelts(ctx, starmap.Belts); err != nil {
		log.Fatalf("insert belts failed: %v", err)
	}

	slog.Info("starmap generation and insertion complete",
		"total_duration", time.Since(start),
	)

	var sysCount int
	db.Get(&sysCount, "SELECT COUNT(*) FROM star_systems")
	var gateCount int
	db.Get(&gateCount, "SELECT COUNT(*) FROM stargates")

	fmt.Printf("\n=== 星陨战歌 · 星图生成完毕 ===\n")
	fmt.Printf("星系总数: %d\n", sysCount)
	fmt.Printf("星门总数: %d\n", gateCount)
	fmt.Printf("行星总数: %d\n", len(starmap.Planets))
	fmt.Printf("矿带总数: %d\n", len(starmap.Belts))
}
