package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

type updateUserModelSelectionsRequest struct {
	Models []string `json:"models"`
}

func GetUserModelSelections(c *gin.Context) {
	userId := c.GetInt("id")
	selections, err := model.GetUserModelSelections(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    selections,
	})
}

func UpdateUserModelSelections(c *gin.Context) {
	var req updateUserModelSelectionsRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	selected, err := service.ReplaceUserModelSelectionsForUsableModels(c.GetInt("id"), req.Models)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    selected,
	})
}
