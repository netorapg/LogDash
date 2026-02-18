package parser

import (
	"strings"
	"testing"
)

func TestPythonLoggingParser_Name(t *testing.T) {
	parser := NewPythonLoggingParser()
	if parser.Name() != "python-logging" {
		t.Errorf("Name() = %v, want %v", parser.Name(), "python-logging")
	}
}

func TestPythonLoggingParser_SupportsFormat(t *testing.T) {
	tests := []struct {
		name    string
		sample  string
		wantMin float64
		wantMax float64
	}{
		{
			name: "typical python logging format",
			sample: `[INFO] 2026-02-13 14:23:42,181 produtos_sem_giro:143 produtos_sem_giro produtos_sem_giro__BIAMercadoCondominio: Iniciando processo
[INFO] 2026-02-13 14:23:42,181 database:29 __init__ produtos_sem_giro__BIAMercadoCondominio: Inicializando Database`,
			wantMin: 0.85,
			wantMax: 1.0,
		},
		{
			name: "single valid line",
			sample: "[ERROR] 2026-02-13 14:23:42,181 module:123 function context: Error message",
			wantMin: 0.8,
			wantMax: 1.0,
		},
		{
			name: "multiple levels",
			sample: `[INFO] 2026-02-13 14:23:42,181 module:1 func ctx: Info
[WARNING] 2026-02-13 14:23:43,000 module:2 func ctx: Warning
[ERROR] 2026-02-13 14:23:44,000 module:3 func ctx: Error`,
			wantMin: 0.85,
			wantMax: 1.0,
		},
		{
			name:    "not python logging - missing brackets",
			sample:  "INFO 2026-02-13 14:23:42,181 module:123 function context: Message",
			wantMin: 0.0,
			wantMax: 0.3,
		},
		{
			name:    "not python logging - wrong timestamp format",
			sample:  "[INFO] 2026-02-13 14:23:42 module:123 function context: Message",
			wantMin: 0.0,
			wantMax: 0.5,
		},
		{
			name:    "generic log format",
			sample:  "2026-02-13 ERROR Something went wrong",
			wantMin: 0.0,
			wantMax: 0.3,
		},
		{
			name:    "empty content",
			sample:  "",
			wantMin: 0.0,
			wantMax: 0.1,
		},
	}

	parser := NewPythonLoggingParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := parser.SupportsFormat([]byte(tt.sample))
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("SupportsFormat() = %v, want between %v and %v", score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestPythonLoggingParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantCount int
		wantErr  bool
		validate func(t *testing.T, entries []LogEntry)
	}{
		{
			name:      "empty content",
			input:     "",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "single valid line",
			input:     "[INFO] 2026-02-13 14:23:42,181 module:143 function context: Test message",
			wantCount: 1,
			wantErr:   false,
			validate: func(t *testing.T, entries []LogEntry) {
				e := entries[0]
				if e.Level != LevelInfo {
					t.Errorf("Level = %v, want %v", e.Level, LevelInfo)
				}
				if e.Message != "Test message" {
					t.Errorf("Message = %v, want %v", e.Message, "Test message")
				}
				if e.Timestamp.IsZero() {
					t.Error("Timestamp should be parsed")
				}
				if e.Timestamp.Year() != 2026 {
					t.Errorf("Year = %v, want 2026", e.Timestamp.Year())
				}
				if e.LineNumber != 1 {
					t.Errorf("LineNumber = %v, want 1", e.LineNumber)
				}
			},
		},
		{
			name: "multiple lines with different levels",
			input: `[INFO] 2026-02-13 14:23:42,181 module:1 func ctx: Info message
[WARNING] 2026-02-13 14:23:43,000 module:2 func ctx: Warning message
[ERROR] 2026-02-13 14:23:44,000 module:3 func ctx: Error message
[DEBUG] 2026-02-13 14:23:45,000 module:4 func ctx: Debug message`,
			wantCount: 4,
			wantErr:   false,
			validate: func(t *testing.T, entries []LogEntry) {
				levels := []LogLevel{LevelInfo, LevelWarning, LevelError, LevelDebug}
				for i, entry := range entries {
					if entry.Level != levels[i] {
						t.Errorf("Entry %d Level = %v, want %v", i, entry.Level, levels[i])
					}
				}
			},
		},
		{
			name:      "real log from file",
			input:     "[INFO] 2026-02-13 14:23:42,181 produtos_sem_giro:143 produtos_sem_giro produtos_sem_giro__BIAMercadoCondominio: Iniciando processo _BIAMercadoCondominio",
			wantCount: 1,
			wantErr:   false,
			validate: func(t *testing.T, entries []LogEntry) {
				e := entries[0]
				if e.Level != LevelInfo {
					t.Errorf("Level = %v, want INFO", e.Level)
				}
				if !strings.Contains(e.Message, "Iniciando processo") {
					t.Errorf("Message should contain 'Iniciando processo', got: %v", e.Message)
				}
				
				// Verificar metadata
				if e.Metadata["module"] != "produtos_sem_giro" {
					t.Errorf("Metadata[module] = %v, want produtos_sem_giro", e.Metadata["module"])
				}
				if e.Metadata["line"] != "143" {
					t.Errorf("Metadata[line] = %v, want 143", e.Metadata["line"])
				}
				if e.Metadata["function"] != "produtos_sem_giro" {
					t.Errorf("Metadata[function] = %v, want produtos_sem_giro", e.Metadata["function"])
				}
				if e.Metadata["context"] != "produtos_sem_giro__BIAMercadoCondominio" {
					t.Errorf("Metadata[context] = %v, want produtos_sem_giro__BIAMercadoCondominio", e.Metadata["context"])
				}
			},
		},
		{
			name: "lines with special characters and encoding",
			input: `[INFO] 2026-02-13 14:23:43,377 database:62 __init__ produtos_sem_giro__BIAMercadoCondominio: 10 conexões criadas com sucesso
[INFO] 2026-02-13 14:24:03,294 funcoes_principais:197 update_temp produtos_sem_giro__BIAMercadoCondominio: Updates na CadProduto e CadProdutoLojas realizados`,
			wantCount: 2,
			wantErr:   false,
			validate: func(t *testing.T, entries []LogEntry) {
				if !strings.Contains(entries[0].Message, "conexões") && !strings.Contains(entries[0].Message, "conexÃµes") {
					t.Errorf("Message should contain 'conexões', got: %v", entries[0].Message)
				}
			},
		},
		{
			name: "empty lines should be skipped",
			input: `[INFO] 2026-02-13 14:23:42,181 module:1 func ctx: First

[INFO] 2026-02-13 14:23:43,000 module:2 func ctx: Second`,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "malformed line mixed with valid",
			input: `[INFO] 2026-02-13 14:23:42,181 module:1 func ctx: Valid
This is not a valid python log line
[ERROR] 2026-02-13 14:23:43,000 module:2 func ctx: Also valid`,
			wantCount: 3, // O parser deve lidar gracefully com linhas malformadas
			wantErr:   false,
		},
		{
			name:      "timestamp with milliseconds parsed correctly",
			input:     "[INFO] 2026-02-13 14:23:42,999 module:1 func ctx: Message",
			wantCount: 1,
			wantErr:   false,
			validate: func(t *testing.T, entries []LogEntry) {
				// Verificar que milissegundos foram parseados (999ms)
				if entries[0].Timestamp.Nanosecond() < 900000000 { // ~999ms em nanosegundos
					t.Errorf("Milliseconds not parsed correctly: %v", entries[0].Timestamp)
				}
			},
		},
		{
			name:      "all log levels",
			input: `[TRACE] 2026-02-13 14:23:42,000 m:1 f c: Trace
[DEBUG] 2026-02-13 14:23:42,000 m:2 f c: Debug
[INFO] 2026-02-13 14:23:42,000 m:3 f c: Info
[WARNING] 2026-02-13 14:23:42,000 m:4 f c: Warning
[WARN] 2026-02-13 14:23:42,000 m:5 f c: Warn
[ERROR] 2026-02-13 14:23:42,000 m:6 f c: Error
[CRITICAL] 2026-02-13 14:23:42,000 m:7 f c: Critical
[FATAL] 2026-02-13 14:23:42,000 m:8 f c: Fatal`,
			wantCount: 8,
			wantErr:   false,
			validate: func(t *testing.T, entries []LogEntry) {
				expectedLevels := []LogLevel{
					LevelTrace, LevelDebug, LevelInfo, LevelWarning, 
					LevelWarning, LevelError, LevelFatal, LevelFatal,
				}
				for i, entry := range entries {
					if entry.Level != expectedLevels[i] {
						t.Errorf("Entry %d Level = %v, want %v", i, entry.Level, expectedLevels[i])
					}
				}
			},
		},
	}

	parser := NewPythonLoggingParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, err := parser.Parse([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(entries) != tt.wantCount {
				t.Errorf("Parse() returned %v entries, want %v", len(entries), tt.wantCount)
				return
			}

			if tt.validate != nil {
				tt.validate(t, entries)
			}
		})
	}
}

