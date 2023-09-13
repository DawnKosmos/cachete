# Cachete
This Go library provides a simple way to cache the results of function calls with specific parameters for a specified duration. It can be useful for reducing the number of repeated calculations or expensive operations by storing their results and reusing them when needed.
![img.png](img.png)

### Install Packete
```bash
go get -u github.com/DawnKosmos/cachete
```
### Crate New Cache
```GO
package main

import (
	"fmt"
	"github.com/DawnKosmos/cachete"
	"time"
)

//Functions or Methods need to have 2 return Values. 1st the Result 2nd is an error
func expensiveFunction(param1 string, param2 int) (string, error) {
	time.Sleep(2 * time.Second)
	return fmt.Sprintf("Processed: %s, %d", param1, param2), nil
}

func main() {
	cache := cachete.NewCache()

	tNow := time.Now()
	var result string
	err := cache.Check(10*time.Second, &result, expensiveFunction, "input", 42)
	if err != nil {
		fmt.Println("Er3ror:", err)
		return
	}
	fmt.Println("First Calculation", time.Since(tNow)) //First Calculation 2.001130973s
	if err = cache.Check(10*time.Second, &result, expensiveFunction, "input", 42); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Second Calculation", time.Since(tNow)) //Second Calculation 2.001160585s
	fmt.Println("Result:", result) //Result: Processed: input, 42
```


## Example with an API

```go
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
```