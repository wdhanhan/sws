package main

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/starfall-warsong/sws/internal/handler"
	"github.com/starfall-warsong/sws/internal/repository"
	"github.com/starfall-warsong/sws/internal/service"
	"github.com/starfall-warsong/sws/pkg/auth"
	"github.com/starfall-warsong/sws/pkg/cache"
	"github.com/starfall-warsong/sws/pkg/config"
	"github.com/starfall-warsong/sws/pkg/database"
	"github.com/starfall-warsong/sws/pkg/logger"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg.Server.Mode)
	slog.Info("starting Starfall Warsong server", "mode", cfg.Server.Mode)

	db, err := database.NewPostgres(&cfg.Database)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	rdb, err := cache.NewRedis(&cfg.Redis)
	if err != nil {
		slog.Warn("redis connection failed, running without cache", "error", err)
	} else {
		defer rdb.Close()
	}
	_ = rdb

	jwtManager := auth.NewJWTManager(&cfg.JWT)

	accountRepo := repository.NewAccountRepo(db)
	characterRepo := repository.NewCharacterRepo(db)
	starmapRepo := repository.NewStarmapRepo(db)
	inventoryRepo := repository.NewInventoryRepo(db)

	accountService := service.NewAccountService(accountRepo, jwtManager)
	characterService := service.NewCharacterService(characterRepo, accountRepo)
	miningService := service.NewMiningService(db, starmapRepo, inventoryRepo)
	refineService := service.NewRefineService(db, inventoryRepo)
	marketService := service.NewMarketService(db, inventoryRepo)

	combatService := service.NewCombatService(db, inventoryRepo)
	skillService := service.NewSkillService(db)
	deathService := service.NewDeathService(db)
	corpService := service.NewCorpService(db)
	buildingService := service.NewBuildingService(db)
	encounterService := service.NewEncounterService(db, inventoryRepo)
	shipService := service.NewShipService(db)
	dungeonService := service.NewDungeonService(db, inventoryRepo)
	combatSiteService := service.NewCombatSiteService(db, inventoryRepo)

	accountHandler := handler.NewAccountHandler(accountService)
	characterHandler := handler.NewCharacterHandler(characterService)
	starmapHandler := handler.NewStarmapHandler(starmapRepo)
	economyHandler := handler.NewEconomyHandler(miningService, refineService, marketService, inventoryRepo)
	combatHandler := handler.NewCombatHandler(combatService)
	skillHandler := handler.NewSkillHandler(skillService, deathService)
	orgHandler := handler.NewOrgHandler(corpService)
	worldHandler := handler.NewWorldHandler(buildingService, encounterService)
	shipHandler := handler.NewShipHandler(shipService)
	dungeonHandler := handler.NewDungeonHandler(dungeonService)
	combatSiteHandler := handler.NewCombatSiteHandler(combatSiteService)

	router := handler.SetupRouter(cfg.Server.Mode, jwtManager, accountHandler, characterHandler, starmapHandler, economyHandler, combatHandler, skillHandler, orgHandler, worldHandler, shipHandler, dungeonHandler, combatSiteHandler)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	slog.Info("server listening", "addr", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
