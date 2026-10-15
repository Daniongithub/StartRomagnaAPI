package announcements

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testPanel(t *testing.T) *Panel {
	t.Helper()
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		t.Fatal(err)
	}
	p := &Panel{
		sessions:     newSessionStore(),
		staticServer: http.FileServer(http.FS(staticFS)),
		mux:          http.NewServeMux(),
		apiRoutes:    make(map[string]struct{}),
	}
	p.routes()
	return p
}

func TestPanelStaticPagesAndProtectedAPI(t *testing.T) {
	panel := testPanel(t)
	if err := panel.RegisterAPI("GET", "/api/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	panel.Mount(mux)

	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, MountPath, nil))
	if response.Code != http.StatusMovedPermanently || response.Header().Get("Location") != MountPath+"/" {
		t.Fatalf("mount senza slash: status=%d location=%q", response.Code, response.Header().Get("Location"))
	}

	response = httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, MountPath+"/", nil))
	if response.Code != http.StatusFound || response.Header().Get("Location") != MountPath+"/login" {
		t.Fatalf("home senza sessione: status=%d location=%q", response.Code, response.Header().Get("Location"))
	}

	response = httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, MountPath+"/login", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "bootstrap@5.3.3") {
		t.Fatalf("pagina login non servita correttamente: status=%d", response.Code)
	}

	response = httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, MountPath+"/api/test", nil))
	if response.Code != http.StatusFound {
		t.Fatalf("API accessibile senza sessione: status=%d", response.Code)
	}

	sessionID, err := panel.sessions.create(session{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, MountPath+"/api/test", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("API non raggiunta con sessione valida: status=%d", response.Code)
	}
}
