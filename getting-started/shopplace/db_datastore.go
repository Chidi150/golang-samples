// Copyright 2015 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package shopplace

import (
	"fmt"

	"cloud.google.com/go/datastore"

	"golang.org/x/net/context"
)

// datastoreDB persists shops to Cloud Datastore.
// https://cloud.google.com/datastore/docs/concepts/overview
type datastoreDB struct {
	client *datastore.Client
}

// Ensure datastoreDB conforms to the ShopDatabase interface.
var _ ShopDatabase = &datastoreDB{}

// newDatastoreDB creates a new ShopDatabase backed by Cloud Datastore.
// See the datastore and google packages for details on creating a suitable Client:
// https://godoc.org/cloud.google.com/go/datastore
func newDatastoreDB(client *datastore.Client) (ShopDatabase, error) {
	ctx := context.Background()
	// Verify that we can communicate and authenticate with the datastore service.
	t, err := client.NewTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("datastoredb: could not connect: %v", err)
	}
	if err := t.Rollback(); err != nil {
		return nil, fmt.Errorf("datastoredb: could not connect: %v", err)
	}
	return &datastoreDB{
		client: client,
	}, nil
}

// Close closes the database.
func (db *datastoreDB) Close() {
	// No op.
}

func (db *datastoreDB) datastoreKey(id int64) *datastore.Key {
	ctx := context.Background()
	return datastore.NewKey(ctx, "Shop", "", id, nil)
}

// GetShop retrieves a shop by its ID.
func (db *datastoreDB) GetShop(id int64) (*Shop, error) {
	ctx := context.Background()
	k := db.datastoreKey(id)
	shop := &Shop{}
	if err := db.client.Get(ctx, k, shop); err != nil {
		return nil, fmt.Errorf("datastoredb: could not get Shop: %v", err)
	}
	shop.ID = id
	return shop, nil
}

// AddShop saves a given shop, assigning it a new ID.
func (db *datastoreDB) AddShop(b *Shop) (id int64, err error) {
	ctx := context.Background()
	k := datastore.NewIncompleteKey(ctx, "Shop", nil)
	k, err = db.client.Put(ctx, k, b)
	if err != nil {
		return 0, fmt.Errorf("datastoredb: could not put Shop: %v", err)
	}
	return k.ID(), nil
}

// DeleteShop removes a given shop by its ID.
func (db *datastoreDB) DeleteShop(id int64) error {
	ctx := context.Background()
	k := db.datastoreKey(id)
	if err := db.client.Delete(ctx, k); err != nil {
		return fmt.Errorf("datastoredb: could not delete Shop: %v", err)
	}
	return nil
}

// UpdateShop updates the entry for a given shop.
func (db *datastoreDB) UpdateShop(b *Shop) error {
	ctx := context.Background()
	k := db.datastoreKey(b.ID)
	if _, err := db.client.Put(ctx, k, b); err != nil {
		return fmt.Errorf("datastoredb: could not update Shop: %v", err)
	}
	return nil
}

// ListShops returns a list of shops, ordered by title.
func (db *datastoreDB) ListShops() ([]*Shop, error) {
	ctx := context.Background()
	shops := make([]*Shop, 0)
	q := datastore.NewQuery("Shop").
		Order("Title")

	keys, err := db.client.GetAll(ctx, q, &shops)

	if err != nil {
		return nil, fmt.Errorf("datastoredb: could not list shops: %v", err)
	}

	for i, k := range keys {
		shops[i].ID = k.ID()
	}

	return shops, nil
}

// ListShopsCreatedBy returns a list of shops, ordered by title, filtered by
// the user who created the shop entry.
func (db *datastoreDB) ListShopsCreatedBy(userID string) ([]*Shop, error) {
	ctx := context.Background()
	if userID == "" {
		return db.ListShops()
	}

	shops := make([]*Shop, 0)
	q := datastore.NewQuery("Shop").
		Filter("CreatedByID =", userID).
		Order("Title")

	keys, err := db.client.GetAll(ctx, q, &shops)

	if err != nil {
		return nil, fmt.Errorf("datastoredb: could not list shops: %v", err)
	}

	for i, k := range keys {
		shops[i].ID = k.ID()
	}

	return shops, nil
}
