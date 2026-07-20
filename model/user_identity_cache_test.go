package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
)

func TestUserBaseCarriesIdentitySnapshot(t *testing.T) {
	originalRedisEnabled := common.RedisEnabled
	t.Cleanup(func() {
		common.RedisEnabled = originalRedisEnabled
	})
	common.RedisEnabled = false

	user := &User{
		Username:           fmt.Sprintf("identity_cache_%d", time.Now().UnixNano()),
		Role:               common.RoleCommonUser,
		Status:             common.UserStatusEnabled,
		Group:              "default",
		AffCode:            fmt.Sprintf("IC%d", time.Now().UnixNano()),
		IdentityVerified:   true,
		IdentityAgeChecked: true,
		IdentityOver16:     true,
	}
	require.NoError(t, DB.Create(user).Error)

	base := user.ToBaseUser()
	require.True(t, base.IdentityVerified)
	require.True(t, base.IdentityAgeChecked)
	require.True(t, base.IdentityOver16)
	require.True(t, base.IdentitySnapshotCached)
	require.True(t, base.HasVerifiedIdentity())

	cached, err := GetUserCache(user.Id)
	require.NoError(t, err)
	require.True(t, cached.IdentityVerified)
	require.True(t, cached.IdentityAgeChecked)
	require.True(t, cached.IdentityOver16)
	require.True(t, cached.IdentitySnapshotCached)
	require.True(t, cached.HasVerifiedIdentity())
}
