package model

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPrivacyRequestModelTest(t *testing.T) {
	t.Helper()
	require.NoError(t, DB.AutoMigrate(&PrivacyRequest{}))
	require.NoError(t, DB.Exec("DELETE FROM privacy_requests").Error)
	t.Cleanup(func() {
		DB.Exec("DELETE FROM privacy_requests")
	})
}

func TestCreatePrivacyRequestCreatesPendingRequest(t *testing.T) {
	setupPrivacyRequestModelTest(t)

	request, err := CreatePrivacyRequest(PrivacyRequestInput{
		UserId:       101,
		Username:     "alice",
		ContactName:  "Alice",
		ContactEmail: "alice@example.com",
		RequestType:  PrivacyRequestTypeAccess,
		Content:      "Please provide a copy of my personal information.",
	})

	require.NoError(t, err)
	require.NotZero(t, request.Id)
	assert.Equal(t, PrivacyRequestStatusPending, request.Status)
	assert.NotZero(t, request.CreatedAt)
	assert.NotZero(t, request.UpdatedAt)
	assert.Zero(t, request.HandledAt)
}

func TestCreatePrivacyRequestRejectsInvalidType(t *testing.T) {
	setupPrivacyRequestModelTest(t)

	_, err := CreatePrivacyRequest(PrivacyRequestInput{
		UserId:      102,
		Username:    "alice",
		RequestType: "export",
		Content:     "Please export my data.",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "privacy request type")
}

func TestCreatePrivacyRequestRejectsEmptyContent(t *testing.T) {
	setupPrivacyRequestModelTest(t)

	_, err := CreatePrivacyRequest(PrivacyRequestInput{
		UserId:      103,
		Username:    "alice",
		RequestType: PrivacyRequestTypeCorrection,
		Content:     "   ",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "content")
}

func TestCreatePrivacyRequestRejectsContentTooLong(t *testing.T) {
	setupPrivacyRequestModelTest(t)

	_, err := CreatePrivacyRequest(PrivacyRequestInput{
		UserId:      104,
		Username:    "alice",
		RequestType: PrivacyRequestTypeAccess,
		Content:     strings.Repeat("x", privacyRequestContentMax+1),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "too long")
}

func TestUserPrivacyRequestReadIsScopedToOwner(t *testing.T) {
	setupPrivacyRequestModelTest(t)

	ownRequest, err := CreatePrivacyRequest(PrivacyRequestInput{
		UserId:      201,
		Username:    "alice",
		RequestType: PrivacyRequestTypeAccess,
		Content:     "Please provide my data.",
	})
	require.NoError(t, err)
	otherRequest, err := CreatePrivacyRequest(PrivacyRequestInput{
		UserId:      202,
		Username:    "bob",
		RequestType: PrivacyRequestTypeDeletion,
		Content:     "Please delete my account.",
	})
	require.NoError(t, err)

	found, err := GetUserPrivacyRequestByID(201, ownRequest.Id)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, ownRequest.Id, found.Id)

	notFound, err := GetUserPrivacyRequestByID(201, otherRequest.Id)
	require.NoError(t, err)
	assert.Nil(t, notFound)

	items, total, err := ListUserPrivacyRequests(201, &common.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	assert.Equal(t, ownRequest.Id, items[0].Id)
}

func TestAdminUpdatePrivacyRequestStatusSetsHandledAt(t *testing.T) {
	setupPrivacyRequestModelTest(t)

	request, err := CreatePrivacyRequest(PrivacyRequestInput{
		UserId:      301,
		Username:    "alice",
		RequestType: PrivacyRequestTypeDeletion,
		Content:     "Please delete my account.",
	})
	require.NoError(t, err)

	updated, err := UpdatePrivacyRequestByAdmin(request.Id, PrivacyRequestAdminUpdate{
		Status:                 PrivacyRequestStatusCompleted,
		AdminId:                9,
		AdminName:              "admin",
		AdminNote:              "account removed",
		ExecuteAccountDeletion: true,
	})

	require.NoError(t, err)
	assert.Equal(t, PrivacyRequestStatusCompleted, updated.Status)
	assert.Equal(t, 9, updated.AdminId)
	assert.Equal(t, "admin", updated.AdminName)
	assert.Equal(t, "account removed", updated.AdminNote)
	assert.True(t, updated.ExecuteAccountDeletion)
	assert.NotZero(t, updated.HandledAt)
}

func TestAdminUpdatePrivacyRequestRejectsInvalidStatus(t *testing.T) {
	setupPrivacyRequestModelTest(t)

	request, err := CreatePrivacyRequest(PrivacyRequestInput{
		UserId:      302,
		Username:    "alice",
		RequestType: PrivacyRequestTypeAccess,
		Content:     "Please provide my data.",
	})
	require.NoError(t, err)

	_, err = UpdatePrivacyRequestByAdmin(request.Id, PrivacyRequestAdminUpdate{
		Status:    "done",
		AdminId:   9,
		AdminName: "admin",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "status")
}

func TestCancelPrivacyRequestRequiresOwner(t *testing.T) {
	setupPrivacyRequestModelTest(t)

	request, err := CreatePrivacyRequest(PrivacyRequestInput{
		UserId:      401,
		Username:    "alice",
		RequestType: PrivacyRequestTypeAccess,
		Content:     "Please provide my data.",
	})
	require.NoError(t, err)

	_, err = CancelUserPrivacyRequest(request.Id, 402)
	require.Error(t, err)

	updated, err := CancelUserPrivacyRequest(request.Id, 401)
	require.NoError(t, err)
	assert.Equal(t, PrivacyRequestStatusCancelled, updated.Status)
	assert.NotZero(t, updated.HandledAt)
}
