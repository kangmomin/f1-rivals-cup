# 🎨 Rivals Cup Design System

**Identity:** "Dark Mode First, High Contrast, Data-Driven"

## 1. 디자인 컨셉 (Design Concept)

* **Core Keyword:** Carbon & Neon
* **Mood:**
* **Immersive:** 실제 F1 스티어링 휠이나 텔레메트리 모니터를 보는 듯한 몰입감.
* **Professional:** 장난스러운 커뮤니티가 아닌, 진지한 감독들의 매니지먼트 툴.
* **Dynamic:** 정적인 자금 장부에서도 느껴지는 스포츠의 역동성과 강렬한 대비.



---

## 2. 컬러 팔레트 (Color Palette)

> 눈의 피로를 줄이고 데이터 가독성을 높이기 위해 **다크 모드**를 기본으로 합니다.

### 2.1. Brand Colors (Primary & Secondary)

* **Racing Red (Main Brand)**
* `Hex: #FF3B30`
* 용도: 로고, 핵심 CTA 버튼, Live 상태, 중요 알림.
* 의미: 속도, 위험, 열정.


* **Electric Blue (Secondary)**
* `Hex: #0A84FF`
* 용도: 링크, 활성화된 탭, 진행률 바, 기술적 요소.
* 의미: 기술(Telemetry), 지성, 신뢰.



### 2.2. Background Colors

* **Carbon Black (App Background)**
* `Hex: #121212` (아주 짙은 회색)
* 용도: 전체 페이지 배경.


* **Asphalt Grey (Card/Surface)**
* `Hex: #1C1C1E`
* 용도: 컨텐츠 카드, 사이드바, 헤더 배경 (미세한 보더로 구분).


* **Steel Grey (Border/Divider)**
* `Hex: #3A3A3C`
* 용도: 카드 테두리, 구분선.



### 2.3. Semantic Colors (상태 및 자금)

* **Profit Green (입금/이익/성공)**
* `Hex: #30D158` (형광 느낌)
* 용도: 상금 입금(+), 순위 상승, 계약 성사, 정상 수치.


* **Loss Red (출금/손실/위험)**
* `Hex: #FF453A` (가독성 조절됨)
* 용도: 벌금/지출(-), 순위 하락, 리타이어(DNF), 오류.


* **Warning Orange (경고/대기)**
* `Hex: #FF9F0A`
* 용도: 심사 중, 페널티 경고, 계약 만료 임박.



---

## 3. 타이포그래피 (Typography)

> F1의 미래지향적 느낌과 데이터의 정확성을 표현 (Google Fonts 기준).

### 3.1. Headings (제목)

* **Font:** Saira 또는 Rajdhani
* **Style:** 스퀘어 형태의 산세리프. 기술적이고 단단한 느낌.
* **Weight:** Bold (700), SemiBold (600)
* **Usage:** 페이지 제목, 카드 타이틀, 등수 표시.

### 3.2. Body (본문)

* **Font:** Inter 또는 Pretendard (한글 최적화)
* **Style:** 가독성이 뛰어난 산세리프.
* **Weight:** Regular (400), Medium (500)
* **Usage:** 뉴스 본문, 일반 텍스트, 설명.

### 3.3. Data & Numbers (데이터)

* **Font:** JetBrains Mono 또는 Roboto Mono
* **Style:** 고정폭(Monospace). 숫자의 자릿수가 딱 맞아야 함.
* **Usage:** 자금(Credit), 랩 타임, 텔레메트리, 날짜.
* **Tip:** 1,000과 10,000의 자릿수 비교가 시각적으로 명확해야 함.

---

## 4. UI 컴포넌트 가이드 (Component Guidelines)

### 4.1. 카드 (Cards) - "The Bento Grid"

* **Style:** Asphalt Grey 배경 + Steel Grey 1px 테두리.
* **Radius:** 8px ~ 12px (약간 각진 느낌 유지).
* **Shadow:** 그림자 대신 Surface Color 차이로 깊이감 표현.
* **Hover:** 마우스 오버 시 테두리를 밝은 회색(`#505050`)으로 강조.

### 4.2. 버튼 (Buttons)

* **Primary:** Racing Red 배경 + 흰색 텍스트 (이적 제안, 뉴스 발행 등).
* **Ghost:** 배경 없음 + Electric Blue 텍스트/테두리 (상세 보기, 취소).
* **Danger:** 투명 배경 + Loss Red 테두리/텍스트 (계약 파기, 삭제).

### 4.3. 데이터 시각화 (Data Viz)

* **Charts:** 어두운 배경 위 네온 계열(Neon Green, Cyan, Magenta) 라인/바 차트 사용.
* **Tables (장부):**
* 헤더는 옅은 회색.
* 행(Row) 간격은 여유 있게.
* 금액 색상: 플러스(+)는 초록색, 마이너스(-)는 빨간색.



### 4.4. 텔레메트리 뷰 (Telemetry View)

* **Track Map:** 짙은 배경 + 밝은 회색 라인.
* **Driver Dot:** 실제 F1 팀 컬러를 적용한 원형 점.

---

## 5. 아이콘 시스템 (Iconography)

* **Library:** Phosphor Icons 또는 Lucide React 권장.
* **Style:** Line 아이콘 (채워진 아이콘보다 세련된 느낌).
* **Weight:** Bold 또는 Regular.

---

## 6. 예시 화면 (Mockup Description)

### 메인 대시보드

* **배경:** Carbon Black.
* **상단:** 다음 경기 D-Day 카운터 (Saira 폰트, 중앙 정렬).
* **중앙:** 라이브 중계(YouTube) 16:9 배치 + Red Shadow 글로우 효과.
* **하단:** 뉴스 헤드라인 가로 스크롤(Ticker).

### 자금 장부 (Ledger)

* **형태:** 테이블 리스트.
* **폰트:** 금액 컬럼에 JetBrains Mono 사용.
* **잔고 표시:** 우측 상단에 매우 큰 폰트로 표시 (예: ₩ 24,500 Cr).
* *특이사항:* 잔고가 마이너스일 경우 빨간색으로 점멸.