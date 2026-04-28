package model

import "time"

type Account struct {
	ID           int64     `db:"id" json:"id"`
	Phone        string    `db:"phone" json:"phone"`
	PasswordHash string    `db:"password_hash" json:"-"`
	MaxSlots     int       `db:"max_slots" json:"max_slots"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

const DefaultFreeSlots = 3
