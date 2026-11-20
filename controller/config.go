package controller

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DBNames   []string
	DestBase  string
	SharePass string
	PGHost    string
	PGUser    string
	PGPass    string
	TimeStr   string
	Hour      int
	Minute    int
	RetainDays int
	EncPass   string
}

func LoadConfig() (*Config, error) {
	keys := []string{
		"DB_NAME_BACKUP", "DB_PATH_DEST", "DB_PASS_BKUSER",
		"DB_HOST_BACKUP", "DB_HOST_USER_BACKUP", "DB_HOST_PASS_BACKUP",
		"TIME_BACKUP", "CHECK_FILE", "ENC_PASS",
	}
	env := map[string]string{}
	var missing []string
	for _, k := range keys {
		env[k] = strings.TrimSpace(os.Getenv(k))
		if env[k] == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("env ไม่ครบ กรุณาตั้งค่า: %s", strings.Join(missing, ", "))
	}

	dbNames := splitTrim(env["DB_NAME_BACKUP"])
	if len(dbNames) == 0 {
		return nil, fmt.Errorf("DB_NAME_BACKUP ไม่มีชื่อ database")
	}

	h, m, err := parseHHMM(env["TIME_BACKUP"])
	if err != nil {
		return nil, fmt.Errorf("TIME_BACKUP '%s' ไม่ถูกต้อง ต้องเป็น HH:MM: %v", env["TIME_BACKUP"], err)
	}

	retain, err := parseRetention(env["CHECK_FILE"])
	if err != nil {
		return nil, err
	}

	return &Config{
		DBNames:   dbNames,
		DestBase:  env["DB_PATH_DEST"],
		SharePass: env["DB_PASS_BKUSER"],
		PGHost:    env["DB_HOST_BACKUP"],
		PGUser:    env["DB_HOST_USER_BACKUP"],
		PGPass:    env["DB_HOST_PASS_BACKUP"],
		TimeStr:   env["TIME_BACKUP"],
		Hour:      h,
		Minute:    m,
		RetainDays: retain,
		EncPass:    env["ENC_PASS"],
	}, nil
}

// parseRetention แปลง "7d" -> 7 (วัน) ต้องเป็นตัวเลข >0 ตามด้วย d เท่านั้น
func parseRetention(s string) (int, error) {
	bad := fmt.Errorf("CHECK_FILE '%s' รูปแบบไม่ถูกต้อง ต้องเป็นตัวเลขตามด้วย d เช่น 7d, 30d", s)
	if !strings.HasSuffix(s, "d") {
		return 0, bad
	}
	n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
	if err != nil || n <= 0 {
		return 0, bad
	}
	return n, nil
}

func splitTrim(s string) []string {
	var out []string
	for _, v := range strings.Split(s, ",") {
		if t := strings.TrimSpace(v); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func parseHHMM(s string) (int, int, error) {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return 0, 0, err
	}
	return h, m, nil
}
