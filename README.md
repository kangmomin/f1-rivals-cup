# F1 Rivals Cup

리그 관리 시스템 - 인증 및 권한 관리 MVP

## 기술 스택

### Backend
- Go 1.21+
- Echo v4 (웹 프레임워크)
- PostgreSQL 14+ (JSONB 권한 관리)
- sqlc (타입 안전 SQL)
- JWT 인증

### Frontend
- React 18
- TypeScript (strict mode)
- Vite
- Tailwind CSS (Carbon & Neon 테마)
- React Hook Form + Zod

## 시작하기

### 사전 요구사항

- Go 1.21+
- Node.js 20+
- Docker & Docker Compose
- Air (Go hot reload): `go install github.com/air-verse/air@latest`

### 설치

1. 저장소 클론
```bash
git clone https://github.com/your-org/f1-rivals-cup.git
cd f1-rivals-cup
```

2. 환경변수 설정
```bash
cp .env.example .env
```

3. Docker 컨테이너 시작
```bash
make docker-up
```

4. 개발 서버 시작
```bash
make dev
```

### 접속 URL

- **Backend API**: http://localhost:8080
- **Frontend**: http://localhost:5173
- **MailHog (이메일 테스트)**: http://localhost:8025
- **PostgreSQL**: localhost:5432

## 개발 명령어

```bash
# 개발 서버 시작
make dev

# Docker 컨테이너 관리
make docker-up      # 시작
make docker-down    # 중지
make docker-logs    # 로그 확인

# 데이터베이스 마이그레이션
make migrate-up     # 마이그레이션 실행
make migrate-down   # 롤백

# 코드 생성
make generate       # sqlc 코드 생성

# 테스트
make test           # 전체 테스트
make test-backend   # 백엔드만
make test-frontend  # 프론트엔드만
```

## 프로젝트 구조

```
f1-rivals-cup/
├── backend/
│   ├── cmd/server/         # 애플리케이션 진입점
│   ├── internal/
│   │   ├── config/         # 환경 설정
│   │   ├── domain/         # 비즈니스 엔티티
│   │   ├── handler/        # HTTP 핸들러
│   │   ├── service/        # 비즈니스 로직
│   │   ├── repository/     # 데이터 접근 (sqlc)
│   │   ├── middleware/     # 미들웨어
│   │   └── errors/         # 에러 코드
│   └── db/
│       ├── migrations/     # SQL 마이그레이션
│       └── queries/        # sqlc 쿼리
├── frontend/
│   └── src/
│       ├── components/     # React 컴포넌트
│       ├── contexts/       # Context API
│       ├── hooks/          # 커스텀 훅
│       ├── pages/          # 페이지 컴포넌트
│       ├── services/       # API 클라이언트
│       └── types/          # TypeScript 타입
├── docker-compose.yml
├── Makefile
└── README.md
```

## API 엔드포인트

### 인증
- `POST /api/v1/auth/register` - 회원가입
- `POST /api/v1/auth/login` - 로그인
- `POST /api/v1/auth/refresh` - 토큰 갱신
- `POST /api/v1/auth/logout` - 로그아웃

### 프로필
- `GET /api/v1/profile` - 내 프로필 조회
- `PUT /api/v1/profile` - 프로필 수정
- `PUT /api/v1/profile/password` - 비밀번호 변경
- `DELETE /api/v1/profile` - 계정 탈퇴

### 유저 관리 (ADMIN)
- `GET /api/v1/members` - 유저 목록
- `GET /api/v1/members/:id` - 유저 상세
- `PUT /api/v1/members/:id/role` - 역할 변경
- `PUT /api/v1/members/:id/permissions` - 권한 변경
- `GET /api/v1/members/:id/history` - 권한 변경 이력

## 라이선스

MIT License
