package utils

import (
	"fmt"
	"net/http"
	"os"
)

var LogFile *os.File

// LogRequest enregistre les journaux des requêtes
func LogRequest(r *http.Request) {
	if LogFile == nil {
		return
	}
	clientIP := GetVisitorIP(r)
	logLine := fmt.Sprintf("%s - %s %s\n", clientIP, r.Method, r.URL.Path)
	LogFile.WriteString(logLine)
}
