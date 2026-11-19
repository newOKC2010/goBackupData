package controller

import (
	"log"
	"strings"
	"time"
)

// Start วนลูปรอเวลาแล้วสั่ง backup ทุกวัน
func Start(cfg *Config) {
	log.Printf("backup พร้อมทำงาน — ทุกวันเวลา %s (database: %s)", cfg.TimeStr, strings.Join(cfg.DBNames, ", "))

	for {
		now := time.Now()
		if now.Hour() == cfg.Hour && now.Minute() == cfg.Minute {
			RunBackup(cfg)
			logNextRun(cfg, now)
			time.Sleep(61 * time.Second) // ข้ามนาทีนี้ป้องกัน backup ซ้ำ
		} else {
			time.Sleep(30 * time.Second)
		}
	}
}

func logNextRun(cfg *Config, from time.Time) {
	next := time.Date(from.Year(), from.Month(), from.Day(), cfg.Hour, cfg.Minute, 0, 0, from.Location())
	if !from.Before(next) {
		next = next.AddDate(0, 0, 1)
	}
	log.Printf("backup ครั้งถัดไป: %s (database: %s)", next.Format("2006-01-02 15:04"), strings.Join(cfg.DBNames, ", "))
}
