package main

import (
	"log"

	"backup_data/controller"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load() // ใช้ .env ถ้ามี ไม่มีก็อ่านจาก env จริง (docker --env-file)

	cfg, err := controller.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	if err := controller.CheckConnections(cfg); err != nil {
		log.Fatal(err)
	}
	controller.Start(cfg)
}
