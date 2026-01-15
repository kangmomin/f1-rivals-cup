package service

import "errors"

// Common service errors
var (
	// League errors
	ErrLeagueNotFound      = errors.New("league not found")
	ErrInvalidDateFormat   = errors.New("invalid date format")
	ErrInvalidStatus       = errors.New("invalid status")
	ErrInvalidLeagueStatus = errors.New("invalid league status")
	ErrInvalidLeagueName   = errors.New("invalid league name")

	// Participant errors
	ErrParticipantNotFound  = errors.New("participant not found")
	ErrAlreadyParticipant   = errors.New("already participant")
	ErrLeagueNotOpen        = errors.New("league is not open for registration")
	ErrInvalidRole          = errors.New("invalid role")
	ErrNoRolesProvided      = errors.New("no roles provided")
	ErrTeamFull             = errors.New("team player limit reached")
	ErrCannotCancelApproved = errors.New("cannot cancel approved participation")
	ErrNilRequest           = errors.New("request cannot be nil")

	// Team errors
	ErrTeamNotFound      = errors.New("team not found")
	ErrTeamAlreadyExists = errors.New("team already exists in this league")
	ErrTeamNameRequired  = errors.New("team name is required")

	// Match errors
	ErrMatchNotFound        = errors.New("match not found")
	ErrDuplicateRound       = errors.New("duplicate round in league")
	ErrInvalidMatchRequest  = errors.New("invalid match request")

	// Finance errors
	ErrAccountNotFound        = errors.New("account not found")
	ErrInsufficientBalance    = errors.New("insufficient balance")
	ErrSameAccount            = errors.New("cannot transfer to same account")
	ErrNotInLeague            = errors.New("account not in league")
	ErrAccountNotInLeague     = errors.New("account not in league")
	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrAmountMustBePositive   = errors.New("amount must be positive")
	ErrNotApproved            = errors.New("participant not approved")
	ErrSystemAccountForbidden = errors.New("cannot use system account")
	ErrNotDirector            = errors.New("user is not a director")
	ErrNotTeamDirector        = errors.New("not director of this team")
	ErrNotAccountOwner        = errors.New("not owner of this account")

	// News errors
	ErrNewsNotFound         = errors.New("news not found")
	ErrNewsEmptyTitle       = errors.New("title is required")
	ErrNewsTitleTooShort    = errors.New("title must be at least 2 characters")
	ErrNewsTitleTooLong     = errors.New("title must be at most 200 characters")
	ErrNewsEmptyContent     = errors.New("content is required")
	ErrNewsEmptyInput       = errors.New("input is required")
	ErrNewsAIUnavailable    = errors.New("AI service is not configured")
	ErrNewsLeagueNotFound   = errors.New("league not found")

	// Comment errors
	ErrCommentNotFound = errors.New("comment not found")
)
