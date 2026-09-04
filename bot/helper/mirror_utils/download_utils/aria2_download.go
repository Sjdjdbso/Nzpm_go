package download_utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
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

type FileItem struct {
	Index           string `json:"index"`
	Path            string `json:"path"`
	Length          string `json:"length"`
	CompletedLength string `json:"completedLength"`
}

type AriaStatus struct {
	GID             string     `json:"gid"`
	Status          string     `json:"status"` // active, waiting, paused, error, complete, removed
	TotalLength     string     `json:"totalLength"`
	CompletedLength string     `json:"completedLength"`
	DownloadSpeed   string     `json:"downloadSpeed"`
	Files           []FileItem `json:"files"`
	ErrorMessage    string     `json:"errorMessage"`
	FollowedBy      []string   `json:"followedBy"`
}

func (c *AriaClient) Call(method string, params []interface{}, result interface{}) error {
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      fmt.Sprintf("%d", time.Now().UnixNano()),
		"method":  method,
		"params":  params,
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

	var rpcResp struct {
		Result json.RawMessage `json:"result,omitempty"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

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

func (c *AriaClient) AddTorrent(torrentBytes []byte, dir string, filename string) (string, error) {
	b64 := base64.StdEncoding.EncodeToString(torrentBytes)
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
		b64,
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
