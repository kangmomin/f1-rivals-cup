package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/f1-rivals-cup/backend/internal/auth"
	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ProductHandler handles product requests
type ProductHandler struct {
	productRepo *repository.ProductRepository
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(productRepo *repository.ProductRepository) *ProductHandler {
	return &ProductHandler{
		productRepo: productRepo,
	}
}

// List handles GET /api/v1/products (public, active only)
func (h *ProductHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	ctx := c.Request().Context()

	products, total, err := h.productRepo.List(ctx, page, limit, "active")
	if err != nil {
		slog.Error("Product.List: failed to list products", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "상품 목록을 불러오는데 실패했습니다",
		})
	}

	if products == nil {
		products = []*model.Product{}
	}

	totalPages := (total + limit - 1) / limit

	return c.JSON(http.StatusOK, model.ListProductsResponse{
		Products:   products,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// Get handles GET /api/v1/products/:id (public)
func (h *ProductHandler) Get(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 상품 ID입니다",
		})
	}

	ctx := c.Request().Context()

	product, err := h.productRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "상품을 찾을 수 없습니다",
			})
		}
		slog.Error("Product.Get: failed to get product", "error", err, "product_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "상품을 불러오는데 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, product)
}

// Create handles POST /api/v1/products (requires store.create)
func (h *ProductHandler) Create(c echo.Context) error {
	var req model.CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	if err := validateCreateProductRequest(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
	}

	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "인증이 필요합니다",
		})
	}

	ctx := c.Request().Context()

	product := &model.Product{
		SellerID:    userID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		Status:      "active",
	}

	// Convert option requests to options
	for _, optReq := range req.Options {
		product.Options = append(product.Options, model.ProductOption{
			OptionName:      optReq.OptionName,
			OptionValue:     optReq.OptionValue,
			AdditionalPrice: optReq.AdditionalPrice,
		})
	}

	if err := h.productRepo.Create(ctx, product); err != nil {
		slog.Error("Product.Create: failed to create product", "error", err, "user_id", userID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "상품 등록에 실패했습니다",
		})
	}

	// Set seller nickname from context
	nickname, _ := c.Get("nickname").(string)
	product.SellerNickname = nickname

	return c.JSON(http.StatusCreated, product)
}

// Update handles PUT /api/v1/products/:id (requires store.edit + owner or store.manage)
func (h *ProductHandler) Update(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 상품 ID입니다",
		})
	}

	var req model.UpdateProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	ctx := c.Request().Context()

	// Get existing product
	product, err := h.productRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "상품을 찾을 수 없습니다",
			})
		}
		slog.Error("Product.Update: failed to get product", "error", err, "product_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "상품을 불러오는데 실패했습니다",
		})
	}

	// Check ownership or manage permission
	if !h.canManageProduct(c, product) {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "이 상품을 수정할 권한이 없습니다",
		})
	}

	// Update fields
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "validation_error",
				Message: "상품명을 입력해주세요",
			})
		}
		if len(name) < 2 || len(name) > 200 {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "validation_error",
				Message: "상품명은 2자 이상 200자 이하여야 합니다",
			})
		}
		product.Name = name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		if *req.Price < 0 {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "validation_error",
				Message: "가격은 0 이상이어야 합니다",
			})
		}
		product.Price = *req.Price
	}
	if req.ImageURL != nil {
		product.ImageURL = *req.ImageURL
	}
	if req.Status != nil {
		status := *req.Status
		if status != "active" && status != "inactive" {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "validation_error",
				Message: "상태는 active 또는 inactive여야 합니다",
			})
		}
		product.Status = status
	}

	if err := h.productRepo.Update(ctx, product); err != nil {
		slog.Error("Product.Update: failed to update product", "error", err, "product_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "상품 수정에 실패했습니다",
		})
	}

	// Reload to get updated_at
	product, _ = h.productRepo.GetByID(ctx, id)

	return c.JSON(http.StatusOK, product)
}

// Delete handles DELETE /api/v1/products/:id (requires store.delete + owner or store.manage)
func (h *ProductHandler) Delete(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 상품 ID입니다",
		})
	}

	ctx := c.Request().Context()

	// Get product for ownership check
	product, err := h.productRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "상품을 찾을 수 없습니다",
			})
		}
		slog.Error("Product.Delete: failed to get product", "error", err, "product_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "상품을 불러오는데 실패했습니다",
		})
	}

	// Check ownership or manage permission
	if !h.canManageProduct(c, product) {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "이 상품을 삭제할 권한이 없습니다",
		})
	}

	if err := h.productRepo.Delete(ctx, id); err != nil {
		slog.Error("Product.Delete: failed to delete product", "error", err, "product_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "상품 삭제에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "상품이 삭제되었습니다",
	})
}

