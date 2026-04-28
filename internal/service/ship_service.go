package service

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/starfall-warsong/sws/internal/model"
)

var (
	ErrShipNotFound   = errors.New("舰船不存在")
	ErrShipNotOwned   = errors.New("这不是你的舰船")
	ErrSlotInvalid    = errors.New("无效的槽位")
	ErrSlotOccupied   = errors.New("槽位已被占用")
	ErrPGExceeded     = errors.New("能量栅格超出上限")
	ErrCPUExceeded    = errors.New("CPU超出上限")
	ErrModuleNotFound = errors.New("模块不存在")
)

type ShipService struct {
	db *sqlx.DB
}

func NewShipService(db *sqlx.DB) *ShipService {
	return &ShipService{db: db}
}

type ShipInfo struct {
	ID           int64  `db:"id" json:"id"`
	CharacterID  int64  `db:"character_id" json:"character_id"`
	ShipDefID    int64  `db:"ship_def_id" json:"ship_def_id"`
	ShipName     string `db:"name" json:"name"`
	DefName      string `json:"def_name"`
	ShipClass    string `json:"ship_class"`
	ShipRole     string `json:"ship_role"`
	Tier         int    `json:"tier"`
	IsActive     bool   `db:"is_active" json:"is_active"`
	IsDestroyed  bool   `db:"is_destroyed" json:"is_destroyed"`
	ShieldCur    int    `db:"shield_current" json:"shield_current"`
	ArmorCur     int    `db:"armor_current" json:"armor_current"`
	StructCur    int    `db:"structure_current" json:"structure_current"`
	// 舰船属性（从定义表填充）
	ShieldMax    int    `json:"shield_max,omitempty"`
	ArmorMax     int    `json:"armor_max,omitempty"`
	StructMax    int    `json:"structure_max,omitempty"`
	MaxSpeed     int    `json:"max_speed,omitempty"`
	HighSlots    int    `json:"high_slots,omitempty"`
	MidSlots     int    `json:"mid_slots,omitempty"`
	LowSlots     int    `json:"low_slots,omitempty"`
	Powergrid    int    `json:"powergrid,omitempty"`
	CPU          int    `json:"cpu,omitempty"`
}

type FittingInfo struct {
	SlotType   string `db:"slot_type" json:"slot_type"`
	SlotIndex  int    `db:"slot_index" json:"slot_index"`
	ModuleName string `json:"module_name"`
	ModuleID   int64  `db:"module_item_id" json:"module_item_id"`
	IsActive   bool   `db:"is_active" json:"is_active"`
}

func (s *ShipService) GetShipDefs(ctx context.Context, raceID int) ([]model.ShipDef, error) {
	var defs []model.ShipDef
	if raceID > 0 {
		return defs, s.db.SelectContext(ctx, &defs,
			`SELECT id,name,race_id,tier,ship_class,ship_role,shield_hp,armor_hp,structure_hp,
			 max_speed,align_ticks,signature,high_slots,mid_slots,low_slots,cargo_m3,powergrid,cpu
			 FROM ship_defs WHERE race_id = $1 ORDER BY tier,id`, raceID)
	}
	return defs, s.db.SelectContext(ctx, &defs,
		`SELECT id,name,race_id,tier,ship_class,ship_role,shield_hp,armor_hp,structure_hp,
		 max_speed,align_ticks,signature,high_slots,mid_slots,low_slots,cargo_m3,powergrid,cpu
		 FROM ship_defs ORDER BY race_id,tier,id`)
}

