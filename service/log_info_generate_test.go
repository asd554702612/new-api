package service

import (
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/stretchr/testify/require"
)

func TestAppendBillingInfoRecordsDefaultBillingPreference(t *testing.T) {
	other := map[string]interface{}{}
	relayInfo := &relaycommon.RelayInfo{
		BillingSource: BillingSourceWallet,
	}

	appendBillingInfo(relayInfo, other)

	require.Equal(t, BillingSourceWallet, other["billing_source"])
	require.Equal(t, "subscription_first", other["billing_preference"])
}
