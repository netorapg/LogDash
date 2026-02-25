package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/netorapg/LogDash/internal/core/aggregator"
	"github.com/netorapg/LogDash/internal/core/parser"
	"github.com/netorapg/LogDash/internal/service"
	"github.com/netorapg/LogDash/internal/tui"
	"github.com/netorapg/LogDash/internal/web/server"
)

const version = "0.1.0"

func main() {
	// Definir subcomandos
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "analyze":
		analyzeCommand()
	case "tui":
		tuiCommand()
	case "web":
		webCommand()
	case "version":
		versionCommand()
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func analyzeCommand() {
	// Definir flags para analyze
	analyzeFlags := flag.NewFlagSet("analyze", flag.ExitOnError)
	
	path := analyzeFlags.String("path", ".", "Root path to search for logs")
	maxFiles := analyzeFlags.Int("max-files", 0, "Maximum number of files to process (0 = no limit)")
	maxDepth := analyzeFlags.Int("depth", 10, "Maximum directory depth to search")
	patterns := analyzeFlags.String("patterns", "*.log,*.txt", "File patterns to match (comma-separated)")
	last := analyzeFlags.String("last", "", "Analyze logs from last duration (e.g., '24h', '7d')")
	verbose := analyzeFlags.Bool("verbose", false, "Verbose output")
	noAnomalies := analyzeFlags.Bool("no-anomalies", false, "Disable anomaly detection")
	topMessages := analyzeFlags.Int("top", 10, "Number of top messages to show")

	analyzeFlags.Parse(os.Args[2:])

	// Validar path
	if *path == "" {
		fmt.Fprintln(os.Stderr, "Error: path cannot be empty")
		os.Exit(1)
	}

	// Criar service
	svc := service.NewLogDashService()
	opts := service.DefaultAnalyzeOptions()
	
	opts.RootPath = *path
	opts.MaxFilesToProcess = *maxFiles
	opts.DiscoverOpts.MaxDepth = *maxDepth
	
	// Configurar patterns
	if *patterns != "" {
		patternList := strings.Split(*patterns, ",")
		for i := range patternList {
			patternList[i] = strings.TrimSpace(patternList[i])
		}
		opts.DiscoverOpts.FilePatterns = patternList
	}

	// Configurar time range se especificado
	if *last != "" {
		duration, err := parseDuration(*last)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing duration '%s': %v\n", *last, err)
			os.Exit(1)
		}
		opts.AggregateOpts.TimeRange = aggregator.TimeRange{
			Start: time.Now().Add(-duration),
			End:   time.Now(),
		}
	}

	// Configurar anomaly detection
	opts.AggregateOpts.DetectAnomalies = !*noAnomalies
	
	// Configurar top messages
	opts.AggregateOpts.TopMessagesLimit = *topMessages

	// Executar análise
	fmt.Printf("🔍 Analyzing logs in: %s\n\n", *path)
	
	result, err := svc.AnalyzeLogs(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Analysis failed: %v\n", err)
		os.Exit(1)
	}

	// Exibir resultados
	printResults(result, *verbose)
}

