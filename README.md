# GCache

`GCache` is a simple in-memory cache implementation in Go, designed for easy storage and retrieval of items with optional expiration.

## Features

- **In-memory cache**: Stores items in memory for fast access.
- **Expiration**: Items can have an expiration time or be set to never expire.
- **Automatic cleanup**: Optionally run a background routine to clean up expired items.
- **Thread-safe**: Concurrent-safe with proper locking.

## Installation

```sh
go get github.com/tgrziminiar/gcache
```


# Example Usage of `gcache` Package

This example demonstrates how to use the `gcache` package to cache different types of data, including simple values and complex structs.

```go
package main

import (
	"fmt"
	"time"
	"tgrziminiar/gcache/cache"
)

type (
	// User represents a user with an ID and Name.
	User struct {
		ID   int
		Name string
	}

	// Product represents a product with an ID, Title, and Price.
	Product struct {
		ID    int
		Title string
		Price float64
	}

	// Order represents an order with an OrderID, UserID, and a list of Products.
	Order struct {
		OrderID  int
		UserID   int
		Products []int
	}
)

func main() {
	// Create a new cache with a default expiration of 2 seconds and a clear interval of 2 seconds.
	c := cache.NewCache(time.Second*2, time.Second*2)
	key := "foo"
	data := "bar"
	// Set cache with a 1-second expiration
	c.Set(key, data, time.Second*1)

	// Retrieve the value from the cache
	val, found := c.Get(key)
	fmt.Printf("Value for key '%s': %v, Found: %v\n", key, val, found)

	// Store multiple data items under the same key
	key2 := "manydata"
	user := User{ID: 1, Name: "Alice"}
	product := Product{ID: 2, Title: "Smartphone", Price: 699.99}
	order := Order{OrderID: 1, UserID: 1, Products: []int{1, 2}}
	c.Set(key2, []interface{}{user, product, order}, 6*time.Hour)

	// Retrieve and check the stored data
	if cachedData, found := c.Get(key2); found {
		if data, ok := cachedData.([]interface{}); ok && len(data) == 3 {
			if userData, ok := data[0].(User); ok {
				if productData, ok := data[1].(Product); ok {
					if orderData, ok := data[2].(Order); ok {
						fmt.Println("User Data:", userData)
						fmt.Println("Product Data:", productData)
						fmt.Println("Order Data:", orderData)
					}
				}
			}
		}
	}
}