func TestPythonLoggingParser_Parse_EdgeCases(t *testing.T) {
	parser := NewPythonLoggingParser()

	t.Run("very long message", func(t *testing.T) {
		longMsg := strings.Repeat("a", 5000)
		input := "[INFO] 2026-02-13 14:23:42,181 module:1 func ctx: " + longMsg
		entries, err := parser.Parse([]byte(input))
		if err != nil {
			t.Errorf("Parse() should handle long messages, got error: %v", err)
		}
		if len(entries) != 1 {
			t.Errorf("Parse() returned %v entries, want 1", len(entries))
		}
		if !strings.Contains(entries[0].Message, "aaa") {
			t.Error("Long message not preserved")
		}
	})

	t.Run("colon in message", func(t *testing.T) {
		input := "[INFO] 2026-02-13 14:23:42,181 module:1 func ctx: Message with: multiple: colons"
		entries, err := parser.Parse([]byte(input))
		if err != nil {
			t.Errorf("Parse() error: %v", err)
		}
		if entries[0].Message != "Message with: multiple: colons" {
			t.Errorf("Message = %v, colons should be preserved", entries[0].Message)
		}
	})

	t.Run("multiline message (continuation)", func(t *testing.T) {
		// Python logging às vezes tem stack traces em múltiplas linhas
		input := `[ERROR] 2026-02-13 14:23:42,181 module:1 func ctx: Error occurred
Traceback line 1
Traceback line 2`
		entries, err := parser.Parse([]byte(input))
		if err != nil {
			t.Errorf("Parse() error: %v", err)
		}
		// Deve parsear cada linha separadamente por enquanto
		if len(entries) < 1 {
			t.Error("Should parse at least the main error line")
		}
	})
}