package config

// GrayScaleConfig 灰度开关配置
// 对标：技术方案5.14节 shouldProcessByGo函数
type GrayScaleConfig struct {
	Enabled bool `yaml:"enabled"` // 是否启用Go处理（true=Go, false=Ruby）

	// 灰度策略
	Percentage int `yaml:"percentage"` // 灰度百分比（0-100）
	// 例如：percentage=30 表示30%流量走Go，70%走Ruby

	// 渠道白名单
	EnabledChannels []string `yaml:"enabled_channels"` // 完全走Go的渠道列表
	// 例如：["YunExpress", "GE-YT"] 表示这些渠道100%走Go

	// 单号白名单
	EnabledTrackingNumbers []string `yaml:"enabled_tracking_numbers"` // 测试单号列表
	// 例如：["YT240TEST103"] 用于测试验证
}

// ShouldProcessByGo 判断是否应该由Go处理（灰度开关）
// 对标：技术方案5.14节
// 返回：
//   - true: Go处理（Webhook推送）
//   - false: Ruby处理（降级到Pull API）
func (c *GrayScaleConfig) ShouldProcessByGo(trackingNumber string, channel string) bool {
	// 1. 全局开关检查
	if !c.Enabled {
		return false // 完全走Ruby
	}

	// 2. 单号白名单检查（测试单号100%走Go）
	for _, num := range c.EnabledTrackingNumbers {
		if trackingNumber == num {
			return true
		}
	}

	// 3. 渠道白名单检查（指定渠道100%走Go）
	for _, ch := range c.EnabledChannels {
		if channel == ch {
			return true
		}
	}

	// 4. 百分比灰度检查
	// 根据tracking_number的哈希值判断
	if c.Percentage > 0 {
		hash := hashTrackingNumber(trackingNumber)
		if hash%100 < c.Percentage {
			return true
		}
	}

	// 默认：Ruby处理
	return false
}

// hashTrackingNumber 计算单号的哈希值（用于灰度分流）
func hashTrackingNumber(trackingNumber string) int {
	hash := 0
	for _, c := range trackingNumber {
		hash = (hash * 31 + int(c)) % 100
	}
	return hash
}