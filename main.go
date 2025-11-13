package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	toast "github.com/go-toast/toast"
)

// config.json
type Config struct {
	Port  string `json:"port"`
	AppID string `json:"appid"`
	Title string `json:"title"`
	Sub   string `json:"sub"`
	Btn   string `json:"btn"`
	URL   string `json:"url"`
	Icon  string `json:"icon"`
}

func main() {

	// variables from config.json
	cfg := loadConfig()
	addr := ":" + cfg.Port
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("Unable to get exe path: %v", err)
	}
	baseDir := filepath.Dir(exe)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handle(w, r, cfg, baseDir)
	})

	log.Printf("Listening on %s\n", addr)
	log.Printf("Try: http://localhost%v/?msg=hello", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func handle(w http.ResponseWriter, r *http.Request, cfg Config, baseDir string) {
	msg := getParamFromRequest(r, "msg")
	subParam := getParamFromRequest(r, "sub")
	icon := getParamFromRequest(r, "icon")

	// parameter for msg is required
	if msg == "" {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "OK (no msg)")
		return
	}

	// parameter for icon
	var iconPath string
	if icon == "" {
		iconPath = filepath.Join(baseDir, "img", cfg.Icon)
		log.Printf("Using default icon: %s", iconPath)
	} else {
		iconPath = filepath.Join(baseDir, "img", icon)
		log.Printf("Using custom icon: %s", iconPath)
	}

	// default parameter for sub
	sub := subParam
	if subParam == "" {
		sub = cfg.Sub
	}

	// replace {msg} in title from config.json
	title := strings.ReplaceAll(cfg.Title, "{msg}", msg)
	log.Printf("Received: %s", msg)

	// send the Windows Notification
	notification := toast.Notification{
		AppID:   cfg.AppID,
		Title:   title,
		Message: sub,
		Icon:    iconPath,
		Actions: []toast.Action{
			{
				Type:      "protocol",
				Label:     cfg.Btn,
				Arguments: cfg.URL,
			},
		},
	}

	if err := notification.Push(); err != nil {
		log.Printf("toast error: %v", err)
		http.Error(w, "toast error: "+err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "OK: %s\n", msg)
}

// fetches parameter from web request
func getParamFromRequest(r *http.Request, param string) string {
	// ?msg=foo for example
	q := r.URL.Query().Get(param)
	if q != "" {
		return q
	}
	return ""
}

// fetches parameters from config.json
func loadConfig() Config {
	// fallback
	cfg := Config{
		Port:  "8080",
		AppID: "WebToast",
		Title: "{msg}",
		Sub:   "",
		URL:   "https://orthexgroup.com/404",
	}

	// current path where .exe is running
	exe, err := os.Executable()
	if err != nil {
		log.Printf("unable to get exe path: %v", err)
		return cfg
	}

	// read config.json file from root folder
	configPath := filepath.Join(filepath.Dir(exe), "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("config.json missing, using defaults")
		writeDefaultConfig(configPath, cfg)
		return cfg
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("config parse error: %v", err)
	}

	return cfg
}

// creates config.json using fallback values if missing
func writeDefaultConfig(path string, cfg Config) {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	_ = os.WriteFile(path, b, 0644)
	log.Printf("created default config.json")
}
