package controller

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// cleanOldFiles ลบไฟล์ .dump ที่อายุเกิน RetainDays ในทุกโฟลเดอร์ database
func cleanOldFiles(cfg *Config, base string) {
	log.Printf("[CLEANUP] ตรวจสอบไฟล์เก่า — เก็บย้อนหลัง %d วัน", cfg.RetainDays)
	limit := time.Duration(cfg.RetainDays) * 24 * time.Hour
	deleted := 0

	for _, db := range cfg.DBNames {
		dir := filepath.Join(base, db)
		entries, err := os.ReadDir(dir)
		if err != nil {
			log.Printf("[CLEANUP][ERROR] อ่านโฟลเดอร์ %s ไม่ได้: %v", dir, err)
			continue
		}

		for _, e := range entries {
			if e.IsDir() || (!strings.HasSuffix(e.Name(), ".dump") && !strings.HasSuffix(e.Name(), ".enc")) {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			age := time.Since(info.ModTime())
			if age <= limit {
				continue
			}

			p := filepath.Join(dir, e.Name())
			if err := os.Remove(p); err != nil {
				log.Printf("[CLEANUP][ERROR] ลบ %s ไม่ได้: %v", p, err)
			} else {
				deleted++
				log.Printf("[CLEANUP] ลบ %s (อายุ %.0f วัน)", p, age.Hours()/24)
			}
		}
	}

	if deleted == 0 {
		log.Println("[CLEANUP] ไม่มีไฟล์เกินกำหนด")
	} else {
		log.Printf("[CLEANUP] ลบไปทั้งหมด %d ไฟล์", deleted)
	}
}
