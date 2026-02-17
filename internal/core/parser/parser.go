package parser

import "time"

// LogLevel representa o nível de severidade de um log
type LogLevel int

const (
	LevelUnknown LogLevel = iota
	LevelTrace
	LevelDebug
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
)

// String retorna a representação em string do LogLevel
func (l LogLevel) String() string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry representa uma entrada de log processada
type LogEntry struct {
	Timestamp  time.Time              // quando o log aconteceu
	Level      LogLevel               // severidade
	Message    string                 // mensagem principal
	Source     string                 // arquivo de origem
	LineNumber int                    // linha no arquivo
	Raw        string                 // linha original completa
	Metadata   map[string]interface{} // campos adicionais específicos do formato
}

// Parser define a interface que todos os parsers devem implementar
type Parser interface {
	// Parse processa o conteúdo e retorna as entradas de log encontradas
	Parse(content []byte) ([]LogEntry, error)

	// SupportsFormat retorna um score de confiança (0.0 a 1.0) indicando
	// quão bem este parser consegue processar o formato fornecido
	SupportsFormat(sample []byte) float64

	// Name retorna o nome identificador deste parser
	Name() string
}