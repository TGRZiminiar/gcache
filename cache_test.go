package gcache

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	c := NewGCache(DefaultExpiration)
	if c == nil {
		t.Errorf("NewGCache returned value is not of type Cache")
	}
}

func TestSetWithCustomExpiration(t *testing.T) {
	c := NewGCache(DefaultExpiration)

	key := "foo"
	data := "bar"
	expiration := 10 * time.Second
	c.Set(key, data, expiration)

	val, ok := c.Get(key)
	if !ok {
		t.Errorf("can't find key name")
	}

	if val != data {
		t.Errorf("data doesn't match")
	}

	// Check if the expiration time is set correctly
	item, ok := c.items[key]
	if !ok {
		t.Errorf("can't find key in cache")
	}

	if item.Expiration == 0 {
		t.Errorf("expiration time is not set")
	}
}

func TestSetWithDefaultExpiration(t *testing.T) {
	c := NewGCache(10 * time.Second)

	key := "foo"
	data := "bar"
	c.Set(key, data, DefaultExpiration)

	val, ok := c.Get(key)
	if !ok {
		t.Errorf("can't find key name")
	}

	if val != data {
		t.Errorf("data doesn't match")
	}

	// Check if the expiration time is set correctly
	item, ok := c.items[key]
	if !ok {
		t.Errorf("can't find key in cache")
	}

	if item.Expiration == 0 {
		t.Errorf("expiration time is not set")
	}
}

func TestSetWithNoExpiration(t *testing.T) {
	c := NewGCache(DefaultExpiration)

	key := "foo"
	data := "bar"
	c.Set(key, data, NoExpiration)

	val, ok := c.Get(key)
	if !ok {
		t.Errorf("can't find key name")
	}

	if val != data {
		t.Errorf("data doesn't match")
	}

	// Check if the expiration time is set correctly
	item, ok := c.items[key]
	if !ok {
		t.Errorf("can't find key in cache")
	}

	if item.Expiration != 0 {
		t.Errorf("expiration time is set")
	}
}

func TestSetConcurrent(t *testing.T) {
	c := NewGCache(DefaultExpiration)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("foo%d", i)
			data := fmt.Sprintf("bar%d", i)
			c.Set(key, data, NoExpiration)
		}(i)
	}

	wg.Wait()

	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("foo%d", i)
		data := fmt.Sprintf("bar%d", i)
		val, ok := c.Get(key)
		if !ok {
			t.Errorf("can't find key name")
		}

		if val != data {
			t.Errorf("data doesn't match")
		}
	}
}

func TestDelete(t *testing.T) {
	c := NewGCache(DefaultExpiration)

	key := "foo"
	data := "bar"
	c.Set(key, data, NoExpiration)

	// Check if the key exists before deleting
	val, ok := c.Get(key)
	if !ok {
		t.Errorf("can't find key name")
	}

	if val != data {
		t.Errorf("data doesn't match")
	}

	// Delete the key
	c.Delete(key)

	// Check if the key is deleted
	_, ok = c.Get(key)
	if ok {
		t.Errorf("key is not deleted")
	}
}

func TestDeleteNonExistingKey(t *testing.T) {
	c := NewGCache(DefaultExpiration)

	key := "foo"

	// Delete a non-existing key
	c.Delete(key)

	// Check if the cache is still intact
	if len(c.items) != 0 {
		t.Errorf("cache is not empty")
	}
}

