package model

import "time"

type ArmID int

const (
	ArmFire  ArmID = 1 // 焚天之臂
	ArmEarth ArmID = 2 // 厚土之臂
	ArmWind  ArmID = 3 // 罡风之臂
	ArmWater ArmID = 4 // 渊水之臂
	ArmCore  ArmID = 5 // 中枢核心
	ArmVoid  ArmID = 6 // 旋臂间虚空
	ArmOuter ArmID = 7 // 外缘未知
)

var ArmNames = map[ArmID]string{
	ArmFire:  "焚天之臂",
	ArmEarth: "厚土之臂",
	ArmWind:  "罡风之臂",
	ArmWater: "渊水之臂",
	ArmCore:  "中枢核心",
	ArmVoid:  "旋臂间虚空",
	ArmOuter: "外缘未知",
}

type StarType int

const (
	StarTypeO       StarType = 1  // 蓝超巨星
	StarTypeB       StarType = 2  // 蓝白巨星
	StarTypeA       StarType = 3  // 白色主序
	StarTypeF       StarType = 4  // 黄白矮星
	StarTypeG       StarType = 5  // 黄矮星(类太阳)
	StarTypeK       StarType = 6  // 橙矮星
	StarTypeM       StarType = 7  // 红矮星
	StarTypeNeutron StarType = 8  // 中子星
	StarTypeBlackH  StarType = 9  // 黑洞
	StarTypePulsar  StarType = 10 // 脉冲星
	StarTypeBinary  StarType = 11 // 双星系统
)

var StarTypeNames = map[StarType]string{
	StarTypeO:       "蓝超巨星",
	StarTypeB:       "蓝白巨星",
	StarTypeA:       "白色主序星",
	StarTypeF:       "黄白矮星",
	StarTypeG:       "黄矮星",
	StarTypeK:       "橙矮星",
	StarTypeM:       "红矮星",
	StarTypeNeutron: "中子星",
	StarTypeBlackH:  "黑洞",
	StarTypePulsar:  "脉冲星",
	StarTypeBinary:  "双星系统",
}

type StarSystem struct {
	ID            int64    `db:"id" json:"id"`
	Name          string   `db:"name" json:"name"`
	ArmID         ArmID    `db:"arm_id" json:"arm_id"`
	CoordX        float64  `db:"coord_x" json:"coord_x"`
	CoordY        float64  `db:"coord_y" json:"coord_y"`
	CoordZ        float64  `db:"coord_z" json:"coord_z"`
	SecurityLevel float64  `db:"security_level" json:"security_level"` // -1.0 ~ 1.0
	StarType      StarType `db:"star_type" json:"star_type"`
	PlanetCount   int      `db:"planet_count" json:"planet_count"`
	BeltCount     int      `db:"belt_count" json:"belt_count"`
	HasAnomaly    bool     `db:"has_anomaly" json:"has_anomaly"`
	HasRuins      bool     `db:"has_ruins" json:"has_ruins"`
	OwnerID       *int64   `db:"owner_id" json:"owner_id,omitempty"` // 主权持有国家ID
	Seed          int64    `db:"seed" json:"seed"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type Stargate struct {
	ID       int64 `db:"id" json:"id"`
	FromID   int64 `db:"from_system_id" json:"from_system_id"`
	ToID     int64 `db:"to_system_id" json:"to_system_id"`
	IsNatural bool `db:"is_natural" json:"is_natural"` // 天然星门(先驱者遗留) vs 玩家建造
}

type Planet struct {
	ID          int64   `db:"id" json:"id"`
	SystemID    int64   `db:"system_id" json:"system_id"`
	Name        string  `db:"name" json:"name"`
	PlanetType  string  `db:"planet_type" json:"planet_type"`
	OrbitAU     float64 `db:"orbit_au" json:"orbit_au"`
	MoonCount   int     `db:"moon_count" json:"moon_count"`
	HasStation  bool    `db:"has_station" json:"has_station"`
	PosX        float64 `db:"pos_x" json:"pos_x"`
	PosY        float64 `db:"pos_y" json:"pos_y"`
	PosZ        float64 `db:"pos_z" json:"pos_z"`
}

type AsteroidBelt struct {
	ID         int64   `db:"id" json:"id"`
	SystemID   int64   `db:"system_id" json:"system_id"`
	Name       string  `db:"name" json:"name"`
	BeltType   string  `db:"belt_type" json:"belt_type"`
	OrbitAU    float64 `db:"orbit_au" json:"orbit_au"`
	Remaining  int     `db:"remaining_pct" json:"remaining_pct"`
	PosX       float64 `db:"pos_x" json:"pos_x"`
	PosY       float64 `db:"pos_y" json:"pos_y"`
	PosZ       float64 `db:"pos_z" json:"pos_z"`
}

type Station struct {
	ID          int64   `db:"id" json:"id"`
	SystemID    int64   `db:"system_id" json:"system_id"`
	PlanetID    *int64  `db:"planet_id" json:"planet_id,omitempty"`
	Name        string  `db:"name" json:"name"`
	OwnerType   string  `db:"owner_type" json:"owner_type"`
	OwnerID     *int64  `db:"owner_id" json:"owner_id,omitempty"`
	PosX        float64 `db:"pos_x" json:"pos_x"`
	PosY        float64 `db:"pos_y" json:"pos_y"`
	PosZ        float64 `db:"pos_z" json:"pos_z"`
	HasMarket   bool    `db:"has_market" json:"has_market"`
	HasRefinery bool    `db:"has_refinery" json:"has_refinery"`
	HasFactory  bool    `db:"has_factory" json:"has_factory"`
	HasCloneBay bool    `db:"has_clone_bay" json:"has_clone_bay"`
	HasRepair   bool    `db:"has_repair" json:"has_repair"`
	DockingFee  int64   `db:"docking_fee" json:"docking_fee"`
}

type SecurityZone string

const (
	ZoneHighSec SecurityZone = "high"  // 0.5 ~ 1.0
	ZoneLowSec  SecurityZone = "low"   // 0.1 ~ 0.4
	ZoneNullSec SecurityZone = "null"  // 0.0
	ZoneAbyss   SecurityZone = "abyss" // < 0
)

func (s *StarSystem) SecurityZone() SecurityZone {
	if s.SecurityLevel >= 0.5 {
		return ZoneHighSec
	}
	if s.SecurityLevel > 0 {
		return ZoneLowSec
	}
	if s.SecurityLevel == 0 {
		return ZoneNullSec
	}
	return ZoneAbyss
}
