package controller

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// RunBackup mount share แล้ว pg_dump ทุก database
func RunBackup(cfg *Config) {
	log.Println("=== เริ่ม backup ===")

	target, unmount, err := mountShare(cfg.DestBase, cfg.SharePass)
	if err != nil {
		log.Printf("[ERROR] เชื่อมต่อ share ไม่ได้: %v", err)
		return
	}
	defer unmount()

	ts := time.Now().Format("2006-01-02_15-04-05")

	for _, db := range cfg.DBNames {
		dbDir := filepath.Join(target, db)
		if err := os.MkdirAll(dbDir, 0o755); err != nil {
			log.Printf("[ERROR] สร้างโฟลเดอร์ %s ไม่ได้: %v", dbDir, err)
			continue
		}

		outFile := filepath.Join(dbDir, db+"_"+ts+".dump")
		cmd := exec.Command("pg_dump", "-Fc",
			"-h", cfg.PGHost,
			"-U", cfg.PGUser,
			"-d", db,
			"-f", outFile,
		)
		cmd.Env = append(os.Environ(), "PGPASSWORD="+cfg.PGPass)

		if result, err := cmd.CombinedOutput(); err != nil {
			log.Printf("[ERROR] backup %s ล้มเหลว: %v\n%s", db, err, result)
		} else {
			log.Printf("[OK] backup %s -> %s", db, outFile)
		}
	}

	cleanOldFiles(cfg, target)

	log.Println("=== backup เสร็จสิ้น ===")
}
