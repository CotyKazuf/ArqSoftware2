package repositories

import (
	"context"
	"fmt"
	"time"

	"products-api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PurchaseRepository defines operations for persisting purchases.
type PurchaseRepository interface {
	Create(purchase *models.Purchase) error
	FindByUserID(userID string) ([]models.Purchase, error)
}

// MongoPurchaseRepository stores purchases inside MongoDB.
type MongoPurchaseRepository struct {
	collection *mongo.Collection
}

// NewMongoPurchaseRepository builds a repository backed by a Mongo collection.
func NewMongoPurchaseRepository(collection *mongo.Collection) *MongoPurchaseRepository {
	return &MongoPurchaseRepository{collection: collection}
}

// Create persists a purchase record.
func (r *MongoPurchaseRepository) Create(purchase *models.Purchase) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := r.collection.InsertOne(ctx, purchase)
	if err != nil {
		return fmt.Errorf("insert purchase: %w", err)
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		purchase.ID = oid
	}
	return nil
}

// FindByUserID returns purchases for a given user sorted by newest first.
func (r *MongoPurchaseRepository) FindByUserID(userID string) ([]models.Purchase, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "fecha_compra", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("find purchases: %w", err)
	}
	defer cursor.Close(ctx)

	var purchases []models.Purchase
	if err := cursor.All(ctx, &purchases); err != nil {
		return nil, fmt.Errorf("decode purchases: %w", err)
	}

	return purchases, nil
}
