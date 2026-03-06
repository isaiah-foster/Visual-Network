package seed

import (
	"visual-network/auth_service/auth"
	"visual-network/auth_service/store"

	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := getEnv("MYSQL_DSN", "")

	if dsn == "" {
		log.Fatal("MYSQL_DSN is required")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	log.Println("Connected to MySQL")

	var password string = "p@ssword"
	passHash, err := auth.HashPasswordArgon2id(password, auth.DefaultArgon2idParameters)
	if err != nil {
		log.Fatal("failed to hash passoword: ", err)
	}

	st := &store.Store{DB: db}
	err = st.CreateUser(context.Background(), "admin", "admin@visual-network.com", passHash, "admin")
	if err != nil {
		log.Fatal("failed to create admin user: ", err)
	}

	log.Println("Admin user created with user_id 'admin' and password 'p@ssword'")

}

func getEnv(key, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultValue
}
