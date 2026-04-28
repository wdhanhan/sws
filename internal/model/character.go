package model

import "time"

type RaceID int

const (
	RaceAries       RaceID = 1  // 白羊座
	RaceTaurus      RaceID = 2  // 金牛座
	RaceGemini      RaceID = 3  // 双子座
	RaceCancer      RaceID = 4  // 巨蟹座
	RaceLeo         RaceID = 5  // 狮子座
	RaceVirgo       RaceID = 6  // 处女座
	RaceLibra       RaceID = 7  // 天秤座
	RaceScorpio     RaceID = 8  // 天蝎座
	RaceSagittarius RaceID = 9  // 射手座
	RaceCapricorn   RaceID = 10 // 摩羯座
	RaceAquarius    RaceID = 11 // 水瓶座
	RacePisces      RaceID = 12 // 双鱼座
)

var RaceNames = map[RaceID]string{
	RaceAries:       "白羊座·冲锋者",
	RaceTaurus:      "金牛座·铸造者",
	RaceGemini:      "双子座·幻影者",
	RaceCancer:      "巨蟹座·守护者",
	RaceLeo:         "狮子座·统御者",
	RaceVirgo:       "处女座·精工者",
	RaceLibra:       "天秤座·裁量者",
	RaceScorpio:     "天蝎座·蚀刻者",
	RaceSagittarius: "射手座·游猎者",
	RaceCapricorn:   "摩羯座·筑垒者",
	RaceAquarius:    "水瓶座·革新者",
	RacePisces:      "双鱼座·共生者",
}

type Character struct {
	ID                   int64     `db:"id" json:"id"`
	AccountID            int64     `db:"account_id" json:"account_id"`
	Name                 string    `db:"name" json:"name"`
	RaceID               RaceID    `db:"race_id" json:"race_id"`
	CurrentSystemID      int64     `db:"current_system_id" json:"current_system_id"`
	IsDocked             bool      `db:"is_docked" json:"is_docked"`
	DockedStationID      *int64    `db:"docked_station_id" json:"docked_station_id,omitempty"`
	PosX                 float64   `db:"pos_x" json:"pos_x"` // 星系内3D坐标(米)
	PosY                 float64   `db:"pos_y" json:"pos_y"`
	PosZ                 float64   `db:"pos_z" json:"pos_z"`
	Balance              int64     `db:"balance" json:"balance"`
	FatiguePoints        int       `db:"fatigue_points" json:"fatigue_points"`
	ConsciousnessPercent int       `db:"consciousness_pct" json:"consciousness_pct"`
	CreatedAt            time.Time `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time `db:"updated_at" json:"updated_at"`
}

const (
	MaxFatiguePoints        = 480 // 8小时
	MaxConsciousnessPercent = 100
)
