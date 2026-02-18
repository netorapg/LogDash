package parser

import (
	"bytes"
	"regexp"
	"strings"
	"time"
)

// pythonLoggingParser parseia logs gerados pela biblioteca logging do Python
// Formato: [LEVEL] YYYY-MM-DD HH:MM:SS,mmm module:line function context: message
type pythonLoggingParser struct {
	// Regex para detectar o padrão completo
	pattern *regexp.Regexp
}

// NewPythonLoggingParser cria uma nova instância do parser Python Logging
func NewPythonLoggingParser() Parser {
	// Padrão: [LEVEL] timestamp module:line function context: message
	pattern := regexp.MustCompile(`^\[([A-Z]+)\]\s+(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2},\d{3})\s+(\S+):(\d+)\s+(\S+)\s+(\S+):\s+(.+)$`)
	
	return &pythonLoggingParser{
		pattern: pattern,
	}
}

// Name retorna o identificador deste parser
func (p *pythonLoggingParser) Name() string {
	return "python-logging"
}

// SupportsFormat analisa a amostra e retorna score de confiança
func (p *pythonLoggingParser) SupportsFormat(sample []byte) float64 {
	if len(sample) == 0 {
		return 0.0
	}

	lines := bytes.Split(sample, []byte("\n"))
	if len(lines) == 0 {
		return 0.0
	}

	matches := 0
	validLines := 0

	for _, line := range lines {
		lineStr := strings.TrimSpace(string(line))
		if len(lineStr) == 0 {
			continue
		}
		
		validLines++

		// Verificar características principais do formato Python logging
		// 1. Começa com [LEVEL]
		if !strings.HasPrefix(lineStr, "[") {
			continue
		}

		// 2. Tem timestamp com vírgula (milissegundos)
		if !strings.Contains(lineStr, ",") {
			continue
		}

		// 3. Tem padrão module:line
		if !strings.Contains(lineStr, ":") {
			continue
		}

		// 4. Tenta match completo com regex
		if p.pattern.MatchString(lineStr) {
			matches++
		}
	}

	if validLines == 0 {
		return 0.1
	}

	// Calcula porcentagem de linhas que batem o padrão
	matchRate := float64(matches) / float64(validLines)

	// Score baseado na taxa de match
	if matchRate >= 0.8 {
		return 0.95 // Alta confiança
	} else if matchRate >= 0.5 {
		return 0.7 // Média-alta confiança
	} else if matchRate >= 0.3 {
		return 0.5 // Média confiança
	} else if matchRate > 0 {
		return 0.3 // Baixa confiança
	}

	return 0.1 // Muito baixa confiança
}

// Parse processa o conteúdo e extrai as entradas de log
func (p *pythonLoggingParser) Parse(content []byte) ([]LogEntry, error) {
	if len(content) == 0 {
		return []LogEntry{}, nil
	}

	var entries []LogEntry
	lines := splitLines(content)

	for i, line := range lines {
		// Pula linhas vazias
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		entry := p.parseLine(line, i+1)
		entries = append(entries, entry)
	}

	return entries, nil
}

// parseLine parseia uma linha individual do log
func (p *pythonLoggingParser) parseLine(line string, lineNumber int) LogEntry {
	entry := LogEntry{
		LineNumber: lineNumber,
		Raw:        line,
		Level:      LevelUnknown,
		Metadata:   make(map[string]interface{}),
	}

	// Tenta fazer match com o padrão completo
	matches := p.pattern.FindStringSubmatch(line)
	
	if len(matches) == 8 {
		// Match completo - extrair todos os campos
		// matches[0] = linha completa
		// matches[1] = level
		// matches[2] = timestamp
		// matches[3] = module
		// matches[4] = line number
		// matches[5] = function
		// matches[6] = context
		// matches[7] = message

		entry.Level = parsePythonLogLevel(matches[1])
		entry.Timestamp = parsePythonTimestamp(matches[2])
		entry.Message = matches[7]

		// Metadados extras
		entry.Metadata["module"] = matches[3]
		entry.Metadata["line"] = matches[4]
		entry.Metadata["function"] = matches[5]
		entry.Metadata["context"] = matches[6]
	} else {
		// Linha não bateu com padrão completo - tentar extração parcial
		entry.Level = detectLevel(line)
		entry.Message = line

		// Tenta pelo menos extrair timestamp se houver
		if ts := extractPythonTimestamp(line); !ts.IsZero() {
			entry.Timestamp = ts
		}
	}

	return entry
}

// parsePythonLogLevel converte string de nível para LogLevel
func parsePythonLogLevel(levelStr string) LogLevel {
	switch strings.ToUpper(levelStr) {
	case "TRACE":
		return LevelTrace
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARNING", "WARN":
		return LevelWarning
	case "ERROR":
		return LevelError
	case "CRITICAL", "FATAL":
		return LevelFatal
	default:
		return LevelUnknown
	}
}

// parsePythonTimestamp parseia timestamp no formato Python logging
// Formato: 2026-02-13 14:23:42,181
func parsePythonTimestamp(timestampStr string) time.Time {
	// Python usa vírgula para milissegundos, Go usa ponto
	// Converter vírgula para ponto
	timestampStr = strings.Replace(timestampStr, ",", ".", 1)
	
	// Formato Go: 2006-01-02 15:04:05.000
	format := "2006-01-02 15:04:05.000"
	
	t, err := time.Parse(format, timestampStr)
	if err != nil {
		return time.Time{}
	}
	
	return t
}

// extractPythonTimestamp tenta extrair timestamp de qualquer parte da linha
func extractPythonTimestamp(line string) time.Time {
	// Procura por padrão de timestamp Python
	timestampPattern := regexp.MustCompile(`\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2},\d{3}`)
	match := timestampPattern.FindString(line)
	
	if match != "" {
		return parsePythonTimestamp(match)
	}
	
	return time.Time{}
}