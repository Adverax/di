package di

import (
	"context"
	"fmt"
	"sync"
)

type Variables interface {
	Get(key string) interface{}
	GetInt(key string) int
	GetString(key string) string
	GetBool(key string) bool
	GetFloat(key string) float64

	TryGet(key string, defVal interface{}) interface{}
	TryGetInt(key string, defVal int) int
	TryGetString(key string, defVal string) string
	TryGetBool(key string, defVal bool) bool
	TryGetFloat(key string, defVal float64) float64

	Set(key string, val interface{})
	SetInt(key string, val int)
	SetString(key string, val string)
	SetBool(key string, val bool)
	SetFloat(key string, val float64)
}

type variables struct {
	sync.RWMutex
	values map[string]interface{}
}

func (v *variables) Get(key string) interface{} {
	v.RLock()
	defer v.RUnlock()

	if val, ok := v.values[key]; ok {
		return val
	}
	panic(fmt.Errorf("Variable not found: %s", key))
}

func (v *variables) GetInt(key string) int {
	v.RLock()
	defer v.RUnlock()

	if val, ok := v.values[key]; ok {
		if v, ok := val.(int); ok {
			return v
		}
		panic(fmt.Errorf("Type mismatch for variable: %s", key))
	}
	panic(fmt.Errorf("Variable not found: %s", key))
}

func (v *variables) GetString(key string) string {
	v.RLock()
	defer v.RUnlock()

	if val, ok := v.values[key]; ok {
		if v, ok := val.(string); ok {
			return v
		}
		panic(fmt.Errorf("Type mismatch for variable: %s", key))
	}
	panic(fmt.Errorf("Variable not found: %s", key))
}

func (v *variables) GetBool(key string) bool {
	v.RLock()
	defer v.RUnlock()

	if val, ok := v.values[key]; ok {
		if v, ok := val.(bool); ok {
			return v
		}
		panic(fmt.Errorf("Type mismatch for variable: %s", key))
	}
	panic(fmt.Errorf("Variable not found: %s", key))
}

func (v *variables) GetFloat(key string) float64 {
	v.RLock()
	defer v.RUnlock()

	if val, ok := v.values[key]; ok {
		if v, ok := val.(float64); ok {
			return v
		}
		panic(fmt.Errorf("Type mismatch for variable: %s", key))
	}
	panic(fmt.Errorf("Variable not found: %s", key))
}

func (v *variables) TryGet(key string, defVal interface{}) interface{} {
	v.RLock()
	defer v.RUnlock()

	if val, ok := v.values[key]; ok {
		if v, ok := val.(int); ok {
			return v
		}
	}
	return defVal
}

func (v *variables) TryGetInt(key string, defVal int) int {
	v.RLock()
	defer v.RUnlock()

	if val, ok := v.values[key]; ok {
		if v, ok := val.(int); ok {
			return v
		}
	}
	return defVal
}

func (v *variables) TryGetString(key string, defVal string) string {
	v.RLock()
	defer v.RUnlock()

	if val, ok := v.values[key]; ok {
		if v, ok := val.(string); ok {
			return v
		}
	}
	return defVal
}

func (v *variables) TryGetBool(key string, defVal bool) bool {
	v.RLock()
	defer v.RUnlock()

	if val, ok := v.values[key]; ok {
		if v, ok := val.(bool); ok {
			return v
		}
	}
	return defVal
}

func (v *variables) TryGetFloat(key string, defVal float64) float64 {
	v.RLock()
	defer v.RUnlock()

	if val, ok := v.values[key]; ok {
		if v, ok := val.(float64); ok {
			return v
		}
	}
	return defVal
}

func (v *variables) Set(key string, val interface{}) {
	v.Lock()
	defer v.Unlock()

	v.values[key] = val
}

func (v *variables) SetInt(key string, val int) {
	v.Lock()
	defer v.Unlock()

	v.values[key] = val
}

func (v *variables) SetString(key string, val string) {
	v.Lock()
	defer v.Unlock()

	v.values[key] = val
}

func (v *variables) SetBool(key string, val bool) {
	v.Lock()
	defer v.Unlock()

	v.values[key] = val
}

func (v *variables) SetFloat(key string, val float64) {
	v.Lock()
	defer v.Unlock()

	v.values[key] = val
}

func NewVariables(values map[string]interface{}) Variables {
	if values == nil {
		values = make(map[string]interface{})
	}
	return &variables{
		values: values,
	}
}

func GetVariables(ctx context.Context) Variables {
	app := GetAppFromContext(ctx)
	return app.Variables()
}
