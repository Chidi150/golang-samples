// Copyright 2015 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package shopplace

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type mongoDB struct {
	conn *mgo.Session
	c    *mgo.Collection
}

// Ensure mongoDB conforms to the ShopDatabase interface.
var _ ShopDatabase = &mongoDB{}

// newMongoDB creates a new ShopDatabase backed by a given Mongo server,
// authenticated with given credentials.
func newMongoDB(addr string, cred *mgo.Credential) (ShopDatabase, error) {
	conn, err := mgo.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("mongo: could not dial: %v", err)
	}

	if cred != nil {
		if err := conn.Login(cred); err != nil {
			return nil, err
		}
	}

	return &mongoDB{
		conn: conn,
		c:    conn.DB("shopplace").C("shops"),
	}, nil
}

// Close closes the database.
func (db *mongoDB) Close() {
	db.conn.Close()
}

// GetShop retrieves a shop by its ID.
func (db *mongoDB) GetShop(id int64) (*Shop, error) {
	b := &Shop{}
	if err := db.c.Find(bson.D{{Name: "id", Value: id}}).One(b); err != nil {
		return nil, err
	}
	return b, nil
}

var maxRand = big.NewInt(1<<63 - 1)

// randomID returns a positive number that fits within an int64.
func randomID() (int64, error) {
	// Get a random number within the range [0, 1<<63-1)
	n, err := rand.Int(rand.Reader, maxRand)
	if err != nil {
		return 0, err
	}
	// Don't assign 0.
	return n.Int64() + 1, nil
}

// AddShop saves a given shop, assigning it a new ID.
func (db *mongoDB) AddShop(b *Shop) (id int64, err error) {
	id, err = randomID()
	if err != nil {
		return 0, fmt.Errorf("mongodb: could not assign an new ID: %v", err)
	}

	b.ID = id
	if err := db.c.Insert(b); err != nil {
		return 0, fmt.Errorf("mongodb: could not add shop: %v", err)
	}
	return id, nil
}

// DeleteShop removes a given shop by its ID.
func (db *mongoDB) DeleteShop(id int64) error {
	return db.c.Remove(bson.D{{Name: "id", Value: id}})
}

// UpdateShop updates the entry for a given shop.
func (db *mongoDB) UpdateShop(b *Shop) error {
	return db.c.Update(bson.D{{Name: "id", Value: b.ID}}, b)
}

// ListShops returns a list of shops, ordered by title.
func (db *mongoDB) ListShops() ([]*Shop, error) {
	var result []*Shop
	if err := db.c.Find(nil).Sort("title").All(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListShopsCreatedBy returns a list of shops, ordered by title, filtered by
// the user who created the shop entry.
func (db *mongoDB) ListShopsCreatedBy(userID string) ([]*Shop, error) {
	var result []*Shop
	if err := db.c.Find(bson.D{{Name: "createdbyid", Value: userID}}).Sort("title").All(&result); err != nil {
		return nil, err
	}
	return result, nil
}
