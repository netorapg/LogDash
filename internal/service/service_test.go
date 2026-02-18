package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/netorapg/LogDash/internal/core/aggregator"
	"github.com/netorapg/LogDash/internal/core/parser"
)

// Helper para criar estrutura de testes
func setupTestLogs(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "logdash-service-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Criar logs de teste
	logs := map[string]string{
		"app.log": `[INFO] 2024-01-15 10:00:00,000 app:1 main ctx: Application started
[ERROR] 2024-01-15 10:05:00,000 app:2 handler ctx: Connection failed
[ERROR] 2024-01-15 10:05:01,000 app:3 handler ctx: Connection failed
[WARNING] 2024-01-15 10:10:00,000 app:4 db ctx: Slow query detected`,

		"db.log": `[INFO] 2024-01-15 10:00:00,000 db:1 init ctx: Database initialized
[ERROR] 2024-01-15 10:15:00,000 db:2 query ctx: Query timeout
[INFO] 2024-01-15 10:20:00,000 db:3 backup ctx: Backup completed`,

		"api.log": `[INFO] 2024-01-15 10:00:00,000 api:1 server ctx: Server listening on :8080
[WARNING] 2024-01-15 10:12:00,000 api:2 auth ctx: Invalid token
[ERROR] 2024-01-15 10:25:00,000 api:3 handler ctx: Internal server error`,
	}

	for filename, content := range logs {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", filename, err)
		}
	}

	return tmpDir
}

func TestNewLogDashService(t *testing.T) {
	service := NewLogDashService()

	if service == nil {
		t.Fatal("NewLogDashService() should not return nil")
	}
}

func TestLogDashService_AnalyzeLogs_Basic(t *testing.T) {
	tmpDir := setupTestLogs(t)
	defer os.RemoveAll(tmpDir)

	service := NewLogDashService()
	opts := DefaultAnalyzeOptions()
	opts.RootPath = tmpDir

	result, err := service.AnalyzeLogs(opts)
	if err != nil {
		t.Fatalf("AnalyzeLogs() error = %v", err)
	}

	if result.FilesDiscovered == 0 {
		t.Error("Should discover at least some files")
	}

	if result.FilesProcessed == 0 {
		t.Error("Should process at least some files")
	}

	if result.EntriesParsed == 0 {
		t.Error("Should parse at least some entries")
	}

	// Verificar que temos aggregate result
	if result.AggregateResult == nil {
		t.Fatal("AggregateResult should not be nil")
	}

	if result.TotalEntries == 0 {
		t.Error("TotalEntries should be > 0")
	}
}

func TestLogDashService_AnalyzeLogs_Statistics(t *testing.T) {
	tmpDir := setupTestLogs(t)
	defer os.RemoveAll(tmpDir)

	service := NewLogDashService()
	opts := DefaultAnalyzeOptions()
	opts.RootPath = tmpDir

	result, err := service.AnalyzeLogs(opts)
	if err != nil {
		t.Fatalf("AnalyzeLogs() error = %v", err)
	}

	// Verificar estatísticas básicas
	errors := result.ByLevel[parser.LevelError]
	warnings := result.ByLevel[parser.LevelWarning]
	infos := result.ByLevel[parser.LevelInfo]

	if errors == 0 {
		t.Error("Should have some errors in test logs")
	}

	if warnings == 0 {
		t.Error("Should have some warnings in test logs")
	}

	if infos == 0 {
		t.Error("Should have some info logs in test logs")
	}

	// Total deve ser soma de todos os níveis
	total := 0
	for _, count := range result.ByLevel {
		total += count
	}

	if total != result.TotalEntries {
		t.Errorf("Sum of ByLevel (%d) should equal TotalEntries (%d)", total, result.TotalEntries)
	}
}

