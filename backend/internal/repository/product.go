package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/f1-rivals-cup/backend/internal/database"
	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/google/uuid"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

// ProductRepository handles product database operations
type ProductRepository struct {
	db *database.DB
}

// NewProductRepository creates a new ProductRepository
func NewProductRepository(db *database.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create creates a new product with its options
func (r *ProductRepository) Create(ctx context.Context, product *model.Product) error {
	tx, err := r.db.Pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO products (seller_id, name, description, price, image_url, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRowContext(ctx, query,
		product.SellerID,
		product.Name,
		product.Description,
		product.Price,
		product.ImageURL,
		product.Status,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return err
	}

	// Insert options
	for i := range product.Options {
		opt := &product.Options[i]
		optQuery := `
			INSERT INTO product_options (product_id, option_name, option_value, additional_price)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at
		`
		err = tx.QueryRowContext(ctx, optQuery,
			product.ID,
			opt.OptionName,
			opt.OptionValue,
			opt.AdditionalPrice,
		).Scan(&opt.ID, &opt.CreatedAt)
		if err != nil {
			return err
		}
		opt.ProductID = product.ID
	}

	return tx.Commit()
}

// GetByID retrieves a product by ID with seller info and options
func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	query := `
		SELECT p.id, p.seller_id, p.name, p.description, p.price, p.image_url, p.status, p.created_at, p.updated_at, u.nickname
		FROM products p
		JOIN users u ON p.seller_id = u.id
		WHERE p.id = $1
	`

	product := &model.Product{}
	var imageURL sql.NullString
	err := r.db.Pool.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.SellerID,
		&product.Name,
		&product.Description,
		&product.Price,
		&imageURL,
		&product.Status,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.SellerNickname,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	if imageURL.Valid {
		product.ImageURL = imageURL.String
	}

	// Fetch options
	options, err := r.getOptionsByProductID(ctx, id)
	if err != nil {
		return nil, err
	}
	product.Options = options

	return product, nil
}

// getOptionsByProductID retrieves all options for a product
func (r *ProductRepository) getOptionsByProductID(ctx context.Context, productID uuid.UUID) ([]model.ProductOption, error) {
	query := `
		SELECT id, product_id, option_name, option_value, additional_price, created_at
		FROM product_options
		WHERE product_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []model.ProductOption
	for rows.Next() {
		var opt model.ProductOption
		if err := rows.Scan(
			&opt.ID,
			&opt.ProductID,
			&opt.OptionName,
			&opt.OptionValue,
			&opt.AdditionalPrice,
			&opt.CreatedAt,
		); err != nil {
			return nil, err
		}
		options = append(options, opt)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return options, nil
}

// List retrieves a paginated list of products
func (r *ProductRepository) List(ctx context.Context, page, limit int, status string) ([]*model.Product, int, error) {
	offset := (page - 1) * limit

	// Count total
	countQuery := `SELECT COUNT(*) FROM products WHERE status = $1`
	var total int
	if err := r.db.Pool.QueryRowContext(ctx, countQuery, status).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get products
	query := `
		SELECT p.id, p.seller_id, p.name, p.description, p.price, p.image_url, p.status, p.created_at, p.updated_at, u.nickname
		FROM products p
		JOIN users u ON p.seller_id = u.id
		WHERE p.status = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		product := &model.Product{}
		var imageURL sql.NullString
		if err := rows.Scan(
			&product.ID,
			&product.SellerID,
			&product.Name,
			&product.Description,
			&product.Price,
			&imageURL,
			&product.Status,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.SellerNickname,
		); err != nil {
			return nil, 0, err
		}
		if imageURL.Valid {
			product.ImageURL = imageURL.String
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// ListBySeller retrieves a paginated list of products for a seller
func (r *ProductRepository) ListBySeller(ctx context.Context, sellerID uuid.UUID, page, limit int) ([]*model.Product, int, error) {
	offset := (page - 1) * limit

	// Count total
	countQuery := `SELECT COUNT(*) FROM products WHERE seller_id = $1`
	var total int
	if err := r.db.Pool.QueryRowContext(ctx, countQuery, sellerID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get products
	query := `
		SELECT p.id, p.seller_id, p.name, p.description, p.price, p.image_url, p.status, p.created_at, p.updated_at, u.nickname
		FROM products p
		JOIN users u ON p.seller_id = u.id
		WHERE p.seller_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, sellerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		product := &model.Product{}
		var imageURL sql.NullString
		if err := rows.Scan(
			&product.ID,
			&product.SellerID,
			&product.Name,
			&product.Description,
			&product.Price,
			&imageURL,
			&product.Status,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.SellerNickname,
		); err != nil {
			return nil, 0, err
		}
		if imageURL.Valid {
			product.ImageURL = imageURL.String
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// Update updates a product
func (r *ProductRepository) Update(ctx context.Context, product *model.Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, image_url = $4, status = $5, updated_at = NOW()
		WHERE id = $6
	`

	result, err := r.db.Pool.ExecContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.ImageURL,
		product.Status,
		product.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrProductNotFound
	}

	return nil
}

// Delete deletes a product (CASCADE deletes options)
func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.Pool.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrProductNotFound
	}

	return nil
}

// ReplaceOptions deletes existing options and inserts new ones
func (r *ProductRepository) ReplaceOptions(ctx context.Context, productID uuid.UUID, options []model.ProductOption) ([]model.ProductOption, error) {
	tx, err := r.db.Pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Delete existing options
	_, err = tx.ExecContext(ctx, `DELETE FROM product_options WHERE product_id = $1`, productID)
	if err != nil {
		return nil, err
	}

	// Insert new options
	var result []model.ProductOption
	for _, opt := range options {
		optQuery := `
			INSERT INTO product_options (product_id, option_name, option_value, additional_price)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at
		`
		var newOpt model.ProductOption
		newOpt.ProductID = productID
		newOpt.OptionName = opt.OptionName
		newOpt.OptionValue = opt.OptionValue
		newOpt.AdditionalPrice = opt.AdditionalPrice

		err = tx.QueryRowContext(ctx, optQuery,
			productID,
			opt.OptionName,
			opt.OptionValue,
			opt.AdditionalPrice,
		).Scan(&newOpt.ID, &newOpt.CreatedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, newOpt)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}
