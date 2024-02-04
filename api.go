package main

import (
	"encoding/json"
	"net/http"
	"os"
)

func authCheckHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	is_authenticated := get_and_validate_jwt(r)
	if !is_authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	is_authenticated := get_and_validate_jwt(r)
	if !is_authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
	}
	if r.Method == "GET" {
		db := establish_database_connection()
		defer db.Close()

		rows, err := db.Query("SELECT key, value FROM settings WHERE key IN (?, ?, ?, ?)", CORRELATION_API_SECRET_SETTINGS_KEY, CHAINLOAD_URI_SETTINGS_KEY, PAGES_TO_COLLECT_SETTINGS_KEY, SEND_ALERTS_SETTINGS_KEY)
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		settings := map[string]string{}
		for rows.Next() {
			var key, value string
			err = rows.Scan(&key, &value)
			if err != nil {
				http.Error(w, "Error scanning database", http.StatusInternalServerError)
				return
			}
			settings[key] = value
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(settings)
	} else if r.Method == "PUT" {
		var setting_key = r.FormValue("key")
		var setting_value = r.FormValue("value")
		if setting_key == "" || setting_value == "" {
			http.Error(w, "Invalid key or value", http.StatusBadRequest)
			return
		}

		switch setting_key {
		case ADMIN_PASSWORD_SETTINGS_KEY:
			hashed_password, err := hash_string(setting_value)
			if err != nil {
				http.Error(w, "Error hashing password", http.StatusInternalServerError)
				return
			}
			update_setting(ADMIN_PASSWORD_SETTINGS_KEY, hashed_password)
		case CORRELATION_API_SECRET_SETTINGS_KEY:
			update_setting(CORRELATION_API_SECRET_SETTINGS_KEY, setting_value)
		case CHAINLOAD_URI_SETTINGS_KEY:
			update_setting(CHAINLOAD_URI_SETTINGS_KEY, setting_value)
		case PAGES_TO_COLLECT_SETTINGS_KEY:
			update_setting(PAGES_TO_COLLECT_SETTINGS_KEY, setting_value)
		case SEND_ALERTS_SETTINGS_KEY:
			update_setting(SEND_ALERTS_SETTINGS_KEY, setting_value)
		default:
			http.Error(w, "Invalid key", http.StatusBadRequest)
		}
	} else {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	if r.Method != "POST" {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	db := establish_database_connection()
	defer db.Close()

	var password string
	db.QueryRow("SELECT value FROM settings WHERE key = ?", ADMIN_PASSWORD_SETTINGS_KEY).Scan(&password)

	if password == "" {
		http.Error(w, "No password set", http.StatusInternalServerError)
		return
	}

	if check_hash(r.FormValue("password"), password) {
		generate_and_set_jwt(w, r)
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
	}
}

func payloadFiresHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	is_authenticated := get_and_validate_jwt(r)
	if !is_authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
	}
	if r.Method == "GET" {
		page_string := r.URL.Query().Get("page")
		limit_string := r.URL.Query().Get("limit")
		page := parameter_to_int(page_string, 1) - 1
		limit := parameter_to_int(limit_string, 10)
		offset := page * limit

		db := establish_database_connection()
		defer db.Close()

		rows, err := db.Query("SELECT id, url, ip_address, referer, user_agent, cookies, title, dom, text, origin, screenshot_id, was_iframe, browser_timestamp FROM payload_fire_results ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		payload_fires := []PayloadFireResults{}
		for rows.Next() {
			var payload_fire PayloadFireResults
			err = rows.Scan(&payload_fire.ID, &payload_fire.url, &payload_fire.ip_address, &payload_fire.referer, &payload_fire.user_agent, &payload_fire.cookies, &payload_fire.title, &payload_fire.dom, &payload_fire.text, &payload_fire.origin, &payload_fire.screenshot_id, &payload_fire.was_iframe, &payload_fire.browser_timestamp)
			if err != nil {
				http.Error(w, "Error scanning database", http.StatusInternalServerError)
				return
			}
			payload_fires = append(payload_fires, payload_fire)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload_fires)
	} else if r.Method == "DELETE" {
		ids_to_delete := r.URL.Query()["ids"]
		if len(ids_to_delete) == 0 {
			http.Error(w, "No ids to delete", http.StatusBadRequest)
			return
		}
		db := establish_database_connection()
		defer db.Close()

		rows, err := db.Query("SELECT screenshot_id FROM payload_fire_results WHERE id IN (?)", ids_to_delete)
		if err != nil {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var screenshot_id string
			err = rows.Scan(&screenshot_id)
			if err != nil {
				http.Error(w, "Error scanning database", http.StatusInternalServerError)
				return
			}
			payload_fire_image_filename := get_env("SCREENSHOT_DIRECTORY") + "/" + screenshot_id + ".png.gz"
			os.Remove(payload_fire_image_filename)
			db.Exec("DELETE FROM payload_fire_results WHERE screenshot_id = ?", screenshot_id)
		}
	} else {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
	}
}

func collectedPagesHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	is_authenticated := get_and_validate_jwt(r)
	if !is_authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
	}
	if r.Method == "GET" {

	} else if r.Method == "DELETE" {

	} else {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
	}
}

func recordInjectionHandler(w http.ResponseWriter, r *http.Request) {
	set_secure_headers(w, r)
	is_authenticated := get_and_validate_jwt(r)
	if !is_authenticated {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
	}
	if r.Method != "POST" {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
}
