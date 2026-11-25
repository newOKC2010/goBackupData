package handlerMophAlert

import (
	mophAlertUtils "backup_data/controller/alert/moph/utils"
)

// CreateBackupFlexMessage flex message สรุปผล backup
func CreateBackupFlexMessage(ok bool, ts string, results []string) mophAlertUtils.FlexMessage {
	status, color := "✅ สำเร็จทั้งหมด", "#16a34a"
	if !ok {
		status, color = "❌ มีรายการล้มเหลว", "#dc2626"
	}

	body := []interface{}{
		mophAlertUtils.FlexText{
			Type:   "text",
			Text:   status,
			Weight: "bold",
			Size:   "sm",
			Color:  color,
			Wrap:   true,
		},
		mophAlertUtils.FlexText{
			Type:  "text",
			Text:  "เวลา: " + ts,
			Size:  "sm",
			Color: "#333333",
		},
	}
	for _, r := range results {
		body = append(body, mophAlertUtils.FlexText{
			Type:  "text",
			Text:  "• " + r,
			Size:  "sm",
			Color: "#333333",
			Wrap:  true,
		})
	}

	return mophAlertUtils.FlexMessage{
		Type:    "flex",
		AltText: "[BACKUP] " + status,
		Contents: mophAlertUtils.FlexBubble{
			Type: "bubble",
			Header: &mophAlertUtils.FlexBox{
				Type:   "box",
				Layout: "vertical",
				Contents: []interface{}{
					mophAlertUtils.FlexText{
						Type:   "text",
						Text:   "💾 DB Backup Report",
						Weight: "bold",
						Size:   "md",
						Color:  "#6ee778ff",
					},
				},
			},
			Body: &mophAlertUtils.FlexBox{
				Type:     "box",
				Layout:   "vertical",
				Spacing:  "md",
				Contents: body,
			},
		},
	}
}
