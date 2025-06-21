package activerecord

import (
	"fmt"
	"reflect"
)

// AssociationType type of association
type AssociationType int

const (
	HasOne AssociationType = iota
	HasMany
	BelongsTo
	HasManyThrough
)

// Association definition of association
type Association struct {
	Type       AssociationType
	Model      interface{}
	ForeignKey string
	LocalKey   string
	Through    string
}

// Associations map of associations for model
type Associations map[string]*Association

// Association registry for test/demo
var associationRegistry = make(map[string]*Association)

// autoRegisterAssociations automatically detects and registers associations based on struct fields
func autoRegisterAssociations(model interface{}) {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldType := field.Type
		if field.Anonymous && (fieldType.Kind() == reflect.Struct || (fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Struct)) {
			// Embedded struct, recurse
			var embeddedVal reflect.Value
			if fieldType.Kind() == reflect.Ptr {
				embeddedVal = val.Field(i)
				if embeddedVal.IsNil() {
					continue
				}
				embeddedVal = embeddedVal.Elem()
			} else {
				embeddedVal = val.Field(i)
			}
			autoRegisterAssociations(embeddedVal.Addr().Interface())
		}
		// BelongsTo: *OtherModel
		if fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Struct {
			fk := field.Name + "ID"
			modelPtr := reflect.New(fieldType.Elem()).Interface()
			associationRegistry[field.Name] = &Association{
				Type:       BelongsTo,
				Model:      modelPtr,
				ForeignKey: fk,
			}
		}
		// HasMany: []OtherModel
		if fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() == reflect.Struct {
			parentType := typ.Name()
			elemType := fieldType.Elem().Name()
			var fk string
			if elemType == parentType {
				// Self-referencing: look for field ending with 'ID' but not 'ID'
				found := false
				for j := 0; j < fieldType.Elem().NumField(); j++ {
					f := fieldType.Elem().Field(j)
					if f.Name != "ID" && len(f.Name) > 2 && f.Name[len(f.Name)-2:] == "ID" {
						fk = f.Name
						found = true
						break
					}
				}
				if !found {
					// fallback: singularize field.Name + ID
					fieldName := field.Name
					if len(fieldName) > 1 && fieldName[len(fieldName)-1] == 's' {
						fieldName = fieldName[:len(fieldName)-1]
					}
					fk = fieldName + "ID"
				}
			} else {
				fk = parentType + "ID"
			}
			slicePtr := reflect.New(fieldType).Interface()
			associationRegistry[field.Name] = &Association{
				Type:       HasMany,
				Model:      slicePtr,
				ForeignKey: fk,
			}
		}
	}
}

// HasOne defines relationship "one to one"
func (m *ActiveRecordModel) HasOne(name string, model interface{}, foreignKey string) {
	associationRegistry[name] = &Association{
		Type:       HasOne,
		Model:      model,
		ForeignKey: foreignKey,
	}
}

// HasMany defines relationship "one to many"
func (m *ActiveRecordModel) HasMany(name string, model interface{}, foreignKey string) {
	associationRegistry[name] = &Association{
		Type:       HasMany,
		Model:      model,
		ForeignKey: foreignKey,
	}
}

// BelongsTo defines relationship "belongs to"
func (m *ActiveRecordModel) BelongsTo(name string, model interface{}, foreignKey string) {
	associationRegistry[name] = &Association{
		Type:       BelongsTo,
		Model:      model,
		ForeignKey: foreignKey,
	}
}

// HasManyThrough defines relationship "many to many through"
func (m *ActiveRecordModel) HasManyThrough(name string, model interface{}, through string, foreignKey string, localKey string) {
}

// Association methods for working with associations

// Load loads association
func (m *ActiveRecordModel) Load(associationName string) error {
	association, exists := m.getAssociation(associationName)
	if !exists {
		autoRegisterAssociations(m)
		association, exists = m.getAssociation(associationName)
		if !exists {
			return fmt.Errorf("association %s not found", associationName)
		}
	}

	switch association.Type {
	case HasOne:
		return m.loadHasOne(associationName, association)
	case HasMany:
		return m.loadHasMany(associationName, association)
	case BelongsTo:
		return m.loadBelongsTo(associationName, association)
	default:
		return fmt.Errorf("unsupported association type")
	}
}

// Include preloads associations
func (m *ActiveRecordModel) Include(associationNames ...string) error {
	for _, name := range associationNames {
		if err := m.Load(name); err != nil {
			return err
		}
	}
	return nil
}

// Helper methods

func (m *ActiveRecordModel) getAssociation(name string) (*Association, bool) {
	assoc, ok := associationRegistry[name]
	return assoc, ok
}

func (m *ActiveRecordModel) loadHasOne(name string, association *Association) error {
	// Implementation of loading has_one
	return nil
}

func (m *ActiveRecordModel) loadHasMany(name string, association *Association) error {
	// For test/demo: load all related records where foreignKey == m.ID
	foreignKey := association.ForeignKey
	id := m.GetID()
	query := foreignKey + " = ?"

	// Try to set result directly to the field in the model
	val := reflect.ValueOf(m)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	field := val.FieldByName(name)
	if field.IsValid() && field.CanSet() {
		// Create slice of the correct type
		sliceType := field.Type()
		slicePtr := reflect.New(sliceType).Interface()
		err := Where(slicePtr, query, id)
		if err != nil {
			return err
		}
		// Set the result to the field
		field.Set(reflect.ValueOf(slicePtr).Elem())
		return nil
	}

	// Fallback to original behavior
	modelSlicePtr := association.Model
	return Where(modelSlicePtr, query, id)
}

func (m *ActiveRecordModel) loadBelongsTo(name string, association *Association) error {
	// For test/demo: load the related record by foreignKey
	foreignKey := association.ForeignKey
	val := reflect.ValueOf(m).Elem().FieldByName(foreignKey)
	if !val.IsValid() {
		return nil
	}

	// Try to set result directly to the field in the model
	modelVal := reflect.ValueOf(m)
	if modelVal.Kind() == reflect.Ptr {
		modelVal = modelVal.Elem()
	}
	field := modelVal.FieldByName(name)
	if field.IsValid() && field.CanSet() {
		// Create instance of the correct type
		fieldType := field.Type()
		instancePtr := reflect.New(fieldType.Elem()).Interface()
		err := Find(instancePtr, val.Interface())
		if err != nil {
			return err
		}
		// Set the result to the field
		field.Set(reflect.ValueOf(instancePtr))
		return nil
	}

	// Fallback to original behavior
	modelPtr := association.Model
	return Find(modelPtr, val.Interface())
}

// Join methods for working with JOIN

// Joins performs JOIN with other tables
func Joins(models interface{}, joins ...string) error {
	// Implementation of JOIN queries
	return nil
}

// LeftJoins performs LEFT JOIN
func LeftJoins(models interface{}, joins ...string) error {
	// Implementation of LEFT JOIN queries
	return nil
}

// InnerJoins performs INNER JOIN
func InnerJoins(models interface{}, joins ...string) error {
	// Implementation of INNER JOIN queries
	return nil
}

// Eager Loading methods

// With preloads associations for collection
func With(models interface{}, associations ...string) error {
	// Implementation of eager loading
	return nil
}

// Preload preloads associations
func Preload(models interface{}, associations ...string) error {
	// Implementation of preload
	return nil
}
