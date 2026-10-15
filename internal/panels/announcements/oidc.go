package announcements

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// idTokenClaims è il sottoinsieme di claim dell'ID token che ci interessano.
// Il claim "groups" dipende dal mapping configurato lato Authentik (va
// esposto sullo scope "groups" nel provider).
type idTokenClaims struct {
	Subject           string   `json:"sub"`
	Email             string   `json:"email"`
	Name              string   `json:"name"`
	PreferredUsername string   `json:"preferred_username"`
	Groups            []string `json:"groups"`
}

// redirectURL calcola la callback OIDC per il prefisso con cui è montato il pannello.
func redirectURL(r *http.Request) string {
	scheme := "http"
	if isSecureRequest(r) {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s%s/auth/callback", scheme, r.Host, MountPath)
}

func (p *Panel) oauth2Config(r *http.Request) oauth2.Config {
	cfg := p.oauth2Base
	cfg.RedirectURL = redirectURL(r)
	return cfg
}

// handleAuthStart avvia il flusso Authorization Code + PKCE verso il
// provider OIDC. È l'azione dietro al bottone "Accedi con Authentik" di
// login.html.
func (p *Panel) handleAuthStart(w http.ResponseWriter, r *http.Request) {
	state, err := randomToken(16)
	if err != nil {
		log.Printf("announcements: generazione state OIDC fallita: %v", err)
		http.Error(w, "errore interno", http.StatusInternalServerError)
		return
	}
	nonce, err := randomToken(16)
	if err != nil {
		log.Printf("announcements: generazione nonce OIDC fallita: %v", err)
		http.Error(w, "errore interno", http.StatusInternalServerError)
		return
	}
	verifier := oauth2.GenerateVerifier()

	setCookie(w, r, stateCookieName, state, oidcFlowTTL)
	setCookie(w, r, nonceCookieName, nonce, oidcFlowTTL)
	setCookie(w, r, verifierCookieName, verifier, oidcFlowTTL)

	oauthCfg := p.oauth2Config(r)
	authURL := oauthCfg.AuthCodeURL(
		state,
		oauth2.AccessTypeOnline,
		oidc.Nonce(nonce),
		oauth2.S256ChallengeOption(verifier),
	)

	http.Redirect(w, r, authURL, http.StatusFound)
}

// failLogin riporta l'utente sulla pagina di login mostrando un messaggio
// d'errore leggibile, invece di una nuda pagina di errore HTTP.
func failLogin(w http.ResponseWriter, r *http.Request, message string) {
	http.Redirect(w, r, MountPath+"/login?error="+url.QueryEscape(message), http.StatusFound)
}

// handleCallback riceve il redirect dal provider, scambia il code per i
// token, verifica l'ID token e, se l'utente appartiene a un gruppo
// autorizzato, crea la sessione locale.
func (p *Panel) handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil || stateCookie.Value == "" || stateCookie.Value != r.URL.Query().Get("state") {
		failLogin(w, r, "state non valido o scaduto, riprova il login")
		return
	}
	nonceCookie, err := r.Cookie(nonceCookieName)
	if err != nil {
		failLogin(w, r, "flusso di login non valido, riprova")
		return
	}
	verifierCookie, err := r.Cookie(verifierCookieName)
	if err != nil {
		failLogin(w, r, "flusso di login non valido, riprova")
		return
	}

	clearCookie(w, r, stateCookieName)
	clearCookie(w, r, nonceCookieName)
	clearCookie(w, r, verifierCookieName)

	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		failLogin(w, r, "provider OIDC: "+errMsg+" - "+r.URL.Query().Get("error_description"))
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		failLogin(w, r, "parametro 'code' mancante")
		return
	}

	oauthCfg := p.oauth2Config(r)
	token, err := oauthCfg.Exchange(ctx, code, oauth2.VerifierOption(verifierCookie.Value))
	if err != nil {
		log.Printf("announcements: scambio code fallito: %v", err)
		failLogin(w, r, "autenticazione fallita")
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		failLogin(w, r, "risposta del provider priva di id_token")
		return
	}

	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		log.Printf("announcements: verifica id_token fallita: %v", err)
		failLogin(w, r, "autenticazione fallita")
		return
	}
	if idToken.Nonce != nonceCookie.Value {
		failLogin(w, r, "nonce non corrispondente")
		return
	}

	var claims idTokenClaims
	if err := idToken.Claims(&claims); err != nil {
		failLogin(w, r, "claims non leggibili")
		return
	}

	if err := p.checkGroups(claims.Groups); err != nil {
		failLogin(w, r, "accesso negato: "+err.Error())
		return
	}

	displayName := claims.Name
	if displayName == "" {
		displayName = claims.PreferredUsername
	}

	sessID, err := p.sessions.create(session{
		Subject: claims.Subject,
		Email:   claims.Email,
		Name:    displayName,
		Groups:  claims.Groups,
	})
	if err != nil {
		failLogin(w, r, "errore interno")
		return
	}

	setCookie(w, r, sessionCookieName, sessID, sessionTTL)
	http.Redirect(w, r, MountPath+"/", http.StatusFound)
}

// checkGroups verifica che l'utente appartenga ad almeno uno dei gruppi
// elencati in config.OIDC_ALLOWED_GROUPS. Se la lista è vuota, chiunque sia
// autenticato dal provider viene ammesso.
func (p *Panel) checkGroups(userGroups []string) error {
	if len(p.allowedGrps) == 0 {
		return nil
	}
	for _, g := range userGroups {
		if _, ok := p.allowedGrps[g]; ok {
			return nil
		}
	}
	return errors.New("nessuno dei gruppi dell'utente è autorizzato")
}

func (p *Panel) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookieName); err == nil {
		p.sessions.delete(c.Value)
	}
	clearCookie(w, r, sessionCookieName)
	http.Redirect(w, r, MountPath+"/login", http.StatusFound)
}

// handleMe espone in JSON la sessione corrente, utile per popolare
// dinamicamente index.html via fetch('/api/me').
func (p *Panel) handleMe(w http.ResponseWriter, r *http.Request) {
	sess, _ := sessionFromContext(r)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sess)
}
