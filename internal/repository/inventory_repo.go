package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/starfall-warsong/sws/internal/model"
)

type InventoryRepo struct {
	db *sqlx.DB
}

func NewInventoryRepo(db *sqlx.DB) *InventoryRepo {
	return &InventoryRepo{db: db}
}

func (r *InventoryRepo) DB() *sqlx.DB { return r.db }

func (r *InventoryRepo) AddItemWithType(ctx context.Context, ownerType string, ownerID, itemDefID, quantity, locationID int64, locType string) error {
	var existing model.InventoryItem
	err := r.db.GetContext(ctx, &existing,
		`SELECT * FROM inventory WHERE owner_type=$1 AND owner_id=$2 AND item_def_id=$3 AND location_id=$4 AND location_type=$5`,
		ownerType, ownerID, itemDefID, locationID, locType)
	if err != nil {
		_, err = r.db.ExecContext(ctx,
			`INSERT INTO inventory (owner_type, owner_id, item_def_id, quantity, location_id, location_type) VALUES ($1,$2,$3,$4,$5,$6)`,
			ownerType, ownerID, itemDefID, quantity, locationID, locType)
	} else {
		_, err = r.db.ExecContext(ctx,
			`UPDATE inventory SET quantity=quantity+$1, updated_at=NOW() WHERE id=$2`, quantity, existing.ID)
	}
	return err
}

func (r *InventoryRepo) RemoveItemByType(ctx context.Context, ownerType string, ownerID, itemDefID, quantity, locationID int64, locType string) (bool, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE inventory SET quantity=quantity-$1, updated_at=NOW()
		 WHERE owner_type=$2 AND owner_id=$3 AND item_def_id=$4 AND location_id=$5 AND location_type=$6 AND quantity>=$1`,
		quantity, ownerType, ownerID, itemDefID, locationID, locType)
	if err != nil { return false, err }
	rows, _ := result.RowsAffected()
	if rows == 0 { return false, nil }
	r.db.ExecContext(ctx,
		`DELETE FROM inventory WHERE owner_type=$1 AND owner_id=$2 AND item_def_id=$3 AND location_id=$4 AND location_type=$5 AND quantity<=0`,
		ownerType, ownerID, itemDefID, locationID, locType)
	return true, nil
}

func (r *InventoryRepo) AddItem(ctx context.Context, ownerType string, ownerID, itemDefID, quantity, locationID int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO inventory (owner_type, owner_id, item_def_id, quantity, location_id)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT DO NOTHING`, // will use upsert logic below
		ownerType, ownerID, itemDefID, quantity, locationID)

	if err != nil {
		return err
	}

	// Try upsert: if same owner+item+location exists, add quantity
	_, err = r.db.ExecContext(ctx,
		`UPDATE inventory SET quantity = quantity + $1, updated_at = NOW()
		 WHERE owner_type = $2 AND owner_id = $3 AND item_def_id = $4 AND location_id = $5`,
		quantity, ownerType, ownerID, itemDefID, locationID)
	return err
}

