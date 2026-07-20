package controller

import (
	"errors"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type publicFeedbackCreateRequest struct {
	Type         string `json:"type"`
	FeedbackType string `json:"feedback_type"`
	ContactName  string `json:"contact_name"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
	Title        string `json:"title"`
	Content      string `json:"content"`
}

type publicFeedbackAdminUpdateRequest struct {
	Status    string `json:"status"`
	AdminNote string `json:"admin_note"`
}

type publicFeedbackTrackResponse struct {
	TrackingCode string `json:"tracking_code"`
	FeedbackType string `json:"feedback_type"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Status       string `json:"status"`
	AdminNote    string `json:"admin_note"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
	HandledAt    int64  `json:"handled_at"`
}

func buildPublicFeedbackTrackResponse(record *model.PublicFeedback) publicFeedbackTrackResponse {
	return publicFeedbackTrackResponse{
		TrackingCode: record.TrackingCode,
		FeedbackType: record.FeedbackType,
		Title:        record.Title,
		Content:      record.Content,
		Status:       record.Status,
		AdminNote:    record.AdminNote,
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
		HandledAt:    record.HandledAt,
	}
}

func currentOptionalFeedbackUser(c *gin.Context) (int, string) {
	session := sessions.Default(c)
	idValue := session.Get("id")
	id, ok := idValue.(int)
	if !ok || id <= 0 {
		return 0, ""
	}
	user, err := model.GetUserById(id, false)
	if err != nil {
		return 0, ""
	}
	return user.Id, user.Username
}

func CreatePublicFeedback(c *gin.Context) {
	var req publicFeedbackCreateRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	userId, username := currentOptionalFeedbackUser(c)
	feedbackType := req.FeedbackType
	if feedbackType == "" {
		feedbackType = req.Type
	}
	record, err := model.CreatePublicFeedback(model.PublicFeedbackInput{
		UserId:       userId,
		Username:     username,
		ContactName:  req.ContactName,
		ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone,
		FeedbackType: feedbackType,
		Title:        req.Title,
		Content:      req.Content,
		IpHash:       common.GenerateHMAC(c.ClientIP()),
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{
		"id":            record.Id,
		"tracking_code": record.TrackingCode,
	})
}

func ListMyFeedback(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	records, total, err := model.ListUserPublicFeedback(c.GetInt("id"), pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(records)
	common.ApiSuccess(c, pageInfo)
}

func GetMyFeedbackDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	record, err := model.GetUserPublicFeedbackByID(c.GetInt("id"), id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if record == nil {
		common.ApiError(c, errors.New("feedback record not found"))
		return
	}
	common.ApiSuccess(c, record)
}

func TrackPublicFeedback(c *gin.Context) {
	record, err := model.GetPublicFeedbackByTrackingCode(c.Param("tracking_code"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if record == nil {
		common.ApiError(c, errors.New("feedback record not found"))
		return
	}
	common.ApiSuccess(c, buildPublicFeedbackTrackResponse(record))
}

func AdminListFeedback(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	filter := model.PublicFeedbackFilter{
		Status:       c.Query("status"),
		FeedbackType: c.Query("feedback_type"),
	}
	if filter.FeedbackType == "" {
		filter.FeedbackType = c.Query("type")
	}
	if userIdStr := c.Query("user_id"); userIdStr != "" {
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			common.ApiErrorI18n(c, i18n.MsgInvalidParams)
			return
		}
		filter.UserId = userId
	}
	records, total, err := model.ListPublicFeedback(pageInfo, filter)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(records)
	common.ApiSuccess(c, pageInfo)
}

func AdminGetFeedback(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	record, err := model.GetPublicFeedbackById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, record)
}

func AdminUpdateFeedback(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	var req publicFeedbackAdminUpdateRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	record, err := model.UpdatePublicFeedbackStatus(id, req.Status, c.GetInt("id"), c.GetString("username"), req.AdminNote)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAuditFor(c, record.UserId, "compliance.feedback_update", map[string]interface{}{
		"id":              record.Id,
		"feedback_type":   record.FeedbackType,
		"status":          record.Status,
		"tracking_code":   record.TrackingCode,
		"target_username": record.Username,
	})
	common.ApiSuccess(c, record)
}
