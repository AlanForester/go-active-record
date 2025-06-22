package activerecord

import (
	"fmt"
	"reflect"
	"sort"
)

// HookType represents the type of hook
type HookType string

const (
	BeforeCreate HookType = "before_create"
	AfterCreate  HookType = "after_create"
	BeforeSave   HookType = "before_save"
	AfterSave    HookType = "after_save"
	BeforeUpdate HookType = "before_update"
	AfterUpdate  HookType = "after_update"
	BeforeDelete HookType = "before_delete"
	AfterDelete  HookType = "after_delete"
	BeforeFind   HookType = "before_find"
	AfterFind    HookType = "after_find"
)

// Hook represents a hook callback
type Hook struct {
	Type     HookType
	Priority int
	Callback func(interface{}) error
}

// Hookable interface for models that support hooks
type Hookable interface {
	AddHook(hookType HookType, callback func(interface{}) error)
	AddHookWithPriority(hookType HookType, priority int, callback func(interface{}) error)
	RunHooks(hookType HookType) error
	ClearHooks(hookType HookType)
}

// HookableModel embeds ActiveRecordModel and adds hook functionality
type HookableModel struct {
	ActiveRecordModel
	hooks map[HookType][]*Hook
}

// NewHookableModel creates a new HookableModel
func NewHookableModel() *HookableModel {
	return &HookableModel{
		hooks: make(map[HookType][]*Hook),
	}
}

// AddHook adds a hook with default priority (0)
func (m *HookableModel) AddHook(hookType HookType, callback func(interface{}) error) {
	m.AddHookWithPriority(hookType, 0, callback)
}

// AddHookWithPriority adds a hook with specified priority
func (m *HookableModel) AddHookWithPriority(hookType HookType, priority int, callback func(interface{}) error) {
	if m.hooks == nil {
		m.hooks = make(map[HookType][]*Hook)
	}

	hook := &Hook{
		Type:     hookType,
		Priority: priority,
		Callback: callback,
	}

	m.hooks[hookType] = append(m.hooks[hookType], hook)

	// Sort hooks by priority (lower numbers = higher priority)
	sort.Slice(m.hooks[hookType], func(i, j int) bool {
		return m.hooks[hookType][i].Priority < m.hooks[hookType][j].Priority
	})
}

// RunHooks executes all hooks of the specified type
func (m *HookableModel) RunHooks(hookType HookType) error {
	if m.hooks == nil {
		return nil
	}

	hooks, exists := m.hooks[hookType]
	if !exists {
		return nil
	}

	for _, hook := range hooks {
		if err := hook.Callback(m); err != nil {
			return fmt.Errorf("hook %s failed: %w", hookType, err)
		}
	}

	return nil
}

// ClearHooks removes all hooks of the specified type
func (m *HookableModel) ClearHooks(hookType HookType) {
	if m.hooks != nil {
		delete(m.hooks, hookType)
	}
}

// ClearAllHooks removes all hooks
func (m *HookableModel) ClearAllHooks() {
	m.hooks = make(map[HookType][]*Hook)
}

// Override Create to include hooks
func (m *HookableModel) Create() error {
	if err := m.RunHooks(BeforeCreate); err != nil {
		return err
	}

	if err := Create(m); err != nil {
		return err
	}

	return m.RunHooks(AfterCreate)
}

// Override Update to include hooks
func (m *HookableModel) Update() error {
	if err := m.RunHooks(BeforeUpdate); err != nil {
		return err
	}

	if err := Update(m); err != nil {
		return err
	}

	return m.RunHooks(AfterUpdate)
}

// Override Save to include hooks
func (m *HookableModel) Save() error {
	if err := m.RunHooks(BeforeSave); err != nil {
		return err
	}

	var err error
	if m.IsNewRecord() {
		err = m.Create()
	} else {
		err = m.Update()
	}

	if err != nil {
		return err
	}

	return m.RunHooks(AfterSave)
}

// Override Delete to include hooks
func (m *HookableModel) Delete() error {
	if err := m.RunHooks(BeforeDelete); err != nil {
		return err
	}

	if err := Delete(m); err != nil {
		return err
	}

	return m.RunHooks(AfterDelete)
}

// Override Find to include hooks
func (m *HookableModel) Find(id interface{}) error {
	if err := m.RunHooks(BeforeFind); err != nil {
		return err
	}

	if err := Find(m, id); err != nil {
		return err
	}

	return m.RunHooks(AfterFind)
}

// Global hook registry for models that don't embed HookableModel
var globalHooks = make(map[string]map[HookType][]*Hook)

// AddGlobalHook adds a hook for a specific model type
func AddGlobalHook(modelType string, hookType HookType, callback func(interface{}) error) {
	AddGlobalHookWithPriority(modelType, hookType, 0, callback)
}

// AddGlobalHookWithPriority adds a hook with priority for a specific model type
func AddGlobalHookWithPriority(modelType string, hookType HookType, priority int, callback func(interface{}) error) {
	if globalHooks[modelType] == nil {
		globalHooks[modelType] = make(map[HookType][]*Hook)
	}

	hook := &Hook{
		Type:     hookType,
		Priority: priority,
		Callback: callback,
	}

	globalHooks[modelType][hookType] = append(globalHooks[modelType][hookType], hook)

	// Sort hooks by priority
	sort.Slice(globalHooks[modelType][hookType], func(i, j int) bool {
		return globalHooks[modelType][hookType][i].Priority < globalHooks[modelType][hookType][j].Priority
	})
}

// RunGlobalHooks executes global hooks for a model
func RunGlobalHooks(model interface{}, hookType HookType) error {
	modelType := reflect.TypeOf(model).String()

	if globalHooks[modelType] == nil {
		return nil
	}

	hooks, exists := globalHooks[modelType][hookType]
	if !exists {
		return nil
	}

	for _, hook := range hooks {
		if err := hook.Callback(model); err != nil {
			return fmt.Errorf("global hook %s failed: %w", hookType, err)
		}
	}

	return nil
}

// ClearGlobalHooks removes all global hooks for a model type
func ClearGlobalHooks(modelType string, hookType HookType) {
	if globalHooks[modelType] != nil {
		delete(globalHooks[modelType], hookType)
	}
}
