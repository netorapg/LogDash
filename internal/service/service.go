// Package service fornece a camada de aplicação que orquestra os componentes
// do LogDash (Discovery, Parser, Aggregator).
//
// O service é responsável por:
// - Descobrir arquivos de log
// - Selecionar parsers apropriados
// - Parsear conteúdo dos arquivos
// - Agregar resultados
// - Retornar análise completa
//
// Exemplo de uso:
//
//	service := service.NewLogDashService()
//	opts := service.AnalyzeOptions{
//	    RootPath: "/var/log",
//	    TimeRange: aggregator.TimeRange{
//	        Start: time.Now().Add(-24 * time.Hour),
//	        End:   time.Now(),
//	    },
//	}
//	result, err := service.AnalyzeLogs(opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.Summary)
package service

import (
	"github.com/netorapg/LogDash/internal/core/aggregator"
	"github.com/netorapg/LogDash/internal/core/discovery"
	"github.com/netorapg/LogDash/internal/core/parser"
)

// LogDashService é o serviço principal que orquestra todos os componentes
type LogDashService struct {
	discoverer discovery.Discoverer
	parsers    []parser.Parser
	aggregator aggregator.Aggregator
}

// AnalyzeOptions configura a análise de logs
type AnalyzeOptions struct {
	// RootPath é o caminho raiz para buscar logs
	RootPath string

	// DiscoverOpts opções para descoberta de arquivos
	DiscoverOpts discovery.DiscoverOptions

	// AggregateOpts opções para agregação
	AggregateOpts aggregator.AggregateOptions

	// MaxFilesToProcess limita quantos arquivos processar (0 = sem limite)
	MaxFilesToProcess int

	// PrioritizeRecent prioriza arquivos mais recentes
	PrioritizeRecent bool
}

// AnalysisResult contém o resultado completo da análise
type AnalysisResult struct {
	// FilesDiscovered total de arquivos encontrados
	FilesDiscovered int

	// FilesProcessed total de arquivos efetivamente processados
	FilesProcessed int

	// EntriesParsed total de entradas de log parseadas
	EntriesParsed int

	// AggregateResult resultado da agregação
	*aggregator.AggregateResult

	// FileResults resultados individuais por arquivo
	FileResults []FileResult

	// Errors erros encontrados durante processamento
	Errors []ProcessingError
}

// FileResult resultado do processamento de um arquivo individual
type FileResult struct {
	FilePath       string
	EntriesFound   int
	ParserUsed     string
	ProcessingTime int64 // em millisegundos
	Error          error
}

// ProcessingError representa um erro durante o processamento
type ProcessingError struct {
	FilePath    string
	Stage       string // "discovery", "parsing", "aggregation"
	Error       error
	Recoverable bool
}

// DefaultAnalyzeOptions retorna opções padrão para análise
func DefaultAnalyzeOptions() AnalyzeOptions {
	return AnalyzeOptions{
		DiscoverOpts:      discovery.DefaultDiscoverOptions(),
		AggregateOpts:     aggregator.DefaultAggregateOptions(),
		MaxFilesToProcess: 0, // sem limite
		PrioritizeRecent:  true,
	}
}
