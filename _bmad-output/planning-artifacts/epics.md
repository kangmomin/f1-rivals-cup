---
stepsCompleted: ["step-01-validate-prerequisites", "step-02-design-epics", "step-03-create-stories", "step-04-final-validation"]
status: complete
inputDocuments:
  - "C:\\projects\\f1 rivals cup\\_bmad-output\\planning-artifacts\\prd.md"
  - "C:\\projects\\f1 rivals cup\\_bmad-output\\planning-artifacts\\architecture.md"
  - "C:\\projects\\f1 rivals cup\\_bmad-output\\project-context.md"
workflowType: 'epics'
project_name: 'f1 rivals cup'
user_name: 'Chm48'
date: '2026-01-10'
---

# f1 rivals cup - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for f1 rivals cup, decomposing the requirements from the PRD, UX Design if it exists, and Architecture requirements into implementable stories.

## Requirements Inventory

### Functional Requirements

**User Authentication (FR1-FR11)**
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

**User Profile (FR12-FR15)**
- FR12: 사용자는 자신의 프로필 정보를 조회할 수 있다
- FR13: 사용자는 자신의 이름을 수정할 수 있다
- FR14: 사용자는 자신의 비밀번호를 변경할 수 있다
- FR15: 사용자는 자신의 계정을 탈퇴 요청할 수 있다

**User Management (FR16-FR25)**
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

**Account Status Management (FR26-FR31)**
- FR26: ADMIN은 유저 계정을 비활성화할 수 있다 (status 변경)
- FR27: ADMIN은 비활성화된 계정을 다시 활성화할 수 있다
- FR28: 시스템은 비활성화된 계정의 로그인을 차단한다
- FR29: ADMIN은 유저 계정을 soft delete할 수 있다 (deleted_at 설정)
- FR30: 시스템은 soft delete된 계정을 목록에서 제외한다
- FR31: 시스템은 모든 삭제를 soft delete로 처리한다

**Access Control (FR32-FR37)**
- FR32: 시스템은 모든 API 요청에서 인증 상태를 확인한다
- FR33: 시스템은 Role 기반 접근 제어를 수행한다
- FR34: 시스템은 Permission 기반 접근 제어를 수행한다
- FR35: ADMIN은 와일드카드 권한(*)으로 모든 기능에 접근할 수 있다
- FR36: 시스템은 권한 없는 접근 시 403 Forbidden을 반환한다
- FR37: USER는 조회 기능에 자유롭게 접근할 수 있다 (관리 기능 제외)

**Permission Management (FR38-FR41)**
- FR38: 시스템은 Go enum으로 Permission 코드를 중앙 관리한다
- FR39: 시스템은 JSONB 배열로 유저별 Permission을 저장한다
- FR40: 시스템은 GIN 인덱스로 Permission 검색 성능을 최적화한다
- FR41: 새로운 Permission 추가 시 스키마 변경 없이 enum만 추가한다

**Audit & History (FR42-FR47)**
- FR42: 시스템은 모든 권한 변경을 히스토리에 기록한다
- FR43: 히스토리는 변경자, 대상 유저, 변경 전/후 값, 타임스탬프를 포함한다
- FR44: ADMIN은 특정 유저의 권한 변경 히스토리를 조회할 수 있다
- FR45: 히스토리 조회는 최신순으로 정렬된다
- FR46: 히스토리 조회는 기본 10개씩 LIMIT으로 제한된다
- FR47: 시스템은 히스토리를 3년간 보관한다

**User Interface (FR48-FR55)**
- FR48: 시스템은 Carbon & Neon 다크모드 디자인을 적용한다
- FR49: 시스템은 반응형 디자인으로 모바일 접근을 지원한다
- FR50: 시스템은 Role과 Permission에 따라 메뉴를 표시/숨김한다
- FR51: ADMIN은 유저 관리 페이지에 접근할 수 있다
- FR52: 유저 관리 페이지는 테이블 형태로 유저 목록을 표시한다
- FR53: 유저 관리 페이지는 Role 변경 UI를 제공한다
- FR54: 유저 관리 페이지는 Permission 편집 UI를 제공한다
- FR55: 유저 관리 페이지는 권한 변경 히스토리 UI를 제공한다

**Error Handling & Feedback (FR56-FR61)**
- FR56: 시스템은 표준 에러 형식으로 에러 응답을 반환한다
- FR57: 에러 응답은 에러 코드, 메시지, 필요 권한을 포함한다
- FR58: 시스템은 모든 에러 메시지를 한글로 표시한다
- FR59: 시스템은 성공/실패에 대한 Toast 알림을 표시한다
- FR60: 시스템은 폼 검증 실패 시 즉시 피드백을 제공한다
- FR61: 시스템은 권한 오류 시 명확한 안내 메시지를 표시한다

**System Operations (FR62-FR63)**
- FR62: 시스템은 동시 권한 수정 시 Optimistic locking으로 충돌을 방지한다
- FR63: 시스템은 충돌 발생 시 "다른 관리자가 수정 중입니다" 메시지를 표시한다

### NonFunctional Requirements

**Performance**
- NFR-P1: 전체 API 요청 처리 시간 100-150ms 이내 (P95 기준)
- NFR-P2: 권한 조회 쿼리 50ms 이하 (JSONB + GIN 인덱스)
- NFR-P3: 히스토리 조회 (LIMIT 10) 100ms 이하
- NFR-P4: 동시 접속자 20명 기준 성능 유지
- NFR-P5: First Contentful Paint (FCP) 3초 이내
- NFR-P6: Time to Interactive (TTI) 5초 이내
- NFR-P7: 유저 목록 페이지 로딩 2초 이내
- NFR-P8: 로딩 인디케이터 버튼 클릭 후 100ms 이내 표시