func TestDeleteConcurrent(t *testing.T) {
	c := NewGCache(DefaultExpiration)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("foo%d", i)
			data := fmt.Sprintf("bar%d", i)
			c.Set(key, data, NoExpiration)
		}(i)
	}

	wg.Wait()

	// Delete all keys concurrently
	var wgDelete sync.WaitGroup
	for i := 0; i < 10; i++ {
		wgDelete.Add(1)
		go func(i int) {
			defer wgDelete.Done()
			key := fmt.Sprintf("foo%d", i)
			c.Delete(key)
		}(i)
	}

	wgDelete.Wait()

	// Check if all keys are deleted
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("foo%d", i)
		_, ok := c.Get(key)
		if ok {
			t.Errorf("key is not deleted")
		}
	}
}
func TestSetDefault(t *testing.T) {
	c := NewCache(time.Minute, 0)
	key := "foo"
	val := "bar"

	c.SetDefault(key, val)
	got, ok := c.Get(key)
	if !ok || got != val {
		t.Errorf("expected %v, got %v", val, got)
	}
}
func TestGetWithExpiration(t *testing.T) {
	c := NewCache(DefaultExpiration, 0)
	key := "foo"
	val := "bar"
	exp := time.Now().Add(time.Minute)

	// Test with a specific expiration time
	c.Set(key, val, time.Minute)
	got, expTime, ok := c.GetWithExpiration(key)
	if !ok || got != val {
		t.Errorf("expected %v, got %v", val, got)
	}

	if expTime.Before(time.Now()) || expTime.After(exp) {
		t.Errorf("expected expiration before %v, got %v", exp, expTime)
	}

	// Test with NoExpiration
	c.Set(key, val, NoExpiration)
	got, expTime, ok = c.GetWithExpiration(key)
	if !ok || got != val {
		t.Errorf("expected %v, got %v", val, got)
	}

	if !expTime.IsZero() {
		t.Errorf("expected expiration time to be zero, got %v", expTime)
	}
}

func TestDeleteExpired(t *testing.T) {
	c := NewCache(DefaultExpiration, 0)
	key1 := "foo1"
	val1 := "bar1"
	key2 := "foo2"
	val2 := "bar2"
	key3 := "foo3"
	val3 := "bar3"

	// Set items with different expiration times
	c.Set(key1, val1, time.Millisecond*50)  // Expires first
	c.Set(key2, val2, time.Millisecond*150) // Expires second
	c.Set(key3, val3, NoExpiration)         // Does not expire

	// Wait for the first item to expire
	time.Sleep(time.Millisecond * 100)
	c.DeleteExpired()

	// Check that only the first item has been deleted
	if _, ok := c.Get(key1); ok {
		t.Errorf("expected key1 to be expired and deleted")
	}
	if _, ok := c.Get(key2); !ok {
		t.Errorf("expected key2 to be present")
	}
	if _, ok := c.Get(key3); !ok {
		t.Errorf("expected key3 to be present")
	}

	// Wait for the second item to expire
	time.Sleep(time.Millisecond * 100)
	c.DeleteExpired()

	// Check that the second item has now been deleted
	if _, ok := c.Get(key2); ok {
		t.Errorf("expected key2 to be expired and deleted")
	}
	if _, ok := c.Get(key3); !ok {
		t.Errorf("expected key3 to be present")
	}
}

func TestDeleteExpired_NoExpiration(t *testing.T) {
	c := NewCache(DefaultExpiration, 0)
	key := "foo"
	val := "bar"

	// Set an item with NoExpiration
	c.Set(key, val, NoExpiration)

	// Wait some time and call DeleteExpired
	time.Sleep(time.Millisecond * 100)
	c.DeleteExpired()

	// Check that the item with NoExpiration is still present
	if _, ok := c.Get(key); !ok {
		t.Errorf("expected key to be present")
	}
}

func TestDeleteExpired_EmptyCache(t *testing.T) {
	c := NewCache(DefaultExpiration, 0)

	// Call DeleteExpired on an empty cache
	c.DeleteExpired()

	// Ensure no panic occurs and cache is still empty
	if len(c.items) != 0 {
		t.Errorf("expected cache to be empty")
	}
}

type User struct {
	ID   int
	Name string
}

type Product struct {
	ID    int
	Title string
	Price float64
}

type Order struct {
	OrderID  int
	UserID   int
	Products []int
}

