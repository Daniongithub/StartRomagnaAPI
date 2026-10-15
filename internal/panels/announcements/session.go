package announcements

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

const (
	sessionCookieName = "panel_session"
	sessionTTL        = 12 * time.Hour

	// cookie temporanei usati solo durante il flusso OIDC
	stateCookieName    = "panel_oidc_state"
	nonceCookieName    = "panel_oidc_nonce"
	verifierCookieName = "panel_oidc_verifier"
	oidcFlowTTL        = 5 * time.Minute
)

// session rappresenta un utente autenticato.
type session struct {
	Subject   string    `json:"sub"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Groups    []string  `json:"groups"`
	ExpiresAt time.Time `json:"-"`
}

// sessionStore è un semplice store in memoria. Per deployment multi-replica
// andrebbe sostituito con un backend condiviso (Redis, DB...), ma per un
// pannello interno single-instance va benissimo.
type sessionStore struct {
	mu   sync.RWMutex
	data map[string]session
}

func newSessionStore() *sessionStore {
	s := &sessionStore{data: make(map[string]session)}
	go s.janitor()
	return s
}

// janitor ripulisce periodicamente le sessioni scadute.
func (s *sessionStore) janitor() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		for id, sess := range s.data {
			if time.Now().After(sess.ExpiresAt) {
				delete(s.data, id)
			}
		}
		s.mu.Unlock()
	}
}

func (s *sessionStore) create(sess session) (id string, err error) {
	id, err = randomToken(32)
	if err != nil {
		return "", err
	}
	sess.ExpiresAt = time.Now().Add(sessionTTL)

	s.mu.Lock()
	s.data[id] = sess
	s.mu.Unlock()

	return id, nil
}

func (s *sessionStore) get(id string) (session, bool) {
	s.mu.RLock()
	sess, ok := s.data[id]
	s.mu.RUnlock()

	if !ok || time.Now().After(sess.ExpiresAt) {
		return session{}, false
	}
	return sess, true
}

func (s *sessionStore) delete(id string) {
	s.mu.Lock()
	delete(s.data, id)
	s.mu.Unlock()
}

// randomToken genera un token casuale URL-safe di n byte di entropia.
func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// isSecureRequest determina se impostare il flag Secure sui cookie,
// tenendo conto di eventuali reverse proxy TLS-terminating.
func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return r.Header.Get("X-Forwarded-Proto") == "https"
}

func setCookie(w http.ResponseWriter, r *http.Request, name, value string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     MountPath + "/",
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(ttl),
	})
}

func clearCookie(w http.ResponseWriter, r *http.Request, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     MountPath + "/",
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}
