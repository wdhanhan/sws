package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/starfall-warsong/sws/internal/engine"
	"github.com/starfall-warsong/sws/internal/model"
	"github.com/starfall-warsong/sws/internal/repository"
)

var (
	ErrDungeonNotFound   = errors.New("远征副本不存在")
	ErrAlreadyInDungeon  = errors.New("角色已在远征中")
	ErrNotInDungeon      = errors.New("角色不在远征中")
	ErrDungeonCompleted  = errors.New("远征已完成")
)

type DungeonService struct {
	db      *sqlx.DB
	invRepo *repository.InventoryRepo
}

func NewDungeonService(db *sqlx.DB, invRepo *repository.InventoryRepo) *DungeonService {
	return &DungeonService{db: db, invRepo: invRepo}
}

type DungeonBrief struct {
	ID          int64  `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
	RaceTheme   int    `db:"race_theme" json:"race_theme"`
	Difficulty  int    `db:"difficulty" json:"difficulty"`
	WaveCount   int    `db:"wave_count" json:"wave_count"`
	Reward      int64  `db:"reward_credits" json:"reward_credits"`
}

type WaveInfo struct {
	WaveNumber int    `json:"wave_number"`
	TotalWaves int    `json:"total_waves"`
	WaveText   string `json:"wave_text"`
	EnemyCount int    `json:"enemy_count"`
	IsBoss     bool   `json:"is_boss"`
	BossName   string `json:"boss_name,omitempty"`
	Combat     *model.CombatState `json:"combat,omitempty"`
}

type DungeonResult struct {
	Status       string   `json:"status"`
	TotalKills   int      `json:"total_kills"`
	RewardCredits int64   `json:"reward_credits"`
	Loot         []string `json:"loot"`
	Messages     []string `json:"messages"`
}

func (s *DungeonService) ListDungeons(ctx context.Context, raceTheme, difficulty int) ([]DungeonBrief, error) {
	q := `SELECT id,name,description,race_theme,difficulty,wave_count,reward_credits FROM dungeon_defs WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if raceTheme > 0 {
		q += fmt.Sprintf(` AND race_theme=$%d`, idx)
		args = append(args, raceTheme)
		idx++
	}
	if difficulty > 0 {
		q += fmt.Sprintf(` AND difficulty=$%d`, idx)
		args = append(args, difficulty)
		idx++
	}
	q += ` ORDER BY race_theme, difficulty LIMIT 50`
	var dungeons []DungeonBrief
	return dungeons, s.db.SelectContext(ctx, &dungeons, q, args...)
}

