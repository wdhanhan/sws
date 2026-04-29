package service

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/starfall-warsong/sws/internal/engine"
	"github.com/starfall-warsong/sws/internal/model"
	"github.com/starfall-warsong/sws/internal/repository"
)

type siteSession struct {
	SiteID     int64
	InstID     int64
	DungeonID  int64
	Wave       int
	TotalWaves int
	Engine     *engine.CombatEngine
	WaveText   string
	IsBoss     bool
	BossName   string
	NPCBounty  int64
	NPCCount   int
	FleetID    int64
	MemberIDs  []int64
	ShipIDs    map[int64]int64 // charID -> shipID
}

type CombatSiteService struct {
	db       *sqlx.DB
	invRepo  *repository.InventoryRepo
	fleetSvc *FleetService
	sessions map[int64]*siteSession // charID -> session (fleet members share pointer)
	mu       sync.RWMutex
}

func NewCombatSiteService(db *sqlx.DB, invRepo *repository.InventoryRepo) *CombatSiteService {
	return &CombatSiteService{db: db, invRepo: invRepo, sessions: make(map[int64]*siteSession)}
}

func (s *CombatSiteService) SetFleetService(fs *FleetService) { s.fleetSvc = fs }

func (s *CombatSiteService) DB() *sqlx.DB { return s.db }

type SiteInfo struct {
	ID         int64  `db:"id" json:"id"`
	SystemID   int64  `db:"system_id" json:"system_id"`
	SiteType   string `db:"site_type" json:"site_type"`
	Name       string `db:"name" json:"name"`
	Difficulty int    `db:"difficulty" json:"difficulty"`
	IsScanned  bool   `db:"is_scanned" json:"is_scanned"`
	Status     string `db:"status" json:"status"`
}

var siteTypeNames = map[string]string{
	"small_anomaly":  "小型异常",
	"medium_anomaly": "中型异常",
	"large_anomaly":  "大型异常",
	"signal":         "战斗信号",
	"expedition":     "远征入口",
}

var armRaces = map[int][]int{
	1: {1, 5, 9},      // 焚天(火象)
	2: {2, 6, 10},     // 厚土(土象)
	3: {3, 7, 11},     // 罡风(风象)
	4: {4, 8, 12},     // 渊水(水象)
	5: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, // 核心
	6: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, // 虚空
	7: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, // 外缘
}

var raceThemeNames = map[int][]string{
	1:  {"烈焰巡逻队", "战焰前哨", "冲锋者残骸"},
	2:  {"废矿掘进队", "铁心采掘场", "锻造者遗迹"},
	3:  {"幻影侦察队", "镜像干扰站", "双面陷阱"},
	4:  {"潮汐守卫站", "甲壳防线", "月蚀哨所"},
	5:  {"皇家巡逻队", "荣耀竞技场", "王座守卫"},
	6:  {"数据采集站", "精工实验室", "秩序核心"},
	7:  {"走私者据点", "黑市交易所", "贸易护卫"},
	8:  {"毒雾渗透点", "蚀刻实验场", "暗影巢穴"},
	9:  {"猎手伏击点", "游猎营地", "追踪信号站"},
	10: {"堡垒哨塔", "基石防线", "时间遗迹"},
	11: {"实验泄漏区", "范式异常点", "革新废墟"},
	12: {"深海孵化场", "共生培养皿", "有机信号源"},
}

// SpawnSitesForSystem 为指定星系生成战斗地点
func (s *CombatSiteService) SpawnSitesForSystem(ctx context.Context, systemID int64) (int, error) {
	// 获取星系信息
	type SysInfo struct {
		ArmID    int     `db:"arm_id"`
		Security float64 `db:"security_level"`
	}
	var sys SysInfo
	if err := s.db.GetContext(ctx, &sys, `SELECT arm_id, security_level FROM star_systems WHERE id=$1`, systemID); err != nil {
		return 0, err
	}

	// 清理过期/已完成的冷却地点
	s.db.ExecContext(ctx, `DELETE FROM combat_sites WHERE system_id=$1 AND status='cooldown' AND cooldown_until < NOW()`, systemID)

	// 计算目标地点数
	var targetCount int
	switch {
	case sys.Security >= 0.5:
		targetCount = 3 + rand.Intn(4) // 3-6
	case sys.Security > 0:
		targetCount = 9 + rand.Intn(7) // 9-15
	case sys.Security == 0:
		targetCount = 15 + rand.Intn(10) // 15-24
	default:
		targetCount = 24 + rand.Intn(13) // 24-36
	}

	// 当前活跃数
	var currentCount int
	s.db.GetContext(ctx, &currentCount, `SELECT COUNT(*) FROM combat_sites WHERE system_id=$1 AND status IN ('active','in_progress')`, systemID)

	needed := targetCount - currentCount
	if needed <= 0 {
		return 0, nil
	}

	// 确定种族池
	races := armRaces[sys.ArmID]
	if races == nil {
		races = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	}

	spawned := 0
	for i := 0; i < needed; i++ {
		// 随机地点类型(按安全等级权重)
		siteType, diff := s.rollSiteType(sys.Security)

		// 随机种族主题
		raceID := races[rand.Intn(len(races))]
		if rand.Float64() > 0.7 { // 30%随机种族
			raceID = rand.Intn(12) + 1
		}

		themes := raceThemeNames[raceID]
		themeName := themes[rand.Intn(len(themes))]
		siteName := fmt.Sprintf("%s %s", model.RaceNames[model.RaceID(raceID)][:6], themeName)
		if diff > 1 {
			siteName += fmt.Sprintf(" %s", []string{"", "I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X"}[diff])
		}

		// 找对应难度的dungeon_def
		var dungeonDefID int64
		err := s.db.GetContext(ctx, &dungeonDefID,
			`SELECT id FROM dungeon_defs WHERE race_theme=$1 AND difficulty=$2 LIMIT 1`, raceID, diff)
		if err != nil {
			// 没找到精确匹配，用最近的
			s.db.GetContext(ctx, &dungeonDefID,
				`SELECT id FROM dungeon_defs WHERE difficulty=$1 ORDER BY RANDOM() LIMIT 1`, diff)
		}
		if dungeonDefID == 0 {
			continue
		}

		isScanned := siteType == "small_anomaly" || siteType == "medium_anomaly" // 异常自动可见
		expiresAt := time.Now().Add(time.Duration(2+rand.Intn(4)) * time.Hour)

		s.db.ExecContext(ctx,
			`INSERT INTO combat_sites (system_id, dungeon_def_id, site_type, name, difficulty, is_scanned, status, spawned_at, expires_at)
			 VALUES ($1,$2,$3,$4,$5,$6,'active',NOW(),$7)`,
			systemID, dungeonDefID, siteType, siteName, diff, isScanned, expiresAt)
		spawned++
	}

	return spawned, nil
}

