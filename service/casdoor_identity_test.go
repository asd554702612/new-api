package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
)

func casdoorIdentitySignature(secret string, timestamp string, nonce string, rawBody string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "\n" + nonce + "\n" + rawBody))
	return hex.EncodeToString(mac.Sum(nil))
}

func casdoorIdentityURLSignature(secret string, timestamp string, nonce string, clientID string, userID string, redirectURI string, state string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "\n" + nonce + "\n" + clientID + "\n" + userID + "\n" + redirectURI + "\n" + state))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestCasdoorIdentityClientSyncSignsRawBody(t *testing.T) {
	const clientID = "app-token-gepinkeji-dd884f73a91f"
	const clientSecret = "secret-for-test"
	const expectedBody = `{"userId":"casdoor-sub"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/external/user/sync", r.URL.Path)
		require.Equal(t, clientID, r.Header.Get("X-Casdoor-App-Id"))
		require.NotEmpty(t, r.Header.Get("X-Casdoor-Timestamp"))
		require.NotEmpty(t, r.Header.Get("X-Casdoor-Nonce"))
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Equal(t, expectedBody, string(body))

		expectedSignature := casdoorIdentitySignature(
			clientSecret,
			r.Header.Get("X-Casdoor-Timestamp"),
			r.Header.Get("X-Casdoor-Nonce"),
			string(body),
		)
		require.Equal(t, expectedSignature, r.Header.Get("X-Casdoor-Signature"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","msg":"","data":{"userId":"casdoor-sub","owner":"gepin","name":"alice","displayName":"Alice","email":"alice@example.com","phone":"13800000000","isVerified":true,"ageChecked":true,"isOver16":true}}`))
	}))
	defer server.Close()

	client := NewCasdoorIdentityClient(server.URL, clientID, clientSecret)
	identity, err := client.SyncUser(context.Background(), "casdoor-sub")

	require.NoError(t, err)
	require.Equal(t, "casdoor-sub", identity.UserID)
	require.Equal(t, "Alice", identity.DisplayName)
	require.True(t, CanEnterCasdoorIdentityBusiness(identity))
}

func TestCasdoorIdentityClientSyncFailsClosedOnProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"status":"error","msg":"invalid signature"}`))
	}))
	defer server.Close()

	client := NewCasdoorIdentityClient(server.URL, "client", "secret")
	identity, err := client.SyncUser(context.Background(), "casdoor-sub")

	require.ErrorContains(t, err, "invalid signature")
	require.Nil(t, identity)
}

func TestCasdoorIdentityClientBuildVerificationURLSignsQuery(t *testing.T) {
	const clientID = "app-token-gepinkeji-dd884f73a91f"
	const clientSecret = "secret-for-test"
	const userID = "casdoor-sub"
	const redirectURI = "https://token.gepinkeji.com/identity/callback"
	const state = "identity-state"

	client := NewCasdoorIdentityClient("https://login.gepinkeji.com", clientID, clientSecret)
	rawURL, err := client.BuildVerificationURL(userID, redirectURI, state)
	require.NoError(t, err)

	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)
	require.Equal(t, "https", parsed.Scheme)
	require.Equal(t, "login.gepinkeji.com", parsed.Host)
	require.Equal(t, "/identity-verification/submit", parsed.Path)

	query := parsed.Query()
	require.Equal(t, clientID, query.Get("clientId"))
	require.Equal(t, userID, query.Get("userId"))
	require.Equal(t, redirectURI, query.Get("redirectUri"))
	require.Equal(t, state, query.Get("state"))
	require.NotEmpty(t, query.Get("timestamp"))
	require.NotEmpty(t, query.Get("nonce"))

	expectedSignature := casdoorIdentityURLSignature(
		clientSecret,
		query.Get("timestamp"),
		query.Get("nonce"),
		clientID,
		userID,
		redirectURI,
		state,
	)
	require.Equal(t, expectedSignature, query.Get("signature"))
}

func TestCasdoorIdentityResponseUsesCommonJSON(t *testing.T) {
	var response casdoorIdentitySyncResponse
	err := common.Unmarshal([]byte(`{"status":"ok","data":{"userId":"u"}}`), &response)
	require.NoError(t, err)
	require.Equal(t, "u", response.Data.UserID)
}
