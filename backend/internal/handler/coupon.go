package handler

import (
	"errors"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type CouponHandler struct {
	couponRepo  *repository.CouponRepository
	productRepo *repository.ProductRepository
}

func NewCouponHandler(couponRepo *repository.CouponRepository, productRepo *repository.ProductRepository) *CouponHandler {
	return &CouponHandler{
		couponRepo:  couponRepo,
		productRepo: productRepo,
	}
}

// Create handles POST /api/v1/products/:id/coupons
func (h *CouponHandler) Create(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "인증이 필요합니다",
		})
	}

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 상품 ID입니다",
		})
	}

	ctx := c.Request().Context()

	// Verify product ownership
	product, err := h.productRepo.GetByID(ctx, productID)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "상품을 찾을 수 없습니다",
			})
		}
		slog.Error("Coupon.Create: failed to get product", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "상품 정보를 불러오는데 실패했습니다",
		})
	}

	if product.SellerID != userID {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "본인의 상품에만 쿠폰을 생성할 수 있습니다",
		})
	}

	var req model.CreateCouponRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	// Validate discount type
	req.DiscountType = strings.ToLower(strings.TrimSpace(req.DiscountType))
	if req.DiscountType != "fixed" && req.DiscountType != "percentage" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "할인 타입은 fixed 또는 percentage여야 합니다",
		})
	}

	if req.DiscountValue < 1 {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "할인 값은 1 이상이어야 합니다",
		})
	}

	if req.DiscountType == "percentage" && req.DiscountValue > 100 {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "할인율은 1~100 사이여야 합니다",
		})
	}

	if req.ExpiresAt.IsZero() {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "만료일을 입력해주세요",
		})
	}

	coupon := &model.Coupon{
		ProductID:     productID,
		Code:          strings.TrimSpace(req.Code),
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		MaxUses:       req.MaxUses,
		OncePerUser:   req.OncePerUser,
		ExpiresAt:     req.ExpiresAt,
	}

	if err := h.couponRepo.Create(ctx, coupon); err != nil {
		if errors.Is(err, repository.ErrCouponCodeExists) {
			return c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "conflict",
				Message: "이미 동일한 코드의 쿠폰이 존재합니다",
			})
		}
		slog.Error("Coupon.Create: failed to create coupon", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "쿠폰 생성에 실패했습니다",
		})
	}

	coupon.ProductName = product.Name
	return c.JSON(http.StatusCreated, coupon)
}

// List handles GET /api/v1/products/:id/coupons
func (h *CouponHandler) List(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "인증이 필요합니다",
		})
	}

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 상품 ID입니다",
		})
	}

	ctx := c.Request().Context()

	// Verify product ownership
	product, err := h.productRepo.GetByID(ctx, productID)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "상품을 찾을 수 없습니다",
			})
		}
		slog.Error("Coupon.List: failed to get product", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "상품 정보를 불러오는데 실패했습니다",
		})
	}

	if product.SellerID != userID {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "본인의 상품 쿠폰만 조회할 수 있습니다",
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
	offset := (page - 1) * limit

	coupons, total, err := h.couponRepo.ListByProduct(ctx, productID, limit, offset)
	if err != nil {
		slog.Error("Coupon.List: failed to list coupons", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "쿠폰 목록을 불러오는데 실패했습니다",
		})
	}

	if coupons == nil {
		coupons = []*model.Coupon{}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return c.JSON(http.StatusOK, map[string]interface{}{
		"coupons":     coupons,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// ListMy handles GET /api/v1/me/coupons
func (h *CouponHandler) ListMy(c echo.Context) error {
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
	offset := (page - 1) * limit

	ctx := c.Request().Context()
	coupons, total, err := h.couponRepo.ListBySeller(ctx, userID, limit, offset)
	if err != nil {
		slog.Error("Coupon.ListMy: failed to list coupons", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "쿠폰 목록을 불러오는데 실패했습니다",
		})
	}

	if coupons == nil {
		coupons = []*model.Coupon{}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return c.JSON(http.StatusOK, map[string]interface{}{
		"coupons":     coupons,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// Delete handles DELETE /api/v1/coupons/:id
func (h *CouponHandler) Delete(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "인증이 필요합니다",
		})
	}

	couponID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 쿠폰 ID입니다",
		})
	}

	ctx := c.Request().Context()

	// Get coupon to verify ownership
	coupon, err := h.couponRepo.GetByID(ctx, couponID)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "쿠폰을 찾을 수 없습니다",
			})
		}
		slog.Error("Coupon.Delete: failed to get coupon", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "쿠폰 정보를 불러오는데 실패했습니다",
		})
	}

	// Verify product ownership
	product, err := h.productRepo.GetByID(ctx, coupon.ProductID)
	if err != nil {
		slog.Error("Coupon.Delete: failed to get product", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "상품 정보를 불러오는데 실패했습니다",
		})
	}

	if product.SellerID != userID {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "본인의 쿠폰만 삭제할 수 있습니다",
		})
	}

	if err := h.couponRepo.Delete(ctx, couponID); err != nil {
		slog.Error("Coupon.Delete: failed to delete coupon", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "쿠폰 삭제에 실패했습니다",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// Validate handles POST /api/v1/coupons/validate
func (h *CouponHandler) Validate(c echo.Context) error {
	var req model.ValidateCouponRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	if req.Code == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "쿠폰 코드를 입력해주세요",
		})
	}

	ctx := c.Request().Context()

	coupon, err := h.couponRepo.GetByCodeAndProduct(ctx, req.Code, req.ProductID)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return c.JSON(http.StatusOK, model.ValidateCouponResponse{
				Valid:   false,
				Message: "유효하지 않은 쿠폰 코드입니다",
			})
		}
		slog.Error("Coupon.Validate: failed to get coupon", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "쿠폰 검증에 실패했습니다",
		})
	}

	if err := repository.ValidateCoupon(coupon); err != nil {
		msg := "유효하지 않은 쿠폰입니다"
		if errors.Is(err, repository.ErrCouponExpired) {
			msg = "만료된 쿠폰입니다"
		} else if errors.Is(err, repository.ErrCouponMaxUsed) {
			msg = "사용 횟수가 초과된 쿠폰입니다"
		}
		return c.JSON(http.StatusOK, model.ValidateCouponResponse{
			Valid:   false,
			Message: msg,
		})
	}

	// Check once_per_user restriction
	if coupon.OncePerUser {
		if userID, ok := c.Get("user_id").(uuid.UUID); ok {
			used, err := h.couponRepo.HasUserUsedCoupon(ctx, coupon.ID, userID)
			if err != nil {
				slog.Error("Coupon.Validate: failed to check user usage", "error", err)
				return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
					Error:   "server_error",
					Message: "쿠폰 검증에 실패했습니다",
				})
			}
			if used {
				return c.JSON(http.StatusOK, model.ValidateCouponResponse{
					Valid:   false,
					Message: "이미 사용한 쿠폰입니다",
				})
			}
		}
	}

	// Get product price for discount calculation
	product, err := h.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		slog.Error("Coupon.Validate: failed to get product", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "상품 정보를 불러오는데 실패했습니다",
		})
	}

	discount := repository.CalculateDiscount(coupon, product.Price)

	return c.JSON(http.StatusOK, model.ValidateCouponResponse{
		Valid:          true,
		DiscountAmount: discount,
		Message:        "쿠폰이 적용되었습니다",
	})
}
