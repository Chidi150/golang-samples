// Copyright 2015 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package shopplace

// Shop holds metadata about a shop.
type Shop struct {
	ID            int64
	Title         string
	Author        string
	PublishedDate string
	ImageURL      string
	Description   string
	CreatedBy     string
	CreatedByID   string
}

// CreatedByDisplayName returns a string appropriate for displaying the name of
// the user who created this shop object.
func (b *Shop) CreatedByDisplayName() string {
	if b.CreatedByID == "anonymous" {
		return "Anonymous"
	}
	return b.CreatedBy
}

// SetCreatorAnonymous sets the CreatedByID field to the "anonymous" ID.
func (b *Shop) SetCreatorAnonymous() {
	b.CreatedBy = ""
	b.CreatedByID = "anonymous"
}

// ShopDatabase provides thread-safe access to a database of shops.
type ShopDatabase interface {
	// ListShops returns a list of shops, ordered by title.
	ListShops() ([]*Shop, error)

	// ListShopsCreatedBy returns a list of shops, ordered by title, filtered by
	// the user who created the shop entry.
	ListShopsCreatedBy(userID string) ([]*Shop, error)

	// GetShop retrieves a shop by its ID.
	GetShop(id int64) (*Shop, error)

	// AddShop saves a given shop, assigning it a new ID.
	AddShop(b *Shop) (id int64, err error)

	// DeleteShop removes a given shop by its ID.
	DeleteShop(id int64) error

	// UpdateShop updates the entry for a given shop.
	UpdateShop(b *Shop) error

	// Close closes the database, freeing up any available resources.
	// TODO(cbro): Close() should return an error.
	Close()
}
