package controller

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// username ของ network share (domain\user)
const shareUser = `ADMIN\bkuser`

// mount point ใน container (linux)
const mountPoint = "/mnt/backup"

// mountShare mount network share แล้วคืน (local path, ฟังก์ชัน unmount, error)
func mountShare(dest, pass string) (string, func(), error) {
	if runtime.GOOS == "windows" {
		exec.Command("net", "use", dest, "/delete", "/y").Run()
		out, err := exec.Command("net", "use", dest, "/user:"+shareUser, pass).CombinedOutput()
		if err != nil {
			return "", nil, fmt.Errorf("net use: %v\n%s", err, out)
		}
		return dest, func() {
			exec.Command("net", "use", dest, "/delete", "/y").Run()
		}, nil
	}

	// linux: mount cifs ที่ mountPoint
	os.MkdirAll(mountPoint, 0o755)
	src := "//" + strings.ReplaceAll(strings.TrimPrefix(dest, `\\`), `\`, "/")
	out, err := exec.Command("mount", "-t", "cifs", src, mountPoint, "-o", cifsOpts(pass)).CombinedOutput()
	if err != nil {
		return "", nil, fmt.Errorf("mount cifs %s: %v\n%s", src, err, out)
	}
	return mountPoint, func() {
		exec.Command("umount", "-l", mountPoint).Run()
	}, nil
}

func cifsOpts(pass string) string {
	u := shareUser
	opts := "user=" + u
	if i := strings.LastIndex(u, `\`); i >= 0 {
		opts = "user=" + u[i+1:] + ",domain=" + u[:i]
	}
	return opts + ",password=" + pass + ",iocharset=utf8"
}