func tuiCommand() {
	// Definir flags para tui
	tuiFlags := flag.NewFlagSet("tui", flag.ExitOnError)
	
	path := tuiFlags.String("path", ".", "Root path to search for logs")
	maxFiles := tuiFlags.Int("max-files", 0, "Maximum number of files to process (0 = no limit)")
	maxDepth := tuiFlags.Int("depth", 10, "Maximum directory depth to search")
	patterns := tuiFlags.String("patterns", "*.log,*.txt", "File patterns to match (comma-separated)")
	last := tuiFlags.String("last", "", "Analyze logs from last duration (e.g., '24h', '7d')")
	noAnomalies := tuiFlags.Bool("no-anomalies", false, "Disable anomaly detection")

	tuiFlags.Parse(os.Args[2:])

	// Validar path
	if *path == "" {
		fmt.Fprintln(os.Stderr, "Error: path cannot be empty")
		os.Exit(1)
	}

	// Preparar opções
	opts := service.DefaultAnalyzeOptions()
	opts.RootPath = *path
	opts.MaxFilesToProcess = *maxFiles
	opts.DiscoverOpts.MaxDepth = *maxDepth
	
	// Configurar patterns
	if *patterns != "" {
		patternList := strings.Split(*patterns, ",")
		for i := range patternList {
			patternList[i] = strings.TrimSpace(patternList[i])
		}
		opts.DiscoverOpts.FilePatterns = patternList
	}

	// Configurar time range se especificado
	if *last != "" {
		duration, err := parseDuration(*last)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing duration '%s': %v\n", *last, err)
			os.Exit(1)
		}
		opts.AggregateOpts.TimeRange = aggregator.TimeRange{
			Start: time.Now().Add(-duration),
			End:   time.Now(),
		}
	}

	// Configurar anomaly detection
	opts.AggregateOpts.DetectAnomalies = !*noAnomalies

	// Iniciar TUI
	model := tui.NewModel(*path, opts)
	p := tea.NewProgram(model, tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func webCommand() {
	// Definir flags para web
	webFlags := flag.NewFlagSet("web", flag.ExitOnError)
	
	path := webFlags.String("path", ".", "Root path to search for logs")
	port := webFlags.Int("port", 8080, "Port to run web server")

	webFlags.Parse(os.Args[2:])

	// Validar path
	if *path == "" {
		fmt.Fprintln(os.Stderr, "Error: path cannot be empty")
		os.Exit(1)
	}

	// Iniciar servidor web
	server := web.NewServer(*path, *port)
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting web server: %v\n", err)
		os.Exit(1)
	}
}

func versionCommand() {
	fmt.Printf("LogDash v%s\n", version)
	fmt.Println("A fast and intelligent log analysis tool")
}

func printUsage() {
	fmt.Println("LogDash - Fast and intelligent log analysis")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  logdash <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  analyze    Analyze log files (command-line output)")
	fmt.Println("  tui        Interactive terminal UI")
	fmt.Println("  web        Web dashboard UI")
	fmt.Println("  version    Show version information")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Analyze Options:")
	fmt.Println("  -path string")
	fmt.Println("        Root path to search for logs (default: current directory)")
	fmt.Println("  -max-files int")
	fmt.Println("        Maximum number of files to process (default: no limit)")
	fmt.Println("  -depth int")
	fmt.Println("        Maximum directory depth to search (default: 10)")
	fmt.Println("  -patterns string")
	fmt.Println("        File patterns to match, comma-separated (default: '*.log,*.txt')")
	fmt.Println("  -last string")
	fmt.Println("        Analyze logs from last duration (e.g., '24h', '7d', '30m')")
	fmt.Println("  -top int")
	fmt.Println("        Number of top messages to show (default: 10)")
	fmt.Println("  -no-anomalies")
	fmt.Println("        Disable anomaly detection")
	fmt.Println("  -verbose")
	fmt.Println("        Verbose output")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Command-line analysis")
	fmt.Println("  logdash analyze")
	fmt.Println("  logdash analyze -path /var/log")
	fmt.Println("  logdash analyze -path /var/log -last 24h")
	fmt.Println("  logdash analyze -path ./logs -max-files 5 -verbose")
	fmt.Println("  logdash analyze -path /app/logs -patterns '*.log' -depth 3")
	fmt.Println()
	fmt.Println("  # Interactive TUI")
	fmt.Println("  logdash tui")
	fmt.Println("  logdash tui -path /var/log")
	fmt.Println("  logdash tui -path /var/log -last 24h")
	fmt.Println()
	fmt.Println("  # Web Dashboard")
	fmt.Println("  logdash web")
	fmt.Println("  logdash web -path /var/log")
	fmt.Println("  logdash web -path /var/log -port 3000")
}

