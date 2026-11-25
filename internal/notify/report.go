package notify

import (
	"log"
	"strings"
	"sync/atomic"
	"time"
)

const repeatInterval = 30 * time.Minute

var telegram atomic.Pointer[Telegram]

func ConfigureTelegram(token string, chatIDs []int64, hostName, stateDir string) {
	sender := newTelegram(token, chatIDs, hostName, stateDir)
	telegram.Store(sender)
	if sender != nil {
		go sender.deliverErrors()
	}
}

func Startup() {
	if sender := telegram.Load(); sender != nil {
		go sender.welcome()
	}
}

// ReportError logs diagnostic details and queues the operation name for Telegram.
// The operation and summary must contain only text safe to share with recipients.
func ReportError(operation string, err error) {
	ReportErrorWithSummary(operation, err, operation)
}

func ReportErrorWithSummary(operation string, err error, summary string) {
	log.Printf("%s: %v", operation, err)
	telegram.Load().queueError(operation, summary)
}

// Fatal attempts delivery before log.Fatal exits the process.
func Fatal(summary string, err error) {
	if sender := telegram.Load(); sender != nil {
		sender.deliver("Fatal error: " + summary + "\nSee the application logs for details.")
	}
	if err != nil {
		log.Fatalf("%s: %v", summary, err)
	}
	log.Fatal(summary)
}

func (t *Telegram) queueError(operation, summary string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	if now.Sub(t.last[operation]) < repeatInterval {
		return
	}
	select {
	case t.errors <- summary:
		t.last[operation] = now
	default:
		// The diagnostic is still logged if the notification queue is full.
	}
}

func (t *Telegram) deliverErrors() {
	for first := range t.errors {
		// Give concurrent failures a chance to arrive before sending one message.
		time.Sleep(2 * time.Second)
		batch := []string{first}
		for len(t.errors) > 0 && len(batch) < 10 {
			batch = append(batch, <-t.errors)
		}
		t.deliver("Operational errors\n\n" + strings.Join(batch, "\n\n") + "\n\nSee the application logs for details.")
	}
}
