---
stepsCompleted: ["step-01-init", "step-02-discovery", "step-03-success", "step-04-journeys", "step-07-project-type", "step-08-scoping", "step-09-functional", "step-10-nonfunctional", "step-11-complete"]
workflowStatus: "complete"
completedAt: "2026-01-10"
inputDocuments:
  - "C:\\projects\\f1 rivals cup\\기획서.md"
  - "C:\\projects\\f1 rivals cup\\Design.md"
workflowType: 'prd'
lastStep: 9
documentCounts:
  briefCount: 0
  researchCount: 0
  brainstormingCount: 0
  projectDocsCount: 2
---

# Product Requirements Document - f1 rivals cup

**Author:** Chm48
**Date:** 2026-01-10

## Executive Summary

F1 Rivals Cup은 F1 팀 감독들을 위한 리그 매니지먼트 시스템입니다. 참가자들은 팀을 운영하고, 자금을 관리하며, 경기 결과를 추적하고, 리그 뉴스를 공유하는 플랫폼입니다.

현재 시스템은 권한 분리가 없어 모든 사용자가 동일한 접근 권한을 가지고 있습니다. 이번 구현은 **확장 가능한 역할 기반 권한 시스템**을 도입하여, 향후 추가될 모든 기능(뉴스 발행, 자금 관리, 경기 결과 등록 등)에 대해 세밀한 접근 제어를 할 수 있는 기반을 마련합니다.

**구현 범위:**
- Member 테이블 기반 권한 시스템 (Role: USER, STAFF, ADMIN)
- 동적 권한(permissions) 배열 관리 (PostgreSQL JSONB + GIN 인덱스)
- Go enum으로 권한 코드 중앙 관리 (예: `news.create`, `fund.manage`)
- ADMIN 전용 유저 관리 페이지 (유저 조회, 검색, Role 변경, Permission 편집)

**권한 모델:**
- **조회**: 모든 role이 가능 (관리 기능 제외)
- **수정/삭제/등록**: 해당 기능의 permission 필요
- **ADMIN**: 와일드카드(`*`) 권한으로 모든 기능 접근 (1-3개 계정 제한)

### What Makes This Special

기존 RBAC 설계안(5개 테이블 구조)에서 **실용적 단순화**로 방향을 전환했습니다. 복잡한 Role-Permission 매핑 테이블 대신, Member 테이블에 role과 permissions를 직접 저장하여 조회 성능을 최적화하면서도, 향후 확장성은 JSONB와 동적 권한 코드로 확보했습니다.

**핵심 가치:**

1. **미래 확장성**: 새로운 기능 추가 시 Go enum에 권한 코드만 추가하면 즉시 세밀한 접근 제어 가능
2. **성능 최적화**: JSONB + GIN 인덱스로 권한 검색 속도 보장
3. **명확한 운영 구조**:
   - ADMIN (1-3명): 슈퍼유저, 모든 권한
   - STAFF: 특정 권한만 가진 운영자 (예: 뉴스 담당)
   - USER: 일반 참가자, 조회 중심 + 필요 권한만 부여

이를 통해 리그 초기에는 단순하게 운영하되, 성장에 따라 권한을 유연하게 추가할 수 있는 확장 가능한 기반을 확보합니다.

## Project Classification

**Technical Type:** Web App (SaaS Dashboard)
**Domain:** Gaming (F1 League Management)
**Complexity:** Medium
**Project Context:** Brownfield - extending existing system

**Existing Tech Stack:**
- Backend: Go (Echo framework), PostgreSQL
- Architecture: RESTful API with sqlc query generation
- Design: Dark mode first, Carbon & Neon theme, F1 telemetry-inspired UI

**New Feature Integration:**
이 권한 시스템은 향후 모든 기능(자금 관리, 뉴스 발행, 경기 결과 등록 등)의 접근 제어 기반이 됩니다. Echo middleware를 통한 권한 체크 패턴을 확립하며, JSONB 기반의 유연한 권한 관리로 기능 추가 시 스키마 변경 없이 확장 가능합니다.

**Technical Decisions:**
- **Permissions Storage:** PostgreSQL JSONB with GIN index (검색 성능 + 확장성)
- **Permission Codes:** Go const enum for centralized management
- **ADMIN Wildcard:** Application-layer handling of `["*"]` permission

## Success Criteria

### User Success

**권한별 접근 제어가 올바르게 작동:**
- **ADMIN**: 유저 관리 페이지에 접근하여 모든 유저 조회, 검색, Role 변경, Permission 편집, 권한 변경 히스토리 조회 가능
- **STAFF**: 부여된 특정 permissions만으로 기능 접근 가능
- **USER**: 관리 기능 제외한 조회는 자유롭게, 수정/삭제/등록은 해당 permission 있을 때만 가능
- **권한 없는 사용자**: 접근이 차단되고 명확한 권한 오류 메시지 표시 (표준 에러 형식)

**성공 순간:**
사용자가 자신의 권한 범위를 명확히 인지하고, 권한 밖의 기능은 접근조차 할 수 없으며, ADMIN이 권한 변경 이력을 로그로 추적할 수 있을 때.

### Business Success

**향후 기능 확장 기반 확보:**
이 권한 시스템 완성 후, 뉴스 발행, 자금 관리, 경기 결과 등록 등 모든 신규 기능에 즉시 권한 체크를 적용할 수 있는 기반 확보.

