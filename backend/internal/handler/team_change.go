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

type TeamChangeHandler struct {
	teamChangeRepo  *repository.TeamChangeRepository
	participantRepo *repository.ParticipantRepository
	teamRepo        *repository.TeamRepository
	leagueRepo      *repository.LeagueRepository
}

func NewTeamChangeHandler(
	teamChangeRepo *repository.TeamChangeRepository,
	participantRepo *repository.ParticipantRepository,
	teamRepo *repository.TeamRepository,
	leagueRepo *repository.LeagueRepository,
) *TeamChangeHandler {
	return &TeamChangeHandler{
		teamChangeRepo:  teamChangeRepo,
		participantRepo: participantRepo,
		teamRepo:        teamRepo,
		leagueRepo:      leagueRepo,
	}
}

// CreateRequest handles POST /api/v1/leagues/:id/team-change-requests
func (h *TeamChangeHandler) CreateRequest(c echo.Context) error {
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

	var req model.CreateTeamChangeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	if req.RequestedTeamName == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "이적할 팀 이름을 입력해주세요",
		})
	}

	ctx := c.Request().Context()

	// Check if league exists
	_, err = h.leagueRepo.GetByID(ctx, leagueID)
	if err != nil {
		if errors.Is(err, repository.ErrLeagueNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "리그를 찾을 수 없습니다",
			})
		}
		slog.Error("TeamChange.CreateRequest: failed to get league", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "리그 정보를 불러오는데 실패했습니다",
		})
	}

	// Get participant
	participant, err := h.participantRepo.GetByLeagueAndUser(ctx, leagueID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error:   "forbidden",
				Message: "해당 리그의 참가자가 아닙니다",
			})
		}
		slog.Error("TeamChange.CreateRequest: failed to get participant", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "참가자 정보를 불러오는데 실패했습니다",
		})
	}

	// Check if participant is approved
	if participant.Status != model.ParticipantStatusApproved {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "승인된 참가자만 팀 변경 신청이 가능합니다",
		})
	}

	// Check if requested team exists in the league
	teams, err := h.teamRepo.ListByLeague(ctx, leagueID)
	if err != nil {
		slog.Error("TeamChange.CreateRequest: failed to list teams", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "팀 목록을 불러오는데 실패했습니다",
		})
	}

	teamExists := false
	for _, t := range teams {
		if t.Name == req.RequestedTeamName {
			teamExists = true
			break
		}
	}

	if !teamExists {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "team_not_found",
			Message: "해당 팀이 리그에 존재하지 않습니다",
		})
	}

	// Check if trying to transfer to current team
	if participant.TeamName != nil && *participant.TeamName == req.RequestedTeamName {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "same_team",
			Message: "현재 소속된 팀과 동일한 팀으로는 이적 신청할 수 없습니다",
		})
	}

	// Check player count in target team (max 2)
	// Check if participant has player role
	hasPlayerRole := false
	for _, role := range participant.Roles {
		if role == string(model.RolePlayer) {
			hasPlayerRole = true
			break
		}
	}

	if hasPlayerRole {
		playerCount, err := h.participantRepo.CountPlayersByTeam(ctx, leagueID, req.RequestedTeamName)
		if err != nil {
			slog.Error("TeamChange.CreateRequest: failed to count players", "error", err)
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

	// Create the team change request
	teamChangeReq := &model.TeamChangeRequest{
		ParticipantID:     participant.ID,
		CurrentTeamName:   participant.TeamName,
		RequestedTeamName: req.RequestedTeamName,
		Status:            model.TeamChangeStatusPending,
		Reason:            req.Reason,
	}

	if err := h.teamChangeRepo.CreateRequest(ctx, teamChangeReq); err != nil {
		if errors.Is(err, repository.ErrPendingRequestExists) {
			return c.JSON(http.StatusConflict, model.ErrorResponse{
				Error:   "pending_request_exists",
				Message: "이미 대기 중인 팀 변경 신청이 있습니다",
			})
		}
		slog.Error("TeamChange.CreateRequest: failed to create request", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "팀 변경 신청에 실패했습니다",
		})
	}

	return c.JSON(http.StatusCreated, teamChangeReq)
}

// ListByLeague handles GET /api/v1/leagues/:id/team-change-requests (for directors/admin)
func (h *TeamChangeHandler) ListByLeague(c echo.Context) error {
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

	requests, err := h.teamChangeRepo.ListRequestsByLeague(ctx, leagueID, status)
	if err != nil {
		slog.Error("TeamChange.ListByLeague: failed to list requests", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "팀 변경 신청 목록을 불러오는데 실패했습니다",
		})
	}

	if requests == nil {
		requests = []*model.TeamChangeRequest{}
	}

	return c.JSON(http.StatusOK, model.TeamChangeRequestListResponse{
		Requests: requests,
		Total:    len(requests),
	})
}

// ListMyRequests handles GET /api/v1/leagues/:id/my-team-change-requests
func (h *TeamChangeHandler) ListMyRequests(c echo.Context) error {
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

	// Get participant
	participant, err := h.participantRepo.GetByLeagueAndUser(ctx, leagueID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return c.JSON(http.StatusOK, model.TeamChangeRequestListResponse{
				Requests: []*model.TeamChangeRequest{},
				Total:    0,
			})
		}
		slog.Error("TeamChange.ListMyRequests: failed to get participant", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "참가자 정보를 불러오는데 실패했습니다",
		})
	}

	requests, err := h.teamChangeRepo.ListRequestsByParticipant(ctx, participant.ID)
	if err != nil {
		slog.Error("TeamChange.ListMyRequests: failed to list requests", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "팀 변경 신청 목록을 불러오는데 실패했습니다",
		})
	}

	if requests == nil {
		requests = []*model.TeamChangeRequest{}
	}

	return c.JSON(http.StatusOK, model.TeamChangeRequestListResponse{
		Requests: requests,
		Total:    len(requests),
	})
}

