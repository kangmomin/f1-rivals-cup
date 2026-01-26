package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ParticipantHandler struct {
	participantRepo *repository.ParticipantRepository
	leagueRepo      *repository.LeagueRepository
	accountRepo     *repository.AccountRepository
}

func NewParticipantHandler(participantRepo *repository.ParticipantRepository, leagueRepo *repository.LeagueRepository, accountRepo *repository.AccountRepository) *ParticipantHandler {
	return &ParticipantHandler{
		participantRepo: participantRepo,
		leagueRepo:      leagueRepo,
		accountRepo:     accountRepo,
	}
}

// Join handles POST /api/v1/leagues/:id/join
func (h *ParticipantHandler) Join(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "로그인이 필요합니다",
		})
	}

	var req model.JoinLeagueRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	// Validate roles
	if len(req.Roles) == 0 {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "최소 하나의 역할을 선택해주세요",
		})
	}

	validRoles := map[string]bool{
		string(model.RoleDirector): true,
		string(model.RolePlayer):   true,
		string(model.RoleReserve):  true,
		string(model.RoleEngineer): true,
	}
	for _, role := range req.Roles {
		if !validRoles[role] {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_role",
				Message: "유효하지 않은 역할입니다: " + role,
			})
		}
	}

	ctx := c.Request().Context()

	// Check if league exists and is open
	league, err := h.leagueRepo.GetByID(ctx, leagueID)
	if err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		slog.Error("Participant.Join: failed to get league", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그 정보를 불러오는데 실패했습니다",
		})
	}

	if league.Status != model.LeagueStatusOpen {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "league_not_open",
			Message: "현재 참가 신청을 받지 않는 리그입니다",
		})
	}

	// Check player limit per team (2 players max)
	hasPlayerRole := false
	for _, role := range req.Roles {
		if role == string(model.RolePlayer) {
			hasPlayerRole = true
			break
		}
	}

	if hasPlayerRole && req.TeamName != nil && *req.TeamName != "" {
		playerCount, err := h.participantRepo.CountPlayersByTeam(ctx, leagueID, *req.TeamName)
		if err != nil {
			slog.Error("Participant.Join: failed to count players by team", "error", err, "league_id", leagueID, "team_name", *req.TeamName)
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "server_error",
				Message: "팀 정보를 확인하는데 실패했습니다",
			})
		}
		if playerCount >= 2 {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "team_full",
				Message: "해당 팀의 선수 정원(2명)이 이미 찼습니다",
			})
		}
	}

	participant := &model.LeagueParticipant{
		LeagueID: leagueID,
		UserID:   userID,
		Status:   model.ParticipantStatusPending,
		Roles:    req.Roles,
		TeamName: req.TeamName,
		Message:  req.Message,
	}

	if err := h.participantRepo.Create(ctx, participant); err != nil {
		if errors.Is(err, repository.ErrAlreadyParticipating) {
			return c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "already_participating",
				Message: "이미 참가 신청한 리그입니다",
			})
		}
		slog.Error("Participant.Join: failed to create participant", "error", err, "league_id", leagueID, "user_id", userID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "참가 신청에 실패했습니다",
		})
	}

	return c.JSON(http.StatusCreated, participant)
}

// GetMyStatus handles GET /api/v1/leagues/:id/my-status
func (h *ParticipantHandler) GetMyStatus(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"is_participating": false,
			"participant":      nil,
		})
	}

	ctx := c.Request().Context()

	participant, err := h.participantRepo.GetByLeagueAndUser(ctx, leagueID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"is_participating": false,
				"participant":      nil,
			})
		}
		slog.Error("Participant.GetMyStatus: failed to get participant status", "error", err, "league_id", leagueID, "user_id", userID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "참가 상태를 확인하는데 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"is_participating": true,
		"participant":      participant,
	})
}

// Cancel handles DELETE /api/v1/leagues/:id/join
func (h *ParticipantHandler) Cancel(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "로그인이 필요합니다",
		})
	}

	ctx := c.Request().Context()

	participant, err := h.participantRepo.GetByLeagueAndUser(ctx, leagueID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "참가 신청 내역이 없습니다",
			})
		}
		slog.Error("Participant.Cancel: failed to get participant", "error", err, "league_id", leagueID, "user_id", userID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "참가 상태를 확인하는데 실패했습니다",
		})
	}

	// Only pending or rejected can be deleted by user
	if participant.Status == model.ParticipantStatusApproved {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "cannot_cancel",
			Message: "이미 승인된 참가는 취소할 수 없습니다",
		})
	}

	if err := h.participantRepo.Delete(ctx, participant.ID); err != nil {
		slog.Error("Participant.Cancel: failed to delete participant", "error", err, "participant_id", participant.ID, "league_id", leagueID, "user_id", userID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "참가 취소에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "참가 신청이 취소되었습니다",
	})
}