**Security**
- NFR-S1: JWT Access Token 15-30분 만료
- NFR-S2: Refresh Token 7일 만료, Secure + HttpOnly Cookie 저장
- NFR-S3: 비밀번호 bcrypt 또는 argon2로 해시 처리
- NFR-S4: 비밀번호 재설정 링크 1시간 만료
- NFR-S5: Echo Middleware에서 모든 보호된 엔드포인트 권한 검증
- NFR-S6: HTTPS 전용 통신
- NFR-S7: SQL Injection 방지 (sqlc parameterized queries)
- NFR-S8: XSS 방지 (React 기본 이스케이핑)
- NFR-S9: CSRF 방지 (JWT + SameSite Cookie)

**Scalability**
- NFR-SC1: 총 등록 유저 100명+ 지원
- NFR-SC2: 피크 타임 동시 접속자 30명 지원
- NFR-SC3: JSONB 권한 배열 최대 20-30개 권한 per 유저
- NFR-SC4: PostgreSQL 14+ 필수

**Accessibility**
- NFR-A1: WCAG 2.1 Level AA 준수
- NFR-A2: 본문 텍스트 White (#FFFFFF) on Carbon Black (#121212) 대비 15:1+
- NFR-A3: Racing Red (#FF3B30) 버튼/강조 요소만 사용 (텍스트 사용 금지)
- NFR-A4: Tab 키로 모든 인터랙티브 요소 순회
- NFR-A5: ARIA labels 필수 적용
- NFR-A6: 명확한 포커스 아웃라인 (Electric Blue)
- NFR-A7: 모바일 (768px 이하) 카드 레이아웃으로 전환

**Reliability**
- NFR-R1: Optimistic locking으로 동시 권한 수정 충돌 방지
- NFR-R2: 모든 삭제는 Soft delete (deleted_at)
- NFR-R3: 목표 가용성 99% (계획된 점검 제외)
- NFR-R4: 5xx 서버 에러 전체 API 호출의 0.1% 이하
- NFR-R5: 데이터베이스 일일 자동 백업, 7일간 보관
- NFR-R6: RTO/RPO 24시간 이내

### Additional Requirements

**Architecture - Starter Template**
- ARCH-ST1: 수동 구성 방식 (기존 보일러플레이트 미사용)
- ARCH-ST2: Monorepo 구조 (backend/, frontend/, docker-compose.yml, Makefile)
- ARCH-ST3: Backend - Echo + sqlc + pgx/v5 직접 구성
- ARCH-ST4: Frontend - vite-react-template 기반 커스터마이징

**Architecture - Backend Structure**
- ARCH-BE1: 레이어드 아키텍처 (Handler → Service → Repository)
- ARCH-BE2: cmd/server/main.go 진입점
- ARCH-BE3: internal/ 패키지 구조 (config, domain, handler, service, repository, middleware, errors)
- ARCH-BE4: db/migrations/ 마이그레이션 파일
- ARCH-BE5: db/queries/ sqlc 쿼리 정의
- ARCH-BE6: tests/integration/ 통합 테스트

**Architecture - Frontend Structure**
- ARCH-FE1: src/components/ (common, auth, admin)
- ARCH-FE2: src/contexts/AuthContext.tsx 전역 인증 상태
- ARCH-FE3: src/hooks/ 커스텀 훅
- ARCH-FE4: src/pages/ 라우트 컴포넌트
- ARCH-FE5: src/services/api.ts Axios 인스턴스
- ARCH-FE6: src/styles/tokens.css 디자인 토큰

**Architecture - Development Tools**
- ARCH-DT1: Makefile (dev, generate, migrate, test 명령)
- ARCH-DT2: docker-compose.yml (PostgreSQL 14, MailHog)
- ARCH-DT3: Air (.air.toml) 백엔드 Hot Reload
- ARCH-DT4: sqlc.yaml 설정 (pgx/v5, emit_json_tags)

**Architecture - Testing Infrastructure**
- ARCH-TI1: Backend - testcontainers-go로 실제 PostgreSQL 테스트
- ARCH-TI2: Frontend - MSW (Mock Service Worker) API 모킹
- ARCH-TI3: Frontend - Vitest + Testing Library 컴포넌트 테스트
- ARCH-TI4: TDD (Test-Driven Development) 필수

**Architecture - API & Integration**
- ARCH-API1: REST API with /api/v1/ prefix
- ARCH-API2: echo-swagger API 문서화
- ARCH-API3: 표준 에러 형식 (code, message, details)
- ARCH-API4: JWT Claims 구조 (member_id, email, role, permissions)

**Architecture - Type Sharing**
- ARCH-TS1: Permission 코드 Go → React 수동 동기화
- ARCH-TS2: 에러 코드 Go → React 수동 동기화

### FR Coverage Map

| FR | Epic | 설명 |
|----|------|------|
| FR1 | Epic 1 | 회원가입 |
| FR2 | Epic 1 | 로그인 |
| FR3 | Epic 2 | 로그아웃 |
| FR4 | Epic 1 | JWT Access Token 발급 |
| FR5 | Epic 2 | Token Refresh |
| FR6 | Epic 3 | 이메일 찾기 |
| FR7 | Epic 3 | 비밀번호 재설정 |
| FR8 | Epic 3 | 재설정 링크 만료 |
| FR9 | Epic 1 | 이메일 중복 검증 |
| FR10 | Epic 1 | 비밀번호 암호화 |
| FR11 | Epic 1 | 로그인 에러 메시지 |
| FR12 | Epic 2 | 프로필 조회 |
| FR13 | Epic 2 | 이름 수정 |
| FR14 | Epic 2 | 비밀번호 변경 |
| FR15 | Epic 2 | 계정 탈퇴 |
| FR16 | Epic 5 | 유저 목록 조회 |
| FR17 | Epic 5 | 페이지네이션 |
| FR18 | Epic 5 | 이름 검색 |
| FR19 | Epic 5 | 이메일 검색 |
| FR20 | Epic 5 | 유저 상세 조회 |
| FR21 | Epic 5 | Role 변경 |
| FR22 | Epic 5 | Permission 추가 |
| FR23 | Epic 5 | Permission 제거 |
| FR24 | Epic 5 | Role 필터링 |
| FR25 | Epic 5 | 가입일 정렬 |
| FR26 | Epic 5 | 계정 비활성화 |
| FR27 | Epic 5 | 계정 활성화 |
| FR28 | Epic 5 | 비활성화 로그인 차단 |
| FR29 | Epic 5 | Soft delete |
| FR30 | Epic 5 | Soft delete 목록 제외 |
| FR31 | Epic 5 | 모든 삭제 Soft delete |
| FR32 | Epic 4 | 인증 상태 확인 |
| FR33 | Epic 4 | Role 기반 접근 제어 |
| FR34 | Epic 4 | Permission 기반 접근 제어 |
| FR35 | Epic 4 | ADMIN 와일드카드 |
| FR36 | Epic 4 | 403 Forbidden |
| FR37 | Epic 4 | USER 조회 접근 |
| FR38 | Epic 4 | Go enum Permission |
| FR39 | Epic 4 | JSONB Permission |
| FR40 | Epic 4 | GIN 인덱스 |
| FR41 | Epic 4 | enum 추가 확장 |
| FR42 | Epic 6 | 권한 변경 기록 |
| FR43 | Epic 6 | 히스토리 데이터 구조 |
| FR44 | Epic 6 | 히스토리 조회 |
| FR45 | Epic 6 | 최신순 정렬 |
| FR46 | Epic 6 | LIMIT 10 |
| FR47 | Epic 6 | 3년 보관 |
| FR48 | Epic 7 | Carbon & Neon 테마 |
| FR49 | Epic 7 | 반응형 디자인 |
| FR50 | Epic 7 | 메뉴 표시/숨김 |
| FR51 | Epic 7 | 유저 관리 페이지 접근 |
| FR52 | Epic 7 | 유저 테이블 |
| FR53 | Epic 7 | Role 변경 UI |
| FR54 | Epic 7 | Permission 편집 UI |
| FR55 | Epic 7 | 히스토리 UI |
| FR56 | Epic 7 | 표준 에러 형식 |
| FR57 | Epic 7 | 에러 코드/메시지 |
| FR58 | Epic 7 | 한글 에러 메시지 |
| FR59 | Epic 7 | Toast 알림 |
| FR60 | Epic 7 | 폼 검증 피드백 |
| FR61 | Epic 7 | 권한 오류 안내 |
| FR62 | Epic 6 | Optimistic locking |
| FR63 | Epic 6 | 충돌 메시지 |

## Epic List

### Epic 1: 프로젝트 기반 및 기본 인증
**목표**: 사용자가 시스템에 회원가입하고 로그인할 수 있습니다.

이 에픽은 전체 시스템의 기반을 구축하고 핵심 인증 기능을 제공합니다.
- Monorepo 구조 설정 (backend/, frontend/)
- Docker Compose (PostgreSQL 14, MailHog)
- Member 테이블 (role, permissions JSONB, GIN 인덱스)
- Permission History 테이블
- 회원가입 API (이메일, 비밀번호, 이름)
- 로그인 API (JWT Access Token 발급)
- 비밀번호 암호화 (bcrypt)
- 이메일 중복 검증
- 로그인 실패 에러 메시지

**FRs:** FR1, FR2, FR4, FR9, FR10, FR11
**ARCH:** ST1-ST4, BE1-BE6, DT1-DT4, TI1-TI4, API1-API4

---

### Epic 2: 세션 관리 및 프로필
**목표**: 사용자가 세션을 관리하고 자신의 프로필을 편집할 수 있습니다.

- Token Refresh 메커니즘 (Refresh Token → Access Token 갱신)
- 로그아웃 기능
- 프로필 조회
- 프로필 수정 (이름)
- 비밀번호 변경
- 계정 탈퇴 요청 (Soft delete)

**FRs:** FR3, FR5, FR12, FR13, FR14, FR15

---

### Epic 3: 계정 복구
**목표**: 사용자가 분실한 자격 증명을 복구할 수 있습니다.

- 이름/전화번호로 이메일 찾기
- 비밀번호 재설정 이메일 링크 발송
- 재설정 링크 만료 처리 (1시간)
- 새 비밀번호 설정

**FRs:** FR6, FR7, FR8
**참고:** PRD에서 이메일 서비스는 Phase 2로 연기 가능 (MVP Fallback: ADMIN이 직접 임시 비밀번호 설정)

---

### Epic 4: 접근 제어 시스템
**목표**: 시스템이 역할과 권한에 따라 접근을 제어합니다.

- JWT 검증 미들웨어 (모든 보호된 엔드포인트)
- Role 기반 접근 제어 미들웨어 (`RequireRole`)
- Permission 기반 접근 제어 미들웨어 (`RequirePermission`)
- ADMIN 와일드카드(*) 권한 처리
- 표준 에러 응답 (403 Forbidden)
- Permission 코드 Go enum 중앙 관리
- JSONB 권한 배열 저장
- GIN 인덱스 권한 검색 최적화

**FRs:** FR32, FR33, FR34, FR35, FR36, FR37, FR38, FR39, FR40, FR41

---

### Epic 5: ADMIN 유저 관리
**목표**: ADMIN이 모든 사용자를 조회하고 관리할 수 있습니다.

- 유저 목록 조회 API
- 페이지네이션 (page, per_page)
- 유저 검색 (이름, 이메일)
- Role 필터링
- 가입일순 정렬
- 유저 상세 정보 조회
- Role 변경 기능 (USER, STAFF, ADMIN)
- Permission 추가/제거 기능
- 계정 활성화/비활성화 (status 변경)
- 비활성화 계정 로그인 차단
- Soft delete (deleted_at)
- Soft delete 계정 목록 제외

**FRs:** FR16, FR17, FR18, FR19, FR20, FR21, FR22, FR23, FR24, FR25, FR26, FR27, FR28, FR29, FR30, FR31

---

### Epic 6: 감사 및 히스토리
**목표**: ADMIN이 모든 권한 변경 이력을 추적할 수 있습니다.

- 권한 변경 자동 기록 (Role, Permission 변경 시)
- 히스토리 데이터 구조 (변경자, 대상, 변경 전/후, 타임스탬프)
- 히스토리 조회 API (최신 10개)
- 최신순 정렬
- Optimistic locking (동시 수정 충돌 방지)
- 충돌 에러 메시지 ("다른 관리자가 수정 중입니다")
- 3년 보관 정책

**FRs:** FR42, FR43, FR44, FR45, FR46, FR47, FR62, FR63

---

### Epic 7: 사용자 인터페이스 및 경험
**목표**: 완성된 UI/UX로 모든 기능을 편리하게 사용할 수 있습니다.

- Carbon & Neon 다크모드 테마
- Tailwind CSS 설정 (tokens.css)
- 반응형 디자인 (Desktop First, Mobile 카드 레이아웃)
- 권한 기반 메뉴 표시/숨김
- ADMIN 유저 관리 페이지
  - 유저 테이블 (데스크톱)
  - 유저 카드 (모바일)
  - Role 변경 UI
  - Permission 편집 UI
  - 히스토리 조회 UI
- Toast 알림 (성공/실패)
- 폼 검증 피드백 (React Hook Form + Zod)
- 한글 에러 메시지
- 로딩 인디케이터 (100ms 이내 표시)
- 표준 에러 형식 처리
- WCAG 2.1 AA 접근성 (ARIA, 키보드 네비게이션, 포커스 관리)

**FRs:** FR48, FR49, FR50, FR51, FR52, FR53, FR54, FR55, FR56, FR57, FR58, FR59, FR60, FR61

---

## Epic Summary

| Epic | 제목 | FR 수 | 주요 가치 |
|------|-----|-------|----------|
| 1 | 프로젝트 기반 및 기본 인증 | 6 + ARCH | 회원가입/로그인 가능 |
| 2 | 세션 관리 및 프로필 | 5 | 프로필 관리 가능 |
| 3 | 계정 복구 | 3 | 자격 증명 복구 가능 |
| 4 | 접근 제어 시스템 | 10 | 권한 기반 접근 제어 |
| 5 | ADMIN 유저 관리 | 16 | 사용자 관리 가능 |
| 6 | 감사 및 히스토리 | 8 | 변경 추적 가능 |
| 7 | 사용자 인터페이스 및 경험 | 14 | 완성된 UX |
| **합계** | | **63 FRs** | |

## Epic Dependencies

```
Epic 1 (Foundation & Auth)
    ↓
Epic 2 (Session/Profile) ←── Epic 3 (Recovery)
    ↓
Epic 4 (Access Control)
    ↓
Epic 5 (Admin Management)
    ↓
Epic 6 (Audit & History)
    ↓
Epic 7 (UI/UX) ← 모든 이전 에픽의 UI 통합
```

---

# Detailed Stories

## Epic 1: 프로젝트 기반 및 기본 인증 - Stories

### Story 1.1: 프로젝트 초기 설정

**As a** 개발자,
**I want** Monorepo 구조와 개발 환경이 설정된 프로젝트를,
**So that** 백엔드와 프론트엔드 개발을 즉시 시작할 수 있습니다.

**Acceptance Criteria:**

**Given** 빈 프로젝트 디렉토리
**When** 프로젝트 초기 설정을 완료하면
**Then** Monorepo 구조가 생성된다 (backend/, frontend/, docker-compose.yml, Makefile)

**Given** docker-compose.yml이 설정된 상태
**When** `docker-compose up -d` 실행 시
**Then** PostgreSQL 14와 MailHog 컨테이너가 시작된다

**Given** Makefile이 설정된 상태
**When** `make dev` 실행 시
**Then** 백엔드(Air hot reload)와 프론트엔드(Vite dev server)가 시작된다

**Given** 백엔드 서버가 실행 중일 때
**When** `GET /health` 요청 시
**Then** 200 OK와 `{"status": "ok"}` 응답을 반환한다

---

### Story 1.2: 데이터베이스 스키마 및 sqlc 설정

**As a** 개발자,
**I want** Member와 Permission History 테이블이 생성되고 sqlc 쿼리가 준비된 상태를,
**So that** 유저 데이터와 권한 변경 이력을 저장할 수 있습니다.

**Acceptance Criteria:**

**Given** PostgreSQL 데이터베이스가 실행 중일 때
**When** `make migrate-up` 실행 시
**Then** members 테이블이 생성된다 (id, email, password_hash, name, role, permissions JSONB, status, created_at, updated_at, deleted_at, version)

**And** permission_histories 테이블이 생성된다 (id, member_id, changer_id, change_type, old_value, new_value, created_at)

**And** members 테이블에 GIN 인덱스가 생성된다 (idx_members_permissions)

**Given** sqlc.yaml 설정 파일이 있을 때
**When** `make generate` 실행 시
**Then** internal/repository/ 디렉토리에 Go 코드가 생성된다

---

### Story 1.3: 회원가입 API 구현

**As a** 방문자,
**I want** 이메일, 비밀번호, 이름으로 회원가입할 수 있기를,
**So that** 시스템에 계정을 생성하고 서비스를 이용할 수 있습니다.

**Acceptance Criteria:**

**Given** 유효한 회원가입 정보
**When** `POST /api/v1/auth/register` 요청 시
**Then** 201 Created와 생성된 유저 정보를 반환한다
**And** 비밀번호는 bcrypt로 해시되어 저장된다
**And** 기본 role은 'USER', permissions는 빈 배열로 설정된다

**Given** 이미 등록된 이메일로 회원가입 시도
**When** `POST /api/v1/auth/register` 요청 시
**Then** 409 Conflict와 "이미 사용 중인 이메일입니다" 에러를 반환한다

---

### Story 1.4: 로그인 API 구현

**As a** 등록된 사용자,
**I want** 이메일과 비밀번호로 로그인할 수 있기를,
**So that** 인증된 사용자로서 서비스를 이용할 수 있습니다.

**Acceptance Criteria:**

**Given** 등록된 사용자의 유효한 이메일과 비밀번호
**When** `POST /api/v1/auth/login` 요청 시
**Then** 200 OK와 access_token, member 정보를 반환한다
**And** Refresh Token은 HttpOnly + Secure + SameSite=Strict 쿠키로 설정된다

**Given** 잘못된 자격 증명
**When** `POST /api/v1/auth/login` 요청 시
**Then** 401 Unauthorized와 "이메일 또는 비밀번호가 올바르지 않습니다" 에러를 반환한다

**Given** 비활성화된 계정
**When** `POST /api/v1/auth/login` 요청 시
**Then** 403 Forbidden과 "비활성화된 계정입니다" 에러를 반환한다

---

## Epic 2: 세션 관리 및 프로필 - Stories

### Story 2.1: 토큰 갱신 API 구현

**As a** 로그인된 사용자,
**I want** Access Token이 만료되기 전에 자동으로 갱신되기를,
**So that** 세션이 끊기지 않고 서비스를 계속 이용할 수 있습니다.

**Acceptance Criteria:**

**Given** 유효한 Refresh Token 쿠키
**When** `POST /api/v1/auth/refresh` 요청 시
**Then** 200 OK와 새로운 Access Token을 반환한다

**Given** 만료된 Refresh Token 쿠키
**When** `POST /api/v1/auth/refresh` 요청 시
**Then** 401 Unauthorized와 "세션이 만료되었습니다" 에러를 반환한다

---

### Story 2.2: 로그아웃 API 구현

**As a** 로그인된 사용자,
**I want** 로그아웃할 수 있기를,
**So that** 세션을 안전하게 종료할 수 있습니다.

**Acceptance Criteria:**

**Given** 로그인된 상태
**When** `POST /api/v1/auth/logout` 요청 시
**Then** 200 OK를 반환하고 Refresh Token 쿠키가 삭제된다

---

### Story 2.3: 프로필 조회 API 구현

**As a** 로그인된 사용자,
**I want** 내 프로필 정보를 조회할 수 있기를,
**So that** 내 계정 정보를 확인할 수 있습니다.

**Acceptance Criteria:**

**Given** 유효한 Access Token
**When** `GET /api/v1/profile` 요청 시
**Then** 200 OK와 사용자 정보를 반환한다 (비밀번호 제외)

---

### Story 2.4: 프로필 수정 API 구현

**As a** 로그인된 사용자,
**I want** 내 이름을 수정할 수 있기를,
**So that** 프로필 정보를 최신 상태로 유지할 수 있습니다.

**Acceptance Criteria:**

**Given** 유효한 Access Token과 새로운 이름
**When** `PUT /api/v1/profile` 요청 시
**Then** 200 OK와 업데이트된 프로필을 반환한다

---

### Story 2.5: 비밀번호 변경 API 구현

**As a** 로그인된 사용자,
**I want** 비밀번호를 변경할 수 있기를,
**So that** 계정 보안을 강화할 수 있습니다.

**Acceptance Criteria:**

**Given** 유효한 Access Token, 현재 비밀번호, 새 비밀번호
**When** `PUT /api/v1/profile/password` 요청 시
**Then** 200 OK를 반환하고 비밀번호가 변경된다

**Given** 잘못된 현재 비밀번호
**When** `PUT /api/v1/profile/password` 요청 시
**Then** 400 Bad Request와 "현재 비밀번호가 올바르지 않습니다" 에러를 반환한다

---

### Story 2.6: 계정 탈퇴 API 구현

**As a** 로그인된 사용자,
**I want** 내 계정을 탈퇴할 수 있기를,
**So that** 더 이상 서비스를 이용하지 않을 때 계정을 삭제할 수 있습니다.

**Acceptance Criteria:**

**Given** 유효한 Access Token과 현재 비밀번호 확인
**When** `DELETE /api/v1/profile` 요청 시
**Then** 200 OK를 반환하고 계정이 Soft delete 처리된다

---

## Epic 3: 계정 복구 - Stories

### Story 3.1: 이메일 찾기 API 구현

**As a** 이메일을 잊어버린 방문자,
**I want** 이름과 전화번호로 이메일을 찾을 수 있기를,
**So that** 내 계정 이메일을 확인하고 로그인할 수 있습니다.

**Acceptance Criteria:**

**Given** 등록된 사용자의 이름과 전화번호
**When** `POST /api/v1/auth/find-email` 요청 시
**Then** 200 OK와 마스킹된 이메일을 반환한다 (ho***@example.com)

**Given** 일치하는 사용자가 없을 때
**When** `POST /api/v1/auth/find-email` 요청 시
**Then** 404 Not Found를 반환한다

---

### Story 3.2: 비밀번호 재설정 요청 API 구현

**As a** 비밀번호를 잊어버린 방문자,
**I want** 이메일로 비밀번호 재설정 링크를 받을 수 있기를,
**So that** 비밀번호를 재설정하고 다시 로그인할 수 있습니다.

**Acceptance Criteria:**

**Given** 등록된 이메일 주소
**When** `POST /api/v1/auth/reset-password` 요청 시
**Then** 200 OK를 반환하고 재설정 링크가 이메일로 전송된다
**And** 재설정 토큰 만료 시간은 1시간으로 설정된다

---

### Story 3.3: 비밀번호 재설정 확인 API 구현

**As a** 비밀번호 재설정 링크를 받은 방문자,
**I want** 링크를 통해 새 비밀번호를 설정할 수 있기를,
**So that** 계정에 다시 접근할 수 있습니다.

**Acceptance Criteria:**

**Given** 유효한 재설정 토큰과 새 비밀번호
**When** `POST /api/v1/auth/reset-password/confirm` 요청 시
**Then** 200 OK를 반환하고 비밀번호가 변경된다

**Given** 만료된 재설정 토큰
**When** `POST /api/v1/auth/reset-password/confirm` 요청 시
**Then** 400 Bad Request와 "링크가 만료되었습니다" 에러를 반환한다

---

## Epic 4: 접근 제어 시스템 - Stories

### Story 4.1: JWT 인증 미들웨어 구현

**As a** 시스템,
**I want** 모든 보호된 API 요청에서 JWT를 검증하기를,
**So that** 인증된 사용자만 보호된 리소스에 접근할 수 있습니다.

**Acceptance Criteria:**

**Given** 유효한 JWT Access Token
**When** 보호된 엔드포인트에 접근 시
**Then** 요청이 다음 핸들러로 전달되고 Context에 사용자 정보가 저장된다

**Given** 만료되거나 유효하지 않은 Token
**When** 보호된 엔드포인트에 접근 시
**Then** 401 Unauthorized를 반환한다

---

### Story 4.2: Permission 코드 중앙 관리 구현

**As a** 개발자,
**I want** Permission 코드가 Go enum으로 중앙 관리되기를,
**So that** 타입 안정성을 확보하고 일관된 권한 체크를 할 수 있습니다.

**Acceptance Criteria:**

**Given** Permission 코드 정의가 필요할 때
**When** internal/domain/permission.go 파일을 확인하면
**Then** user.view, user.manage, user.role.change, user.permission.edit, * 코드가 정의되어 있다

---

### Story 4.3: Role 기반 접근 제어 미들웨어 구현

**As a** 시스템,
**I want** Role에 따라 API 접근을 제어하기를,
**So that** 특정 역할의 사용자만 해당 기능에 접근할 수 있습니다.

**Acceptance Criteria:**

**Given** ADMIN Role이 필요한 엔드포인트
**When** ADMIN이 아닌 사용자가 접근 시
**Then** 403 Forbidden과 에러 응답을 반환한다

---

### Story 4.4: Permission 기반 접근 제어 미들웨어 구현

**As a** 시스템,
**I want** Permission에 따라 API 접근을 제어하기를,
**So that** 특정 권한을 가진 사용자만 해당 기능에 접근할 수 있습니다.

**Acceptance Criteria:**

**Given** 특정 Permission이 필요한 엔드포인트
**When** 해당 Permission이 없는 사용자가 접근 시
**Then** 403 Forbidden과 INSUFFICIENT_PERMISSION 에러를 반환한다

---

### Story 4.5: ADMIN 와일드카드 권한 처리 구현

**As a** ADMIN 사용자,
**I want** 와일드카드(*) 권한으로 모든 기능에 접근할 수 있기를,
**So that** 모든 관리 작업을 수행할 수 있습니다.

**Acceptance Criteria:**

**Given** permissions에 "*"가 포함된 ADMIN 사용자
**When** 어떤 Permission이 필요한 엔드포인트에 접근 시
**Then** Permission 체크를 통과한다

---

### Story 4.6: 표준 에러 응답 형식 구현

**As a** 클라이언트 개발자,
**I want** 권한 오류 시 일관된 에러 형식을 받기를,
**So that** 에러 처리 로직을 일관되게 구현할 수 있습니다.

**Acceptance Criteria:**

**Given** 인증/권한 관련 에러가 발생할 때
**When** 에러 응답을 반환하면
**Then** { error: { code, message, required_permission, details } } 형식을 따른다

---

## Epic 5: ADMIN 유저 관리 - Stories

### Story 5.1: 유저 목록 조회 API 구현

**As a** ADMIN,
**I want** 전체 유저 목록을 페이지네이션으로 조회할 수 있기를,
**So that** 시스템에 등록된 모든 사용자를 확인할 수 있습니다.

**Acceptance Criteria:**

**Given** ADMIN 권한을 가진 사용자
**When** `GET /api/v1/members?page=1&per_page=20` 요청 시
**Then** 200 OK와 유저 목록, 메타 정보 (total, total_pages)를 반환한다
**And** Soft delete된 유저는 제외된다

---

### Story 5.2: 유저 검색 API 구현

**As a** ADMIN,
**I want** 이름이나 이메일로 유저를 검색할 수 있기를,
**So that** 특정 사용자를 빠르게 찾을 수 있습니다.

**Acceptance Criteria:**

**Given** ADMIN 권한을 가진 사용자
**When** `GET /api/v1/members?search=keyword` 요청 시
**Then** 이름 또는 이메일에 keyword가 포함된 유저 목록을 반환한다

---

### Story 5.3: 유저 필터링 및 정렬 API 구현

**As a** ADMIN,
**I want** Role별로 필터링하고 가입일순으로 정렬할 수 있기를,
**So that** 원하는 조건의 사용자만 조회할 수 있습니다.

**Acceptance Criteria:**

**Given** ADMIN 권한을 가진 사용자
**When** `GET /api/v1/members?role=STAFF&sort=created_at&order=desc` 요청 시
**Then** STAFF Role 유저만 최신순으로 정렬되어 반환된다

---

### Story 5.4: 유저 상세 조회 API 구현

**As a** ADMIN,
**I want** 특정 유저의 상세 정보를 조회할 수 있기를,
**So that** 사용자의 전체 정보를 확인하고 관리할 수 있습니다.

**Acceptance Criteria:**

**Given** ADMIN 권한과 유효한 member_id
**When** `GET /api/v1/members/:id` 요청 시
**Then** 200 OK와 유저 상세 정보 (version 포함)를 반환한다

---

### Story 5.5: Role 변경 API 구현

**As a** ADMIN,
**I want** 유저의 Role을 변경할 수 있기를,
**So that** 사용자에게 적절한 역할을 부여할 수 있습니다.

**Acceptance Criteria:**

**Given** ADMIN 권한과 user.role.change Permission
**When** `PUT /api/v1/members/:id/role` 요청 시
**Then** Role이 변경되고 version이 증가하고 히스토리가 기록된다

**Given** version 불일치 (다른 ADMIN이 수정)
**When** 변경 요청 시
**Then** 409 Conflict와 "다른 관리자가 수정 중입니다" 에러를 반환한다

---

### Story 5.6: Permission 관리 API 구현

**As a** ADMIN,
**I want** 유저에게 Permission을 추가하거나 제거할 수 있기를,
**So that** 세밀한 권한 제어를 할 수 있습니다.

**Acceptance Criteria:**

**Given** ADMIN 권한과 user.permission.edit Permission
**When** `PUT /api/v1/members/:id/permissions` 요청 시
**Then** Permission이 업데이트되고 히스토리가 기록된다

**Given** 유효하지 않은 Permission 코드
**When** 요청 시
**Then** 400 Bad Request를 반환한다

---

### Story 5.7: 계정 상태 관리 API 구현

**As a** ADMIN,
**I want** 유저 계정을 활성화/비활성화할 수 있기를,
**So that** 문제가 있는 계정을 일시적으로 차단할 수 있습니다.

**Acceptance Criteria:**

**Given** ADMIN 권한을 가진 사용자
**When** `PUT /api/v1/members/:id/status` 요청으로 INACTIVE 설정 시
**Then** 계정이 비활성화되고 해당 유저는 로그인이 차단된다

---

### Story 5.8: 유저 삭제 API 구현 (Soft Delete)

**As a** ADMIN,
**I want** 유저 계정을 삭제할 수 있기를,
**So that** 더 이상 사용하지 않는 계정을 정리할 수 있습니다.

**Acceptance Criteria:**

**Given** ADMIN 권한을 가진 사용자
**When** `DELETE /api/v1/members/:id` 요청 시
**Then** deleted_at 필드가 설정되고 (Soft Delete) 목록에서 제외된다

---

## Epic 6: 감사 및 히스토리 - Stories

### Story 6.1: 권한 변경 히스토리 자동 기록 구현

**As a** 시스템,
**I want** 모든 권한 변경을 자동으로 기록하기를,
**So that** 누가 언제 어떤 권한을 변경했는지 추적할 수 있습니다.

**Acceptance Criteria:**

**Given** Role, Permission, Status 변경 API가 호출될 때
**When** 변경이 성공하면
**Then** permission_histories에 (member_id, changer_id, change_type, old_value, new_value, created_at)가 기록된다

---

### Story 6.2: 히스토리 조회 API 구현

**As a** ADMIN,
**I want** 특정 유저의 권한 변경 히스토리를 조회할 수 있기를,
**So that** 해당 유저의 권한 변경 내역을 확인할 수 있습니다.

**Acceptance Criteria:**

**Given** ADMIN 권한을 가진 사용자
**When** `GET /api/v1/members/:id/history?limit=10` 요청 시
**Then** 200 OK와 히스토리 목록 (변경자 정보 포함)을 반환한다

---

### Story 6.3: 히스토리 정렬 및 페이징 구현

**As a** ADMIN,
**I want** 히스토리가 최신순으로 정렬되고 더 많은 이력을 조회할 수 있기를,
**So that** 최근 변경 사항을 먼저 확인할 수 있습니다.

**Acceptance Criteria:**

**Given** 히스토리 조회 요청
**When** API가 결과를 반환할 때
**Then** created_at DESC (최신순)로 정렬되고 기본 limit=10이 적용된다

---

### Story 6.4: Optimistic Locking 구현

**As a** 시스템,
**I want** 동시 수정 시 충돌을 방지하기를,
**So that** 두 ADMIN이 동시에 같은 유저를 수정할 때 데이터 일관성이 보장됩니다.

**Acceptance Criteria:**

**Given** 유저의 현재 version이 N일 때
**When** version=N-1로 변경 요청 시
**Then** 409 Conflict와 "다른 관리자가 수정 중입니다" 에러를 반환한다

---

### Story 6.5: 히스토리 보관 정책 구현

**As a** 시스템 관리자,
**I want** 히스토리가 3년간 보관되기를,
**So that** 감사 요구사항을 충족할 수 있습니다.

**Acceptance Criteria:**

**Given** 히스토리 테이블 설계
**When** 3년 이상 된 데이터에 대해 정리 작업이 실행되면
**Then** 해당 데이터가 삭제된다

---

## Epic 7: 사용자 인터페이스 및 경험 - Stories

### Story 7.1: 프론트엔드 기본 구조 및 디자인 시스템 설정

**As a** 개발자,
**I want** Carbon & Neon 디자인 시스템이 적용된 프론트엔드 기반을,
**So that** 일관된 디자인으로 UI 컴포넌트를 개발할 수 있습니다.

**Acceptance Criteria:**

**Given** Tailwind CSS 설정
**When** 설정을 확인하면
**Then** carbon, neon, racing 커스텀 색상이 정의되어 있다

**Given** 기본 레이아웃
**When** 페이지를 렌더링하면
**Then** 다크 모드 배경 (#121212)과 WCAG AA 대비율을 충족하는 텍스트가 적용된다

---

### Story 7.2: 인증 페이지 UI 구현

**As a** 방문자,
**I want** 회원가입/로그인 페이지를 사용할 수 있기를,
**So that** 시스템에 가입하고 접근할 수 있습니다.

**Acceptance Criteria:**

**Given** 로그인/회원가입 페이지
**When** 폼 검증 실패 시
**Then** 즉시 에러 메시지가 해당 필드 아래에 표시된다

**Given** 로그인 실패 시
**When** API 에러 응답을 받으면
**Then** Toast 알림으로 한글 에러 메시지가 표시된다

---

### Story 7.3: 인증 상태 관리 및 Protected Route 구현

**As a** 로그인된 사용자,
**I want** 인증 상태가 전역으로 관리되고 보호된 페이지에 접근할 수 있기를,
**So that** 권한에 맞는 기능만 사용할 수 있습니다.

**Acceptance Criteria:**

**Given** AuthContext 구현
**When** 사용하면
**Then** user, isAuthenticated, login, logout, refreshToken이 제공된다

**Given** 보호된 라우트
**When** 인증되지 않은 사용자가 접근 시
**Then** /login으로 리다이렉트된다

---

### Story 7.4: ADMIN 유저 목록 페이지 구현

**As a** ADMIN,
**I want** 유저 목록을 테이블/카드 형태로 볼 수 있기를,
**So that** 시스템의 모든 사용자를 한눈에 파악할 수 있습니다.

**Acceptance Criteria:**

**Given** 데스크톱 화면 (768px 이상)
**When** 유저 관리 페이지를 렌더링하면
**Then** 테이블 형태로 유저 목록이 표시된다

**Given** 모바일 화면 (768px 미만)
**When** 유저 관리 페이지를 렌더링하면
**Then** 카드 형태로 유저 목록이 표시된다

---

### Story 7.5: Role 변경 UI 구현

**As a** ADMIN,
**I want** 유저의 Role을 변경할 수 있는 UI를,
**So that** 쉽게 역할을 부여할 수 있습니다.

**Acceptance Criteria:**

**Given** Role 변경 섹션
**When** 새로운 Role을 선택하고 저장하면
**Then** 확인 모달이 표시되고 성공 시 Toast 알림이 표시된다

---

### Story 7.6: Permission 편집 UI 구현

**As a** ADMIN,
**I want** 유저의 Permission을 추가/제거할 수 있는 UI를,
**So that** 세밀하게 권한을 관리할 수 있습니다.

**Acceptance Criteria:**

**Given** Permission 편집 섹션
**When** Permission을 추가/제거하고 저장하면
**Then** API 호출로 업데이트되고 성공 시 Toast 알림이 표시된다

---

### Story 7.7: 히스토리 조회 UI 구현

**As a** ADMIN,
**I want** 유저의 권한 변경 히스토리를 볼 수 있는 UI를,
**So that** 변경 내역을 확인할 수 있습니다.

**Acceptance Criteria:**

**Given** 히스토리 섹션
**When** 렌더링하면
**Then** 타임라인 형태로 변경 내역 (일시, 변경자, 변경 유형, 이전→새 값)이 표시된다

---

### Story 7.8: 공통 UI 컴포넌트 및 피드백 시스템 구현

**As a** 사용자,
**I want** 일관된 피드백을 받을 수 있기를,
**So that** 시스템 상태를 명확히 인지할 수 있습니다.

**Acceptance Criteria:**

**Given** Toast 알림 컴포넌트
**When** 성공/실패 메시지가 발생하면
**Then** 화면 우상단에 알림이 표시되고 3초 후 사라진다

**Given** 로딩 인디케이터
**When** API 요청 중일 때
**Then** 버튼 클릭 후 100ms 이내에 로딩 상태가 표시된다

---

### Story 7.9: 접근성 및 키보드 네비게이션 구현

**As a** 스크린 리더 또는 키보드 사용자,
**I want** 접근성이 보장된 UI를,
**So that** 보조 기술로도 시스템을 사용할 수 있습니다.

**Acceptance Criteria:**

**Given** 모든 인터랙티브 요소
**When** Tab 키를 누르면
**Then** 논리적 순서로 포커스가 이동하고 Electric Blue 아웃라인이 표시된다

**Given** 모달이 열렸을 때
**When** Escape 키를 누르면
**Then** 모달이 닫히고 포커스가 트랩된다

---

# Story Summary

| Epic | 스토리 수 | FR 커버리지 |
|------|----------|------------|
| Epic 1: 프로젝트 기반 및 기본 인증 | 4 | FR1, FR2, FR4, FR9, FR10, FR11 |
| Epic 2: 세션 관리 및 프로필 | 6 | FR3, FR5, FR12, FR13, FR14, FR15 |
| Epic 3: 계정 복구 | 3 | FR6, FR7, FR8 |
| Epic 4: 접근 제어 시스템 | 6 | FR32-FR41 |
| Epic 5: ADMIN 유저 관리 | 8 | FR16-FR31 |
| Epic 6: 감사 및 히스토리 | 5 | FR42-FR47, FR62, FR63 |
| Epic 7: 사용자 인터페이스 및 경험 | 9 | FR48-FR61 |
| **합계** | **41** | **63 FRs** |