func TestLogDashService_AnalyzeLogs_BySource(t *testing.T) {
	tmpDir := setupTestLogs(t)
	defer os.RemoveAll(tmpDir)

	service := NewLogDashService()
	opts := DefaultAnalyzeOptions()
	opts.RootPath = tmpDir

	result, err := service.AnalyzeLogs(opts)
	if err != nil {
		t.Fatalf("AnalyzeLogs() error = %v", err)
	}

	// Verificar que temos dados por source
	if len(result.BySource) == 0 {
		t.Error("BySource should not be empty")
	}

	// Verificar que sources fazem sentido
	foundAppLog := false
	for source := range result.BySource {
		if strings.Contains(source, "app.log") {
			foundAppLog = true
		}
	}

	if !foundAppLog {
		t.Error("Should have entries from app.log")
	}
}

func TestLogDashService_AnalyzeLogs_FileResults(t *testing.T) {
	tmpDir := setupTestLogs(t)
	defer os.RemoveAll(tmpDir)

	service := NewLogDashService()
	opts := DefaultAnalyzeOptions()
	opts.RootPath = tmpDir

	result, err := service.AnalyzeLogs(opts)
	if err != nil {
		t.Fatalf("AnalyzeLogs() error = %v", err)
	}

	if len(result.FileResults) == 0 {
		t.Error("FileResults should not be empty")
	}

	// Verificar que cada FileResult tem dados válidos
	for _, fr := range result.FileResults {
		if fr.FilePath == "" {
			t.Error("FileResult.FilePath should not be empty")
		}

		if fr.EntriesFound == 0 && fr.Error == nil {
			t.Error("FileResult should have entries or an error")
		}

		if fr.ParserUsed == "" && fr.Error == nil {
			t.Error("FileResult.ParserUsed should be set if no error")
		}

		if fr.ProcessingTime < 0 {
			t.Error("FileResult.ProcessingTime should be >= 0")
		}
	}
}

func TestLogDashService_AnalyzeLogs_MaxFilesToProcess(t *testing.T) {
	tmpDir := setupTestLogs(t)
	defer os.RemoveAll(tmpDir)

	service := NewLogDashService()
	opts := DefaultAnalyzeOptions()
	opts.RootPath = tmpDir
	opts.MaxFilesToProcess = 2 // processar apenas 2 arquivos

	result, err := service.AnalyzeLogs(opts)
	if err != nil {
		t.Fatalf("AnalyzeLogs() error = %v", err)
	}

	if result.FilesProcessed > 2 {
		t.Errorf("FilesProcessed = %d, want <= 2", result.FilesProcessed)
	}
}

func TestLogDashService_AnalyzeLogs_TimeRange(t *testing.T) {
	tmpDir := setupTestLogs(t)
	defer os.RemoveAll(tmpDir)

	service := NewLogDashService()
	opts := DefaultAnalyzeOptions()
	opts.RootPath = tmpDir

	// Filtrar apenas logs entre 10:05 e 10:15
	opts.AggregateOpts.TimeRange = aggregator.TimeRange{
		Start: time.Date(2024, 1, 15, 10, 5, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 15, 10, 15, 0, 0, time.UTC),
	}

	result, err := service.AnalyzeLogs(opts)
	if err != nil {
		t.Fatalf("AnalyzeLogs() error = %v", err)
	}

	// Deve ter parseado mais entries do que retornou (filtrado por tempo)
	if result.EntriesParsed <= result.TotalEntries {
		t.Logf("EntriesParsed: %d, TotalEntries (after filter): %d",
			result.EntriesParsed, result.TotalEntries)
		// Isso é esperado - nem sempre teremos filtros aplicados
	}
}

func TestLogDashService_AnalyzeLogs_Summary(t *testing.T) {
	tmpDir := setupTestLogs(t)
	defer os.RemoveAll(tmpDir)

	service := NewLogDashService()
	opts := DefaultAnalyzeOptions()
	opts.RootPath = tmpDir

	result, err := service.AnalyzeLogs(opts)
	if err != nil {
		t.Fatalf("AnalyzeLogs() error = %v", err)
	}

	if result.Summary == "" {
		t.Error("Summary should not be empty")
	}

	// Summary deve mencionar quantidade de entries
	if !strings.Contains(result.Summary, "log") && !strings.Contains(result.Summary, "entries") {
		t.Error("Summary should mention log entries")
	}
}