func (s *ShipService) BoardShip(ctx context.Context, charID, shipDefID int64, shipName string) (*ShipInfo, error) {
	// Get ship def for initial HP
	type DefHP struct {
		ShieldHP    int `db:"shield_hp"`
		ArmorHP     int `db:"armor_hp"`
		StructureHP int `db:"structure_hp"`
		Capacitor   int `db:"capacitor"`
	}
	var hp DefHP
	err := s.db.GetContext(ctx, &hp,
		`SELECT shield_hp, armor_hp, structure_hp, capacitor FROM ship_defs WHERE id = $1`, shipDefID)
	if err != nil {
		return nil, ErrShipNotFound
	}

	var systemID int64
	s.db.GetContext(ctx, &systemID, `SELECT current_system_id FROM characters WHERE id = $1`, charID)

	// Deactivate current ship
	s.db.ExecContext(ctx, `UPDATE ships SET is_active = false WHERE character_id = $1`, charID)

	// Create and activate new ship
	var ship ShipInfo
	err = s.db.QueryRowxContext(ctx,
		`INSERT INTO ships (character_id, ship_def_id, name, shield_current, armor_current, structure_current, cap_current, is_active, location_system_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,true,$8) RETURNING id,character_id,ship_def_id,name,is_active,is_destroyed,shield_current,armor_current,structure_current`,
		charID, shipDefID, shipName, hp.ShieldHP, hp.ArmorHP, hp.StructureHP, hp.Capacitor, systemID,
	).StructScan(&ship)
	if err != nil {
		return nil, err
	}

	return &ship, nil
}

func (s *ShipService) GetMyShips(ctx context.Context, charID int64) ([]ShipInfo, error) {
	var ships []ShipInfo
	err := s.db.SelectContext(ctx, &ships,
		`SELECT id,character_id,ship_def_id,name,is_active,is_destroyed,shield_current,armor_current,structure_current
		 FROM ships WHERE character_id = $1 AND is_destroyed = false ORDER BY is_active DESC, id`, charID)
	if err != nil {
		return nil, err
	}
	for i := range ships {
		var def struct {
			Name      string `db:"name"`
			ShipClass string `db:"ship_class"`
			ShipRole  string `db:"ship_role"`
			Tier      int    `db:"tier"`
			ShieldHP  int    `db:"shield_hp"`
			ArmorHP   int    `db:"armor_hp"`
			StructHP  int    `db:"structure_hp"`
			MaxSpeed  int    `db:"max_speed"`
			HighSlots int    `db:"high_slots"`
			MidSlots  int    `db:"mid_slots"`
			LowSlots  int    `db:"low_slots"`
			PG        int    `db:"powergrid"`
			CPU       int    `db:"cpu"`
		}
		s.db.GetContext(ctx, &def,
			`SELECT name,ship_class,ship_role,tier,shield_hp,armor_hp,structure_hp,max_speed,high_slots,mid_slots,low_slots,powergrid,cpu
			 FROM ship_defs WHERE id = $1`, ships[i].ShipDefID)
		ships[i].DefName = def.Name
		ships[i].ShipClass = def.ShipClass
		ships[i].ShipRole = def.ShipRole
		ships[i].Tier = def.Tier
		ships[i].ShieldMax = def.ShieldHP
		ships[i].ArmorMax = def.ArmorHP
		ships[i].StructMax = def.StructHP
		ships[i].MaxSpeed = def.MaxSpeed
		ships[i].HighSlots = def.HighSlots
		ships[i].MidSlots = def.MidSlots
		ships[i].LowSlots = def.LowSlots
		ships[i].Powergrid = def.PG
		ships[i].CPU = def.CPU
	}
	return ships, err
}

type FitModuleRequest struct {
	ShipID       int64  `json:"ship_id" binding:"required"`
	SlotType     string `json:"slot_type" binding:"required"` // high, mid, low
	SlotIndex    int    `json:"slot_index" binding:"required"`
	ModuleItemID int64  `json:"module_item_id" binding:"required"`
}

