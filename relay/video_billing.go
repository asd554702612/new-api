package relay

import (
	"io"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/pkg/videobilling"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/video_billing_setting"
	"github.com/gin-gonic/gin"
)

func applyConfiguredVideoBillingPriceData(c *gin.Context, info *relaycommon.RelayInfo) (bool, error) {
	if _, ok := video_billing_setting.GetRule(info.OriginModelName); !ok {
		return false, nil
	}

	info.PriceData.GroupRatioInfo = helper.HandleGroupRatio(c, info)
	return applyConfiguredVideoBilling(c, info)
}

func applyConfiguredVideoBilling(c *gin.Context, info *relaycommon.RelayInfo) (bool, error) {
	rule, ok := video_billing_setting.GetRule(info.OriginModelName)
	if !ok {
		return false, nil
	}

	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return false, err
	}

	action := ""
	if info.TaskRelayInfo != nil {
		action = info.Action
	}

	var body []byte
	if storageAny, exists := c.Get(common.KeyBodyStorage); exists && storageAny != nil {
		if storage, ok := storageAny.(common.BodyStorage); ok {
			body, _ = storage.Bytes()
			_, _ = storage.Seek(0, io.SeekStart)
		}
	}

	result, err := videobilling.Calculate(rule, videobilling.Input{
		Model:       info.OriginModelName,
		Action:      action,
		ChannelType: videoBillingChannelType(info),
		Request:     req,
		Body:        body,
	})
	if err != nil {
		return false, err
	}

	groupRatio := info.PriceData.GroupRatioInfo.GroupRatio
	quota := int(result.AmountUSD * common.QuotaPerUnit * groupRatio)
	freeModel := false
	if !operation_setting.GetQuotaSetting().EnableFreeModelPreConsume && (groupRatio == 0 || result.AmountUSD == 0) {
		quota = 0
		freeModel = true
	}

	otherRatios := map[string]float64{
		"video_seconds": float64(result.Seconds),
	}
	for key, value := range result.Factors {
		otherRatios["video_"+key] = value
	}

	info.PriceData.FreeModel = freeModel
	info.PriceData.ModelPrice = result.AmountUSD
	info.PriceData.ModelRatio = 0
	info.PriceData.UsePrice = true
	info.PriceData.Quota = quota
	info.PriceData.OtherRatios = otherRatios
	return true, nil
}

func videoBillingChannelType(info *relaycommon.RelayInfo) int {
	if info == nil || info.ChannelMeta == nil {
		return 0
	}
	return info.ChannelType
}
