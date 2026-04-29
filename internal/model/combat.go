package model

type DamageType string

const (
	DamageKinetic   DamageType = "kinetic"
	DamageThermal   DamageType = "thermal"
	DamageEM        DamageType = "em"
	DamageExplosive DamageType = "explosive"
)

type ShipDef struct {
	ID             int64   `db:"id" json:"id"`
	Name           string  `db:"name" json:"name"`
	RaceID         int     `db:"race_id" json:"race_id"`
	Tier           int     `db:"tier" json:"tier"`
	ShipClass      string  `db:"ship_class" json:"ship_class"`
	ShipRole       string  `db:"ship_role" json:"ship_role"`
	ShieldHP       int     `db:"shield_hp" json:"shield_hp"`
	ArmorHP        int     `db:"armor_hp" json:"armor_hp"`
	StructureHP    int     `db:"structure_hp" json:"structure_hp"`
	ShieldRecharge int     `db:"shield_recharge" json:"shield_recharge"`
	Capacitor      int     `db:"capacitor" json:"capacitor"`
	CapRecharge    int     `db:"cap_recharge" json:"cap_recharge"`
	MaxSpeed       int     `db:"max_speed" json:"max_speed"`         // 亚光速最大速度(m/s)
	WarpSpeed      float64 `db:"warp_speed" json:"warp_speed"`       // 跃迁速度(AU/s)
	WarpCapCost    int     `db:"warp_cap_cost" json:"warp_cap_cost"` // 每次跃迁电容消耗
	AlignTicks     int     `db:"align_ticks" json:"align_ticks"`     // 对齐Tick数(越小越快进入跃迁)
	JumpRange      float64 `db:"jump_range" json:"jump_range"`       // 跳跃引擎范围(光年)，0=无
	Mass           int64   `db:"mass" json:"mass"`                   // 质量(kg)
	Signature      int     `db:"signature" json:"signature"`
	HighSlots      int     `db:"high_slots" json:"high_slots"`
	MidSlots       int     `db:"mid_slots" json:"mid_slots"`
	LowSlots       int     `db:"low_slots" json:"low_slots"`
	CargoM3        float64 `db:"cargo_m3" json:"cargo_m3"`
	DroneBayM3     float64 `db:"drone_bay_m3" json:"drone_bay_m3"`
	Powergrid      int     `db:"powergrid" json:"powergrid"`
	CPU            int     `db:"cpu" json:"cpu"`
}

type NPCDef struct {
	ID             int64      `db:"id" json:"id"`
	Name           string     `db:"name" json:"name"`
	NPCType        string     `db:"npc_type" json:"npc_type"`
	Tier           int        `db:"tier" json:"tier"`
	ShieldHP       int        `db:"shield_hp" json:"shield_hp"`
	ArmorHP        int        `db:"armor_hp" json:"armor_hp"`
	StructureHP    int        `db:"structure_hp" json:"structure_hp"`
	ShieldRecharge int        `db:"shield_recharge" json:"shield_recharge"`
	DamagePerTick  int        `db:"damage_per_tick" json:"damage_per_tick"`
	DamageType     DamageType `db:"damage_type" json:"damage_type"`
	OptimalRange   int        `db:"optimal_range" json:"optimal_range"`
	Speed          int        `db:"speed" json:"speed"`
	Signature      int        `db:"signature" json:"signature"`
	Bounty         int64      `db:"bounty" json:"bounty"`
	AIBehavior     string     `db:"ai_behavior" json:"ai_behavior"`
}

type ResistProfile struct {
	Kinetic   float64 `json:"kinetic"`
	Thermal   float64 `json:"thermal"`
	EM        float64 `json:"em"`
	Explosive float64 `json:"explosive"`
}

type CombatParticipant struct {
	ID               int64      `json:"id"`
	Name             string     `json:"name"`
	Type             string     `json:"type"` // player / npc
	Team             string     `json:"team"`
	ShieldCurrent    int        `json:"shield_current"`
	ShieldMax        int        `json:"shield_max"`
	ArmorCurrent     int        `json:"armor_current"`
	ArmorMax         int        `json:"armor_max"`
	StructureCurrent int        `json:"structure_current"`
	StructureMax     int        `json:"structure_max"`
	CapCurrent       int        `json:"cap_current"`
	CapMax           int        `json:"cap_max"`
	CapRecharge      int        `json:"cap_recharge"`
	Distance         int        `json:"distance"`
	IsDestroyed      bool       `json:"is_destroyed"`
	TargetID         *int64     `json:"target_id,omitempty"`
	DamagePerTick    int        `json:"damage_per_tick"`
	DamageType       DamageType `json:"damage_type"`
	RateOfFire       int        `json:"rate_of_fire"`
	WeaponName       string     `json:"weapon_name,omitempty"`
	TrackingSpeed    float64    `json:"tracking_speed"`
	FalloffRange     int        `json:"falloff_range"`
	CapCost          int        `json:"cap_cost"`
	ShieldRecharge   int        `json:"shield_recharge"`
	ArmorRepair      int        `json:"armor_repair"`
	Speed            int        `json:"speed"`
	Signature        int        `json:"signature"`
	OptimalRange     int        `json:"optimal_range"`
	ShieldResist     ResistProfile `json:"shield_resist"`
	ArmorResist      ResistProfile `json:"armor_resist"`
}

type CombatState struct {
	CombatID     int64                `json:"combat_id"`
	Tick         int                  `json:"tick"`
	Status       string               `json:"status"`
	Participants []CombatParticipant  `json:"participants"`
	Logs         []string             `json:"logs"`
}
