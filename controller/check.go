package controller

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

// CheckConnections ตรวจสอบตอน start — เชื่อมต่อ db ทุกตัว + share ปลายทาง
// คืน error ถ้ามีตัวไหนเชื่อมต่อไม่ได้
func CheckConnections(cfg *Config) error {
	log.Println("=== ตรวจสอบการเชื่อมต่อ ===")

	var failed []string

	for _, db := range cfg.DBNames {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=5",
			pqVal(cfg.PGHost), pqVal(cfg.PGUser), pqVal(cfg.PGPass), pqVal(db))
		conn, err := sql.Open("postgres", dsn)
		if err == nil {
			err = conn.Ping()
			conn.Close()
		}
		if err != nil {
			failed = append(failed, "db "+db)
			log.Printf("[CHECK][ERROR] database %s เชื่อมต่อไม่ได้: %v", db, err)
		} else {
			log.Printf("[CHECK] database %s เชื่อมต่อสำเร็จ", db)
		}
	}

	_, unmount, err := mountShare(cfg.DestBase, cfg.SharePass)
	if err != nil {
		failed = append(failed, "share "+cfg.DestBase)
		log.Printf("[CHECK][ERROR] เครื่องปลายทาง %s เชื่อมต่อไม่ได้: %v", cfg.DestBase, err)
	} else {
		unmount()
		log.Printf("[CHECK] เครื่องปลายทาง %s เชื่อมต่อสำเร็จ", cfg.DestBase)
	}

	log.Println("=== ตรวจสอบเสร็จสิ้น ===")

	if len(failed) > 0 {
		return fmt.Errorf("เชื่อมต่อไม่ได้: %s — หยุดการทำงาน", strings.Join(failed, ", "))
	}
	return nil
}

func pqVal(s string) string {
	return "'" + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `'`, `\'`) + "'"
}
