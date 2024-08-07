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