**확장 가능성 검증 방법 (문서 기반):**
- 새로운 기능 추가 시: Go enum에 permission 코드 추가 → Echo middleware 적용 → 테스트 작성/통과 프로세스가 30분~1시간 내 완료 가능함을 문서화
- 샘플 시나리오: "뉴스 조회 API 추가" 과정을 PRD에 예시로 기술하여 확장 패턴 명시

**운영 효율성:**
- 예상 ADMIN 수: 1-3명 (기술적 강제 아님, 운영 예측치)
- 권한 변경 히스토리로 감사 로그 확보 (보관 기간: 3년)
- 새로운 기능 추가 시 스키마 변경 없이 enum 추가만으로 권한 관리

### Technical Success

**테스트 주도 개발(TDD) 적용:**
- **필수 요구사항**: 모든 기능은 TDD(Test-Driven Development) 방식으로 개발
- **Red-Green-Refactor 사이클**: 실패하는 테스트 작성 → 구현 → 테스트 통과 → 리팩토링
- **테스트 작성 원칙**:
  - 테스트는 비즈니스 로직과 요구사항에 따라 작성되며, 명세(specification) 역할을 함
  - 아래 "핵심 테스트 시나리오"를 기반으로 테스트 먼저 작성
- **테스트 코드 수정 금지 원칙**:
  - 테스트가 실패했을 때, 독자적 판단으로 테스트 코드를 수정해서는 안 됨
  - 테스트 실패는 구현 코드의 문제이거나 요구사항 재검증이 필요한 신호
  - 예외: 요구사항 변경 시 PM/Analyst와 협의 후 테스트 수정, 테스트 자체에 버그가 있는 경우 리뷰 후 수정

**핵심 테스트 시나리오:**

```
권한 체크 테스트:
- Given: USER role with permissions ["news.read"]
  When: Accessing endpoint requiring "user.manage"
  Then: Return 403 Forbidden with standard error format

- Given: ADMIN with permissions ["*"]
  When: Accessing any endpoint
  Then: Always pass permission check

- Given: STAFF with permissions ["user.view"]
  When: Accessing user list page
  Then: Pass permission check

히스토리 기록 테스트:
- Given: ADMIN changes USER's role from USER to STAFF
  When: Permission change committed
  Then: History record created with (changer_id, target_id, old_value, new_value, timestamp)

- Given: ADMIN views user history
  When: Fetching history with LIMIT 10
  Then: Return latest 10 records ordered by created_at DESC

동시성 테스트:
- Given: Two ADMINs editing same user permissions simultaneously
  When: Second request arrives while first is processing
  Then: Second request fails with "다른 관리자가 수정 중입니다" error
```

**아키텍처 목표:**
- PostgreSQL JSONB + GIN 인덱스로 권한 조회 성능 최적화
- Echo middleware 패턴으로 모든 엔드포인트에 일관된 권한 체크 적용
- Go const enum으로 권한 코드 중앙 관리하여 타입 안정성 확보
- Row-level locking + Optimistic locking으로 동시성 처리

**에러 처리 표준:**
```json
{
  "error": {
    "code": "INSUFFICIENT_PERMISSION",
    "message": "이 작업을 수행할 권한이 없습니다",
    "required_permission": "user.manage",
    "details": {
      "user_role": "USER",
      "user_permissions": ["news.read"]
    }
  }
}
```

### Measurable Outcomes (Acceptance Criteria)

**기능 검증:**
- [ ] USER가 ADMIN 전용 페이지 접근 시 403 Forbidden + 표준 에러 형식 반환
- [ ] ADMIN이 유저 목록 조회, 검색(이름/이메일), Role 변경, Permission 편집 가능
- [ ] 권한 변경 시 히스토리 기록 생성 (변경자, 대상 유저, 변경 전/후, 타임스탬프)
- [ ] 히스토리 조회 시 최신순 정렬 + LIMIT으로 성능 최적화
- [ ] `permissions = ["*"]` ADMIN은 모든 권한 체크 통과
- [ ] STAFF가 부여된 permission만 접근 가능
- [ ] 동시 권한 수정 시 Optimistic locking으로 충돌 방지
- [ ] 모든 권한 체크 기능에 대한 단위 테스트 작성 및 통과 (TDD)
- [ ] 통합 테스트로 전체 권한 플로우 검증

**성능 기준:**
- 권한 조회 쿼리 50ms 이하 (JSONB + GIN 인덱스 활용)
- 히스토리 조회(LIMIT 10) 100ms 이하

**데이터 관리:**
- 히스토리 보관 기간: 생성일로부터 3년

## Product Scope

### MVP - Minimum Viable Product

**필수 구현 항목:**

1. **Member 테이블 및 권한 모델**
   - Role: USER, STAFF, ADMIN (enum)
   - Permissions: JSONB 배열 + GIN 인덱스
   - Go const enum으로 permission 코드 관리:
     - `user.manage`: 유저 정보 수정
     - `user.role.change`: Role 변경
     - `user.permission.edit`: Permission 편집
     - `user.view`: 유저 조회 (ADMIN 페이지)

2. **Permission History 테이블**
   - 변경자 ID, 대상 유저 ID, 변경 전/후 값, 타임스탬프
   - 인덱스: `(member_id, created_at DESC)`
   - 보관 기간: 3년

