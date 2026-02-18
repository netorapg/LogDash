package parser

import (
	"bytes"
	"strings"
	"time"
)

// genericParser é um parser fallback que tenta extrair informações básicas
// de qualquer formato de log, mesmo que não seja estruturado
type genericParser struct{}

// NewGenericParser cria uma nova instância do parser genérico
func NewGenericParser() Parser {
	return &genericParser{}
}

// Name retorna o identificador deste parser
func (p *genericParser) Name() string {
	return "generic"
}

// SupportsFormat retorna um score baixo a médio, já que este parser
// é um fallback e deve ter menor prioridade que parsers especializados
func (p *genericParser) SupportsFormat(sample []byte) float64 {
	if len(sample) == 0 {
		return 0.1
	}

	// Se tem múltiplas linhas, aumenta um pouco a confiança
	lines := bytes.Count(sample, []byte("\n"))
	if lines > 0 {
		return 0.3
	}

	// Conteúdo não vazio tem pelo menos algum suporte
	return 0.2
}

// Parse processa o conteúdo linha por linha, tentando extrair
// informações básicas como timestamp e nível
func (p *genericParser) Parse(content []byte) ([]LogEntry, error) {
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

		entry := LogEntry{
			LineNumber: i + 1,
			Raw:        line,
			Message:    line,
			Level:      detectLevel(line),
			Timestamp:  detectTimestamp(line),
			Metadata:   make(map[string]interface{}),
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// splitLines divide o conteúdo em linhas, lidando com diferentes terminadores
func splitLines(content []byte) []string {
	// Normaliza line endings para \n
	normalized := bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	normalized = bytes.ReplaceAll(normalized, []byte("\r"), []byte("\n"))

	text := string(normalized)
	return strings.Split(text, "\n")
}

// detectLevel tenta identificar o nível do log através de palavras-chave
func detectLevel(line string) LogLevel {
	upperLine := strings.ToUpper(line)

	// Ordem importa: verificamos do mais crítico para o menos
	if strings.Contains(upperLine, "FATAL") {
		return LevelFatal
	}
	if strings.Contains(upperLine, "ERROR") || strings.Contains(upperLine, "ERR") {
		return LevelError
	}
	if strings.Contains(upperLine, "WARN") {
		return LevelWarning
	}
	if strings.Contains(upperLine, "INFO") {
		return LevelInfo
	}
	if strings.Contains(upperLine, "DEBUG") {
		return LevelDebug
	}
	if strings.Contains(upperLine, "TRACE") {
		return LevelTrace
	}

	return LevelUnknown
}

// detectTimestamp tenta encontrar um timestamp no início da linha
// usando formatos comuns
func detectTimestamp(line string) time.Time {
	// Formatos comuns de timestamp em logs
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.000",
		"2006-01-02T15:04:05.000",
		"2006/01/02 15:04:05",
		"Jan 02 15:04:05",
		"02/Jan/2006:15:04:05",
	}

	// Tenta cada formato nos primeiros 30 caracteres da linha
	prefix := line
	if len(line) > 30 {
		prefix = line[:30]
	}

	for _, format := range formats {
		if t, err := time.Parse(format, prefix[:min(len(prefix), len(format))]); err == nil {
			return t
		}
	}

	// Se não encontrou timestamp, retorna zero value
	return time.Time{}
}

// min retorna o menor de dois inteiros
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}