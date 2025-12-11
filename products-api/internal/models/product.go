package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product models a perfume entry stored in MongoDB.
type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Descripcion string             `bson:"descripcion" json:"descripcion"`
	Precio      float64            `bson:"precio" json:"precio"`
	Stock       int                `bson:"stock" json:"stock"`
	Tipo        string             `bson:"tipo" json:"tipo"`
	Estacion    string             `bson:"estacion" json:"estacion"`
	Ocasion     string             `bson:"ocasion" json:"ocasion"`
	Notas       []string           `bson:"notas" json:"notas"`
	Genero      string             `bson:"genero" json:"genero"`
	Marca       string             `bson:"marca" json:"marca"`
	Imagen      string             `bson:"imagen,omitempty" json:"imagen"`
	OwnerID     string             `bson:"owner_id,omitempty" json:"owner_id"`
	Score       float64            `bson:"score,omitempty" json:"score"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}
