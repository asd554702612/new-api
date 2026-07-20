package model

import (
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	PublicFeedbackTypeComplaint = "complaint"
	PublicFeedbackTypeFeedback  = "feedback"
	PublicFeedbackTypeOther     = "other"

	PublicFeedbackStatusPending    = "pending"
	PublicFeedbackStatusProcessing = "processing"
	PublicFeedbackStatusResolved   = "resolved"
	PublicFeedbackStatusClosed     = "closed"
	PublicFeedbackStatusRejected   = "rejected"
)

const (
	publicFeedbackTitleMax   = 120
	publicFeedbackContentMax = 4000
	publicFeedbackNoteMax    = 2000
)

type PublicFeedback struct {
	Id           int    `json:"id"`
	UserId       int    `json:"user_id" gorm:"index"`
	Username     string `json:"username" gorm:"type:varchar(64);index"`
	ContactName  string `json:"contact_name" gorm:"type:varchar(64)"`
	ContactEmail string `json:"contact_email" gorm:"type:varchar(128)"`
	ContactPhone string `json:"contact_phone" gorm:"type:varchar(32)"`
	FeedbackType string `json:"feedback_type" gorm:"type:varchar(32);index"`
	Title        string `json:"title" gorm:"type:varchar(160)"`
	Content      string `json:"content" gorm:"type:text"`
	Status       string `json:"status" gorm:"type:varchar(32);index"`
	TrackingCode string `json:"tracking_code" gorm:"type:varchar(64);uniqueIndex"`
	IpHash       string `json:"-" gorm:"type:varchar(128);index"`
	AdminId      int    `json:"admin_id" gorm:"index"`
	AdminName    string `json:"admin_name" gorm:"type:varchar(64)"`
	AdminNote    string `json:"admin_note" gorm:"type:text"`
	CreatedAt    int64  `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt    int64  `json:"updated_at" gorm:"autoUpdateTime"`
	HandledAt    int64  `json:"handled_at" gorm:"default:0;index"`
}

type PublicFeedbackInput struct {
	UserId       int
	Username     string
	ContactName  string
	ContactEmail string
	ContactPhone string
	FeedbackType string
	Title        string
	Content      string
	IpHash       string
}

type PublicFeedbackFilter struct {
	UserId       int
	Status       string
	FeedbackType string
}

type PublicFeedbackAdminUpdate struct {
	Status    string
	AdminId   int
	AdminName string
	AdminNote string
}

func isValidPublicFeedbackType(value string) bool {
	switch value {
	case PublicFeedbackTypeComplaint, PublicFeedbackTypeFeedback, PublicFeedbackTypeOther:
		return true
	default:
		return false
	}
}

func isValidPublicFeedbackStatus(value string) bool {
	switch value {
	case PublicFeedbackStatusPending, PublicFeedbackStatusProcessing, PublicFeedbackStatusResolved, PublicFeedbackStatusClosed, PublicFeedbackStatusRejected:
		return true
	default:
		return false
	}
}

func isTerminalPublicFeedbackStatus(value string) bool {
	switch value {
	case PublicFeedbackStatusResolved, PublicFeedbackStatusClosed, PublicFeedbackStatusRejected:
		return true
	default:
		return false
	}
}

func validatePublicFeedbackTitleAndContent(title string, content string) (string, string, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" {
		return "", "", errors.New("feedback title is required")
	}
	if content == "" {
		return "", "", errors.New("feedback content is required")
	}
	if len([]rune(title)) > publicFeedbackTitleMax {
		return "", "", errors.New("feedback title is too long")
	}
	if len([]rune(content)) > publicFeedbackContentMax {
		return "", "", errors.New("feedback content is too long")
	}
	return title, content, nil
}

func newPublicFeedbackTrackingCode() string {
	return fmt.Sprintf("FB%s%s", common.GetTimeString(), common.GetRandomString(6))
}

func normalizePublicFeedbackInput(input any) (PublicFeedbackInput, error) {
	switch value := input.(type) {
	case PublicFeedbackInput:
		return value, nil
	case *PublicFeedbackInput:
		if value == nil {
			return PublicFeedbackInput{}, errors.New("public feedback is required")
		}
		return *value, nil
	case PublicFeedback:
		return PublicFeedbackInput{
			UserId:       value.UserId,
			Username:     value.Username,
			ContactName:  value.ContactName,
			ContactEmail: value.ContactEmail,
			ContactPhone: value.ContactPhone,
			FeedbackType: value.FeedbackType,
			Title:        value.Title,
			Content:      value.Content,
			IpHash:       value.IpHash,
		}, nil
	case *PublicFeedback:
		if value == nil {
			return PublicFeedbackInput{}, errors.New("public feedback is required")
		}
		return PublicFeedbackInput{
			UserId:       value.UserId,
			Username:     value.Username,
			ContactName:  value.ContactName,
			ContactEmail: value.ContactEmail,
			ContactPhone: value.ContactPhone,
			FeedbackType: value.FeedbackType,
			Title:        value.Title,
			Content:      value.Content,
			IpHash:       value.IpHash,
		}, nil
	default:
		return PublicFeedbackInput{}, errors.New("public feedback is required")
	}
}

func CreatePublicFeedback(input any) (*PublicFeedback, error) {
	normalized, err := normalizePublicFeedbackInput(input)
	if err != nil {
		return nil, err
	}
	feedbackType := strings.TrimSpace(normalized.FeedbackType)
	if !isValidPublicFeedbackType(feedbackType) {
		return nil, errors.New("invalid public feedback type")
	}
	title, content, err := validatePublicFeedbackTitleAndContent(normalized.Title, normalized.Content)
	if err != nil {
		return nil, err
	}

	record := &PublicFeedback{
		UserId:       normalized.UserId,
		Username:     strings.TrimSpace(normalized.Username),
		ContactName:  strings.TrimSpace(normalized.ContactName),
		ContactEmail: strings.TrimSpace(normalized.ContactEmail),
		ContactPhone: strings.TrimSpace(normalized.ContactPhone),
		FeedbackType: feedbackType,
		Title:        title,
		Content:      content,
		Status:       PublicFeedbackStatusPending,
		TrackingCode: newPublicFeedbackTrackingCode(),
		IpHash:       strings.TrimSpace(normalized.IpHash),
	}
	if err := DB.Create(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

func publicFeedbackQuery(filter PublicFeedbackFilter) *gorm.DB {
	query := DB.Model(&PublicFeedback{})
	if filter.UserId > 0 {
		query = query.Where("user_id = ?", filter.UserId)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.FeedbackType != "" {
		query = query.Where("feedback_type = ?", filter.FeedbackType)
	}
	return query
}

func ListPublicFeedback(pageInfo *common.PageInfo, filter PublicFeedbackFilter) ([]*PublicFeedback, int64, error) {
	if pageInfo == nil {
		pageInfo = &common.PageInfo{Page: 1, PageSize: 20}
	}
	query := publicFeedbackQuery(filter)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []*PublicFeedback
	err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&records).Error
	return records, total, err
}

func ListUserPublicFeedback(userId int, pageInfo *common.PageInfo) ([]*PublicFeedback, int64, error) {
	return ListPublicFeedback(pageInfo, PublicFeedbackFilter{UserId: userId})
}

func GetPublicFeedbackById(id int) (*PublicFeedback, error) {
	if id <= 0 {
		return nil, errors.New("invalid feedback id")
	}
	var record PublicFeedback
	if err := DB.First(&record, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func GetPublicFeedbackByID(id int) (*PublicFeedback, error) {
	return GetPublicFeedbackById(id)
}

func GetPublicFeedbackByTrackingCode(trackingCode string) (*PublicFeedback, error) {
	trackingCode = strings.TrimSpace(trackingCode)
	if trackingCode == "" {
		return nil, errors.New("feedback tracking code is required")
	}
	var record PublicFeedback
	if err := DB.Where("tracking_code = ?", trackingCode).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func GetUserPublicFeedbackByID(userId int, id int) (*PublicFeedback, error) {
	if userId <= 0 {
		return nil, errors.New("user id is required")
	}
	if id <= 0 {
		return nil, errors.New("invalid feedback id")
	}
	var record PublicFeedback
	if err := DB.Where("id = ? AND user_id = ?", id, userId).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func UpdatePublicFeedbackByAdmin(id int, input PublicFeedbackAdminUpdate) (*PublicFeedback, error) {
	record, err := GetPublicFeedbackById(id)
	if err != nil {
		return nil, err
	}
	status := strings.TrimSpace(input.Status)
	if !isValidPublicFeedbackStatus(status) {
		return nil, errors.New("invalid feedback status")
	}
	adminNote := strings.TrimSpace(input.AdminNote)
	if len([]rune(adminNote)) > publicFeedbackNoteMax {
		return nil, errors.New("admin note is too long")
	}
	now := common.GetTimestamp()
	updates := map[string]interface{}{
		"status":     status,
		"admin_id":   input.AdminId,
		"admin_name": strings.TrimSpace(input.AdminName),
		"admin_note": adminNote,
		"updated_at": now,
	}
	if isTerminalPublicFeedbackStatus(status) {
		updates["handled_at"] = now
	} else {
		updates["handled_at"] = int64(0)
	}
	if err := DB.Model(record).Updates(updates).Error; err != nil {
		return nil, err
	}
	return GetPublicFeedbackById(id)
}

func UpdatePublicFeedbackStatus(id int, status string, adminId int, adminName string, adminNote string) (*PublicFeedback, error) {
	return UpdatePublicFeedbackByAdmin(id, PublicFeedbackAdminUpdate{
		Status:    status,
		AdminId:   adminId,
		AdminName: adminName,
		AdminNote: adminNote,
	})
}
