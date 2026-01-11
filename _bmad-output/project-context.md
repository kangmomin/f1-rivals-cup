---
project_name: 'F1 Rivals Cup'
user_name: 'Chm48'
date: '2026-01-10'
sections_completed: ['technology_stack', 'language_rules', 'framework_rules', 'testing_rules', 'security_rules', 'database_rules', 'api_rules', 'anti_patterns']
status: 'complete'
rule_count: 35
optimized_for_llm: true
---

# Project Context for AI Agents

_F1 Rivals Cup 프로젝트의 구현 규칙. AI 에이전트는 코드 작성 시 반드시 이 규칙을 따라야 합니다._

---

## Technology Stack & Versions

### Backend (Go)
- **Go**: 1.21+ (slog 사용을 위해 필수)
- **Echo**: v4
- **sqlc**: latest (pgx/v5 드라이버)
- **pgx**: v5.x
- **golang-migrate**: v4.x
- **golang-jwt/jwt**: v5
- **bcrypt**: 표준 라이브러리

### Frontend (React)
- **React**: 18
- **Vite**: latest
- **TypeScript**: strict mode
- **Tailwind CSS**: latest (커스텀 테마)
- **React Hook Form**: latest
- **Zod**: latest
- **Axios**: latest
- **React Router**: v6

### Infrastructure
- **PostgreSQL**: 14+ (JSONB, GIN 인덱스)
- **Docker Compose**: 3.8
- **GitHub Actions**: CI/CD
- **Air**: Hot Reload (개발용)

---

## Critical Implementation Rules

### Go Backend Rules

**레이어드 아키텍처 (필수):**
- Handler → Service → Repository 단방향 의존성
- Handler는 Service만 호출, Repository 직접 호출 금지
- Service는 Repository만 호출
- 레이어 간 순환 의존성 절대 금지

**sqlc + JSONB 패턴:**
- JSONB 필드는 `pgtype.JSONB` 타입 사용
- 복잡한 JSONB 연산자(`@>`, `?`, `?|`)는 raw SQL 쿼리로 작성
- Permission 검색: `WHERE permissions @> '["user.manage"]'::jsonb`

**네이밍 규칙:**
- 패키지명: lowercase, 단일 단어 (`handler`, `service`)
- 파일명: snake_case (`auth_handler.go`)
- Struct: PascalCase (`MemberService`)
- Interface: -er 접미사 (`MemberRepository`)
- Error: Err 접두사 (`ErrNotFound`)
- JSON 태그: snake_case (`json:"member_id"`)

**에러 처리:**
```go
const (
    ErrCodeNotFound             = "NOT_FOUND"
    ErrCodeUnauthorized         = "UNAUTHORIZED"
    ErrCodeForbidden            = "FORBIDDEN"
    ErrCodeInsufficientPermission = "INSUFFICIENT_PERMISSION"
    ErrCodeValidation           = "VALIDATION_ERROR"
    ErrCodeConflict             = "CONFLICT"
)
```

### React Frontend Rules

**컴포넌트 구조:**
- 기능별 폴더: `components/auth/`, `components/admin/`
- 공통 컴포넌트: `components/common/`
- 테스트 co-location: `LoginForm.tsx` + `LoginForm.test.tsx`

**네이밍 규칙:**
- 컴포넌트 파일: PascalCase (`UserTable.tsx`)
- 유틸 파일: camelCase (`authUtils.ts`)
- Hook: use 접두사 (`useAuth`, `useMembers`)
- 상수: SCREAMING_SNAKE_CASE (`API_BASE_URL`)

**상태 관리:**
- 전역 인증: Context API (`AuthContext`)
- 로컬 상태: useState
- 서버 상태: 직접 fetch (SWR/React Query 미사용)

**로딩 상태 네이밍:**
```typescript
const [isLoading, setIsLoading] = useState(false);
const [isSubmitting, setIsSubmitting] = useState(false);
```

