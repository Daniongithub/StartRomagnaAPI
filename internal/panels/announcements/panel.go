package announcements

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"startromagnaapi/config"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const MountPath = "/panel/announcements"

//go:embed static
var staticFiles embed.FS

type Panel struct {
	oauth2Base   oauth2.Config
	verifier     *oidc.IDTokenVerifier
	sessions     *sessionStore
	allowedGrps  map[string]struct{}
	staticServer http.Handler
	mux          *http.ServeMux
	apiRoutes    map[string]struct{}
}

// New inizializza OIDC e le pagine statiche. Va chiamata solo quando
// config.ENABLE_AUTH è true.
func New(ctx context.Context) (*Panel, error) {
	provider, err := oidc.NewProvider(ctx, config.OIDC_ISSUER_URL)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery fallita su %s: %w", config.OIDC_ISSUER_URL, err)
	}

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return nil, fmt.Errorf("inizializzazione file statici del pannello: %w", err)
	}

	allowed := make(map[string]struct{}, len(config.OIDC_ALLOWED_GROUPS))
	for _, group := range config.OIDC_ALLOWED_GROUPS {
		if group = strings.TrimSpace(group); group != "" {
			allowed[group] = struct{}{}
		}
	}

	p := &Panel{
		oauth2Base: oauth2.Config{
			ClientID: config.OIDC_CLIENT_ID,
			Endpoint: provider.Endpoint(),
			Scopes:   []string{oidc.ScopeOpenID, "profile", "email", "groups"},
		},
		verifier:     provider.Verifier(&oidc.Config{ClientID: config.OIDC_CLIENT_ID}),
		sessions:     newSessionStore(),
		allowedGrps:  allowed,
		staticServer: http.FileServer(http.FS(staticFS)),
		mux:          http.NewServeMux(),
		apiRoutes:    make(map[string]struct{}),
	}
	p.routes()
	return p, nil
}

// RegisterAPI collega un handler Go al pannello. Il percorso è relativo al
// pannello e deve iniziare con /api/; l'handler richiede una sessione OIDC.
func (p *Panel) RegisterAPI(method, path string, handler http.Handler) error {
	method = strings.TrimSpace(method)
	if method == "" || path != strings.TrimSpace(path) || !strings.HasPrefix(path, "/api/") ||
		strings.ContainsAny(path, "?#") || handler == nil {
		return fmt.Errorf("route API del pannello non valida: method=%q path=%q", method, path)
	}

	pattern := method + " " + path
	if _, exists := p.apiRoutes[pattern]; exists {
		return fmt.Errorf("route API del pannello già registrata: %s", pattern)
	}
	p.mux.Handle(pattern, p.requireAuth(handler))
	p.apiRoutes[pattern] = struct{}{}
	return nil
}

func (p *Panel) routes() {
	p.mux.Handle("GET /", p.requireAuth(http.HandlerFunc(p.handleIndex)))
	p.mux.HandleFunc("GET /login", p.handleLoginPage)
	p.mux.HandleFunc("GET /auth/start", p.handleAuthStart)
	p.mux.HandleFunc("GET /auth/callback", p.handleCallback)
	p.mux.HandleFunc("GET /logout", p.handleLogout)
	p.mux.Handle("GET /api/me", p.requireAuth(http.HandlerFunc(p.handleMe)))
	p.mux.Handle("GET /static/", http.StripPrefix("/static/", p.staticServer))
}

// Mux restituisce l'handler da montare sotto MountPath, normalmente con
// http.StripPrefix(MountPath, panel.Mux()).
func (p *Panel) Mux() http.Handler {
	return p.mux
}

// Mount registra il pannello sul mux dell'applicazione, incluso il redirect
// dall'URL senza slash finale al percorso canonico.
func (p *Panel) Mount(mux *http.ServeMux) {
	mux.HandleFunc(MountPath, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, MountPath+"/", http.StatusMovedPermanently)
	})
	mux.Handle(MountPath+"/", http.StripPrefix(MountPath, p.Mux()))
}

func (p *Panel) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	p.serveStatic(w, r, "index.html")
}

func (p *Panel) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	p.serveStatic(w, r, "login.html")
}

func (p *Panel) serveStatic(w http.ResponseWriter, r *http.Request, path string) {
	request := r.Clone(r.Context())
	request.URL.Path = "/" + path
	p.staticServer.ServeHTTP(w, request)
}