func (s *CombatSiteService) rollSiteType(security float64) (string, int) {
	roll := rand.Float64()
	if security >= 0.5 { // 高安
		if roll < 0.70 {
			return "small_anomaly", 1 + rand.Intn(2)
		}
		return "medium_anomaly", 2 + rand.Intn(2)
	}
	if security > 0 { // 低安
		switch {
		case roll < 0.35:
			return "small_anomaly", 2 + rand.Intn(2)
		case roll < 0.65:
			return "medium_anomaly", 3 + rand.Intn(2)
		case roll < 0.85:
			return "large_anomaly", 4 + rand.Intn(2)
		default:
			return "signal", 4 + rand.Intn(3)
		}
	}
	if security == 0 { // 零安
		switch {
		case roll < 0.25:
			return "small_anomaly", 3 + rand.Intn(2)
		case roll < 0.45:
			return "medium_anomaly", 4 + rand.Intn(2)
		case roll < 0.65:
			return "large_anomaly", 5 + rand.Intn(2)
		case roll < 0.85:
			return "signal", 5 + rand.Intn(3)
		default:
			return "expedition", 6 + rand.Intn(3)
		}
	}
	// 深渊
	switch {
	case roll < 0.15:
		return "medium_anomaly", 5 + rand.Intn(2)
	case roll < 0.40:
		return "large_anomaly", 6 + rand.Intn(2)
	case roll < 0.65:
		return "signal", 7 + rand.Intn(2)
	default:
		return "expedition", 8 + rand.Intn(3)
	}
}

// ListSites 列出星系中的战斗地点(自动触发刷新)
func (s *CombatSiteService) ListSites(ctx context.Context, systemID int64, showHidden bool) ([]SiteInfo, error) {
	// 先触发刷新
	s.SpawnSitesForSystem(ctx, systemID)

	q := `SELECT id, system_id, site_type, name, difficulty, is_scanned, status
		  FROM combat_sites WHERE system_id=$1 AND status='active'`
	if !showHidden {
		q += ` AND is_scanned = true`
	}
	q += ` ORDER BY difficulty, name`

	var sites []SiteInfo
	return sites, s.db.SelectContext(ctx, &sites, q, systemID)
}

// ScanSystem 扫描星系，发现隐藏地点
func (s *CombatSiteService) ScanSystem(ctx context.Context, systemID int64) ([]SiteInfo, error) {
	s.SpawnSitesForSystem(ctx, systemID)
	// 扫描发现未扫描的地点(概率成功)
	s.db.ExecContext(ctx,
		`UPDATE combat_sites SET is_scanned=true WHERE system_id=$1 AND is_scanned=false AND RANDOM() < 0.6`, systemID)

	var sites []SiteInfo
	return sites, s.db.SelectContext(ctx, &sites,
		`SELECT id, system_id, site_type, name, difficulty, is_scanned, status
		 FROM combat_sites WHERE system_id=$1 AND status='active' AND is_scanned=true
		 ORDER BY difficulty, name`, systemID)
}

type SiteFightResult struct {
	SiteID      int64              `json:"site_id"`
	SiteName    string             `json:"site_name"`
	WaveNumber  int                `json:"wave_number"`
	TotalWaves  int                `json:"total_waves"`
	WaveText    string             `json:"wave_text"`
	IsBoss      bool               `json:"is_boss"`
	BossName    string             `json:"boss_name,omitempty"`
	Completed   bool               `json:"completed"`
	Failed      bool               `json:"failed"`
	Combat      *model.CombatState `json:"combat"`
	Rewards     []string           `json:"rewards,omitempty"`
}

