package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type SubscriptionHandler struct {
	subscriptionRepo *repository.SubscriptionRepository
	productRepo      *repository.ProductRepository
	accountRepo      *repository.AccountRepository
	participantRepo  *repository.ParticipantRepository
}

func NewSubscriptionHandler(
	subscriptionRepo *repository.SubscriptionRepository,
	productRepo *repository.ProductRepository,
	accountRepo *repository.AccountRepository,
	participantRepo *repository.ParticipantRepository,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionRepo: subscriptionRepo,
		productRepo:      productRepo,
		accountRepo:      accountRepo,
		participantRepo:  participantRepo,
	}
}

// Subscribe handles POST /api/v1/subscriptions
func (h *SubscriptionHandler) Subscribe(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "인증이 필요합니다",
		})
	}

	var req model.SubscribeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	return h.processSubscription(c, userID, req)
}

// Renew handles POST /api/v1/subscriptions/:id/renew
func (h *SubscriptionHandler) Renew(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "인증이 필요합니다",
		})
	}

	subID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 구독 ID입니다",
		})
	}

	ctx := c.Request().Context()

	// Get existing subscription
	sub, err := h.subscriptionRepo.GetByID(ctx, subID)
	if err != nil {
		if errors.Is(err, repository.ErrSubscriptionNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "구독을 찾을 수 없습니다",
			})
		}
		slog.Error("Subscription.Renew: failed to get subscription", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "구독 정보를 불러오는데 실패했습니다",
		})
	}

	if sub.UserID != userID {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "본인의 구독만 갱신할 수 있습니다",
		})
	}

	// Bind optional option_id
	var renewReq struct {
		OptionID *uuid.UUID `json:"option_id,omitempty"`
	}
	c.Bind(&renewReq)

	return h.processSubscription(c, userID, model.SubscribeRequest{
		ProductID: sub.ProductID,
		LeagueID:  sub.LeagueID,
		OptionID:  renewReq.OptionID,
	})
}

// ListMy handles GET /api/v1/me/subscriptions
func (h *SubscriptionHandler) ListMy(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "인증이 필요합니다",
		})
	}

	ctx := c.Request().Context()

	subs, err := h.subscriptionRepo.GetActiveByUser(ctx, userID)
	if err != nil {
		slog.Error("Subscription.ListMy: failed to list subscriptions", "error", err, "user_id", userID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "구독 목록을 불러오는데 실패했습니다",
		})
	}

	if subs == nil {
		subs = []*model.Subscription{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"subscriptions": subs,
		"total":         len(subs),
	})
}

// CheckAccess handles GET /api/v1/products/:id/access
// Returns whether the current user has access to the product via subscription permission.
func (h *SubscriptionHandler) CheckAccess(c echo.Context) error {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 상품 ID입니다",
		})
	}

	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"has_access":   false,
			"subscription": nil,
		})
	}

	// Check JWT permissions
	permKey := fmt.Sprintf("product.%s", productID.String())
	hasAccess := false
	if perms, ok := c.Get("permissions").([]string); ok {
		for _, p := range perms {
			if p == permKey {
				hasAccess = true
				break
			}
		}
	}

	// Also look up the active subscription for additional info
	ctx := c.Request().Context()
	subs, err := h.subscriptionRepo.GetActiveByUser(ctx, userID)
	if err != nil {
		slog.Error("Subscription.CheckAccess: failed to get subscriptions", "error", err)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"has_access":   hasAccess,
			"subscription": nil,
		})
	}

	var activeSub *model.Subscription
	for _, s := range subs {
		if s.ProductID == productID {
			activeSub = s
			break
		}
	}

	if activeSub != nil {
		hasAccess = true
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"has_access":   hasAccess,
		"subscription": activeSub,
	})
}

// processSubscription is the shared logic for Subscribe and Renew.
func (h *SubscriptionHandler) processSubscription(
	c echo.Context,
	userID uuid.UUID,
	req model.SubscribeRequest,
) error {
	reqCtx := c.Request().Context()

	// 1. Get product
	product, err := h.productRepo.GetByID(reqCtx, req.ProductID)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "상품을 찾을 수 없습니다",
			})
		}
		slog.Error("Subscription: failed to get product", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "상품 정보를 불러오는데 실패했습니다",
		})
	}

	if product.Status != "active" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "비활성 상품은 구독할 수 없습니다",
		})
	}

	if product.SubscriptionDurationDays == nil || *product.SubscriptionDurationDays <= 0 {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "구독 상품이 아닙니다",
		})
	}

	// 2. Calculate price
	totalPrice := product.Price
	if req.OptionID != nil {
		found := false
		for _, opt := range product.Options {
			if opt.ID == *req.OptionID {
				totalPrice += opt.AdditionalPrice
				found = true
				break
			}
		}
		if !found {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "잘못된 옵션입니다",
			})
		}
	}

	// 3. Get buyer participant (must be approved)
	participant, err := h.participantRepo.GetByLeagueAndUser(reqCtx, req.LeagueID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error:   "forbidden",
				Message: "해당 리그에 참여하고 있지 않습니다",
			})
		}
		slog.Error("Subscription: failed to get participant", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "참가자 정보를 불러오는데 실패했습니다",
		})
	}

	if participant.Status != model.ParticipantStatusApproved {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "승인된 참가자만 구독할 수 있습니다",
		})
	}

	// 4. Get buyer account
	buyerAccount, err := h.accountRepo.GetByOwner(reqCtx, req.LeagueID, participant.ID, model.OwnerTypeParticipant)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "구매자 계좌를 찾을 수 없습니다",
			})
		}
		slog.Error("Subscription: failed to get buyer account", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "계좌 정보를 불러오는데 실패했습니다",
		})
	}

	if buyerAccount.Balance < totalPrice {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "insufficient_balance",
			Message: "잔액이 부족합니다",
		})
	}

	// 5. Get seller participant & account in the same league
	sellerParticipant, err := h.participantRepo.GetByLeagueAndUser(reqCtx, req.LeagueID, product.SellerID)
	if err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "판매자가 해당 리그에 참여하고 있지 않습니다",
			})
		}
		slog.Error("Subscription: failed to get seller participant", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "판매자 정보를 불러오는데 실패했습니다",
		})
	}

	sellerAccount, err := h.accountRepo.GetByOwner(reqCtx, req.LeagueID, sellerParticipant.ID, model.OwnerTypeParticipant)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "판매자 계좌를 찾을 수 없습니다",
			})
		}
		slog.Error("Subscription: failed to get seller account", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "판매자 계좌 정보를 불러오는데 실패했습니다",
		})
	}

	// 6. Execute subscription
	desc := fmt.Sprintf("구독: %s", product.Name)
	sub, err := h.subscriptionRepo.Subscribe(
		reqCtx,
		userID, product.ID, req.LeagueID,
		buyerAccount.ID, sellerAccount.ID,
		totalPrice,
		*product.SubscriptionDurationDays,
		desc,
	)
	if err != nil {
		if errors.Is(err, repository.ErrInsufficientBalance) {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "insufficient_balance",
				Message: "잔액이 부족합니다",
			})
		}
		slog.Error("Subscription: failed to subscribe", "error", err, "user_id", userID, "product_id", product.ID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "구독 처리에 실패했습니다",
		})
	}

	sub.ProductName = product.Name

	return c.JSON(http.StatusCreated, sub)
}
