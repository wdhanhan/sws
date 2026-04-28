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
	ID          int64        `db:"id" json:"id"`
	Name        string       `db:"name" json:"name"`
	Category    ItemCategory `db:"category" json:"category"`
	Description string       `db:"description" json:"description"`
	Volume      float64      `db:"volume" json:"volume"`       // 体积(m³)
	BasePrice   int64        `db:"base_price" json:"base_price"` // NPC保底价(分)
	Stackable   bool         `db:"stackable" json:"stackable"`
	TechLevel   int          `db:"tech_level" json:"tech_level"` // 产业链层级 1-9
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
