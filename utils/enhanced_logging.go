package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel représente le niveau de journalisation
type LogLevel int

const (
	// Niveaux de log
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
	LogLevelFatal
)

// Logger représente un journaliseur
type Logger struct {
	level     LogLevel
	output    io.Writer
	errorOut  io.Writer
	prefix    string
	mutex     sync.Mutex
	logToFile bool
	logFile   *os.File
	showColor bool
}

var (
	defaultLogger *Logger
	loggerMutex   sync.Mutex
)

// Couleurs ANSI pour la console
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[37m"
)

// LogLevelToString convertit un niveau de log en chaîne
func LogLevelToString(level LogLevel) string {
	switch level {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarning:
		return "WARNING"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LevelColor retourne la couleur ANSI pour un niveau de log
func LevelColor(level LogLevel) string {
	switch level {
	case LogLevelDebug:
		return colorGray
	case LogLevelInfo:
		return colorGreen
	case LogLevelWarning:
		return colorYellow
	case LogLevelError:
		return colorRed
	case LogLevelFatal:
		return colorPurple
	default:
		return colorReset
	}
}

// InitLogger initialise le journaliseur par défaut
func InitLogger(level LogLevel, logToFile bool, logDir string) error {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	// Détecter si nous sommes dans un terminal colorisé
	showColor := false
	if runtime.GOOS != "windows" {
		// Sur Linux/Mac, nous activons la couleur par défaut
		showColor = true
	}

	var logFile *os.File
	var err error

	if logToFile {
		// Créer le répertoire de logs si nécessaire
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("impossible de créer le répertoire de logs: %w", err)
		}

		// Ouvrir ou créer le fichier de log
		logFilePath := filepath.Join(logDir, fmt.Sprintf("bkc_%s.log", time.Now().Format("2006-01-02")))
		logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("impossible d'ouvrir le fichier de logs: %w", err)
		}
	}

	// Créer le logger
	defaultLogger = &Logger{
		level:     level,
		output:    os.Stdout,
		errorOut:  os.Stderr,
		prefix:    "",
		logToFile: logToFile,
		logFile:   logFile,
		showColor: showColor,
	}

	return nil
}

// Close ferme le logger
func (l *Logger) Close() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.logToFile && l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// SetLevel change le niveau de log
func (l *Logger) SetLevel(level LogLevel) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.level = level
}

// SetPrefix définit un préfixe pour tous les logs
func (l *Logger) SetPrefix(prefix string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.prefix = prefix
}

// EnableColor active ou désactive la coloration des logs
func (l *Logger) EnableColor(enable bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.showColor = enable
}

// SetOutputs définit les sorties pour les logs
func (l *Logger) SetOutputs(out io.Writer, errOut io.Writer) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.output = out
	l.errorOut = errOut
}

// log écrit un message dans le journal
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Obtenir l'heure actuelle
	now := time.Now()

	// Construire le message
	message := fmt.Sprintf(format, args...)

	// Obtenir les informations d'appel
	_, file, line, ok := runtime.Caller(2)
	fileInfo := "???"
	if ok {
		// Extraire seulement le nom du fichier, pas le chemin complet
		fileInfo = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	// Formater le message complet
	levelStr := LogLevelToString(level)
	fullMessage := fmt.Sprintf("[%s] %s [%s] %s%s\n",
		now.Format("2006-01-02 15:04:05.000"),
		levelStr,
		fileInfo,
		l.prefix,
		message)

	// Message coloré pour la console
	coloredMessage := fullMessage
	if l.showColor {
		colorCode := LevelColor(level)
		coloredMessage = fmt.Sprintf("%s%s%s", colorCode, fullMessage, colorReset)
	}

	// Écrire dans la sortie appropriée
	var out io.Writer
	if level >= LogLevelError {
		out = l.errorOut
	} else {
		out = l.output
	}

	fmt.Fprint(out, coloredMessage)

	// Si configuré, écrire aussi dans un fichier
	if l.logToFile && l.logFile != nil {
		fmt.Fprint(l.logFile, fullMessage)
	}

	// Si fatal, terminer le programme
	if level == LogLevelFatal {
		os.Exit(1)
	}
}

// Debug écrit un message de niveau debug
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LogLevelDebug, format, args...)
}

// Info écrit un message de niveau info
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LogLevelInfo, format, args...)
}

// Warning écrit un message de niveau warning
func (l *Logger) Warning(format string, args ...interface{}) {
	l.log(LogLevelWarning, format, args...)
}

// Error écrit un message de niveau error
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LogLevelError, format, args...)
}

// Fatal écrit un message de niveau fatal et termine le programme
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LogLevelFatal, format, args...)
	// Le programme se terminera dans la méthode log
}

// Fonctions d'accès au logger par défaut

// Debug écrit un message de niveau debug
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(format, args...)
	}
}

// Info écrit un message de niveau info
func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(format, args...)
	}
}

// Warning écrit un message de niveau warning
func Warning(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warning(format, args...)
	}
}

// Error écrit un message de niveau error
func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(format, args...)
	}
}

// Fatal écrit un message de niveau fatal et termine le programme
func Fatal(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Fatal(format, args...)
	} else {
		// Si le logger n'est pas initialisé, afficher quand même le message
		fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", args...)
		os.Exit(1)
	}
}

// CloseLogger ferme le logger par défaut
func CloseLogger() error {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	if defaultLogger != nil {
		err := defaultLogger.Close()
		defaultLogger = nil
		return err
	}
	return nil
}

// GetLogLevel récupère un niveau de log à partir d'une chaîne
func GetLogLevel(levelStr string) LogLevel {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return LogLevelDebug
	case "INFO":
		return LogLevelInfo
	case "WARNING", "WARN":
		return LogLevelWarning
	case "ERROR":
		return LogLevelError
	case "FATAL":
		return LogLevelFatal
	default:
		return LogLevelInfo // Par défaut
	}
}

// LogMiddlewareHandler crée un middleware de journalisation HTTP
func LogMiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrapper de ResponseWriter pour capturer le code de statut
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Défaut
		}

		// Traiter la requête
		next.ServeHTTP(rw, r)

		// Calculer la durée
		duration := time.Since(start)

		// Déterminer le niveau de log en fonction du code de statut
		level := LogLevelInfo
		if rw.statusCode >= 400 && rw.statusCode < 500 {
			level = LogLevelWarning
		} else if rw.statusCode >= 500 {
			level = LogLevelError
		}

		// Journaliser la requête
		clientIP := GetVisitorIP(r)
		if defaultLogger != nil {
			defaultLogger.log(level, "%s %s %s [%d] in %v", r.Method, r.URL.Path, clientIP, rw.statusCode, duration)
		}
	})
}

// responseWriter est un wrapper pour http.ResponseWriter qui capture le code de statut
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader capture le code de statut
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
