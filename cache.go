package cachete

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Item struct {
	expirationTime time.Time
	value          any
}

type Cache struct {
	mu   sync.RWMutex
	data map[string]Item
	tag  map[string]int64
}

func NewCache() *Cache {
	return &Cache{data: make(map[string]Item)}
}

func (c *Cache) Set(validFor time.Duration, key string, value any) {
	c.mu.Lock()
	c.data[key] = Item{
		value:          value,
		expirationTime: time.Now().Add(validFor),
	}
	c.mu.Unlock()
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if item, ok := c.data[key]; ok {
		if time.Now().Before(item.expirationTime) {
			return item.value, true
		} else {
			c.Delete("key")
		}

	}
	return nil, false
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}

// Kills Deletes are expired Items
func (c *Cache) Kills() {
	tNow := time.Now()
	for k, v := range c.data {
		if v.expirationTime.Before(tNow) {
			c.Delete(k)
		}
	}
}

func AutoCleaner(cachete *Cache, checkInterval time.Duration) (stop chan struct{}) {
	t := time.NewTicker(checkInterval)
	stop = make(chan struct{})
	clean(cachete, t, stop)
	return stop
}

func clean(cachete *Cache, t *time.Ticker, stop chan struct{}) {
	for {
		select {
		case <-t.C:
			cachete.Kills()
		case <-stop:
			break
		}
	}
}

/*
Check checks if the result of a given function with specific parameters is already cached.
If it is, the method sets the result to the cached value.
If not, it calls the function with the given parameters, caches the result, and sets it.
*/
func (c *Cache) Check(validFor time.Duration, result any, fn any, parameters ...any) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v \n", r)
			return
		}
	}()
	// Check if the provided fn argument is a function. Return an error if it's not.
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("fn must be a function")
	}

	hash := createHash(fn, parameters)

	// If the result is already cached, set the result value to the cached value.
	if value, ok := c.Get(hash); ok {
		reflectValue(result, reflect.ValueOf(value))
	} else {
		// If the result is not cached, call the function with the given parameters.
		res := reflect.ValueOf(fn).Call(convertToValue(parameters))
		// Check if the function returned an error and return it if it did.
		if err, ok := res[1].Interface().(error); ok {
			return err
		}
		// Cache the result for the specified duration and set the result value.
		c.Set(validFor, hash, res[0].Interface())
		reflectValue(result, res[0])
	}
	return nil
}

func reflectValue(result any, value reflect.Value) {
	if value.Kind() == reflect.Ptr {
		reflect.ValueOf(result).Elem().Set(value.Elem())
	} else {
		reflect.ValueOf(result).Elem().Set(value)
	}
}

func createHash(fn interface{}, parameters []interface{}) string {
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

	hash := sha256.New()
	hash.Write([]byte(paramBuilder.String()))
	return hex.EncodeToString(hash.Sum(nil))
}

func convertToValue(params []interface{}) []reflect.Value {
	values := make([]reflect.Value, len(params))
	for i, param := range params {
		values[i] = reflect.ValueOf(param)
	}
	return values
}
