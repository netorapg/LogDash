package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/netorapg/LogDash/internal/core/aggregator"
	"github.com/netorapg/LogDash/internal/core/parser"
	"github.com/netorapg/LogDash/internal/service"
)

// Server representa o servidor web
type Server struct {
	service  *service.LogDashService
	rootPath string
	opts     service.AnalyzeOptions
	result   *service.AnalysisResult
	port     int
}

// NewServer cria novo servidor web
func NewServer(rootPath string, port int) *Server {
	return &Server{
		service:  service.NewLogDashService(),
		rootPath: rootPath,
		opts:     service.DefaultAnalyzeOptions(),
		port:     port,
	}
}

// Start inicia o servidor web
func (s *Server) Start() error {
	// Configurar opções
	s.opts.RootPath = s.rootPath
	s.opts.AggregateOpts.TopMessagesLimit = 1000 // Capturar todas as mensagens

	// Fazer análise inicial
	if err := s.analyze(); err != nil {
		return fmt.Errorf("initial analysis failed: %w", err)
	}

	// Configurar rotas
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/api/refresh", s.handleRefresh)
	http.HandleFunc("/api/stats", s.handleStats)
	http.HandleFunc("/api/errors", s.handleErrors)
	http.HandleFunc("/api/warnings", s.handleWarnings)
	http.HandleFunc("/api/timeline", s.handleTimeline)
	http.HandleFunc("/api/anomalies", s.handleAnomalies)

	// Iniciar servidor
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("🚀 LogDash Web UI starting at http://localhost%s", addr)
	log.Printf("📁 Analyzing logs in: %s", s.rootPath)
	
	return http.ListenAndServe(addr, nil)
}

// analyze executa análise de logs
func (s *Server) analyze() error {
	start := time.Now()
	result, err := s.service.AnalyzeLogs(s.opts)
	if err != nil {
		return err
	}
	
	s.result = result
	log.Printf("✅ Analysis complete in %v", time.Since(start))
	return nil
}

// handleIndex serve a página principal
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("internal/web/templates/index.html"))
	
	data := map[string]interface{}{
		"RootPath": s.rootPath,
		"Result":   s.result,
	}
	
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleRefresh reanalisa os logs
func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if err := s.analyze(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Retornar stats atualizadas
	s.handleStats(w, r)
}

// handleStats retorna estatísticas gerais
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_entries":    s.result.EntriesParsed,
		"files_discovered": s.result.FilesDiscovered,
		"files_processed":  s.result.FilesProcessed,
		"by_level":         s.result.ByLevel,
		"summary":          s.result.Summary,
		"anomalies_count":  len(s.result.Anomalies),
	})
}

// handleErrors retorna lista de erros
func (s *Server) handleErrors(w http.ResponseWriter, r *http.Request) {
	errors := []aggregator.MessageFrequency{}
	
	for _, msg := range s.result.TopMessages {
		if msg.Level == parser.LevelError || msg.Level == parser.LevelFatal {
			errors = append(errors, msg)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(errors)
}

// handleWarnings retorna lista de warnings
func (s *Server) handleWarnings(w http.ResponseWriter, r *http.Request) {
	warnings := []aggregator.MessageFrequency{}
	
	for _, msg := range s.result.TopMessages {
		if msg.Level == parser.LevelWarning {
			warnings = append(warnings, msg)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(warnings)
}

// handleTimeline retorna dados da timeline
func (s *Server) handleTimeline(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.result.Timeline)
}

// handleAnomalies retorna anomalias detectadas
func (s *Server) handleAnomalies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.result.Anomalies)
}