package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Telegram struct {
	client   *http.Client
	endpoint string
	botID    string
	chatIDs  []int64
	hostName string
	stateDir string
	errors   chan string
	mu       sync.Mutex
	last     map[string]time.Time
}

func newTelegram(token string, chatIDs []int64, hostName, stateDir string) *Telegram {
	if token == "" {
		return nil
	}
	botID, _, _ := strings.Cut(token, ":")
	return &Telegram{
		client: &http.Client{
			Timeout: 5 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		endpoint: "https://api.telegram.org/bot" + token + "/sendMessage",
		botID:    botID,
		chatIDs:  chatIDs,
		hostName: hostName,
		stateDir: stateDir,
		errors:   make(chan string, 64),
		last:     make(map[string]time.Time),
	}
}

func (t *Telegram) deliver(message string) {
	text := t.formatMessage(message)
	t.sendToChats(func(ctx context.Context, chatID int64) error {
		return t.send(ctx, chatID, text)
	})
}

func (t *Telegram) formatMessage(message string) string {
	message = fmt.Sprintf("[StartRomagnaAPI | %s]\n%s", t.hostName, message)
	if chars := []rune(message); len(chars) > 4000 {
		return string(chars[:3999]) + "…"
	}
	return message
}

func (t *Telegram) sendToChats(send func(context.Context, int64) error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	for i, chatID := range t.chatIDs {
		wg.Go(func() {
			if err := send(ctx, chatID); err != nil {
				log.Printf("Failed to send Telegram notification to recipient %d: %v", i+1, err)
			}
		})
	}
	wg.Wait()
}

func (t *Telegram) send(ctx context.Context, chatID int64, message string) error {
	form := url.Values{
		"chat_id": {strconv.FormatInt(chatID, 10)},
		"text":    {message},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, t.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("Invalid Telegram request")
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := t.client.Do(request)
	if err != nil {
		// The HTTP error can contain the bot token in the request URL.
		return fmt.Errorf("Telegram request failed")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram returned HTTP %d", response.StatusCode)
	}
	var result struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(&result); err != nil {
		return fmt.Errorf("Invalid response from Telegram")
	}
	if !result.OK {
		return fmt.Errorf("Telegram rejected the message")
	}
	return nil
}
