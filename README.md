# goBackupData

โปรแกรม backup PostgreSQL อัตโนมัติทุกวัน → เก็บบน network share → เข้ารหัสไฟล์ด้วย AES-256 → ลบไฟล์เก่าตามอายุที่กำหนด

## ไฟล์ที่ได้หลัง backup

```
\\192.168.2.202\dbbackup\
├── manager\
│   ├── manager_2026-11-19_00-00-09.dump.enc
│   └── manager_2026-11-20_00-00-09.dump.enc
└── check_ambulance\
    └── check_ambulance_2026-11-20_00-00-09.dump.enc
```

ไฟล์ `.enc` = ไฟล์ที่ถูกเข้ารหัสแล้ว (เปิดอ่านตรงๆ ไม่ได้) ต้องถอดรหัสก่อนถึงจะ restore เข้า database ได้

---

## วิธี Restore (ทีละขั้น)

ไฟล์ backup ผ่าน 2 ชั้น — restore ต้องทำย้อนกลับ 2 ขั้น:

```
ไฟล์ .dump.enc --[ถอดรหัส openssl]--> .dump --[pg_restore]--> database
```

### เตรียมของก่อน

ต้องมี 2 โปรแกรมในเครื่องที่จะ restore:

| โปรแกรม | เอาไว้ทำอะไร | เช็คว่ามีไหม |
|---------|------------|-------------|
| `openssl` | ถอดรหัสไฟล์ .enc | พิมพ์ `openssl version` |
| `pg_restore` | เทข้อมูลเข้า database | พิมพ์ `pg_restore --version` |

- **Windows:** openssl มีใน Git Bash อยู่แล้ว — แต่ `pg_restore` ต้องลง PostgreSQL เอง ถ้าไม่มีให้ใช้วิธี Docker ด้านล่าง
- **Ubuntu/Docker:** container `pg-backup` มีครบทั้งคู่ แนะนำใช้อันนี้
- เตรียมรหัส `ENC_PASS` (ดูใน `.env`) และรหัสของ user `it_admin` (ค่า `DB_HOST_PASS_BACKUP`)

### ขั้นที่ 1 — เอาไฟล์มาไว้ที่เครื่อง

copy ไฟล์ `.enc` ที่จะ restore จาก share `\\192.168.2.202\dbbackup\<ชื่อdb>\` มาไว้ที่เครื่อง เช่น:

```
C:\Users\spurs\OneDrive\Desktop\backup\check_ambulance_2026-11-20_11-28-06.dump.enc
```

### ขั้นที่ 2 — ถอดรหัสไฟล์ (decrypt)

**บน Windows:** เปิด Git Bash แล้วพิมพ์ (ปรับชื่อไฟล์+รหัสให้ตรง)

```bash
cd /c/Users/spurs/OneDrive/Desktop/backup

openssl enc -d -aes-256-cbc -pbkdf2 -in check_ambulance_2026-11-20_12-46-10.dump.enc -out check.dump -pass pass:รหัสผ่าน

openssl enc -d -aes-256-cbc -pbkdf2 \
  -in  ไฟล์.enc        # -in  = ไฟล์เข้า: ไฟล์ล็อคที่จะถอด
  -out ไฟล์.dump       # -out = ไฟล์ออก: ชื่อผลลัพธ์หลังถอด (ตั้งเองได้)
  -pass pass:รหัส      # -pass = รหัสปลดล็อค (ค่า ENC_PASS ใน .env)
