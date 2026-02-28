package blame

import (
	"errors"
	"fmt"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/google/uuid"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/actorctx"
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
	if err := db.Callback().Create().Before("gorm:create").Replace(cbBeforeCreate, beforeCreate); err != nil {
		return fmt.Errorf("gormblame: register create callback: %w", err)
	}
	if err := db.Callback().Update().Before("gorm:update").Replace(cbBeforeUpdate, beforeUpdate); err != nil {
		return fmt.Errorf("gormblame: register update callback: %w", err)
	}

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
		return
	}

	if tx.Statement.Schema == nil {
		if err := tx.Statement.Parse(tx.Statement.Dest); err != nil {
			return
		}
	}

	sch := tx.Statement.Schema
	if sch == nil {
		return
	}

	if setCreated {
		_ = setUUIDFieldIfEmpty(tx, sch, "created_by", actorID)
	}

	if setUpdated {
		_ = setUUIDField(tx, sch, "updated_by", actorID)
	}
}

func setUUIDField(tx *gorm.DB, sch *schema.Schema, dbColumn string, actorID uuid.UUID) error {
	f := sch.LookUpField(dbColumn)
	if f == nil {
		return nil
	}

	value, err := buildUUIDFieldValue(f, actorID)
	if err != nil {
		wrapped := fmt.Errorf("gormblame: set %s: %w", dbColumn, err)
		_ = tx.AddError(wrapped)
		return wrapped
	}

	if err := f.Set(tx.Statement.Context, tx.Statement.ReflectValue, value); err != nil {
		wrapped := fmt.Errorf("gormblame: set %s: %w", dbColumn, err)
		_ = tx.AddError(wrapped)
		return wrapped
	}

	return nil
}

func setUUIDFieldIfEmpty(tx *gorm.DB, sch *schema.Schema, dbColumn string, actorID uuid.UUID) error {
	f := sch.LookUpField(dbColumn)
	if f == nil {
		return nil
	}

	val, isZero := f.ValueOf(tx.Statement.Context, tx.Statement.ReflectValue)
	if !isUUIDEmpty(val, isZero) {
		return nil
	}

	return setUUIDField(tx, sch, dbColumn, actorID)
}

var uuidType = reflect.TypeOf(uuid.UUID{})

func buildUUIDFieldValue(f *schema.Field, actorID uuid.UUID) (any, error) {
	switch {
	case f.FieldType == uuidType:
		return actorID, nil

	case f.FieldType.Kind() == reflect.Ptr && f.FieldType.Elem() == uuidType:
		return new(actorID), nil

	case f.FieldType.ConvertibleTo(uuidType):
		return reflect.ValueOf(actorID).Convert(f.FieldType).Interface(), nil

	case f.FieldType.Kind() == reflect.Ptr && f.FieldType.Elem().ConvertibleTo(uuidType):
		v := reflect.New(f.FieldType.Elem())
		v.Elem().Set(reflect.ValueOf(actorID).Convert(f.FieldType.Elem()))
		return v.Interface(), nil

	default:
		return nil, fmt.Errorf("unsupported field type %s", f.FieldType)
	}
}

func isUUIDEmpty(val any, isZero bool) bool {
	if isZero || val == nil {
		return true
	}

	switch v := val.(type) {
	case uuid.UUID:
		return v == uuid.Nil
	case *uuid.UUID:
		return v == nil || *v == uuid.Nil
	default:
		return false
	}
}
