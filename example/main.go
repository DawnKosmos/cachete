package main

import (
	"fmt"
	"github.com/DawnKosmos/bybit-go5"
	"github.com/DawnKosmos/bybit-go5/models"
	"github.com/DawnKosmos/cachete"
	"github.com/dgraph-io/ristretto"
	"time"
)

func expensiveFunction(param1 string, param2 int) (string, error) {
	time.Sleep(2 * time.Second)
	return fmt.Sprintf("Processed: %s, %d", param1, param2), nil
}

func ss() {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		panic(err)
	}

	// set a value with a cost of 1
	cache.Set("key", "value", 1)

}

func main() {
	cache := cachete.NewCache()

	tNow := time.Now()
	var result string
	err := cache.Check(10*time.Second, &result, expensiveFunction, "input", 42)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("First Calculation", time.Since(tNow))

	if err = cache.Check(10*time.Second, &result, expensiveFunction, "input", 42); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Second Calculation", time.Since(tNow))
	fmt.Println("Result:", result)

	by, _ := bybit.New(nil, bybit.URL, nil, false)

	var res models.GetTickersLinearResponse
	parameters := models.GetTickersRequest{
		Category: "linear",
		Symbol:   "BTCUSDT",
	}
	// res, err := by.GetTickersLinear(parameters) this is how the function usually is called

	err = cache.Check(5*time.Second, &res, by.GetTickersLinear, parameters)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(res.List[0].Volume24h)
	err = cache.Check(5*time.Second, &res, by.GetTickersLinear, parameters)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(res.List[0].Volume24h)

}
