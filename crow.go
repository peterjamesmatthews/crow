package crow

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Database map[any][]any

type Crow struct {
	db   *gorm.DB
	seed Database
}

func New(db *gorm.DB) Crow { return Crow{db: db, seed: nil} }

func (c *Crow) Seed(seed Database) error {
	c.seed = seed

	var models []any
	var records []any
	for model, recs := range seed {
		models = append(models, model)
		records = append(records, recs...)
	}

	if err := c.db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto migrate models: %w", err)
	}

	// TODO collect and remove foreign key constraints (all constraints?)

	for _, record := range records {
		if err := c.db.Create(getPointerTo(record)).Error; err != nil {
			return fmt.Errorf("failed to create record %v: %w", record, err)
		}
	}

	// TODO restore collected foreign key constraints (all constraints?)

	return nil
}

func (c *Crow) Dump() (Database, error) {
	dump := make(Database)

	for model := range c.seed {
		tabler, ok := getPointerTo(model).(schema.Tabler)
		if !ok {
			return nil, fmt.Errorf("model %v does not implement schema.Tabler", model)
		}

		// create pointer to slice of model's concrete type
		m := reflect.TypeOf(getPointedTo(model))
		ms := reflect.SliceOf(m)
		records := reflect.New(ms).Interface()

		if err := c.db.Table(tabler.TableName()).Find(records).Error; err != nil {
			return nil, fmt.Errorf("failed to find records for model %v: %w", model, err)
		}

		// convert records to []any
		models := reflect.ValueOf(records).Elem()
		anys := make([]any, models.Len())
		for i := range anys {
			anys[i] = models.Index(i).Interface()
		}
		dump[getPointedTo(model)] = anys
	}

	return dump, nil
}

func getPointerTo(i any) any {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		return i
	}
	ptr := reflect.New(v.Type())
	ptr.Elem().Set(v)
	return ptr.Interface()
}

func getPointedTo(i any) any {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr {
		return i
	}
	return v.Elem().Interface()
}
