package service

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrAlreadyInFleet  = errors.New("你已在舰队中")
	ErrFleetNotFound   = errors.New("舰队不存在")
	ErrNotFleetLeader  = errors.New("你不是队长")
	ErrNotInvited      = errors.New("没有收到该舰队的邀请")
	ErrNotInFleet      = errors.New("你不在任何舰队中")
	ErrCannotKickSelf  = errors.New("不能踢出自己")
	ErrTargetNotInFleet = errors.New("目标不在舰队中")
)

type FleetService struct {
	db *sqlx.DB
}

func NewFleetService(db *sqlx.DB) *FleetService {
	return &FleetService{db: db}
}

type FleetInfo struct {
	FleetID   int64         `json:"fleet_id"`
	LeaderID  int64         `json:"leader_id"`
	LeaderName string       `json:"leader_name"`
	Status    string        `json:"status"`
	Members   []FleetMember `json:"members"`
	CreatedAt time.Time     `json:"created_at"`
}

type FleetMember struct {
	CharacterID int64  `db:"character_id" json:"character_id"`
	Name        string `db:"name" json:"name"`
	Status      string `db:"status" json:"status"`
	ShipName    string `json:"ship_name"`
}

func (s *FleetService) Create(ctx context.Context, charID int64) (*FleetInfo, error) {
	if _, err := s.getActiveFleetID(ctx, charID); err == nil {
		return nil, ErrAlreadyInFleet
	}

	tx, _ := s.db.BeginTxx(ctx, nil)
	defer tx.Rollback()

	var fleetID int64
	err := tx.QueryRowContext(ctx,
		`INSERT INTO fleets (leader_id) VALUES ($1) RETURNING id`, charID).Scan(&fleetID)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO fleet_members (fleet_id, character_id, status) VALUES ($1, $2, 'joined')`,
		fleetID, charID)
	if err != nil {
		return nil, err
	}

	tx.Commit()
	return s.GetFleet(ctx, charID)
}

func (s *FleetService) Invite(ctx context.Context, leaderID, targetCharID int64) error {
	fleetID, err := s.getActiveFleetID(ctx, leaderID)
	if err != nil {
		return ErrNotInFleet
	}

	var actualLeader int64
	s.db.GetContext(ctx, &actualLeader, `SELECT leader_id FROM fleets WHERE id=$1`, fleetID)
	if actualLeader != leaderID {
		return ErrNotFleetLeader
	}

	if fid, _ := s.getActiveFleetID(ctx, targetCharID); fid > 0 {
		return errors.New("目标角色已在舰队中")
	}

	// Same account = auto join, no need to accept
	var leaderAccountID, targetAccountID int64
	s.db.GetContext(ctx, &leaderAccountID, `SELECT account_id FROM characters WHERE id=$1`, leaderID)
	s.db.GetContext(ctx, &targetAccountID, `SELECT account_id FROM characters WHERE id=$1`, targetCharID)

	status := "pending"
	if leaderAccountID > 0 && leaderAccountID == targetAccountID {
		status = "joined"
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO fleet_members (fleet_id, character_id, status) VALUES ($1, $2, $3)
		 ON CONFLICT (fleet_id, character_id) DO UPDATE SET status=$3`,
		fleetID, targetCharID, status)
	return err
}

