package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
	"github.com/stretchr/testify/require"
)

func TestUpdateOptionMapSetsCasdoorIdentityAPIRequirement(t *testing.T) {
	originalRequired := setting.CasdoorIdentityApiRequired
	common.OptionMapRWMutex.Lock()
	originalMap := common.OptionMap
	common.OptionMap = map[string]string{}
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		setting.CasdoorIdentityApiRequired = originalRequired
		common.OptionMapRWMutex.Lock()
		common.OptionMap = originalMap
		common.OptionMapRWMutex.Unlock()
	})

	require.NoError(t, updateOptionMap("CasdoorIdentityApiRequired", "true"))
	require.True(t, setting.CasdoorIdentityApiRequired)

	require.NoError(t, updateOptionMap("CasdoorIdentityApiRequired", "false"))
	require.False(t, setting.CasdoorIdentityApiRequired)
}
