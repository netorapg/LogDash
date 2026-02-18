package aggregator

import (
	"strings"
	"testing"
	"time"

	"github.com/netorapg/LogDash/internal/core/parser"
)

// Helper para criar LogEntry de teste
func createEntry(level parser.LogLevel, message string, timestamp time.Time, source string) parser.LogEntry {
	return parser.LogEntry{
		Timestamp:  timestamp,
		Level:      level,
		Message:    message,
		Source:     source,
		LineNumber: 1,
		Raw:        message,
		Metadata:   make(map[string]interface{}),
	}
}

func TestLogAggregator_Aggregate_Basic(t *testing.T) {
	now := time.Now()

	entries := []parser.LogEntry{
		createEntry(parser.LevelInfo, "App started", now, "app.log"),
		createEntry(parser.LevelError, "Connection failed", now.Add(1*time.Minute), "app.log"),
		createEntry(parser.LevelWarning, "Slow query", now.Add(2*time.Minute), "db.log"),
		createEntry(parser.LevelInfo, "Request processed", now.Add(3*time.Minute), "app.log"),
	}

	aggregator := NewLogAggregator()
	opts := DefaultAggregateOptions()

	result, err := aggregator.Aggregate(entries, opts)
	if err != nil {
		t.Fatalf("Aggregate() error = %v", err)
	}

	if result.TotalEntries != 4 {
		t.Errorf("TotalEntries = %d, want 4", result.TotalEntries)
	}

	if result.ByLevel[parser.LevelInfo] != 2 {
		t.Errorf("ByLevel[Info] = %d, want 2", result.ByLevel[parser.LevelInfo])
	}

	if result.ByLevel[parser.LevelError] != 1 {
		t.Errorf("ByLevel[Error] = %d, want 1", result.ByLevel[parser.LevelError])
	}

	if result.ByLevel[parser.LevelWarning] != 1 {
		t.Errorf("ByLevel[Warning] = %d, want 1", result.ByLevel[parser.LevelWarning])
	}
}

func TestLogAggregator_Aggregate_BySource(t *testing.T) {
	now := time.Now()

	entries := []parser.LogEntry{
		createEntry(parser.LevelInfo, "msg1", now, "app.log"),
		createEntry(parser.LevelInfo, "msg2", now, "app.log"),
		createEntry(parser.LevelInfo, "msg3", now, "db.log"),
		createEntry(parser.LevelInfo, "msg4", now, "api.log"),
		createEntry(parser.LevelInfo, "msg5", now, "api.log"),
		createEntry(parser.LevelInfo, "msg6", now, "api.log"),
	}

	aggregator := NewLogAggregator()
	opts := DefaultAggregateOptions()

	result, err := aggregator.Aggregate(entries, opts)
	if err != nil {
		t.Fatalf("Aggregate() error = %v", err)
	}

	if result.BySource["app.log"] != 2 {
		t.Errorf("BySource[app.log] = %d, want 2", result.BySource["app.log"])
	}

	if result.BySource["db.log"] != 1 {
		t.Errorf("BySource[db.log] = %d, want 1", result.BySource["db.log"])
	}

	if result.BySource["api.log"] != 3 {
		t.Errorf("BySource[api.log] = %d, want 3", result.BySource["api.log"])
	}
}

func TestLogAggregator_Aggregate_TimeRange(t *testing.T) {
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	entries := []parser.LogEntry{
		createEntry(parser.LevelInfo, "msg1", base.Add(-2*time.Hour), "app.log"), // antes
		createEntry(parser.LevelInfo, "msg2", base.Add(-1*time.Hour), "app.log"), // dentro
		createEntry(parser.LevelInfo, "msg3", base, "app.log"),                   // dentro
		createEntry(parser.LevelInfo, "msg4", base.Add(1*time.Hour), "app.log"),  // dentro
		createEntry(parser.LevelInfo, "msg5", base.Add(3*time.Hour), "app.log"),  // depois
	}

	aggregator := NewLogAggregator()
	opts := DefaultAggregateOptions()
	opts.TimeRange = TimeRange{
		Start: base.Add(-1 * time.Hour),
		End:   base.Add(2 * time.Hour),
	}

	result, err := aggregator.Aggregate(entries, opts)
	if err != nil {
		t.Fatalf("Aggregate() error = %v", err)
	}

	// Deve contar apenas os 3 entries dentro do range
	if result.TotalEntries != 3 {
		t.Errorf("TotalEntries = %d, want 3 (filtered by time range)", result.TotalEntries)
	}
}

