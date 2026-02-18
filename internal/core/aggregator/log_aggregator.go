package aggregator

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/netorapg/LogDash/internal/core/parser"
)

// logAggregator implementa agregação e análise de logs
type logAggregator struct{}

// NewLogAggregator cria uma nova instância do agregador
func NewLogAggregator() Aggregator {
	return &logAggregator{}
}

// Aggregate implementa a interface Aggregator
func (a *logAggregator) Aggregate(entries []parser.LogEntry, opts AggregateOptions) (*AggregateResult, error) {
	result := &AggregateResult{
		ByLevel:  make(map[parser.LogLevel]int),
		BySource: make(map[string]int),
	}

	// Filtrar por TimeRange se especificado
	filteredEntries := a.filterByTimeRange(entries, opts.TimeRange)
	result.TotalEntries = len(filteredEntries)

	// Se vazio, retornar resultado vazio
	if len(filteredEntries) == 0 {
		return result, nil
	}

	// Calcular TimeRange efetivo
	result.EffectiveTimeRange = a.calculateEffectiveTimeRange(filteredEntries)

	// Agregar por nível
	for _, entry := range filteredEntries {
		result.ByLevel[entry.Level]++
	}

	// Agregar por source
	for _, entry := range filteredEntries {
		if entry.Source != "" {
			result.BySource[entry.Source]++
		}
	}

	// Gerar timeline
	result.Timeline = a.generateTimeline(filteredEntries, opts.TimelineBucketSize)

	// Calcular top messages
	result.TopMessages = a.calculateTopMessages(filteredEntries, opts.TopMessagesLimit)

	// Detectar anomalias
	if opts.DetectAnomalies {
		result.Anomalies = a.detectAnomalies(filteredEntries, result)
	}

	// Gerar summary
	result.Summary = a.generateSummary(result)

	return result, nil
}

// filterByTimeRange filtra entries por intervalo de tempo
func (a *logAggregator) filterByTimeRange(entries []parser.LogEntry, timeRange TimeRange) []parser.LogEntry {
	if timeRange.IsZero() {
		return entries
	}

	filtered := make([]parser.LogEntry, 0, len(entries))
	for _, entry := range entries {
		if timeRange.Contains(entry.Timestamp) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// calculateEffectiveTimeRange calcula o range real dos dados
func (a *logAggregator) calculateEffectiveTimeRange(entries []parser.LogEntry) TimeRange {
	if len(entries) == 0 {
		return TimeRange{}
	}

	start := entries[0].Timestamp
	end := entries[0].Timestamp

	for _, entry := range entries {
		if !entry.Timestamp.IsZero() {
			if entry.Timestamp.Before(start) {
				start = entry.Timestamp
			}
			if entry.Timestamp.After(end) {
				end = entry.Timestamp
			}
		}
	}

	return TimeRange{Start: start, End: end}
}

// generateTimeline cria timeline com buckets temporais
func (a *logAggregator) generateTimeline(entries []parser.LogEntry, bucketSize time.Duration) []TimePoint {
	if len(entries) == 0 {
		return []TimePoint{}
	}

	// Agrupar por bucket
	buckets := make(map[time.Time]*TimePoint)

	for _, entry := range entries {
		if entry.Timestamp.IsZero() {
			continue
		}

		// Calcular bucket timestamp (arredonda para baixo)
		bucketTime := entry.Timestamp.Truncate(bucketSize)

		if buckets[bucketTime] == nil {
			buckets[bucketTime] = &TimePoint{
				Timestamp: bucketTime,
				ByLevel:   make(map[parser.LogLevel]int),
			}
		}

		buckets[bucketTime].Count++
		buckets[bucketTime].ByLevel[entry.Level]++
	}

	// Converter map para slice e ordenar
	timeline := make([]TimePoint, 0, len(buckets))
	for _, point := range buckets {
		timeline = append(timeline, *point)
	}

	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].Timestamp.Before(timeline[j].Timestamp)
	})

	return timeline
}

// calculateTopMessages encontra mensagens mais frequentes
func (a *logAggregator) calculateTopMessages(entries []parser.LogEntry, limit int) []MessageFrequency {
	if len(entries) == 0 {
		return []MessageFrequency{}
	}

	// Contar frequências
	frequencies := make(map[string]*MessageFrequency)

	for _, entry := range entries {
		msg := entry.Message
		if msg == "" {
			continue
		}

		if frequencies[msg] == nil {
			frequencies[msg] = &MessageFrequency{
				Message:   msg,
				Level:     entry.Level,
				FirstSeen: entry.Timestamp,
				LastSeen:  entry.Timestamp,
				Sources:   []string{},
			}
		}

		freq := frequencies[msg]
		freq.Count++

		// Atualizar FirstSeen/LastSeen
		if !entry.Timestamp.IsZero() {
			if freq.FirstSeen.IsZero() || entry.Timestamp.Before(freq.FirstSeen) {
				freq.FirstSeen = entry.Timestamp
			}
			if freq.LastSeen.IsZero() || entry.Timestamp.After(freq.LastSeen) {
				freq.LastSeen = entry.Timestamp
			}
		}

		// Adicionar source se não existir
		if entry.Source != "" && !contains(freq.Sources, entry.Source) {
			freq.Sources = append(freq.Sources, entry.Source)
		}
	}

	// Converter para slice e ordenar por frequência
	topMessages := make([]MessageFrequency, 0, len(frequencies))
	for _, freq := range frequencies {
		topMessages = append(topMessages, *freq)
	}

	sort.Slice(topMessages, func(i, j int) bool {
		return topMessages[i].Count > topMessages[j].Count
	})

	// Limitar ao tamanho solicitado
	if limit > 0 && len(topMessages) > limit {
		topMessages = topMessages[:limit]
	}

	return topMessages
}

