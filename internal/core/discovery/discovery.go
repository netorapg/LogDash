// Package discovery é responsável por encontrar arquivos de log no sistema de arquivos.
//
// O discovery vasculha recursivamente diretórios, identifica arquivos que parecem ser logs
// (por extensão ou conteúdo) e retorna metadados sobre eles sem ler o conteúdo completo.
//
// Exemplo de uso:
//
//	discoverer := NewFileDiscoverer()
//	opts := DiscoverOptions{
//	    MaxDepth:     5,
//	    FilePatterns: []string{"*.log", "*.txt"},
//	    ExcludePaths: []string{"node_modules", ".git"},
//	}
//	files, err := discoverer.Discover("/path/to/project", opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, file := range files {
//	    fmt.Printf("Found: %s (%d bytes)\n", file.Path, file.Size)
//	}
package discovery

import "time"

// Discoverer define a interface para descoberta de arquivos de log
type Discoverer interface {
	// Discover procura por arquivos de log a partir de rootPath
	// Retorna slice vazio (não nil) quando não encontrar arquivos
	Discover(rootPath string, opts DiscoverOptions) ([]LogFile, error)
}

// LogFile representa um arquivo de log encontrado
type LogFile struct {
	Path           string    // caminho absoluto do arquivo
	Size           int64     // tamanho em bytes
	LastModified   time.Time // última modificação
	DetectedFormat string    // formato detectado (vazio se desconhecido)
}

// DiscoverOptions configura o comportamento da descoberta
type DiscoverOptions struct {
	// MaxDepth limita a profundidade de recursão (0 = sem limite)
	MaxDepth int

	// FilePatterns são padrões glob para filtrar arquivos (ex: "*.log", "*.txt")
	// Se vazio, usa padrões padrão: ["*.log", "*.txt", "*.out"]
	FilePatterns []string

	// ExcludePaths são diretórios/arquivos a ignorar (ex: "node_modules", ".git")
	// Padrões padrão sempre aplicados: [".git", "node_modules", "vendor"]
	ExcludePaths []string

	// FollowSymlinks indica se deve seguir links simbólicos
	FollowSymlinks bool

	// MinSize é o tamanho mínimo do arquivo em bytes (0 = sem limite)
	MinSize int64

	// MaxSize é o tamanho máximo do arquivo em bytes (0 = sem limite)
	MaxSize int64

	// SampleSize é quantos bytes ler do arquivo para detectar formato (padrão: 1024)
	SampleSize int
}

// DefaultDiscoverOptions retorna opções padrão sensatas
func DefaultDiscoverOptions() DiscoverOptions {
	return DiscoverOptions{
		MaxDepth:       10,
		FilePatterns:   []string{"*.log", "*.txt", "*.out"},
		ExcludePaths:   []string{".git", "node_modules", "vendor", "__pycache__", ".venv"},
		FollowSymlinks: false,
		MinSize:        0,
		MaxSize:        0, // sem limite
		SampleSize:     1024,
	}
}
