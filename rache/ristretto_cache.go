package rache

import (
	"fmt"
	"github.com/dgraph-io/ristretto"
	"log"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type Rache struct {
	cache *ristretto.Cache
}

type CashSize *ristretto.Config

var (
	SMALL CashSize = &ristretto.Config{
		NumCounters: 1000,
		MaxCost:     1 << 10,
		BufferItems: 64,
	}
	MEDIUM CashSize = &ristretto.Config{
		NumCounters: 100000,
		MaxCost:     1 << 20,
		BufferItems: 64,
	}
	BIG CashSize = &ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	}
)

func New(size CashSize) (*Rache, error) {
	cache, err := ristretto.NewCache(size)
	if err != nil {
		return nil, err
	}
	return &Rache{cache: cache}, nil
}

func (r *Rache) Set(key any, value any, ttl time.Duration) bool {
	return r.cache.SetWithTTL(key, value, 0, ttl)
}

func (r *Rache) Get(key any) (any, bool) {
	return r.cache.Get(key)
}

func (r *Rache) Delete(key any) {
	r.cache.Del(key)
}

/*
FuncCache checks if the result of a given function with specific parameters is already cached.
If it is, the method sets the result to the cached value.
If not, it calls the function with the given parameters, caches the result, and sets it.
*/
func (r *Rache) FuncCache(fn any, result any, ttl time.Duration, parameter ...any) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v \n", r)
			return
		}
	}()
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("fn must be a function")
	}
	key := toString(fn, parameter)

	if v, ok := r.Get(key); ok {
		reflectValue(result, reflect.ValueOf(v))
	} else {
		// If the result is not cached, call the function with the given parameters.
		res := reflect.ValueOf(fn).Call(convertToValue(parameter))
		// Check if the function returned an error and return it if it did.
		if err, ok := res[1].Interface().(error); ok {
			return err
		}
		// Cache the result for the specified duration and set the result value.
		r.Set(key, res[0].Interface(), ttl)
		reflectValue(result, res[0])
	}
	return nil
}

func toString(fn any, parameters []any) string {
	fnName := getFunctionName(fn)
	hash := createParameterHash(fnName, parameters)
	return hash
}

func getFunctionName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

func createParameterHash(fnName string, parameters []interface{}) string {
	var paramBuilder strings.Builder
	paramBuilder.WriteString(fnName)
	for _, param := range parameters {
		paramBuilder.WriteString(fmt.Sprintf("_%v", param))
	}
	return paramBuilder.String()
}

func reflectValue(result any, value reflect.Value) {
	if value.Kind() == reflect.Ptr {
		reflect.ValueOf(result).Elem().Set(value.Elem())
	} else {
		reflect.ValueOf(result).Elem().Set(value)
	}
}

func convertToValue(params []interface{}) []reflect.Value {
	values := make([]reflect.Value, len(params))
	for i, param := range params {
		values[i] = reflect.ValueOf(param)
	}
	return values
}
