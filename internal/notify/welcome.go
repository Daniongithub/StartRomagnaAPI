package notify

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func (t *Telegram) welcome() {
	if err := os.MkdirAll(t.stateDir, 0700); err != nil {
		log.Printf("Failed to open Telegram welcome state directory: %v", err)
		return
	}
	message := t.formatMessage("Telegram notifications are enabled for this host.")
	t.sendToChats(func(ctx context.Context, chatID int64) error {
		return t.welcomeChat(ctx, chatID, message)
	})
}

func (t *Telegram) welcomeChat(ctx context.Context, chatID int64, message string) error {
	association := fmt.Sprintf("%s\x00%s\x00%d", t.botID, t.hostName, chatID)
	digest := sha256.Sum256([]byte(association))
	path := filepath.Join(t.stateDir, fmt.Sprintf("welcome-%x", digest))

	// Reserve the association before sending so overlapping instances send only once.
	marker, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Reserve Telegram welcome: %w", err)
	}
	if err := marker.Close(); err != nil {
		return errors.Join(fmt.Errorf("Close Telegram welcome marker: %w", err), removeWelcomeMarker(path))
	}
	if err := t.send(ctx, chatID, message); err != nil {
		// A failed delivery should be retried at the next successful startup.
		return errors.Join(err, removeWelcomeMarker(path))
	}
	return nil
}

func removeWelcomeMarker(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("Remove Telegram welcome marker: %w", err)
	}
	return nil
}
