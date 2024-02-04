package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	// "gorm.io/driver/postgres"
	// "gorm.io/driver/sqlite"
	// "gorm.io/gorm"
)

type Settings struct {
	ID    uint
	key   *string
	value *string
}

type PayloadFireResults struct {
	ID                 string // `json:"id"`
	url                string // `json:"url"`
	ip_address         string // `json:"ip_address"`
	referer            string // `json:"referer"`
	user_agent         string // `json:"user_agent"`
	cookies            string // `json:"cookies"`
	title              string // `json:"title"`
	dom                string // `json:"dom"`
	text               string // `json:"text"`
	origin             string // `json:"origin"`
	screenshot_id      string // `json:"screenshot_id"`
	was_iframe         bool   // `json:"was_iframe"`
	browser_timestamp  uint   // `json:"browser_timestamp"`
	correlated_request string // `json:"correlated_request"`
}

type CollectedPages struct {
	ID   uint
	uri  string
	html string
}

type InjectionRequests struct {
	ID            uint
	request       string
	injection_key string
}

func initialize_database() {
	if get_env("DATABASE_URL") != "" {
		initialize_postgres_database()
	} else {
		initialize_sqlite_database()
	}
	initialize_settings()
}

func establish_database_connection() *sql.DB {
	if get_env("DATABASE_URL") != "" {
		return establish_postgres_database_connection()
	}
	return establish_sqlite_database_connection()
}

func initialize_sqlite_database() {
	if _, err := os.Stat("db"); os.IsNotExist(err) {
		err = os.MkdirAll("db", 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	create_sqlite_tables()
}

func initialize_postgres_database() {
	create_postgres_tables()
}

func establish_sqlite_database_connection() *sql.DB {
	dbPath := get_sqlite_database_path()
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func establish_postgres_database_connection() *sql.DB {
	db, err := sql.Open("postgres", get_env("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func create_sqlite_tables() {
	db := establish_sqlite_database_connection()
	defer db.Close()

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT,
		value TEXT
	);
	CREATE TABLE IF NOT EXISTS payload_fire_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT,
		ip_address TEXT,
		referer TEXT,
		user_agent TEXT,
		cookies TEXT,
		title TEXT,
		dom TEXT,
		text TEXT,
		origin TEXT,
		screenshot_id TEXT,
		was_iframe BOOLEAN,
		browser_timestamp INTEGER
	);
	CREATE TABLE IF NOT EXISTS collected_pages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uri TEXT,
		html TEXT
	);
	CREATE TABLE IF NOT EXISTS injection_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		request TEXT,
		injection_key TEXT
	);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func create_postgres_tables() {
	db := establish_postgres_database_connection()
	defer db.Close()

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS settings (
		id SERIAL PRIMARY KEY,
		key TEXT,
		value TEXT
	);
	CREATE TABLE IF NOT EXISTS payload_fire_results (
		id SERIAL PRIMARY KEY,
		url TEXT,
		ip_address TEXT,
		referer TEXT,
		user_agent TEXT,
		cookies TEXT,
		title TEXT,
		dom TEXT,
		text TEXT,
		origin TEXT,
		screenshot_id TEXT,
		was_iframe BOOLEAN,
		browser_timestamp INTEGER
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS collected_pages (
		id SERIAL PRIMARY KEY,
		uri TEXT,
		html TEXT
	);
	CREATE TABLE IF NOT EXISTS injection_requests (
		id SERIAL PRIMARY KEY,
		request TEXT,
		injection_key TEXT
	);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func initialize_settings() {
	initialize_users()
	initialize_configs()
	initialize_correlation_api()
	initialize_setting_helper(PAGES_TO_COLLECT_SETTINGS_KEY, "")
	initialize_setting_helper(SEND_ALERTS_SETTINGS_KEY, "true")
	initialize_setting_helper(CHAINLOAD_URI_SETTINGS_KEY, "")
}

func initialize_users() {
	new_password, err := get_secure_random_string(32)
	if err != nil {
		log.Fatal(err)
	}

	new_user := setup_admin_user(new_password)

	if new_user {
		return
	}

	banner_message := get_default_user_created_banner(new_password)
	fmt.Println(banner_message)
}

func setup_admin_user(password string) bool {
	db := establish_database_connection()
	defer db.Close()

	hashed_password, err := hash_string(password)
	if err != nil {
		log.Fatal(err)
	}

	return initialize_setting_helper(ADMIN_PASSWORD_SETTINGS_KEY, hashed_password)
}

func initialize_configs() {
	session_secret, err := get_secure_random_string(64)
	if err != nil {
		log.Fatal(err)
	}
	initialize_setting_helper(session_secret_key, session_secret)
}

func initialize_correlation_api() {
	api_key, err := get_secure_random_string(64)
	if err != nil {
		log.Fatal(err)
	}
	initialize_setting_helper(CORRELATION_API_SECRET_SETTINGS_KEY, api_key)
}

func initialize_setting_helper(key string, value string) bool {
	db := establish_database_connection()
	defer db.Close()

	var has_setting int
	db.QueryRow("SELECT COUNT(1) FROM settings WHERE key = ?", key).Scan(&has_setting)
	if has_setting != 1 {
		_, err := db.Exec("INSERT INTO settings (key, value) VALUES (?, ?)", key, value)
		if err != nil {
			log.Fatal(err)
		}
		return false
	}
	return true
}

func get_default_user_created_banner(password string) string {
	return `
   ============================================================================
    █████╗ ████████╗████████╗███████╗███╗   ██╗████████╗██╗ ██████╗ ███╗   ██╗
   ██╔══██╗╚══██╔══╝╚══██╔══╝██╔════╝████╗  ██║╚══██╔══╝██║██╔═══██╗████╗  ██║
   ███████║   ██║      ██║   █████╗  ██╔██╗ ██║   ██║   ██║██║   ██║██╔██╗ ██║
   ██╔══██║   ██║      ██║   ██╔══╝  ██║╚██╗██║   ██║   ██║██║   ██║██║╚██╗██║
   ██║  ██║   ██║      ██║   ███████╗██║ ╚████║   ██║   ██║╚██████╔╝██║ ╚████║
   ╚═╝  ╚═╝   ╚═╝      ╚═╝   ╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
																			  
   vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv
	   An admin user (for the admin control panel) has been created
	   with the following password:
   
	   PASSWORD: ` + password + `
   
	   XSS Hunter Go has only one user for the instance. Do not
	   share this password with anyone who you don't trust. Save it
	   in your password manager and don't change it to anything that
	   is bruteforcable.
   
   ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
    █████╗ ████████╗████████╗███████╗███╗   ██╗████████╗██╗ ██████╗ ███╗   ██╗
   ██╔══██╗╚══██╔══╝╚══██╔══╝██╔════╝████╗  ██║╚══██╔══╝██║██╔═══██╗████╗  ██║
   ███████║   ██║      ██║   █████╗  ██╔██╗ ██║   ██║   ██║██║   ██║██╔██╗ ██║
   ██╔══██║   ██║      ██║   ██╔══╝  ██║╚██╗██║   ██║   ██║██║   ██║██║╚██╗██║
   ██║  ██║   ██║      ██║   ███████╗██║ ╚████║   ██║   ██║╚██████╔╝██║ ╚████║
   ╚═╝  ╚═╝   ╚═╝      ╚═╝   ╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
																			  
   ============================================================================`
}
