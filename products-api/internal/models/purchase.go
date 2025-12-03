package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PurchaseItem stores a snapshot of a product included in a purchase.
type PurchaseItem struct {
	ProductID      primitive.ObjectID `bson:"product_id" json:"product_id"`
	Nombre         string             `bson:"nombre" json:"nombre"`
	Marca          string             `bson:"marca" json:"marca"`
	Imagen         string             `bson:"imagen" json:"imagen"`
	PrecioUnitario float64            `bson:"precio_unitario" json:"precio_unitario"`
	Cantidad       int                `bson:"cantidad" json:"cantidad"`
}

// Purchase describes a completed checkout operation.
type Purchase struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	FechaCompra time.Time          `bson:"fecha_compra" json:"fecha_compra"`
	Total       float64            `bson:"total" json:"total"`
	Items       []PurchaseItem     `bson:"items" json:"items"`
}