```

**บน PowerShell** เขียนบรรทัดเดียว:

```powershell
openssl enc -d -aes-256-cbc -pbkdf2 -in check_ambulance_2026-11-20_11-28-06.dump.enc -out check.dump -pass pass:ใส่รหัสENC_PASSตรงนี้
```

**บน Ubuntu:**

```bash
openssl enc -d -aes-256-cbc -pbkdf2 -in ไฟล์.enc -out ไฟล์.dump -pass pass:ใส่รหัสENC_PASSตรงนี้
```

ผลลัพธ์:

- ถ้า**รหัสถูก** → ได้ไฟล์ `.dump` ใหม่โผล่มา (เช็คด้วย `ls`)
- ถ้า**รหัสผิด** → ขึ้น `bad decrypt` แล้วไม่มีทางเปิดได้เลย (ลืมรหัส = ไฟล์ทิ้ง)

> ถ้ารหัสมีตัวอักษรพิเศษ เช่น `!` `$` ให้ครอบด้วย single quote เช่น `-pass 'pass:!abc123'`

### ขั้นที่ 3 — เทข้อมูลเข้า database (restore)

สร้าง database ว่างขึ้นมาก่อน แล้วเทไฟล์เข้า:

```bash
# สร้าง db ใหม่ (จะถามรหัสของ it_admin)
createdb -h 192.168.2.23 -U it_admin check_restore

# เทข้อมูลจากไฟล์เข้า db
pg_restore -h 192.168.2.23 -U it_admin -d check_restore check.dump
```

จบ — เปิด pgAdmin ดู database `check_restore` จะเห็นข้อมูลครบเหมือนตอน backup

**กรณี restore ทับ db เดิม** (ไม่ใช่ db ใหม่) ใส่ `--clean`:

```bash
pg_restore -h 192.168.2.23 -U it_admin -d check_ambulance --clean check.dump
```

> `--clean` = ลบข้อมูลเก่าใน db นั้นก่อนแล้วค่อยเทของใหม่เข้า — ระวังใช้กับ db จริงที่ยังใช้งานอยู่

### ทางลัด — ทำทุกอย่างใน Docker container (แนะนำ)

container `pg-backup` มี `openssl` + `pg_restore` ครบอยู่แล้ว:

```bash
# บน Ubuntu server — โยนไฟล์เข้า container
docker cp ~/backup/check_ambulance_2026-11-20_11-28-06.dump.enc pg-backup:/tmp/

# เข้าไปใน container
docker exec -it pg-backup sh

# ใน container
cd /tmp
openssl enc -d -aes-256-cbc -pbkdf2 -in check_ambulance_2026-11-20_11-28-06.dump.enc -out check.dump -pass pass:รหัสENC_PASS
createdb -h 192.168.2.23 -U it_admin check_restore
pg_restore -h 192.168.2.23 -U it_admin -d check_restore check.dump
```

---

## คำสั่งใช้งานโปรแกรม

```bash
# รัน local (ต้องมี pg_dump + openssl ในเครื่อง)
go run .

# รันแบบ Docker
docker compose up -d --build

# ดู log
docker compose logs -f
```

## ตัวแปรใน .env

| ตัวแปร | ความหมาย |
|--------|----------|
| `DB_NAME_BACKUP` | ชื่อ database ที่จะ backup คั่นด้วย `,` |
| `DB_PATH_DEST` | path ปลายทาง network share |
| `DB_PASS_BKUSER` | รหัส user ของ share (`ADMIN\bkuser`) |
| `DB_HOST_BACKUP` / `DB_HOST_USER_BACKUP` / `DB_HOST_PASS_BACKUP` | host, user, รหัส PostgreSQL |
| `TIME_BACKUP` | เวลา backup ทุกวัน `HH:MM` |
| `CHECK_FILE` | เก็บไฟล์ย้อนหลังกี่วัน `<เลข>d` เช่น `7d` |
| `ENC_PASS` | รหัสเข้ารหัสไฟล์ — **เก็บไว้ให้ดี ห้ามลืม** |
| `MOPH_ALERT_METHOD` / `MOPH_ALERT_URL` | method + endpoint ของ MOPH Alert API |
| `MOPH_ALERT_CLIENT_ID` / `MOPH_ALERT_SECRET_KEY` | credential ของ MOPH Alert |
| `CID_ALERT` | CID ผู้รับ alert คั่นด้วย `,` |
