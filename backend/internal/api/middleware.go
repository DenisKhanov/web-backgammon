package api

import (
	"bufio"
	"context"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/denis/web-backgammon/internal/db"
)

// --- Logging ---

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)
		slog.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.status,
			"dur", time.Since(start).String(),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// Hijack delegates to the underlying ResponseWriter if it implements http.Hijacker.
// This is required for WebSocket upgrade requests that pass through loggingMiddleware.
func (sw *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := sw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// --- CORS ---

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, allowed := range s.origins {
			if origin == allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				break
			}
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- Rate limiting ---

type ipLimiter struct {
	mu       sync.Mutex
	visitors map[string]*rate.Limiter
	r        rate.Limit
	b        int
}

func newIPLimiter(r rate.Limit, b int) *ipLimiter {
	return &ipLimiter{visitors: make(map[string]*rate.Limiter), r: r, b: b}
}

func (l *ipLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if lim, ok := l.visitors[ip]; ok {
		return lim
	}
	lim := rate.NewLimiter(l.r, l.b)
	l.visitors[ip] = lim
	return lim
}

func (l *ipLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if !l.get(ip).Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- Session auth ---

type contextKey string

const playerCtxKey contextKey = "player"

func (s *Server) requireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		player, err := s.players.FindBySession(r.Context(), cookie.Value)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), playerCtxKey, player)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func playerFromCtx(ctx context.Context) *db.Player {
	p, _ := ctx.Value(playerCtxKey).(*db.Player)
	return p
}
