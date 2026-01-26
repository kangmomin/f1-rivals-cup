package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

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
	activityRepo    *repository.TeamChangeActivityRepository
}

func NewTeamChangeHandler(
	teamChangeRepo *repository.TeamChangeRepository,
	participantRepo *repository.ParticipantRepository,
	teamRepo *repository.TeamRepository,
	leagueRepo *repository.LeagueRepository,
	activityRepo *repository.TeamChangeActivityRepository,
) *TeamChangeHandler {
	return &TeamChangeHandler{
		teamChangeRepo:  teamChangeRepo,
		participantRepo: participantRepo,
		teamRepo:        teamRepo,
		leagueRepo:      leagueRepo,
		activityRepo:    activityRepo,
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

	// Validate requested roles if provided
	validRoles := map[string]bool{
		string(model.RoleDirector): true,
		string(model.RolePlayer):   true,
		string(model.RoleReserve):  true,
		string(model.RoleEngineer): true,
	}
	for _, role := range req.RequestedRoles {
		if !validRoles[role] {
			return c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_role",
				Message: "유효하지 않은 역할입니다: " + role,
			})
		}
	}

	// Determine effective roles (use requested_roles if provided, otherwise keep current)
	effectiveRoles := req.RequestedRoles
	if len(effectiveRoles) == 0 {
		effectiveRoles = []string(participant.Roles)
	}

	// Check player count in target team (max 2) if becoming/remaining a player
	willBePlayer := false
	for _, role := range effectiveRoles {
		if role == string(model.RolePlayer) {
			willBePlayer = true
			break
		}
	}

	if willBePlayer {
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
		CurrentRoles:      participant.Roles,
		Status:            model.TeamChangeStatusPending,
		Reason:            req.Reason,
	}
	// Only set requested_roles if different from current
	if len(req.RequestedRoles) > 0 {
		teamChangeReq.RequestedRoles = req.RequestedRoles
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

	// Log activity (non-blocking)
	details := map[string]any{
		"requested_team_name": req.RequestedTeamName,
		"current_roles":       []string(participant.Roles),
	}
	if participant.TeamName != nil {
		details["current_team_name"] = *participant.TeamName
	}
	if len(req.RequestedRoles) > 0 {
		details["requested_roles"] = req.RequestedRoles
	}
	if req.Reason != nil {
		details["reason"] = *req.Reason
	}
	activityLog := &model.TeamChangeActivityLog{
		ActorID:       userID,
		RequestID:     teamChangeReq.ID,
		ParticipantID: participant.ID,
		ActionType:    model.TeamChangeActionCreate,
		Details:       details,
	}
	if err := h.activityRepo.Create(ctx, activityLog); err != nil {
		slog.Error("TeamChange: failed to log activity", "error", err)
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
		// Determine effective roles after approval
		effectiveRoles := changeRequest.RequestedRoles
		if len(effectiveRoles) == 0 {
			effectiveRoles = changeRequest.CurrentRoles
		}

		// Re-check player count before approving (in case it changed)
		willBePlayer := false
		for _, role := range effectiveRoles {
			if role == string(model.RolePlayer) {
				willBePlayer = true
				break
			}
		}

		if willBePlayer {
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

		// Log activity (non-blocking)
		details := map[string]any{
			"requested_team_name": changeRequest.RequestedTeamName,
		}
		if changeRequest.CurrentTeamName != nil {
			details["current_team_name"] = *changeRequest.CurrentTeamName
		}
		if len(changeRequest.CurrentRoles) > 0 {
			details["current_roles"] = []string(changeRequest.CurrentRoles)
		}
		if len(changeRequest.RequestedRoles) > 0 {
			details["requested_roles"] = []string(changeRequest.RequestedRoles)
		}
		activityLog := &model.TeamChangeActivityLog{
			ActorID:       userID,
			RequestID:     requestID,
			ParticipantID: changeRequest.ParticipantID,
			ActionType:    model.TeamChangeActionApprove,
			Details:       details,
		}
		if err := h.activityRepo.Create(ctx, activityLog); err != nil {
			slog.Error("TeamChange: failed to log activity", "error", err)
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

		// Log activity (non-blocking)
		details := map[string]any{
			"requested_team_name": changeRequest.RequestedTeamName,
		}
		if changeRequest.CurrentTeamName != nil {
			details["current_team_name"] = *changeRequest.CurrentTeamName
		}
		if len(changeRequest.CurrentRoles) > 0 {
			details["current_roles"] = []string(changeRequest.CurrentRoles)
		}
		if len(changeRequest.RequestedRoles) > 0 {
			details["requested_roles"] = []string(changeRequest.RequestedRoles)
		}
		activityLog := &model.TeamChangeActivityLog{
			ActorID:       userID,
			RequestID:     requestID,
			ParticipantID: changeRequest.ParticipantID,
			ActionType:    model.TeamChangeActionReject,
			Details:       details,
		}
		if err := h.activityRepo.Create(ctx, activityLog); err != nil {
			slog.Error("TeamChange: failed to log activity", "error", err)
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

	// Prepare activity log details before deletion
	details := map[string]any{
		"requested_team_name": changeRequest.RequestedTeamName,
	}
	if changeRequest.CurrentTeamName != nil {
		details["current_team_name"] = *changeRequest.CurrentTeamName
	}
	if len(changeRequest.CurrentRoles) > 0 {
		details["current_roles"] = []string(changeRequest.CurrentRoles)
	}
	if len(changeRequest.RequestedRoles) > 0 {
		details["requested_roles"] = []string(changeRequest.RequestedRoles)
	}

	// Delete the request
	if err := h.teamChangeRepo.DeleteRequest(ctx, requestID); err != nil {
		slog.Error("TeamChange.CancelRequest: failed to delete request", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "팀 변경 신청 취소에 실패했습니다",
		})
	}

	// Log activity (non-blocking) - Note: log will be cascade-deleted with request
	activityLog := &model.TeamChangeActivityLog{
		ActorID:       userID,
		RequestID:     requestID,
		ParticipantID: participant.ID,
		ActionType:    model.TeamChangeActionCancel,
		Details:       details,
	}
	if err := h.activityRepo.Create(ctx, activityLog); err != nil {
		// Expected to fail after request deletion due to FK constraint
		slog.Debug("TeamChange: activity log for cancel skipped (request deleted)", "error", err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "팀 변경 신청이 취소되었습니다",
	})
}

// ListActivity handles GET /api/v1/admin/leagues/:id/team-change-activity
func (h *TeamChangeHandler) ListActivity(c echo.Context) error {
	leagueIDStr := c.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "잘못된 리그 ID입니다",
		})
	}

	// Parse pagination parameters
	page := 1
	limit := 20
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	ctx := c.Request().Context()

	activities, total, err := h.activityRepo.ListByLeague(ctx, leagueID, page, limit)
	if err != nil {
		slog.Error("TeamChange.ListActivity: failed to list activities", "error", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "server_error",
			Message: "활동 로그를 불러오는데 실패했습니다",
		})
	}

	if activities == nil {
		activities = []*model.TeamChangeActivityLog{}
	}

	return c.JSON(http.StatusOK, model.TeamChangeActivityListResponse{
		Activities: activities,
		Total:      total,
		Page:       page,
		Limit:      limit,
	})
}
