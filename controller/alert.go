package controller

import mophAlert "backup_data/controller/alert/moph"

// sendAlert ส่งสรุปผล backup ไป MOPH Alert ตาม CID ที่ตั้งใน env
func sendAlert(cfg *Config, ok bool, ts string, results []string) {
	mophAlert.SendBackupAlert(mophAlert.Config{
		Method:    cfg.AlertMethod,
		URL:       cfg.AlertURL,
		ClientID:  cfg.AlertClientID,
		SecretKey: cfg.AlertSecret,
		CIDs:      cfg.AlertCIDs,
	}, ok, ts, results)
}
