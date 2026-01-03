package startup

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/herman-xphp/bukuo/internal/config"
)

// Version info (set via ldflags during build)
var (
	Version   = "1.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// Colors for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// PrintBanner prints a clean startup banner with system info
func PrintBanner(cfg *config.Config, dbConnected bool, configErrors, configWarnings int) {
	// Clear screen for fresh start (optional)
	fmt.Print("\033[H\033[2J")

	// ASCII Art Logo
	logo := `
    ____        __
   / __ )__  __/ /____  ____
  / __  / / / / //_/ / / / _ \
 / /_/ / /_/ / ,< / /_/ / (_) |
/_____/\__,_/_/|_|\__,_/\___/
`
	fmt.Print(colorCyan + logo + colorReset)

	// App info line
	fmt.Printf("%s%s Financial Accounting Platform %s\n", colorBold, colorPurple, colorReset)
	fmt.Printf("%sVersion %s%s\n\n", colorGray, Version, colorReset)

	// Status box
	printDivider("─", 55)

	// Server Info
	printSection("⚙️  SERVER")
	printKeyValue("Environment", formatEnv(cfg.Server.Env))
	printKeyValue("Port", cfg.Server.Port)
	printKeyValue("Status", formatStatus("Running", true))

	fmt.Println()

	// Database Info
	printSection("🗄️  DATABASE")
	printKeyValue("Host", fmt.Sprintf("%s:%s", cfg.Database.Host, cfg.Database.Port))
	printKeyValue("Database", cfg.Database.Name)
	printKeyValue("Status", formatStatus("Connected", dbConnected))

	fmt.Println()

	// System Info
	printSection("💻 SYSTEM")
	printKeyValue("Go Version", runtime.Version())
	printKeyValue("OS/Arch", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH))
	printKeyValue("CPUs", fmt.Sprintf("%d cores", runtime.NumCPU()))
	printKeyValue("Goroutines", fmt.Sprintf("%d", runtime.NumGoroutine()))

	fmt.Println()

	// Security Status
	printSection("🔐 SECURITY")
	printKeyValue("JWT Expiry", formatDuration(cfg.JWT.Expiry))
	printKeyValue("Rate Limit", fmt.Sprintf("%d req/min", cfg.RateLimit.GlobalPerMin))
	printKeyValue("CORS Origins", formatOrigins(cfg.Security.AllowedOrigins))

	fmt.Println()

	// Config Validation Summary
	printSection("📋 CONFIG VALIDATION")
	if configErrors == 0 && configWarnings == 0 {
		printKeyValue("Status", fmt.Sprintf("%s✓ All checks passed%s", colorGreen, colorReset))
	} else {
		if configErrors > 0 {
			printKeyValue("Errors", fmt.Sprintf("%s%d issue(s)%s", colorRed, configErrors, colorReset))
		}
		if configWarnings > 0 {
			printKeyValue("Warnings", fmt.Sprintf("%s%d warning(s)%s", colorYellow, configWarnings, colorReset))
		}
	}

	printDivider("─", 55)
	fmt.Println()

	// Quick Links
	fmt.Printf("%s📡 ENDPOINTS%s\n", colorBold, colorReset)
	baseURL := fmt.Sprintf("http://localhost:%s", cfg.Server.Port)
	printKeyValue("API", baseURL)
	printKeyValue("Swagger", baseURL+"/swagger/index.html")
	printKeyValue("Health", baseURL+"/health")

	printDivider("─", 55)

	// Ready message
	fmt.Printf("\n%s%s🚀 Server is ready to accept connections%s\n", colorBold, colorGreen, colorReset)
	fmt.Printf("%sStarted at %s%s\n\n", colorGray, time.Now().Format("2006-01-02 15:04:05"), colorReset)
}

// PrintShutdownBanner prints a clean shutdown message
func PrintShutdownBanner() {
	fmt.Println()
	printDivider("─", 55)
	fmt.Printf("%s%s⏳ Graceful shutdown initiated...%s\n", colorBold, colorYellow, colorReset)
}

// PrintShutdownComplete prints shutdown complete message
func PrintShutdownComplete() {
	fmt.Printf("%s%s✅ Server shutdown complete%s\n", colorBold, colorGreen, colorReset)
	printDivider("─", 55)
	fmt.Println()
}

// Helper functions
func printDivider(char string, length int) {
	fmt.Printf("%s%s%s\n", colorGray, strings.Repeat(char, length), colorReset)
}

func printSection(title string) {
	fmt.Printf("%s%s%s\n", colorBold, title, colorReset)
}

func printKeyValue(key, value string) {
	fmt.Printf("  %s%-14s%s %s\n", colorGray, key+":", colorReset, value)
}

func formatEnv(env string) string {
	switch env {
	case "production":
		return fmt.Sprintf("%s%s PRODUCTION %s", colorRed, colorBold, colorReset)
	case "staging":
		return fmt.Sprintf("%s%s STAGING %s", colorYellow, colorBold, colorReset)
	case "development":
		return fmt.Sprintf("%s%s DEVELOPMENT %s", colorGreen, colorBold, colorReset)
	default:
		return fmt.Sprintf("%s%s %s %s", colorGray, colorBold, strings.ToUpper(env), colorReset)
	}
}

func formatStatus(label string, ok bool) string {
	if ok {
		return fmt.Sprintf("%s● %s%s", colorGreen, label, colorReset)
	}
	return fmt.Sprintf("%s● Failed%s", colorRed, colorReset)
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	if hours >= 24 {
		days := hours / 24
		return fmt.Sprintf("%d day(s)", days)
	}
	return fmt.Sprintf("%d hour(s)", hours)
}

func formatOrigins(origins []string) string {
	if len(origins) == 0 {
		return fmt.Sprintf("%sNone%s", colorRed, colorReset)
	}
	if len(origins) == 1 {
		return origins[0]
	}
	if len(origins) <= 3 {
		return strings.Join(origins, ", ")
	}
	return fmt.Sprintf("%s (+%d more)", origins[0], len(origins)-1)
}

// GetMemoryUsage returns current memory usage in MB
func GetMemoryUsage() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024
}

// GetHostname returns the hostname
func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