func (s *DungeonService) Enter(ctx context.Context, charID, dungeonDefID int64) (*WaveInfo, error) {
	var active int
	s.db.GetContext(ctx, &active, `SELECT COUNT(*) FROM dungeon_instances WHERE character_id=$1 AND status='running'`, charID)
	if active > 0 {
		return nil, ErrAlreadyInDungeon
	}

	var def DungeonBrief
	if err := s.db.GetContext(ctx, &def,
		`SELECT id,name,description,race_theme,difficulty,wave_count,reward_credits FROM dungeon_defs WHERE id=$1`, dungeonDefID); err != nil {
		return nil, ErrDungeonNotFound
	}

	var instID int64
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO dungeon_instances (dungeon_def_id, character_id, current_wave, status)
		 VALUES ($1,$2,1,'running') RETURNING id`,
		dungeonDefID, charID).Scan(&instID)
	if err != nil {
		return nil, err
	}

	return s.getWaveInfo(ctx, dungeonDefID, 1, def.WaveCount)
}

func (s *DungeonService) GetStatus(ctx context.Context, charID int64) (*WaveInfo, error) {
	var inst struct {
		ID          int64 `db:"id"`
		DungeonDefID int64 `db:"dungeon_def_id"`
		CurrentWave int   `db:"current_wave"`
		Status      string `db:"status"`
	}
	err := s.db.GetContext(ctx, &inst,
		`SELECT id,dungeon_def_id,current_wave,status FROM dungeon_instances WHERE character_id=$1 AND status='running' LIMIT 1`, charID)
	if err != nil {
		return nil, ErrNotInDungeon
	}

	var waveCount int
	s.db.GetContext(ctx, &waveCount, `SELECT wave_count FROM dungeon_defs WHERE id=$1`, inst.DungeonDefID)

	return s.getWaveInfo(ctx, inst.DungeonDefID, inst.CurrentWave, waveCount)
}

func (s *DungeonService) FightWave(ctx context.Context, charID int64) (*WaveInfo, error) {
	var inst struct {
		ID           int64  `db:"id"`
		DungeonDefID int64  `db:"dungeon_def_id"`
		CurrentWave  int    `db:"current_wave"`
		TotalKills   int    `db:"total_kills"`
	}
	err := s.db.GetContext(ctx, &inst,
		`SELECT id,dungeon_def_id,current_wave,total_kills FROM dungeon_instances WHERE character_id=$1 AND status='running' LIMIT 1`, charID)
	if err != nil {
		return nil, ErrNotInDungeon
	}

	// Get wave data
	type WaveData struct {
		NPCDefID    int64  `db:"npc_def_id"`
		NPCCount    int    `db:"npc_count"`
		IsBoss      bool   `db:"is_boss"`
		BossName    string `db:"boss_name"`
		BossHPOver  int    `db:"boss_hp_override"`
		WaveText    string `db:"wave_text"`
	}
	var wave WaveData
	s.db.GetContext(ctx, &wave,
		`SELECT npc_def_id,npc_count,is_boss,boss_name,boss_hp_override,wave_text
		 FROM dungeon_waves WHERE dungeon_id=$1 AND wave_number=$2`, inst.DungeonDefID, inst.CurrentWave)

	// Get NPC base stats
	type NPCStat struct {
		Name      string `db:"name"`
		ShieldHP  int    `db:"shield_hp"`
		ArmorHP   int    `db:"armor_hp"`
		StructHP  int    `db:"structure_hp"`
		DPT       int    `db:"damage_per_tick"`
		DmgType   string `db:"damage_type"`
		Speed     int    `db:"speed"`
		Signature int    `db:"signature"`
		OptRange  int    `db:"optimal_range"`
		Bounty    int64  `db:"bounty"`
	}
	var npcStat NPCStat
	s.db.GetContext(ctx, &npcStat,
		`SELECT name,shield_hp,armor_hp,structure_hp,damage_per_tick,damage_type,speed,signature,optimal_range,bounty
		 FROM npc_defs WHERE id=$1`, wave.NPCDefID)

	// Build combat with all enemies
	combatID := time.Now().UnixNano()
	eng := engine.NewCombatEngine(combatID)

	// Player
	playerP := model.CombatParticipant{
		ID: charID, Name: "你", Type: "player", Team: "a",
		ShieldCurrent: 2000, ShieldMax: 2000,
		ArmorCurrent: 1500, ArmorMax: 1500,
		StructureCurrent: 1000, StructureMax: 1000,
		DamagePerTick: 80, DamageType: model.DamageKinetic, RateOfFire: 1,
		ShieldRecharge: 30, Speed: 300, Signature: 100, OptimalRange: 15000, Distance: 20000,
		ShieldResist: model.ResistProfile{Kinetic: 0.15, Thermal: 0.40, EM: 0.30, Explosive: 0.10},
		ArmorResist:  model.ResistProfile{Kinetic: 0.40, Thermal: 0.20, EM: 0.10, Explosive: 0.30},
	}
	// Load actual ship stats
	var shipHP struct {
		ShieldHP int `db:"shield_hp"`
		ArmorHP  int `db:"armor_hp"`
		StructHP int `db:"structure_hp"`
		Speed    int `db:"max_speed"`
		Sig      int `db:"signature"`
	}
	err = s.db.GetContext(ctx, &shipHP,
		`SELECT sd.shield_hp,sd.armor_hp,sd.structure_hp,sd.max_speed,sd.signature
		 FROM ship_defs sd JOIN ships sh ON sh.ship_def_id=sd.id
		 WHERE sh.character_id=$1 AND sh.is_active=true LIMIT 1`, charID)
	if err == nil {
		playerP.ShieldCurrent = shipHP.ShieldHP
		playerP.ShieldMax = shipHP.ShieldHP
		playerP.ArmorCurrent = shipHP.ArmorHP
		playerP.ArmorMax = shipHP.ArmorHP
		playerP.StructureCurrent = shipHP.StructHP
		playerP.StructureMax = shipHP.StructHP
		playerP.Speed = shipHP.Speed
		playerP.Signature = shipHP.Sig
	}
	eng.AddParticipant(playerP)

	// Enemies
	for i := 0; i < wave.NPCCount; i++ {
		hp := npcStat.ShieldHP + npcStat.ArmorHP + npcStat.StructHP
		if wave.IsBoss && wave.BossHPOver > 0 {
			scale := float64(wave.BossHPOver) / float64(hp)
			npcStat.ShieldHP = int(float64(npcStat.ShieldHP) * scale)
			npcStat.ArmorHP = int(float64(npcStat.ArmorHP) * scale)
			npcStat.StructHP = int(float64(npcStat.StructHP) * scale)
			npcStat.DPT = int(float64(npcStat.DPT) * scale * 0.5)
		}
		eName := npcStat.Name
		if wave.IsBoss && wave.BossName != "" {
			eName = wave.BossName
		} else if wave.NPCCount > 1 {
			eName = fmt.Sprintf("%s #%d", npcStat.Name, i+1)
		}
		ep := model.CombatParticipant{
			ID: 10000 + int64(i), Name: eName, Type: "npc", Team: "b",
			ShieldCurrent: npcStat.ShieldHP, ShieldMax: npcStat.ShieldHP,
			ArmorCurrent: npcStat.ArmorHP, ArmorMax: npcStat.ArmorHP,
			StructureCurrent: npcStat.StructHP, StructureMax: npcStat.StructHP,
			DamagePerTick: npcStat.DPT, DamageType: model.DamageType(npcStat.DmgType),
			RateOfFire: 1, Speed: npcStat.Speed, Signature: npcStat.Signature,
			OptimalRange: npcStat.OptRange, Distance: 20000,
			ShieldResist: model.ResistProfile{Kinetic: 0.15, Thermal: 0.30, EM: 0.20, Explosive: 0.10},
			ArmorResist:  model.ResistProfile{Kinetic: 0.30, Thermal: 0.20, EM: 0.15, Explosive: 0.20},
		}
		eng.AddParticipant(ep)
	}

	// Auto-fight the wave
	var allLogs []string
	for tick := 0; tick < 200; tick++ {
		logs := eng.ProcessTick()
		allLogs = append(allLogs, logs...)
		if eng.State.Status != "active" {
			break
		}
	}

	// Check result
	playerAlive := false
	kills := 0
	for _, p := range eng.State.Participants {
		if p.Type == "player" && !p.IsDestroyed {
			playerAlive = true
		}
		if p.Type == "npc" && p.IsDestroyed {
			kills++
		}
	}

	var waveCount int
	s.db.GetContext(ctx, &waveCount, `SELECT wave_count FROM dungeon_defs WHERE id=$1`, inst.DungeonDefID)

	if !playerAlive {
		s.db.ExecContext(ctx, `UPDATE dungeon_instances SET status='failed', completed_at=NOW() WHERE id=$1`, inst.ID)
		return &WaveInfo{
			WaveNumber: inst.CurrentWave, TotalWaves: waveCount,
			WaveText: "你的舰船被击毁了！远征失败。", IsBoss: wave.IsBoss,
			Combat: eng.State,
		}, nil
	}

	s.db.ExecContext(ctx, `UPDATE dungeon_instances SET total_kills=total_kills+$1 WHERE id=$2`, kills, inst.ID)

	// Award bounty per kill
	totalBounty := npcStat.Bounty * int64(kills)
	s.db.ExecContext(ctx, `UPDATE characters SET balance=balance+$1 WHERE id=$2`, totalBounty, charID)

	// Check if dungeon completed
	if inst.CurrentWave >= waveCount {
		// Dungeon completed!
		var reward int64
		s.db.GetContext(ctx, &reward, `SELECT reward_credits FROM dungeon_defs WHERE id=$1`, inst.DungeonDefID)
		s.db.ExecContext(ctx, `UPDATE characters SET balance=balance+$1 WHERE id=$2`, reward, charID)
		s.db.ExecContext(ctx, `UPDATE dungeon_instances SET status='completed', completed_at=NOW() WHERE id=$1`, inst.ID)

		// Loot roll
		var lootMsgs []string
		var systemID int64
		s.db.GetContext(ctx, &systemID, `SELECT current_system_id FROM characters WHERE id=$1`, charID)

		oreIDs := []int64{1001, 1002, 1003, 1005, 1006, 1008}
		for i := 0; i < 2+rand.Intn(3); i++ {
			itemID := oreIDs[rand.Intn(len(oreIDs))]
			qty := int64(10 + rand.Intn(50))
			s.invRepo.AddOrUpsertItem(ctx, "character", charID, itemID, qty, systemID)
			def, _ := s.invRepo.GetItemDef(ctx, itemID)
			name := "物品"
			if def != nil {
				name = def.Name
			}
			lootMsgs = append(lootMsgs, fmt.Sprintf("%s x%d", name, qty))
		}

		allLogs = append(allLogs,
			fmt.Sprintf("═══ 远征通关! ═══"),
			fmt.Sprintf("▸ 通关奖励: +%d 星币", reward),
			fmt.Sprintf("▸ 击杀赏金: +%d 星币", totalBounty),
		)
		for _, l := range lootMsgs {
			allLogs = append(allLogs, "▸ 战利品: "+l)
		}

		eng.State.Logs = allLogs
		eng.State.Status = "completed"

		return &WaveInfo{
			WaveNumber: inst.CurrentWave, TotalWaves: waveCount,
			WaveText: "远征通关！所有敌人已被消灭。", IsBoss: wave.IsBoss, BossName: wave.BossName,
			Combat: eng.State,
		}, nil
	}

	// Advance to next wave
	s.db.ExecContext(ctx, `UPDATE dungeon_instances SET current_wave=current_wave+1 WHERE id=$1`, inst.ID)

	allLogs = append(allLogs, fmt.Sprintf("▸ 第%d波清除! 击杀%d个敌人, 赏金+%d星币", inst.CurrentWave, kills, totalBounty))

	// 30% chance trigger encounter between waves
	if rand.Float64() < 0.3 {
		allLogs = append(allLogs, "★ 在残骸中发现了一些有趣的东西...")
	}

	eng.State.Logs = allLogs

	return &WaveInfo{
		WaveNumber: inst.CurrentWave, TotalWaves: waveCount,
		WaveText: fmt.Sprintf("第%d波清除完毕。准备迎接第%d波...", inst.CurrentWave, inst.CurrentWave+1),
		EnemyCount: kills, IsBoss: wave.IsBoss,
		Combat: eng.State,
	}, nil
}

func (s *DungeonService) Leave(ctx context.Context, charID int64) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE dungeon_instances SET status='abandoned', completed_at=NOW() WHERE character_id=$1 AND status='running'`, charID)
	return err
}

func (s *DungeonService) getWaveInfo(ctx context.Context, dungeonDefID int64, wave, totalWaves int) (*WaveInfo, error) {
	type W struct {
		NPCCount int    `db:"npc_count"`
		IsBoss   bool   `db:"is_boss"`
		BossName string `db:"boss_name"`
		WaveText string `db:"wave_text"`
	}
	var w W
	s.db.GetContext(ctx, &w,
		`SELECT npc_count,is_boss,boss_name,wave_text FROM dungeon_waves WHERE dungeon_id=$1 AND wave_number=$2`,
		dungeonDefID, wave)

	return &WaveInfo{
		WaveNumber: wave, TotalWaves: totalWaves,
		WaveText: w.WaveText, EnemyCount: w.NPCCount,
		IsBoss: w.IsBoss, BossName: w.BossName,
	}, nil
}

func init() { _ = json.Marshal }
