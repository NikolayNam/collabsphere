package gormblame

import (
	"errors"
	"fmt"

	"github.com/NikolayNam/collabsphere-go/internal/platform/actorctx"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const (
	cbBeforeCreate = "gormblame:before_create"
	cbBeforeUpdate = "gormblame:before_update"
)

// Register registers GORM callbacks that set created_by / updated_by from context actor ID.
func Register(db *gorm.DB) error {
	if db == nil {
		return errors.New("gormblame: db is nil")
	}

	// Avoid double registration (panic in gorm if same name registered twice).
	// gorm doesn't expose "exists" directly; common pragmatic approach is:
	// - call Replace instead of Register
	_ = db.Callback().Create().Before("gorm:create").Replace(cbBeforeCreate, beforeCreate)
	_ = db.Callback().Update().Before("gorm:update").Replace(cbBeforeUpdate, beforeUpdate)

	return nil
}

func beforeCreate(tx *gorm.DB) {
	apply(tx, true, true) // on create: set created_by (if nil) and updated_by
}

func beforeUpdate(tx *gorm.DB) {
	apply(tx, false, true) // on update: set updated_by only
}

func apply(tx *gorm.DB, setCreated bool, setUpdated bool) {
	if tx == nil || tx.Statement == nil || tx.Error != nil {
		return
	}

	actorID, ok := actorctx.ActorID(tx.Statement.Context)
	if !ok || actorID == uuid.Nil {
		// No actor in context: do nothing (system tasks/migrations/etc.)
		return
	}

	// If schema not parsed (rare), try parse.
	if tx.Statement.Schema == nil {
		if err := tx.Statement.Parse(tx.Statement.Dest); err != nil {
			// Don't break write path
			return
		}
	}

	sch := tx.Statement.Schema
	if sch == nil {
		return
	}

	if setCreated {
		// Only set created_by if it's currently NULL / zero pointer
		_ = setUUIDPtrFieldIfEmpty(tx, sch, "created_by", actorID)
	}

	if setUpdated {
		// Always set updated_by (overwrite)
		_ = setUUIDPtrField(tx, sch, "updated_by", actorID)
	}
}

func setUUIDPtrField(tx *gorm.DB, sch *schema.Schema, dbColumn string, actorID uuid.UUID) error {
	f := sch.LookUpField(dbColumn)
	if f == nil {
		// model doesn't have this column — fine
		return nil
	}

	// Set dest field value for all records (handles struct/slice)
	if err := f.Set(tx.Statement.Context, tx.Statement.ReflectValue, new(actorID)); err != nil {
		// keep error but don't crash; set it to tx.Error so gorm will return it
		err := tx.AddError(fmt.Errorf("gormblame: set %s: %w", dbColumn, err))
		if err != nil {
			return err
		}
		return err
	}
	return nil
}

func setUUIDPtrFieldIfEmpty(tx *gorm.DB, sch *schema.Schema, dbColumn string, actorID uuid.UUID) error {
	f := sch.LookUpField(dbColumn)
	if f == nil {
		return nil
	}

	// Check current value; if already set, do not overwrite.
	// For slices, this checks first element only; if you batch-create mixed states, that's on you.
	val, isZero := f.ValueOf(tx.Statement.Context, tx.Statement.ReflectValue)
	if !isZero && val != nil {
		return nil
	}

	return setUUIDPtrField(tx, sch, dbColumn, actorID)
}
