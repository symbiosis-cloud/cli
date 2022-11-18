package util

import (
	"os"
	"os/exec"
	"reflect"
)

func GetEnvOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func GetKeys(src map[string]interface{}) []string {
	keys := make([]string, len(src))

	i := 0
	for k := range src {
		keys[i] = k
		i++
	}

	return keys
}

func IsPointer(v interface{}) bool {
	if reflect.ValueOf(v).Kind() == reflect.Ptr {
		return true
	}
	return false
}
