package model

import (
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	PrivacyRequestTypeAccess     = "access"
	PrivacyRequestTypeCorrection = "correction"
	PrivacyRequestTypeDeletion   = "deletion"

	PrivacyRequestStatusPending    = "pending"
	PrivacyRequestStatusProcessing = "processing"
	PrivacyRequestStatusCompleted  = "completed"
	PrivacyRequestStatusRejected   = "rejected"
	PrivacyRequestStatusCancelled  = "cancelled"
)

const (
	privacyRequestContentMax = 4000
	privacyRequestNoteMax    = 2000
)

var (
	ErrPrivacyRequestInvalidAccountDeletion = errors.New("account deletion can only be executed for completed deletion requests")
	ErrPrivacyRequestRootDeletion           = errors.New("root user cannot be deleted")
)

type PrivacyRequest struct {
	Id                     int    `json:"id"`
	UserId                 int    `json:"user_id" gorm:"index"`
	Username               string `json:"username" gorm:"type:varchar(64);index"`
	ContactName            string `json:"contact_name" gorm:"type:varchar(64)"`
	ContactEmail           string `json:"contact_email" gorm:"type:varchar(128)"`
	ContactPhone           string `json:"contact_phone" gorm:"type:varchar(32)"`
	RequestType            string `json:"request_type" gorm:"type:varchar(32);index"`
	Content                string `json:"content" gorm:"type:text"`
	Status                 string `json:"status" gorm:"type:varchar(32);index"`
	AdminId                int    `json:"admin_id" gorm:"index"`
	AdminName              string `json:"admin_name" gorm:"type:varchar(64)"`
	AdminNote              string `json:"admin_note" gorm:"type:text"`
	ExecuteAccountDeletion bool   `json:"execute_account_deletion" gorm:"default:false"`
	CreatedAt              int64  `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt              int64  `json:"updated_at" gorm:"autoUpdateTime"`
	HandledAt              int64  `json:"handled_at" gorm:"default:0;index"`
}

type PrivacyRequestInput struct {
	UserId       int
	Username     string
	ContactName  string
	ContactEmail string
	ContactPhone string
	RequestType  string
	Content      string
}

type PrivacyRequestFilter struct {
	UserId      int
	Status      string
	RequestType string
}

type PrivacyRequestAdminUpdate struct {
	Status                 string
	AdminId                int
	AdminName              string
	AdminNote              string
	ExecuteAccountDeletion bool
}

func isValidPrivacyRequestType(value string) bool {
	switch value {
	case PrivacyRequestTypeAccess, PrivacyRequestTypeCorrection, PrivacyRequestTypeDeletion:
		return true
	default:
		return false
	}
}

func isValidPrivacyRequestStatus(value string) bool {
	switch value {
	case PrivacyRequestStatusPending, PrivacyRequestStatusProcessing, PrivacyRequestStatusCompleted, PrivacyRequestStatusRejected, PrivacyRequestStatusCancelled:
		return true
	default:
		return false
	}
}

func isTerminalPrivacyRequestStatus(value string) bool {
	switch value {
	case PrivacyRequestStatusCompleted, PrivacyRequestStatusRejected, PrivacyRequestStatusCancelled:
		return true
	default:
		return false
	}
}

func validatePrivacyRequestContent(content string) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", errors.New("request content is required")
	}
	if len([]rune(content)) > privacyRequestContentMax {
		return "", errors.New("request content is too long")
	}
	return content, nil
}

func CreatePrivacyRequest(input PrivacyRequestInput) (*PrivacyRequest, error) {
	requestType := strings.TrimSpace(input.RequestType)
	if !isValidPrivacyRequestType(requestType) {
		return nil, errors.New("invalid privacy request type")
	}
	content, err := validatePrivacyRequestContent(input.Content)
	if err != nil {
		return nil, err
	}
	if input.UserId <= 0 {
		return nil, errors.New("user id is required")
	}

	record := &PrivacyRequest{
		UserId:       input.UserId,
		Username:     strings.TrimSpace(input.Username),
		ContactName:  strings.TrimSpace(input.ContactName),
		ContactEmail: strings.TrimSpace(input.ContactEmail),
		ContactPhone: strings.TrimSpace(input.ContactPhone),
		RequestType:  requestType,
		Content:      content,
		Status:       PrivacyRequestStatusPending,
	}
	if err := DB.Create(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

func privacyRequestQuery(filter PrivacyRequestFilter) *gorm.DB {
	query := DB.Model(&PrivacyRequest{})
	if filter.UserId > 0 {
		query = query.Where("user_id = ?", filter.UserId)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.RequestType != "" {
		query = query.Where("request_type = ?", filter.RequestType)
	}
	return query
}

func ListPrivacyRequests(pageInfo *common.PageInfo, filter PrivacyRequestFilter) ([]*PrivacyRequest, int64, error) {
	if pageInfo == nil {
		pageInfo = &common.PageInfo{Page: 1, PageSize: 20}
	}
	query := privacyRequestQuery(filter)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []*PrivacyRequest
	err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&records).Error
	return records, total, err
}

func ListUserPrivacyRequests(userId int, pageInfo *common.PageInfo) ([]*PrivacyRequest, int64, error) {
	return ListPrivacyRequests(pageInfo, PrivacyRequestFilter{UserId: userId})
}

func GetPrivacyRequestById(id int) (*PrivacyRequest, error) {
	if id <= 0 {
		return nil, errors.New("invalid privacy request id")
	}
	var record PrivacyRequest
	if err := DB.First(&record, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func GetUserPrivacyRequestByID(userId int, id int) (*PrivacyRequest, error) {
	if userId <= 0 {
		return nil, errors.New("invalid user id")
	}
	if id <= 0 {
		return nil, errors.New("invalid privacy request id")
	}
	var record PrivacyRequest
	if err := DB.Where("id = ? AND user_id = ?", id, userId).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func CancelUserPrivacyRequest(id int, userId int) (*PrivacyRequest, error) {
	record, err := GetPrivacyRequestById(id)
	if err != nil {
		return nil, err
	}
	if record.UserId != userId {
		return nil, errors.New("privacy request does not belong to user")
	}
	if record.Status != PrivacyRequestStatusPending && record.Status != PrivacyRequestStatusProcessing {
		return nil, errors.New("privacy request cannot be cancelled")
	}
	now := common.GetTimestamp()
	updates := map[string]interface{}{
		"status":     PrivacyRequestStatusCancelled,
		"updated_at": now,
		"handled_at": now,
	}
	if err := DB.Model(record).Updates(updates).Error; err != nil {
		return nil, err
	}
	return GetPrivacyRequestById(id)
}

func normalizePrivacyRequestAdminUpdate(input PrivacyRequestAdminUpdate) (string, string, error) {
	status := strings.TrimSpace(input.Status)
	if !isValidPrivacyRequestStatus(status) {
		return "", "", errors.New("invalid privacy request status")
	}
	adminNote := strings.TrimSpace(input.AdminNote)
	if len([]rune(adminNote)) > privacyRequestNoteMax {
		return "", "", errors.New("admin note is too long")
	}
	return status, adminNote, nil
}

func buildPrivacyRequestAdminUpdates(input PrivacyRequestAdminUpdate, status string, adminNote string, now int64) map[string]interface{} {
	updates := map[string]interface{}{
		"status":                   status,
		"admin_id":                 input.AdminId,
		"admin_name":               strings.TrimSpace(input.AdminName),
		"admin_note":               adminNote,
		"execute_account_deletion": input.ExecuteAccountDeletion,
		"updated_at":               now,
	}
	if isTerminalPrivacyRequestStatus(status) {
		updates["handled_at"] = now
	} else {
		updates["handled_at"] = int64(0)
	}
	return updates
}

func UpdatePrivacyRequestByAdmin(id int, input PrivacyRequestAdminUpdate) (*PrivacyRequest, error) {
	record, err := GetPrivacyRequestById(id)
	if err != nil {
		return nil, err
	}
	status, adminNote, err := normalizePrivacyRequestAdminUpdate(input)
	if err != nil {
		return nil, err
	}
	now := common.GetTimestamp()
	if err := DB.Model(record).Updates(buildPrivacyRequestAdminUpdates(input, status, adminNote, now)).Error; err != nil {
		return nil, err
	}
	return GetPrivacyRequestById(id)
}

func UpdatePrivacyRequestByAdminAndDeleteUser(id int, input PrivacyRequestAdminUpdate) (*PrivacyRequest, error) {
	status, adminNote, err := normalizePrivacyRequestAdminUpdate(input)
	if err != nil {
		return nil, err
	}
	if !input.ExecuteAccountDeletion {
		return UpdatePrivacyRequestByAdmin(id, input)
	}

	var updated PrivacyRequest
	deletedUserId := 0
	err = DB.Transaction(func(tx *gorm.DB) error {
		var record PrivacyRequest
		if err := tx.First(&record, "id = ?", id).Error; err != nil {
			return err
		}
		if record.RequestType != PrivacyRequestTypeDeletion || status != PrivacyRequestStatusCompleted {
			return ErrPrivacyRequestInvalidAccountDeletion
		}

		var targetUser User
		if err := tx.First(&targetUser, "id = ?", record.UserId).Error; err != nil {
			return err
		}
		if targetUser.Role == common.RoleRootUser {
			return ErrPrivacyRequestRootDeletion
		}

		now := common.GetTimestamp()
		if err := tx.Model(&record).Updates(buildPrivacyRequestAdminUpdates(input, status, adminNote, now)).Error; err != nil {
			return err
		}
		if err := tx.Delete(&User{Id: record.UserId}).Error; err != nil {
			return err
		}
		if err := tx.First(&updated, "id = ?", id).Error; err != nil {
			return err
		}
		deletedUserId = record.UserId
		return nil
	})
	if err != nil {
		return nil, err
	}
	if deletedUserId > 0 {
		if err := invalidateUserCache(deletedUserId); err != nil {
			return nil, err
		}
		if err := InvalidateUserTokensCache(deletedUserId); err != nil {
			common.SysLog(fmt.Sprintf("failed to invalidate tokens cache for user %d: %s", deletedUserId, err.Error()))
		}
	}
	return &updated, nil
}
