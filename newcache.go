package cachete

import (
	"fmt"
	"github.com/DawnKosmos/cachete/expire"
	"hash/fnv"
	"log"
	"reflect"
	"sync"
	"time"
)

type item struct {
	expiration expire.Expirator
	value      any
}

type Cachete struct {
	mu   sync.RWMutex
	data map[int64]item
	tag  map[string][]int64
}

func New() *Cachete {
	return &Cachete{data: make(map[int64]item), tag: make(map[string][]int64)}
}

func (c *Cachete) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.data[stringToInt64Hash(key)]
	//delete when expired
	return item.value, ok
}

func (c *Cachete) Set(e expire.Expirator, key string, value any) {
	c.mu.Lock()
	hash := stringToInt64Hash(key)
	c.data[hash] = item{
		expiration: e,
		value:      value,
	}
	switch v := e.(type) {
	case expire.Tag:
		c.addTag(v.GetValue(), hash)
	case expire.Tags:
		for _, tag := range v.GetValue() {
			c.addTag(tag, hash)
		}
	}

	c.mu.Unlock()
}

func (c *Cachete) addTag(tag string, hash int64) {
	hashs, ok := c.tag[tag]
	if !ok {
		c.tag[tag] = []int64{hash}
	} else {
		c.tag[tag] = append(hashs, hash)
	}
}

func (c *Cachete) Delete(key string) {
	c.mu.Lock()
	delete(c.data, stringToInt64Hash(key))
	c.mu.Unlock()
}

func (c *Cachete) delete(key int64) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}

func (c *Cachete) DeleteWithTag(tag string) {
	tags, ok := c.tag[tag]
	c.mu.Lock()
	if !ok {
		return
	}
	for _, v := range tags {
		delete(c.data, v)
	}
	delete(c.tag, tag)
	c.mu.Unlock()
}

func stringToInt64Hash(s string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return int64(h.Sum64())
}

/*
ExecuteAndCache checks if the result of a given function with specific parameters is already cached.
If it is, the method sets the result to the cached value.
If not, it calls the function with the given parameters, caches the result, and sets it.
*/
func (c *Cachete) ExecuteAndCache(ex expire.Expirator, result any, fn any, parameters ...any) error {
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
		c.Set(ex, hash, res[0].Interface())
		reflectValue(result, res[0])
	}
	return nil
}

func (c *Cachete) Kills() {
	tNow := time.Now()
	for k, v := range c.data {
		if v.expiration.Expire(tNow) {
			c.delete(k)
		}
	}
}