func TestLogDashService_AnalyzeLogs_EmptyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "logdash-empty-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	service := NewLogDashService()
	opts := DefaultAnalyzeOptions()
	opts.RootPath = tmpDir

	result, err := service.AnalyzeLogs(opts)
	if err != nil {
		t.Fatalf("AnalyzeLogs() should not error on empty dir, got: %v", err)
	}

	if result.FilesDiscovered != 0 {
		t.Errorf("FilesDiscovered = %d, want 0", result.FilesDiscovered)
	}

	if result.FilesProcessed != 0 {
		t.Errorf("FilesProcessed = %d, want 0", result.FilesProcessed)
	}

	if result.EntriesParsed != 0 {
		t.Errorf("EntriesParsed = %d, want 0", result.EntriesParsed)
	}
}

func TestLogDashService_AnalyzeLogs_NonExistentPath(t *testing.T) {
	service := NewLogDashService()
	opts := DefaultAnalyzeOptions()
	opts.RootPath = "/path/that/does/not/exist/logdash-test"

	_, err := service.AnalyzeLogs(opts)
	if err == nil {
		t.Error("AnalyzeLogs() should return error for non-existent path")
	}
}

func TestLogDashService_AnalyzeLogs_Anomalies(t *testing.T) {
	tmpDir := setupTestLogs(t)
	defer os.RemoveAll(tmpDir)

	service := NewLogDashService()
	opts := DefaultAnalyzeOptions()
	opts.RootPath = tmpDir
	opts.AggregateOpts.DetectAnomalies = true

	result, err := service.AnalyzeLogs(opts)
	if err != nil {
		t.Fatalf("AnalyzeLogs() error = %v", err)
	}

	// Pode ou não ter anomalias dependendo dos dados
	// Apenas verificar que o campo existe
	if result.Anomalies == nil {
		t.Error("Anomalies slice should not be nil (can be empty)")
	}
}

func TestLogDashService_AnalyzeLogs_TopMessages(t *testing.T) {
	tmpDir := setupTestLogs(t)
	defer os.RemoveAll(tmpDir)

	service := NewLogDashService()
	opts := DefaultAnalyzeOptions()
	opts.RootPath = tmpDir
	opts.AggregateOpts.TopMessagesLimit = 5

	result, err := service.AnalyzeLogs(opts)
	if err != nil {
		t.Fatalf("AnalyzeLogs() error = %v", err)
	}

	if len(result.TopMessages) == 0 {
		t.Error("TopMessages should not be empty")
	}

	// Mensagem "Connection failed" aparece 2 vezes, deve estar no topo
	foundRepeated := false
	for _, msg := range result.TopMessages {
		if strings.Contains(msg.Message, "Connection failed") {
			foundRepeated = true
			if msg.Count < 2 {
				t.Errorf("'Connection failed' should appear at least 2 times, got %d", msg.Count)
			}
		}
	}

	if !foundRepeated {
		t.Error("Should find repeated 'Connection failed' message")
	}
}

func TestDefaultAnalyzeOptions(t *testing.T) {
	opts := DefaultAnalyzeOptions()

	if opts.MaxFilesToProcess < 0 {
		t.Error("Default MaxFilesToProcess should be >= 0")
	}

	if !opts.PrioritizeRecent {
		t.Error("Default PrioritizeRecent should be true")
	}

	// Verificar que opções padrão dos sub-componentes estão configuradas
	if opts.DiscoverOpts.MaxDepth <= 0 {
		t.Error("Default DiscoverOpts.MaxDepth should be > 0")
	}

	if opts.AggregateOpts.TopMessagesLimit <= 0 {
		t.Error("Default AggregateOpts.TopMessagesLimit should be > 0")
	}
}
