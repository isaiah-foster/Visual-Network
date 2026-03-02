package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	ListenAddr  string
	FrontendDir string
	AuthService string
	BackendAddr string
	TLSCert     string
	TLSKey      string
	MySQLDSN    string
	UseHTTPS    bool
	MetricsPort string
	HTTPAddr    string
}

type User struct {
	UserID   string
	Email    string
	PassHash string
	Role     string
	IsActive bool
}

type Store struct {
	DB *sql.DB
}

func (s *Store) GetUserBySessionHash(ctx context.Context, tokenHash []byte) (*User, error) {
	const query = `
	SELECT user_id, email, pass_hash, role, is_active
	FROM user_auth
	WHERE session_token_hash = ?
	AND session_expiration > NOW()
	AND is_active = 1
	LIMIT 1;`

	user := &User{}
	var isActiveInt int
	err := s.DB.QueryRowContext(ctx, query, tokenHash).Scan(
		&user.UserID,
		&user.Email,
		&user.PassHash,
		&user.Role,
		&isActiveInt,
	)
	if err != nil {
		return nil, err
	}
	user.IsActive = isActiveInt == 1
	return user, nil
}

type Server struct {
	config      Config
	store       *Store
	upgrader    websocket.Upgrader
	metricsHub  *MetricsHub
	httpServer  *http.Server
	httpsServer *http.Server
}

type MetricsHub struct {
	clients    map[*MetricsClient]bool
	broadcast  chan interface{}
	register   chan *MetricsClient
	unregister chan *MetricsClient
	mu         sync.RWMutex
}

type MetricsClient struct {
	conn   *websocket.Conn
	send   chan interface{}
	userID string
}

