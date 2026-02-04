package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type UserServiceClient struct {
	baseURL string
	client  *http.Client
}

func NewUserServiceClient(baseURL string) *UserServiceClient {
	return &UserServiceClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// LinkTelegramAccount posts { "username": "...", "chat_id": 12345 }
func (c *UserServiceClient) LinkTelegramAccount(ctx context.Context, username string, chatID int64, tgnickname string) error {
	reqBody := map[string]interface{}{
		"username":    username,
		"chat_id":     chatID,
		"tg_nickname": tgnickname,
	}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/link-telegram", strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("user-service returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *UserServiceClient) GetChatIDByUsername(ctx context.Context, username string) (int64, error) {
	reqBody := map[string]interface{}{
		"username": username,
	}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/get-chatid", strings.NewReader(string(b)))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("user-service returned %d: %s", resp.StatusCode, string(body))
	}
	var out struct {
		ChatID int64 `json:"chat_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, err
	}
	return out.ChatID, nil
}
