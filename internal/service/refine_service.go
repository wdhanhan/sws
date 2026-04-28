package service

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/starfall-warsong/sws/internal/repository"
)

var (
	ErrInsufficientOre = errors.New("矿石数量不足")
	ErrNoRecipe        = errors.New("没有找到该矿石的精炼配方")
)

type RefineService struct {
	db      *sqlx.DB
	invRepo *repository.InventoryRepo
}

func NewRefineService(db *sqlx.DB, invRepo *repository.InventoryRepo) *RefineService {
	return &RefineService{db: db, invRepo: invRepo}
}

type RefineRequest struct {
	ItemDefID int64 `json:"item_def_id" binding:"required"`
	Quantity  int64 `json:"quantity" binding:"required"`
}

type RefineOutput struct {
	ItemName string `json:"item_name"`
	Quantity int64  `json:"quantity"`
}

type RefineResult struct {
	InputName  string         `json:"input_name"`
	InputUsed  int64          `json:"input_used"`
	Outputs    []RefineOutput `json:"outputs"`
}

func (s *RefineService) Refine(ctx context.Context, charID int64, req *RefineRequest) (*RefineResult, error) {
	// Get character's location
	var systemID int64
	err := s.db.GetContext(ctx, &systemID, `SELECT current_system_id FROM characters WHERE id = $1`, charID)
	if err != nil {
		return nil, err
	}

	// Check quantity available
	available, err := s.invRepo.GetItemQuantity(ctx, "character", charID, req.ItemDefID, systemID)
	if err != nil {
		return nil, err
	}
	if available < req.Quantity {
		return nil, ErrInsufficientOre
	}

	// Get refine recipes for this input
	type Recipe struct {
		InputQuantity  int   `db:"input_quantity"`
		OutputItemID   int64 `db:"output_item_id"`
		OutputQuantity int   `db:"output_quantity"`
	}
	var recipes []Recipe
	err = s.db.SelectContext(ctx, &recipes,
		`SELECT input_quantity, output_item_id, output_quantity FROM refine_recipes WHERE input_item_id = $1`,
		req.ItemDefID)
	if err != nil || len(recipes) == 0 {
		return nil, ErrNoRecipe
	}

	// Calculate batches (using first recipe's input quantity as batch size)
	batchSize := int64(recipes[0].InputQuantity)
	batches := req.Quantity / batchSize
	if batches <= 0 {
		return nil, ErrInsufficientOre
	}
	actualInput := batches * batchSize

	// Remove input ore
	ok, err := s.invRepo.RemoveItem(ctx, "character", charID, req.ItemDefID, actualInput, systemID)
	if err != nil || !ok {
		return nil, ErrInsufficientOre
	}

	// Produce outputs
	inputDef, _ := s.invRepo.GetItemDef(ctx, req.ItemDefID)
	inputName := "未知"
	if inputDef != nil {
		inputName = inputDef.Name
	}

	var outputs []RefineOutput
	for _, recipe := range recipes {
		outputQty := int64(recipe.OutputQuantity) * batches
		err = s.invRepo.AddOrUpsertItem(ctx, "character", charID, recipe.OutputItemID, outputQty, systemID)
		if err != nil {
			continue
		}

		outDef, _ := s.invRepo.GetItemDef(ctx, recipe.OutputItemID)
		outName := "未知"
		if outDef != nil {
			outName = outDef.Name
		}
		outputs = append(outputs, RefineOutput{ItemName: outName, Quantity: outputQty})
	}

	return &RefineResult{
		InputName: inputName,
		InputUsed: actualInput,
		Outputs:   outputs,
	}, nil
}

func (s *RefineService) GetRecipes(ctx context.Context) ([]map[string]interface{}, error) {
	type RecipeRow struct {
		InputItemID    int64 `db:"input_item_id"`
		InputQuantity  int   `db:"input_quantity"`
		OutputItemID   int64 `db:"output_item_id"`
		OutputQuantity int   `db:"output_quantity"`
	}
	var rows []RecipeRow
	err := s.db.SelectContext(ctx, &rows, `SELECT input_item_id, input_quantity, output_item_id, output_quantity FROM refine_recipes ORDER BY input_item_id`)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, r := range rows {
		inDef, _ := s.invRepo.GetItemDef(ctx, r.InputItemID)
		outDef, _ := s.invRepo.GetItemDef(ctx, r.OutputItemID)
		inName, outName := "?", "?"
		if inDef != nil { inName = inDef.Name }
		if outDef != nil { outName = outDef.Name }

		result = append(result, map[string]interface{}{
			"input":           inName,
			"input_quantity":  r.InputQuantity,
			"output":          outName,
			"output_quantity": r.OutputQuantity,
		})
	}
	return result, nil
}
