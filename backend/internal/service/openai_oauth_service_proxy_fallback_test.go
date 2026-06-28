package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type openAIOAuthClientProxyFallbackStub struct {
	lastProxyURL string
}

func (s *openAIOAuthClientProxyFallbackStub) ExchangeCode(ctx context.Context, code, codeVerifier, redirectURI, proxyURL, clientID string) (*openai.TokenResponse, error) {
	s.lastProxyURL = proxyURL
	return &openai.TokenResponse{
		AccessToken:  "at",
		RefreshToken: "rt",
		ExpiresIn:    3600,
	}, nil
}

func (s *openAIOAuthClientProxyFallbackStub) RefreshToken(ctx context.Context, refreshToken, proxyURL string) (*openai.TokenResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *openAIOAuthClientProxyFallbackStub) RefreshTokenWithClientID(ctx context.Context, refreshToken, proxyURL string, clientID string) (*openai.TokenResponse, error) {
	return nil, errors.New("not implemented")
}

type openAIProxyRepoFallbackStub struct {
	listActiveFunc func(ctx context.Context) ([]Proxy, error)
}

func (m *openAIProxyRepoFallbackStub) Create(ctx context.Context, proxy *Proxy) error {
	panic("not implemented")
}

func (m *openAIProxyRepoFallbackStub) GetByID(ctx context.Context, id int64) (*Proxy, error) {
	return nil, errors.New("proxy not found")
}

func (m *openAIProxyRepoFallbackStub) ListByIDs(ctx context.Context, ids []int64) ([]Proxy, error) {
	panic("not implemented")
}

func (m *openAIProxyRepoFallbackStub) Update(ctx context.Context, proxy *Proxy) error {
	panic("not implemented")
}

func (m *openAIProxyRepoFallbackStub) Delete(ctx context.Context, id int64) error {
	panic("not implemented")
}

func (m *openAIProxyRepoFallbackStub) List(ctx context.Context, params pagination.PaginationParams) ([]Proxy, *pagination.PaginationResult, error) {
	panic("not implemented")
}

func (m *openAIProxyRepoFallbackStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, protocol, status, search string) ([]Proxy, *pagination.PaginationResult, error) {
	panic("not implemented")
}

func (m *openAIProxyRepoFallbackStub) ListWithFiltersAndAccountCount(ctx context.Context, params pagination.PaginationParams, protocol, status, search string) ([]ProxyWithAccountCount, *pagination.PaginationResult, error) {
	panic("not implemented")
}

func (m *openAIProxyRepoFallbackStub) ListActive(ctx context.Context) ([]Proxy, error) {
	if m.listActiveFunc != nil {
		return m.listActiveFunc(ctx)
	}
	return nil, nil
}

func (m *openAIProxyRepoFallbackStub) ListActiveWithAccountCount(ctx context.Context) ([]ProxyWithAccountCount, error) {
	panic("not implemented")
}

func (m *openAIProxyRepoFallbackStub) ExistsByHostPortAuth(ctx context.Context, host string, port int, username, password string) (bool, error) {
	panic("not implemented")
}

func (m *openAIProxyRepoFallbackStub) CountAccountsByProxyID(ctx context.Context, proxyID int64) (int64, error) {
	panic("not implemented")
}

func (m *openAIProxyRepoFallbackStub) ListAccountSummariesByProxyID(ctx context.Context, proxyID int64) ([]ProxyAccountSummary, error) {
	panic("not implemented")
}

func newOpenAIProxyRepoFallbackStub() *openAIProxyRepoFallbackStub {
	return &openAIProxyRepoFallbackStub{
		listActiveFunc: func(ctx context.Context) ([]Proxy, error) {
			return []Proxy{
				{
					ID:       1,
					Protocol: "http",
					Host:     "proxy.test",
					Port:     3128,
					Status:   StatusActive,
				},
			}, nil
		},
	}
}

func TestOpenAIOAuthService_GenerateAuthURL_UsesActiveProxyFallback(t *testing.T) {
	client := &openAIOAuthClientProxyFallbackStub{}
	svc := NewOpenAIOAuthService(newOpenAIProxyRepoFallbackStub(), client)
	defer svc.Stop()

	result, err := svc.GenerateAuthURL(context.Background(), nil, "", PlatformOpenAI)
	require.NoError(t, err)

	session, ok := svc.sessionStore.Get(result.SessionID)
	require.True(t, ok)
	require.Equal(t, "http://proxy.test:3128", session.ProxyURL)
}

func TestOpenAIOAuthService_ExchangeCode_UsesActiveProxyFallback(t *testing.T) {
	client := &openAIOAuthClientProxyFallbackStub{}
	svc := NewOpenAIOAuthService(newOpenAIProxyRepoFallbackStub(), client)
	defer svc.Stop()
	svc.sessionStore.Set("sid", &openai.OAuthSession{
		State:        "expected-state",
		CodeVerifier: "verifier",
		RedirectURI:  openai.DefaultRedirectURI,
		CreatedAt:    time.Now(),
	})

	info, err := svc.ExchangeCode(context.Background(), &OpenAIExchangeCodeInput{
		SessionID: "sid",
		Code:      "auth-code",
		State:     "expected-state",
	})
	require.NoError(t, err)
	require.Equal(t, "at", info.AccessToken)
	require.Equal(t, "http://proxy.test:3128", client.lastProxyURL)
}