### API Rules

**엔드포인트 형식:**
- Base: `/api/v1/`
- Resource: 복수형, kebab-case (`/members`, `/permission-histories`)
- Query params: snake_case (`?page=1&per_page=20`)

**응답 형식:**
```json
// 성공
{ "data": {...} }

// 목록
{ "data": [...], "meta": { "page": 1, "per_page": 20, "total": 45 } }

// 에러
{ "error": { "code": "INSUFFICIENT_PERMISSION", "message": "...", "details": {...} } }
```

**날짜 형식:**
- API 전송: ISO 8601 UTC (`2026-01-10T14:30:00Z`)
- DB 저장: TIMESTAMPTZ
- UI 표시: 로컬 시간 (`2026년 1월 10일`)

### Database Rules

**테이블 네이밍:**
- 테이블: snake_case, 복수형 (`members`, `permission_histories`)
- 컬럼: snake_case (`member_id`, `created_at`)
- PK: `id` (BIGSERIAL)
- FK: `{table_singular}_id` (`member_id`)
- Index: `idx_{table}_{columns}`
- Constraint: `{table}_{type}_{column}`

**필수 컬럼:**
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `version INTEGER NOT NULL DEFAULT 1` (Optimistic lock)

### Testing Rules

**Go 테스트:**
- 동일 디렉토리에 `_test.go` 파일
- 통합 테스트: `tests/integration/`
- testcontainers-go로 실제 PostgreSQL 사용

**React 테스트:**
- 동일 위치에 `.test.tsx`
- MSW로 API 모킹
- Vitest + Testing Library

### Security Rules

**JWT 토큰:**
- Access Token: React 메모리 (state/context)에만 저장
- Refresh Token: HttpOnly + Secure + SameSite=Strict Cookie
- Access 만료: 15-30분
- Refresh 만료: 7일

**JWT Claims 구조:**
```go
type Claims struct {
    MemberID    int64    `json:"member_id"`
    Email       string   `json:"email"`
    Role        string   `json:"role"`
    Permissions []string `json:"permissions"`
    jwt.RegisteredClaims
}
```

**Permission 코드:**
- 형식: `{domain}.{action}`
- 예시: `user.view`, `user.manage`, `user.role.change`
- Wildcard: `*`

---

## Critical Don't-Miss Rules

### NEVER DO (금지 사항)

1. **Handler에서 Repository 직접 호출 금지** - 반드시 Service 경유
2. **Access Token을 localStorage/Cookie에 저장 금지** - 메모리만 사용
3. **JSONB 컬럼에 복잡한 중첩 객체 사용 금지** - 단순 문자열 배열만
4. **테이블명 단수형 사용 금지** - 항상 복수형 (`member` ❌ → `members` ✅)
5. **API 응답에서 camelCase JSON 사용 금지** - snake_case 사용
6. **버전 컬럼 없이 UPDATE 금지** - Optimistic lock 필수

### ALWAYS DO (필수 사항)

1. **모든 API 에러는 표준 형식 사용** - `{ error: { code, message, details } }`
2. **권한 변경 시 History 기록** - `permission_histories` 테이블에 자동 기록
3. **날짜는 ISO 8601 형식으로 API 전송** - `2026-01-10T14:30:00Z`
4. **Soft delete 사용** - `deleted_at` 컬럼, 물리 삭제 금지
5. **Go JSON 태그에 snake_case** - `json:"member_id"`
6. **React 컴포넌트와 테스트 파일 co-location**

---

## File Structure Reference

```
f1-rivals-cup/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── domain/
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repository/
│   │   ├── middleware/
│   │   └── errors/
│   └── db/
│       ├── migrations/
│       └── queries/
└── frontend/
    └── src/
        ├── components/{common,auth,admin}/
        ├── contexts/
        ├── hooks/
        ├── pages/
        ├── services/
        ├── types/
        └── utils/
```
