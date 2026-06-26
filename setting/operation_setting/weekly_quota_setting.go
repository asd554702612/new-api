package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

// WeeklyQuotaSetting 周额度领取配置
type WeeklyQuotaSetting struct {
	Enabled    bool `json:"enabled"`     // 是否启用领取套餐
	Amount     int  `json:"amount"`      // 旧版周额度金额，保留用于兼容历史配置
	PlanId     int  `json:"plan_id"`     // 可领取的订阅套餐 ID
	PeriodDays int  `json:"period_days"` // 领取周期天数
}

var weeklyQuotaSetting = WeeklyQuotaSetting{
	Enabled:    false,
	Amount:     0,
	PlanId:     0,
	PeriodDays: 7,
}

func init() {
	config.GlobalConfig.Register("weekly_quota_setting", &weeklyQuotaSetting)
}

// GetWeeklyQuotaSetting 获取周额度领取配置
func GetWeeklyQuotaSetting() *WeeklyQuotaSetting {
	return &weeklyQuotaSetting
}
