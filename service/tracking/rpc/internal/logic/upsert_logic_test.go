package logic

import (
	"strconv"
	"testing"
)

// TestGetStatusText 测试状态码文本映射
// 对标：Ruby yun_express_track_formatter.rb:6-19 (STATUS_CODES)
func TestGetStatusText(t *testing.T) {
	tests := []struct {
		name       string
		statusCode string
		expected   string
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
		{"Undefined", "999", "Undefined"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStatusText(tt.statusCode)
			if result != tt.expected {
				t.Errorf("getStatusText(%s) = %s, expected %s",
					tt.statusCode, result, tt.expected)
			}
		})
	}
}

// TestMapStatusTextToTrackStatus 测试状态文本 → TrackStatus 映射（不含海外包裹逻辑）
// 对标：Ruby tracking_detail.rb:80-88
func TestMapStatusTextToTrackStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   int32
		expected int32
	}{
		{"InTransit → in_transit", 20, 1},
		{"InTransit_Arrival → in_transit", 1001, 1},
		{"Delivered → delivered", 50, 2},
		{"AvailableForPickup → delivered", 30, 2},
		{"Exception → track_exception", 60, 3},
		{"DeliveryFailure → track_exception", 40, 3},
		{"Expired → track_exception", 70, 3},
		{"海关查验 → track_exception", 80, 3},
		{"Exception_Returned → track_exception", 90, 3},
		{"Exception_Cancel → track_exception", 100, 3},
		{"NotFound → to_receive", 0, 0},
		{"InfoReceived → to_receive（默认，需海外包裹检查）", 10, 0},
		{"Undefined → to_receive", 999, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用 getStatusText + 状态文本映射规则验证
			statusTxt := getStatusText(strconv.Itoa(int(tt.status)))
			result := mapStatusTextToTrackStatusForTest(statusTxt)

			if result != tt.expected {
				t.Errorf("mapStatusTextToTrackStatus(%s) = %d, expected %d",
					statusTxt, result, tt.expected)
			}
		})
	}
}

// mapStatusTextToTrackStatusForTest 状态文本 → TrackStatus 映射（测试用）
// 注意：InfoReceived 海外包裹逻辑在 mapTrackStatus 中实现，这里返回默认值
func mapStatusTextToTrackStatusForTest(statusTxt string) int32 {
	switch statusTxt {
	case "InTransit", "InTransit_Arrival":
		return 1 // in_transit
	case "Delivered", "AvailableForPickup":
		return 2 // delivered
	case "DeliveryFailure", "Exception", "Expired", "Exception_Returned", "Exception_Cancel":
		return 3 // track_exception
	default:
		return 0 // to_receive（包含 InfoReceived、NotFound、Undefined）
	}
}