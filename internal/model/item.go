package model

import "time"

type ItemCategory string

const (
	ItemCatOre       ItemCategory = "ore"        // 矿石
	ItemCatMineral   ItemCategory = "mineral"     // 精炼金属
	ItemCatGas       ItemCategory = "gas"         // 气体
	ItemCatIce       ItemCategory = "ice"         // 冰矿
	ItemCatPlanetary ItemCategory = "planetary"   // 行星资源
	ItemCatAlloy     ItemCategory = "alloy"       // 合金
	ItemCatComponent ItemCategory = "component"   // 组件
	ItemCatModule    ItemCategory = "module"      // 舰船模块
	ItemCatShip      ItemCategory = "ship"        // 舰船
	ItemCatAmmo      ItemCategory = "ammo"        // 弹药
	ItemCatBlueprint ItemCategory = "blueprint"   // 蓝图
	ItemCatImplant   ItemCategory = "implant"     // 植入体
	ItemCatConsume   ItemCategory = "consumable"  // 消耗品
	ItemCatSalvage   ItemCategory = "salvage"     // 打捞物
)

type ItemDef struct {
	ID            int64        `db:"id" json:"id"`
	Name          string       `db:"name" json:"name"`
	Category      ItemCategory `db:"category" json:"category"`
	Description   string       `db:"description" json:"description"`
	Volume        float64      `db:"volume" json:"volume"`
	BasePrice     int64        `db:"base_price" json:"base_price"`
	Stackable     bool         `db:"stackable" json:"stackable"`
	TechLevel     int          `db:"tech_level" json:"tech_level"`
	SlotType      string       `db:"slot_type" json:"slot_type,omitempty"`
	ModuleType    string       `db:"module_type" json:"module_type,omitempty"`
	DamagePerTick int          `db:"damage_per_tick" json:"damage_per_tick,omitempty"`
	DamageType    string       `db:"damage_type" json:"damage_type,omitempty"`
	OptimalRange  int          `db:"optimal_range" json:"optimal_range,omitempty"`
	FalloffRange  int          `db:"falloff_range" json:"falloff_range,omitempty"`
	TrackingSpeed float64      `db:"tracking_speed" json:"tracking_speed,omitempty"`
	RateOfFire    int          `db:"rate_of_fire" json:"rate_of_fire,omitempty"`
	PGCost        int          `db:"pg_cost" json:"pg_cost,omitempty"`
	CPUCost       int          `db:"cpu_cost" json:"cpu_cost,omitempty"`
	CapCost       int          `db:"cap_cost" json:"cap_cost,omitempty"`
	BonusType     string       `db:"bonus_type" json:"bonus_type,omitempty"`
	BonusValue    float64      `db:"bonus_value" json:"bonus_value,omitempty"`
}

type InventoryItem struct {
	ID          int64     `db:"id" json:"id"`
	OwnerType   string    `db:"owner_type" json:"owner_type"` // "character" / "station" / "ship"
	OwnerID     int64     `db:"owner_id" json:"owner_id"`
	ItemDefID   int64     `db:"item_def_id" json:"item_def_id"`
	Quantity    int64     `db:"quantity" json:"quantity"`
	LocationID  int64     `db:"location_id" json:"location_id"` // 所在星系/空间站ID
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
