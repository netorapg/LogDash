package discovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function para criar estrutura de arquivos de teste
func setupTestDir(t *testing.T) string {
	t.Helper()

	// Criar diretório temporário
	tmpDir, err := os.MkdirTemp("", "logdash-discovery-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Estrutura de diretórios e arquivos para teste
	structure := map[string]string{
		"app.log":                    "Sample log content",
		"error.log":                  "Error log content",
		"access.txt":                 "Access log as txt",
		"output.out":                 "Output file",
		"regular.txt":                "Not a log file",
		"data.json":                  "{}",
		"logs/app.log":               "Nested log",
		"logs/archive/old.log":       "Old archived log",
		"logs/archive/very/deep.log": "Very deep log",
		"node_modules/package.log":   "Should be ignored",
		".git/config":                "Git config",
		"vendor/lib.log":             "Vendor log - ignored",
		"src/debug.log":              "Source debug log",
		"src/app/main.log":           "Main app log",
	}

	for path, content := range structure {
		fullPath := filepath.Join(tmpDir, path)

		// Criar diretórios necessários
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}

		// Criar arquivo
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	return tmpDir
}

func TestFileDiscoverer_Discover_Basic(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	discoverer := NewFileDiscoverer()
	opts := DefaultDiscoverOptions()

	files, err := discoverer.Discover(tmpDir, opts)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	if len(files) == 0 {
		t.Error("Discover() should find at least some log files")
	}

	// Verificar que arquivos têm campos preenchidos
	for _, file := range files {
		if file.Path == "" {
			t.Error("File Path should not be empty")
		}
		if file.Size <= 0 {
			t.Errorf("File %s Size should be > 0, got %d", file.Path, file.Size)
		}
		if file.LastModified.IsZero() {
			t.Errorf("File %s LastModified should be set", file.Path)
		}
	}
}

func TestFileDiscoverer_Discover_FilePatterns(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name         string
		patterns     []string
		wantContains []string
		wantExcludes []string
	}{
		{
			name:         "only .log files",
			patterns:     []string{"*.log"},
			wantContains: []string{"app.log", "error.log"},
			wantExcludes: []string{"access.txt", "output.out"},
		},
		{
			name:         "only .txt files",
			patterns:     []string{"*.txt"},
			wantContains: []string{"access.txt", "regular.txt"},
			wantExcludes: []string{"app.log", "error.log"},
		},
		{
			name:         "multiple patterns",
			patterns:     []string{"*.log", "*.out"},
			wantContains: []string{"app.log", "output.out"},
			wantExcludes: []string{"data.json"},
		},
	}

	discoverer := NewFileDiscoverer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultDiscoverOptions()
			opts.FilePatterns = tt.patterns

			files, err := discoverer.Discover(tmpDir, opts)
			if err != nil {
				t.Fatalf("Discover() error = %v", err)
			}

			// Verificar que arquivos esperados estão presentes
			for _, want := range tt.wantContains {
				found := false
				for _, file := range files {
					if strings.Contains(file.Path, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find file containing %q", want)
				}
			}

			// Verificar que arquivos excluídos não estão presentes
			for _, exclude := range tt.wantExcludes {
				for _, file := range files {
					if strings.Contains(file.Path, exclude) {
						t.Errorf("Should not find file containing %q", exclude)
					}
				}
			}
		})
	}
}

func TestFileDiscoverer_Discover_ExcludePaths(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	discoverer := NewFileDiscoverer()
	opts := DefaultDiscoverOptions()

	files, err := discoverer.Discover(tmpDir, opts)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	// Verificar que paths excluídos não aparecem
	excludedPaths := []string{"node_modules", ".git", "vendor"}

	for _, file := range files {
		for _, excluded := range excludedPaths {
			if strings.Contains(file.Path, excluded) {
				t.Errorf("File %s should be excluded (contains %s)", file.Path, excluded)
			}
		}
	}
}

func TestFileDiscoverer_Discover_MaxDepth(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name         string
		maxDepth     int
		wantContains []string
		wantExcludes []string
	}{
		{
			name:         "depth 0 - only root",
			maxDepth:     0,
			wantContains: []string{"app.log", "error.log"},
			wantExcludes: []string{"logs/app.log", "logs/archive/old.log"},
		},
		{
			name:         "depth 1 - one level down",
			maxDepth:     1,
			wantContains: []string{"app.log", "logs/app.log"},
			wantExcludes: []string{"logs/archive/old.log"},
		},
		{
			name:         "depth 2 - two levels down",
			maxDepth:     2,
			wantContains: []string{"app.log", "logs/app.log", "logs/archive/old.log"},
			wantExcludes: []string{"logs/archive/very/deep.log"},
		},
	}

	discoverer := NewFileDiscoverer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultDiscoverOptions()
			opts.MaxDepth = tt.maxDepth

			files, err := discoverer.Discover(tmpDir, opts)
			if err != nil {
				t.Fatalf("Discover() error = %v", err)
			}

			for _, want := range tt.wantContains {
				found := false
				for _, file := range files {
					if strings.HasSuffix(file.Path, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find %q at depth %d", want, tt.maxDepth)
				}
			}

			for _, exclude := range tt.wantExcludes {
				for _, file := range files {
					if strings.HasSuffix(file.Path, exclude) {
						t.Errorf("Should not find %q at depth %d", exclude, tt.maxDepth)
					}
				}
			}
		})
	}
}