// EnterSite 进入地点，支持舰队组队
func (s *CombatSiteService) EnterSite(ctx context.Context, charID, siteID int64) (*SiteFightResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, ok := s.sessions[charID]; ok {
		return &SiteFightResult{
			SiteID: sess.SiteID, SiteName: "", WaveNumber: sess.Wave, TotalWaves: sess.TotalWaves,
			WaveText: sess.WaveText, IsBoss: sess.IsBoss, Combat: sess.Engine.State,
		}, nil
	}

	// Determine participants: solo or fleet
	memberIDs := []int64{charID}
	var fleetID int64
	if s.fleetSvc != nil {
		inFleet, fid, leaderID := s.fleetSvc.IsInFleet(ctx, charID)
		if inFleet {
			if leaderID != charID {
				return nil, fmt.Errorf("请等待队长进入战斗地点")
			}
			fleetID = fid
			ids, _ := s.fleetSvc.GetFleetMemberIDs(ctx, fid)
			if len(ids) > 0 {
				memberIDs = ids
			}
		}
	}

	var site struct {
		ID           int64  `db:"id"`
		DungeonDefID int64  `db:"dungeon_def_id"`
		Name         string `db:"name"`
		Status       string `db:"status"`
	}
	if err := s.db.GetContext(ctx, &site,
		`SELECT id, dungeon_def_id, name, status FROM combat_sites WHERE id=$1`, siteID); err != nil {
		return nil, fmt.Errorf("地点不存在")
	}
	if site.Status != "active" && site.Status != "in_progress" {
		return nil, fmt.Errorf("地点不可用(状态:%s)", site.Status)
	}

	var instID int64
	var currentWave int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, current_wave FROM dungeon_instances WHERE dungeon_def_id=$1 AND character_id=$2 AND status='running'`,
		site.DungeonDefID, charID).Scan(&instID, &currentWave)
	if err != nil {
		s.db.QueryRowContext(ctx,
			`INSERT INTO dungeon_instances (dungeon_def_id, character_id, current_wave, status) VALUES ($1,$2,1,'running') RETURNING id`,
			site.DungeonDefID, charID).Scan(&instID)
		currentWave = 1
		s.db.ExecContext(ctx, `UPDATE combat_sites SET status='in_progress', occupied_by=$1 WHERE id=$2`, charID, siteID)
	}

	var waveCount int
	s.db.GetContext(ctx, &waveCount, `SELECT wave_count FROM dungeon_defs WHERE id=$1`, site.DungeonDefID)

	sess := s.initWaveCombatFleet(ctx, memberIDs, site.DungeonDefID, currentWave)
	sess.SiteID = siteID
	sess.InstID = instID
	sess.DungeonID = site.DungeonDefID
	sess.Wave = currentWave
	sess.TotalWaves = waveCount
	sess.FleetID = fleetID
	sess.MemberIDs = memberIDs
	sess.ShipIDs = make(map[int64]int64)
	for _, mid := range memberIDs {
		var shipID int64
		s.db.GetContext(ctx, &shipID, `SELECT id FROM ships WHERE character_id=$1 AND is_active=true LIMIT 1`, mid)
		if shipID > 0 { sess.ShipIDs[mid] = shipID }
	}

	// Store session for all members
	for _, mid := range memberIDs {
		s.sessions[mid] = sess
	}

	sess.Engine.ProcessTick()

	return &SiteFightResult{
		SiteID: siteID, SiteName: site.Name,
		WaveNumber: currentWave, TotalWaves: waveCount,
		WaveText: sess.WaveText, IsBoss: sess.IsBoss, BossName: sess.BossName,
		Combat: sess.Engine.State,
	}, nil
}

// SiteNextTick 推进一个Tick
func (s *CombatSiteService) SiteNextTick(ctx context.Context, charID int64) (*SiteFightResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[charID]
	if !ok {
		return nil, fmt.Errorf("你不在任何战斗地点中")
	}

	if sess.Engine.State.Status != "active" {
		// 当前波已结束，处理结果
		return s.processWaveEnd(ctx, charID, sess)
	}

	sess.Engine.ProcessTick()

	// 检查这个Tick后是否结束
	if sess.Engine.State.Status != "active" {
		return s.processWaveEnd(ctx, charID, sess)
	}

	return &SiteFightResult{
		SiteID: sess.SiteID, WaveNumber: sess.Wave, TotalWaves: sess.TotalWaves,
		WaveText: sess.WaveText, IsBoss: sess.IsBoss, BossName: sess.BossName,
		Combat: sess.Engine.State,
	}, nil
}

// SiteAutoFight 自动打完当前波
func (s *CombatSiteService) SiteAutoFight(ctx context.Context, charID int64) (*SiteFightResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[charID]
	if !ok {
		return nil, fmt.Errorf("你不在任何战斗地点中")
	}

	for i := 0; i < 200 && sess.Engine.State.Status == "active"; i++ {
		sess.Engine.ProcessTick()
	}

	return s.processWaveEnd(ctx, charID, sess)
}

func (s *CombatSiteService) processWaveEnd(ctx context.Context, charID int64, sess *siteSession) (*SiteFightResult, error) {
	playerAlive := false
	kills := 0
	for _, p := range sess.Engine.State.Participants {
		if p.Type == "player" && !p.IsDestroyed { playerAlive = true }
		if p.Type == "npc" && p.IsDestroyed { kills++ }
	}

	result := &SiteFightResult{
		SiteID: sess.SiteID, WaveNumber: sess.Wave, TotalWaves: sess.TotalWaves,
		WaveText: sess.WaveText, IsBoss: sess.IsBoss, BossName: sess.BossName,
		Combat: sess.Engine.State,
	}

	members := sess.MemberIDs
	if len(members) == 0 {
		members = []int64{charID}
	}

	// Process destroyed player ships -> create wrecks
	for _, p := range sess.Engine.State.Participants {
		if p.Type != "player" || !p.IsDestroyed { continue }
		shipID := sess.ShipIDs[p.ID]
		if shipID == 0 { continue }
		if _, alreadyDone := sess.ShipIDs[p.ID]; !alreadyDone { continue }

		// Mark ship destroyed
		s.db.ExecContext(ctx, `UPDATE ships SET is_destroyed=true, is_active=false WHERE id=$1`, shipID)

		// Get system for wreck
		var systemID int64
		s.db.GetContext(ctx, &systemID, `SELECT current_system_id FROM characters WHERE id=$1`, p.ID)

		// Get ship def info
		var shipDefID int64
		var shipName string
		s.db.QueryRowContext(ctx, `SELECT ship_def_id, name FROM ships WHERE id=$1`, shipID).Scan(&shipDefID, &shipName)

		// Create wreck
		var wreckID int64
		s.db.QueryRowContext(ctx,
			`INSERT INTO wrecks (system_id, owner_name, ship_def_id, ship_name) VALUES ($1,$2,$3,$4) RETURNING id`,
			systemID, p.Name, shipDefID, shipName).Scan(&wreckID)

		// 50% chance each fitting drops into wreck
		type FitRow struct {
			ItemDefID int64 `db:"module_item_id"`
		}
		var fittings []FitRow
		s.db.SelectContext(ctx, &fittings, `SELECT module_item_id FROM ship_fittings WHERE ship_id=$1`, shipID)
		dropped := 0
		for _, f := range fittings {
			if rand.Float64() < 0.5 {
				s.db.ExecContext(ctx, `INSERT INTO wreck_items (wreck_id, item_def_id, quantity) VALUES ($1,$2,1)`, wreckID, f.ItemDefID)
				dropped++
			}
		}

		// Remove all fittings from destroyed ship
		s.db.ExecContext(ctx, `DELETE FROM ship_fittings WHERE ship_id=$1`, shipID)

		// Remove from ShipIDs so we don't process again
		delete(sess.ShipIDs, p.ID)

		sess.Engine.State.Logs = append(sess.Engine.State.Logs,
			fmt.Sprintf("  ▸ %s 的舰船变为残骸(掉落%d件装备)", p.Name, dropped))
	}

	if !playerAlive {
		result.Failed = true
		s.db.ExecContext(ctx, `UPDATE dungeon_instances SET status='failed', completed_at=NOW() WHERE id=$1`, sess.InstID)
		s.db.ExecContext(ctx, `UPDATE combat_sites SET status='active', occupied_by=0 WHERE id=$1`, sess.SiteID)
		for _, mid := range members { delete(s.sessions, mid) }
		sess.Engine.State.Logs = append(sess.Engine.State.Logs, "舰队全灭！")
		return result, nil
	}

	// Distribute bounty to all surviving members
	bounty := sess.NPCBounty * int64(kills)
	sharePerMember := bounty / int64(len(members))
	for _, mid := range members {
		s.db.ExecContext(ctx, `UPDATE characters SET balance=balance+$1 WHERE id=$2`, sharePerMember, mid)
	}
	sess.Engine.State.Logs = append(sess.Engine.State.Logs,
		fmt.Sprintf("▸ 第%d波清除! 击杀%d, 赏金+%d星币(每人%d)", sess.Wave, kills, bounty, sharePerMember))

	if sess.Wave >= sess.TotalWaves {
		result.Completed = true
		var reward int64
		s.db.GetContext(ctx, &reward, `SELECT reward_credits FROM dungeon_defs WHERE id=$1`, sess.DungeonID)
		rewardPerMember := reward / int64(len(members))
		for _, mid := range members {
			s.db.ExecContext(ctx, `UPDATE characters SET balance=balance+$1 WHERE id=$2`, rewardPerMember, mid)
		}
		s.db.ExecContext(ctx, `UPDATE dungeon_instances SET status='completed', completed_at=NOW() WHERE id=$1`, sess.InstID)
		s.db.ExecContext(ctx, `UPDATE combat_sites SET status='cooldown', completed_at=NOW(), cooldown_until=$1 WHERE id=$2`,
			time.Now().Add(20*time.Minute), sess.SiteID)

		// Loot: scale by difficulty - better items for harder content
		var difficulty int
		s.db.GetContext(ctx, &difficulty, `SELECT difficulty FROM dungeon_defs WHERE id=$1`, sess.DungeonID)

		// Choose loot pool based on difficulty
		type LootItem struct {
			ID  int64
			Qty int64
		}
		var lootPool []LootItem
		if difficulty >= 8 {
			// T8+ drops modules and high-value salvage
			modIDs := []int64{5163, 5152, 5171, 8005, 8024, 5451, 8132, 8101, 5302}
			for j := 0; j < 2+rand.Intn(4); j++ {
				lootPool = append(lootPool, LootItem{modIDs[rand.Intn(len(modIDs))], 1})
			}
			// Plus some minerals
			mineralIDs := []int64{1005, 1006, 1008}
			for j := 0; j < 1+rand.Intn(2); j++ {
				lootPool = append(lootPool, LootItem{mineralIDs[rand.Intn(len(mineralIDs))], int64(50 + rand.Intn(200))})
			}
		} else if difficulty >= 5 {
			modIDs := []int64{5112, 5102, 5121, 8004, 8029, 5315, 5306}
			for j := 0; j < 1+rand.Intn(3); j++ {
				lootPool = append(lootPool, LootItem{modIDs[rand.Intn(len(modIDs))], 1})
			}
			mineralIDs := []int64{1003, 1005, 1006}
			for j := 0; j < 1+rand.Intn(3); j++ {
				lootPool = append(lootPool, LootItem{mineralIDs[rand.Intn(len(mineralIDs))], int64(30 + rand.Intn(100))})
			}
		} else {
			oreIDs := []int64{1001, 1002, 1003, 1005, 1006, 1008}
			for j := 0; j < 1+rand.Intn(3); j++ {
				lootPool = append(lootPool, LootItem{oreIDs[rand.Intn(len(oreIDs))], int64(10 + rand.Intn(50))})
			}
		}

		// Distribute loot: each member gets the same loot
		for _, mid := range members {
			var systemID int64
			s.db.GetContext(ctx, &systemID, `SELECT current_system_id FROM characters WHERE id=$1`, mid)
			for _, item := range lootPool {
				s.invRepo.AddOrUpsertItem(ctx, "character", mid, item.ID, item.Qty, systemID)
			}
		}
		// Show loot once (not per-member)
		for _, item := range lootPool {
			def, _ := s.invRepo.GetItemDef(ctx, item.ID)
			n := "物品"
			if def != nil { n = def.Name }
			result.Rewards = append(result.Rewards, fmt.Sprintf("%s x%d", n, item.Qty))
		}

		sess.Engine.State.Logs = append(sess.Engine.State.Logs,
			"═══ 地点通关! ═══",
			fmt.Sprintf("▸ 通关奖励: 每人+%d 星币", rewardPerMember))
		for _, r := range result.Rewards {
			sess.Engine.State.Logs = append(sess.Engine.State.Logs, "▸ 战利品: "+r)
		}
		for _, mid := range members { delete(s.sessions, mid) }
	} else {
		sess.Wave++
		s.db.ExecContext(ctx, `UPDATE dungeon_instances SET current_wave=$1 WHERE id=$2`, sess.Wave, sess.InstID)
		newSess := s.initWaveCombatFleet(ctx, members, sess.DungeonID, sess.Wave)
		newSess.SiteID = sess.SiteID
		newSess.InstID = sess.InstID
		newSess.DungeonID = sess.DungeonID
		newSess.TotalWaves = sess.TotalWaves
		newSess.FleetID = sess.FleetID
		newSess.MemberIDs = members

		prevLogs := sess.Engine.State.Logs
		newSess.Engine.State.Logs = append(prevLogs, fmt.Sprintf("── 第%d波开始 ── %s", newSess.Wave, newSess.WaveText))
		for _, mid := range members { s.sessions[mid] = newSess }

		result.WaveNumber = newSess.Wave
		result.WaveText = newSess.WaveText
		result.IsBoss = newSess.IsBoss
		result.Combat = newSess.Engine.State
	}

	return result, nil
}

func (s *CombatSiteService) initWaveCombat(ctx context.Context, charID, dungeonDefID int64, wave int) *siteSession {
	type WaveData struct {
		NPCDefID   int64  `db:"npc_def_id"`
		NPCCount   int    `db:"npc_count"`
		IsBoss     bool   `db:"is_boss"`
		BossName   string `db:"boss_name"`
		BossHPOver int    `db:"boss_hp_override"`
		WaveText   string `db:"wave_text"`
	}
	var wd WaveData
	s.db.GetContext(ctx, &wd,
		`SELECT npc_def_id,npc_count,is_boss,boss_name,boss_hp_override,wave_text
		 FROM dungeon_waves WHERE dungeon_id=$1 AND wave_number=$2`, dungeonDefID, wave)

	type NPC struct {
		Name string `db:"name"`; SHP int `db:"shield_hp"`; AHP int `db:"armor_hp"`; HHP int `db:"structure_hp"`
		DPT int `db:"damage_per_tick"`; DT string `db:"damage_type"`; Spd int `db:"speed"`
		Sig int `db:"signature"`; Rng int `db:"optimal_range"`; Bty int64 `db:"bounty"`
	}
	var npc NPC
	s.db.GetContext(ctx, &npc,
		`SELECT name,shield_hp,armor_hp,structure_hp,damage_per_tick,damage_type,speed,signature,optimal_range,bounty
		 FROM npc_defs WHERE id=$1`, wd.NPCDefID)

	eng := engine.NewCombatEngine(time.Now().UnixNano())

	// 玩家
	pp := model.CombatParticipant{
		ID: charID, Name: "你", Type: "player", Team: "a",
		ShieldCurrent: 2000, ShieldMax: 2000, ArmorCurrent: 1500, ArmorMax: 1500,
		StructureCurrent: 1000, StructureMax: 1000, CapCurrent: 500, CapMax: 500, Distance: 20000,
		DamagePerTick: 0, DamageType: model.DamageKinetic, RateOfFire: 1,
		ShieldRecharge: 30, Speed: 300, Signature: 100, OptimalRange: 15000,
		ShieldResist: model.ResistProfile{Kinetic: 0.15, Thermal: 0.40, EM: 0.30, Explosive: 0.10},
		ArmorResist:  model.ResistProfile{Kinetic: 0.40, Thermal: 0.20, EM: 0.10, Explosive: 0.30},
	}
	var shipInfo struct {
		ShipID         int64   `db:"ship_id"`
		SHP            int     `db:"shield_hp"`
		AHP            int     `db:"armor_hp"`
		HHP            int     `db:"structure_hp"`
		Spd            int     `db:"max_speed"`
		Sig            int     `db:"signature"`
		SRch           int     `db:"shield_recharge"`
		Cap            int     `db:"capacitor"`
		CapRch         int     `db:"cap_recharge"`
		SResK          float64 `db:"shield_res_kinetic"`
		SResT          float64 `db:"shield_res_thermal"`
		SResE          float64 `db:"shield_res_em"`
		SResX          float64 `db:"shield_res_explosive"`
		AResK          float64 `db:"armor_res_kinetic"`
		AResT          float64 `db:"armor_res_thermal"`
		AResE          float64 `db:"armor_res_em"`
		AResX          float64 `db:"armor_res_explosive"`
	}
	if s.db.GetContext(ctx, &shipInfo,
		`SELECT sh.id as ship_id, sd.shield_hp,sd.armor_hp,sd.structure_hp,sd.max_speed,sd.signature,sd.shield_recharge,
		 sd.capacitor,sd.cap_recharge,
		 sd.shield_res_kinetic,sd.shield_res_thermal,sd.shield_res_em,sd.shield_res_explosive,
		 sd.armor_res_kinetic,sd.armor_res_thermal,sd.armor_res_em,sd.armor_res_explosive
		 FROM ship_defs sd JOIN ships sh ON sh.ship_def_id=sd.id WHERE sh.character_id=$1 AND sh.is_active=true LIMIT 1`, charID) == nil {
		pp.ShieldCurrent, pp.ShieldMax = shipInfo.SHP, shipInfo.SHP
		pp.ArmorCurrent, pp.ArmorMax = shipInfo.AHP, shipInfo.AHP
		pp.StructureCurrent, pp.StructureMax = shipInfo.HHP, shipInfo.HHP
		pp.Speed, pp.Signature = shipInfo.Spd, shipInfo.Sig
		pp.ShieldRecharge = shipInfo.SRch
		pp.CapCurrent, pp.CapMax = shipInfo.Cap, shipInfo.Cap
		pp.CapRecharge = shipInfo.CapRch
		pp.ShieldResist = model.ResistProfile{Kinetic: shipInfo.SResK, Thermal: shipInfo.SResT, EM: shipInfo.SResE, Explosive: shipInfo.SResX}
		pp.ArmorResist = model.ResistProfile{Kinetic: shipInfo.AResK, Thermal: shipInfo.AResT, EM: shipInfo.AResE, Explosive: shipInfo.AResX}

		// High slot weapons
		type WeaponRow struct {
			DPT      int     `db:"damage_per_tick"`
			DmgType  string  `db:"damage_type"`
			ROF      int     `db:"rate_of_fire"`
			OptRng   int     `db:"optimal_range"`
			Falloff  int     `db:"falloff_range"`
			Tracking float64 `db:"tracking_speed"`
			CapCost  int     `db:"cap_cost"`
		}
		var weapons []WeaponRow
		s.db.SelectContext(ctx, &weapons,
			`SELECT COALESCE(i.damage_per_tick,0) as damage_per_tick,
			  COALESCE(i.damage_type,'kinetic') as damage_type,
			  COALESCE(i.rate_of_fire,1) as rate_of_fire,
			  COALESCE(i.optimal_range,15000) as optimal_range,
			  COALESCE(i.falloff_range,0) as falloff_range,
			  COALESCE(i.tracking_speed,0) as tracking_speed,
			  COALESCE(i.cap_cost,0) as cap_cost
			 FROM ship_fittings sf JOIN item_defs i ON i.id=sf.module_item_id
			 WHERE sf.ship_id=$1 AND sf.slot_type='high' AND COALESCE(i.damage_per_tick,0)>0`, shipInfo.ShipID)

		if len(weapons) > 0 {
			totalDPS, totalCapCost, bestRange, bestFalloff := 0, 0, 0, 0
			var bestTracking float64
			dmgCounts := map[string]int{}
			for _, w := range weapons {
				rof := w.ROF
				if rof <= 0 { rof = 1 }
				totalDPS += w.DPT / rof
				totalCapCost += w.CapCost / rof
				if w.OptRng > bestRange { bestRange = w.OptRng }
				if w.Falloff > bestFalloff { bestFalloff = w.Falloff }
				if w.Tracking > 0 && (bestTracking == 0 || w.Tracking < bestTracking) {
					bestTracking = w.Tracking
				}
				dmgCounts[w.DmgType]++
			}
			pp.DamagePerTick = totalDPS
			pp.OptimalRange = bestRange
			pp.FalloffRange = bestFalloff
			pp.TrackingSpeed = bestTracking
			pp.CapCost = totalCapCost
			bestType, bestCount := "kinetic", 0
			for t, c := range dmgCounts {
				if c > bestCount { bestType = t; bestCount = c }
			}
			pp.DamageType = model.DamageType(bestType)
			pp.WeaponName = fmt.Sprintf("%d门武器", len(weapons))
		}

		// Mid/low slot module effects
		type ModEffect struct {
			BonusType  string  `db:"bonus_type"`
			BonusValue float64 `db:"bonus_value"`
		}
		var effects []ModEffect
		s.db.SelectContext(ctx, &effects,
			`SELECT COALESCE(i.bonus_type,'') as bonus_type, COALESCE(i.bonus_value,0) as bonus_value
			 FROM ship_fittings sf JOIN item_defs i ON i.id=sf.module_item_id
			 WHERE sf.ship_id=$1 AND sf.slot_type IN ('mid','low') AND COALESCE(i.bonus_type,'') != ''`, shipInfo.ShipID)

		for _, e := range effects {
			switch e.BonusType {
			case "shield_boost":
				pp.ShieldRecharge += int(e.BonusValue)
			case "shield_hp_bonus":
				bonus := int(float64(pp.ShieldMax) * e.BonusValue)
				pp.ShieldMax += bonus
				pp.ShieldCurrent += bonus
			case "armor_repair":
				pp.ArmorRepair += int(e.BonusValue)
			case "armor_hp":
				pp.ArmorMax += int(e.BonusValue)
				pp.ArmorCurrent += int(e.BonusValue)
			case "armor_kinetic_resist":
				pp.ArmorResist.Kinetic += e.BonusValue
			case "armor_thermal_resist":
				pp.ArmorResist.Thermal += e.BonusValue
			case "armor_em_resist":
				pp.ArmorResist.EM += e.BonusValue
			case "armor_explosive_resist":
				pp.ArmorResist.Explosive += e.BonusValue
			case "armor_omni_resist":
				pp.ArmorResist.Kinetic += e.BonusValue
				pp.ArmorResist.Thermal += e.BonusValue
				pp.ArmorResist.EM += e.BonusValue
				pp.ArmorResist.Explosive += e.BonusValue
			case "shield_kinetic_resist":
				pp.ShieldResist.Kinetic += e.BonusValue
			case "shield_thermal_resist":
				pp.ShieldResist.Thermal += e.BonusValue
			case "shield_em_resist":
				pp.ShieldResist.EM += e.BonusValue
			case "shield_explosive_resist":
				pp.ShieldResist.Explosive += e.BonusValue
			case "shield_omni_resist":
				pp.ShieldResist.Kinetic += e.BonusValue
				pp.ShieldResist.Thermal += e.BonusValue
				pp.ShieldResist.EM += e.BonusValue
				pp.ShieldResist.Explosive += e.BonusValue
			case "all_resist":
				pp.ShieldResist.Kinetic += e.BonusValue
				pp.ShieldResist.Thermal += e.BonusValue
				pp.ShieldResist.EM += e.BonusValue
				pp.ShieldResist.Explosive += e.BonusValue
				pp.ArmorResist.Kinetic += e.BonusValue
				pp.ArmorResist.Thermal += e.BonusValue
				pp.ArmorResist.EM += e.BonusValue
				pp.ArmorResist.Explosive += e.BonusValue
			case "speed_bonus":
				pp.Speed += int(float64(pp.Speed) * e.BonusValue)
			case "cap_boost":
				pp.CapMax += int(e.BonusValue)
				pp.CapCurrent += int(e.BonusValue)
			case "cap_recharge_bonus":
				pp.CapRecharge += int(float64(pp.CapRecharge) * e.BonusValue)
			case "tracking_bonus":
				pp.TrackingSpeed *= (1.0 + e.BonusValue)
			case "signature_reduction":
				pp.Signature -= int(float64(pp.Signature) * e.BonusValue)
				if pp.Signature < 10 { pp.Signature = 10 }
			}
		}

		// Clamp resists to 0.85 max
		clampResist := func(r *model.ResistProfile) {
			if r.Kinetic > 0.85 { r.Kinetic = 0.85 }
			if r.Thermal > 0.85 { r.Thermal = 0.85 }
			if r.EM > 0.85 { r.EM = 0.85 }
			if r.Explosive > 0.85 { r.Explosive = 0.85 }
		}
		clampResist(&pp.ShieldResist)
		clampResist(&pp.ArmorResist)
	}
	eng.AddParticipant(pp)

	for i := 0; i < wd.NPCCount; i++ {
		shp, ahp, stp, dpt := npc.SHP, npc.AHP, npc.HHP, npc.DPT
		if wd.IsBoss && wd.BossHPOver > 0 && (shp+ahp+stp) > 0 {
			sc := float64(wd.BossHPOver) / float64(shp+ahp+stp)
			shp, ahp, stp = int(float64(shp)*sc), int(float64(ahp)*sc), int(float64(stp)*sc)
			dpt = int(float64(dpt) * sc * 0.5)
		}
		en := npc.Name
		if wd.IsBoss && wd.BossName != "" { en = "★ " + wd.BossName } else if wd.NPCCount > 1 { en = fmt.Sprintf("%s #%d", npc.Name, i+1) }
		eng.AddParticipant(model.CombatParticipant{
			ID: 10000 + int64(i), Name: en, Type: "npc", Team: "b",
			ShieldCurrent: shp, ShieldMax: shp, ArmorCurrent: ahp, ArmorMax: ahp,
			StructureCurrent: stp, StructureMax: stp, CapCurrent: 500, Distance: 20000,
			DamagePerTick: dpt, DamageType: model.DamageType(npc.DT), RateOfFire: 1,
			Speed: npc.Spd, Signature: npc.Sig, OptimalRange: npc.Rng,
			ShieldResist: model.ResistProfile{Kinetic: 0.15, Thermal: 0.30, EM: 0.20, Explosive: 0.10},
			ArmorResist:  model.ResistProfile{Kinetic: 0.30, Thermal: 0.20, EM: 0.15, Explosive: 0.20},
		})
	}

	return &siteSession{
		Wave: wave, Engine: eng, WaveText: wd.WaveText,
		IsBoss: wd.IsBoss, BossName: wd.BossName,
		NPCBounty: npc.Bty, NPCCount: wd.NPCCount,
	}
}

// initWaveCombatFleet creates a wave combat with multiple player ships
func (s *CombatSiteService) initWaveCombatFleet(ctx context.Context, charIDs []int64, dungeonDefID int64, wave int) *siteSession {
	if len(charIDs) == 1 {
		return s.initWaveCombat(ctx, charIDs[0], dungeonDefID, wave)
	}

	// Use the single-player version to get wave data and NPC setup
	baseSess := s.initWaveCombat(ctx, charIDs[0], dungeonDefID, wave)

	// Add remaining fleet members as additional team "a" participants
	for idx := 1; idx < len(charIDs); idx++ {
		cid := charIDs[idx]
		pp := s.buildPlayerParticipant(ctx, cid)
		pp.ID = cid
		baseSess.Engine.AddParticipant(pp)
	}

	return baseSess
}

// buildPlayerParticipant creates a CombatParticipant from a character's active ship
func (s *CombatSiteService) buildPlayerParticipant(ctx context.Context, charID int64) model.CombatParticipant {
	var charName string
	s.db.GetContext(ctx, &charName, `SELECT name FROM characters WHERE id=$1`, charID)

	pp := model.CombatParticipant{
		ID: charID, Name: charName, Type: "player", Team: "a",
		ShieldCurrent: 2000, ShieldMax: 2000, ArmorCurrent: 1500, ArmorMax: 1500,
		StructureCurrent: 1000, StructureMax: 1000, CapCurrent: 500, CapMax: 500, Distance: 20000,
		DamagePerTick: 0, DamageType: model.DamageKinetic, RateOfFire: 1,
		ShieldRecharge: 30, Speed: 300, Signature: 100, OptimalRange: 15000,
		ShieldResist: model.ResistProfile{Kinetic: 0.15, Thermal: 0.40, EM: 0.30, Explosive: 0.10},
		ArmorResist:  model.ResistProfile{Kinetic: 0.40, Thermal: 0.20, EM: 0.10, Explosive: 0.30},
	}

	var shipInfo struct {
		ShipID int64   `db:"ship_id"`
		SHP    int     `db:"shield_hp"`
		AHP    int     `db:"armor_hp"`
		HHP    int     `db:"structure_hp"`
		Spd    int     `db:"max_speed"`
		Sig    int     `db:"signature"`
		SRch   int     `db:"shield_recharge"`
		Cap    int     `db:"capacitor"`
		CapRch int     `db:"cap_recharge"`
		SResK  float64 `db:"shield_res_kinetic"`
		SResT  float64 `db:"shield_res_thermal"`
		SResE  float64 `db:"shield_res_em"`
		SResX  float64 `db:"shield_res_explosive"`
		AResK  float64 `db:"armor_res_kinetic"`
		AResT  float64 `db:"armor_res_thermal"`
		AResE  float64 `db:"armor_res_em"`
		AResX  float64 `db:"armor_res_explosive"`
	}
	if s.db.GetContext(ctx, &shipInfo,
		`SELECT sh.id as ship_id, sd.shield_hp,sd.armor_hp,sd.structure_hp,sd.max_speed,sd.signature,sd.shield_recharge,
		 sd.capacitor,sd.cap_recharge,
		 sd.shield_res_kinetic,sd.shield_res_thermal,sd.shield_res_em,sd.shield_res_explosive,
		 sd.armor_res_kinetic,sd.armor_res_thermal,sd.armor_res_em,sd.armor_res_explosive
		 FROM ship_defs sd JOIN ships sh ON sh.ship_def_id=sd.id WHERE sh.character_id=$1 AND sh.is_active=true LIMIT 1`, charID) == nil {
		pp.ShieldCurrent, pp.ShieldMax = shipInfo.SHP, shipInfo.SHP
		pp.ArmorCurrent, pp.ArmorMax = shipInfo.AHP, shipInfo.AHP
		pp.StructureCurrent, pp.StructureMax = shipInfo.HHP, shipInfo.HHP
		pp.Speed, pp.Signature = shipInfo.Spd, shipInfo.Sig
		pp.ShieldRecharge = shipInfo.SRch
		pp.CapCurrent, pp.CapMax = shipInfo.Cap, shipInfo.Cap
		pp.CapRecharge = shipInfo.CapRch
		pp.ShieldResist = model.ResistProfile{Kinetic: shipInfo.SResK, Thermal: shipInfo.SResT, EM: shipInfo.SResE, Explosive: shipInfo.SResX}
		pp.ArmorResist = model.ResistProfile{Kinetic: shipInfo.AResK, Thermal: shipInfo.AResT, EM: shipInfo.AResE, Explosive: shipInfo.AResX}

		// Weapons
		type WeaponRow struct {
			DPT      int     `db:"damage_per_tick"`
			DmgType  string  `db:"damage_type"`
			ROF      int     `db:"rate_of_fire"`
			OptRng   int     `db:"optimal_range"`
			Falloff  int     `db:"falloff_range"`
			Tracking float64 `db:"tracking_speed"`
			CapCost  int     `db:"cap_cost"`
		}
		var weapons []WeaponRow
		s.db.SelectContext(ctx, &weapons,
			`SELECT COALESCE(i.damage_per_tick,0) as damage_per_tick,
			  COALESCE(i.damage_type,'kinetic') as damage_type,
			  COALESCE(i.rate_of_fire,1) as rate_of_fire,
			  COALESCE(i.optimal_range,15000) as optimal_range,
			  COALESCE(i.falloff_range,0) as falloff_range,
			  COALESCE(i.tracking_speed,0) as tracking_speed,
			  COALESCE(i.cap_cost,0) as cap_cost
			 FROM ship_fittings sf JOIN item_defs i ON i.id=sf.module_item_id
			 WHERE sf.ship_id=$1 AND sf.slot_type='high' AND COALESCE(i.damage_per_tick,0)>0`, shipInfo.ShipID)

		if len(weapons) > 0 {
			totalDPS, totalCapCost, bestRange, bestFalloff := 0, 0, 0, 0
			var bestTracking float64
			dmgCounts := map[string]int{}
			for _, w := range weapons {
				rof := w.ROF; if rof <= 0 { rof = 1 }
				totalDPS += w.DPT / rof
				totalCapCost += w.CapCost / rof
				if w.OptRng > bestRange { bestRange = w.OptRng }
				if w.Falloff > bestFalloff { bestFalloff = w.Falloff }
				if w.Tracking > 0 && (bestTracking == 0 || w.Tracking < bestTracking) { bestTracking = w.Tracking }
				dmgCounts[w.DmgType]++
			}
			pp.DamagePerTick = totalDPS
			pp.OptimalRange = bestRange
			pp.FalloffRange = bestFalloff
			pp.TrackingSpeed = bestTracking
			pp.CapCost = totalCapCost
			bestType, bestCount := "kinetic", 0
			for t, c := range dmgCounts { if c > bestCount { bestType = t; bestCount = c } }
			pp.DamageType = model.DamageType(bestType)
			pp.WeaponName = fmt.Sprintf("%d门武器", len(weapons))
		}

		// Module effects
		type ModEffect struct {
			BonusType  string  `db:"bonus_type"`
			BonusValue float64 `db:"bonus_value"`
		}
		var effects []ModEffect
		s.db.SelectContext(ctx, &effects,
			`SELECT COALESCE(i.bonus_type,'') as bonus_type, COALESCE(i.bonus_value,0) as bonus_value
			 FROM ship_fittings sf JOIN item_defs i ON i.id=sf.module_item_id
			 WHERE sf.ship_id=$1 AND sf.slot_type IN ('mid','low') AND COALESCE(i.bonus_type,'') != ''`, shipInfo.ShipID)

		for _, e := range effects {
			switch e.BonusType {
			case "shield_boost": pp.ShieldRecharge += int(e.BonusValue)
			case "shield_hp_bonus":
				bonus := int(float64(pp.ShieldMax) * e.BonusValue)
				pp.ShieldMax += bonus; pp.ShieldCurrent += bonus
			case "armor_repair": pp.ArmorRepair += int(e.BonusValue)
			case "armor_hp": pp.ArmorMax += int(e.BonusValue); pp.ArmorCurrent += int(e.BonusValue)
			case "armor_kinetic_resist": pp.ArmorResist.Kinetic += e.BonusValue
			case "armor_thermal_resist": pp.ArmorResist.Thermal += e.BonusValue
			case "armor_em_resist": pp.ArmorResist.EM += e.BonusValue
			case "armor_explosive_resist": pp.ArmorResist.Explosive += e.BonusValue
			case "armor_omni_resist":
				pp.ArmorResist.Kinetic += e.BonusValue; pp.ArmorResist.Thermal += e.BonusValue
				pp.ArmorResist.EM += e.BonusValue; pp.ArmorResist.Explosive += e.BonusValue
			case "shield_kinetic_resist": pp.ShieldResist.Kinetic += e.BonusValue
			case "shield_thermal_resist": pp.ShieldResist.Thermal += e.BonusValue
			case "shield_em_resist": pp.ShieldResist.EM += e.BonusValue
			case "shield_explosive_resist": pp.ShieldResist.Explosive += e.BonusValue
			case "shield_omni_resist":
				pp.ShieldResist.Kinetic += e.BonusValue; pp.ShieldResist.Thermal += e.BonusValue
				pp.ShieldResist.EM += e.BonusValue; pp.ShieldResist.Explosive += e.BonusValue
			case "all_resist":
				pp.ShieldResist.Kinetic += e.BonusValue; pp.ShieldResist.Thermal += e.BonusValue
				pp.ShieldResist.EM += e.BonusValue; pp.ShieldResist.Explosive += e.BonusValue
				pp.ArmorResist.Kinetic += e.BonusValue; pp.ArmorResist.Thermal += e.BonusValue
				pp.ArmorResist.EM += e.BonusValue; pp.ArmorResist.Explosive += e.BonusValue
			case "speed_bonus": pp.Speed += int(float64(pp.Speed) * e.BonusValue)
			case "cap_boost": pp.CapMax += int(e.BonusValue); pp.CapCurrent += int(e.BonusValue)
			case "cap_recharge_bonus": pp.CapRecharge += int(float64(pp.CapRecharge) * e.BonusValue)
			case "tracking_bonus": pp.TrackingSpeed *= (1.0 + e.BonusValue)
			case "signature_reduction":
				pp.Signature -= int(float64(pp.Signature) * e.BonusValue)
				if pp.Signature < 10 { pp.Signature = 10 }
			}
		}
		clamp := func(r *model.ResistProfile) {
			if r.Kinetic > 0.85 { r.Kinetic = 0.85 }; if r.Thermal > 0.85 { r.Thermal = 0.85 }
			if r.EM > 0.85 { r.EM = 0.85 }; if r.Explosive > 0.85 { r.Explosive = 0.85 }
		}
		clamp(&pp.ShieldResist); clamp(&pp.ArmorResist)
	}
	return pp
}

// LeaveSite 离开当前地点（支持舰队）
func (s *CombatSiteService) LeaveSite(ctx context.Context, charID int64) error {
	s.mu.Lock()
	sess, ok := s.sessions[charID]
	if ok && len(sess.MemberIDs) > 1 {
		// Fleet: clean up all members
		for _, mid := range sess.MemberIDs {
			delete(s.sessions, mid)
		}
	} else {
		delete(s.sessions, charID)
	}
	s.mu.Unlock()
	s.db.ExecContext(ctx, `UPDATE dungeon_instances SET status='abandoned', completed_at=NOW() WHERE character_id=$1 AND status='running'`, charID)
	s.db.ExecContext(ctx, `UPDATE combat_sites SET status='active', occupied_by=0 WHERE occupied_by=$1`, charID)
	return nil
}