func printResults(result *service.AnalysisResult, verbose bool) {
	// Header
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📊 ANALYSIS RESULTS")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Summary
	fmt.Println("📝 Summary:")
	fmt.Printf("   %s\n", result.Summary)
	fmt.Println()

	// Files
	fmt.Println("📁 Files:")
	fmt.Printf("   Discovered: %d\n", result.FilesDiscovered)
	fmt.Printf("   Processed:  %d\n", result.FilesProcessed)
	fmt.Printf("   Entries:    %d\n", result.EntriesParsed)
	fmt.Println()

	// Statistics by Level
	fmt.Println("📈 By Level:")
	printLevelStats(result.ByLevel)
	fmt.Println()

	// Top Messages
	if len(result.TopMessages) > 0 {
		fmt.Println("🔝 Top Messages:")
		for i, msg := range result.TopMessages {
			if i >= 10 {
				break
			}
			icon := getLevelIcon(msg.Level)
			fmt.Printf("   %d. %s [%s] %s (x%d)\n", 
				i+1, icon, msg.Level, truncate(msg.Message, 60), msg.Count)
		}
		fmt.Println()
	}

	// Anomalies
	if len(result.Anomalies) > 0 {
		fmt.Println("⚠️  Anomalies Detected:")
		for _, anomaly := range result.Anomalies {
			severity := getSeverityIcon(anomaly.Severity)
			fmt.Printf("   %s [%s] %s\n", severity, anomaly.Type, anomaly.Description)
		}
		fmt.Println()
	}

	// Verbose: File Results
	if verbose && len(result.FileResults) > 0 {
		fmt.Println("📄 File Details:")
		for _, fr := range result.FileResults {
			status := "✓"
			if fr.Error != nil {
				status = "✗"
			}
			fmt.Printf("   %s %-50s %6d entries %4dms [%s]\n",
				status,
				truncate(filepath.Base(fr.FilePath), 50),
				fr.EntriesFound,
				fr.ProcessingTime,
				fr.ParserUsed)
		}
		fmt.Println()
	}

	// Errors
	if len(result.Errors) > 0 {
		fmt.Println("❌ Errors:")
		for _, e := range result.Errors {
			fmt.Printf("   [%s] %s: %v\n", e.Stage, filepath.Base(e.FilePath), e.Error)
		}
		fmt.Println()
	}

	// Footer
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printLevelStats(byLevel map[parser.LogLevel]int) {
	levels := []struct {
		level parser.LogLevel
		name  string
		icon  string
	}{
		{parser.LevelFatal, "Fatal", "💀"},
		{parser.LevelError, "Error", "❌"},
		{parser.LevelWarning, "Warning", "⚠️ "},
		{parser.LevelInfo, "Info", "ℹ️ "},
		{parser.LevelDebug, "Debug", "🐛"},
		{parser.LevelTrace, "Trace", "🔍"},
	}

	for _, l := range levels {
		count := byLevel[l.level]
		if count > 0 {
			fmt.Printf("   %s %-8s %6d\n", l.icon, l.name+":", count)
		}
	}
}

func getLevelIcon(level parser.LogLevel) string {
	switch level {
	case parser.LevelFatal:
		return "💀"
	case parser.LevelError:
		return "❌"
	case parser.LevelWarning:
		return "⚠️"
	case parser.LevelInfo:
		return "ℹ️"
	case parser.LevelDebug:
		return "🐛"
	case parser.LevelTrace:
		return "🔍"
	default:
		return "  "
	}
}

func getSeverityIcon(severity float64) string {
	if severity >= 0.8 {
		return "🚨"
	} else if severity >= 0.5 {
		return "⚠️"
	}
	return "ℹ️"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func parseDuration(s string) (time.Duration, error) {
	// Suportar formatos como "24h", "7d", "30m"
	s = strings.TrimSpace(s)
	
	if strings.HasSuffix(s, "d") {
		days := s[:len(s)-1]
		return time.ParseDuration(days + "h")
	}
	
	return time.ParseDuration(s)
}