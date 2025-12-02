package models

import "time"

// ProductDocument represents how a product is indexed in Solr and exposed via search-api.
type ProductDocument struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Descripcion string    `json:"descripcion"`
	Precio      float64   `json:"precio"`
	Stock       int       `json:"stock"`
	Tipo        string    `json:"tipo"`
	Estacion    string    `json:"estacion"`
	Ocasion     string    `json:"ocasion"`
	Notas       []string  `json:"notas"`
	Genero      string    `json:"genero"`
	Marca       string    `json:"marca"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
