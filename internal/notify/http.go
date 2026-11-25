package notify

import "net/http"

// HTTPPanics reports a handler panic while preserving net/http's recovery and logging.
func HTTPPanics(next http.Handler) http.Handler {
	if telegram.Load() == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if failure := recover(); failure != nil {
				if failure != http.ErrAbortHandler {
					operation := "HTTP handler panic"
					if r.Pattern != "" {
						operation += " on " + r.Pattern
					}
					telegram.Load().queueError(operation, operation)
				}
				panic(failure)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
