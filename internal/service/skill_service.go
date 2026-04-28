package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrSkillNotFound     = errors.New("技能不存在")
	ErrSkillMaxLevel     = errors.New("技能已达到最高等级")
	ErrPrereqNotMet      = errors.New("前置技能条件未满足")
	ErrQueueFull         = errors.New("训练队列已满(最多10个)")
	ErrAlreadyTraining   = errors.New("该技能已在训练队列中")
	ErrImplantSlotLocked = errors.New("植入体槽位未解锁，需要提升生物适配技能")
	ErrImplantSlotFull   = errors.New("该槽位已有植入体")
	ErrNoCloneSlot       = errors.New("克隆槽位不足，需要提升克隆网络技能")
)

const MaxSkillLevel = 5

type SkillService struct {
	db *sqlx.DB
}

func NewSkillService(db *sqlx.DB) *SkillService {
	return &SkillService{db: db}
}

type SkillDef struct {
	ID             int64  `db:"id" json:"id"`
	Name           string `db:"name" json:"name"`
	Category       string `db:"category" json:"category"`
	Description    string `db:"description" json:"description"`
	Rank           int    `db:"rank" json:"rank"`
	PrimaryAttr    string `db:"primary_attr" json:"primary_attr"`
	SecondaryAttr  string `db:"secondary_attr" json:"secondary_attr"`
	PrereqSkillID  *int64 `db:"prereq_skill_id" json:"prereq_skill_id,omitempty"`
	PrereqLevel    int    `db:"prereq_level" json:"prereq_level"`
}

type CharacterSkill struct {
	SkillDefID  int64  `db:"skill_def_id" json:"skill_def_id"`
	Name        string `db:"name" json:"name"`
	Category    string `db:"category" json:"category"`
	Level       int    `db:"level" json:"level"`
	SkillPoints int64  `db:"skill_points" json:"skill_points"`
	SPForNext   int64  `json:"sp_for_next"`
}

type QueueItem struct {
	SkillDefID   int64      `db:"skill_def_id" json:"skill_def_id"`
	Name         string     `json:"name"`
	TargetLevel  int        `db:"target_level" json:"target_level"`
	Position     int        `db:"queue_position" json:"queue_position"`
	StartTime    *time.Time `db:"start_time" json:"start_time,omitempty"`
	FinishTime   *time.Time `db:"finish_time" json:"finish_time,omitempty"`
	IsActive     bool       `db:"is_active" json:"is_active"`
	TimeRemaining string    `json:"time_remaining,omitempty"`
}

// SPForLevel returns total SP needed to reach a given level (1-5)
// Formula: 250 * rank * 2^(2.5*(level-1))
func SPForLevel(rank, level int) int64 {
	if level <= 0 {
		return 0
	}
	return int64(250 * float64(rank) * math.Pow(2, 2.5*float64(level-1)))
}

// TrainingTime returns duration to train from currentSP to target level
func TrainingTime(rank, targetLevel int, currentSP int64) time.Duration {
	targetSP := SPForLevel(rank, targetLevel)
	remaining := targetSP - currentSP
	if remaining <= 0 {
		return 0
	}
	spPerMinute := 30.0 / float64(rank) // base: rank1=30sp/min, rank5=6sp/min
	minutes := float64(remaining) / spPerMinute
	return time.Duration(minutes) * time.Minute
}

func (s *SkillService) GetSkillDefs(ctx context.Context, category string) ([]SkillDef, error) {
	var skills []SkillDef
	if category != "" {
		return skills, s.db.SelectContext(ctx, &skills, `SELECT * FROM skill_defs WHERE category = $1 ORDER BY id`, category)
	}
	return skills, s.db.SelectContext(ctx, &skills, `SELECT * FROM skill_defs ORDER BY category, id`)
}

