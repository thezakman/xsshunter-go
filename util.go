package main

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func get_secure_random_string(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hash_string(input string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func check_hash(input string, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(input))
	return err == nil
}

// func generate_log(input string) {
// 	datetime :=
// }

func checkFileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

func get_env(key string) string {
	err := godotenv.Load()
	if err != nil {
		return ""
	}
	return os.Getenv(key)
}

func parameter_to_int(input string, default_int int) int {
	if input == "" {
		return default_int
	}
	value, err := strconv.Atoi(input)
	if err != nil {
		return default_int
	}
	return value
}

func update_setting(setting_key string, setting_value string) {
	db := establish_database_connection()
	defer db.Close()
	db.Exec("UPDATE settings SET value = :value WHERE key = :key", setting_value, setting_key)
}

func make_folder_if_not_exists(folder string) {
	if !checkFileExists(folder) {
		os.MkdirAll(folder, 0755)
	}
}
