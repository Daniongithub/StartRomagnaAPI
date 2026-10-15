package announcements

import (
	"context"
	"net/http"
)

type ctxKey int

const sessionCtxKey ctxKey = iota

func (p *Panel) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			http.Redirect(w, r, MountPath+"/login", http.StatusFound)
			return
		}

		sess, ok := p.sessions.get(cookie.Value)
		if !ok {
			clearCookie(w, r, sessionCookieName)
			http.Redirect(w, r, MountPath+"/login", http.StatusFound)
			return
		}

		ctx := context.WithValue(r.Context(), sessionCtxKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func sessionFromContext(r *http.Request) (session, bool) {
	sess, ok := r.Context().Value(sessionCtxKey).(session)
	return sess, ok
}
