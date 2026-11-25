package config

import (
	"fmt"
	"strconv"
	"strings"
)

type TelegramConfig struct {
	BotToken string
	ChatIDs  []int64
	HostName string
	StateDir string
}

func loadTelegram(getenv func(string) string, hostname func() (string, error)) (TelegramConfig, error) {
	value := strings.TrimSpace(getenv("TELEGRAM_ENABLED"))
	if value == "" {
		return TelegramConfig{}, nil
	}
	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return TelegramConfig{}, fmt.Errorf("Invalid boolean value for TELEGRAM_ENABLED")
	}
	if !enabled {
		return TelegramConfig{}, nil
	}

	token := strings.TrimSpace(getenv("TELEGRAM_BOT_TOKEN"))
	id, secret, ok := strings.Cut(token, ":")
	const tokenCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	if _, err := strconv.ParseUint(id, 10, 64); err != nil || !ok || secret == "" || strings.Trim(secret, tokenCharacters) != "" {
		return TelegramConfig{}, fmt.Errorf("Invalid or missing TELEGRAM_BOT_TOKEN")
	}

	var chatIDs []int64
	seen := make(map[int64]bool)
	for _, entry := range strings.Split(getenv("TELEGRAM_CHAT_IDS"), ",") {
		chatID, err := strconv.ParseInt(strings.TrimSpace(entry), 10, 64)
		if err != nil || chatID == 0 {
			return TelegramConfig{}, fmt.Errorf("TELEGRAM_CHAT_IDS must contain comma-separated, nonzero numeric chat IDs")
		}
		if !seen[chatID] {
			seen[chatID] = true
			chatIDs = append(chatIDs, chatID)
		}
	}

	host := strings.TrimSpace(getenv("TELEGRAM_HOST_NAME"))
	if host == "" {
		host, err = hostname()
		if err != nil || strings.TrimSpace(host) == "" {
			return TelegramConfig{}, fmt.Errorf("Set TELEGRAM_HOST_NAME because the system hostname is unavailable")
		}
	}
	stateDir := strings.TrimSpace(getenv("TELEGRAM_STATE_DIR"))
	if stateDir == "" {
		stateDir = ".telegram-state"
	}
	return TelegramConfig{BotToken: token, ChatIDs: chatIDs, HostName: host, StateDir: stateDir}, nil
}
