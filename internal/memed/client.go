// Package memed talks to Memed's Sinapse Prescrição REST API — the backend
// side of the digital-prescription integration. This package only ever
// registers/looks up the prescriber (health professional) and hands back
// the per-professional token the frontend needs to load Memed's own widget;
// it never sees the prescription's content itself (medications, dosage,
// etc.) — that stays entirely between Memed's script and Memed's servers,
// reaching this app only as an external ID logged for audit purposes (see
// internal/domain/models/memed_prescription_log.go).
package memed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Config is exactly what Memed's dashboard gives you for an environment
// (sandbox or production): an API key/secret pair and the REST base URL.
// Memed publishes a shared sandbox key pair in its own public docs
// (doc.memed.com.br) for developers to integrate against before a
// commercial/production key pair is issued — see ROADMAP.md.
type Config struct {
	APIKey    string
	SecretKey string
	// BaseURL defaults to Memed's public sandbox
	// ("https://integrations.api.memed.com.br/v1") when empty.
	BaseURL string
}

const defaultBaseURL = "https://integrations.api.memed.com.br/v1"

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	secretKey  string
}

func NewClient(cfg Config) *Client {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		baseURL:    baseURL,
		apiKey:     cfg.APIKey,
		secretKey:  cfg.SecretKey,
	}
}

// Prescriber is the health professional data Memed's registration endpoint
// requires. ExternalID should be stable and unique (this app uses the
// User's own UUID) — it's what makes FetchOrRegisterToken idempotent
// instead of creating a duplicate Memed user on every call.
type Prescriber struct {
	ExternalID  string
	Name        string
	Surname     string
	CPF         string // 11 digits, no punctuation
	BoardCode   string // "CRM" | "CRO" | ...
	BoardNumber string
	BoardState  string // 2-letter UF
	BirthDate   time.Time
	Email       string
	Phone       string
	Sex         string // "M" | "F", optional
}

type usuariosResponse struct {
	Data struct {
		Attributes struct {
			Token string `json:"token"`
		} `json:"attributes"`
	} `json:"data"`
}

type apiError struct {
	status int
	body   string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("memed: API returned status %d: %s", e.status, e.body)
}

// FetchOrRegisterToken looks up the prescriber by ExternalID first (a
// professional may already be registered from a previous call) and falls
// back to registering them if Memed doesn't know that ID yet. The token
// Memed returns is not long-lived — callers should fetch a fresh one for
// each session rather than caching it, per Memed's own docs.
func (c *Client) FetchOrRegisterToken(ctx context.Context, p Prescriber) (string, error) {
	token, err := c.lookupToken(ctx, p.ExternalID)
	if err == nil {
		return token, nil
	}
	var apiErr *apiError
	if !isNotFound(err, &apiErr) {
		return "", err
	}
	return c.registerToken(ctx, p)
}

func isNotFound(err error, target **apiError) bool {
	ae, ok := err.(*apiError)
	if !ok {
		return false
	}
	*target = ae
	return ae.status == http.StatusNotFound
}

func (c *Client) lookupToken(ctx context.Context, externalID string) (string, error) {
	u := fmt.Sprintf("%s/sinapse-prescricao/usuarios/%s?%s", c.baseURL, url.PathEscape(externalID), c.authQuery())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", fmt.Errorf("memed: building lookup request: %w", err)
	}
	return c.doTokenRequest(req)
}

// registerRequest mirrors the JSON:API-style envelope Memed's own response
// uses (see usuariosResponse) — a flat payload gets rejected with
// `{"errors":[{"detail":"O objeto enviado não contém a propriedade \"data\"."}]}`
// (found by probing the real sandbox, not documented in the fetched docs).
type registerRequest struct {
	Data struct {
		Type       string            `json:"type"`
		Attributes map[string]string `json:"attributes"`
	} `json:"data"`
}

func (c *Client) registerToken(ctx context.Context, p Prescriber) (string, error) {
	attributes := map[string]string{
		"external_id":     p.ExternalID,
		"nome":            p.Name,
		"sobrenome":       p.Surname,
		"cpf":             p.CPF,
		"board_code":      p.BoardCode,
		"board_number":    p.BoardNumber,
		"board_state":     p.BoardState,
		"data_nascimento": p.BirthDate.Format("02/01/2006"),
	}
	if p.Email != "" {
		attributes["email"] = p.Email
	}
	if p.Phone != "" {
		attributes["telefone"] = p.Phone
	}
	if p.Sex != "" {
		attributes["sexo"] = p.Sex
	}
	var payload registerRequest
	payload.Data.Type = "usuarios"
	payload.Data.Attributes = attributes
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("memed: encoding registration payload: %w", err)
	}

	u := fmt.Sprintf("%s/sinapse-prescricao/usuarios?%s", c.baseURL, c.authQuery())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("memed: building registration request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doTokenRequest(req)
}

func (c *Client) doTokenRequest(req *http.Request) (string, error) {
	req.Header.Set("Accept", "application/vnd.api+json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("memed: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("memed: reading response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &apiError{status: resp.StatusCode, body: string(respBody)}
	}

	var parsed usuariosResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("memed: decoding response: %w", err)
	}
	if parsed.Data.Attributes.Token == "" {
		return "", fmt.Errorf("memed: response did not include a token")
	}
	return parsed.Data.Attributes.Token, nil
}

// authQuery never touches the secret-key from anywhere but this
// backend-only client — see the SecretKey doc comment on Config.
func (c *Client) authQuery() string {
	v := url.Values{}
	v.Set("api-key", c.apiKey)
	v.Set("secret-key", c.secretKey)
	return v.Encode()
}