3. **ADMIN 유저 관리 페이지**
   - 유저 목록 조회 (페이지네이션)
   - 검색 기능 (이름, 이메일)
   - Role 변경 UI
   - Permission 편집 UI (추가/제거)
   - 권한 변경 히스토리 조회 (최신 10개)
   - Carbon & Neon 디자인 시스템 적용

4. **Echo Middleware 권한 체크**
   - `RequirePermission(permCode)` middleware
   - `RequireRole(role)` middleware
   - ADMIN 와일드카드 처리 로직
   - 표준 에러 응답 형식

5. **sqlc 쿼리 생성**
   - Member CRUD
   - 권한 조회/검증 쿼리 (JSONB 연산자 활용)
   - 히스토리 기록/조회 쿼리
   - Optimistic locking 지원

6. **TDD 테스트 스위트**
   - 권한 체크 로직 단위 테스트
   - Middleware 테스트
   - 히스토리 기록 테스트
   - 동시성 테스트
   - 통합 테스트

7. **확장 패턴 문서화**
   - 새로운 기능 추가 시 권한 연동 가이드
   - 샘플 시나리오: "뉴스 조회 API 권한 추가" 예시

### Growth Features (Post-MVP)

**향후 고려사항:**
- 권한 템플릿 (STAFF에게 자주 부여하는 권한 세트 저장)
- 권한 만료 기능 (시간 제한 권한)
- 권한 요청/승인 워크플로우
- 감사 로그 고도화 (모든 API 호출 기록)
- 히스토리 롤백 기능

### Vision (Future)

**장기 비전:**
F1 Rivals Cup의 모든 기능(뉴스, 자금, 경기 결과, 텔레메트리)에 대해 세밀한 권한 제어가 가능한 통합 권한 시스템. 리그 규모 확대에 따라 복잡한 역할 구조(팀별 권한, 시즌별 권한)도 유연하게 지원.

## User Journeys

### Journey 1: 김준호 (ADMIN) - 리그 성장과 권한 통제의 균형

김준호는 F1 Rivals Cup을 6개월 전 친구 5명과 시작한 리그 창립자입니다. 처음에는 소규모였기에 모두가 자유롭게 시스템을 사용했습니다. 하지만 입소문을 타고 참가자가 30명으로 늘어나면서 문제가 생기기 시작했습니다. 한 참가자가 실수로 다른 팀의 자금을 수정했고, 또 다른 참가자는 공식 뉴스로 착각하고 개인 의견을 올렸습니다.

어느 토요일 아침, 준호는 시스템에 로그인하여 새로 추가된 유저 관리 페이지를 발견합니다. 30명의 참가자 목록이 깔끔하게 정렬되어 나타나고, 각자의 현재 role과 permissions가 한눈에 보입니다. 그는 먼저 뉴스를 자주 작성하는 박지민을 찾아 검색창에 이름을 입력합니다. 박지민의 프로필을 열고 Role을 USER에서 STAFF로 변경한 후, permissions 섹션에서 `news.create`와 `news.publish` 권한을 추가합니다. 저장 버튼을 누르자 히스토리에 "김준호가 박지민의 권한을 변경함"이라는 로그가 자동으로 기록됩니다.

다음 주, 또 다른 ADMIN인 최수진이 실수로 잘못된 권한을 부여했다는 제보를 받습니다. 준호는 해당 유저의 권한 변경 히스토리를 열어 최근 10개 변경 내역을 확인합니다. 언제, 누가, 무엇을 변경했는지 타임스탬프와 함께 명확하게 기록되어 있어, 문제를 즉시 파악하고 올바른 권한으로 수정합니다.

3개월 후, 리그는 50명으로 성장했고 준호는 권한 시스템 덕분에 자신 있게 리그를 운영합니다. 각 참가자는 자신의 역할에 맞는 권한만 가지고 있고, 모든 권한 변경은 투명하게 추적됩니다. 실수로 인한 혼란은 완전히 사라졌습니다.

### Journey 2: 박지민 (STAFF) - 명확한 역할로 집중하는 뉴스 에디터

박지민은 F1 Rivals Cup의 열렬한 팬이자 글쓰기를 좋아하는 참가자입니다. 그는 경기 리뷰와 팀 인터뷰 기사를 작성하는 것을 즐기지만, 일반 USER로는 뉴스를 직접 발행할 수 없어 항상 ADMIN에게 요청해야 했습니다. 김준호가 그의 재능을 알아보고 뉴스 담당 STAFF 역할을 제안했고, 지민은 기쁘게 수락했습니다.

월요일 아침, 지민이 시스템에 로그인하자 이전에는 보이지 않던 "뉴스 관리" 메뉴가 나타납니다. 그는 주말 경기 결과를 바탕으로 리뷰 기사를 작성하고, 이번에는 직접 "발행" 버튼을 클릭할 수 있습니다. 기사가 즉시 메인 페이지에 게시되고, 다른 참가자들의 댓글이 달리기 시작합니다.

어느 날, 지민은 호기심에 "유저 관리" 메뉴에 접근해봅니다. 하지만 시스템은 명확한 메시지를 보여줍니다: "이 작업을 수행할 권한이 없습니다. 필요 권한: user.manage" 그는 자신의 역할이 뉴스 관리에만 집중되어 있다는 것을 이해하고, 오히려 안심합니다. 다른 참가자의 정보를 실수로 건드릴 걱정이 없기 때문입니다.

