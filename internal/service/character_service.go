package service

import (
	"context"
	"errors"
	"unicode/utf8"

	"github.com/starfall-warsong/sws/internal/model"
	"github.com/starfall-warsong/sws/internal/repository"
)

var (
	ErrSlotsFull     = errors.New("角色位已满，需要购买额外角色位")
	ErrNameTooShort  = errors.New("角色名至少2个字符")
	ErrNameTooLong   = errors.New("角色名最多12个字符")
	ErrNameExists    = errors.New("角色名已被使用")
	ErrInvalidRace   = errors.New("无效的种族选择")
	ErrCharNotFound  = errors.New("角色不存在")
	ErrNotOwner      = errors.New("你没有权限操作此角色")
)

// 12种族起源星系ID（后续由星图生成器填充真实值）
var raceOriginSystems = map[model.RaceID]int64{
	model.RaceAries:       1,
	model.RaceTaurus:      2,
	model.RaceGemini:      3,
	model.RaceCancer:      4,
	model.RaceLeo:         5,
	model.RaceVirgo:       6,
	model.RaceLibra:       7,
	model.RaceScorpio:     8,
	model.RaceSagittarius: 9,
	model.RaceCapricorn:   10,
	model.RaceAquarius:    11,
	model.RacePisces:      12,
}

type CharacterService struct {
	charRepo    *repository.CharacterRepo
	accountRepo *repository.AccountRepo
}

func NewCharacterService(charRepo *repository.CharacterRepo, accountRepo *repository.AccountRepo) *CharacterService {
	return &CharacterService{charRepo: charRepo, accountRepo: accountRepo}
}

type CreateCharacterRequest struct {
	Name   string       `json:"name" binding:"required"`
	RaceID model.RaceID `json:"race_id" binding:"required"`
}

func (s *CharacterService) Create(ctx context.Context, accountID int64, req *CreateCharacterRequest) (*model.Character, error) {
	nameLen := utf8.RuneCountInString(req.Name)
	if nameLen < 2 {
		return nil, ErrNameTooShort
	}
	if nameLen > 12 {
		return nil, ErrNameTooLong
	}

	if _, ok := model.RaceNames[req.RaceID]; !ok {
		return nil, ErrInvalidRace
	}

	exists, err := s.charRepo.NameExists(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrNameExists
	}

	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	count, err := s.charRepo.CountByAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if count >= account.MaxSlots {
		return nil, ErrSlotsFull
	}

	originSystem := raceOriginSystems[req.RaceID]

	char := &model.Character{
		AccountID:            accountID,
		Name:                 req.Name,
		RaceID:               req.RaceID,
		CurrentSystemID:      originSystem,
		IsDocked:             true,
		Balance:              5000000, // 初始50000星币(单位:分)
		FatiguePoints:        model.MaxFatiguePoints,
		ConsciousnessPercent: model.MaxConsciousnessPercent,
	}

	created, err := s.charRepo.Create(ctx, char)
	if err != nil {
		return nil, err
	}

	// 自动赠送种族T1突击护卫舰
	raceShipIDs := map[model.RaceID]int64{
		model.RaceAries: 101, model.RaceTaurus: 201, model.RaceGemini: 301,
		model.RaceCancer: 401, model.RaceLeo: 501, model.RaceVirgo: 601,
		model.RaceLibra: 701, model.RaceScorpio: 801, model.RaceSagittarius: 901,
		model.RaceCapricorn: 1001, model.RaceAquarius: 1101, model.RacePisces: 1201,
	}
	if shipDefID, ok := raceShipIDs[req.RaceID]; ok {
		type DefHP struct {
			ShieldHP int `db:"shield_hp"`
			ArmorHP  int `db:"armor_hp"`
			StructHP int `db:"structure_hp"`
			Cap      int `db:"capacitor"`
		}
		var hp DefHP
		s.charRepo.DB().GetContext(ctx, &hp,
			`SELECT shield_hp, armor_hp, structure_hp, capacitor FROM ship_defs WHERE id = $1`, shipDefID)
		if hp.ShieldHP > 0 {
			s.charRepo.DB().ExecContext(ctx,
				`INSERT INTO ships (character_id, ship_def_id, name, shield_current, armor_current, structure_current, cap_current, is_active, location_system_id)
				 VALUES ($1,$2,'初始座驾',$3,$4,$5,$6,true,$7)`,
				created.ID, shipDefID, hp.ShieldHP, hp.ArmorHP, hp.StructHP, hp.Cap, originSystem)
		}
	}

	return created, nil
}

func (s *CharacterService) List(ctx context.Context, accountID int64) ([]model.Character, error) {
	return s.charRepo.ListByAccount(ctx, accountID)
}

func (s *CharacterService) Get(ctx context.Context, accountID, charID int64) (*model.Character, error) {
	c, err := s.charRepo.GetByID(ctx, charID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCharNotFound
	}
	if c.AccountID != accountID {
		return nil, ErrNotOwner
	}
	return c, nil
}

func (s *CharacterService) Delete(ctx context.Context, accountID, charID int64) error {
	c, err := s.charRepo.GetByID(ctx, charID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCharNotFound
	}
	if c.AccountID != accountID {
		return ErrNotOwner
	}
	return s.charRepo.Delete(ctx, charID)
}
