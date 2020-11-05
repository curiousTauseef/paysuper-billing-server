package repository

import (
	"context"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CustomerRepositoryInterface is abstraction layer for working with customer and representation in database.
type CustomerRepositoryInterface interface {
	// Insert adds the customer to the collection.
	Insert(context.Context, *billingpb.Customer) error

	// Update updates the customer in the collection.
	Update(context.Context, *billingpb.Customer) error

	// GetById returns the customer by unique identity.
	GetById(context.Context, string) (*billingpb.Customer, error)

	// Find return customer by merchant id and token user (user id or email or phone).
	Find(context.Context, string, *billingpb.TokenUser) (*billingpb.Customer, error)

	//Return all customers
	FindAll(ctx context.Context) ([]*billingpb.Customer, error)

	//FindBy returns customers filter by query
	FindBy(ctx context.Context, query bson.M, opts ...*options.FindOptions) ([]*billingpb.Customer, error)

	//CountBy returns count of customers with filter by query
	CountBy(ctx context.Context, query bson.M) (int64, error)
}
