package service

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/netorapg/LogDash/internal/core/aggregator"
	"github.com/netorapg/LogDash/internal/core/discovery"
	"github.com/netorapg/LogDash/internal/core/parser"
)

// NewLogDashService cria uma nova instância do serviço com componentes padrão
func NewLogDashService() *LogDashService {
	return &LogDashService{
		discoverer: discovery.NewFileDiscoverer(),
		parsers: []parser.Parser{
			parser.NewPythonLoggingParser(),
			parser.NewGenericParser(), // fallback
		},
		aggregator: aggregator.NewLogAggregator(),
	}
}

// AnalyzeLogs executa análise completa: discovery → parsing → aggregation
func (s *LogDashService) AnalyzeLogs(opts AnalyzeOptions) (*AnalysisResult, error) {
	result := &AnalysisResult{
		FileResults: []FileResult{},
		Errors:      []ProcessingError{},
	}

	// 1. Discovery - encontrar arquivos de log
	files, err := s.discoverer.Discover(opts.RootPath, opts.DiscoverOpts)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	result.FilesDiscovered = len(files)

	if len(files) == 0 {
		// Não é erro, apenas não há arquivos
		result.AggregateResult = &aggregator.AggregateResult{
			ByLevel:  make(map[parser.LogLevel]int),
			BySource: make(map[string]int),
			Summary:  "No log files found.",
		}
		return result, nil
	}

	// 2. Priorizar arquivos recentes se solicitado
	if opts.PrioritizeRecent {
		s.sortFilesByRecency(files)
	}

	// 3. Limitar número de arquivos se especificado
	if opts.MaxFilesToProcess > 0 && len(files) > opts.MaxFilesToProcess {
		files = files[:opts.MaxFilesToProcess]
	}

	// 4. Parsear cada arquivo
	allEntries := []parser.LogEntry{}

	for _, file := range files {
		fileResult := s.processFile(file)
		result.FileResults = append(result.FileResults, fileResult)

		if fileResult.Error != nil {
			result.Errors = append(result.Errors, ProcessingError{
				FilePath:    file.Path,
				Stage:       "parsing",
				Error:       fileResult.Error,
				Recoverable: true,
			})
			continue
		}

		result.FilesProcessed++
		result.EntriesParsed += fileResult.EntriesFound
	}

	// Coletar todas as entries de arquivos processados com sucesso
	for _, fileResult := range result.FileResults {
		if fileResult.Error == nil && fileResult.EntriesFound > 0 {
			// Reprocessar arquivo para pegar entries (já fizemos isso em processFile)
			// Por simplicidade, vamos confiar que está correto
			// Em produção, poderíamos cachear as entries
		}
	}

	// Na verdade, precisamos coletar as entries. Vamos refatorar processFile
	// para retornar as entries também
	allEntries = s.collectAllEntries(files, result.FileResults)

	// 5. Agregar resultados
	aggregateResult, err := s.aggregator.Aggregate(allEntries, opts.AggregateOpts)
	if err != nil {
		result.Errors = append(result.Errors, ProcessingError{
			Stage:       "aggregation",
			Error:       err,
			Recoverable: false,
		})
		// Continuar com resultado parcial
		aggregateResult = &aggregator.AggregateResult{
			ByLevel:  make(map[parser.LogLevel]int),
			BySource: make(map[string]int),
			Summary:  "Aggregation failed.",
		}
	}

	result.AggregateResult = aggregateResult

	return result, nil
}

// processFile processa um arquivo individual
func (s *LogDashService) processFile(file discovery.LogFile) FileResult {
	startTime := time.Now()

	result := FileResult{
		FilePath: file.Path,
	}

	// Ler conteúdo do arquivo
	content, err := os.ReadFile(file.Path)
	if err != nil {
		result.Error = fmt.Errorf("failed to read file: %w", err)
		return result
	}

	// Selecionar parser apropriado
	selectedParser := s.selectParser(content, file.DetectedFormat)
	result.ParserUsed = selectedParser.Name()

	// Parsear
	entries, err := selectedParser.Parse(content)
	if err != nil {
		result.Error = fmt.Errorf("parsing failed: %w", err)
		return result
	}

	// Adicionar source aos entries
	for i := range entries {
		entries[i].Source = file.Path
	}

	result.EntriesFound = len(entries)
	result.ProcessingTime = time.Since(startTime).Milliseconds()

	return result
}

// selectParser seleciona o parser mais apropriado para o conteúdo
func (s *LogDashService) selectParser(content []byte, detectedFormat string) parser.Parser {
	if len(content) == 0 {
		return s.parsers[len(s.parsers)-1] // fallback (GenericParser)
	}

	// Se detection sugeriu um formato específico, dar preferência
	if detectedFormat != "" && detectedFormat != "generic" {
		for _, p := range s.parsers {
			if p.Name() == detectedFormat {
				return p
			}
		}
	}

	// Calcular score de cada parser
	var bestParser parser.Parser
	var bestScore float64

	for _, p := range s.parsers {
		score := p.SupportsFormat(content)
		if score > bestScore {
			bestScore = score
			bestParser = p
		}
	}

	if bestParser != nil {
		return bestParser
	}

	// Fallback para o último parser (GenericParser)
	return s.parsers[len(s.parsers)-1]
}

// collectAllEntries re-processa arquivos para coletar entries
func (s *LogDashService) collectAllEntries(files []discovery.LogFile, fileResults []FileResult) []parser.LogEntry {
	allEntries := []parser.LogEntry{}

	for i, file := range files {
		if i >= len(fileResults) {
			break
		}

		fileResult := fileResults[i]
		if fileResult.Error != nil {
			continue
		}

		// Ler e parsear novamente
		content, err := os.ReadFile(file.Path)
		if err != nil {
			continue
		}

		selectedParser := s.selectParser(content, file.DetectedFormat)
		entries, err := selectedParser.Parse(content)
		if err != nil {
			continue
		}

		// Adicionar source
		for i := range entries {
			entries[i].Source = file.Path
		}

		allEntries = append(allEntries, entries...)
	}

	return allEntries
}

// sortFilesByRecency ordena arquivos por data de modificação (mais recente primeiro)
func (s *LogDashService) sortFilesByRecency(files []discovery.LogFile) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].LastModified.After(files[j].LastModified)
	})
}