지민은 이제 매주 3-4개의 고품질 기사를 자신 있게 발행하고, 리그의 공식 뉴스 담당자로서 명확한 역할을 수행합니다. 그는 자신의 권한 범위 내에서 자유롭게 창작할 수 있고, 불필요한 기능에는 접근조차 하지 않아 집중할 수 있습니다.

### Journey 3: 이태훈 (USER) - 안전하게 즐기는 리그 참가자

이태훈은 F1을 사랑하는 직장인으로, Rivals Cup에서 Mercedes 팀을 운영합니다. 그는 기술에 익숙하지 않아 복잡한 시스템을 어려워하지만, 경기를 보고 자신의 팀 전적을 확인하는 것만으로도 충분히 즐겁습니다.

수요일 저녁, 태훈은 퇴근 후 시스템에 접속하여 이번 주 경기 일정을 확인합니다. 다른 팀들의 순위와 자금 현황을 조회하고, 박지민이 작성한 경기 리뷰 기사를 읽습니다. 모든 정보가 잘 정리되어 있고, 그는 편안하게 리그를 즐깁니다.

어느 날, 태훈은 실수로 다른 참가자 프로필의 "수정" 버튼을 클릭합니다. 하지만 시스템은 즉시 "권한이 없습니다"라는 메시지를 보여주며 접근을 차단합니다. 태훈은 처음에는 당황했지만, 곧 "아, 내가 실수로 다른 사람 정보를 망칠 뻔했구나"라고 안도합니다. 시스템이 자신의 실수로부터 보호해준다는 것을 깨닫습니다.

태훈은 자신의 팀 정보를 업데이트하고, 경기 결과를 조회하고, 뉴스를 읽는 데 필요한 모든 것을 자유롭게 할 수 있습니다. 하지만 다른 사람의 데이터나 관리 기능에는 접근할 수 없어, 실수로 무언가를 망칠 걱정 없이 안전하게 리그를 즐깁니다. 그는 이제 시스템을 더 자신 있게 사용하며, F1 Rivals Cup이 자신 같은 일반 사용자도 편안하게 참여할 수 있는 곳이라고 느낍니다.

### Journey Requirements Summary

이 3가지 여정을 통해 다음과 같은 기능 요구사항이 도출됩니다:

**유저 관리 기능 (ADMIN):**
- 전체 유저 목록 조회 (페이지네이션)
- 유저 검색 (이름, 이메일)
- Role 변경 UI (USER ↔ STAFF ↔ ADMIN)
- Permission 추가/제거 UI
- 권한 변경 히스토리 조회 (최신 10개, 타임스탬프 포함)
- 변경 사항 자동 로그 기록

**권한 기반 메뉴/기능 접근 (STAFF):**
- Role과 permission에 따라 메뉴 표시/숨김
- 부여된 권한으로 기능 실행 (예: 뉴스 작성/발행)
- 권한 없는 기능 접근 시 명확한 에러 메시지
- 자신의 권한 범위 인지 가능

**안전한 접근 제어 (USER):**
- 조회 기능은 자유롭게 접근
- 수정/삭제/관리 기능은 권한 체크로 차단
- 실수 방지를 위한 즉각적인 권한 에러 표시
- 자신의 데이터는 자유롭게 관리 가능

**공통 요구사항:**
- 명확한 권한 에러 메시지 (표준 형식)
- Role별 차별화된 사용자 경험
- 투명한 권한 변경 추적 시스템

## Web App Specific Requirements

### Project-Type Overview

F1 Rivals Cup의 Member 권한 시스템은 **React 18 기반 Single Page Application (SPA)**으로 구축됩니다. 최신 브라우저를 타겟으로 하며, JWT 인증을 사용한 stateless 아키텍처입니다. 반응형 디자인으로 모바일에서도 사용 가능하지만 주 사용 환경은 PC입니다.

### Technical Architecture Considerations

**Frontend Stack:**
- **Framework**: React 18 (stable 최신 버전)
- **Routing**: React Router v6
- **State Management**: Context API (인증 상태, 권한 정보)
- **HTTP Client**: Axios 또는 Fetch API
- **Styling**: CSS-in-JS 또는 Tailwind CSS (Carbon & Neon 디자인 시스템 적용)

**Backend Integration:**
- Go Echo 프레임워크로 RESTful API 제공
- JWT 기반 인증 (Access Token + Refresh Token)
- Echo middleware에서 권한 체크

**Authentication Flow:**
1. 로그인 → JWT Access Token 발급 (짧은 만료 시간)
2. Refresh Token으로 Access Token 갱신
3. Context API로 전역 인증 상태 관리
4. Protected Route로 권한별 페이지 접근 제어

### Browser Support Matrix

**지원 브라우저:**
- Chrome (최신 2개 버전)
- Firefox (최신 2개 버전)
- Safari (최신 2개 버전)
- Edge Chromium (최신 2개 버전)

**명시적 비지원:**
- Internet Explorer (모든 버전)
- Legacy Edge (EdgeHTML)

**기술적 함의:**
- Modern JavaScript (ES6+) 전체 사용
- CSS Grid, Flexbox 지원
- No Polyfills 필요

### API Endpoints

