package controller

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RunBackup mount share แล้ว pg_dump ทุก database จบด้วยส่ง alert สรุปผล
func RunBackup(cfg *Config) {
	log.Println("=== เริ่ม backup ===")

	ts := time.Now().Format("2006-01-02_15-04-05")

	target, unmount, err := mountShare(cfg.DestBase, cfg.SharePass)
	if err != nil {
		log.Printf("[ERROR] เชื่อมต่อ share ไม่ได้: %v", err)
		sendAlert(cfg, false, ts, []string{"เชื่อมต่อ share ปลายทางไม่ได้: " + cfg.DestBase})
		return
	}
	defer unmount()

	var results []string
	fail := false

	for _, db := range cfg.DBNames {
		dbDir := filepath.Join(target, db)
		if err := os.MkdirAll(dbDir, 0o755); err != nil {
			log.Printf("[ERROR] สร้างโฟลเดอร์ %s ไม่ได้: %v", dbDir, err)
			results = append(results, "FAIL "+db+" (สร้างโฟลเดอร์ไม่ได้)")
			fail = true
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
			results = append(results, "FAIL "+db+": "+firstLine(string(result)))
			fail = true
			continue
		}
		log.Printf("[OK] backup %s -> %s", db, outFile)

		if err := encryptFile(outFile, cfg.EncPass); err != nil {
			log.Printf("[ERROR] encrypt %s ล้มเหลว: %v (ไฟล์ .dump ยังอยู่)", outFile, err)
			results = append(results, "FAIL "+db+": encrypt ล้มเหลว")
			fail = true
			continue
		}
		results = append(results, "OK   "+db)
	}

	cleanOldFiles(cfg, target)

	sendAlert(cfg, !fail, ts, results)

	log.Println("=== backup เสร็จสิ้น ===")
}

// encryptFile เข้ารหัสไฟล์ด้วย openssl แล้วลบไฟล์ต้นฉบับ เหลือ <path>.enc
func encryptFile(path, pass string) error {
	enc := path + ".enc"
	cmd := exec.Command("openssl", "enc", "-aes-256-cbc", "-pbkdf2",
		"-in", path, "-out", enc, "-pass", "env:BK_ENC")
	cmd.Env = append(os.Environ(), "BK_ENC="+pass)

	if out, err := cmd.CombinedOutput(); err != nil {
		os.Remove(enc) // ลบไฟล์เสียที่อาจเหลือ
		return fmt.Errorf("openssl: %v\n%s", err, out)
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	log.Printf("[OK] encrypt -> %s", enc)
	return nil
}

// firstLine ตัดเอาแค่บรรทัดแรก ไม่เกิน 150 ตัว
func firstLine(s string) string {
	s, _, _ = strings.Cut(s, "\n")
	if len(s) > 150 {
		s = s[:150] + "..."
	}
	return strings.TrimSpace(s)
}
