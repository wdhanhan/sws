package service

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	ErrCorpNameExists    = errors.New("军团名称已存在")
	ErrTickerExists      = errors.New("军团代号已存在")
	ErrAlreadyInCorp     = errors.New("角色已在一个军团中")
	ErrNotInCorp         = errors.New("角色不在军团中")
	ErrNotCEO            = errors.New("只有团长可以执行此操作")
	ErrCorpNotFound      = errors.New("军团不存在")
	ErrCannotLeaveCEO    = errors.New("团长不能离开军团，请先转让团长")
	ErrAllianceNotFound  = errors.New("联盟不存在")
	ErrNationNotFound    = errors.New("国家不存在")
)

type CorpService struct {
	db *sqlx.DB
}

func NewCorpService(db *sqlx.DB) *CorpService {
	return &CorpService{db: db}
}

func (s *CorpService) DB() *sqlx.DB {
	return s.db
}

type CreateCorpRequest struct {
	Name        string  `json:"name" binding:"required"`
	Ticker      string  `json:"ticker" binding:"required"`
	Description string  `json:"description"`
	TaxRate     float64 `json:"tax_rate"`
}

type CorpInfo struct {
	ID           int64   `db:"id" json:"id"`
	Name         string  `db:"name" json:"name"`
	Ticker       string  `db:"ticker" json:"ticker"`
	Description  string  `db:"description" json:"description"`
	CEOCharID    int64   `db:"ceo_character_id" json:"ceo_character_id"`
	MemberCount  int     `db:"member_count" json:"member_count"`
	TaxRate      float64 `db:"tax_rate" json:"tax_rate"`
	Balance      int64   `db:"balance" json:"balance"`
	HomeSystemID *int64  `db:"home_system_id" json:"home_system_id,omitempty"`
	CreatedAt    string  `db:"created_at" json:"created_at"`
}

type CorpMemberInfo struct {
	CharacterID int64  `db:"character_id" json:"character_id"`
	Name        string `json:"name"`
	Role        string `db:"role" json:"role"`
	JoinedAt    string `db:"joined_at" json:"joined_at"`
}

func (s *CorpService) Create(ctx context.Context, charID int64, req *CreateCorpRequest) (*CorpInfo, error) {
	// Check not already in a corp
	var inCorp int
	s.db.GetContext(ctx, &inCorp, `SELECT COUNT(*) FROM corp_members WHERE character_id = $1`, charID)
	if inCorp > 0 {
		return nil, ErrAlreadyInCorp
	}

	if req.TaxRate < 0 || req.TaxRate > 0.5 {
		req.TaxRate = 0.05
	}

	var corp CorpInfo
	err := s.db.QueryRowxContext(ctx,
		`INSERT INTO corporations (name, ticker, description, ceo_character_id, tax_rate)
		 VALUES ($1, $2, $3, $4, $5) RETURNING *`,
		req.Name, req.Ticker, req.Description, charID, req.TaxRate,
	).StructScan(&corp)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrCorpNameExists
		}
		return nil, err
	}

	// Add creator as CEO
	s.db.ExecContext(ctx,
		`INSERT INTO corp_members (corp_id, character_id, role) VALUES ($1, $2, 'ceo')`,
		corp.ID, charID)

	return &corp, nil
}

func (s *CorpService) Join(ctx context.Context, charID, corpID int64) error {
	var inCorp int
	s.db.GetContext(ctx, &inCorp, `SELECT COUNT(*) FROM corp_members WHERE character_id = $1`, charID)
	if inCorp > 0 {
		return ErrAlreadyInCorp
	}

	var exists int
	s.db.GetContext(ctx, &exists, `SELECT COUNT(*) FROM corporations WHERE id = $1`, corpID)
	if exists == 0 {
		return ErrCorpNotFound
	}

	s.db.ExecContext(ctx,
		`INSERT INTO corp_members (corp_id, character_id, role) VALUES ($1, $2, 'member')`,
		corpID, charID)
	s.db.ExecContext(ctx,
		`UPDATE corporations SET member_count = member_count + 1 WHERE id = $1`, corpID)

	return nil
}

func (s *CorpService) Leave(ctx context.Context, charID int64) error {
	var member struct {
		CorpID int64  `db:"corp_id"`
		Role   string `db:"role"`
	}
	err := s.db.GetContext(ctx, &member,
		`SELECT corp_id, role FROM corp_members WHERE character_id = $1`, charID)
	if err != nil {
		return ErrNotInCorp
	}
	if member.Role == "ceo" {
		return ErrCannotLeaveCEO
	}

	s.db.ExecContext(ctx, `DELETE FROM corp_members WHERE character_id = $1`, charID)
	s.db.ExecContext(ctx,
		`UPDATE corporations SET member_count = member_count - 1 WHERE id = $1`, member.CorpID)

	return nil
}

func (s *CorpService) GetInfo(ctx context.Context, corpID int64) (*CorpInfo, error) {
	var corp CorpInfo
	err := s.db.GetContext(ctx, &corp, `SELECT * FROM corporations WHERE id = $1`, corpID)
	if err != nil {
		return nil, ErrCorpNotFound
	}
	return &corp, nil
}

func (s *CorpService) GetMembers(ctx context.Context, corpID int64) ([]CorpMemberInfo, error) {
	var members []CorpMemberInfo
	err := s.db.SelectContext(ctx, &members,
		`SELECT cm.character_id, cm.role, cm.joined_at FROM corp_members cm
		 WHERE cm.corp_id = $1 ORDER BY cm.role, cm.joined_at`, corpID)
	if err != nil {
		return nil, err
	}
	for i := range members {
		var name string
		s.db.GetContext(ctx, &name, `SELECT name FROM characters WHERE id = $1`, members[i].CharacterID)
		members[i].Name = name
	}
	return members, nil
}

func (s *CorpService) GetCharCorp(ctx context.Context, charID int64) (*CorpInfo, string, error) {
	var member struct {
		CorpID int64  `db:"corp_id"`
		Role   string `db:"role"`
	}
	err := s.db.GetContext(ctx, &member,
		`SELECT corp_id, role FROM corp_members WHERE character_id = $1`, charID)
	if err != nil {
		return nil, "", ErrNotInCorp
	}
	corp, err := s.GetInfo(ctx, member.CorpID)
	return corp, member.Role, err
}

// CreateAlliance creates an alliance from the CEO's corp
func (s *CorpService) CreateAlliance(ctx context.Context, charID int64, name, ticker string) (int64, error) {
	corp, role, err := s.GetCharCorp(ctx, charID)
	if err != nil {
		return 0, err
	}
	if role != "ceo" {
		return 0, ErrNotCEO
	}

	var allianceID int64
	err = s.db.QueryRowContext(ctx,
		`INSERT INTO alliances (name, ticker, leader_corp_id) VALUES ($1, $2, $3) RETURNING id`,
		name, ticker, corp.ID).Scan(&allianceID)
	if err != nil {
		return 0, err
	}

	s.db.ExecContext(ctx,
		`INSERT INTO alliance_members (alliance_id, corp_id, role) VALUES ($1, $2, 'executor')`,
		allianceID, corp.ID)

	return allianceID, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (errors.Is(err, errors.New("unique")) ||
		containsStr(err.Error(), "unique") ||
		containsStr(err.Error(), "duplicate"))
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && findSubstr(s, sub))
}

func findSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
