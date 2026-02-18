package discovery

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// fileDiscoverer implementa a descoberta de arquivos de log no sistema de arquivos
type fileDiscoverer struct{}

// NewFileDiscoverer cria uma nova instância do descobridor de arquivos
func NewFileDiscoverer() Discoverer {
	return &fileDiscoverer{}
}

// Discover implementa a interface Discoverer
func (d *fileDiscoverer) Discover(rootPath string, opts DiscoverOptions) ([]LogFile, error) {
	// Verificar se o path existe
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat root path: %w", err)
	}

	// Inicializar slice vazio (não nil)
	files := []LogFile{}

	// Se for um arquivo único, retornar apenas ele (se passar nos filtros)
	if !info.IsDir() {
		if d.shouldIncludeFile(rootPath, info, opts, 0) {
			logFile := d.createLogFile(rootPath, info, opts)
			return []LogFile{logFile}, nil
		}
		return files, nil
	}

	// Percorrer diretório recursivamente
	err = d.walkDirectory(rootPath, rootPath, 0, opts, &files)
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}

// walkDirectory percorre recursivamente um diretório
func (d *fileDiscoverer) walkDirectory(rootPath, currentPath string, currentDepth int, opts DiscoverOptions, files *[]LogFile) error {
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		// Não falhar por permissão negada, apenas pular
		return nil
	}

	for _, entry := range entries {
		fullPath := filepath.Join(currentPath, entry.Name())

		// Verificar se deve excluir este path
		if d.shouldExcludePath(fullPath, entry.Name(), opts) {
			continue
		}

		// Se for diretório, verificar profundidade antes de recursão
		if entry.IsDir() {
			// MaxDepth 0 = apenas root, não entra em subdiretórios
			// MaxDepth 1 = root + 1 nível de subdiretórios
			// currentDepth 0 = root, 1 = primeiro nível, etc
			if opts.MaxDepth == 0 && currentDepth == 0 {
				// MaxDepth 0: não entrar em nenhum subdiretório do root
				continue
			}
			if opts.MaxDepth > 0 && currentDepth >= opts.MaxDepth {
				// Atingiu limite de profundidade
				continue
			}

			err := d.walkDirectory(rootPath, fullPath, currentDepth+1, opts, files)
			if err != nil {
				// Continuar mesmo com erros em subdiretórios
				continue
			}
			continue
		}

		// Se for symlink e não devemos seguir, pular
		if entry.Type()&fs.ModeSymlink != 0 && !opts.FollowSymlinks {
			continue
		}

		// Obter info do arquivo
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Verificar se deve incluir este arquivo
		if d.shouldIncludeFile(fullPath, info, opts, currentDepth) {
			logFile := d.createLogFile(fullPath, info, opts)
			*files = append(*files, logFile)
		}
	}

	return nil
}

// shouldExcludePath verifica se um path deve ser excluído
func (d *fileDiscoverer) shouldExcludePath(fullPath, name string, opts DiscoverOptions) bool {
	// Verificar contra paths excluídos
	for _, exclude := range opts.ExcludePaths {
		// Pode ser nome exato ou padrão
		if name == exclude || strings.Contains(fullPath, exclude) {
			return true
		}
	}
	return false
}

// shouldIncludeFile verifica se um arquivo deve ser incluído
func (d *fileDiscoverer) shouldIncludeFile(path string, info fs.FileInfo, opts DiscoverOptions, depth int) bool {
	// Verificar tamanho mínimo
	if opts.MinSize > 0 && info.Size() < opts.MinSize {
		return false
	}

	// Verificar tamanho máximo
	if opts.MaxSize > 0 && info.Size() > opts.MaxSize {
		return false
	}

	// Verificar contra padrões de arquivo
	matched := false
	for _, pattern := range opts.FilePatterns {
		match, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && match {
			matched = true
			break
		}
	}

	return matched
}

// createLogFile cria um LogFile a partir de FileInfo
func (d *fileDiscoverer) createLogFile(path string, info fs.FileInfo, opts DiscoverOptions) LogFile {
	logFile := LogFile{
		Path:         path,
		Size:         info.Size(),
		LastModified: info.ModTime(),
	}

	// Tentar detectar formato lendo amostra do arquivo
	if opts.SampleSize > 0 {
		logFile.DetectedFormat = d.detectFormat(path, opts.SampleSize)
	}

	return logFile
}

// detectFormat tenta detectar o formato do log lendo uma amostra
func (d *fileDiscoverer) detectFormat(path string, sampleSize int) string {
	// Ler amostra do arquivo
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	sample := make([]byte, sampleSize)
	n, err := file.Read(sample)
	if err != nil && n == 0 {
		return ""
	}

	sample = sample[:n]
	content := string(sample)

	// Heurísticas simples para detectar formatos conhecidos
	// Formato Python Logging: [LEVEL] timestamp
	if strings.Contains(content, "[INFO]") || strings.Contains(content, "[ERROR]") || strings.Contains(content, "[WARNING]") {
		if strings.Contains(content, ",") { // timestamp com milissegundos
			return "python-logging"
		}
	}

	// JSON logs
	if strings.HasPrefix(strings.TrimSpace(content), "{") {
		return "json"
	}

	// Log4j/Logback: timestamp seguido de level
	if strings.Contains(content, "ERROR") || strings.Contains(content, "WARN") || strings.Contains(content, "INFO") {
		return "structured"
	}

	return "generic"
}
