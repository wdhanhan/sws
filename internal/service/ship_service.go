package service

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/starfall-warsong/sws/internal/model"
)

var (
	ErrShipNotFound    = errors.New("舰船不存在")
	ErrShipNotOwned    = errors.New("这不是你的舰船")
	ErrSlotInvalid     = errors.New("无效的槽位")
	ErrSlotOccupied    = errors.New("槽位已被占用")
	ErrPGExceeded      = errors.New("能量栅格超出上限")
	ErrCPUExceeded     = errors.New("CPU超出上限")
	ErrModuleNotFound  = errors.New("模块不存在")
	ErrSlotTypeMismatch = errors.New("该模块不适配此槽位")
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

	ShieldMax       int     `json:"shield_max,omitempty"`
	ArmorMax        int     `json:"armor_max,omitempty"`
	StructMax       int     `json:"structure_max,omitempty"`
	MaxSpeed        int     `json:"max_speed,omitempty"`
	HighSlots       int     `json:"high_slots,omitempty"`
	MidSlots        int     `json:"mid_slots,omitempty"`
	LowSlots        int     `json:"low_slots,omitempty"`
	Powergrid       int     `json:"powergrid,omitempty"`
	CPU             int     `json:"cpu,omitempty"`
	Capacitor       int     `json:"capacitor,omitempty"`
	CapRecharge     int     `json:"cap_recharge,omitempty"`
	ShieldRecharge  int     `json:"shield_recharge,omitempty"`
	Signature       int     `json:"signature,omitempty"`
	DPS             int     `json:"dps"`
	UsedPG          int     `json:"used_pg"`
	UsedCPU         int     `json:"used_cpu"`

	ShieldResKinetic   float64 `json:"shield_res_kinetic,omitempty"`
	ShieldResThermal   float64 `json:"shield_res_thermal,omitempty"`
	ShieldResEM        float64 `json:"shield_res_em,omitempty"`
	ShieldResExplosive float64 `json:"shield_res_explosive,omitempty"`
	ArmorResKinetic    float64 `json:"armor_res_kinetic,omitempty"`
	ArmorResThermal    float64 `json:"armor_res_thermal,omitempty"`
	ArmorResEM         float64 `json:"armor_res_em,omitempty"`
	ArmorResExplosive  float64 `json:"armor_res_explosive,omitempty"`
}

