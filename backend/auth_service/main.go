package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"visual-network/auth_service/handlers"
	"visual-network/auth_service/store"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := getEnv("MYSQL_DSN", "")
	addr := getEnv("SERVER_ADDR", ":8081")
	cookieSecure := getEnv("COOKIE_SECURE", "false") == "true"

	if dsn == "" {
		log.Fatal("MYSQL_DSN is required")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	log.Println("Connected to MySQL")

	st := &store.Store{DB: db}

	authHandlers := &handlers.AuthHandlers{
		Store:        st,
		CookieSecure: cookieSecure,
		SessionTTL:   8 * time.Hour,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/login", authHandlers.Login)

	mux.Handle("/auth/logout",
		handlers.WithAuth(st, http.HandlerFunc(authHandlers.Logout)))

	mux.Handle("/auth/admin/users",
		handlers.WithAuth(st, http.HandlerFunc(authHandlers.CreateUser)))

	mux.Handle("/auth/me",
		handlers.WithAuth(st, http.HandlerFunc(authHandlers.Me)))

	server := &http.Server{
		Addr:         addr,
		Handler:      corsMiddleware(loggingMiddleware(mux)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server is running on %s\n", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced server shutdown: %v", err)
	}

	log.Println("Server exited cleanly")
}

func getEnv(key, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultValue
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		next.ServeHTTP(writer, request)
		log.Printf("%s %s %s", request.Method, request.URL.Path, time.Since(start))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Replace with frontend origin
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, request)
	})
}