func (s *ShipService) FitModule(ctx context.Context, charID int64, req *FitModuleRequest) error {
	// Verify ship ownership
	var shipCharID int64
	err := s.db.GetContext(ctx, &shipCharID, `SELECT character_id FROM ships WHERE id = $1`, req.ShipID)
	if err != nil {
		return ErrShipNotFound
	}
	if shipCharID != charID {
		return ErrShipNotOwned
	}

	// Check slot exists on ship
	type SlotCount struct {
		High int `db:"high_slots"`
		Mid  int `db:"mid_slots"`
		Low  int `db:"low_slots"`
		PG   int `db:"powergrid"`
		CPU  int `db:"cpu"`
	}
	var slots SlotCount
	s.db.GetContext(ctx, &slots,
		`SELECT sd.high_slots, sd.mid_slots, sd.low_slots, sd.powergrid, sd.cpu
		 FROM ship_defs sd JOIN ships s ON s.ship_def_id = sd.id WHERE s.id = $1`, req.ShipID)

	maxSlots := 0
	switch req.SlotType {
	case "high":
		maxSlots = slots.High
	case "mid":
		maxSlots = slots.Mid
	case "low":
		maxSlots = slots.Low
	default:
		return ErrSlotInvalid
	}
	if req.SlotIndex < 0 || req.SlotIndex >= maxSlots {
		return ErrSlotInvalid
	}

	// Check slot not occupied
	var occupied int
	s.db.GetContext(ctx, &occupied,
		`SELECT COUNT(*) FROM ship_fittings WHERE ship_id=$1 AND slot_type=$2 AND slot_index=$3`,
		req.ShipID, req.SlotType, req.SlotIndex)
	if occupied > 0 {
		return ErrSlotOccupied
	}

	// Check PG/CPU
	type ModCost struct {
		PG  int `db:"pg_cost"`
		CPU int `db:"cpu_cost"`
	}
	var modCost ModCost
	err = s.db.GetContext(ctx, &modCost,
		`SELECT COALESCE(pg_cost,0) as pg_cost, COALESCE(cpu_cost,0) as cpu_cost FROM item_defs WHERE id = $1`, req.ModuleItemID)
	if err != nil {
		return ErrModuleNotFound
	}

	var usedPG, usedCPU int
	s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(i.pg_cost),0), COALESCE(SUM(i.cpu_cost),0)
		 FROM ship_fittings sf JOIN item_defs i ON i.id = sf.module_item_id
		 WHERE sf.ship_id = $1`, req.ShipID).Scan(&usedPG, &usedCPU)

	if usedPG+modCost.PG > slots.PG {
		return ErrPGExceeded
	}
	if usedCPU+modCost.CPU > slots.CPU {
		return ErrCPUExceeded
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO ship_fittings (ship_id, slot_type, slot_index, module_item_id, is_active)
		 VALUES ($1,$2,$3,$4,true)`,
		req.ShipID, req.SlotType, req.SlotIndex, req.ModuleItemID)
	return err
}

func (s *ShipService) GetFitting(ctx context.Context, shipID int64) ([]FittingInfo, error) {
	type Row struct {
		SlotType     string `db:"slot_type"`
		SlotIndex    int    `db:"slot_index"`
		ModuleItemID int64  `db:"module_item_id"`
		IsActive     bool   `db:"is_active"`
	}
	var rows []Row
	err := s.db.SelectContext(ctx, &rows,
		`SELECT slot_type, slot_index, module_item_id, is_active
		 FROM ship_fittings WHERE ship_id = $1 ORDER BY slot_type, slot_index`, shipID)
	if err != nil {
		return nil, err
	}

	fittings := make([]FittingInfo, len(rows))
	for i, r := range rows {
		var modName string
		s.db.GetContext(ctx, &modName, `SELECT name FROM item_defs WHERE id = $1`, r.ModuleItemID)
		fittings[i] = FittingInfo{
			SlotType:   r.SlotType,
			SlotIndex:  r.SlotIndex,
			ModuleName: modName,
			ModuleID:   r.ModuleItemID,
			IsActive:   r.IsActive,
		}
	}
	return fittings, nil
}

func (s *ShipService) RemoveModule(ctx context.Context, charID, shipID int64, slotType string, slotIndex int) error {
	var shipCharID int64
	s.db.GetContext(ctx, &shipCharID, `SELECT character_id FROM ships WHERE id = $1`, shipID)
	if shipCharID != charID {
		return ErrShipNotOwned
	}
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM ship_fittings WHERE ship_id=$1 AND slot_type=$2 AND slot_index=$3`,
		shipID, slotType, slotIndex)
	return err
}
