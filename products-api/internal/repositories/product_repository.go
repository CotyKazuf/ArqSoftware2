package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"products-api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrNotFound indicates that a product was not located in the database.
var ErrNotFound = errors.New("product not found")

// ProductFilter encapsulates optional filters for listing products.
type ProductFilter struct {
	Tipo     string
	Estacion string
	Ocasion  string
	Genero   string
	Marca    string
	Texto    string
}

// Pagination controls paging for list operations.
type Pagination struct {
	Page     int
	PageSize int
}

// ProductRepository defines the data access contract.
type ProductRepository interface {
	Create(p *models.Product) error
	Update(p *models.Product) error
	Delete(id string) error
	FindByID(id string) (*models.Product, error)
	FindAll(filter ProductFilter, pagination Pagination) ([]models.Product, int64, error)
}

// MongoProductRepository persists products in MongoDB.
type MongoProductRepository struct {
	collection *mongo.Collection
}

// NewMongoProductRepository builds a Mongo-backed repository.
func NewMongoProductRepository(collection *mongo.Collection) *MongoProductRepository {
	return &MongoProductRepository{collection: collection}
}

// Create inserts a new product document.
func (r *MongoProductRepository) Create(p *models.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := r.collection.InsertOne(ctx, p)
	if err != nil {
		return fmt.Errorf("insert product: %w", err)
	}

	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		p.ID = oid
	}
	return nil
}

// Update persists changes for an existing product.
func (r *MongoProductRepository) Update(p *models.Product) error {
	if p.ID.IsZero() {
		return errors.New("product id is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": p.ID}
	update := bson.M{
		"$set": bson.M{
			"owner_id":    p.OwnerID,
			"name":        p.Name,
			"descripcion": p.Descripcion,
			"slug":        p.Slug,
			"precio":      p.Precio,
			"stock":       p.Stock,
			"tipo":        p.Tipo,
			"estacion":    p.Estacion,
			"ocasion":     p.Ocasion,
			"notas":       p.Notas,
			"genero":      p.Genero,
			"marca":       p.Marca,
			"imagen":      p.Imagen,
			"tags":        p.Tags,
			"updated_at":  p.UpdatedAt,
		},
	}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a product by ID.
func (r *MongoProductRepository) Delete(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// FindByID locates a product by ID.
func (r *MongoProductRepository) FindByID(id string) (*models.Product, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var product models.Product
	if err := r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&product); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}
	return &product, nil
}

// FindAll returns a filtered, paginated slice of products plus the total count.
func (r *MongoProductRepository) FindAll(filter ProductFilter, pagination Pagination) ([]models.Product, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := buildQuery(filter)
	findOptions := options.Find()
	if pagination.PageSize > 0 {
		findOptions.SetLimit(int64(pagination.PageSize))
		skip := int64((pagination.Page - 1) * pagination.PageSize)
		if skip < 0 {
			skip = 0
		}
		findOptions.SetSkip(skip)
	}
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, query, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("find products: %w", err)
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, 0, fmt.Errorf("decode products: %w", err)
	}

	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	return products, total, nil
}

func buildQuery(filter ProductFilter) bson.M {
	query := bson.M{}
	if filter.Tipo != "" {
		query["tipo"] = filter.Tipo
	}
	if filter.Estacion != "" {
		query["estacion"] = filter.Estacion
	}
	if filter.Ocasion != "" {
		query["ocasion"] = filter.Ocasion
	}
	if filter.Genero != "" {
		query["genero"] = filter.Genero
	}
	if filter.Marca != "" {
		query["marca"] = filter.Marca
	}
	if filter.Texto != "" {
		regex := primitive.Regex{Pattern: filter.Texto, Options: "i"}
		query["$or"] = []bson.M{
			{"name": regex},
			{"descripcion": regex},
		}
	}
	return query
}
