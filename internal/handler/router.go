package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/starfall-warsong/sws/internal/middleware"
	"github.com/starfall-warsong/sws/pkg/auth"
)

func SetupRouter(
	mode string,
	jwtManager *auth.JWTManager,
	accountHandler *AccountHandler,
	characterHandler *CharacterHandler,
	starmapHandler *StarmapHandler,
	economyHandler *EconomyHandler,
	combatHandler *CombatHandler,
	skillHandler *SkillHandler,
	orgHandler *OrgHandler,
	worldHandler *WorldHandler,
	shipHandler *ShipHandler,
	dungeonHandler *DungeonHandler,
	combatSiteHandler *CombatSiteHandler,
	combatSiteWS *CombatSiteWS,
	fleetHandler *FleetHandler,
	wreckHandler *WreckHandler,
) *gin.Engine {
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "game": "星陨战歌"})
	})

	api := r.Group("/api/v1")
	{
		api.POST("/login", accountHandler.Login)
	}

	starmap := api.Group("/starmap")
	{
		starmap.GET("/systems/:id", starmapHandler.GetSystem)
		starmap.GET("/systems", starmapHandler.SearchSystems)
		starmap.GET("/stats", starmapHandler.GetStats)
	}

	authed := api.Group("")
	authed.Use(middleware.JWTAuth(jwtManager))
	{
		chars := authed.Group("/characters")
		{
			chars.POST("", characterHandler.Create)
			chars.GET("", characterHandler.List)
			chars.GET("/:id", characterHandler.Get)
			chars.DELETE("/:id", characterHandler.Delete)
		}

		// 舰船
		ships := authed.Group("/ships")
		{
			ships.GET("", shipHandler.GetMyShips)
			ships.POST("/board", shipHandler.BoardShip)
			ships.POST("/fit", shipHandler.FitModule)
			ships.DELETE("/fit", shipHandler.RemoveModule)
			ships.GET("/:id/fitting", shipHandler.GetFitting)
		}

		// 经济系统
		mining := authed.Group("/mining")
		{
			mining.POST("/start", economyHandler.StartMining)
			mining.POST("/collect", economyHandler.CollectMining)
			mining.POST("/stop", economyHandler.StopMining)
		}

		refine := authed.Group("/refine")
		{
			refine.POST("", economyHandler.Refine)
			refine.GET("/recipes", economyHandler.GetRefineRecipes)
		}

		inv := authed.Group("/inventory")
		{
			inv.GET("", economyHandler.GetInventory)
			inv.POST("/transfer", economyHandler.TransferItem)
			inv.GET("/assets", economyHandler.GetAssets)
		}

		market := authed.Group("/market")
		{
			market.POST("/sell", economyHandler.CreateSellOrder)
			market.POST("/buy", economyHandler.CreateBuyOrder)
			market.POST("/orders/:id/buy", economyHandler.BuyFromOrder)
			market.GET("/orders", economyHandler.SearchOrders)
			market.DELETE("/orders/:id", economyHandler.CancelOrder)
		}

		// 技能系统
		skills := authed.Group("/skills")
		{
			skills.GET("", skillHandler.GetCharacterSkills)
			skills.POST("/train", skillHandler.AddToQueue)
			skills.GET("/queue", skillHandler.GetQueue)
		}

		// 死亡/克隆
		authed.POST("/death", skillHandler.ProcessDeath)

		// 军团
		corps := authed.Group("/corps")
		{
			corps.POST("", orgHandler.CreateCorp)
			corps.POST("/:id/join", orgHandler.JoinCorp)
			corps.POST("/leave", orgHandler.LeaveCorp)
			corps.GET("/mine", orgHandler.GetMyCorp)
		}

		// 联盟
		authed.POST("/alliances", orgHandler.CreateAlliance)

		// 聊天
		chat := authed.Group("/chat")
		{
			chat.POST("/send", orgHandler.SendChat)
			chat.GET("/messages", orgHandler.GetChat)
		}

		// 邮件
		mail := authed.Group("/mail")
		{
			mail.POST("/send", orgHandler.SendMail)
			mail.GET("", orgHandler.GetMails)
		}

		// 战斗系统
		combat := authed.Group("/combat")
		{
			combat.GET("/scan", combatHandler.ScanEnemies)
			combat.POST("/engage", combatHandler.EngageNPC)
			combat.POST("/tick", combatHandler.NextTick)
			combat.GET("/state", combatHandler.GetState)
			combat.POST("/command", combatHandler.Command)
			combat.POST("/auto", combatHandler.AutoFight)
		}

		// 建筑
		bld := authed.Group("/buildings")
		{
			bld.POST("/deploy", worldHandler.DeployBuilding)
		}

		// 奇遇
		enc := authed.Group("/encounters")
		{
			enc.POST("/try", worldHandler.TryEncounter)
			enc.POST("/choose", worldHandler.MakeChoice)
		}

		// NPC商店
		authed.POST("/shop/buy", economyHandler.NPCShop)

		// 战斗地点
		sites := authed.Group("/sites")
		{
			sites.GET("", combatSiteHandler.ListSites)
			sites.POST("/scan", combatSiteHandler.Scan)
			sites.POST("/enter", combatSiteHandler.Enter)
			sites.POST("/tick", combatSiteHandler.Tick)
			sites.POST("/auto", combatSiteHandler.AutoWave)
			sites.POST("/leave", combatSiteHandler.Leave)
		}

		// 残骸
		authed.GET("/wrecks", wreckHandler.ListWrecks)
		authed.POST("/wrecks/:id/loot", wreckHandler.LootWreck)

		// 舰队
		fleet := authed.Group("/fleet")
		{
			fleet.POST("/create", fleetHandler.Create)
			fleet.POST("/invite", fleetHandler.Invite)
			fleet.POST("/accept", fleetHandler.Accept)
			fleet.POST("/decline", fleetHandler.Decline)
			fleet.POST("/leave", fleetHandler.Leave)
			fleet.POST("/kick", fleetHandler.Kick)
			fleet.POST("/disband", fleetHandler.Disband)
			fleet.GET("", fleetHandler.GetFleet)
		}

		// 远征
		dng := authed.Group("/dungeons")
		{
			dng.GET("", dungeonHandler.ListDungeons)
			dng.POST("/enter", dungeonHandler.Enter)
			dng.GET("/status", dungeonHandler.GetStatus)
			dng.POST("/fight", dungeonHandler.FightWave)
			dng.POST("/leave", dungeonHandler.Leave)
		}
	}

	// WebSocket (auth handled inside handler)
	r.GET("/ws/sites", gin.WrapH(combatSiteWS.Handler()))

	// 公开API
	api.GET("/items", economyHandler.GetItemDefs)
	api.GET("/skills/defs", skillHandler.GetSkillDefs)
	api.GET("/corps/:id", orgHandler.GetCorpInfo)
	api.GET("/buildings/defs", worldHandler.GetBuildingDefs)
	api.GET("/buildings/system/:system_id", worldHandler.GetSystemBuildings)
	api.GET("/ships/defs", shipHandler.GetShipDefs)

	return r
}
