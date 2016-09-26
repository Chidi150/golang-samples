// Copyright 2015 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package shopplace

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Ensure memoryDB conforms to the ShopDatabase interface.
var _ ShopDatabase = &memoryDB{}

// memoryDB is a simple in-memory persistence layer for shops.
type memoryDB struct {
	mu     sync.Mutex
	nextID int64           // next ID to assign to a shop.
	shops  map[int64]*Shop // maps from Shop ID to Shop.
}

func newMemoryDB() *memoryDB {
	return &memoryDB{
		shops:  make(map[int64]*Shop),
		nextID: 1,
	}
}

// Close closes the database.
func (db *memoryDB) Close() {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.shops = nil
}

// GetShop retrieves a shop by its ID.
func (db *memoryDB) GetShop(id int64) (*Shop, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	shop, ok := db.shops[id]
	if !ok {
		return nil, fmt.Errorf("memorydb: shop not found with ID %d", id)
	}
	return shop, nil
}

// AddShop saves a given shop, assigning it a new ID.
func (db *memoryDB) AddShop(b *Shop) (id int64, err error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	b.ID = db.nextID
	db.shops[b.ID] = b

	db.nextID++

	return b.ID, nil
}

// DeleteShop removes a given shop by its ID.
func (db *memoryDB) DeleteShop(id int64) error {
	if id == 0 {
		return errors.New("memorydb: shop with unassigned ID passed into deleteShop")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.shops[id]; !ok {
		return fmt.Errorf("memorydb: could not delete shop with ID %d, does not exist", id)
	}
	delete(db.shops, id)
	return nil
}

// UpdateShop updates the entry for a given shop.
func (db *memoryDB) UpdateShop(b *Shop) error {
	if b.ID == 0 {
		return errors.New("memorydb: shop with unassigned ID passed into updateShop")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	db.shops[b.ID] = b
	return nil
}

// shopsByTitle implements sort.Interface, ordering shops by Title.
// https://golang.org/pkg/sort/#example__sortWrapper
type shopsByTitle []*Shop

func (s shopsByTitle) Less(i, j int) bool { return s[i].Title < s[j].Title }
func (s shopsByTitle) Len() int           { return len(s) }
func (s shopsByTitle) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// ListShops returns a list of shops, ordered by title.
func (db *memoryDB) ListShops() ([]*Shop, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var shops []*Shop
	for _, b := range db.shops {
		shops = append(shops, b)
	}

	sort.Sort(shopsByTitle(shops))
	return shops, nil
}

// ListShopsCreatedBy returns a list of shops, ordered by title, filtered by
// the user who created the shop entry.
func (db *memoryDB) ListShopsCreatedBy(userID string) ([]*Shop, error) {
	if userID == "" {
		return db.ListShops()
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	var shops []*Shop
	for _, b := range db.shops {
		if b.CreatedByID == userID {
			shops = append(shops, b)
		}
	}

	sort.Sort(shopsByTitle(shops))
	return shops, nil
}
