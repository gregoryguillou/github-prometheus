package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

type Token struct {
	Type           string    `json:"type"`
	TokenType      string    `json:"tokenType"`
	Token          string    `json:"token"`
	InstallationID int64     `json:"installationId"`
	CreatedAt      time.Time `json:"createdAt"`
	ExpiresAt      time.Time `json:"expiresAt"`
}

type Tokens struct {
	*sync.Mutex
	Tokens map[string]Token
}

var tokens = Tokens{
	Mutex:  &sync.Mutex{},
	Tokens: map[string]Token{},
}

// get the current token
func token(ctx context.Context, endpoint string) (*Token, error) {
	tokens.Lock()
	defer tokens.Unlock()
	if endpoint == "" {
		endpoint = "http://localhost:3000/get-token"
	}
	if currentToken, ok := tokens.Tokens[endpoint]; ok && currentToken.ExpiresAt.Add(-60*time.Second).After(time.Now()) {
		return &currentToken, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	t := Token{}
	err = json.Unmarshal(payload, &t)
	if err == nil {
		tokens.Tokens[endpoint] = t
	}
	return &t, err
}