// ListMy handles GET /api/v1/me/products (requires store.create)
func (h *ProductHandler) ListMy(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "인증이 필요합니다",
		})
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	ctx := c.Request().Context()

	products, total, err := h.productRepo.ListBySeller(ctx, userID, page, limit)
	if err != nil {
		slog.Error("Product.ListMy: failed to list products", "error", err, "user_id", userID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "상품 목록을 불러오는데 실패했습니다",
		})
	}

	if products == nil {
		products = []*model.Product{}
	}

	totalPages := (total + limit - 1) / limit

	return c.JSON(http.StatusOK, model.ListProductsResponse{
		Products:   products,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// ManageOptions handles PUT /api/v1/products/:id/options (requires store.edit + owner or store.manage)
func (h *ProductHandler) ManageOptions(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 상품 ID입니다",
		})
	}

	var req struct {
		Options []model.CreateProductOptionRequest `json:"options"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	ctx := c.Request().Context()

	// Get product for ownership check
	product, err := h.productRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "상품을 찾을 수 없습니다",
			})
		}
		slog.Error("Product.ManageOptions: failed to get product", "error", err, "product_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "상품을 불러오는데 실패했습니다",
		})
	}

	// Check ownership or manage permission
	if !h.canManageProduct(c, product) {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "이 상품의 옵션을 수정할 권한이 없습니다",
		})
	}

	// Validate options
	for _, opt := range req.Options {
		if strings.TrimSpace(opt.OptionName) == "" {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "validation_error",
				Message: "옵션 이름을 입력해주세요",
			})
		}
		if strings.TrimSpace(opt.OptionValue) == "" {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "validation_error",
				Message: "옵션 값을 입력해주세요",
			})
		}
	}

	// Convert to model options
	var options []model.ProductOption
	for _, optReq := range req.Options {
		options = append(options, model.ProductOption{
			OptionName:      strings.TrimSpace(optReq.OptionName),
			OptionValue:     strings.TrimSpace(optReq.OptionValue),
			AdditionalPrice: optReq.AdditionalPrice,
		})
	}

	newOptions, err := h.productRepo.ReplaceOptions(ctx, id, options)
	if err != nil {
		slog.Error("Product.ManageOptions: failed to replace options", "error", err, "product_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "옵션 수정에 실패했습니다",
		})
	}

	if newOptions == nil {
		newOptions = []model.ProductOption{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"options": newOptions,
	})
}

// canManageProduct checks if the current user can manage (edit/delete) a product
func (h *ProductHandler) canManageProduct(c echo.Context, product *model.Product) bool {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return false
	}

	// Owner can always manage
	if product.SellerID == userID {
		return true
	}

	// ADMIN role can manage all
	userRole := c.Get("role")
	if userRole != nil && auth.Role(userRole.(string)) == auth.RoleAdmin {
		return true
	}

	// Check store.manage permission
	userPerms := c.Get("permissions")
	if userPerms != nil {
		perms, ok := userPerms.([]string)
		if ok && auth.HasPermission(perms, auth.PermStoreManage) {
			return true
		}
	}

	return false
}

func validateCreateProductRequest(req *model.CreateProductRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)

	if req.Name == "" {
		return errors.New("상품명을 입력해주세요")
	}
	if len(req.Name) < 2 {
		return errors.New("상품명은 최소 2자 이상이어야 합니다")
	}
	if len(req.Name) > 200 {
		return errors.New("상품명은 최대 200자까지 가능합니다")
	}
	if req.Price < 0 {
		return errors.New("가격은 0 이상이어야 합니다")
	}

	// Validate options
	for i, opt := range req.Options {
		req.Options[i].OptionName = strings.TrimSpace(opt.OptionName)
		req.Options[i].OptionValue = strings.TrimSpace(opt.OptionValue)

		if req.Options[i].OptionName == "" {
			return errors.New("옵션 이름을 입력해주세요")
		}
		if req.Options[i].OptionValue == "" {
			return errors.New("옵션 값을 입력해주세요")
		}
	}

	return nil
}
