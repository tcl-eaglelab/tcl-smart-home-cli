package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
)

// buildJWT constructs a minimal JWT with the given exp claim for testing.
func buildJWT(exp int64) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`))
	payload, _ := json.Marshal(map[string]interface{}{"exp": exp, "sub": 123})
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	return fmt.Sprintf("%s.%s.fakesig", header, encodedPayload)
}

func TestTokenExpiryMs(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantMs  int64
		wantErr bool
	}{
		{
			name:   "valid JWT with exp",
			token:  buildJWT(1777961527),
			wantMs: 1777961527 * 1000,
		},
		{
			name:    "not a JWT",
			token:   "notavalidtoken",
			wantErr: true,
		},
		{
			name:    "only two parts",
			token:   "header.payload",
			wantErr: true,
		},
		{
			name:    "invalid base64 payload",
			token:   "header.!!!.sig",
			wantErr: true,
		},
		{
			name: "missing exp claim",
			token: func() string {
				header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`))
				payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":123}`))
				return fmt.Sprintf("%s.%s.fakesig", header, payload)
			}(),
			wantErr: true,
		},
		{
			name: "exp is zero",
			token: func() string {
				header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`))
				payload := base64.RawURLEncoding.EncodeToString([]byte(`{"exp":0}`))
				return fmt.Sprintf("%s.%s.fakesig", header, payload)
			}(),
			wantErr: true,
		},
		{
			name:    "empty string",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TokenExpiryMs(tt.token)
			if tt.wantErr {
				if err == nil {
					t.Errorf("TokenExpiryMs() expected error, got nil (result: %d)", got)
				}
				return
			}
			if err != nil {
				t.Errorf("TokenExpiryMs() unexpected error: %v", err)
				return
			}
			if got != tt.wantMs {
				t.Errorf("TokenExpiryMs() = %d, want %d", got, tt.wantMs)
			}
		})
	}
}
