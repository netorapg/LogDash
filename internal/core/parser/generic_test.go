package parser

import (
	"strings"
	"testing"
)

func TestGenericParser_Name(t *testing.T) {
	parser := NewGenericParser()
	if parser.Name() != "generic" {
		t.Errorf("Name() = %v, want %v", parser.Name(), "generic")
	}
}

func TestGenericParser_SupportsFormat(t *testing.T) {
	tests := []struct {
		name    string
		sample  string
		wantMin float64 // score mínimo esperado
		wantMax float64 // score máximo esperado
	}{
		{
			name:    "any text should have some support",
			sample:  "This is just plain text\nAnother line",
			wantMin: 0.1,
			wantMax: 0.5,
		},
		{
			name:    "empty content should have low support",
			sample:  "",
			wantMin: 0.0,
			wantMax: 0.2,
		},
		{
			name:    "structured log should have lower support than specialized parser",
			sample:  "[2024-01-15 10:30:00] ERROR - Connection failed",
			wantMin: 0.2,
			wantMax: 0.6,
		},
	}

	parser := NewGenericParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := parser.SupportsFormat([]byte(tt.sample))
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("SupportsFormat() = %v, want between %v and %v", score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestGenericParser_Parse(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantCount  int
		wantErr    bool
		validate   func(t *testing.T, entries []LogEntry)
	}{
		{
			name:      "empty content",
			input:     "",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "single line",
			input:     "This is a log line",
			wantCount: 1,
			wantErr:   false,
			validate: func(t *testing.T, entries []LogEntry) {
				if entries[0].Message != "This is a log line" {
					t.Errorf("Message = %v, want %v", entries[0].Message, "This is a log line")
				}
				if entries[0].Raw != "This is a log line" {
					t.Errorf("Raw = %v, want %v", entries[0].Raw, "This is a log line")
				}
				if entries[0].Level != LevelUnknown {
					t.Errorf("Level = %v, want %v", entries[0].Level, LevelUnknown)
				}
			},
		},
		{
			name:      "multiple lines",
			input:     "First line\nSecond line\nThird line",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "lines with different levels detected by keywords",
			input:     "ERROR: Something went wrong\nWARNING: Be careful\nINFO: All good",
			wantCount: 3,
			wantErr:   false,
			validate: func(t *testing.T, entries []LogEntry) {
				if entries[0].Level != LevelError {
					t.Errorf("First entry Level = %v, want %v", entries[0].Level, LevelError)
				}
				if entries[1].Level != LevelWarning {
					t.Errorf("Second entry Level = %v, want %v", entries[1].Level, LevelWarning)
				}
				if entries[2].Level != LevelInfo {
					t.Errorf("Third entry Level = %v, want %v", entries[2].Level, LevelInfo)
				}
			},
		},
		{
			name:      "line numbers are assigned correctly",
			input:     "Line one\nLine two\nLine three",
			wantCount: 3,
			wantErr:   false,
			validate: func(t *testing.T, entries []LogEntry) {
				for i, entry := range entries {
					expectedLineNum := i + 1
					if entry.LineNumber != expectedLineNum {
						t.Errorf("Entry %d LineNumber = %v, want %v", i, entry.LineNumber, expectedLineNum)
					}
				}
			},
		},
		{
			name:      "timestamp detection - simple date format",
			input:     "2024-01-15 10:30:00 Application started",
			wantCount: 1,
			wantErr:   false,
			validate: func(t *testing.T, entries []LogEntry) {
				if entries[0].Timestamp.IsZero() {
					t.Error("Timestamp should be detected but got zero value")
				}
				expectedYear := 2024
				if entries[0].Timestamp.Year() != expectedYear {
					t.Errorf("Timestamp year = %v, want %v", entries[0].Timestamp.Year(), expectedYear)
				}
			},
		},
		{
			name:      "empty lines should be skipped",
			input:     "First line\n\n\nSecond line",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "windows line endings",
			input:     "First line\r\nSecond line\r\nThird line",
			wantCount: 3,
			wantErr:   false,
		},
	}

	parser := NewGenericParser()
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

func TestGenericParser_Parse_EdgeCases(t *testing.T) {
	parser := NewGenericParser()

	t.Run("very long line", func(t *testing.T) {
		longLine := strings.Repeat("a", 10000)
		entries, err := parser.Parse([]byte(longLine))
		if err != nil {
			t.Errorf("Parse() should handle long lines, got error: %v", err)
		}
		if len(entries) != 1 {
			t.Errorf("Parse() should return 1 entry for long line, got %v", len(entries))
		}
	})

	t.Run("mixed content with special characters", func(t *testing.T) {
		input := "Line with émojis 🎉\nLine with tabs\t\there\nLine with quotes \"test\""
		entries, err := parser.Parse([]byte(input))
		if err != nil {
			t.Errorf("Parse() should handle special chars, got error: %v", err)
		}
		if len(entries) != 3 {
			t.Errorf("Parse() returned %v entries, want 3", len(entries))
		}
	})
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level LogLevel
		want  string
	}{
		{LevelTrace, "TRACE"},
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarning, "WARNING"},
		{LevelError, "ERROR"},
		{LevelFatal, "FATAL"},
		{LevelUnknown, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.want)
			}
		})
	}
}