type FittingInfo struct {
	SlotType      string  `db:"slot_type" json:"slot_type"`
	SlotIndex     int     `db:"slot_index" json:"slot_index"`
	ModuleName    string  `db:"module_name" json:"module_name"`
	ModuleID      int64   `db:"module_item_id" json:"module_item_id"`
	IsActive      bool    `db:"is_active" json:"is_active"`
	ModuleType    string  `db:"module_type" json:"module_type"`
	DamagePerTick int     `db:"damage_per_tick" json:"damage_per_tick,omitempty"`
	DamageType    string  `db:"damage_type" json:"damage_type,omitempty"`
	OptimalRange  int     `db:"optimal_range" json:"optimal_range,omitempty"`
	FalloffRange  int     `db:"falloff_range" json:"falloff_range,omitempty"`
	TrackingSpeed float64 `db:"tracking_speed" json:"tracking_speed,omitempty"`
	RateOfFire    int     `db:"rate_of_fire" json:"rate_of_fire,omitempty"`
	CapCost       int     `db:"cap_cost" json:"cap_cost,omitempty"`
	BonusType     string  `db:"bonus_type" json:"bonus_type,omitempty"`
	BonusValue    float64 `db:"bonus_value" json:"bonus_value,omitempty"`
	PGCost        int     `db:"pg_cost" json:"pg_cost"`
	CPUCost       int     `db:"cpu_cost" json:"cpu_cost"`
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
			Name               string  `db:"name"`
			ShipClass          string  `db:"ship_class"`
			ShipRole           string  `db:"ship_role"`
			Tier               int     `db:"tier"`
			ShieldHP           int     `db:"shield_hp"`
			ArmorHP            int     `db:"armor_hp"`
			StructHP           int     `db:"structure_hp"`
			MaxSpeed           int     `db:"max_speed"`
			HighSlots          int     `db:"high_slots"`
			MidSlots           int     `db:"mid_slots"`
			LowSlots           int     `db:"low_slots"`
			PG                 int     `db:"powergrid"`
			CPU                int     `db:"cpu"`
			Capacitor          int     `db:"capacitor"`
			CapRecharge        int     `db:"cap_recharge"`
			ShieldRecharge     int     `db:"shield_recharge"`
			Signature          int     `db:"signature"`
			ShieldResKinetic   float64 `db:"shield_res_kinetic"`
			ShieldResThermal   float64 `db:"shield_res_thermal"`
			ShieldResEM        float64 `db:"shield_res_em"`
			ShieldResExplosive float64 `db:"shield_res_explosive"`
			ArmorResKinetic    float64 `db:"armor_res_kinetic"`
			ArmorResThermal    float64 `db:"armor_res_thermal"`
			ArmorResEM         float64 `db:"armor_res_em"`
			ArmorResExplosive  float64 `db:"armor_res_explosive"`
		}
		s.db.GetContext(ctx, &def,
			`SELECT name,ship_class,ship_role,tier,shield_hp,armor_hp,structure_hp,max_speed,
			 high_slots,mid_slots,low_slots,powergrid,cpu,capacitor,cap_recharge,shield_recharge,signature,
			 shield_res_kinetic,shield_res_thermal,shield_res_em,shield_res_explosive,
			 armor_res_kinetic,armor_res_thermal,armor_res_em,armor_res_explosive
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
		ships[i].Capacitor = def.Capacitor
		ships[i].CapRecharge = def.CapRecharge
		ships[i].ShieldRecharge = def.ShieldRecharge
		ships[i].Signature = def.Signature
		ships[i].ShieldResKinetic = def.ShieldResKinetic
		ships[i].ShieldResThermal = def.ShieldResThermal
		ships[i].ShieldResEM = def.ShieldResEM
		ships[i].ShieldResExplosive = def.ShieldResExplosive
		ships[i].ArmorResKinetic = def.ArmorResKinetic
		ships[i].ArmorResThermal = def.ArmorResThermal
		ships[i].ArmorResEM = def.ArmorResEM
		ships[i].ArmorResExplosive = def.ArmorResExplosive

		// DPS and used PG/CPU from fittings
		type FitStats struct {
			DPS     int `db:"dps"`
			UsedPG  int `db:"used_pg"`
			UsedCPU int `db:"used_cpu"`
		}
		var fs FitStats
		s.db.GetContext(ctx, &fs,
			`SELECT COALESCE(SUM(CASE WHEN COALESCE(i.damage_per_tick,0)>0 THEN i.damage_per_tick/GREATEST(i.rate_of_fire,1) ELSE 0 END),0) as dps,
			 COALESCE(SUM(i.pg_cost),0) as used_pg, COALESCE(SUM(i.cpu_cost),0) as used_cpu
			 FROM ship_fittings sf JOIN item_defs i ON i.id=sf.module_item_id WHERE sf.ship_id=$1`, ships[i].ID)
		ships[i].DPS = fs.DPS
		ships[i].UsedPG = fs.UsedPG
		ships[i].UsedCPU = fs.UsedCPU

		// Apply mid/low module bonuses to displayed stats
		type ModBonus struct {
			BonusType  string  `db:"bonus_type"`
			BonusValue float64 `db:"bonus_value"`
		}
		var bonuses []ModBonus
		s.db.SelectContext(ctx, &bonuses,
			`SELECT COALESCE(i.bonus_type,'') as bonus_type, COALESCE(i.bonus_value,0) as bonus_value
			 FROM ship_fittings sf JOIN item_defs i ON i.id=sf.module_item_id
			 WHERE sf.ship_id=$1 AND COALESCE(i.bonus_type,'') != ''`, ships[i].ID)

		for _, b := range bonuses {
			switch b.BonusType {
			case "shield_boost":
				ships[i].ShieldRecharge += int(b.BonusValue)
			case "shield_hp_bonus":
				bonus := int(float64(ships[i].ShieldMax) * b.BonusValue)
				ships[i].ShieldMax += bonus
				ships[i].ShieldCur += bonus
			case "armor_repair":
				// displayed via fittings detail
			case "armor_hp":
				add := int(b.BonusValue)
				ships[i].ArmorMax += add
				ships[i].ArmorCur += add
			case "armor_kinetic_resist":
				ships[i].ArmorResKinetic += b.BonusValue
			case "armor_thermal_resist":
				ships[i].ArmorResThermal += b.BonusValue
			case "armor_em_resist":
				ships[i].ArmorResEM += b.BonusValue
			case "armor_explosive_resist":
				ships[i].ArmorResExplosive += b.BonusValue
			case "armor_omni_resist":
				ships[i].ArmorResKinetic += b.BonusValue
				ships[i].ArmorResThermal += b.BonusValue
				ships[i].ArmorResEM += b.BonusValue
				ships[i].ArmorResExplosive += b.BonusValue
			case "shield_kinetic_resist":
				ships[i].ShieldResKinetic += b.BonusValue
			case "shield_thermal_resist":
				ships[i].ShieldResThermal += b.BonusValue
			case "shield_em_resist":
				ships[i].ShieldResEM += b.BonusValue
			case "shield_explosive_resist":
				ships[i].ShieldResExplosive += b.BonusValue
			case "shield_omni_resist":
				ships[i].ShieldResKinetic += b.BonusValue
				ships[i].ShieldResThermal += b.BonusValue
				ships[i].ShieldResEM += b.BonusValue
				ships[i].ShieldResExplosive += b.BonusValue
			case "all_resist":
				ships[i].ShieldResKinetic += b.BonusValue
				ships[i].ShieldResThermal += b.BonusValue
				ships[i].ShieldResEM += b.BonusValue
				ships[i].ShieldResExplosive += b.BonusValue
				ships[i].ArmorResKinetic += b.BonusValue
				ships[i].ArmorResThermal += b.BonusValue
				ships[i].ArmorResEM += b.BonusValue
				ships[i].ArmorResExplosive += b.BonusValue
			case "speed_bonus":
				ships[i].MaxSpeed += int(float64(ships[i].MaxSpeed) * b.BonusValue)
			case "cap_boost":
				ships[i].Capacitor += int(b.BonusValue)
			case "cap_recharge_bonus":
				ships[i].CapRecharge += int(float64(ships[i].CapRecharge) * b.BonusValue)
			case "tracking_bonus", "signature_reduction", "agility_bonus":
				// affects combat but not displayed in summary stats
			}
		}

		// Clamp current HP to not exceed max
		if ships[i].ShieldCur > ships[i].ShieldMax { ships[i].ShieldCur = ships[i].ShieldMax }
		if ships[i].ArmorCur > ships[i].ArmorMax { ships[i].ArmorCur = ships[i].ArmorMax }

		// Clamp resists to 85%
		clamp := func(v *float64) { if *v > 0.85 { *v = 0.85 } }
		clamp(&ships[i].ShieldResKinetic); clamp(&ships[i].ShieldResThermal)
		clamp(&ships[i].ShieldResEM); clamp(&ships[i].ShieldResExplosive)
		clamp(&ships[i].ArmorResKinetic); clamp(&ships[i].ArmorResThermal)
		clamp(&ships[i].ArmorResEM); clamp(&ships[i].ArmorResExplosive)
	}
	return ships, err
}

type FitModuleRequest struct {
	ShipID       int64  `json:"ship_id" binding:"required"`
	SlotType     string `json:"slot_type" binding:"required"` // high, mid, low
	SlotIndex    int    `json:"slot_index" binding:"min=0"`
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

	// Check module exists and slot type matches
	type ModInfo struct {
		PG       int    `db:"pg_cost"`
		CPU      int    `db:"cpu_cost"`
		SlotType string `db:"slot_type"`
	}
	var modInfo ModInfo
	err = s.db.GetContext(ctx, &modInfo,
		`SELECT COALESCE(pg_cost,0) as pg_cost, COALESCE(cpu_cost,0) as cpu_cost, COALESCE(slot_type,'') as slot_type FROM item_defs WHERE id = $1`, req.ModuleItemID)
	if err != nil {
		return ErrModuleNotFound
	}
	if modInfo.SlotType != "" && modInfo.SlotType != req.SlotType {
		return ErrSlotTypeMismatch
	}
	modCost := struct{ PG, CPU int }{modInfo.PG, modInfo.CPU}

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
	var fittings []FittingInfo
	err := s.db.SelectContext(ctx, &fittings,
		`SELECT sf.slot_type, sf.slot_index, sf.module_item_id, sf.is_active,
		 i.name as module_name, COALESCE(i.module_type,'') as module_type,
		 COALESCE(i.damage_per_tick,0) as damage_per_tick, COALESCE(i.damage_type,'') as damage_type,
		 COALESCE(i.optimal_range,0) as optimal_range, COALESCE(i.falloff_range,0) as falloff_range,
		 COALESCE(i.tracking_speed,0) as tracking_speed, COALESCE(i.rate_of_fire,1) as rate_of_fire,
		 COALESCE(i.cap_cost,0) as cap_cost, COALESCE(i.bonus_type,'') as bonus_type,
		 COALESCE(i.bonus_value,0) as bonus_value, COALESCE(i.pg_cost,0) as pg_cost,
		 COALESCE(i.cpu_cost,0) as cpu_cost
		 FROM ship_fittings sf JOIN item_defs i ON i.id = sf.module_item_id
		 WHERE sf.ship_id = $1 ORDER BY sf.slot_type, sf.slot_index`, shipID)
	if err != nil {
		return nil, err
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
