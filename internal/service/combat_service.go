package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/starfall-warsong/sws/internal/engine"
	"github.com/starfall-warsong/sws/internal/model"
	"github.com/starfall-warsong/sws/internal/repository"
)

var (
	ErrNoCombat         = errors.New("当前没有进行中的战斗")
	ErrAlreadyInCombat  = errors.New("角色已经在战斗中")
	ErrNPCNotFound      = errors.New("目标NPC不存在")
	ErrCombatFinished   = errors.New("战斗已经结束")
)

type CombatService struct {
	db      *sqlx.DB
	invRepo *repository.InventoryRepo
	mu      sync.RWMutex
	combats map[int64]*engine.CombatEngine // charID -> combat
}

func NewCombatService(db *sqlx.DB, invRepo *repository.InventoryRepo) *CombatService {
	return &CombatService{
		db:      db,
		invRepo: invRepo,
		combats: make(map[int64]*engine.CombatEngine),
	}
}

func (s *CombatService) DB() *sqlx.DB { return s.db }

type EngageNPCRequest struct {
	NPCDefID int64 `json:"npc_def_id" binding:"required"`
}

func (s *CombatService) EngageNPC(ctx context.Context, charID int64, req *EngageNPCRequest) (*model.CombatState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.combats[charID]; exists {
		return nil, ErrAlreadyInCombat
	}

	// Get player ship stats (simplified: use race default ship)
	var char model.Character
	err := s.db.GetContext(ctx, &char, `SELECT * FROM characters WHERE id = $1`, charID)
	if err != nil {
		return nil, err
	}

	// Get first available ship or use defaults
	defaultShieldResist := model.ResistProfile{Kinetic: 0.15, Thermal: 0.40, EM: 0.30, Explosive: 0.10}
	defaultArmorResist := model.ResistProfile{Kinetic: 0.40, Thermal: 0.20, EM: 0.10, Explosive: 0.30}

	playerShip := model.CombatParticipant{
		ID:              charID,
		Name:            char.Name,
		Type:            "player",
		Team:            "a",
		ShieldCurrent:   2000,
		ShieldMax:       2000,
		ArmorCurrent:    1500,
		ArmorMax:        1500,
		StructureCurrent: 1000,
		StructureMax:    1000,
		CapCurrent:      500,
		Distance:        20000,
		DamagePerTick:   80,
		DamageType:      model.DamageKinetic,
		ShieldRecharge:  30,
		Speed:           300,
		Signature:       100,
		OptimalRange:    15000,
		ShieldResist:    defaultShieldResist,
		ArmorResist:     defaultArmorResist,
	}

	type ShipDefWithResist struct {
		ShieldHP       int     `db:"shield_hp"`
		ArmorHP        int     `db:"armor_hp"`
		StructureHP    int     `db:"structure_hp"`
		ShieldRecharge int     `db:"shield_recharge"`
		MaxSpeed       int     `db:"max_speed"`
		Signature      int     `db:"signature"`
		ShieldResKinetic   float64 `db:"shield_res_kinetic"`
		ShieldResThermal   float64 `db:"shield_res_thermal"`
		ShieldResEM        float64 `db:"shield_res_em"`
		ShieldResExplosive float64 `db:"shield_res_explosive"`
		ArmorResKinetic    float64 `db:"armor_res_kinetic"`
		ArmorResThermal    float64 `db:"armor_res_thermal"`
		ArmorResEM         float64 `db:"armor_res_em"`
		ArmorResExplosive  float64 `db:"armor_res_explosive"`
	}
	var shipDef ShipDefWithResist
	err = s.db.GetContext(ctx, &shipDef,
		`SELECT sd.shield_hp,sd.armor_hp,sd.structure_hp,sd.shield_recharge,sd.max_speed,sd.signature,
		 sd.shield_res_kinetic,sd.shield_res_thermal,sd.shield_res_em,sd.shield_res_explosive,
		 sd.armor_res_kinetic,sd.armor_res_thermal,sd.armor_res_em,sd.armor_res_explosive
		 FROM ship_defs sd JOIN ships s ON s.ship_def_id = sd.id
		 WHERE s.character_id = $1 AND s.is_active = true AND s.is_destroyed = false
		 LIMIT 1`, charID)
	if err == nil {
		playerShip.ShieldCurrent = shipDef.ShieldHP
		playerShip.ShieldMax = shipDef.ShieldHP
		playerShip.ArmorCurrent = shipDef.ArmorHP
		playerShip.ArmorMax = shipDef.ArmorHP
		playerShip.StructureCurrent = shipDef.StructureHP
		playerShip.StructureMax = shipDef.StructureHP
		playerShip.ShieldRecharge = shipDef.ShieldRecharge
		playerShip.Speed = shipDef.MaxSpeed
		playerShip.Signature = shipDef.Signature
		playerShip.ShieldResist = model.ResistProfile{
			Kinetic: shipDef.ShieldResKinetic, Thermal: shipDef.ShieldResThermal,
			EM: shipDef.ShieldResEM, Explosive: shipDef.ShieldResExplosive,
		}
		playerShip.ArmorResist = model.ResistProfile{
			Kinetic: shipDef.ArmorResKinetic, Thermal: shipDef.ArmorResThermal,
			EM: shipDef.ArmorResEM, Explosive: shipDef.ArmorResExplosive,
		}
	}

	type NPCDefWithResist struct {
		ID             int64              `db:"id"`
		Name           string             `db:"name"`
		NPCType        string             `db:"npc_type"`
		Tier           int                `db:"tier"`
		ShieldHP       int                `db:"shield_hp"`
		ArmorHP        int                `db:"armor_hp"`
		StructureHP    int                `db:"structure_hp"`
		ShieldRecharge int                `db:"shield_recharge"`
		DamagePerTick  int                `db:"damage_per_tick"`
		DamageType     model.DamageType   `db:"damage_type"`
		OptimalRange   int                `db:"optimal_range"`
		Speed          int                `db:"speed"`
		Signature      int                `db:"signature"`
		Bounty         int64              `db:"bounty"`
		AIBehavior     string             `db:"ai_behavior"`
		ShieldResKinetic   float64 `db:"shield_res_kinetic"`
		ShieldResThermal   float64 `db:"shield_res_thermal"`
		ShieldResEM        float64 `db:"shield_res_em"`
		ShieldResExplosive float64 `db:"shield_res_explosive"`
		ArmorResKinetic    float64 `db:"armor_res_kinetic"`
		ArmorResThermal    float64 `db:"armor_res_thermal"`
		ArmorResEM         float64 `db:"armor_res_em"`
		ArmorResExplosive  float64 `db:"armor_res_explosive"`
	}
	var npc NPCDefWithResist
	err = s.db.GetContext(ctx, &npc,
		`SELECT id,name,npc_type,tier,shield_hp,armor_hp,structure_hp,shield_recharge,
		 damage_per_tick,damage_type,optimal_range,speed,signature,bounty,ai_behavior,
		 shield_res_kinetic,shield_res_thermal,shield_res_em,shield_res_explosive,
		 armor_res_kinetic,armor_res_thermal,armor_res_em,armor_res_explosive
		 FROM npc_defs WHERE id = $1`, req.NPCDefID)
	if err != nil {
		return nil, ErrNPCNotFound
	}

	npcParticipant := model.CombatParticipant{
		ID:              10000 + npc.ID,
		Name:            npc.Name,
		Type:            "npc",
		Team:            "b",
		ShieldCurrent:   npc.ShieldHP,
		ShieldMax:       npc.ShieldHP,
		ArmorCurrent:    npc.ArmorHP,
		ArmorMax:        npc.ArmorHP,
		StructureCurrent: npc.StructureHP,
		StructureMax:    npc.StructureHP,
		CapCurrent:      500,
		Distance:        20000,
		DamagePerTick:   npc.DamagePerTick,
		DamageType:      npc.DamageType,
		RateOfFire:      1, // NPC默认每Tick开火
		ShieldRecharge:  npc.ShieldRecharge,
		Speed:           npc.Speed,
		Signature:       npc.Signature,
		OptimalRange:    npc.OptimalRange,
		ShieldResist: model.ResistProfile{
			Kinetic: npc.ShieldResKinetic, Thermal: npc.ShieldResThermal,
			EM: npc.ShieldResEM, Explosive: npc.ShieldResExplosive,
		},
		ArmorResist: model.ResistProfile{
			Kinetic: npc.ArmorResKinetic, Thermal: npc.ArmorResThermal,
			EM: npc.ArmorResEM, Explosive: npc.ArmorResExplosive,
		},
	}

	combatID := time.Now().UnixNano()
	eng := engine.NewCombatEngine(combatID)
	eng.AddParticipant(playerShip)
	eng.AddParticipant(npcParticipant)

	s.combats[charID] = eng

	// Process first tick
	eng.ProcessTick()

	return eng.State, nil
}