// detectAnomalies detecta padrões anormais nos logs
func (a *logAggregator) detectAnomalies(entries []parser.LogEntry, result *AggregateResult) []Anomaly {
	anomalies := []Anomaly{}

	// Detectar error burst (muitos erros em pouco tempo)
	errorBurst := a.detectErrorBurst(entries)
	if errorBurst != nil {
		anomalies = append(anomalies, *errorBurst)
	}

	// Detectar mensagem repetida excessivamente
	repeatedErrors := a.detectRepeatedErrors(result.TopMessages)
	anomalies = append(anomalies, repeatedErrors...)

	// Detectar spike na timeline
	spike := a.detectSpike(result.Timeline)
	if spike != nil {
		anomalies = append(anomalies, *spike)
	}

	return anomalies
}

// detectErrorBurst detecta explosão de erros em curto período
func (a *logAggregator) detectErrorBurst(entries []parser.LogEntry) *Anomaly {
	const burstWindow = 10 * time.Minute
	const burstThreshold = 5 // 5+ erros em 10 minutos = burst

	errors := []parser.LogEntry{}
	for _, entry := range entries {
		if entry.Level == parser.LevelError || entry.Level == parser.LevelFatal {
			errors = append(errors, entry)
		}
	}

	if len(errors) < burstThreshold {
		return nil
	}

	// Procurar janela de tempo com muitos erros
	for i := 0; i < len(errors)-burstThreshold+1; i++ {
		windowStart := errors[i].Timestamp
		windowEnd := windowStart.Add(burstWindow)

		count := 0
		var windowErrors []parser.LogEntry

		for j := i; j < len(errors); j++ {
			if errors[j].Timestamp.After(windowEnd) {
				break
			}
			count++
			windowErrors = append(windowErrors, errors[j])
		}

		if count >= burstThreshold {
			severity := float64(count) / 10.0
			if severity > 1.0 {
				severity = 1.0
			}

			return &Anomaly{
				Type:           AnomalyErrorBurst,
				Description:    fmt.Sprintf("%d errors in %v", count, burstWindow),
				Severity:       severity,
				Timestamp:      windowStart,
				RelatedEntries: windowErrors,
			}
		}
	}

	return nil
}

// detectRepeatedErrors detecta mensagens de erro repetidas
func (a *logAggregator) detectRepeatedErrors(topMessages []MessageFrequency) []Anomaly {
	anomalies := []Anomaly{}

	for _, msg := range topMessages {
		// Se é erro e aparece muitas vezes, é suspeito
		if (msg.Level == parser.LevelError || msg.Level == parser.LevelFatal) && msg.Count >= 3 {
			severity := float64(msg.Count) / 10.0
			if severity > 1.0 {
				severity = 1.0
			}

			anomalies = append(anomalies, Anomaly{
				Type:        AnomalyRepeatedError,
				Description: fmt.Sprintf("Error message repeated %d times: %s", msg.Count, truncateString(msg.Message, 50)),
				Severity:    severity,
				Timestamp:   msg.FirstSeen,
			})
		}
	}

	return anomalies
}

// detectSpike detecta picos súbitos na timeline
func (a *logAggregator) detectSpike(timeline []TimePoint) *Anomaly {
	if len(timeline) < 3 {
		return nil
	}

	// Calcular média
	var total int
	for _, point := range timeline {
		total += point.Count
	}
	avg := float64(total) / float64(len(timeline))

	// Procurar ponto que seja 3x a média
	for _, point := range timeline {
		if float64(point.Count) > avg*3 {
			severity := (float64(point.Count) / avg) / 10.0
			if severity > 1.0 {
				severity = 1.0
			}

			return &Anomaly{
				Type:        AnomalySpike,
				Description: fmt.Sprintf("Spike of %d logs (avg: %.0f)", point.Count, avg),
				Severity:    severity,
				Timestamp:   point.Timestamp,
			}
		}
	}

	return nil
}

// generateSummary cria resumo textual da análise
func (a *logAggregator) generateSummary(result *AggregateResult) string {
	if result.TotalEntries == 0 {
		return "No log entries found."
	}

	parts := []string{
		fmt.Sprintf("Analyzed %d log entries", result.TotalEntries),
	}

	// Adicionar info sobre níveis
	errors := result.ByLevel[parser.LevelError] + result.ByLevel[parser.LevelFatal]
	warnings := result.ByLevel[parser.LevelWarning]

	if errors > 0 {
		parts = append(parts, fmt.Sprintf("%d errors", errors))
	}
	if warnings > 0 {
		parts = append(parts, fmt.Sprintf("%d warnings", warnings))
	}

	// Adicionar info sobre anomalias
	if len(result.Anomalies) > 0 {
		parts = append(parts, fmt.Sprintf("%d anomalies detected", len(result.Anomalies)))
	}

	// Adicionar período
	if !result.EffectiveTimeRange.IsZero() {
		duration := result.EffectiveTimeRange.End.Sub(result.EffectiveTimeRange.Start)
		parts = append(parts, fmt.Sprintf("spanning %v", duration.Round(time.Second)))
	}

	return strings.Join(parts, ", ") + "."
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
