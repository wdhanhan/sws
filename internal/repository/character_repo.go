package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/starfall-warsong/sws/internal/model"
)

type CharacterRepo struct {
	db *sqlx.DB
}

func NewCharacterRepo(db *sqlx.DB) *CharacterRepo {
	return &CharacterRepo{db: db}
}

func (r *CharacterRepo) DB() *sqlx.DB { return r.db }

func (r *CharacterRepo) Create(ctx context.Context, c *model.Character) (*model.Character, error) {
	var char model.Character
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO characters (account_id, name, race_id, current_system_id, is_docked, docked_station_id, balance, fatigue_points, consciousness_pct)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING *`,
		c.AccountID, c.Name, c.RaceID, c.CurrentSystemID, c.IsDocked, c.DockedStationID,
		c.Balance, c.FatiguePoints, c.ConsciousnessPercent,
	).StructScan(&char)
	if err != nil {
		return nil, err
	}
	return &char, nil
}

func (r *CharacterRepo) GetByID(ctx context.Context, id int64) (*model.Character, error) {
	var c model.Character
	err := r.db.GetContext(ctx, &c, `SELECT * FROM characters WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &c, err
}

func (r *CharacterRepo) ListByAccount(ctx context.Context, accountID int64) ([]model.Character, error) {
	var chars []model.Character
	err := r.db.SelectContext(ctx, &chars, `SELECT * FROM characters WHERE account_id = $1 ORDER BY created_at`, accountID)
	return chars, err
}

func (r *CharacterRepo) CountByAccount(ctx context.Context, accountID int64) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM characters WHERE account_id = $1`, accountID)
	return count, err
}

func (r *CharacterRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM characters WHERE id = $1`, id)
	return err
}

func (r *CharacterRepo) NameExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM characters WHERE name = $1)`, name)
	return exists, err
}
