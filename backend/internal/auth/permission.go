package auth

// Role represents a user role in the system
type Role string

const (
	RoleUser  Role = "USER"
	RoleStaff Role = "STAFF"
	RoleAdmin Role = "ADMIN"
)

// IsValid checks if the role is valid
func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleStaff, RoleAdmin:
		return true
	}
	return false
}

// Permission represents a permission code in the system
type Permission string

// User management permissions
const (
	PermUserView           Permission = "user.view"           // View user list (ADMIN page)
	PermUserManage         Permission = "user.manage"         // Edit user info
	PermUserRoleChange     Permission = "user.role.change"    // Change user role
	PermUserPermissionEdit Permission = "user.permission.edit" // Edit user permissions
)

// News permissions (for future use)
const (
	PermNewsCreate  Permission = "news.create"  // Create news
	PermNewsPublish Permission = "news.publish" // Publish news
	PermNewsEdit    Permission = "news.edit"    // Edit news
	PermNewsDelete  Permission = "news.delete"  // Delete news
)

// Fund management permissions (for future use)
const (
	PermFundView   Permission = "fund.view"   // View fund info
	PermFundManage Permission = "fund.manage" // Manage funds
)

// Match permissions (for future use)
const (
	PermMatchEdit   Permission = "match.edit"   // Edit match info
	PermMatchResult Permission = "match.result" // Enter match results
)

// League permissions
const (
	PermLeagueCreate Permission = "league.create" // Create league
	PermLeagueEdit   Permission = "league.edit"   // Edit league
	PermLeagueDelete Permission = "league.delete" // Delete league
)

// Wildcard permission for ADMIN
const PermWildcard Permission = "*"

// AllPermissions returns all available permission codes
func AllPermissions() []Permission {
	return []Permission{
		// User management
		PermUserView,
		PermUserManage,
		PermUserRoleChange,
		PermUserPermissionEdit,
		// News
		PermNewsCreate,
		PermNewsPublish,
		PermNewsEdit,
		PermNewsDelete,
		// Fund
		PermFundView,
		PermFundManage,
		// Match
		PermMatchEdit,
		PermMatchResult,
		// League
		PermLeagueCreate,
		PermLeagueEdit,
		PermLeagueDelete,
	}
}

// PermissionInfo provides description for a permission
type PermissionInfo struct {
	Code        Permission `json:"code"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
}

// GetPermissionInfo returns information about all permissions
func GetPermissionInfo() []PermissionInfo {
	return []PermissionInfo{
		// User management
		{PermUserView, "유저 조회", "관리자 페이지에서 유저 목록 조회", "user"},
		{PermUserManage, "유저 관리", "유저 정보 수정", "user"},
		{PermUserRoleChange, "역할 변경", "유저의 역할(Role) 변경", "user"},
		{PermUserPermissionEdit, "권한 편집", "유저의 권한(Permission) 추가/제거", "user"},
		// News
		{PermNewsCreate, "뉴스 작성", "뉴스 기사 작성", "news"},
		{PermNewsPublish, "뉴스 발행", "뉴스 기사 발행", "news"},
		{PermNewsEdit, "뉴스 수정", "뉴스 기사 수정", "news"},
		{PermNewsDelete, "뉴스 삭제", "뉴스 기사 삭제", "news"},
		// Fund
		{PermFundView, "자금 조회", "팀 자금 정보 조회", "fund"},
		{PermFundManage, "자금 관리", "팀 자금 수정", "fund"},
		// Match
		{PermMatchEdit, "경기 수정", "경기 정보 수정", "match"},
		{PermMatchResult, "결과 입력", "경기 결과 입력", "match"},
		// League
		{PermLeagueCreate, "리그 생성", "새 리그 생성", "league"},
		{PermLeagueEdit, "리그 수정", "리그 정보 수정", "league"},
		{PermLeagueDelete, "리그 삭제", "리그 삭제", "league"},
	}
}

// IsValid checks if the permission code is valid
func (p Permission) IsValid() bool {
	if p == PermWildcard {
		return true
	}
	for _, perm := range AllPermissions() {
		if p == perm {
			return true
		}
	}
	return false
}

// HasPermission checks if a permission slice contains the required permission
func HasPermission(permissions []string, required Permission) bool {
	for _, p := range permissions {
		// Wildcard grants all permissions
		if Permission(p) == PermWildcard {
			return true
		}
		if Permission(p) == required {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if a permission slice contains any of the required permissions
func HasAnyPermission(permissions []string, required []Permission) bool {
	for _, req := range required {
		if HasPermission(permissions, req) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a permission slice contains all required permissions
func HasAllPermissions(permissions []string, required []Permission) bool {
	for _, req := range required {
		if !HasPermission(permissions, req) {
			return false
		}
	}
	return true
}
