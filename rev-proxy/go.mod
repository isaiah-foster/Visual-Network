module visual-network-proxy

go 1.25.0

require (
	github.com/go-sql-driver/mysql v1.9.3
	github.com/gorilla/websocket v1.5.0
)

require filippo.io/edwards25519 v1.1.0 // indirect

replace visual-network => ../rev-proxy/