func TestFileDiscoverer_Discover_SizeFilters(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	// Criar arquivo grande para teste
	largeFile := filepath.Join(tmpDir, "large.log")
	largeContent := strings.Repeat("x", 10000)
	if err := os.WriteFile(largeFile, []byte(largeContent), 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	discoverer := NewFileDiscoverer()

	t.Run("min size filter", func(t *testing.T) {
		opts := DefaultDiscoverOptions()
		opts.MinSize = 5000 // apenas arquivos maiores que 5KB

		files, err := discoverer.Discover(tmpDir, opts)
		if err != nil {
			t.Fatalf("Discover() error = %v", err)
		}

		// Verificar que todos os arquivos são >= MinSize
		for _, file := range files {
			if file.Size < opts.MinSize {
				t.Errorf("File %s size %d is less than MinSize %d", file.Path, file.Size, opts.MinSize)
			}
		}

		// Deve encontrar o arquivo grande
		found := false
		for _, file := range files {
			if strings.Contains(file.Path, "large.log") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Should find large.log with MinSize filter")
		}
	})

	t.Run("max size filter", func(t *testing.T) {
		opts := DefaultDiscoverOptions()
		opts.MaxSize = 100 // apenas arquivos menores que 100 bytes

		files, err := discoverer.Discover(tmpDir, opts)
		if err != nil {
			t.Fatalf("Discover() error = %v", err)
		}

		// Verificar que todos os arquivos são <= MaxSize
		for _, file := range files {
			if file.Size > opts.MaxSize {
				t.Errorf("File %s size %d exceeds MaxSize %d", file.Path, file.Size, opts.MaxSize)
			}
		}

		// Não deve encontrar o arquivo grande
		for _, file := range files {
			if strings.Contains(file.Path, "large.log") {
				t.Error("Should not find large.log with MaxSize filter")
			}
		}
	})
}

func TestFileDiscoverer_Discover_EmptyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "logdash-empty-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	discoverer := NewFileDiscoverer()
	opts := DefaultDiscoverOptions()

	files, err := discoverer.Discover(tmpDir, opts)
	if err != nil {
		t.Fatalf("Discover() should not error on empty dir, got: %v", err)
	}

	if files == nil {
		t.Error("Discover() should return empty slice, not nil")
	}

	if len(files) != 0 {
		t.Errorf("Discover() should return empty slice for empty dir, got %d files", len(files))
	}
}

func TestFileDiscoverer_Discover_NonExistentPath(t *testing.T) {
	discoverer := NewFileDiscoverer()
	opts := DefaultDiscoverOptions()

	_, err := discoverer.Discover("/path/that/does/not/exist", opts)
	if err == nil {
		t.Error("Discover() should return error for non-existent path")
	}
}

func TestFileDiscoverer_Discover_PathIsFile(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	discoverer := NewFileDiscoverer()
	opts := DefaultDiscoverOptions()

	// Apontar para um arquivo específico, não um diretório
	singleFile := filepath.Join(tmpDir, "app.log")

	files, err := discoverer.Discover(singleFile, opts)
	if err != nil {
		t.Fatalf("Discover() should handle file path, got error: %v", err)
	}

	// Deve retornar apenas esse arquivo
	if len(files) != 1 {
		t.Errorf("Discover() on single file should return 1 file, got %d", len(files))
	}

	if len(files) > 0 && files[0].Path != singleFile {
		t.Errorf("File path = %v, want %v", files[0].Path, singleFile)
	}
}

func TestDefaultDiscoverOptions(t *testing.T) {
	opts := DefaultDiscoverOptions()

	if opts.MaxDepth <= 0 {
		t.Error("Default MaxDepth should be > 0")
	}

	if len(opts.FilePatterns) == 0 {
		t.Error("Default FilePatterns should not be empty")
	}

	if len(opts.ExcludePaths) == 0 {
		t.Error("Default ExcludePaths should not be empty")
	}

	if opts.FollowSymlinks {
		t.Error("Default FollowSymlinks should be false for safety")
	}

	if opts.SampleSize <= 0 {
		t.Error("Default SampleSize should be > 0")
	}
}