func TestSetGet_MixedTypes(t *testing.T) {
	c := NewCache(DefaultExpiration, 0)
	items := []interface{}{
		User{ID: 1, Name: "Alice"},
		User{ID: 2, Name: "Bob"},
		Product{ID: 1, Title: "Laptop", Price: 1299.99},
		Product{ID: 2, Title: "Smartphone", Price: 699.99},
		Order{OrderID: 1, UserID: 1, Products: []int{1, 2}},
	}

	keys := []string{"user1", "user2", "product1", "product2", "order1"}

	// Set items in the cache
	for i, item := range items {
		c.Set(keys[i], item, DefaultExpiration)
	}

	// Get and check items from the cache
	for i, key := range keys {
		got, ok := c.Get(key)
		if !ok {
			t.Errorf("expected key %s to be present", key)
		}
		if !reflect.DeepEqual(got, items[i]) {
			t.Errorf("expected %v, got %v", items[i], got)
		}
	}

	// Check type assertions
	user1, ok := c.Get("user1")
	if !ok {
		t.Errorf("expected user1 to be present")
	}
	if _, ok := user1.(User); !ok {
		t.Errorf("expected type User, got %T", user1)
	}

	product1, ok := c.Get("product1")
	if !ok {
		t.Errorf("expected product1 to be present")
	}
	if _, ok := product1.(Product); !ok {
		t.Errorf("expected type Product, got %T", product1)
	}

	order1, ok := c.Get("order1")
	if !ok {
		t.Errorf("expected order1 to be present")
	}
	if _, ok := order1.(Order); !ok {
		t.Errorf("expected type Order, got %T", order1)
	}
}

func TestSetGetMixedTypesWithExpiration(t *testing.T) {
	c := NewCache(DefaultExpiration, 0)
	items := []interface{}{
		User{ID: 1, Name: "Alice"},
		Product{ID: 1, Title: "Laptop", Price: 1299.99},
		Order{OrderID: 1, UserID: 1, Products: []int{1}},
	}

	keys := []string{"user1", "product1", "order1"}

	// Set items in the cache with different expiration times
	c.Set(keys[0], items[0], time.Millisecond*50)
	c.Set(keys[1], items[1], time.Millisecond*100)
	c.Set(keys[2], items[2], NoExpiration)

	// Wait and check expiration
	time.Sleep(time.Millisecond * 60)
	c.DeleteExpired()

	// The first item should be expired
	if _, ok := c.Get(keys[0]); ok {
		t.Errorf("expected key %s to be expired", keys[0])
	}

	// The second and third items should be present
	if _, ok := c.Get(keys[1]); !ok {
		t.Errorf("expected key %s to be present", keys[1])
	}
	if _, ok := c.Get(keys[2]); !ok {
		t.Errorf("expected key %s to be present", keys[2])
	}
}
func TestItems(t *testing.T) {
	// Create a new cache with a default expiration
	c := NewCache(DefaultExpiration, 0)

	// Define test items
	user := User{ID: 1, Name: "Alice"}
	product := Product{ID: 2, Title: "Smartphone", Price: 699.99}
	order := Order{OrderID: 1, UserID: 1, Products: []int{1, 2}}

	// Set items in the cache with different expiration times
	c.Set("user", user, 5*time.Second)
	c.Set("product", product, 10*time.Second)
	c.Set("order", order, -1) // No expiration

	// Wait for a short period to ensure some items have not yet expired
	time.Sleep(1 * time.Second)

	// Retrieve items from cache
	items := c.Items()

	// Check that all items are returned
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}

	// Check individual items using reflect.DeepEqual
	if item, ok := items["user"]; !ok || !reflect.DeepEqual(item.Object, user) {
		t.Errorf("expected 'user' item to be %v, got %v", user, item.Object)
	}

	if item, ok := items["product"]; !ok || !reflect.DeepEqual(item.Object, product) {
		t.Errorf("expected 'product' item to be %v, got %v", product, item.Object)
	}

	if item, ok := items["order"]; !ok || !reflect.DeepEqual(item.Object, order) {
		t.Errorf("expected 'order' item to be %v, got %v", order, item.Object)
	}

	// Wait for the 'user' item to expire
	time.Sleep(5 * time.Second)

	// Retrieve items from cache again
	items = c.Items()

	// Check that the expired item is no longer in the cache
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}

	if _, ok := items["user"]; ok {
		t.Error("expected 'user' item to be expired")
	}

	if item, ok := items["product"]; !ok || !reflect.DeepEqual(item.Object, product) {
		t.Errorf("expected 'product' item to be %v, got %v", product, item.Object)
	}

	if item, ok := items["order"]; !ok || !reflect.DeepEqual(item.Object, order) {
		t.Errorf("expected 'order' item to be %v, got %v", order, item.Object)
	}
}