func (r *InventoryRepo) AddOrUpsertItem(ctx context.Context, ownerType string, ownerID, itemDefID, quantity, locationID int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO inventory (owner_type, owner_id, item_def_id, quantity, location_id)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (id) DO UPDATE SET quantity = inventory.quantity + EXCLUDED.quantity, updated_at = NOW()`,
		ownerType, ownerID, itemDefID, quantity, locationID)

	if err != nil {
		// Fallback: check exists then insert or update
		var existing model.InventoryItem
		findErr := r.db.GetContext(ctx, &existing,
			`SELECT * FROM inventory WHERE owner_type=$1 AND owner_id=$2 AND item_def_id=$3 AND location_id=$4`,
			ownerType, ownerID, itemDefID, locationID)
		if findErr != nil {
			// Does not exist, insert
			_, err = r.db.ExecContext(ctx,
				`INSERT INTO inventory (owner_type, owner_id, item_def_id, quantity, location_id) VALUES ($1,$2,$3,$4,$5)`,
				ownerType, ownerID, itemDefID, quantity, locationID)
		} else {
			// Exists, update
			_, err = r.db.ExecContext(ctx,
				`UPDATE inventory SET quantity = quantity + $1, updated_at = NOW() WHERE id = $2`,
				quantity, existing.ID)
		}
	}
	return err
}

func (r *InventoryRepo) RemoveItem(ctx context.Context, ownerType string, ownerID, itemDefID, quantity, locationID int64) (bool, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE inventory SET quantity = quantity - $1, updated_at = NOW()
		 WHERE owner_type = $2 AND owner_id = $3 AND item_def_id = $4 AND location_id = $5 AND quantity >= $1`,
		quantity, ownerType, ownerID, itemDefID, locationID)
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return false, nil
	}

	// Clean up zero-quantity rows
	r.db.ExecContext(ctx,
		`DELETE FROM inventory WHERE owner_type = $1 AND owner_id = $2 AND item_def_id = $3 AND location_id = $4 AND quantity <= 0`,
		ownerType, ownerID, itemDefID, locationID)

	return true, nil
}

func (r *InventoryRepo) GetItemQuantity(ctx context.Context, ownerType string, ownerID, itemDefID, locationID int64) (int64, error) {
	var qty int64
	err := r.db.GetContext(ctx, &qty,
		`SELECT COALESCE(SUM(quantity), 0) FROM inventory
		 WHERE owner_type = $1 AND owner_id = $2 AND item_def_id = $3 AND location_id = $4`,
		ownerType, ownerID, itemDefID, locationID)
	return qty, err
}

func (r *InventoryRepo) ListByOwner(ctx context.Context, ownerType string, ownerID, locationID int64) ([]model.InventoryItem, error) {
	var items []model.InventoryItem
	err := r.db.SelectContext(ctx, &items,
		`SELECT id, owner_type, owner_id, item_def_id, quantity, location_id, created_at, updated_at
		 FROM inventory WHERE owner_type = $1 AND owner_id = $2 AND location_id = $3 ORDER BY item_def_id`,
		ownerType, ownerID, locationID)
	return items, err
}

func (r *InventoryRepo) GetItemDef(ctx context.Context, id int64) (*model.ItemDef, error) {
	var item model.ItemDef
	err := r.db.GetContext(ctx, &item,
		`SELECT id, name, category, description, volume, base_price, stackable, tech_level
		 FROM item_defs WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *InventoryRepo) ListItemDefs(ctx context.Context, category, slotType string) ([]model.ItemDef, error) {
	q := `SELECT id, name, category, description, volume, base_price, stackable, tech_level,
	      COALESCE(slot_type,'') as slot_type, COALESCE(module_type,'') as module_type,
	      COALESCE(damage_per_tick,0) as damage_per_tick, COALESCE(damage_type,'') as damage_type,
	      COALESCE(optimal_range,0) as optimal_range, COALESCE(falloff_range,0) as falloff_range,
	      COALESCE(tracking_speed,0) as tracking_speed, COALESCE(rate_of_fire,1) as rate_of_fire,
	      COALESCE(pg_cost,0) as pg_cost, COALESCE(cpu_cost,0) as cpu_cost, COALESCE(cap_cost,0) as cap_cost,
	      COALESCE(bonus_type,'') as bonus_type, COALESCE(bonus_value,0) as bonus_value
	      FROM item_defs`
	var items []model.ItemDef
	switch {
	case category != "" && slotType != "":
		err := r.db.SelectContext(ctx, &items, q+` WHERE category = $1 AND slot_type = $2 ORDER BY id`, category, slotType)
		return items, err
	case category != "":
		err := r.db.SelectContext(ctx, &items, q+` WHERE category = $1 ORDER BY id`, category)
		return items, err
	case slotType != "":
		err := r.db.SelectContext(ctx, &items, q+` WHERE slot_type = $1 ORDER BY id`, slotType)
		return items, err
	default:
		err := r.db.SelectContext(ctx, &items, q+` ORDER BY id`)
		return items, err
	}
}
