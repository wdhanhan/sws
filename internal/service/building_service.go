package service

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	ErrBuildingDefNotFound = errors.New("建筑类型不存在")
	ErrBuildingNotFound    = errors.New("建筑不存在")
)

type BuildingService struct {
	db *sqlx.DB
}

func NewBuildingService(db *sqlx.DB) *BuildingService {
	return &BuildingService{db: db}
}

type BuildingDef struct {
	ID            int64   `db:"id" json:"id"`
	Name          string  `db:"name" json:"name"`
	Category      string  `db:"category" json:"category"`
	BuildingType  string  `db:"building_type" json:"building_type"`
	Description   string  `db:"description" json:"description"`
	BuildTimeHrs  int     `db:"build_time_hours" json:"build_time_hours"`
	FuelPerHour   int     `db:"fuel_per_hour" json:"fuel_per_hour"`
	ShieldHP      int     `db:"shield_hp" json:"shield_hp"`
	ArmorHP       int     `db:"armor_hp" json:"armor_hp"`
	StructureHP   int     `db:"structure_hp" json:"structure_hp"`
	HasDefense    bool    `db:"has_defense" json:"has_defense"`
	DefenseDPS    int     `db:"defense_dps" json:"defense_dps"`
	CargoCapacity float64 `db:"cargo_capacity" json:"cargo_capacity"`
}

type BuildingInfo struct {
	ID            int64   `db:"id" json:"id"`
	DefID         int64   `db:"building_def_id" json:"building_def_id"`
	OwnerName     string  `db:"owner_name" json:"owner_name"`
	Name          string  `db:"name" json:"name"`
	SystemID      int64   `db:"system_id" json:"system_id"`
	Status        string  `db:"status" json:"status"`
	BuildProgress int     `db:"build_progress" json:"build_progress"`
	ShieldCurrent int     `db:"shield_current" json:"shield_current"`
	ArmorCurrent  int     `db:"armor_current" json:"armor_current"`
	StructCurrent int     `db:"structure_current" json:"structure_current"`
	FuelRemaining int     `db:"fuel_remaining" json:"fuel_remaining"`
	IsPowered     bool    `db:"is_powered" json:"is_powered"`
}

func (s *BuildingService) GetBuildingDefs(ctx context.Context, category string) ([]BuildingDef, error) {
	var defs []BuildingDef
	if category != "" {
		return defs, s.db.SelectContext(ctx, &defs, `SELECT * FROM building_defs WHERE category = $1 ORDER BY id`, category)
	}
	return defs, s.db.SelectContext(ctx, &defs, `SELECT * FROM building_defs ORDER BY category, id`)
}

type DeployRequest struct {
	BuildingDefID int64   `json:"building_def_id" binding:"required"`
	Name          string  `json:"name" binding:"required"`
	SystemID      int64   `json:"system_id" binding:"required"`
	PosX          float64 `json:"pos_x"`
	PosY          float64 `json:"pos_y"`
	PosZ          float64 `json:"pos_z"`
}

func (s *BuildingService) Deploy(ctx context.Context, ownerType string, ownerID int64, ownerName string, req *DeployRequest) (*BuildingInfo, error) {
	var def BuildingDef
	err := s.db.GetContext(ctx, &def, `SELECT * FROM building_defs WHERE id = $1`, req.BuildingDefID)
	if err != nil {
		return nil, ErrBuildingDefNotFound
	}

	var building BuildingInfo
	err = s.db.QueryRowxContext(ctx,
		`INSERT INTO buildings (building_def_id, owner_type, owner_id, owner_name, name, system_id, pos_x, pos_y, pos_z,
		  status, shield_current, armor_current, structure_current, fuel_remaining)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'anchoring',$10,$11,$12,0) RETURNING *`,
		req.BuildingDefID, ownerType, ownerID, ownerName, req.Name, req.SystemID,
		req.PosX, req.PosY, req.PosZ,
		def.ShieldHP, def.ArmorHP, def.StructureHP,
	).StructScan(&building)

	return &building, err
}

func (s *BuildingService) GetSystemBuildings(ctx context.Context, systemID int64) ([]BuildingInfo, error) {
	var buildings []BuildingInfo
	err := s.db.SelectContext(ctx, &buildings,
		`SELECT * FROM buildings WHERE system_id = $1 AND status != 'destroyed' ORDER BY created_at`, systemID)
	return buildings, err
}
