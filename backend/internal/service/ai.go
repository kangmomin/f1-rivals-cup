package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AIService handles AI-related operations using Gemini API
type AIService struct {
	apiKey     string
	httpClient *http.Client
}

// NewAIService creates a new AIService instance
func NewAIService(apiKey string) *AIService {
	return &AIService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GeminiRequest represents the request body for Gemini API
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent represents content in Gemini API request
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of content in Gemini API request
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse represents the response from Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

// NewsContent represents the structured news content from AI
type NewsContent struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	NewsProvider string `json:"news_provider"`
}

// ErrNoAPIKey is returned when API key is not configured
var ErrNoAPIKey = errors.New("Gemini API key is not configured")

// ErrNoContent is returned when no content is generated
var ErrNoContent = errors.New("no content generated from AI")

// ErrInvalidJSON is returned when AI response is not valid JSON
var ErrInvalidJSON = errors.New("AI response is not valid JSON")

// ErrAPIUnavailable is returned when Gemini API returns 5xx error
var ErrAPIUnavailable = errors.New("AI service temporarily unavailable")

// ErrAPIRateLimit is returned when Gemini API returns 429 error
var ErrAPIRateLimit = errors.New("AI service rate limit exceeded")

// ErrAPIBadRequest is returned when Gemini API returns 4xx error
var ErrAPIBadRequest = errors.New("invalid request to AI service")

// maxResponseSize limits the response body to 1MB to prevent memory exhaustion
const maxResponseSize = 1 * 1024 * 1024

// GenerateNewsContent generates news content from user input using Gemini API
func (s *AIService) GenerateNewsContent(ctx context.Context, userInput string) (*NewsContent, error) {
	if s.apiKey == "" {
		return nil, ErrNoAPIKey
	}

	// Gemini API endpoint
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite:generateContent?key=%s", s.apiKey)

	prompt := fmt.Sprintf(`당신은 F1 Rivals Cup의 전문 뉴스 기자입니다. 아래 정보를 바탕으로 뉴스 기사를 작성해주세요.

# 응답 형식

반드시 아래 JSON 형식으로만 응답하세요. 다른 텍스트는 포함하지 마세요.

{
  "title": "말머리와 제목 (예: [속보] 홍길동, 레드불로 이적 확정)",
  "description": "Markdown 형식의 본문 내용",
  "news_provider": "언론사 이름"
}

# 기사 작성 규칙

1. **title (제목):**
   - 말머리와 함께 기사의 핵심을 담은 제목
   - 말머리 예시: [속보], [오피셜], [이슈], [단독], [광고]
   - Markdown 형식 사용하지 않음 (순수 텍스트)

2. **description (본문):**
   - Markdown 형식 사용 (## 부제, **강조**, - 목록 등)
   - 부제(##)로 시작하여 핵심 요약
   - 2~3개의 문단으로 구성
   - 핵심 인물, 금액, 중요한 발언은 **볼드체**로 강조

3. **news_provider (언론사):**
   - 사용자가 지정하면 해당 이름 사용
   - 지정하지 않으면 창의적인 언론사 이름 생성 (예: F1RC 데일리, 라이벌스 타임즈)

# 입력된 정보

%s

# JSON 응답:`, userInput)

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		// Check for context cancellation/timeout
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-2xx HTTP status codes
	if resp.StatusCode != http.StatusOK {
		switch {
		case resp.StatusCode == http.StatusTooManyRequests:
			return nil, ErrAPIRateLimit
		case resp.StatusCode >= 500:
			return nil, fmt.Errorf("%w: status %d", ErrAPIUnavailable, resp.StatusCode)
		case resp.StatusCode >= 400:
			return nil, fmt.Errorf("%w: status %d", ErrAPIBadRequest, resp.StatusCode)
		}
	}

	// Limit response size to prevent memory exhaustion
	limitedReader := io.LimitReader(resp.Body, maxResponseSize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error in response body
	if geminiResp.Error != nil {
		return nil, fmt.Errorf("Gemini API error: %s (code: %d)", geminiResp.Error.Message, geminiResp.Error.Code)
	}

	// Extract generated text
	if len(geminiResp.Candidates) == 0 ||
		len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, ErrNoContent
	}

	generatedText := geminiResp.Candidates[0].Content.Parts[0].Text

	// Parse JSON from generated text
	// Remove markdown code block if present
	generatedText = strings.TrimSpace(generatedText)
	if strings.HasPrefix(generatedText, "```json") {
		generatedText = strings.TrimPrefix(generatedText, "```json")
		generatedText = strings.TrimSuffix(generatedText, "```")
		generatedText = strings.TrimSpace(generatedText)
	} else if strings.HasPrefix(generatedText, "```") {
		generatedText = strings.TrimPrefix(generatedText, "```")
		generatedText = strings.TrimSuffix(generatedText, "```")
		generatedText = strings.TrimSpace(generatedText)
	}

	var newsContent NewsContent
	if err := json.Unmarshal([]byte(generatedText), &newsContent); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	return &newsContent, nil
}

// IsConfigured returns whether the AI service is properly configured
func (s *AIService) IsConfigured() bool {
	return s.apiKey != ""
}
