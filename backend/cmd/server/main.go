package main

import (
	"context"
	"humanguard/auth"
	"humanguard/handlers"
	"humanguard/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const requestIDKey contextKey = "requestID"

func main() {
	store := connectToDatabase()
	defer store.Close()

	server := startHTTPServer(store)
	waitForShutdown(server)

	log.Println("Shutting down...")
}

func connectToDatabase() storage.Storage {
	cfg := &storage.Config{
		DBURL:       getEnv("DATABASE_URL", "postgres://postgres:123@postgres:5432/humanguard?sslmode=disable"),
		UploadDir:   getEnv("UPLOAD_DIR", "./data/uploads"),
		MaxFileSize: 100 * 1024 * 1024,
	}

	store, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Connected to database")

	if err := store.Ping(); err != nil {
		log.Fatal("Database ping failed:", err)
	}
	log.Println("Database ping successful")

	return store
}

func startHTTPServer(store storage.Storage) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	jwtService := auth.NewJWTService(getEnv("JWT_SECRET", "super-secret-key"))
	totpService := auth.NewTOTPService()

	oauthService := auth.NewOAuthService(
		"humanguard",
		getEnv("OAUTH_CLIENT_SECRET", "1meWH6qPeEhd17APBADgo20Mth1J5pzP"),
		"http://localhost:8080/api/auth/keycloak/callback",
		getEnv("KEYCLOAK_URL", "http://localhost:8081"),
	)

	authMiddleware := auth.AuthMiddleware(jwtService)

	{
		userHandler := handlers.NewUserHandler(store, jwtService, totpService, oauthService)

		mux.HandleFunc("POST /api/users", userHandler.CreateUser)
		mux.HandleFunc("POST /api/login", userHandler.Login)

		mux.HandleFunc("GET /api/auth/keycloak/login", userHandler.KeycloakLogin)
		mux.HandleFunc("GET /api/auth/keycloak/callback", userHandler.KeycloakCallback)

		// Protected
		mux.Handle("GET /api/users", authMiddleware(http.HandlerFunc(userHandler.ListUsers)))
		mux.Handle("GET /api/me", authMiddleware(http.HandlerFunc(userHandler.GetCurrentUser)))
		mux.Handle("GET /api/users/{id}", authMiddleware(http.HandlerFunc(userHandler.GetUser)))
		mux.Handle("GET /api/users/email/{email}", authMiddleware(http.HandlerFunc(userHandler.GetUserByEmail)))
		mux.Handle("GET /api/users/exists", authMiddleware(http.HandlerFunc(userHandler.CheckEmailExists)))
		mux.Handle("GET /api/users/oauth/{provider}/{oauthId}", authMiddleware(http.HandlerFunc(userHandler.GetUserByOAuth)))
		mux.Handle("PUT /api/users/{id}", authMiddleware(http.HandlerFunc(userHandler.UpdateUser)))
		mux.Handle("DELETE /api/users/{id}", authMiddleware(http.HandlerFunc(userHandler.DeleteUser)))
		mux.Handle("POST /api/users/{id}/password", authMiddleware(http.HandlerFunc(userHandler.ChangePassword)))
		mux.Handle("POST /api/users/{id}/avatar", authMiddleware(http.HandlerFunc(userHandler.UpdateAvatar)))
	}

	{
		siteHandler := handlers.NewSiteHandler(store)

		mux.Handle("POST /api/sites", authMiddleware(http.HandlerFunc(siteHandler.CreateSite)))
		mux.Handle("GET /api/sites", authMiddleware(http.HandlerFunc(siteHandler.ListSites)))
		mux.Handle("GET /api/sites/{id}", authMiddleware(http.HandlerFunc(siteHandler.GetSite)))
		mux.Handle("PUT /api/sites/{id}", authMiddleware(http.HandlerFunc(siteHandler.UpdateSite)))
		mux.Handle("DELETE /api/sites/{id}", authMiddleware(http.HandlerFunc(siteHandler.DeleteSite)))
		mux.Handle("POST /api/sites/{id}/activate", authMiddleware(http.HandlerFunc(siteHandler.ActivateSite)))
		mux.Handle("POST /api/sites/{id}/suspend", authMiddleware(http.HandlerFunc(siteHandler.SuspendSite)))
		mux.Handle("GET /api/sites/{id}/settings", authMiddleware(http.HandlerFunc(siteHandler.GetSiteSettings)))
		mux.Handle("PUT /api/sites/{id}/settings", authMiddleware(http.HandlerFunc(siteHandler.UpdateSiteSettings)))
	}

	{
		sessionHandler := handlers.NewSessionHandler(store)

		mux.Handle("POST /api/sessions", authMiddleware(http.HandlerFunc(sessionHandler.CreateSession)))
		mux.Handle("GET /api/sessions/{id}", authMiddleware(http.HandlerFunc(sessionHandler.GetSession)))
		mux.Handle("PUT /api/sessions/{id}", authMiddleware(http.HandlerFunc(sessionHandler.UpdateSession)))
		mux.Handle("DELETE /api/sessions/{id}", authMiddleware(http.HandlerFunc(sessionHandler.DeactivateSession)))
		mux.Handle("POST /api/sessions/{id}/block", authMiddleware(http.HandlerFunc(sessionHandler.BlockSession)))
		mux.Handle("POST /api/sessions/{id}/unblock", authMiddleware(http.HandlerFunc(sessionHandler.UnblockSession)))
		mux.Handle("PATCH /api/sessions/{id}/risk", authMiddleware(http.HandlerFunc(sessionHandler.UpdateRiskScore)))
		mux.Handle("POST /api/sessions/{id}/activity", authMiddleware(http.HandlerFunc(sessionHandler.UpdateSessionActivity)))
		mux.Handle("POST /api/sessions/{id}/captcha", authMiddleware(http.HandlerFunc(sessionHandler.MarkCaptchaShown)))
		mux.Handle("POST /api/sessions/cleanup", authMiddleware(http.HandlerFunc(sessionHandler.CleanupExpiredSessions)))
		mux.Handle("GET /api/sites/{id}/sessions", authMiddleware(http.HandlerFunc(sessionHandler.GetSessionsBySite)))
		mux.Handle("GET /api/sites/{id}/sessions/suspicious", authMiddleware(http.HandlerFunc(sessionHandler.GetSuspiciousSessions)))
		mux.Handle("GET /api/sites/{id}/stats", authMiddleware(http.HandlerFunc(sessionHandler.GetSessionStats)))
	}

	fs := storage.NewLocalS3("./data/uploads")
	fileHandler := handlers.NewFileHandler(store, fs)

	mux.Handle("POST /api/files/upload", authMiddleware(http.HandlerFunc(fileHandler.Upload)))
	mux.Handle("GET /api/files/{id}", authMiddleware(http.HandlerFunc(fileHandler.Download)))
	mux.Handle("DELETE /api/files/{id}", authMiddleware(http.HandlerFunc(fileHandler.Delete)))
	mux.Handle("GET /api/files", authMiddleware(http.HandlerFunc(fileHandler.List)))
	mux.Handle("POST /api/files/share", authMiddleware(http.HandlerFunc(fileHandler.CreateShare)))
	mux.HandleFunc("GET /api/files/share/{token}", fileHandler.GetByShareToken)

	handler := loggingMiddleware(corsMiddleware(rateLimitMiddleware(newRateLimiter(10, time.Minute))(securityHeadersMiddleware(mux))))

	server := &http.Server{
		Addr:         ":" + getEnv("PORT", "8080"),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Server starting on http://localhost:" + getEnv("PORT", "8080"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	return server
}

func waitForShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server stopped")
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		reqID, err := uuid.NewV7()
		var reqIDStr string
		if err != nil {
			reqIDStr = uuid.New().String()
		} else {
			reqIDStr = reqID.String()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, reqIDStr)
		w.Header().Set("X-Request-Id", reqIDStr)

		next.ServeHTTP(w, r.WithContext(ctx))
		log.Printf("[%s] %s %s %s", reqIDStr, r.Method, r.URL.Path, time.Since(start))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	rawOrigins := getEnv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")
	allowedOrigins := make(map[string]bool)
	for _, o := range strings.Split(rawOrigins, ",") {
		if o = strings.TrimSpace(o); o != "" {
			allowedOrigins[o] = true
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if allowedOrigins[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			} else if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// rateLimiter implements a sliding-window per-IP rate limiter.
type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()
	return rl
}

func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for key, times := range rl.requests {
		var valid []time.Time
		for _, t := range times {
			if now.Sub(t) < rl.window {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = valid
		}
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	var recent []time.Time
	for _, t := range rl.requests[key] {
		if now.Sub(t) < rl.window {
			recent = append(recent, t)
		}
	}
	if len(recent) >= rl.limit {
		rl.requests[key] = recent
		return false
	}
	rl.requests[key] = append(recent, now)
	return true
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.SplitN(xff, ",", 2)[0]
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	addr := r.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i >= 0 {
		return addr[:i]
	}
	return addr
}

func rateLimitMiddleware(rl *rateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.allow(getClientIP(r)) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit exceeded"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
