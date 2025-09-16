package context

import (
	"context"
	"sync"
)

type Current struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

func NewCurrent() *Current {
	return &Current{
		data: make(map[string]interface{}),
	}
}

func (c *Current) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *Current) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value := c.data[key]
	return value
}

func (c *Current) GetString(key string) (string, bool) {
	value := c.Get(key)
	if value == nil {
		return "", false
	}
	if str, ok := value.(string); ok {
		return str, true
	}
	return "", false
}

func (c *Current) GetInt(key string) (int, bool) {
	value := c.Get(key)
	if value == nil {
		return 0, false
	}
	if i, ok := value.(int); ok {
		return i, true
	}
	return 0, false
}

func (c *Current) GetFloat64(key string) (float64, bool) {
	value := c.Get(key)
	if value == nil {
		return 0, false
	}
	if f, ok := value.(float64); ok {
		return f, true
	}
	return 0, false
}

func (c *Current) GetBool(key string) (bool, bool) {
	value := c.Get(key)
	if value == nil {
		return false, false
	}
	if b, ok := value.(bool); ok {
		return b, true
	}
	return false, false
}

func (c *Current) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *Current) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]interface{})
}

func (c *Current) All() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

func (c *Current) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys
}

func (c *Current) Exists(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.data[key]
	return exists
}

type contextKey string

const currentKey contextKey = "current"

func WithCurrent(ctx context.Context, current *Current) context.Context {
	return context.WithValue(ctx, currentKey, current)
}

func FromContext(ctx context.Context) (*Current, bool) {
	current, ok := ctx.Value(currentKey).(*Current)
	return current, ok
}

func GetCurrent(ctx context.Context) *Current {
	if current, ok := FromContext(ctx); ok {
		return current
	}

	return NewCurrent()
}

func SetCurrent(ctx context.Context, current *Current) context.Context {
	return WithCurrent(ctx, current)
}

var globalCurrent *Current

func Set(key string, value interface{}) {
	if globalCurrent == nil {
		globalCurrent = NewCurrent()
	}
	globalCurrent.Set(key, value)
}

func Get(key string) interface{} {
	if globalCurrent == nil {
		return nil
	}
	return globalCurrent.Get(key)
}

func GetString(key string) (string, bool) {
	if globalCurrent == nil {
		return "", false
	}
	return globalCurrent.GetString(key)
}

func GetInt(key string) (int, bool) {
	if globalCurrent == nil {
		return 0, false
	}
	return globalCurrent.GetInt(key)
}

func GetFloat64(key string) (float64, bool) {
	if globalCurrent == nil {
		return 0, false
	}
	return globalCurrent.GetFloat64(key)
}

func GetBool(key string) (bool, bool) {
	if globalCurrent == nil {
		return false, false
	}
	return globalCurrent.GetBool(key)
}

func Delete(key string) {
	if globalCurrent != nil {
		globalCurrent.Delete(key)
	}
}

func Clear() {
	if globalCurrent != nil {
		globalCurrent.Clear()
	}
}

func Exists(key string) bool {
	if globalCurrent == nil {
		return false
	}
	return globalCurrent.Exists(key)
}

func All() map[string]interface{} {
	if globalCurrent == nil {
		return make(map[string]interface{})
	}
	return globalCurrent.All()
}

func Keys() []string {
	if globalCurrent == nil {
		return []string{}
	}
	return globalCurrent.Keys()
}

func SetGlobalCurrent(current *Current) {
	globalCurrent = current
}

func GetGlobalCurrent() *Current {
	if globalCurrent == nil {
		globalCurrent = NewCurrent()
	}
	return globalCurrent
}

func ResetGlobalCurrent() {
	globalCurrent = nil
}
