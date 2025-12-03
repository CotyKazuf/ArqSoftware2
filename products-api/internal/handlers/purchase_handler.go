package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"products-api/internal/middleware"
	"products-api/internal/repositories"
	"products-api/internal/responses"
	"products-api/internal/services"
)

// PurchaseHandler exposes checkout endpoints.
type PurchaseHandler struct {
	service *services.PurchaseService
}

// NewPurchaseHandler builds a PurchaseHandler.
func NewPurchaseHandler(service *services.PurchaseService) *PurchaseHandler {
	return &PurchaseHandler{service: service}
}

type checkoutRequest struct {
	Items []checkoutItem `json:"items"`
}

type checkoutItem struct {
	ProductID string `json:"producto_id"`
	Cantidad  int    `json:"cantidad"`
}

// CreatePurchase handles POST /compras.
func (h *PurchaseHandler) CreatePurchase(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		responses.WriteError(w, http.StatusUnauthorized, "AUTHENTICATION_FAILED", "User session required")
		return
	}

	var req checkoutRequest
	if err := decodeJSON(r, &req); err != nil {
		responses.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload")
		return
	}

	purchase, err := h.service.Checkout(userID, toCheckoutInput(req.Items))
	if err != nil {
		handlePurchaseError(w, err)
		return
	}

	responses.WriteJSON(w, http.StatusCreated, purchase)
}

// ListMyPurchases handles GET /compras/mias.
func (h *PurchaseHandler) ListMyPurchases(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		responses.WriteError(w, http.StatusUnauthorized, "AUTHENTICATION_FAILED", "User session required")
		return
	}

	purchases, err := h.service.ListByUser(userID)
	if err != nil {
		log.Printf("list purchases failed: %v", err)
		responses.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not obtain purchases")
		return
	}

	responses.WriteJSON(w, http.StatusOK, purchases)
}

func toCheckoutInput(items []checkoutItem) []services.CheckoutItemInput {
	result := make([]services.CheckoutItemInput, 0, len(items))
	for _, item := range items {
		result = append(result, services.CheckoutItemInput{
			ProductID: item.ProductID,
			Cantidad:  item.Cantidad,
		})
	}
	return result
}

func handlePurchaseError(w http.ResponseWriter, err error) {
	var valErr services.ValidationError
	if errors.As(err, &valErr) {
		code := valErr.Code
		if code == "" {
			code = "VALIDATION_ERROR"
		}
		responses.WriteError(w, http.StatusBadRequest, code, valErr.Error())
		return
	}

	var stockErr services.InsufficientStockError
	if errors.As(err, &stockErr) {
		message := "No hay stock suficiente para completar la compra."
		if stockErr.ProductID != "" {
			if stockErr.Available > 0 {
				message = fmt.Sprintf("Solo quedan %d unidades del producto solicitado (ID %s).", stockErr.Available, stockErr.ProductID)
			} else {
				message = fmt.Sprintf("El producto solicitado (ID %s) no tiene stock disponible.", stockErr.ProductID)
			}
		}
		responses.WriteError(w, http.StatusBadRequest, "INSUFFICIENT_STOCK", message)
		return
	}

	if errors.Is(err, repositories.ErrNotFound) {
		responses.WriteError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Uno de los productos solicitados no existe.")
		return
	}

	log.Printf("purchase error: %v", err)
	responses.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not create purchase")
}