func (s *SkillService) GetCharacterSkills(ctx context.Context, charID int64) ([]CharacterSkill, error) {
	var skills []CharacterSkill
	err := s.db.SelectContext(ctx, &skills,
		`SELECT cs.skill_def_id, sd.name, sd.category, cs.level, cs.skill_points
		 FROM character_skills cs
		 JOIN skill_defs sd ON sd.id = cs.skill_def_id
		 WHERE cs.character_id = $1
		 ORDER BY sd.category, sd.id`, charID)
	for i := range skills {
		var rank int
		s.db.GetContext(ctx, &rank, `SELECT rank FROM skill_defs WHERE id = $1`, skills[i].SkillDefID)
		if skills[i].Level < MaxSkillLevel {
			skills[i].SPForNext = SPForLevel(rank, skills[i].Level+1)
		}
	}
	return skills, err
}

type TrainRequest struct {
	SkillDefID int64 `json:"skill_def_id" binding:"required"`
}

func (s *SkillService) AddToQueue(ctx context.Context, charID int64, req *TrainRequest) (*QueueItem, error) {
	var skill SkillDef
	if err := s.db.GetContext(ctx, &skill, `SELECT * FROM skill_defs WHERE id = $1`, req.SkillDefID); err != nil {
		return nil, ErrSkillNotFound
	}

	// Check current level
	var currentLevel int
	var currentSP int64
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(level, 0), COALESCE(skill_points, 0) FROM character_skills WHERE character_id = $1 AND skill_def_id = $2`,
		charID, req.SkillDefID).Scan(&currentLevel, &currentSP)
	if err != nil {
		currentLevel = 0
		currentSP = 0
	}

	targetLevel := currentLevel + 1
	if targetLevel > MaxSkillLevel {
		return nil, ErrSkillMaxLevel
	}

	// Check prerequisites
	if skill.PrereqSkillID != nil && *skill.PrereqSkillID > 0 {
		var prereqLevel int
		s.db.QueryRowContext(ctx,
			`SELECT COALESCE(level, 0) FROM character_skills WHERE character_id = $1 AND skill_def_id = $2`,
			charID, skill.PrereqSkillID).Scan(&prereqLevel)
		if prereqLevel < skill.PrereqLevel {
			return nil, ErrPrereqNotMet
		}
	}

	// Check queue size
	var queueSize int
	s.db.GetContext(ctx, &queueSize, `SELECT COUNT(*) FROM skill_queue WHERE character_id = $1`, charID)
	if queueSize >= 10 {
		return nil, ErrQueueFull
	}

	// Check not already in queue
	var inQueue int
	s.db.GetContext(ctx, &inQueue, `SELECT COUNT(*) FROM skill_queue WHERE character_id = $1 AND skill_def_id = $2`, charID, req.SkillDefID)
	if inQueue > 0 {
		return nil, ErrAlreadyTraining
	}

	duration := TrainingTime(skill.Rank, targetLevel, currentSP)
	now := time.Now()
	isFirst := queueSize == 0

	var startTime, finishTime *time.Time
	if isFirst {
		st := now
		ft := now.Add(duration)
		startTime = &st
		finishTime = &ft
	}

	s.db.ExecContext(ctx,
		`INSERT INTO skill_queue (character_id, skill_def_id, target_level, queue_position, start_time, finish_time, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		charID, req.SkillDefID, targetLevel, queueSize, startTime, finishTime, isFirst)

	// Ensure skill row exists
	s.db.ExecContext(ctx,
		`INSERT INTO character_skills (character_id, skill_def_id, level, skill_points)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (character_id, skill_def_id) DO NOTHING`,
		charID, req.SkillDefID, currentLevel, currentSP)

	timeStr := ""
	if duration > 0 {
		timeStr = formatDuration(duration)
	}

	return &QueueItem{
		SkillDefID:    req.SkillDefID,
		Name:          skill.Name,
		TargetLevel:   targetLevel,
		Position:      queueSize,
		StartTime:     startTime,
		FinishTime:    finishTime,
		IsActive:      isFirst,
		TimeRemaining: timeStr,
	}, nil
}

