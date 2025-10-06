package main

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type request struct {
	Username string `json:"username"` // required
	Password string `json:"password"` // required
	Pin      string `json:"pin"`      // required (4 digits)
	Gender   int    `json:"gender"`   // required (0 male, 1 female)
	Dob      int    `json:"dob"`      // required (YYYYMMDD)
}

type response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

var (
	db           *sql.DB
	maxBodyBytes int64 = 1 << 20
	usernameRe         = regexp.MustCompile(`^[A-Za-z0-9_.-]{3,32}$`)
)

func main() {
	// Build DSN from your existing VALHALLA_* envs
	addr := os.Getenv("VALHALLA_DATABASE_ADDRESS")
	port := os.Getenv("VALHALLA_DATABASE_PORT")
	user := os.Getenv("VALHALLA_DATABASE_USER")
	pass := os.Getenv("VALHALLA_DATABASE_PASSWORD")
	name := os.Getenv("VALHALLA_DATABASE_DATABASE")

	if addr == "" || port == "" || user == "" || name == "" {
		log.Fatal("missing required VALHALLA_DATABASE_* environment variables")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4", user, pass, addr, port, name)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(16)
	db.SetConnMaxLifetime(30 * time.Minute)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	http.HandleFunc("/register", registerHandler)

	log.Println("register service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	defer r.Body.Close()

	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(strings.ToLower(ct), "application/json") {
		http.Error(w, "use application/json", http.StatusUnsupportedMediaType)
		return
	}

	// Read once so we can verify required keys then unmarshal
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	// Ensure required JSON fields exist
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	for _, k := range []string{"username", "password", "pin", "gender", "dob"} {
		if _, ok := raw[k]; !ok {
			writeErr(w, http.StatusBadRequest, k+" is required")
			return
		}
	}

	var req request
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	// --- Validation (all required) ---
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	req.Pin = strings.TrimSpace(req.Pin)

	if !usernameRe.MatchString(req.Username) {
		writeErr(w, http.StatusBadRequest, "username must be 3-32 chars [A-Za-z0-9_.-]")
		return
	}
	if len(req.Password) < 8 {
		writeErr(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	// PIN must be exactly 4 digits
	if !isFourDigits(req.Pin) {
		writeErr(w, http.StatusBadRequest, "pin must be exactly 4 digits")
		return
	}
	// Gender must be explicitly 0 (male) or 1 (female)
	if req.Gender != 0 && req.Gender != 1 {
		writeErr(w, http.StatusBadRequest, "gender must be 0 (male) or 1 (female)")
		return
	}
	// DOB must be a valid date before today
	dobStr := itoa(req.Dob)
	if len(dobStr) != 8 || !allDigits(dobStr) {
		writeErr(w, http.StatusBadRequest, "dob must be YYYYMMDD (8 digits)")
		return
	}

	dob, err := time.Parse("20060102", dobStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "dob must be a valid date (YYYYMMDD)")
		return
	}
	if dob.After(time.Now()) {
		writeErr(w, http.StatusBadRequest, "dob must be a date before today")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Uniqueness check
	var exists int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM accounts WHERE username=?`, req.Username).Scan(&exists); err != nil {
		log.Printf("count error: %v", err)
		writeErr(w, http.StatusInternalServerError, "server error")
		return
	}
	if exists > 0 {
		writeErr(w, http.StatusConflict, "username already taken")
		return
	}

	// SHA-512 (hex) password to match your server’s scheme
	h := sha512.New()
	h.Write([]byte(req.Password))
	hashed := hex.EncodeToString(h.Sum(nil))

	// Insert row
	_, err = db.ExecContext(ctx, `
		INSERT INTO accounts
			(username, password, pin, isLogedIn, adminLevel, isBanned, gender, dob, eula, nx, maplepoints)
		VALUES (?, ?, ?, 0, 0, 0, ?, ?, 1, 0, 0)
	`, req.Username, hashed, req.Pin, req.Gender, req.Dob)
	if err != nil {
		// If you later add a DB UNIQUE index on username, duplicates will also hit here
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			writeErr(w, http.StatusConflict, "username already taken")
			return
		}
		log.Printf("insert error: %v", err)
		writeErr(w, http.StatusInternalServerError, "could not create account")
		return
	}

	writeJSON(w, http.StatusOK, response{Status: "ok"})
}

// --- helpers ---

func writeErr(w http.ResponseWriter, code int, msg string) { writeJSON(w, code, response{Status: "error", Error: msg}) }

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func isFourDigits(s string) bool { return len(s) == 4 && allDigits(s) }

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// tiny itoa to avoid importing strconv for one use
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	pos := len(b)
	n := i
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		pos--
		b[pos] = byte('0' + (n % 10))
		n /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}
