package model

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a product in the shop
type Product struct {
	ID             uuid.UUID       `json:"id"`
	SellerID       uuid.UUID       `json:"seller_id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Price          int64           `json:"price"`
	ImageURL       string          `json:"image_url,omitempty"`
	Status         string          `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	SellerNickname string          `json:"seller_nickname,omitempty"`
	Options        []ProductOption `json:"options,omitempty"`
}

// ProductOption represents an option for a product
type ProductOption struct {
	ID              uuid.UUID `json:"id"`
	ProductID       uuid.UUID `json:"product_id"`
	OptionName      string    `json:"option_name"`
	OptionValue     string    `json:"option_value"`
	AdditionalPrice int64     `json:"additional_price"`
	CreatedAt       time.Time `json:"created_at"`
}

// CreateProductRequest represents a request to create a product
type CreateProductRequest struct {
	Name        string                       `json:"name" validate:"required,min=2,max=200"`
	Description string                       `json:"description"`
	Price       int64                        `json:"price" validate:"min=0"`
	ImageURL    string                       `json:"image_url,omitempty"`
	Options     []CreateProductOptionRequest `json:"options,omitempty"`
}

// CreateProductOptionRequest represents a request to create a product option
type CreateProductOptionRequest struct {
	OptionName      string `json:"option_name" validate:"required,min=1,max=100"`
	OptionValue     string `json:"option_value" validate:"required,min=1,max=200"`
	AdditionalPrice int64  `json:"additional_price"`
}

// UpdateProductRequest represents a request to update a product
type UpdateProductRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=2,max=200"`
	Description *string `json:"description,omitempty"`
	Price       *int64  `json:"price,omitempty" validate:"omitempty,min=0"`
	ImageURL    *string `json:"image_url,omitempty"`
	Status      *string `json:"status,omitempty"`
}

// ListProductsResponse represents the response for listing products
type ListProductsResponse struct {
	Products   []*Product `json:"products"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	TotalPages int        `json:"total_pages"`
}