**인증 API:**
```
POST /api/auth/register - 회원가입
POST /api/auth/login - 로그인
POST /api/auth/refresh - 토큰 갱신
POST /api/auth/find-email - 이메일 찾기
POST /api/auth/reset-password - 비밀번호 재설정
```

**유저 관리 API (ADMIN):**
```
GET /api/members?page=1&limit=20&search=keyword - 유저 목록 조회
GET /api/members/:id - 특정 유저 조회
PUT /api/members/:id/role - Role 변경
PUT /api/members/:id/permissions - Permission 편집
GET /api/members/:id/history?limit=10 - 권한 변경 히스토리
```

### Route Structure

```
/ - 메인 페이지 (공개)
/login - 로그인
/register - 회원가입
/find-email - 이메일 찾기
/reset-password - 비밀번호 재설정
/admin/users - ADMIN 유저 관리 (ADMIN 전용)
/schedule - 경기 일정 (공개)
/news - 뉴스 (공개)
```

**Protected Route 구현:**
- Context에서 role/permissions 확인
- 권한 없으면 `/login`으로 리다이렉트 또는 403 페이지

### Responsive Design Requirements

**디바이스 지원:**
- **Desktop**: 1920x1080 (기본), 1366x768 (최소)
- **Tablet**: 768px 이상
- **Mobile**: 375px 이상

**반응형 전략:**
- Desktop First 접근
- Breakpoints: 1920px, 1366px, 768px, 375px
- 모든 기능 모바일에서도 사용 가능
- 주 사용 환경은 PC (UX 최적화 우선순위)