// ListByLeague handles GET /api/v1/admin/leagues/:id/participants
func (h *ParticipantHandler) ListByLeague(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	status := c.QueryParam("status")
	ctx := c.Request().Context()

	participants, err := h.participantRepo.ListByLeague(ctx, leagueID, status)
	if err != nil {
		slog.Error("Participant.ListByLeague: failed to list participants", "error", err, "league_id", leagueID, "status", status)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "참가자 목록을 불러오는데 실패했습니다",
		})
	}

	if participants == nil {
		participants = []*model.LeagueParticipant{}
	}

	return c.JSON(http.StatusOK, model.ListParticipantsResponse{
		Participants: participants,
		Total:        len(participants),
	})
}

// ListApprovedByLeague handles GET /api/v1/leagues/:id/participants (public endpoint)
// Returns only approved participants for public viewing
func (h *ParticipantHandler) ListApprovedByLeague(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	ctx := c.Request().Context()

	// Only return approved participants for public access
	participants, err := h.participantRepo.ListByLeague(ctx, leagueID, string(model.ParticipantStatusApproved))
	if err != nil {
		slog.Error("Participant.ListApprovedByLeague: failed to list participants", "error", err, "league_id", leagueID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "참가자 목록을 불러오는데 실패했습니다",
		})
	}

	if participants == nil {
		participants = []*model.LeagueParticipant{}
	}

	return c.JSON(http.StatusOK, model.ListParticipantsResponse{
		Participants: participants,
		Total:        len(participants),
	})
}

// ListMyParticipations handles GET /api/v1/me/participations
func (h *ParticipantHandler) ListMyParticipations(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "로그인이 필요합니다",
		})
	}

	ctx := c.Request().Context()

	participants, err := h.participantRepo.ListByUser(ctx, userID)
	if err != nil {
		slog.Error("Participant.ListMyParticipations: failed to list user participations", "error", err, "user_id", userID)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "참가 목록을 불러오는데 실패했습니다",
		})
	}

	if participants == nil {
		participants = []*model.LeagueParticipant{}
	}

	return c.JSON(http.StatusOK, model.ListParticipantsResponse{
		Participants: participants,
		Total:        len(participants),
	})
}

// UpdateStatus handles PUT /api/v1/admin/participants/:id/status
func (h *ParticipantHandler) UpdateStatus(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 참가자 ID입니다",
		})
	}

	var req model.UpdateParticipantRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	if req.Status != model.ParticipantStatusApproved && req.Status != model.ParticipantStatusRejected {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_status",
			Message: "유효하지 않은 상태입니다",
		})
	}

	ctx := c.Request().Context()

	// Get participant to check current status and get league ID
	participant, err := h.participantRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "참가자를 찾을 수 없습니다",
			})
		}
		slog.Error("Participant.UpdateStatus: failed to get participant", "error", err, "participant_id", id)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "참가자 조회에 실패했습니다",
		})
	}

	if err := h.participantRepo.UpdateStatus(ctx, id, req.Status); err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "참가자를 찾을 수 없습니다",
			})
		}
		slog.Error("Participant.UpdateStatus: failed to update participant status", "error", err, "participant_id", id, "status", req.Status)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "상태 변경에 실패했습니다",
		})
	}

	// Create participant account when approved (if not already exists)
	if req.Status == model.ParticipantStatusApproved && participant.Status != model.ParticipantStatusApproved {
		// Check if account already exists
		_, err := h.accountRepo.GetByOwner(ctx, participant.LeagueID, participant.ID, model.OwnerTypeParticipant)
		if errors.Is(err, repository.ErrAccountNotFound) {
			// Create new account for participant
			account := &model.Account{
				LeagueID:  participant.LeagueID,
				OwnerID:   participant.ID,
				OwnerType: model.OwnerTypeParticipant,
				Balance:   0,
			}
			if err := h.accountRepo.Create(ctx, account); err != nil {
				slog.Error("Participant.UpdateStatus: failed to create participant account", "error", err, "participant_id", id)
				// Don't fail the request, account creation is secondary
			} else {
				slog.Info("Participant.UpdateStatus: created participant account", "participant_id", id, "account_id", account.ID)
			}
		} else if err != nil {
			slog.Error("Participant.UpdateStatus: failed to check existing account", "error", err, "participant_id", id)
		}
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "상태가 변경되었습니다",
	})
}

// UpdateTeam handles PUT /api/v1/admin/participants/:id/team
func (h *ParticipantHandler) UpdateTeam(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 참가자 ID입니다",
		})
	}

	var req struct {
		TeamName *string `json:"team_name"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	ctx := c.Request().Context()

	if err := h.participantRepo.UpdateTeam(ctx, id, req.TeamName); err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "참가자를 찾을 수 없습니다",
			})
		}
		slog.Error("Participant.UpdateTeam: failed to update participant team", "error", err, "participant_id", id, "team_name", req.TeamName)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "팀 배정에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "팀이 배정되었습니다",
	})
}
