package main

import (
	"BkC/blockchain"
	"BkC/handlers"
	"BkC/network"
	"BkC/utils"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

// Version of the application
const appVersion = "1.3.0"

// openBrowser opens the default browser with the specified URL
func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	default:
		log.Println("⚠️ Unsupported system for automatic browser opening")
		return
	}

	args = append(args, url)
	if err := exec.Command(cmd, args...).Start(); err != nil {
		log.Printf("❌ Error opening browser: %v", err)
	}
}

// initDirectories creates necessary directories
func initDirectories() error {
	directories := []string{
		"logs",
		"wallets",
		"data",
		"data/contracts",
		"data/metrics",
		"static/js",
		"templates/layouts",
		"network",
		"data/db", // For LevelDB
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("could not create directory %s: %w", dir, err)
		}
	}

	return nil
}

// setupLogging configures logging
func setupLogging() error {
	logFilePath := filepath.Join(utils.Config.LogsDir, "server.log")
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening log file: %w", err)
	}
	utils.LogFile = logFile
	return nil
}

// setupAuditAndSecurity initializes audit and security systems
func setupAuditAndSecurity() {
	// Initialize audit system
	auditPath := filepath.Join(utils.Config.DataDir, "audit.json")
	if err := utils.InitAuditTrail(auditPath); err != nil {
		log.Printf("⚠️ Error initializing audit trail: %v", err)
	}

	// Initialize security risk assessment system
	securityPath := filepath.Join(utils.Config.DataDir, "security.json")
	if err := utils.InitSecurityRiskAssessment(securityPath); err != nil {
		log.Printf("⚠️ Error initializing security risk assessment: %v", err)
	}

	// Initialize performance monitor
	metricsPath := filepath.Join(utils.Config.DataDir, "metrics", "performance.json")
	utils.InitPerformanceMonitor(metricsPath)
}

// setupBlockchainProcessor configures blockchain processor
func setupBlockchainProcessor(bc *blockchain.Blockchain) {
	// Start periodic processing of pending smart contracts
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				bc.ProcessPendingContracts()
			}
		}
	}()

	// Start periodic staking rewards distribution
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				bc.StakingPool.DistributeRewards()
				bc.StakingPool.ProcessExpiredStakes()
				bc.StakingPool.CalculateAPY()
			}
		}
	}()
}