func TestLogAggregator_Aggregate_Timeline(t *testing.T) {
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	entries := []parser.LogEntry{
		createEntry(parser.LevelInfo, "msg1", base, "app.log"),
		createEntry(parser.LevelError, "msg2", base.Add(30*time.Minute), "app.log"),
		createEntry(parser.LevelInfo, "msg3", base.Add(1*time.Hour), "app.log"),
		createEntry(parser.LevelInfo, "msg4", base.Add(1*time.Hour+30*time.Minute), "app.log"),
		createEntry(parser.LevelWarning, "msg5", base.Add(2*time.Hour), "app.log"),
	}

	aggregator := NewLogAggregator()
	opts := DefaultAggregateOptions()
	opts.TimelineBucketSize = 1 * time.Hour

	result, err := aggregator.Aggregate(entries, opts)
	if err != nil {
		t.Fatalf("Aggregate() error = %v", err)
	}

	if len(result.Timeline) == 0 {
		t.Error("Timeline should not be empty")
	}

	// Verificar que timeline está ordenada
	for i := 1; i < len(result.Timeline); i++ {
		if result.Timeline[i].Timestamp.Before(result.Timeline[i-1].Timestamp) {
			t.Error("Timeline should be sorted by timestamp")
		}
	}
}

func TestLogAggregator_Aggregate_TopMessages(t *testing.T) {
	now := time.Now()

	entries := []parser.LogEntry{
		createEntry(parser.LevelError, "Connection timeout", now, "app.log"),
		createEntry(parser.LevelError, "Connection timeout", now.Add(1*time.Minute), "app.log"),
		createEntry(parser.LevelError, "Connection timeout", now.Add(2*time.Minute), "app.log"),
		createEntry(parser.LevelWarning, "Slow query", now.Add(3*time.Minute), "db.log"),
		createEntry(parser.LevelWarning, "Slow query", now.Add(4*time.Minute), "db.log"),
		createEntry(parser.LevelInfo, "User login", now.Add(5*time.Minute), "app.log"),
	}

	aggregator := NewLogAggregator()
	opts := DefaultAggregateOptions()
	opts.TopMessagesLimit = 5

	result, err := aggregator.Aggregate(entries, opts)
	if err != nil {
		t.Fatalf("Aggregate() error = %v", err)
	}

	if len(result.TopMessages) == 0 {
		t.Fatal("TopMessages should not be empty")
	}

	// Primeira mensagem deve ser a mais frequente
	if result.TopMessages[0].Message != "Connection timeout" {
		t.Errorf("TopMessages[0] = %v, want 'Connection timeout'", result.TopMessages[0].Message)
	}

	if result.TopMessages[0].Count != 3 {
		t.Errorf("TopMessages[0].Count = %d, want 3", result.TopMessages[0].Count)
	}

	// Verificar que está ordenado por frequência (decrescente)
	for i := 1; i < len(result.TopMessages); i++ {
		if result.TopMessages[i].Count > result.TopMessages[i-1].Count {
			t.Error("TopMessages should be sorted by count (descending)")
		}
	}

	// Verificar FirstSeen e LastSeen
	top := result.TopMessages[0]
	if top.FirstSeen.IsZero() {
		t.Error("FirstSeen should be set")
	}
	if top.LastSeen.IsZero() {
		t.Error("LastSeen should be set")
	}
	if top.LastSeen.Before(top.FirstSeen) {
		t.Error("LastSeen should be after or equal to FirstSeen")
	}
}

func TestLogAggregator_Aggregate_AnomalyDetection(t *testing.T) {
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	// Criar padrão normal + spike de erros
	entries := []parser.LogEntry{
		// Normal: 1-2 logs por hora
		createEntry(parser.LevelInfo, "msg1", base, "app.log"),
		createEntry(parser.LevelInfo, "msg2", base.Add(1*time.Hour), "app.log"),

		// SPIKE: 10 erros em 5 minutos
		createEntry(parser.LevelError, "error1", base.Add(2*time.Hour), "app.log"),
		createEntry(parser.LevelError, "error2", base.Add(2*time.Hour+1*time.Minute), "app.log"),
		createEntry(parser.LevelError, "error3", base.Add(2*time.Hour+2*time.Minute), "app.log"),
		createEntry(parser.LevelError, "error4", base.Add(2*time.Hour+3*time.Minute), "app.log"),
		createEntry(parser.LevelError, "error5", base.Add(2*time.Hour+4*time.Minute), "app.log"),
		createEntry(parser.LevelError, "error6", base.Add(2*time.Hour+5*time.Minute), "app.log"),
		createEntry(parser.LevelError, "error7", base.Add(2*time.Hour+6*time.Minute), "app.log"),
		createEntry(parser.LevelError, "error8", base.Add(2*time.Hour+7*time.Minute), "app.log"),
		createEntry(parser.LevelError, "error9", base.Add(2*time.Hour+8*time.Minute), "app.log"),
		createEntry(parser.LevelError, "error10", base.Add(2*time.Hour+9*time.Minute), "app.log"),

		// Volta ao normal
		createEntry(parser.LevelInfo, "msg3", base.Add(3*time.Hour), "app.log"),
	}

	aggregator := NewLogAggregator()
	opts := DefaultAggregateOptions()
	opts.DetectAnomalies = true

	result, err := aggregator.Aggregate(entries, opts)
	if err != nil {
		t.Fatalf("Aggregate() error = %v", err)
	}

	// Deve detectar pelo menos 1 anomalia (error burst ou spike)
	if len(result.Anomalies) == 0 {
		t.Error("Should detect anomalies (error burst)")
	}

	// Verificar que anomalia tem dados relevantes
	for _, anomaly := range result.Anomalies {
		if anomaly.Type == "" {
			t.Error("Anomaly Type should be set")
		}
		if anomaly.Description == "" {
			t.Error("Anomaly Description should be set")
		}
		if anomaly.Severity <= 0 || anomaly.Severity > 1 {
			t.Errorf("Anomaly Severity should be between 0 and 1, got %f", anomaly.Severity)
		}
	}
}

