package utils

// MaskTrackingNumber 脱敏运单号（用于日志记录）
// 规则：保留前3位和后3位，中间替换为 ***
// 例如：YT2606500704802225 → YT2***225
func MaskTrackingNumber(tn string) string {
	if len(tn) <= 6 {
		return "***" // 过短的号码完全隐藏
	}
	return tn[:3] + "***" + tn[len(tn)-3:]
}

// MaskWaybillNumber 脱敏运单号（别名，语义更明确）
func MaskWaybillNumber(wn string) string {
	return MaskTrackingNumber(wn)
}