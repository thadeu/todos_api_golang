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

// Delete remove um valor do contexto atual
func (c *Current) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// Clear remove todos os valores do contexto atual
func (c *Current) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]interface{})
}

// All retorna todos os dados do contexto atual
func (c *Current) All() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Retorna uma cópia para evitar race conditions
	result := make(map[string]interface{})
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// Keys retorna todas as chaves do contexto atual
func (c *Current) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys
}

// Exists verifica se uma chave existe
func (c *Current) Exists(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.data[key]
	return exists
}

// ContextKey é usado para armazenar o Current no context.Context
type contextKey string

const currentKey contextKey = "current"

// WithCurrent adiciona o Current ao context.Context
func WithCurrent(ctx context.Context, current *Current) context.Context {
	return context.WithValue(ctx, currentKey, current)
}

// FromContext extrai o Current do context.Context
func FromContext(ctx context.Context) (*Current, bool) {
	current, ok := ctx.Value(currentKey).(*Current)
	return current, ok
}

// GetCurrent retorna o Current do contexto ou cria um novo
func GetCurrent(ctx context.Context) *Current {
	if current, ok := FromContext(ctx); ok {
		return current
	}

	return NewCurrent()
}

// SetCurrent define o Current no contexto
func SetCurrent(ctx context.Context, current *Current) context.Context {
	return WithCurrent(ctx, current)
}

// =============================================================================
// FUNÇÕES GLOBAIS - Estilo Rails CurrentAttributes
// =============================================================================

// Variável global para armazenar o Current da requisição atual
var globalCurrent *Current

// Set define um valor no contexto global (estilo Rails)
func Set(key string, value interface{}) {
	if globalCurrent == nil {
		globalCurrent = NewCurrent()
	}
	globalCurrent.Set(key, value)
}

// Get retorna um valor do contexto global (estilo Rails)
func Get(key string) interface{} {
	if globalCurrent == nil {
		return nil
	}
	return globalCurrent.Get(key)
}

// GetString retorna um valor como string (estilo Rails)
func GetString(key string) (string, bool) {
	if globalCurrent == nil {
		return "", false
	}
	return globalCurrent.GetString(key)
}

// GetInt retorna um valor como int (estilo Rails)
func GetInt(key string) (int, bool) {
	if globalCurrent == nil {
		return 0, false
	}
	return globalCurrent.GetInt(key)
}

// GetFloat64 retorna um valor como float64 (estilo Rails)
func GetFloat64(key string) (float64, bool) {
	if globalCurrent == nil {
		return 0, false
	}
	return globalCurrent.GetFloat64(key)
}

// GetBool retorna um valor como bool (estilo Rails)
func GetBool(key string) (bool, bool) {
	if globalCurrent == nil {
		return false, false
	}
	return globalCurrent.GetBool(key)
}

// Delete remove um valor do contexto global (estilo Rails)
func Delete(key string) {
	if globalCurrent != nil {
		globalCurrent.Delete(key)
	}
}

// Clear remove todos os valores do contexto global (estilo Rails)
func Clear() {
	if globalCurrent != nil {
		globalCurrent.Clear()
	}
}

// Exists verifica se uma chave existe no contexto global (estilo Rails)
func Exists(key string) bool {
	if globalCurrent == nil {
		return false
	}
	return globalCurrent.Exists(key)
}

// All retorna todos os dados do contexto global (estilo Rails)
func All() map[string]interface{} {
	if globalCurrent == nil {
		return make(map[string]interface{})
	}
	return globalCurrent.All()
}

// Keys retorna todas as chaves do contexto global (estilo Rails)
func Keys() []string {
	if globalCurrent == nil {
		return []string{}
	}
	return globalCurrent.Keys()
}

// SetGlobalCurrent define o Current global (usado pelos middlewares)
func SetGlobalCurrent(current *Current) {
	globalCurrent = current
}

// GetGlobalCurrent retorna o Current global
func GetGlobalCurrent() *Current {
	if globalCurrent == nil {
		globalCurrent = NewCurrent()
	}
	return globalCurrent
}

// ResetGlobalCurrent limpa o Current global (usado para limpeza entre requisições)
func ResetGlobalCurrent() {
	globalCurrent = nil
}
