package core

import (
	"log"
	"os/exec"
	"time"
)

func EnsureAria2Daemon() error {
	log.Println("[INFO] Memeriksa status Aria2c Daemon...")

	var stat map[string]interface{}
	err := Aria.Call("aria2.getGlobalStat", []interface{}{}, &stat)
	if err == nil {
		log.Println("[INFO] Aria2c Daemon sudah aktif dan siap menerima perintah.")
		return nil
	}

	log.Println("[INFO] Aria2c Daemon belum berjalan, memulai aria.sh...")
	cmd := exec.Command("bash", "aria.sh")
	if err := cmd.Start(); err != nil {
		return err
	}

	// Tunggu hingga Aria2 siap (maksimal 15 detik)
	for i := 0; i < 15; i++ {
		time.Sleep(1 * time.Second)
		err = Aria.Call("aria2.getGlobalStat", []interface{}{}, &stat)
		if err == nil {
			log.Println("[INFO] Aria2c Daemon berhasil dijalankan dan siap.")
			return nil
		}
	}

	return err
}
