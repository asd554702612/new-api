package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCasdoorPaymentSignatureUsesTimestampNonceAndRawBody(t *testing.T) {
	body := []byte(`{"externalOrderId":"gptk-order-202606250001","amount":88.66}`)

	signature := BuildCasdoorPaymentSignature("client-secret", "1782362400", "nonce-123", body)

	require.Equal(t, "10c1dfde56fb4167befbee595cc1ee88aecf1ab35fa6b25132824a185578aa2b", signature)
}

func TestCasdoorWebhookSignatureVerificationUsesRawBody(t *testing.T) {
	body := []byte(`{"event":"payment.paid","externalOrderId":"casdoor-topup-1","amount":12.34}`)
	header := BuildCasdoorWebhookSignature("client-secret", body)

	require.True(t, VerifyCasdoorWebhookSignature("client-secret", body, header))
	require.False(t, VerifyCasdoorWebhookSignature("client-secret", []byte(`{"event":"payment.paid","externalOrderId":"casdoor-topup-1","amount":12.35}`), header))
	require.False(t, VerifyCasdoorWebhookSignature("wrong-secret", body, header))
	require.False(t, VerifyCasdoorWebhookSignature("client-secret", body, "sha256=bad"))
}