func (s *CombatService) NextTick(ctx context.Context, charID int64) (*model.CombatState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	eng, exists := s.combats[charID]
	if !exists {
		return nil, ErrNoCombat
	}

	if eng.State.Status != "active" {
		return nil, ErrCombatFinished
	}

	eng.ProcessTick()

	// If combat finished, process rewards
	if eng.State.Status == "finished" {
		s.processCombatEnd(ctx, charID, eng)
	}

	return eng.State, nil
}

func (s *CombatService) GetCombatState(ctx context.Context, charID int64) (*model.CombatState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	eng, exists := s.combats[charID]
	if !exists {
		return nil, ErrNoCombat
	}
	return eng.State, nil
}

type CombatCommand struct {
	Action   string `json:"action" binding:"required"` // change_target, set_distance, retreat
	TargetID *int64 `json:"target_id,omitempty"`
	Distance *int   `json:"distance,omitempty"`
}

func (s *CombatService) IssueCommand(ctx context.Context, charID int64, cmd *CombatCommand) (*model.CombatState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	eng, exists := s.combats[charID]
	if !exists {
		return nil, ErrNoCombat
	}

	// Find player participant
	for i := range eng.State.Participants {
		p := &eng.State.Participants[i]
		if p.ID == charID && p.Type == "player" {
			switch cmd.Action {
			case "change_target":
				if cmd.TargetID != nil {
					p.TargetID = cmd.TargetID
				}
			case "set_distance":
				if cmd.Distance != nil {
					p.Distance = *cmd.Distance
				}
			case "retreat":
				// Remove from combat
				s.processCombatEnd(ctx, charID, eng)
				delete(s.combats, charID)
				eng.State.Status = "retreated"
				eng.State.Logs = append(eng.State.Logs, "你成功跃迁撤离了战场！")
				return eng.State, nil
			}
			break
		}
	}

	return eng.State, nil
}

