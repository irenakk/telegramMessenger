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

// GetChatIDByUsername GET /users/{username}/chatid -> { "chat_id": 12345 }
func (c *UserServiceClient) GetChatIDByUsername(ctx context.Context, username string) (int64, error) {
	url := fmt.Sprintf("%s/users/%s/chatid", c.baseURL, username)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return 0, nil // no chat id
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("user-service returned %d: %s", resp.StatusCode, string(b))
	}
	var out struct {
		ChatID int64 `json:"chat_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, err
	}
	return out.ChatID, nil
}
