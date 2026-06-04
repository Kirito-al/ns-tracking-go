package formatter

import "testing"

// ===================== 状态码映射测试 =====================

func TestGetStatusText(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"NotFound", "0", "NotFound"},
		{"InfoReceived", "10", "InfoReceived"},
		{"InTransit", "20", "InTransit"},
		{"AvailableForPickup", "30", "AvailableForPickup"},
		{"DeliveryFailure", "40", "DeliveryFailure"},
		{"Delivered", "50", "Delivered"},
		{"Exception", "60", "Exception"},
		{"Expired", "70", "Expired"},
		{"海关查验", "80", "Exception"},
		{"Exception_Returned", "90", "Exception_Returned"},
		{"Exception_Cancel", "100", "Exception_Cancel"},
		{"InTransit_Arrival", "1001", "InTransit_Arrival"},
		{"Unknown", "999", "Undefined"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStatusText(tt.status)
			if result != tt.expected {
				t.Errorf("GetStatusText(%s) = %s, want %s", tt.status, result, tt.expected)
			}
		})
	}
}

func TestGetPackageState(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"NotFound", "0", "0"},
		{"InfoReceived", "10", "4"},
		{"InTransit", "20", "2"},
		{"AvailableForPickup", "30", "2"},
		{"DeliveryFailure", "40", "6"},
		{"Delivered", "50", "3"},
		{"Exception", "60", "6"},
		{"Expired", "70", "6"},
		{"海关查验改写前", "80", "6"},
		{"Exception_Returned", "90", "7"},
		{"Exception_Cancel", "100", "5"},
		{"InTransit_Arrival", "1001", "2"},
		{"Unknown", "999", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPackageState(tt.status)
			if result != tt.expected {
				t.Errorf("GetPackageState(%s) = %s, want %s", tt.status, result, tt.expected)
			}
		})
	}
}

// ===================== 海关查验改写测试 =====================

func TestFixCustomsInspectionStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"海关查验改写", "80", "20"},
		{"其他状态不变", "20", "20"},
		{"其他状态不变", "50", "50"},
		{"Unknown状态不变", "999", "999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FixCustomsInspectionStatus(tt.status)
			if result != tt.expected {
				t.Errorf("FixCustomsInspectionStatus(%s) = %s, want %s", tt.status, result, tt.expected)
			}
		})
	}
}

func TestFixCustomsInspectionStatusAndPackageState(t *testing.T) {
	// 测试：海关查验改写后，PackageState 应该自动正确
	status := "80"

	// 1. 改写状态码
	fixedStatus := FixCustomsInspectionStatus(status)
	if fixedStatus != "20" {
		t.Errorf("FixCustomsInspectionStatus(80) = %s, want 20", fixedStatus)
	}

	// 2. 查表获取 PackageState（应该自动正确）
	packageState := GetPackageState(fixedStatus)
	if packageState != "2" {
		t.Errorf("GetPackageState(20) = %s, want 2 (InTransit)", packageState)
	}

	t.Logf("海关查验改写成功：80 → 20, PackageState: 6 → 2")
}

// ===================== 完整流程测试（可选）====================

func TestYunExpressFormatter_Format(t *testing.T) {
	// 模拟云途 Webhook Payload
	raw := map[string]interface{}{
		"trackingNumber": "YTTEST001",
		"wayBillNumber":  "YTTEST001",
		"providerName":   "云途物流",
		"orderTrackingDetails": []map[string]interface{}{
			{
				"processDate":     "2025-01-01T10:00:00Z",
				"processLocation": "深圳仓库",
				"processContent":  "快件电子信息已收到",
				"trackingStatus":  "10",
			},
			{
				"processDate":     "2025-01-02T15:00:00Z",
				"processLocation": "深圳口岸",
				"processContent":  "快件已发出",
				"trackingStatus":  "20",
			},
		},
	}

	// 先测试数据提取
	details := ExtractOrderTrackingDetails(raw)
	t.Logf("ExtractOrderTrackingDetails: len=%d", len(details))
	if len(details) == 0 {
		t.Fatal("ExtractOrderTrackingDetails returned 0 events")
	}

	// 测试字段映射
	mappedFirst := MapEventFields(details[0])
	t.Logf("MapEventFields first event: time=%v, location=%v, content=%v, TrackingStatus=%v",
		mappedFirst["time"], mappedFirst["location"], mappedFirst["content"], mappedFirst["TrackingStatus"])

	// 测试事件过滤
	mappedAll := MapAllEvents(details)
	filtered := FilterOrderTrackingDetails(mappedAll)
	t.Logf("FilterOrderTrackingDetails: input=%d, output=%d", len(mappedAll), len(filtered))

	// 执行格式化
	formatter := NewYunExpressFormatter(raw)
	result := formatter.Format()

	// 验证结果
	if result.Response.Item.TrackingStatus != "20" {
		t.Errorf("TrackingStatus = %s, want 20", result.Response.Item.TrackingStatus)
	}

	if result.Response.Item.PackageState != "2" {
		t.Errorf("PackageState = %s, want 2", result.Response.Item.PackageState)
	}

	if len(result.Response.Item.OrderTrackingDetails) != 2 {
		t.Errorf("OrderTrackingDetails length = %d, want 2", len(result.Response.Item.OrderTrackingDetails))
	}

	t.Logf("格式化成功：状态码=%s, 包裹状态=%s, synced_at=%s", 
		result.Response.Item.TrackingStatus, result.Response.Item.PackageState, result.SyncedAt)
}