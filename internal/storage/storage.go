package storage

import (
	"fmt"
	"sync"
)

type InMemoryDB struct {
	data sync.Map
}

func NewInMemoryDB() *InMemoryDB {
	db := InMemoryDB{
		data: sync.Map{},
	}
	return &db
}

func (db *InMemoryDB) WriteValue(key string, value string) {
	db.data.Store(key, value)
}

func (db *InMemoryDB) GetValue(key string) (string, bool) {
	value, ok := db.data.Load(key)
	return fmt.Sprintf("%v", value), ok
}
