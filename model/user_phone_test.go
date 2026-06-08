package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
)

func TestCheckUserExistOrDeletedDetectsPhoneNumber(t *testing.T) {
	truncateTables(t)

	require.NoError(t, DB.Create(&User{
		Username:    "alice",
		Password:    "hashed",
		Status:      common.UserStatusEnabled,
		PhoneNumber: "+8613800138000",
	}).Error)

	exists, err := CheckUserExistOrDeleted("bob", "", "+8613800138000")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestValidateAndFillAllowsPhoneNumberPasswordLogin(t *testing.T) {
	truncateTables(t)

	hashed, err := common.Password2Hash("password123")
	require.NoError(t, err)
	require.NoError(t, DB.Create(&User{
		Username:    "alice",
		Password:    hashed,
		Status:      common.UserStatusEnabled,
		PhoneNumber: "+8613800138000",
	}).Error)

	user := User{
		Username: "13800138000",
		Password: "password123",
	}
	require.NoError(t, user.ValidateAndFill())
	require.Equal(t, "alice", user.Username)
}

func TestSearchUsersMatchesPhoneNumber(t *testing.T) {
	truncateTables(t)

	require.NoError(t, DB.Create(&User{
		Username:    "alice",
		Password:    "hashed",
		Status:      common.UserStatusEnabled,
		PhoneNumber: "+8613800138000",
	}).Error)

	users, total, err := SearchUsers("13800138000", "", nil, nil, 0, 10)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, users, 1)
	require.Equal(t, "alice", users[0].Username)
}
