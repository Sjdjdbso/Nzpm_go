package core

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

type AriaClient struct {
	RPCUrl     string
	HTTPClient *http.Client
}

var Aria *AriaClient

func InitAriaClient(rpcUrl string) {
	if rpcUrl == "" {
		rpcUrl = "http://127.0.0.1:6800/jsonrpc"
	}
	Aria = &AriaClient{
		RPCUrl: rpcUrl,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type FileItem struct {
	Index           string `json:"index"`
	Path            string `json:"path"`
	Length          string `json:"length"`
	CompletedLength string `json:"completedLength"`
	Selected        string `json:"selected"`
}

type AriaStatus struct {
	GID             string     `json:"gid"`
	Status          string     `json:"status"` // active, waiting, paused, error, complete, removed
	TotalLength     string     `json:"totalLength"`
	CompletedLength string     `json:"completedLength"`
	DownloadSpeed   string     `json:"downloadSpeed"`
	Files           []FileItem `json:"files"`
	ErrorMessage    string     `json:"errorMessage"`
	FollowedBy      []string   `json:"followedBy"` // untuk transisi metadata magnet ke download utama
}

func (c *AriaClient) Call(method string, params []interface{}, result interface{}) error {
	reqBody := RPCRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
		Method:  method,
		Params:  params,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Post(c.RPCUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var rpcResp RPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return err
	}

	if rpcResp.Error != nil {
		return fmt.Errorf("aria2 error [%d]: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	if result != nil && len(rpcResp.Result) > 0 {
		return json.Unmarshal(rpcResp.Result, result)
	}

	return nil
}

func (c *AriaClient) AddURI(uri string, dir string, filename string) (string, error) {
	options := map[string]interface{}{}
	if dir != "" {
		absDir, err := filepath.Abs(dir)
		if err == nil {
			dir = absDir
		}
		options["dir"] = dir
	}
	if filename != "" {
		options["out"] = filename
	}

	params := []interface{}{
		[]string{uri},
		options,
	}

	var gid string
	err := c.Call("aria2.addUri", params, &gid)
	return gid, err
}

// AddTorrent menambahkan file .torrent dalam bentuk byte array (base64)
func (c *AriaClient) AddTorrent(torrentBytes []byte, dir string, filename string) (string, error) {
	b64Torrent := base64.StdEncoding.EncodeToString(torrentBytes)
	options := map[string]interface{}{}
	if dir != "" {
		absDir, err := filepath.Abs(dir)
		if err == nil {
			dir = absDir
		}
		options["dir"] = dir
	}
	if filename != "" {
		options["out"] = filename
	}

	params := []interface{}{
		b64Torrent,
		[]string{},
		options,
	}

	var gid string
	err := c.Call("aria2.addTorrent", params, &gid)
	return gid, err
}

func (c *AriaClient) TellStatus(gid string) (*AriaStatus, error) {
	params := []interface{}{
		gid,
		[]string{"gid", "status", "totalLength", "completedLength", "downloadSpeed", "files", "errorMessage", "followedBy"},
	}

	var status AriaStatus
	err := c.Call("aria2.tellStatus", params, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (c *AriaClient) Remove(gid string) error {
	var res string
	return c.Call("aria2.forceRemove", []interface{}{gid}, &res)
}

func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func FormatSpeed(bytesPerSec int64) string {
	return fmt.Sprintf("%s/s", FormatBytes(bytesPerSec))
}

func CalculateETA(total, completed, speed int64) string {
	if speed <= 0 || completed >= total {
		return "0s"
	}
	remainingBytes := total - completed
	seconds := remainingBytes / speed

	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if seconds < 3600 {
		return fmt.Sprintf("%dm %ds", seconds/60, seconds%60)
	}
	return fmt.Sprintf("%dh %dm", seconds/3600, (seconds%3600)/60)
}

func StringToInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}
