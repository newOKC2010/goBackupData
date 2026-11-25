package mophAlert

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	handlerMophAlert "backup_data/controller/alert/moph/handler"
	mophAlertUtils "backup_data/controller/alert/moph/utils"
)

// Config ค่าจาก env MOPH_ALERT_* + CID_ALERT
type Config struct {
	Method    string
	URL       string
	ClientID  string
	SecretKey string
	CIDs      []string
}

// SendBackupAlert ส่ง flex message สรุปผล backup ให้ทุก CID ที่กำหนด
func SendBackupAlert(cfg Config, ok bool, ts string, results []string) mophAlertUtils.AlertResponse {
	payload := mophAlertUtils.FlexAlertPayload{
		CID:      cfg.CIDs,
		Messages: []mophAlertUtils.FlexMessage{handlerMophAlert.CreateBackupFlexMessage(ok, ts, results)},
	}
	return SendAlert(cfg, payload)
}

// SendAlert POST payload ไป MOPH Alert API
func SendAlert(cfg Config, payload mophAlertUtils.FlexAlertPayload) mophAlertUtils.AlertResponse {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[ALERT][ERROR] JSON error: %v", err)
		return mophAlertUtils.AlertResponse{Message: "ส่งแจ้งเตือนไม่สำเร็จ", MessageCode: 500}
	}

	method := cfg.Method
	if method == "" {
		method = "POST"
	}

	req, err := http.NewRequest(strings.ToUpper(method), cfg.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[ALERT][ERROR] สร้าง request ไม่ได้: %v", err)
		return mophAlertUtils.AlertResponse{Message: "ส่งแจ้งเตือนไม่สำเร็จ", MessageCode: 500}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("client-key", cfg.ClientID)
	req.Header.Set("secret-key", cfg.SecretKey)

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ALERT][ERROR] ส่ง alert ไม่ได้: %v", err)
		return mophAlertUtils.AlertResponse{Message: "ส่งแจ้งเตือนไม่สำเร็จ", MessageCode: 500}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 2048))
	if err != nil {
		log.Printf("[ALERT][ERROR] อ่าน response ไม่ได้: %v", err)
		return mophAlertUtils.AlertResponse{Message: "ส่งแจ้งเตือนไม่สำเร็จ", MessageCode: 500}
	}

	var result mophAlertUtils.AlertResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		log.Printf("[ALERT][ERROR] response decode ไม่ได้: %v (status %d): %s", err, resp.StatusCode, bodyBytes)
		return mophAlertUtils.AlertResponse{Message: "ส่งแจ้งเตือนไม่สำเร็จ", MessageCode: 500}
	}

	log.Printf("[ALERT] ส่งถึง %d CID — status %d: %s", len(payload.CID), result.MessageCode, result.Message)
	return result
}