**디자인 시스템:**
- Carbon & Neon 다크모드 테마
- F1 텔레메트리 스타일 UI
- Racing Red (#FF3B30): 버튼, 아이콘, Live 배지 (텍스트 X)
- Electric Blue (#0A84FF): 링크, 활성 탭
- 주요 텍스트: White (#FFFFFF)

### SEO Strategy

**최소 SEO (Meta Tags만):**

**공개 페이지:**
- 메인, 경기 일정, 뉴스
- `<title>`, `<meta description>` 설정
- OpenGraph 태그 (선택적)

**비공개 페이지:**
- 로그인, 회원가입, ADMIN 페이지
- `<meta name="robots" content="noindex">`

**구현 방식:**
- Client-side rendering (순수 SPA)
- React Helmet으로 Meta 태그 관리

### Accessibility Requirements

**WCAG 2.1 Level AA 준수:**

**색상 대비:**
- Racing Red (#FF3B30): **버튼/강조 요소만 사용** (텍스트로 사용 금지)
- 텍스트: White (#FFFFFF) on Carbon Black (#121212) - 충분한 대비
- Electric Blue (#0A84FF): 링크/버튼 배경에 흰색 텍스트

**키보드 네비게이션:**
- Tab 키로 모든 인터랙티브 요소 접근
- Enter/Space로 버튼 활성화
- Escape로 모달 닫기

**스크린 리더:**
- ARIA labels (예: `aria-label="유저 검색"`)
- ARIA roles (예: `role="navigation"`)
- Form labels 필수

**포커스 표시:**
- 명확한 포커스 아웃라인 (Electric Blue)

### Localization

**언어:**
- 한글만 지원
- i18n 불필요
- 에러 메시지, UI 텍스트 모두 한글

### Performance Targets

**로딩 성능:**
- First Contentful Paint: 3초 이내
- Time to Interactive: 5초 이내
- API 응답: 50ms (권한 조회), 100ms (히스토리)

**최적화 전략:**
- Code Splitting (React.lazy)
- 라우트별 Lazy Loading
- 이미지 최적화 (WebP)
- 번들 크기 제한 없음 (최적화만)

**실시간 업데이트:**
- 불필요 (명시적 제외)
- 권한 변경 시 수동 새로고침 또는 재조회

### Implementation Considerations

**권한 기반 UI:**
- Context에서 `user.role`, `user.permissions` 제공
- 권한별 메뉴 표시/숨김
- Protected Route로 페이지 접근 제어

**폼 검증:**
- Client-side: 즉시 피드백 (React Hook Form)
- Server-side: 최종 검증 (보안)

**에러 처리:**
- 표준 에러 형식 파싱
- 한글 에러 메시지 표시
- Toast 알림 (성공/실패)

### MVP Scope Update

**MVP Phase 1 - 인증 시스템:**
1. 메인 페이지
2. 회원가입 (이메일, 비밀번호, 이름)
3. 로그인 (JWT 발급)
4. 이메일 찾기 (이름, 전화번호로 찾기)
5. 비밀번호 재설정 (이메일 인증 링크)

**MVP Phase 2 - 권한 관리:**
6. ADMIN 유저 관리 페이지
7. Role 변경, Permission 편집
8. 권한 변경 히스토리 조회

## Project Scoping & Phased Development

### MVP Strategy & Philosophy

**MVP Approach:** Platform MVP (통합 출시)

F1 Rivals Cup의 권한 시스템은 인증과 권한 관리를 **한 번에 출시**하여, 첫 릴리즈부터 완전한 권한 제어가 작동합니다. Phase를 분리하면 인증만으로는 사용자 가치가 없으므로, 통합 개발로 효율성을 높입니다.

**Resource Requirements:**
- **개발팀**: 1-3명 (풀스택 또는 백엔드 1명 + 프론트엔드 1명)
- **필수 기술**: Go (Echo), React 18, PostgreSQL 12+, JWT
- **예상 개발 기간**: 4주 (인증 + 권한 통합, TDD 테스트 포함)

### MVP Feature Set (통합 출시)

**Core User Journeys Supported:**
- **ADMIN**: 유저 관리 페이지에서 권한 설정 및 히스토리 추적
- **STAFF**: 부여받은 권한으로 특정 기능 접근
- **USER**: 안전한 조회 중심 사용, 실수 방지

**MVP - 인증 + 권한 관리 (Must-Have):**

**인증 시스템:**
1. 메인 페이지 (공개, 간단한 소개)
2. 회원가입 (이메일, 비밀번호, 이름)
3. 로그인 (JWT Access + Refresh Token)
4. 이메일 찾기 (이름, 전화번호 인증)
5. 비밀번호 재설정 (이메일 인증 링크)

**권한 관리 시스템:**
6. Member 테이블 (role + permissions JSONB) + Permission History 테이블
7. ADMIN 유저 관리 페이지
   - 유저 목록 조회 (페이지네이션, 검색)
   - Role 변경 UI
   - Permission 편집 UI (추가/제거)
8. 권한 변경 히스토리 조회 (최신 10개)
9. Echo Middleware 권한 체크 (`RequirePermission`, `RequireRole`)
10. TDD 테스트 스위트 (단위, 통합, 동시성 테스트)

**개발 일정 (4주):**
- 1주차: DB 스키마 + API 골격
- 2주차: 인증 API + Middleware
- 3주차: ADMIN React 페이지 + 권한 UI
- 4주차: TDD 테스트 작성 + 버그 수정

### Post-MVP Features

**Phase 2 (공개 기능 - 유저 유입):**
- 경기 일정 페이지 (공개, 로그인 전 접근)
- 뉴스 페이지 (공개)
- 팀 로그 (공개)
- React Helmet으로 Meta 태그 관리 (기본 SEO)

**Phase 3 (Growth Features - 운영 효율화):**
- 권한 템플릿 저장 (STAFF에게 자주 부여하는 권한 세트)
- 권한 만료 기능 (시간 제한 권한)
- 권한 요청/승인 워크플로우
- 감사 로그 고도화 (모든 API 호출 기록)
- 히스토리 롤백 기능

**Phase 4 (Expansion - 기능 확장):**
- 뉴스 발행 시스템 (권한: `news.create`, `news.publish`)
- 자금 관리 시스템 (권한: `fund.manage`)
- 경기 결과 등록 (권한: `match.edit`)

**Phase 5 (장기 비전 - 텔레메트리):**
- Go UDP 패킷 수신기
- 실시간 웹소켓 중계
- 텔레메트리 데이터 시각화

### Risk Mitigation Strategy

**Technical Risks:**
- **성능 목표**: 전체 API 요청 처리 시간 100-150ms (JWT 검증 + 권한 조회 + 비즈니스 로직)
  - 권한 조회 쿼리만: 50ms 이하 (JSONB + GIN 인덱스)
- **완화 방법**:
  - 초기부터 GIN 인덱스 설계 및 적용
  - 성능 테스트를 TDD에 포함 (벤치마크 자동화)
  - PostgreSQL 12+ 사용 필수 (JSONB 최적화 지원)
- **리스크 가정**:
  - JSONB 배열 크기 제한 (최대 20-30개 권한)
  - PostgreSQL 버전이 12 미만이면 업그레이드 필요 (마이그레이션 비용 추가)

**Market Risks:**
- **가장 큰 리스크**: 사용자가 권한 시스템을 복잡하게 느낄 수 있음
- **MVP 대응**:
  - 단순화된 3-role 모델 (USER, STAFF, ADMIN)
  - 직관적인 Carbon & Neon 디자인
  - 명확한 한글 에러 메시지
- **검증 방법**: 베타 테스터 5-10명과 초기 피드백 수집 (2주 베타 테스트)

**Resource Risks:**
- **최소 팀 크기**: 1명 풀스택 개발자 (Go + React 경험)
- **축소 시나리오**:
  - 이메일 찾기 / 비밀번호 재설정 기능 제외 (나중에 추가)
  - 히스토리 조회를 간소화 (최신 5개만)
  - ADMIN 페이지 UI를 기본 테이블로 (디자인 단순화)
- **절대 최소 기능**: 로그인 + ADMIN 유저 관리 (role 변경만) + 기본 권한 체크

**Contingency Plan:**
- 개발 지연 시: 이메일/비밀번호 찾기 기능을 Phase 2로 이동
- 성능 이슈 시: JSONB 대신 별도 `member_permissions` 테이블로 마이그레이션 (스키마 변경 비용 발생)
- 테스트 작성 부담 시: 핵심 권한 체크 로직만 TDD, 나머지는 통합 테스트로 커버

## Functional Requirements

### User Authentication

- FR1: 방문자는 이메일, 비밀번호, 이름으로 회원가입할 수 있다
- FR2: 방문자는 이메일과 비밀번호로 로그인할 수 있다
- FR3: 사용자는 로그아웃할 수 있다
- FR4: 시스템은 로그인 시 JWT Access Token을 발급한다
- FR5: 시스템은 Refresh Token으로 Access Token을 갱신할 수 있다
- FR6: 방문자는 이름과 전화번호로 이메일을 찾을 수 있다
- FR7: 방문자는 이메일 인증 링크로 비밀번호를 재설정할 수 있다
- FR8: 시스템은 비밀번호 재설정 링크에 만료 시간을 적용한다
- FR9: 시스템은 회원가입 시 이메일 중복 여부를 검증한다
- FR10: 시스템은 비밀번호를 암호화하여 저장한다
- FR11: 시스템은 로그인 실패 시 명확한 에러 메시지를 표시한다

### User Profile

- FR12: 사용자는 자신의 프로필 정보를 조회할 수 있다
- FR13: 사용자는 자신의 이름을 수정할 수 있다
- FR14: 사용자는 자신의 비밀번호를 변경할 수 있다
- FR15: 사용자는 자신의 계정을 탈퇴 요청할 수 있다

### User Management

- FR16: ADMIN은 전체 유저 목록을 조회할 수 있다
- FR17: ADMIN은 유저 목록을 페이지네이션으로 탐색할 수 있다
- FR18: ADMIN은 이름으로 유저를 검색할 수 있다
- FR19: ADMIN은 이메일로 유저를 검색할 수 있다
- FR20: ADMIN은 특정 유저의 상세 정보를 조회할 수 있다
- FR21: ADMIN은 유저의 Role을 변경할 수 있다 (USER, STAFF, ADMIN)
- FR22: ADMIN은 유저에게 Permission을 추가할 수 있다
- FR23: ADMIN은 유저의 Permission을 제거할 수 있다
- FR24: ADMIN은 유저 목록을 Role별로 필터링할 수 있다
- FR25: ADMIN은 유저 목록을 가입일순으로 정렬할 수 있다

### Account Status Management

- FR26: ADMIN은 유저 계정을 비활성화할 수 있다 (status 변경)
- FR27: ADMIN은 비활성화된 계정을 다시 활성화할 수 있다
- FR28: 시스템은 비활성화된 계정의 로그인을 차단한다
- FR29: ADMIN은 유저 계정을 soft delete할 수 있다 (deleted_at 설정)
- FR30: 시스템은 soft delete된 계정을 목록에서 제외한다
- FR31: 시스템은 모든 삭제를 soft delete로 처리한다

### Access Control

- FR32: 시스템은 모든 API 요청에서 인증 상태를 확인한다
- FR33: 시스템은 Role 기반 접근 제어를 수행한다
- FR34: 시스템은 Permission 기반 접근 제어를 수행한다
- FR35: ADMIN은 와일드카드 권한(*)으로 모든 기능에 접근할 수 있다
- FR36: 시스템은 권한 없는 접근 시 403 Forbidden을 반환한다
- FR37: USER는 조회 기능에 자유롭게 접근할 수 있다 (관리 기능 제외)

### Permission Management

- FR38: 시스템은 Go enum으로 Permission 코드를 중앙 관리한다
- FR39: 시스템은 JSONB 배열로 유저별 Permission을 저장한다
- FR40: 시스템은 GIN 인덱스로 Permission 검색 성능을 최적화한다
- FR41: 새로운 Permission 추가 시 스키마 변경 없이 enum만 추가한다

### Audit & History

- FR42: 시스템은 모든 권한 변경을 히스토리에 기록한다
- FR43: 히스토리는 변경자, 대상 유저, 변경 전/후 값, 타임스탬프를 포함한다
- FR44: ADMIN은 특정 유저의 권한 변경 히스토리를 조회할 수 있다
- FR45: 히스토리 조회는 최신순으로 정렬된다
- FR46: 히스토리 조회는 기본 10개씩 LIMIT으로 제한된다
- FR47: 시스템은 히스토리를 3년간 보관한다

### User Interface

- FR48: 시스템은 Carbon & Neon 다크모드 디자인을 적용한다
- FR49: 시스템은 반응형 디자인으로 모바일 접근을 지원한다
- FR50: 시스템은 Role과 Permission에 따라 메뉴를 표시/숨김한다
- FR51: ADMIN은 유저 관리 페이지에 접근할 수 있다
- FR52: 유저 관리 페이지는 테이블 형태로 유저 목록을 표시한다
- FR53: 유저 관리 페이지는 Role 변경 UI를 제공한다
- FR54: 유저 관리 페이지는 Permission 편집 UI를 제공한다
- FR55: 유저 관리 페이지는 권한 변경 히스토리 UI를 제공한다

### Error Handling & Feedback

- FR56: 시스템은 표준 에러 형식으로 에러 응답을 반환한다
- FR57: 에러 응답은 에러 코드, 메시지, 필요 권한을 포함한다
- FR58: 시스템은 모든 에러 메시지를 한글로 표시한다
- FR59: 시스템은 성공/실패에 대한 Toast 알림을 표시한다
- FR60: 시스템은 폼 검증 실패 시 즉시 피드백을 제공한다
- FR61: 시스템은 권한 오류 시 명확한 안내 메시지를 표시한다

### System Operations

- FR62: 시스템은 동시 권한 수정 시 Optimistic locking으로 충돌을 방지한다
- FR63: 시스템은 충돌 발생 시 "다른 관리자가 수정 중입니다" 메시지를 표시한다

## Non-Functional Requirements

### Performance

**API 응답 시간 (측정 조건: 개발 환경, P95 기준):**
- 전체 API 요청 처리: 100-150ms 이내 (JWT 검증 + 권한 조회 + 비즈니스 로직)
- 권한 조회 쿼리: 50ms 이하 (JSONB + GIN 인덱스)
- 히스토리 조회 (LIMIT 10): 100ms 이하
- 동시 접속자 기준: 20명 동시 요청 시 성능 유지

**Frontend 성능 (측정 조건: 광대역 네트워크, 개발 환경):**
- First Contentful Paint (FCP): 3초 이내
- Time to Interactive (TTI): 5초 이내
- 유저 목록 페이지 로딩: 2초 이내
- 로딩 인디케이터: 버튼 클릭 후 100ms 이내 표시

**최적화 전략:**
- React.lazy를 활용한 Code Splitting
- 라우트별 Lazy Loading
- JSONB + GIN 인덱스로 권한 검색 최적화
- 페이지네이션으로 대량 데이터 분할 조회

### Security

**인증 보안:**
- JWT Access Token: 15-30분 만료
- Refresh Token: 7일 만료, Secure + HttpOnly Cookie 저장
- 비밀번호: bcrypt 또는 argon2로 해시 처리
- 비밀번호 재설정 링크: 1시간 만료

**접근 제어:**
- Echo Middleware에서 모든 보호된 엔드포인트 권한 검증
- Role + Permission 이중 검증
- ADMIN 와일드카드(*) 권한 애플리케이션 레이어에서 처리
- 권한 오류 시 403 Forbidden + 표준 에러 형식

**데이터 보안:**
- HTTPS 전용 통신
- SQL Injection 방지 (sqlc parameterized queries)
- XSS 방지 (React 기본 이스케이핑)
- CSRF 방지 (JWT + SameSite Cookie)

**감사 추적:**
- 모든 권한 변경 히스토리 기록
- 변경자, 대상, 변경 전/후, 타임스탬프 포함
- 히스토리 보관 기간: 3년

### Scalability

**사용자 규모:**
- 총 등록 유저: 100명+ 지원
- 피크 타임 동시 접속자: 30명 (경기 시간대 기준)
- JSONB 권한 배열: 최대 20-30개 권한 per 유저

**확장 전략:**
- JSONB 기반으로 스키마 변경 없이 권한 추가
- Go enum으로 Permission 코드 중앙 관리
- 페이지네이션으로 대량 유저 조회 지원
- 인덱스 설계로 조회 성능 유지

**제약 사항 및 대응:**
- PostgreSQL 14+ 필수 (JSONB 최적화 및 장기 지원)
- JSONB 권한 수 30개 초과 시: 별도 `member_permissions` 테이블 마이그레이션 검토
- 동시 접속자 50명 초과 예상 시: 커넥션 풀링 및 캐시 레이어 도입 검토

### Accessibility

**WCAG 2.1 Level AA 준수:**

**색상 대비:**
- 본문 텍스트: White (#FFFFFF) on Carbon Black (#121212) - 대비 15:1+
- Racing Red (#FF3B30): 버튼/강조 요소만 사용 (텍스트 사용 금지)
- Electric Blue (#0A84FF): 링크, 버튼 배경에 흰색 텍스트

**키보드 접근성:**
- Tab 키로 모든 인터랙티브 요소 순회
- Enter/Space로 버튼 활성화
- Escape로 모달/드롭다운 닫기
- Skip to main content 링크 제공

**스크린 리더 지원:**
- ARIA labels 필수 적용 (예: `aria-label="유저 검색"`)
- ARIA roles 적용 (예: `role="navigation"`)
- Form 요소에 label 연결 필수
- 에러 메시지 aria-live로 알림

**포커스 관리:**
- 명확한 포커스 아웃라인 (Electric Blue)
- 모달 열림 시 포커스 트래핑
- 작업 완료 시 적절한 포커스 이동

**반응형 레이아웃:**
- Desktop: 테이블 레이아웃 (유저 관리 페이지)
- Mobile (768px 이하): 카드 레이아웃으로 전환

### Reliability

**동시성 처리:**
- Optimistic locking으로 동시 권한 수정 충돌 방지
- Row-level locking으로 데이터 무결성 보장
- 충돌 시 명확한 에러 메시지: "다른 관리자가 수정 중입니다"

**데이터 무결성:**
- 모든 삭제는 Soft delete (deleted_at)
- 외래 키 제약으로 참조 무결성 보장
- 트랜잭션으로 원자성 보장

**가용성 목표:**
- 목표 가용성: 99% (계획된 점검 제외)
- 계획된 점검: 월 1회, 사전 공지
- 비계획 다운타임: 월 7시간 이하

**에러율 목표:**
- 5xx 서버 에러: 전체 API 호출의 0.1% 이하
- 4xx 클라이언트 에러: 로깅 및 모니터링 (목표 없음)

**백업 및 복구:**
- 데이터베이스 백업: 일일 자동 백업
- 백업 보관: 7일간 보관
- RTO (복구 목표 시간): 24시간 이내
- RPO (복구 시점 목표): 24시간 이내 (일일 백업 기준)

**오류 복구:**
- 권한 변경 히스토리로 수동 롤백 가능
- 에러 로깅으로 문제 추적
- Graceful error handling으로 사용자 경험 유지
- 에러 발생 시 사용자에게 다음 행동 안내 제공
