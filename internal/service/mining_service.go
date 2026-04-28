package service

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/starfall-warsong/sws/internal/repository"
)

var (
	ErrBeltNotFound   = errors.New("矿带不存在")
	ErrBeltDepleted   = errors.New("矿带已经枯竭")
	ErrAlreadyMining  = errors.New("角色已经在采矿中")
	ErrNotMining      = errors.New("角色当前没有在采矿")
)

var beltOreMapping = map[string][]int64{
	"普通矿带": {1001, 1002},             // 铁陨石, 铜辉矿
	"富矿带":  {1001, 1002, 1003, 1004},   // +钛铁矿, 铬尖晶石
	"稀有矿带": {1005, 1006, 1007, 1008},  // 钨锰矿, 铌钽铁矿, 钼铅矿, 铂族砂矿
	"冰矿带":  {1201, 1202, 1203},         // 水冰, 重水冰, 氨冰
	"异常矿带": {1008, 1009, 1010},        // 铂族砂矿, 铪锆英石, 锕系重矿
}

type MiningService struct {
	db          *sqlx.DB
	starmapRepo *repository.StarmapRepo
	invRepo     *repository.InventoryRepo
}

func NewMiningService(db *sqlx.DB, starmapRepo *repository.StarmapRepo, invRepo *repository.InventoryRepo) *MiningService {
	return &MiningService{db: db, starmapRepo: starmapRepo, invRepo: invRepo}
}

type MiningStartRequest struct {
	BeltID int64 `json:"belt_id" binding:"required"`
}

type MiningStatus struct {
	IsMining     bool   `json:"is_mining"`
	BeltName     string `json:"belt_name,omitempty"`
	OreName      string `json:"ore_name,omitempty"`
	CycleTimeSec int    `json:"cycle_time_sec,omitempty"`
	YieldPerCycle int   `json:"yield_per_cycle,omitempty"`
	TotalMined   int64  `json:"total_mined"`
}

func (s *MiningService) StartMining(ctx context.Context, charID int64, req *MiningStartRequest) (*MiningStatus, error) {
	// Check not already mining
	var count int
	s.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM mining_sessions WHERE character_id = $1 AND status = 'active'`, charID)
	if count > 0 {
		return nil, ErrAlreadyMining
	}

	// Get belt
	var belt struct {
		ID           int64  `db:"id"`
		SystemID     int64  `db:"system_id"`
		BeltType     string `db:"belt_type"`
		RemainingPct int    `db:"remaining_pct"`
	}
	err := s.db.GetContext(ctx, &belt, `SELECT id, system_id, belt_type, remaining_pct FROM asteroid_belts WHERE id = $1`, req.BeltID)
	if err != nil {
		return nil, ErrBeltNotFound
	}
	if belt.RemainingPct <= 0 {
		return nil, ErrBeltDepleted
	}

	ores := beltOreMapping[belt.BeltType]
	if len(ores) == 0 {
		ores = []int64{1001} // fallback to iron
	}
	selectedOre := ores[rand.Intn(len(ores))]

	yieldPerCycle := 100
	cycleTime := 10 // 文字游戏10秒一周期

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO mining_sessions (character_id, belt_id, ore_item_id, yield_per_cycle, cycle_time_sec, status)
		 VALUES ($1, $2, $3, $4, $5, 'active')`,
		charID, belt.ID, selectedOre, yieldPerCycle, cycleTime)
	if err != nil {
		return nil, err
	}

	oreDef, _ := s.invRepo.GetItemDef(ctx, selectedOre)
	oreName := "未知矿石"
	if oreDef != nil {
		oreName = oreDef.Name
	}

	return &MiningStatus{
		IsMining:      true,
		BeltName:      belt.BeltType,
		OreName:       oreName,
		CycleTimeSec:  cycleTime,
		YieldPerCycle: yieldPerCycle,
	}, nil
}

func (s *MiningService) CollectMining(ctx context.Context, charID int64) (*MiningStatus, error) {
	var session struct {
		ID            int64     `db:"id"`
		BeltID        int64     `db:"belt_id"`
		OreItemID     int64     `db:"ore_item_id"`
		YieldPerCycle int       `db:"yield_per_cycle"`
		CycleTimeSec  int       `db:"cycle_time_sec"`
		LastCycleAt   time.Time `db:"last_cycle_at"`
	}
	err := s.db.GetContext(ctx, &session,
		`SELECT id, belt_id, ore_item_id, yield_per_cycle, cycle_time_sec, last_cycle_at
		 FROM mining_sessions WHERE character_id = $1 AND status = 'active'`, charID)
	if err != nil {
		return nil, ErrNotMining
	}

	elapsed := time.Since(session.LastCycleAt)
	completedCycles := int(elapsed.Seconds()) / session.CycleTimeSec
	if completedCycles <= 0 {
		return &MiningStatus{
			IsMining:      true,
			CycleTimeSec:  session.CycleTimeSec,
			YieldPerCycle: session.YieldPerCycle,
			TotalMined:    0,
		}, nil
	}

	totalYield := int64(completedCycles * session.YieldPerCycle)

	// Get character location
	var systemID int64
	s.db.GetContext(ctx, &systemID, `SELECT current_system_id FROM characters WHERE id = $1`, charID)

	err = s.invRepo.AddOrUpsertItem(ctx, "character", charID, session.OreItemID, totalYield, systemID)
	if err != nil {
		return nil, err
	}

	// Update last cycle time
	s.db.ExecContext(ctx,
		`UPDATE mining_sessions SET last_cycle_at = $1 WHERE id = $2`,
		session.LastCycleAt.Add(time.Duration(completedCycles*session.CycleTimeSec)*time.Second), session.ID)

	// Reduce belt remaining
	reduction := completedCycles / 10
	if reduction > 0 {
		s.db.ExecContext(ctx,
			`UPDATE asteroid_belts SET remaining_pct = GREATEST(0, remaining_pct - $1) WHERE id = $2`,
			reduction, session.BeltID)
	}

	oreDef, _ := s.invRepo.GetItemDef(ctx, session.OreItemID)
	oreName := "未知"
	if oreDef != nil {
		oreName = oreDef.Name
	}

	return &MiningStatus{
		IsMining:      true,
		OreName:       oreName,
		CycleTimeSec:  session.CycleTimeSec,
		YieldPerCycle: session.YieldPerCycle,
		TotalMined:    totalYield,
	}, nil
}

func (s *MiningService) StopMining(ctx context.Context, charID int64) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE mining_sessions SET status = 'stopped' WHERE character_id = $1 AND status = 'active'`, charID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotMining
	}
	return nil
}
