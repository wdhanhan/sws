package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type DeathService struct {
	db *sqlx.DB
}

func NewDeathService(db *sqlx.DB) *DeathService {
	return &DeathService{db: db}
}

type DeathResult struct {
	RespawnSystem  int64  `json:"respawn_system"`
	RespawnStation string `json:"respawn_station"`
	ConsciousnessLost int `json:"consciousness_lost"`
	ConsciousnessNow  int `json:"consciousness_now"`
	ImplantsAffected  int `json:"implants_damaged"`
	EchoCreated    bool   `json:"echo_created"`
	Messages       []string `json:"messages"`
}

func (s *DeathService) ProcessDeath(ctx context.Context, charID int64, killedBy string) (*DeathResult, error) {
	result := &DeathResult{}

	// 1. 降低意识完整度 -3%
	s.db.ExecContext(ctx,
		`UPDATE characters SET consciousness_pct = GREATEST(1, consciousness_pct - 3) WHERE id = $1`, charID)

	var consPct int
	s.db.GetContext(ctx, &consPct, `SELECT consciousness_pct FROM characters WHERE id = $1`, charID)
	result.ConsciousnessLost = 3
	result.ConsciousnessNow = consPct
	result.Messages = append(result.Messages, "意识完整度 -3%，当前: "+fmt.Sprintf("%d%%", consPct))

	// 2. 降低所有植入体稳定度 -20%
	res, _ := s.db.ExecContext(ctx,
		`UPDATE character_implants SET stability = GREATEST(0, stability - 20)
		 WHERE character_id = $1 AND is_active = TRUE`, charID)
	affected, _ := res.RowsAffected()
	result.ImplantsAffected = int(affected)
	if affected > 0 {
		result.Messages = append(result.Messages, fmt.Sprintf("%d个植入体稳定度-20%%", affected))
	}

	// 3. 创建死亡回响
	var posX, posY, posZ float64
	var systemID int64
	s.db.QueryRowContext(ctx,
		`SELECT current_system_id, pos_x, pos_y, pos_z FROM characters WHERE id = $1`, charID,
	).Scan(&systemID, &posX, &posY, &posZ)

	s.db.ExecContext(ctx,
		`INSERT INTO death_echoes (character_id, system_id, pos_x, pos_y, pos_z, killed_by, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		charID, systemID, posX, posY, posZ, killedBy, time.Now().Add(72*time.Hour))
	result.EchoCreated = true
	result.Messages = append(result.Messages, "死亡回响已在当前位置生成(72小时后消散)")

	// 4. 传送到克隆所在站
	type CloneInfo struct {
		StationID int64 `db:"station_id"`
		SystemID  int64 `db:"system_id"`
	}
	var clone CloneInfo
	err := s.db.GetContext(ctx, &clone,
		`SELECT station_id, system_id FROM clones
		 WHERE character_id = $1 AND clone_type = 'main' AND is_active = TRUE
		 LIMIT 1`, charID)
	if err != nil {
		// No clone set, respawn at origin
		s.db.GetContext(ctx, &clone.SystemID,
			`SELECT current_system_id FROM characters WHERE id = $1`, charID)
		clone.StationID = 0
	}

	s.db.ExecContext(ctx,
		`UPDATE characters SET current_system_id = $1, is_docked = TRUE, docked_station_id = $2,
		 pos_x = 0, pos_y = 0, pos_z = 0 WHERE id = $3`,
		clone.SystemID, clone.StationID, charID)

	result.RespawnSystem = clone.SystemID

	var stationName string
	s.db.GetContext(ctx, &stationName, `SELECT name FROM stations WHERE id = $1`, clone.StationID)
	if stationName == "" {
		stationName = "起源空间站"
	}
	result.RespawnStation = stationName
	result.Messages = append(result.Messages, "意识已传输至克隆体: "+stationName)

	// 5. 低意识完整度特殊效果
	if consPct < 70 {
		result.Messages = append(result.Messages, "⚠ 意识完整度低于70%，技能训练速度-5%")
	}
	if consPct < 50 {
		result.Messages = append(result.Messages, "⚠ 意识完整度低于50%，植入体效果-10%，但解锁了【意识边界】特殊任务线")
	}
	if consPct < 30 {
		result.Messages = append(result.Messages, "⚠ 意识完整度低于30%，偶尔会进入【超验状态】获得随机强力buff")
	}

	return result, nil
}

func (s *DeathService) CollectEcho(ctx context.Context, charID, echoID int64) (string, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE death_echoes SET is_collected = TRUE
		 WHERE id = $1 AND character_id = $2 AND is_collected = FALSE AND expires_at > NOW()`,
		echoID, charID)
	if err != nil {
		return "", err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return "", errors.New("回响不存在或已过期")
	}

	s.db.ExecContext(ctx,
		`UPDATE characters SET consciousness_pct = LEAST(100, consciousness_pct + 2) WHERE id = $1`, charID)

	return "成功回收死亡回响，意识完整度 +2%", nil
}