func TestLogAggregator_Aggregate_EmptyEntries(t *testing.T) {
	aggregator := NewLogAggregator()
	opts := DefaultAggregateOptions()

	result, err := aggregator.Aggregate([]parser.LogEntry{}, opts)
	if err != nil {
		t.Fatalf("Aggregate() should not error on empty entries, got: %v", err)
	}

	if result.TotalEntries != 0 {
		t.Errorf("TotalEntries = %d, want 0", result.TotalEntries)
	}

	if len(result.ByLevel) != 0 {
		t.Error("ByLevel should be empty")
	}

	if len(result.TopMessages) != 0 {
		t.Error("TopMessages should be empty")
	}
}

func TestLogAggregator_Aggregate_Summary(t *testing.T) {
	now := time.Now()

	entries := []parser.LogEntry{
		createEntry(parser.LevelInfo, "msg1", now, "app.log"),
		createEntry(parser.LevelError, "error1", now, "app.log"),
		createEntry(parser.LevelError, "error2", now, "app.log"),
		createEntry(parser.LevelWarning, "warn1", now, "app.log"),
	}

	aggregator := NewLogAggregator()
	opts := DefaultAggregateOptions()

	result, err := aggregator.Aggregate(entries, opts)
	if err != nil {
		t.Fatalf("Aggregate() error = %v", err)
	}

	// Summary deve existir e ter conteúdo útil
	if result.Summary == "" {
		t.Error("Summary should not be empty")
	}

	// Summary deve mencionar totais
	if !strings.Contains(result.Summary, "4") && !strings.Contains(result.Summary, "four") {
		t.Error("Summary should mention total entries")
	}

	// Summary deve mencionar erros se houver
	if !strings.Contains(strings.ToLower(result.Summary), "error") {
		t.Error("Summary should mention errors when present")
	}
}

func TestTimeRange_Contains(t *testing.T) {
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		timeRange TimeRange
		timestamp time.Time
		want      bool
	}{
		{
			name:      "zero range accepts everything",
			timeRange: TimeRange{},
			timestamp: base,
			want:      true,
		},
		{
			name: "timestamp inside range",
			timeRange: TimeRange{
				Start: base.Add(-1 * time.Hour),
				End:   base.Add(1 * time.Hour),
			},
			timestamp: base,
			want:      true,
		},
		{
			name: "timestamp before range",
			timeRange: TimeRange{
				Start: base,
				End:   base.Add(1 * time.Hour),
			},
			timestamp: base.Add(-1 * time.Hour),
			want:      false,
		},
		{
			name: "timestamp after range",
			timeRange: TimeRange{
				Start: base,
				End:   base.Add(1 * time.Hour),
			},
			timestamp: base.Add(2 * time.Hour),
			want:      false,
		},
		{
			name: "timestamp equals start",
			timeRange: TimeRange{
				Start: base,
				End:   base.Add(1 * time.Hour),
			},
			timestamp: base,
			want:      true,
		},
		{
			name: "timestamp equals end",
			timeRange: TimeRange{
				Start: base,
				End:   base.Add(1 * time.Hour),
			},
			timestamp: base.Add(1 * time.Hour),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.timeRange.Contains(tt.timestamp)
			if got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultAggregateOptions(t *testing.T) {
	opts := DefaultAggregateOptions()

	if opts.TopMessagesLimit <= 0 {
		t.Error("Default TopMessagesLimit should be > 0")
	}

	if opts.TimelineBucketSize <= 0 {
		t.Error("Default TimelineBucketSize should be > 0")
	}

	if !opts.DetectAnomalies {
		t.Error("Default DetectAnomalies should be true")
	}
}