// ReviewRequest handles PUT /api/v1/leagues/:id/team-change-requests/:requestId
func (h *TeamChangeHandler) ReviewRequest(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	requestIDStr := c.Param("requestId")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 신청 ID입니다",
		})
	}

	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "로그인이 필요합니다",
		})
	}

	var req model.ReviewTeamChangeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 요청입니다",
		})
	}

	if req.Status != model.TeamChangeStatusApproved && req.Status != model.TeamChangeStatusRejected {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_status",
			Message: "유효하지 않은 상태입니다 (approved 또는 rejected)",
		})
	}

	ctx := c.Request().Context()

	// Get the team change request
	changeRequest, err := h.teamChangeRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, repository.ErrTeamChangeRequestNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "팀 변경 신청을 찾을 수 없습니다",
			})
		}
		slog.Error("TeamChange.ReviewRequest: failed to get request", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "신청 정보를 불러오는데 실패했습니다",
		})
	}

	// Verify the request belongs to the specified league
	if changeRequest.LeagueID == nil || *changeRequest.LeagueID != leagueID {
		return c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "not_found",
			Message: "해당 리그의 팀 변경 신청이 아닙니다",
		})
	}

	// Check if already processed
	if changeRequest.Status != model.TeamChangeStatusPending {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "already_processed",
			Message: "이미 처리된 신청입니다",
		})
	}

	// Check if reviewer is a director of the target team
	directorTeams, err := h.participantRepo.GetDirectorTeams(ctx, leagueID, userID)
	if err != nil {
		slog.Error("TeamChange.ReviewRequest: failed to get director teams", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "권한 확인에 실패했습니다",
		})
	}

	isTargetTeamDirector := false
	for _, teamName := range directorTeams {
		if teamName == changeRequest.RequestedTeamName {
			isTargetTeamDirector = true
			break
		}
	}

	if !isTargetTeamDirector {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "해당 팀의 디렉터만 이적 신청을 승인/거절할 수 있습니다",
		})
	}

	// Process the request
	if req.Status == model.TeamChangeStatusApproved {
		// Re-check player count before approving (in case it changed)
		participant, err := h.participantRepo.GetByID(ctx, changeRequest.ParticipantID)
		if err != nil {
			slog.Error("TeamChange.ReviewRequest: failed to get participant", "error", err)
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "server_error",
				Message: "참가자 정보를 불러오는데 실패했습니다",
			})
		}

		hasPlayerRole := false
		for _, role := range participant.Roles {
			if role == string(model.RolePlayer) {
				hasPlayerRole = true
				break
			}
		}

		if hasPlayerRole {
			playerCount, err := h.participantRepo.CountPlayersByTeam(ctx, leagueID, changeRequest.RequestedTeamName)
			if err != nil {
				slog.Error("TeamChange.ReviewRequest: failed to count players", "error", err)
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

		// Approve and update team
		if err := h.teamChangeRepo.ApproveTeamChange(ctx, requestID, userID); err != nil {
			slog.Error("TeamChange.ReviewRequest: failed to approve", "error", err)
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "server_error",
				Message: "팀 변경 승인에 실패했습니다",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "팀 변경 신청이 승인되었습니다",
		})
	} else {
		// Reject
		if err := h.teamChangeRepo.RejectTeamChange(ctx, requestID, userID, req.Reason); err != nil {
			slog.Error("TeamChange.ReviewRequest: failed to reject", "error", err)
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "server_error",
				Message: "팀 변경 거절에 실패했습니다",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "팀 변경 신청이 거절되었습니다",
		})
	}
}

// CancelRequest handles DELETE /api/v1/leagues/:id/team-change-requests/:requestId
func (h *TeamChangeHandler) CancelRequest(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	requestIDStr := c.Param("requestId")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 신청 ID입니다",
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

	// Get the team change request
	changeRequest, err := h.teamChangeRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, repository.ErrTeamChangeRequestNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "not_found",
				Message: "팀 변경 신청을 찾을 수 없습니다",
			})
		}
		slog.Error("TeamChange.CancelRequest: failed to get request", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "신청 정보를 불러오는데 실패했습니다",
		})
	}

	// Verify the request belongs to the specified league
	if changeRequest.LeagueID == nil || *changeRequest.LeagueID != leagueID {
		return c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "not_found",
			Message: "해당 리그의 팀 변경 신청이 아닙니다",
		})
	}

	// Check if user owns this request
	participant, err := h.participantRepo.GetByLeagueAndUser(ctx, leagueID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrParticipantNotFound) {
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error:   "forbidden",
				Message: "본인의 신청만 취소할 수 있습니다",
			})
		}
		slog.Error("TeamChange.CancelRequest: failed to get participant", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "참가자 정보를 불러오는데 실패했습니다",
		})
	}

	if changeRequest.ParticipantID != participant.ID {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "본인의 신청만 취소할 수 있습니다",
		})
	}

	// Check if request is pending
	if changeRequest.Status != model.TeamChangeStatusPending {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "cannot_cancel",
			Message: "대기 중인 신청만 취소할 수 있습니다",
		})
	}

	// Delete the request
	if err := h.teamChangeRepo.DeleteRequest(ctx, requestID); err != nil {
		slog.Error("TeamChange.CancelRequest: failed to delete request", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "팀 변경 신청 취소에 실패했습니다",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "팀 변경 신청이 취소되었습니다",
	})
}
