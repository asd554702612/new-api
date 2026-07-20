package model

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPublicFeedbackModelTest(t *testing.T) {
	t.Helper()
	require.NoError(t, DB.AutoMigrate(&PublicFeedback{}))
	require.NoError(t, DB.Exec("DELETE FROM public_feedbacks").Error)
	t.Cleanup(func() {
		DB.Exec("DELETE FROM public_feedbacks")
	})
}

func TestCreatePublicFeedbackGeneratesTrackingCodeAndStoresOnlyIpHash(t *testing.T) {
	setupPublicFeedbackModelTest(t)

	feedback, err := CreatePublicFeedback(PublicFeedbackInput{
		UserId:       0,
		ContactName:  "Visitor",
		ContactEmail: "visitor@example.com",
		FeedbackType: PublicFeedbackTypeComplaint,
		Title:        "Billing issue",
		Content:      "Please review the duplicate billing event.",
		IpHash:       "sha256:abc123",
	})

	require.NoError(t, err)
	require.NotZero(t, feedback.Id)
	assert.NotEmpty(t, feedback.TrackingCode)
	assert.Equal(t, PublicFeedbackStatusPending, feedback.Status)
	assert.Equal(t, "sha256:abc123", feedback.IpHash)
	assert.NotEqual(t, "127.0.0.1", feedback.IpHash)
	assert.False(t, DB.Migrator().HasColumn(&PublicFeedback{}, "ip"))
	assert.False(t, DB.Migrator().HasColumn(&PublicFeedback{}, "ip_address"))
}

func TestCreatePublicFeedbackRejectsInvalidType(t *testing.T) {
	setupPublicFeedbackModelTest(t)

	_, err := CreatePublicFeedback(PublicFeedbackInput{
		FeedbackType: "abuse",
		Title:        "Invalid type",
		Content:      "This type should be rejected.",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "feedback type")
}

func TestCreatePublicFeedbackRejectsEmptyContent(t *testing.T) {
	setupPublicFeedbackModelTest(t)

	_, err := CreatePublicFeedback(PublicFeedbackInput{
		FeedbackType: PublicFeedbackTypeFeedback,
		Title:        "Empty content",
		Content:      "   ",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "content")
}

func TestCreatePublicFeedbackRejectsContentTooLong(t *testing.T) {
	setupPublicFeedbackModelTest(t)

	_, err := CreatePublicFeedback(PublicFeedbackInput{
		FeedbackType: PublicFeedbackTypeFeedback,
		Title:        "Too long",
		Content:      strings.Repeat("x", publicFeedbackContentMax+1),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "too long")
}

func TestPublicFeedbackListAndReadByID(t *testing.T) {
	setupPublicFeedbackModelTest(t)

	first, err := CreatePublicFeedback(PublicFeedbackInput{
		UserId:       501,
		Username:     "alice",
		FeedbackType: PublicFeedbackTypeFeedback,
		Title:        "First",
		Content:      "First feedback content.",
		IpHash:       "hash-a",
	})
	require.NoError(t, err)
	_, err = CreatePublicFeedback(PublicFeedbackInput{
		UserId:       502,
		Username:     "bob",
		FeedbackType: PublicFeedbackTypeOther,
		Title:        "Second",
		Content:      "Second feedback content.",
		IpHash:       "hash-b",
	})
	require.NoError(t, err)

	found, err := GetPublicFeedbackByID(first.Id)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, first.Id, found.Id)

	items, total, err := ListPublicFeedback(&common.PageInfo{Page: 1, PageSize: 10}, PublicFeedbackFilter{
		Status:       PublicFeedbackStatusPending,
		FeedbackType: PublicFeedbackTypeFeedback,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	assert.Equal(t, first.Id, items[0].Id)
}

func TestGetPublicFeedbackByTrackingCode(t *testing.T) {
	setupPublicFeedbackModelTest(t)

	feedback, err := CreatePublicFeedback(PublicFeedbackInput{
		UserId:       701,
		Username:     "alice",
		FeedbackType: PublicFeedbackTypeComplaint,
		Title:        "Trackable complaint",
		Content:      "Please provide a status update.",
		IpHash:       "hash-track",
	})
	require.NoError(t, err)

	found, err := GetPublicFeedbackByTrackingCode(feedback.TrackingCode)

	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, feedback.Id, found.Id)
	assert.Equal(t, feedback.TrackingCode, found.TrackingCode)
}

func TestGetPublicFeedbackByTrackingCodeRejectsEmptyCode(t *testing.T) {
	setupPublicFeedbackModelTest(t)

	found, err := GetPublicFeedbackByTrackingCode("   ")

	require.Error(t, err)
	assert.Nil(t, found)
	assert.Contains(t, err.Error(), "tracking code")
}

func TestGetUserPublicFeedbackByIDIsScopedToOwner(t *testing.T) {
	setupPublicFeedbackModelTest(t)

	ownFeedback, err := CreatePublicFeedback(PublicFeedbackInput{
		UserId:       801,
		Username:     "alice",
		FeedbackType: PublicFeedbackTypeFeedback,
		Title:        "Own feedback",
		Content:      "This should be visible to the owner.",
		IpHash:       "hash-owner",
	})
	require.NoError(t, err)
	otherFeedback, err := CreatePublicFeedback(PublicFeedbackInput{
		UserId:       802,
		Username:     "bob",
		FeedbackType: PublicFeedbackTypeComplaint,
		Title:        "Other feedback",
		Content:      "This should not be visible to Alice.",
		IpHash:       "hash-other",
	})
	require.NoError(t, err)

	found, err := GetUserPublicFeedbackByID(801, ownFeedback.Id)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, ownFeedback.Id, found.Id)

	notFound, err := GetUserPublicFeedbackByID(801, otherFeedback.Id)
	require.NoError(t, err)
	assert.Nil(t, notFound)

	invalidUser, err := GetUserPublicFeedbackByID(0, ownFeedback.Id)
	require.Error(t, err)
	assert.Nil(t, invalidUser)
}

func TestAdminUpdatePublicFeedbackRejectsInvalidStatus(t *testing.T) {
	setupPublicFeedbackModelTest(t)

	feedback, err := CreatePublicFeedback(PublicFeedbackInput{
		UserId:       602,
		Username:     "alice",
		FeedbackType: PublicFeedbackTypeComplaint,
		Title:        "Complaint",
		Content:      "Please resolve this complaint.",
		IpHash:       "hash-a",
	})
	require.NoError(t, err)

	_, err = UpdatePublicFeedbackStatus(feedback.Id, "done", 7, "admin", "invalid")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "status")
}

func TestAdminUpdatePublicFeedbackStatusSetsHandledAt(t *testing.T) {
	setupPublicFeedbackModelTest(t)

	feedback, err := CreatePublicFeedback(PublicFeedbackInput{
		UserId:       601,
		Username:     "alice",
		FeedbackType: PublicFeedbackTypeComplaint,
		Title:        "Complaint",
		Content:      "Please resolve this complaint.",
		IpHash:       "hash-a",
	})
	require.NoError(t, err)

	updated, err := UpdatePublicFeedbackStatus(feedback.Id, PublicFeedbackStatusResolved, 7, "admin", "resolved")

	require.NoError(t, err)
	assert.Equal(t, PublicFeedbackStatusResolved, updated.Status)
	assert.Equal(t, 7, updated.AdminId)
	assert.Equal(t, "admin", updated.AdminName)
	assert.Equal(t, "resolved", updated.AdminNote)
	assert.NotZero(t, updated.HandledAt)
}