func (s *SkillService) GetQueue(ctx context.Context, charID int64) ([]QueueItem, error) {
	type row struct {
		SkillDefID  int64      `db:"skill_def_id"`
		TargetLevel int        `db:"target_level"`
		Position    int        `db:"queue_position"`
		StartTime   *time.Time `db:"start_time"`
		FinishTime  *time.Time `db:"finish_time"`
		IsActive    bool       `db:"is_active"`
	}
	var rows []row
	err := s.db.SelectContext(ctx, &rows,
		`SELECT skill_def_id, target_level, queue_position, start_time, finish_time, is_active
		 FROM skill_queue WHERE character_id = $1 ORDER BY queue_position`, charID)
	if err != nil {
		return nil, err
	}

	items := make([]QueueItem, len(rows))
	for i, r := range rows {
		var name string
		s.db.GetContext(ctx, &name, `SELECT name FROM skill_defs WHERE id = $1`, r.SkillDefID)
		timeStr := ""
		if r.FinishTime != nil && r.IsActive {
			remaining := time.Until(*r.FinishTime)
			if remaining > 0 {
				timeStr = formatDuration(remaining)
			} else {
				timeStr = "已完成"
			}
		}
		items[i] = QueueItem{
			SkillDefID:    r.SkillDefID,
			Name:          name,
			TargetLevel:   r.TargetLevel,
			Position:      r.Position,
			StartTime:     r.StartTime,
			FinishTime:    r.FinishTime,
			IsActive:      r.IsActive,
			TimeRemaining: timeStr,
		}
	}
	return items, nil
}

// ProcessCompletedSkills checks and applies finished training
func (s *SkillService) ProcessCompleted(ctx context.Context, charID int64) ([]string, error) {
	type queueRow struct {
		ID          int64  `db:"id"`
		SkillDefID  int64  `db:"skill_def_id"`
		TargetLevel int    `db:"target_level"`
	}
	var completed []queueRow
	err := s.db.SelectContext(ctx, &completed,
		`SELECT id, skill_def_id, target_level FROM skill_queue
		 WHERE character_id = $1 AND is_active = TRUE AND finish_time <= NOW()`, charID)
	if err != nil {
		return nil, err
	}

	var messages []string
	for _, c := range completed {
		targetSP := SPForLevel(1, c.TargetLevel) // simplified
		s.db.ExecContext(ctx,
			`UPDATE character_skills SET level = $1, skill_points = $2 WHERE character_id = $3 AND skill_def_id = $4`,
			c.TargetLevel, targetSP, charID, c.SkillDefID)
		s.db.ExecContext(ctx, `DELETE FROM skill_queue WHERE id = $1`, c.ID)

		var name string
		s.db.GetContext(ctx, &name, `SELECT name FROM skill_defs WHERE id = $1`, c.SkillDefID)
		messages = append(messages, name+" 已训练到 "+romanNumeral(c.TargetLevel)+" 级")
	}

	// Activate next in queue
	if len(completed) > 0 {
		var next queueRow
		err := s.db.GetContext(ctx, &next,
			`SELECT id, skill_def_id, target_level FROM skill_queue
			 WHERE character_id = $1 ORDER BY queue_position LIMIT 1`, charID)
		if err == nil {
			var rank int
			s.db.GetContext(ctx, &rank, `SELECT rank FROM skill_defs WHERE id = $1`, next.SkillDefID)
			var currentSP int64
			s.db.QueryRowContext(ctx,
				`SELECT COALESCE(skill_points, 0) FROM character_skills WHERE character_id=$1 AND skill_def_id=$2`,
				charID, next.SkillDefID).Scan(&currentSP)
			duration := TrainingTime(rank, next.TargetLevel, currentSP)
			now := time.Now()
			finish := now.Add(duration)
			s.db.ExecContext(ctx,
				`UPDATE skill_queue SET is_active = TRUE, start_time = $1, finish_time = $2, queue_position = 0 WHERE id = $3`,
				now, finish, next.ID)
		}
	}

	return messages, nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "不到1分钟"
	}
	if d < time.Hour {
		return fmt.Sprintf("%d分钟", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d小时%d分钟", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%d天%d小时", days, hours)
}

func romanNumeral(n int) string {
	switch n {
	case 1: return "I"
	case 2: return "II"
	case 3: return "III"
	case 4: return "IV"
	case 5: return "V"
	default: return fmt.Sprintf("%d", n)
	}
}
