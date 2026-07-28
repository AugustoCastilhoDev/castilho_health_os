package memed_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/memed"
)

func testPrescriber() memed.Prescriber {
	birth, _ := time.Parse("2006-01-02", "1985-03-10")
	return memed.Prescriber{
		ExternalID:  "user-123",
		Name:        "Joana",
		Surname:     "Souza",
		CPF:         "12345678901",
		BoardCode:   "CRM",
		BoardNumber: "12345",
		BoardState:  "SP",
		BirthDate:   birth,
		Email:       "joana@example.com",
	}
}

// Regression guard for a bug only caught by probing the real sandbox:
// Memed rejects a flat registration payload with a 400/"não contém a
// propriedade \"data\"" error — the body must be wrapped in a JSON:API-style
// {"data": {"type": "usuarios", "attributes": {...}}} envelope.
func TestClient_FetchOrRegisterToken_RegistersWithDataEnvelope(t *testing.T) {
	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPost:
			require.NoError(t, json.NewDecoder(r.Body).Decode(&capturedBody))
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"type":       "usuarios",
					"attributes": map[string]any{"token": "new-token"},
				},
			})
		}
	}))
	defer server.Close()

	client := memed.NewClient(memed.Config{APIKey: "k", SecretKey: "s", BaseURL: server.URL})
	token, err := client.FetchOrRegisterToken(context.Background(), testPrescriber())
	require.NoError(t, err)
	assert.Equal(t, "new-token", token)

	data, ok := capturedBody["data"].(map[string]any)
	require.True(t, ok, "registration body must be wrapped in a top-level \"data\" object")
	assert.Equal(t, "usuarios", data["type"])
	attributes, ok := data["attributes"].(map[string]any)
	require.True(t, ok, "registration body must carry fields under data.attributes")
	assert.Equal(t, "user-123", attributes["external_id"])
	assert.Equal(t, "Joana", attributes["nome"])
	assert.Equal(t, "Souza", attributes["sobrenome"])
	assert.Equal(t, "12345678901", attributes["cpf"])
	assert.Equal(t, "CRM", attributes["board_code"])
	assert.Equal(t, "10/03/1985", attributes["data_nascimento"])
}

func TestClient_FetchOrRegisterToken_ReturnsExistingTokenWithoutRegistering(t *testing.T) {
	registerCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"attributes": map[string]any{"token": "existing-token"}},
			})
		case http.MethodPost:
			registerCalled = true
			w.WriteHeader(http.StatusCreated)
		}
	}))
	defer server.Close()

	client := memed.NewClient(memed.Config{APIKey: "k", SecretKey: "s", BaseURL: server.URL})
	token, err := client.FetchOrRegisterToken(context.Background(), testPrescriber())
	require.NoError(t, err)
	assert.Equal(t, "existing-token", token)
	assert.False(t, registerCalled, "an existing prescriber must not be re-registered")
}

func TestClient_FetchOrRegisterToken_PropagatesNonNotFoundLookupErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"errors":[{"detail":"invalid api key"}]}`))
		}
	}))
	defer server.Close()

	client := memed.NewClient(memed.Config{APIKey: "bad", SecretKey: "bad", BaseURL: server.URL})
	_, err := client.FetchOrRegisterToken(context.Background(), testPrescriber())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}
