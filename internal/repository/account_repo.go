package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/starfall-warsong/sws/internal/model"
)

type AccountRepo struct {
	db *sqlx.DB
}

func NewAccountRepo(db *sqlx.DB) *AccountRepo {
	return &AccountRepo{db: db}
}

func (r *AccountRepo) Create(ctx context.Context, phone, passwordHash string) (*model.Account, error) {
	var account model.Account
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO accounts (phone, password_hash, max_slots)
		 VALUES ($1, $2, $3)
		 RETURNING id, phone, password_hash, max_slots, created_at, updated_at`,
		phone, passwordHash, model.DefaultFreeSlots,
	).StructScan(&account)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepo) GetByPhone(ctx context.Context, phone string) (*model.Account, error) {
	var account model.Account
	err := r.db.GetContext(ctx, &account, `SELECT * FROM accounts WHERE phone = $1`, phone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &account, err
}

func (r *AccountRepo) GetByID(ctx context.Context, id int64) (*model.Account, error) {
	var account model.Account
	err := r.db.GetContext(ctx, &account, `SELECT * FROM accounts WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &account, err
}

func (r *AccountRepo) UpdateMaxSlots(ctx context.Context, id int64, maxSlots int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE accounts SET max_slots = $1, updated_at = NOW() WHERE id = $2`,
		maxSlots, id,
	)
	return err
}
