package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/starfall-warsong/sws/internal/model"
)

type StarmapRepo struct {
	db *sqlx.DB
}

func NewStarmapRepo(db *sqlx.DB) *StarmapRepo {
	return &StarmapRepo{db: db}
}

func (r *StarmapRepo) DB() *sqlx.DB { return r.db }

func (r *StarmapRepo) BulkInsertSystems(ctx context.Context, systems []model.StarSystem) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO star_systems (id, name, arm_id, coord_x, coord_y, coord_z, security_level, star_type, planet_count, belt_count, has_anomaly, has_ruins, seed)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		 ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range systems {
		_, err := stmt.ExecContext(ctx, s.ID, s.Name, s.ArmID, s.CoordX, s.CoordY, s.CoordZ,
			s.SecurityLevel, s.StarType, s.PlanetCount, s.BeltCount, s.HasAnomaly, s.HasRuins, s.Seed)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *StarmapRepo) BulkInsertGates(ctx context.Context, gates []model.Stargate) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO stargates (from_system_id, to_system_id, is_natural) VALUES ($1,$2,$3)
		 ON CONFLICT DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, g := range gates {
		_, err := stmt.ExecContext(ctx, g.FromID, g.ToID, g.IsNatural)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *StarmapRepo) BulkInsertPlanets(ctx context.Context, planets []model.Planet) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO planets (system_id, name, planet_type, orbit_au, moon_count, has_station) VALUES ($1,$2,$3,$4,$5,$6)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range planets {
		_, err := stmt.ExecContext(ctx, p.SystemID, p.Name, p.PlanetType, p.OrbitAU, p.MoonCount, p.HasStation)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *StarmapRepo) BulkInsertBelts(ctx context.Context, belts []model.AsteroidBelt) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO asteroid_belts (system_id, name, belt_type, orbit_au, remaining_pct) VALUES ($1,$2,$3,$4,$5)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, b := range belts {
		_, err := stmt.ExecContext(ctx, b.SystemID, b.Name, b.BeltType, b.OrbitAU, b.Remaining)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *StarmapRepo) GetSystem(ctx context.Context, id int64) (*model.StarSystem, error) {
	var s model.StarSystem
	err := r.db.GetContext(ctx, &s, `SELECT * FROM star_systems WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *StarmapRepo) GetAdjacentSystems(ctx context.Context, systemID int64) ([]model.StarSystem, error) {
	var systems []model.StarSystem
	err := r.db.SelectContext(ctx, &systems,
		`SELECT s.* FROM star_systems s
		 JOIN stargates g ON g.to_system_id = s.id
		 WHERE g.from_system_id = $1`, systemID)
	return systems, err
}

func (r *StarmapRepo) GetSystemPlanets(ctx context.Context, systemID int64) ([]model.Planet, error) {
	var planets []model.Planet
	err := r.db.SelectContext(ctx, &planets,
		`SELECT * FROM planets WHERE system_id = $1 ORDER BY orbit_au`, systemID)
	return planets, err
}

func (r *StarmapRepo) GetSystemBelts(ctx context.Context, systemID int64) ([]model.AsteroidBelt, error) {
	var belts []model.AsteroidBelt
	err := r.db.SelectContext(ctx, &belts,
		`SELECT * FROM asteroid_belts WHERE system_id = $1 ORDER BY orbit_au`, systemID)
	return belts, err
}

func (r *StarmapRepo) GetSystemCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM star_systems`)
	return count, err
}

func (r *StarmapRepo) SearchSystems(ctx context.Context, armID *model.ArmID, secMin, secMax *float64, limit int) ([]model.StarSystem, error) {
	query := `SELECT * FROM star_systems WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if armID != nil {
		query += fmt.Sprintf(` AND arm_id = $%d`, argIdx)
		args = append(args, *armID)
		argIdx++
	}
	if secMin != nil {
		query += fmt.Sprintf(` AND security_level >= $%d`, argIdx)
		args = append(args, *secMin)
		argIdx++
	}
	if secMax != nil {
		query += fmt.Sprintf(` AND security_level <= $%d`, argIdx)
		args = append(args, *secMax)
		argIdx++
	}

	query += fmt.Sprintf(` ORDER BY id LIMIT $%d`, argIdx)
	args = append(args, limit)

	var systems []model.StarSystem
	err := r.db.SelectContext(ctx, &systems, query, args...)
	return systems, err
}