func (s *CombatService) processCombatEnd(ctx context.Context, charID int64, eng *engine.CombatEngine) {
	// Check if player won
	playerAlive := false
	for _, p := range eng.State.Participants {
		if p.ID == charID && !p.IsDestroyed {
			playerAlive = true
		}
	}

	if playerAlive {
		// Award bounty and loot
		for _, p := range eng.State.Participants {
			if p.Type == "npc" && p.IsDestroyed {
				npcID := p.ID - 10000

				var npc model.NPCDef
				if err := s.db.GetContext(ctx, &npc, `SELECT * FROM npc_defs WHERE id = $1`, npcID); err == nil {
					// Award bounty
					s.db.ExecContext(ctx,
						`UPDATE characters SET balance = balance + $1 WHERE id = $2`,
						npc.Bounty, charID)

					// Roll loot
					type LootRow struct {
						ItemDefID   int64   `db:"item_def_id"`
						QuantityMin int     `db:"quantity_min"`
						QuantityMax int     `db:"quantity_max"`
						DropChance  float64 `db:"drop_chance"`
					}
					var loots []LootRow
					s.db.SelectContext(ctx, &loots,
						`SELECT item_def_id, quantity_min, quantity_max, drop_chance FROM npc_loot_table WHERE npc_def_id = $1`, npcID)

					var systemID int64
					s.db.GetContext(ctx, &systemID, `SELECT current_system_id FROM characters WHERE id = $1`, charID)

					for _, l := range loots {
						if rand.Float64() < l.DropChance {
							qty := int64(l.QuantityMin + rand.Intn(l.QuantityMax-l.QuantityMin+1))
							s.invRepo.AddOrUpsertItem(ctx, "character", charID, l.ItemDefID, qty, systemID)

							def, _ := s.invRepo.GetItemDef(ctx, l.ItemDefID)
							name := "物品"
							if def != nil {
								name = def.Name
							}
							eng.State.Logs = append(eng.State.Logs,
								fmt.Sprintf("  ▸ 战利品: %s x%d", name, qty))
						}
					}

					eng.State.Logs = append(eng.State.Logs,
						fmt.Sprintf("  ▸ 赏金: +%d 星币", npc.Bounty))
				}
			}
		}
	}

	delete(s.combats, charID)
}