func (s *FleetService) Accept(ctx context.Context, charID, fleetID int64) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE fleet_members SET status='joined', joined_at=NOW()
		 WHERE fleet_id=$1 AND character_id=$2 AND status='pending'`,
		fleetID, charID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotInvited
	}
	return nil
}

func (s *FleetService) Decline(ctx context.Context, charID, fleetID int64) error {
	s.db.ExecContext(ctx,
		`DELETE FROM fleet_members WHERE fleet_id=$1 AND character_id=$2 AND status='pending'`,
		fleetID, charID)
	return nil
}

func (s *FleetService) Leave(ctx context.Context, charID int64) error {
	fleetID, err := s.getActiveFleetID(ctx, charID)
	if err != nil {
		return ErrNotInFleet
	}

	var leaderID int64
	s.db.GetContext(ctx, &leaderID, `SELECT leader_id FROM fleets WHERE id=$1`, fleetID)

	if leaderID == charID {
		return s.Disband(ctx, charID)
	}

	s.db.ExecContext(ctx,
		`UPDATE fleet_members SET status='left' WHERE fleet_id=$1 AND character_id=$2`,
		fleetID, charID)
	return nil
}

func (s *FleetService) Kick(ctx context.Context, leaderID, targetCharID int64) error {
	if leaderID == targetCharID {
		return ErrCannotKickSelf
	}

	fleetID, err := s.getActiveFleetID(ctx, leaderID)
	if err != nil {
		return ErrNotInFleet
	}

	var actualLeader int64
	s.db.GetContext(ctx, &actualLeader, `SELECT leader_id FROM fleets WHERE id=$1`, fleetID)
	if actualLeader != leaderID {
		return ErrNotFleetLeader
	}

	res, _ := s.db.ExecContext(ctx,
		`UPDATE fleet_members SET status='left' WHERE fleet_id=$1 AND character_id=$2 AND status='joined'`,
		fleetID, targetCharID)
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrTargetNotInFleet
	}
	return nil
}

func (s *FleetService) Disband(ctx context.Context, leaderID int64) error {
	fleetID, err := s.getActiveFleetID(ctx, leaderID)
	if err != nil {
		return ErrNotInFleet
	}

	s.db.ExecContext(ctx, `UPDATE fleets SET status='disbanded' WHERE id=$1`, fleetID)
	s.db.ExecContext(ctx, `UPDATE fleet_members SET status='left' WHERE fleet_id=$1`, fleetID)
	return nil
}

func (s *FleetService) GetFleet(ctx context.Context, charID int64) (*FleetInfo, error) {
	fleetID, err := s.getActiveFleetID(ctx, charID)
	if err != nil {
		return nil, ErrNotInFleet
	}

	var fleet struct {
		ID        int64     `db:"id"`
		LeaderID  int64     `db:"leader_id"`
		Status    string    `db:"status"`
		CreatedAt time.Time `db:"created_at"`
	}
	if err := s.db.GetContext(ctx, &fleet, `SELECT id, leader_id, status, created_at FROM fleets WHERE id=$1`, fleetID); err != nil {
		return nil, ErrFleetNotFound
	}

	var leaderName string
	s.db.GetContext(ctx, &leaderName, `SELECT name FROM characters WHERE id=$1`, fleet.LeaderID)

	var members []FleetMember
	s.db.SelectContext(ctx, &members,
		`SELECT fm.character_id, c.name, fm.status
		 FROM fleet_members fm JOIN characters c ON c.id = fm.character_id
		 WHERE fm.fleet_id=$1 AND fm.status IN ('joined','pending')
		 ORDER BY fm.status, fm.joined_at`, fleetID)

	for i := range members {
		var shipName string
		s.db.GetContext(ctx, &shipName,
			`SELECT sd.name FROM ships sh JOIN ship_defs sd ON sd.id=sh.ship_def_id
			 WHERE sh.character_id=$1 AND sh.is_active=true LIMIT 1`, members[i].CharacterID)
		members[i].ShipName = shipName
	}

	return &FleetInfo{
		FleetID:    fleet.ID,
		LeaderID:   fleet.LeaderID,
		LeaderName: leaderName,
		Status:     fleet.Status,
		Members:    members,
		CreatedAt:  fleet.CreatedAt,
	}, nil
}

func (s *FleetService) GetFleetMemberIDs(ctx context.Context, fleetID int64) ([]int64, error) {
	var ids []int64
	err := s.db.SelectContext(ctx, &ids,
		`SELECT character_id FROM fleet_members WHERE fleet_id=$1 AND status='joined'`, fleetID)
	return ids, err
}

func (s *FleetService) IsInFleet(ctx context.Context, charID int64) (bool, int64, int64) {
	// Check both joined and pending members
	var fleetID int64
	err := s.db.GetContext(ctx, &fleetID,
		`SELECT fm.fleet_id FROM fleet_members fm
		 JOIN fleets f ON f.id = fm.fleet_id
		 WHERE fm.character_id=$1 AND fm.status IN ('joined','pending') AND f.status='active'
		 LIMIT 1`, charID)
	if err != nil || fleetID == 0 {
		return false, 0, 0
	}
	var leaderID int64
	s.db.GetContext(ctx, &leaderID, `SELECT leader_id FROM fleets WHERE id=$1`, fleetID)
	return true, fleetID, leaderID
}

func (s *FleetService) GetPendingInvites(ctx context.Context, charID int64) ([]FleetInfo, error) {
	type PendingRow struct {
		FleetID  int64  `db:"fleet_id"`
		LeaderID int64  `db:"leader_id"`
		LeaderName string `db:"leader_name"`
	}
	var rows []PendingRow
	s.db.SelectContext(ctx, &rows,
		`SELECT fm.fleet_id, f.leader_id, c.name as leader_name
		 FROM fleet_members fm
		 JOIN fleets f ON f.id = fm.fleet_id
		 JOIN characters c ON c.id = f.leader_id
		 WHERE fm.character_id=$1 AND fm.status='pending' AND f.status='active'`, charID)

	var invites []FleetInfo
	for _, r := range rows {
		invites = append(invites, FleetInfo{
			FleetID:    r.FleetID,
			LeaderID:   r.LeaderID,
			LeaderName: r.LeaderName,
		})
	}
	return invites, nil
}

type SimpleChar struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

func (s *FleetService) GetSameAccountChars(ctx context.Context, charID int64) []SimpleChar {
	var chars []SimpleChar
	s.db.SelectContext(ctx, &chars,
		`SELECT c2.id, c2.name FROM characters c1
		 JOIN characters c2 ON c2.account_id = c1.account_id
		 WHERE c1.id = $1 AND c2.id != $1
		 ORDER BY c2.id`, charID)
	return chars
}

func (s *FleetService) getActiveFleetID(ctx context.Context, charID int64) (int64, error) {
	var fleetID int64
	err := s.db.GetContext(ctx, &fleetID,
		`SELECT fm.fleet_id FROM fleet_members fm
		 JOIN fleets f ON f.id = fm.fleet_id
		 WHERE fm.character_id=$1 AND fm.status='joined' AND f.status='active'
		 LIMIT 1`, charID)
	return fleetID, err
}
