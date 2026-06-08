package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

// WeeklyQuotaSetting 周额度领取配置
type WeeklyQuotaSetting struct {
	Enabled bool `json:"enabled"` // 是否启用周额度领取
	Amount  int  `json:"amount"`  // 每个领取周期奖励额度
}

var weeklyQuotaSetting = WeeklyQuotaSetting{
	Enabled: false,
	Amount:  0,
}

func init() {
	config.GlobalConfig.Register("weekly_quota_setting", &weeklyQuotaSetting)
}

// GetWeeklyQuotaSetting 获取周额度领取配置
func GetWeeklyQuotaSetting() *WeeklyQuotaSetting {
	return &weeklyQuotaSetting
}