type MetricsMessage struct {
	Type      string      `json:"type"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

func main() {
	listenAddr := flag.String("listen", ":8443", "HTTPS listen address")
	frontendDir := flag.String("frontend", "../frontend/build", "Frontend build directory")
	authService := flag.String("auth", "http://localhost:8081", "Auth service URL")
	backendAddr := flag.String("backend", "http://localhost:9090", "Backend service URL")
	tlsCert := flag.String("cert", "", "TLS certificate file (leave empty for no HTTPS)")
	tlsKey := flag.String("key", "", "TLS key file (leave empty for no HTTPS)")
	mysqlDSN := flag.String("mysql-dsn", "", "MySQL DSN for auth verification")
	metricsPort := flag.String("metrics-port", "9090", "Port where metrics are available")
	httpPort := flag.String("http-port", ":8080", "HTTP listen address")
	flag.Parse()

	cfg := Config{
		ListenAddr:  *listenAddr,
		FrontendDir: *frontendDir,
		AuthService: *authService,
		BackendAddr: *backendAddr,
		TLSCert:     *tlsCert,
		TLSKey:      *tlsKey,
		MySQLDSN:    *mysqlDSN,
		MetricsPort: *metricsPort,
		HTTPAddr:    *httpPort,
	}
	cfg.UseHTTPS = cfg.TLSCert != "" && cfg.TLSKey != ""

	// Initialize database connection if MySQL DSN provided
	var store *Store
	if cfg.MySQLDSN != "" {
		db, err := sql.Open("mysql", cfg.MySQLDSN)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}

		db.SetMaxOpenConns(20)
		db.SetMaxIdleConns(10)
		db.SetConnMaxLifetime(30 * time.Minute)

		store = &Store{DB: db}
		log.Println("Connected to MySQL for auth verification")
	}

	srv := &Server{
		config:     cfg,
		store:      store,
		metricsHub: NewMetricsHub(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// In production, validate origin header
				return true
			},
		},
	}

	// Start metrics hub
	go srv.metricsHub.run()

	// Setup routes
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", srv.handleHealth)

	// WebSocket for metrics
	mux.HandleFunc("/api/metrics/ws", srv.handleMetricsWebSocket)

	// API proxy to auth service
	authURL, _ := url.Parse(cfg.AuthService)
	authProxy := httputil.NewSingleHostReverseProxy(authURL)
	mux.Handle("/api/auth/", http.StripPrefix("/api/auth", authProxy))

	// API proxy to backend
	backendURL, _ := url.Parse(cfg.BackendAddr)
	backendProxy := httputil.NewSingleHostReverseProxy(backendURL)
	mux.Handle("/api/backend/", http.StripPrefix("/api/backend", backendProxy))

	// Serve frontend (all other paths)
	mux.HandleFunc("/", srv.handleFrontend)

	// Create HTTP server
	srv.httpServer = &http.Server{
		Addr:           cfg.HTTPAddr,
		Handler:        mux,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Create HTTPS server if certificates provided
	if cfg.UseHTTPS {
		cert, err := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)
		if err != nil {
			log.Fatalf("Failed to load TLS certificates: %v", err)
		}

		srv.httpsServer = &http.Server{
			Addr:    cfg.ListenAddr,
			Handler: mux,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
			ReadTimeout:    15 * time.Second,
			WriteTimeout:   15 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
	}

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server
	go func() {
		log.Printf("Starting HTTP server on %s", srv.httpServer.Addr)
		if err := srv.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start HTTPS server if configured
	if cfg.UseHTTPS {
		go func() {
			log.Printf("Starting HTTPS server on %s", srv.httpsServer.Addr)
			if err := srv.httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTPS server error: %v", err)
			}
		}()
	}

	log.Println("Reverse proxy started successfully")
	log.Printf("  HTTP:  http://localhost:8080")
	if cfg.UseHTTPS {
		log.Printf("  HTTPS: https://localhost%s", cfg.ListenAddr)
	}
	log.Println("  Frontend: served at /")
	log.Println("  Auth API: proxied to /api/auth/*")
	log.Println("  Metrics WS: available at /api/metrics/ws")

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal: %v, shutting down...", sig)

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	if cfg.UseHTTPS && srv.httpsServer != nil {
		if err := srv.httpsServer.Shutdown(ctx); err != nil {
			log.Printf("HTTPS server shutdown error: %v", err)
		}
	}

	log.Println("Reverse proxy stopped")
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (s *Server) handleFrontend(w http.ResponseWriter, r *http.Request) {
	// Try to serve the requested file
	filePath := strings.TrimPrefix(r.URL.Path, "/")
	if filePath == "" {
		filePath = "index.html"
	}

	fullPath := s.config.FrontendDir + "/" + filePath

	// Check if file exists
	if _, err := os.Stat(fullPath); err == nil {
		http.ServeFile(w, r, fullPath)
		return
	}

	// For SPA, serve index.html for non-existent routes
	http.ServeFile(w, r, s.config.FrontendDir+"/index.html")
}

func (s *Server) handleMetricsWebSocket(w http.ResponseWriter, r *http.Request) {
	// Verify session/auth if MySQL connected
	userID := "anonymous"
	if s.store != nil {
		sessionCookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		tokenHash := sha256.Sum256([]byte(sessionCookie.Value))
		user, err := s.store.GetUserBySessionHash(r.Context(), tokenHash[:])
		if err != nil || !user.IsActive {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		userID = user.UserID
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &MetricsClient{
		conn:   conn,
		send:   make(chan interface{}, 256),
		userID: userID,
	}

	s.metricsHub.register <- client
	log.Printf("Client connected to metrics: %s", userID)

	go client.readPump()
	go client.writePump()
}

func (s *Server) broadcastMetrics(data interface{}) {
	msg := MetricsMessage{
		Type:      "metrics",
		Timestamp: time.Now().Unix(),
		Data:      data,
	}
	s.metricsHub.broadcast <- msg
}

// MetricsHub manages WebSocket connections
func NewMetricsHub() *MetricsHub {
	return &MetricsHub{
		clients:    make(map[*MetricsClient]bool),
		broadcast:  make(chan interface{}),
		register:   make(chan *MetricsClient),
		unregister: make(chan *MetricsClient),
	}
}

func (h *MetricsHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					// Client send channel full, close it
					go func(c *MetricsClient) {
						h.unregister <- c
					}(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (c *MetricsClient) readPump() {
	defer func() {
		// Cleanup on disconnect
		// Add any necessary cleanup here
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg interface{}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			return
		}
	}
}

func (c *MetricsClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(msg); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
