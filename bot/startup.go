package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"
)

var StartTime time.Time

func LaunchHealthServer(port string) {
	StartTime = time.Now()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		uptime := time.Since(StartTime).Round(time.Second).String()
		fmt.Fprintf(w, `{"status":"running","bot":"%s","version":"%s","uptime":"%s"}`, BotName, Version, uptime)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("[INFO] Health check server aktif di port %s (Koyeb Ready)...", port)
	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Printf("[WARN] Health server error: %v", err)
		}
	}()
}

func LaunchAria2c() error {
	log.Println("[INFO] Memeriksa status Aria2c Daemon...")

	client := &http.Client{Timeout: 3 * time.Second}
	reqBody, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "1",
		"method":  "aria2.getGlobalStat",
		"params":  []interface{}{},
	})

	resp, err := client.Post("http://127.0.0.1:6800/jsonrpc", "application/json", bytes.NewBuffer(reqBody))
	if err == nil {
		resp.Body.Close()
		log.Println("[INFO] Aria2c Daemon sudah aktif.")
		return nil
	}

	log.Println("[INFO] Aria2c belum berjalan, menjalankan aria.sh...")
	cmd := exec.Command("bash", "aria.sh")
	if err := cmd.Start(); err != nil {
		return err
	}

	for i := 0; i < 15; i++ {
		time.Sleep(1 * time.Second)
		resp, err := client.Post("http://127.0.0.1:6800/jsonrpc", "application/json", bytes.NewBuffer(reqBody))
		if err == nil {
			resp.Body.Close()
			log.Println("[INFO] Aria2c Daemon berhasil dijalankan dan siap.")
			return nil
		}
	}
	return fmt.Errorf("timeout menunggu Aria2c siap")
}
