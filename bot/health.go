package bot

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

var StartTime time.Time

func StartHealthServer(port string) {
	StartTime = time.Now()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		uptime := time.Since(StartTime).Round(time.Second).String()
		fmt.Fprintf(w, `{"status":"running","bot":"go-mirror-bot","uptime":"%s"}`, uptime)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("[INFO] Health check web server aktif di port %s (siap untuk Koyeb)...", port)
	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Printf("[WARN] Health server error: %v", err)
		}
	}()
}
