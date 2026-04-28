package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/starfall-warsong/sws/internal/repository"
)

var (
	ErrNoEncounter     = errors.New("当前没有可用的奇遇事件")
	ErrInvalidChoice   = errors.New("无效的选择")
	ErrEncounterOnCooldown = errors.New("奇遇冷却中")
)

type EncounterService struct {
	db      *sqlx.DB
	invRepo *repository.InventoryRepo
}

func NewEncounterService(db *sqlx.DB, invRepo *repository.InventoryRepo) *EncounterService {
	return &EncounterService{db: db, invRepo: invRepo}
}

type EncounterEvent struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	IntroText   string   `json:"intro_text"`
	Choices     []Choice `json:"choices"`
}

type Choice struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

type EncounterResult struct {
	ResultText    string `json:"result_text"`
	RewardItem    string `json:"reward_item,omitempty"`
	RewardQty     int    `json:"reward_quantity,omitempty"`
	RewardCredits int64  `json:"reward_credits,omitempty"`
	Messages      []string `json:"messages"`
}

func (s *EncounterService) TryTrigger(ctx context.Context, charID int64) (*EncounterEvent, error) {
	// Get character info
	var systemID int64
	var consPct int
	s.db.QueryRowContext(ctx,
		`SELECT current_system_id, consciousness_pct FROM characters WHERE id = $1`, charID,
	).Scan(&systemID, &consPct)

	var secLevel float64
	s.db.QueryRowContext(ctx, `SELECT security_level FROM star_systems WHERE id = $1`, systemID).Scan(&secLevel)

	// Check cooldown
	var lastEncounter time.Time
	err := s.db.QueryRowContext(ctx,
		`SELECT created_at FROM character_encounters WHERE character_id = $1 ORDER BY created_at DESC LIMIT 1`,
		charID).Scan(&lastEncounter)
	if err == nil && time.Since(lastEncounter) < 6*time.Hour {
		return nil, ErrEncounterOnCooldown
	}

	// Find eligible encounters
	type EncRow struct {
		ID          int64   `db:"id"`
		Name        string  `db:"name"`
		Type        string  `db:"encounter_type"`
		IntroText   string  `db:"intro_text"`
		BaseProbability float64 `db:"base_probability"`
		MinSecurity float64 `db:"min_security"`
		MaxSecurity float64 `db:"max_security"`
		MinConsciousness int `db:"min_consciousness"`
	}
	var candidates []EncRow
	s.db.SelectContext(ctx, &candidates,
		`SELECT id, name, encounter_type, intro_text, base_probability, min_security, max_security, min_consciousness
		 FROM encounter_defs WHERE is_active = TRUE`)

	// Filter and roll
	var eligible []EncRow
	for _, c := range candidates {
		if secLevel < c.MinSecurity || secLevel > c.MaxSecurity {
			continue
		}
		if consPct < c.MinConsciousness {
			continue
		}
		eligible = append(eligible, c)
	}

	if len(eligible) == 0 {
		return nil, ErrNoEncounter
	}

	// Environment modifier
	envMod := 1.0
	if secLevel >= 0.5 {
		envMod = 0.3
	} else if secLevel < 0 {
		envMod = 2.0
	}
	if consPct < 70 {
		envMod *= 1.5
	}

	// Roll for each
	for _, c := range eligible {
		prob := c.BaseProbability * envMod
		if rand.Float64() < prob {
			// Triggered! Get choices
			type ChoiceRow struct {
				ChoiceIndex int    `db:"choice_index"`
				ChoiceText  string `db:"choice_text"`
			}
			var choices []ChoiceRow
			s.db.SelectContext(ctx, &choices,
				`SELECT choice_index, choice_text FROM encounter_choices WHERE encounter_id = $1 ORDER BY choice_index`, c.ID)

			event := &EncounterEvent{
				ID:        c.ID,
				Name:      c.Name,
				Type:      c.Type,
				IntroText: c.IntroText,
			}
			for _, ch := range choices {
				event.Choices = append(event.Choices, Choice{Index: ch.ChoiceIndex, Text: ch.ChoiceText})
			}

			return event, nil
		}
	}

	return nil, ErrNoEncounter
}

type MakeChoiceRequest struct {
	EncounterID int64 `json:"encounter_id" binding:"required"`
	ChoiceIndex int   `json:"choice_index" binding:"required"`
}

func (s *EncounterService) MakeChoice(ctx context.Context, charID int64, req *MakeChoiceRequest) (*EncounterResult, error) {
	type ChoiceData struct {
		ResultText        string  `db:"result_text"`
		RewardItemID      int64   `db:"reward_item_id"`
		RewardQuantity    int     `db:"reward_quantity"`
		RewardCredits     int64   `db:"reward_credits"`
		ConsciousnessChg  int     `db:"consciousness_change"`
		TriggerCombatNPC  int64   `db:"trigger_combat_npc_id"`
	}
	var choice ChoiceData
	err := s.db.GetContext(ctx, &choice,
		`SELECT result_text, reward_item_id, reward_quantity, reward_credits, consciousness_change, trigger_combat_npc_id
		 FROM encounter_choices WHERE encounter_id = $1 AND choice_index = $2`,
		req.EncounterID, req.ChoiceIndex)
	if err != nil {
		return nil, ErrInvalidChoice
	}

	result := &EncounterResult{
		ResultText: choice.ResultText,
	}

	var systemID int64
	s.db.QueryRowContext(ctx, `SELECT current_system_id FROM characters WHERE id = $1`, charID).Scan(&systemID)

	// Apply rewards
	if choice.RewardItemID > 0 && choice.RewardQuantity > 0 {
		s.invRepo.AddOrUpsertItem(ctx, "character", charID, choice.RewardItemID, int64(choice.RewardQuantity), systemID)
		def, _ := s.invRepo.GetItemDef(ctx, choice.RewardItemID)
		name := "物品"
		if def != nil {
			name = def.Name
		}
		result.RewardItem = name
		result.RewardQty = choice.RewardQuantity
		result.Messages = append(result.Messages, fmt.Sprintf("获得: %s x%d", name, choice.RewardQuantity))
	}

	if choice.RewardCredits > 0 {
		s.db.ExecContext(ctx, `UPDATE characters SET balance = balance + $1 WHERE id = $2`, choice.RewardCredits, charID)
		result.RewardCredits = choice.RewardCredits
		result.Messages = append(result.Messages, fmt.Sprintf("获得: %d 星币", choice.RewardCredits))
	}

	if choice.ConsciousnessChg != 0 {
		s.db.ExecContext(ctx,
			`UPDATE characters SET consciousness_pct = LEAST(100, GREATEST(1, consciousness_pct + $1)) WHERE id = $2`,
			choice.ConsciousnessChg, charID)
		result.Messages = append(result.Messages, fmt.Sprintf("意识完整度 %+d%%", choice.ConsciousnessChg))
	}

	// Record encounter
	s.db.ExecContext(ctx,
		`INSERT INTO character_encounters (character_id, encounter_id, choice_made, system_id) VALUES ($1,$2,$3,$4)`,
		charID, req.EncounterID, req.ChoiceIndex, systemID)

	return result, nil
}
