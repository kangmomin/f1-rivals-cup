package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

// ErrNoAPIKey is returned when API key is not configured
var ErrNoAPIKey = errors.New("Gemini API key is not configured")

// ErrNoContent is returned when no content is generated
var ErrNoContent = errors.New("no content generated from AI")

// ErrAPIUnavailable is returned when Gemini API returns 5xx error
var ErrAPIUnavailable = errors.New("AI service temporarily unavailable")

// ErrAPIRateLimit is returned when Gemini API returns 429 error
var ErrAPIRateLimit = errors.New("AI service rate limit exceeded")

// ErrAPIBadRequest is returned when Gemini API returns 4xx error
var ErrAPIBadRequest = errors.New("invalid request to AI service")

// maxResponseSize limits the response body to 1MB to prevent memory exhaustion
const maxResponseSize = 1 * 1024 * 1024

// GenerateNewsContent generates news content from user input using Gemini API
func (s *AIService) GenerateNewsContent(ctx context.Context, userInput string) (string, error) {
	if s.apiKey == "" {
		return "", ErrNoAPIKey
	}

	// Gemini API endpoint
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s", s.apiKey)

	prompt := fmt.Sprintf(`당신은 F1 e스포츠 리그의 뉴스 기자입니다. 아래 정보를 바탕으로 Markdown 형식의 뉴스 기사를 작성해주세요.

요구사항:
- 제목은 작성하지 마세요 (별도로 입력됨)
- Markdown 형식 사용 (## 소제목, **강조**, - 목록 등)
- 전문적이고 흥미로운 문체 사용
- 한국어로 작성
- 500-1000자 정도의 분량

입력된 정보:
%s

뉴스 본문:`, userInput)

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
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		// Check for context cancellation/timeout
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-2xx HTTP status codes
	if resp.StatusCode != http.StatusOK {
		switch {
		case resp.StatusCode == http.StatusTooManyRequests:
			return "", ErrAPIRateLimit
		case resp.StatusCode >= 500:
			return "", fmt.Errorf("%w: status %d", ErrAPIUnavailable, resp.StatusCode)
		case resp.StatusCode >= 400:
			return "", fmt.Errorf("%w: status %d", ErrAPIBadRequest, resp.StatusCode)
		}
	}

	// Limit response size to prevent memory exhaustion
	limitedReader := io.LimitReader(resp.Body, maxResponseSize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error in response body
	if geminiResp.Error != nil {
		return "", fmt.Errorf("Gemini API error: %s (code: %d)", geminiResp.Error.Message, geminiResp.Error.Code)
	}

	// Extract generated text
	if len(geminiResp.Candidates) == 0 ||
		len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", ErrNoContent
	}

	generatedText := geminiResp.Candidates[0].Content.Parts[0].Text

	return generatedText, nil
}

// IsConfigured returns whether the AI service is properly configured
func (s *AIService) IsConfigured() bool {
	return s.apiKey != ""
}