// initializeStorage initializes the storage layer
func initializeStorage() (*blockchain.LevelDBStorage, error) {
	dbPath := filepath.Join(utils.Config.DataDir, "db")
	storage, err := blockchain.NewLevelDBStorage(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	return storage, nil
}

// setupRoutes configures all application routes
func setupRoutes(router *http.ServeMux, bc *blockchain.Blockchain, enhancedNetwork *network.EnhancedNetworkManager) {
	// Static files
	fs := http.FileServer(http.Dir("static"))
	router.Handle("/static/", http.StripPrefix("/static/", fs))

	// Default route: displays home page (acceuil.html)
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		tmpl, err := template.ParseFiles("templates/acceuil.html")
		if err != nil {
			http.Error(w, "Error loading home page", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	// Authentication routes
	router.HandleFunc("/login", handlers.LoginHandler)
	router.HandleFunc("/login-submit", handlers.LoginSubmitHandler)
	router.HandleFunc("/logout", handlers.LogoutHandler)
	router.HandleFunc("/register", handlers.RegisterHandler)
	router.HandleFunc("/register-submit", handlers.RegisterSubmitHandler)

	// Secure routes
	router.HandleFunc("/home", handlers.HomeHandler)
	router.HandleFunc("/profile", handlers.ProfileHandler)
	router.HandleFunc("/admin", handlers.AdminHandler(bc))

	// Blockchain routes
	router.HandleFunc("/blockchain", handlers.BlockchainHandler(bc))
	router.HandleFunc("/transactions", handlers.TransactionHandler(bc))
	router.HandleFunc("/wallets", handlers.WalletHandler("wallets"))
	router.HandleFunc("/stats", handlers.StatsHandler(bc))

	// Staking routes
	router.HandleFunc("/staking", handlers.StakingHandler(bc))
	router.HandleFunc("/staking/create", handlers.CreateStakeHandler(bc))
	router.HandleFunc("/staking/claim", handlers.ClaimRewardsHandler(bc))
	router.HandleFunc("/staking/unstake", handlers.UnstakeHandler(bc))
	router.HandleFunc("/staking/withdraw", handlers.WithdrawStakeHandler(bc))
	router.HandleFunc("/staking/validators", handlers.ValidatorsHandler(bc))
	router.HandleFunc("/staking/delegate", handlers.DelegateHandler(bc))

	// P2P routes
	router.HandleFunc("/p2p/", handlers.P2PHandler(bc))
	router.HandleFunc("/p2p/message", handlers.P2PHandler(bc))
	router.HandleFunc("/p2p/nodes", handlers.P2PHandler(bc))
	router.HandleFunc("/p2p/node", handlers.P2PHandler(bc))
	router.HandleFunc("/p2p/sync", handlers.P2PHandler(bc))

	// Add enhanced P2P handlers
	enhancedNetwork.ExtendHandlers(router)

	// Smart contract routes
	router.HandleFunc("/contracts", handlers.ContractUIHandler(bc))
	router.HandleFunc("/contract/", handlers.ContractUIHandler(bc))

	// Block detail routes
	router.HandleFunc("/block/", func(w http.ResponseWriter, r *http.Request) {
		// Extract block index from URL
		path := r.URL.Path
		if len(path) <= len("/block/") {
			http.Error(w, "Missing block index", http.StatusBadRequest)
			return
		}

		indexStr := path[len("/block/"):]
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			http.Error(w, "Invalid block index", http.StatusBadRequest)
			return
		}

		// Get the block
		block := bc.GetBlockByIndex(index)
		if block == nil {
			http.Error(w, "Block not found", http.StatusNotFound)
			return
		}

		// Render the template with block data
		tmpl, err := template.ParseFiles("templates/block_detail.html")
		if err != nil {
			http.Error(w, "Error loading template", http.StatusInternalServerError)
			return
		}

		// Format data for the template
		rawBlockData, _ := json.Marshal(block)
		latestBlockIndex := len(bc.Blocks) - 1

		// Calculate estimated mining time (for demo)
		miningTime := 30.0      // seconds (demo value)
		targetBlockTime := 60.0 // seconds (target)
		miningTimeRatio := miningTime / targetBlockTime

		// Estimate hash rate (for demo)
		estimatedHashRate := float64(block.Nonce) / miningTime

		data := map[string]interface{}{
			"Block":             block,
			"RawBlockData":      string(rawBlockData),
			"LatestBlockIndex":  latestBlockIndex,
			"BlockReward":       bc.MiningReward,
			"MiningTime":        miningTime,
			"TargetBlockTime":   targetBlockTime,
			"MiningTimeRatio":   miningTimeRatio,
			"EstimatedHashRate": estimatedHashRate,
		}

		tmpl.Execute(w, data)
	})

	// REST API routes
	router.HandleFunc("/api/", handlers.APIHandler(bc))
	router.HandleFunc("/api/contracts", handlers.ContractsHandler(bc))
	router.HandleFunc("/api/contracts/", handlers.ContractsHandler(bc))
	router.HandleFunc("/api/staking", handlers.StakingAPIHandler(bc))
	router.HandleFunc("/api/staking/", handlers.StakingAPIHandler(bc))
}

// setupGracefulShutdown sets up graceful shutdown
func setupGracefulShutdown() chan os.Signal {
	// Channel to receive shutdown signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	return stop
}

// generateServerKey generates a random server key
func generateServerKey() string {
	key := make([]byte, 32)
	rand.Read(key)
	return hex.EncodeToString(key)
}

func main() {
	startTime := time.Now()

	// Create necessary directories
	if err := initDirectories(); err != nil {
		log.Fatalf("❌ Error initializing directories: %v", err)
	}

	// Initialize configuration
	utils.InitializeConfig()

	// Configure logging
	if err := setupLogging(); err != nil {
		log.Fatalf("❌ Error configuring logging: %v", err)
	}
	defer utils.LogFile.Close()

	// Initialize audit and security systems
	setupAuditAndSecurity()

	// Initialize storage
	storage, err := initializeStorage()
	if err != nil {
		log.Fatalf("❌ Error initializing storage: %v", err)
	}
	defer storage.Close()

	// Initialize blockchain with new features
	bc := blockchain.NewBlockchain()
	bc.Storage = storage                         // Set storage
	bc.StakingPool = blockchain.NewStakingPool() // Initialize staking pool

	// Configure blockchain processor
	setupBlockchainProcessor(bc)

	// Initialize default users
	handlers.InitSampleUsers()

	// Initialize P2P network manager with enhanced features
	isValidator := true // This node is a validator
	port := utils.Config.ServerPort
	nodeURL := fmt.Sprintf("http://localhost:%d", port)
	enhancedNetwork := network.NewEnhancedNetworkManager(nodeURL, bc, isValidator)

	// Bootstrap the network with some seed nodes
	seedNodes := []string{
		fmt.Sprintf("http://localhost:%d", port), // This node
		"http://localhost:8081",                  // Potential other nodes
		"http://localhost:8082",
	}
	enhancedNetwork.Bootstrap(seedNodes)

	// Set up automatic refresh
	enhancedNetwork.AutoRefresh(10 * time.Minute)

	// Initialize router and configure routes
	router := http.NewServeMux()
	setupRoutes(router, bc, enhancedNetwork)

	// Configure graceful shutdown
	stopChan := setupGracefulShutdown()

	// Apply middlewares
	var handler http.Handler = router
	handler = utils.LoggingMiddleware(handler)
	handler = utils.RecoveryMiddleware(handler)
	handler = utils.SecurityHeadersMiddleware(handler)
	handler = utils.CORSMiddleware(handler)

	// Add rate limiting middleware if configured
	if utils.Config.EnableRateLimiting {
		handler = utils.RateLimiterMiddleware(handler)
	}

	// Log server start
	utils.LogAuditEvent(
		utils.EventTypeServerStarted,
		"system",
		"localhost",
		fmt.Sprintf("Server started on port %d", port),
		utils.RiskLow,
		map[string]interface{}{
			"startup_time_ms": time.Since(startTime).Milliseconds(),
			"port":            port,
			"validator":       isValidator,
			"version":         appVersion,
		},
	)

	// Automatically open browser if configured
	if utils.Config.AutoOpenBrowser {
		go func() {
			log.Println("🌍 Opening browser...")
			openBrowser(fmt.Sprintf("http://localhost:%d", port))
		}()
	}

	// Display startup information
	fmt.Printf("🚀 Server launched at: http://localhost:%d\n", port)
	fmt.Println("👤 Default users:")
	fmt.Println("   - Admin: admin/admin")
	fmt.Println("   - User:  user/user")
	fmt.Println("🎨 Enhanced UI with dark theme and animations")
	fmt.Println("🌐 Enhanced P2P network with Kademlia DHT")
	fmt.Println("📊 LevelDB persistent storage implemented")
	fmt.Println("💰 Token staking and validator system added")
	fmt.Printf("⏱️ Startup time: %v\n", time.Since(startTime))
	fmt.Printf("📊 Version %s | © 2025 CryptoChain Go\n", appVersion)

	// Start server in a separate goroutine
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
			log.Fatalf("❌ Error starting server: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-stopChan

	// Clean up
	fmt.Println("\n⏹️ Shutting down server...")

	// Log server shutdown
	utils.LogAuditEvent(
		utils.EventTypeServerStopped,
		"system",
		"localhost",
		"Clean server shutdown",
		utils.RiskLow,
		map[string]interface{}{
			"uptime_seconds": time.Since(startTime).Seconds(),
		},
	)

	// Allow time for pending writes
	time.Sleep(500 * time.Millisecond)

	fmt.Println("✅ Server shutdown successful.")
}
