package zitadeladmin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type getUserResponse struct {
	User struct {
		Human *struct {
			Email *struct {
				Email      string `json:"email"`
				IsVerified bool   `json:"isVerified"`
			} `json:"email"`
		} `json:"human"`
	} `json:"user"`
}

type resendEmailCodeRequest struct {
	ReturnCode map[string]any `json:"returnCode"`
}

type resendEmailCodeResponse struct {
	VerificationCode string `json:"verificationCode"`
}

type verifyEmailRequest struct {
	VerificationCode string `json:"verificationCode"`
}

type apiErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type userEmailState struct {
	Email      string
	IsVerified bool
}

func NewClient(cfg config.Zitadel) (*Client, error) {
	token, err := cfg.AdminTokenValue()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(token) == "" {
		return nil, nil
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.IssuerURL), "/")
	if baseURL == "" {
		return nil, errors.New("auth zitadel issuer url is empty")
	}
	timeout := cfg.HTTPTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		baseURL:    baseURL,
		token:      strings.TrimSpace(token),
		httpClient: &http.Client{Timeout: timeout},
	}, nil
}

func (c *Client) ForceVerifyUserEmail(ctx context.Context, userID string) (*ports.ZitadelUserEmailVerificationResult, error) {
	if c == nil {
		return nil, errors.New("zitadel admin client is nil")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, &ports.ZitadelAdminAPIError{StatusCode: http.StatusBadRequest, Message: "ZITADEL user id is required"}
	}

	state, err := c.getUserEmailState(ctx, userID)
	if err != nil {
		return nil, err
	}
	if state.IsVerified {
		return &ports.ZitadelUserEmailVerificationResult{
			UserID:          userID,
			Email:           state.Email,
			AlreadyVerified: true,
		}, nil
	}

	verificationCode, err := c.resendEmailCode(ctx, userID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(verificationCode) == "" {
		return nil, &ports.ZitadelAdminAPIError{StatusCode: http.StatusBadGateway, Message: "ZITADEL did not return verification code"}
	}
	if err := c.verifyEmail(ctx, userID, verificationCode); err != nil {
		return nil, err
	}
	return &ports.ZitadelUserEmailVerificationResult{
		UserID:          userID,
		Email:           state.Email,
		AlreadyVerified: false,
	}, nil
}

func (c *Client) getUserEmailState(ctx context.Context, userID string) (*userEmailState, error) {
	var response getUserResponse
	if err := c.doJSON(ctx, http.MethodGet, "/v2/users/"+url.PathEscape(userID), nil, &response); err != nil {
		return nil, err
	}
	if response.User.Human == nil || response.User.Human.Email == nil {
		return nil, &ports.ZitadelAdminAPIError{StatusCode: http.StatusBadRequest, Message: "ZITADEL user does not have a human email"}
	}
	email := strings.TrimSpace(response.User.Human.Email.Email)
	if email == "" {
		return nil, &ports.ZitadelAdminAPIError{StatusCode: http.StatusBadRequest, Message: "ZITADEL user email is empty"}
	}
	return &userEmailState{Email: email, IsVerified: response.User.Human.Email.IsVerified}, nil
}

func (c *Client) resendEmailCode(ctx context.Context, userID string) (string, error) {
	body := resendEmailCodeRequest{ReturnCode: map[string]any{}}
	var response resendEmailCodeResponse
	if err := c.doJSON(ctx, http.MethodPost, "/v2/users/"+url.PathEscape(userID)+"/email/resend", body, &response); err != nil {
		return "", err
	}
	return strings.TrimSpace(response.VerificationCode), nil
}

func (c *Client) verifyEmail(ctx context.Context, userID, verificationCode string) error {
	body := verifyEmailRequest{VerificationCode: strings.TrimSpace(verificationCode)}
	return c.doJSON(ctx, http.MethodPost, "/v2/users/"+url.PathEscape(userID)+"/email/verify", body, nil)
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any) error {
	endpoint := c.baseURL + path
	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode zitadel admin request: %w", err)
		}
		payload = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, payload)
	if err != nil {
		return fmt.Errorf("build zitadel admin request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute zitadel admin request: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read zitadel admin response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr apiErrorResponse
		_ = json.Unmarshal(rawBody, &apiErr)
		message := strings.TrimSpace(apiErr.Message)
		if message == "" {
			message = strings.TrimSpace(string(rawBody))
		}
		return &ports.ZitadelAdminAPIError{
			StatusCode: resp.StatusCode,
			Code:       strings.TrimSpace(apiErr.Code),
			Message:    message,
		}
	}
	if out == nil || len(rawBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(rawBody, out); err != nil {
		return fmt.Errorf("decode zitadel admin response: %w", err)
	}
	return nil
}
