// Package aggregator é responsável por processar entradas de log e gerar insights,
// estatísticas e análises agregadas.
//
// O aggregator recebe LogEntry já parseados e produz informações úteis como:
// distribuição por nível, timeline temporal, mensagens mais frequentes, e anomalias.
//
// Exemplo de uso:
//
//	aggregator := NewLogAggregator()
//	opts := AggregateOptions{
//	    TimeRange: TimeRange{
//	        Start: time.Now().Add(-24 * time.Hour),
//	        End:   time.Now(),
//	    },
//	}
//	result, err := aggregator.Aggregate(entries, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Total: %d, Errors: %d\n",
//	    result.TotalEntries, result.ByLevel[parser.LevelError])
package aggregator

import (
	"time"

	"github.com/netorapg/LogDash/internal/core/parser"
)

// Aggregator define a interface para agregação e análise de logs
type Aggregator interface {
	// Aggregate processa entries e retorna estatísticas e insights
	Aggregate(entries []parser.LogEntry, opts AggregateOptions) (*AggregateResult, error)
}

// AggregateOptions configura o comportamento da agregação
type AggregateOptions struct {
	// TimeRange filtra entries por período (opcional)
	TimeRange TimeRange

	// GroupBy define campos para agrupamento adicional
	GroupBy []GroupByField

	// TopMessagesLimit quantas mensagens mais frequentes retornar (padrão: 10)
	TopMessagesLimit int

	// TimelineBucketSize tamanho dos buckets da timeline (padrão: 1 hora)
	TimelineBucketSize time.Duration

	// DetectAnomalies ativa/desativa detecção de anomalias
	DetectAnomalies bool
}

// TimeRange representa um intervalo de tempo
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// IsZero retorna true se o TimeRange não está definido
func (tr TimeRange) IsZero() bool {
	return tr.Start.IsZero() && tr.End.IsZero()
}

// Contains verifica se um timestamp está dentro do range
func (tr TimeRange) Contains(t time.Time) bool {
	if tr.IsZero() {
		return true // sem filtro = aceita tudo
	}
	return (tr.Start.IsZero() || t.After(tr.Start) || t.Equal(tr.Start)) &&
		(tr.End.IsZero() || t.Before(tr.End) || t.Equal(tr.End))
}

// GroupByField representa um campo para agrupamento
type GroupByField string

const (
	GroupBySource   GroupByField = "source"
	GroupByLevel    GroupByField = "level"
	GroupByHour     GroupByField = "hour"
	GroupByDay      GroupByField = "day"
	GroupByMetadata GroupByField = "metadata"
)

// AggregateResult contém os resultados da agregação
type AggregateResult struct {
	// TotalEntries é o número total de entradas processadas
	TotalEntries int

	// ByLevel conta entradas por nível de severidade
	ByLevel map[parser.LogLevel]int

	// BySource conta entradas por arquivo de origem
	BySource map[string]int

	// Timeline mostra distribuição temporal
	Timeline []TimePoint

	// TopMessages lista mensagens mais frequentes
	TopMessages []MessageFrequency

	// Anomalies lista anomalias detectadas
	Anomalies []Anomaly

	// TimeRange efetivo usado (pode ser diferente do solicitado)
	EffectiveTimeRange TimeRange

	// Summary texto resumido da análise
	Summary string
}

// TimePoint representa um ponto na timeline
type TimePoint struct {
	Timestamp time.Time
	Count     int
	ByLevel   map[parser.LogLevel]int
}

// MessageFrequency representa uma mensagem e sua frequência
type MessageFrequency struct {
	Message string
	Count   int
	Level   parser.LogLevel
	// FirstSeen primeiro timestamp onde apareceu
	FirstSeen time.Time
	// LastSeen último timestamp onde apareceu
	LastSeen time.Time
	// Sources arquivos onde apareceu
	Sources []string
}

// Anomaly representa uma anomalia detectada nos logs
type Anomaly struct {
	Type        AnomalyType
	Description string
	Severity    float64 // 0.0 a 1.0
	Timestamp   time.Time
	// Entries relacionadas à anomalia
	RelatedEntries []parser.LogEntry
}

// AnomalyType tipos de anomalias detectáveis
type AnomalyType string

const (
	// AnomalySpike pico súbito de logs
	AnomalySpike AnomalyType = "spike"
	// AnomalySilence período sem logs quando deveria haver
	AnomalySilence AnomalyType = "silence"
	// AnomalyErrorBurst explosão de erros em curto período
	AnomalyErrorBurst AnomalyType = "error_burst"
	// AnomalyRepeatedError mesmo erro repetido muitas vezes
	AnomalyRepeatedError AnomalyType = "repeated_error"
)

// DefaultAggregateOptions retorna opções padrão sensatas
func DefaultAggregateOptions() AggregateOptions {
	return AggregateOptions{
		TimeRange:          TimeRange{}, // sem filtro
		GroupBy:            []GroupByField{},
		TopMessagesLimit:   10,
		TimelineBucketSize: 1 * time.Hour,
		DetectAnomalies:    true,
	}
}
