# Aggregator Package

Pacote responsável por agregar e analisar logs, transformando entradas brutas em insights acionáveis.

## Visão Rápida

```go
aggregator := aggregator.NewLogAggregator()
opts := aggregator.DefaultAggregateOptions()

result, err := aggregator.Aggregate(entries, opts)
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.Summary)
// "Analyzed 1,523 log entries, 47 errors, 12 warnings, 2 anomalies detected, spanning 4h23m."

fmt.Printf("Errors: %d\n", result.ByLevel[parser.LevelError])
```

## Features

- ✅ **Estatísticas**: Total, por nível, por source
- ✅ **Timeline**: Distribuição temporal com buckets configuráveis
- ✅ **Top Messages**: Mensagens mais frequentes com metadados ricos
- ✅ **Detecção de Anomalias**: Error burst, repeated errors, spikes
- ✅ **Summary**: Resumo textual gerado automaticamente
- ✅ **Filtros**: TimeRange para análise de períodos específicos
- ✅ **Performance**: O(n) para agregação básica

## Análises Disponíveis

### 1. Estatísticas Básicas
```go
result.TotalEntries           // total processado
result.ByLevel[LevelError]    // quantos erros
result.BySource["app.log"]    // logs por arquivo
```

### 2. Timeline
```go
opts.TimelineBucketSize = 1 * time.Hour

for _, point := range result.Timeline {
    fmt.Printf("%s: %d logs (%d errors)\n",
        point.Timestamp.Format("15:04"),
        point.Count,
        point.ByLevel[parser.LevelError])
}
```

### 3. Top Messages
```go
opts.TopMessagesLimit = 10

for i, msg := range result.TopMessages {
    fmt.Printf("%d. [%s] %s (x%d)\n", 
        i+1, msg.Level, msg.Message, msg.Count)
}
```

### 4. Anomalias
```go
opts.DetectAnomalies = true

for _, anomaly := range result.Anomalies {
    fmt.Printf("[%s] %.0f%% - %s\n",
        anomaly.Type,
        anomaly.Severity * 100,
        anomaly.Description)
}
```

**Tipos de anomalias:**
- `AnomalyErrorBurst` - 5+ erros em 10 minutos
- `AnomalyRepeatedError` - Mesmo erro 3+ vezes
- `AnomalySpike` - Pico 3x acima da média

## Exemplos

### Análise Simples
```go
agg := NewLogAggregator()
result, _ := agg.Aggregate(entries, DefaultAggregateOptions())

fmt.Println(result.Summary)
```

### Filtrar por Período
```go
opts := DefaultAggregateOptions()
opts.TimeRange = TimeRange{
    Start: time.Now().Add(-24 * time.Hour),
    End:   time.Now(),
}

result, _ := agg.Aggregate(entries, opts)
```

### Health Check
```go
errorRate := float64(result.ByLevel[LevelError]) / float64(result.TotalEntries)

if errorRate > 0.1 {
    fmt.Println("🔴 CRITICAL - Error rate > 10%")
} else if errorRate > 0.05 {
    fmt.Println("🟡 WARNING - Error rate > 5%")
} else {
    fmt.Println("🟢 OK")
}
```

### Comparação Temporal
```go
// Antes do deploy
resultBefore, _ := agg.Aggregate(entriesBefore, opts)

// Depois do deploy
resultAfter, _ := agg.Aggregate(entriesAfter, opts)

if resultAfter.ByLevel[LevelError] > resultBefore.ByLevel[LevelError] * 2 {
    fmt.Println("⚠️  Deploy causou aumento de erros!")
}
```

## Performance

| Entries | Tempo | Memória |
|---------|-------|---------|
| 1,000 | < 10ms | ~1MB |
| 10,000 | < 50ms | ~5MB |
| 100,000 | < 500ms | ~30MB |
| 1,000,000 | < 5s | ~200MB |

**Otimizações:**
- Filtrar entries antes de agregar
- Ajustar bucket size ao período
- Limitar TopMessagesLimit ao necessário

## Opções

```go
type AggregateOptions struct {
    TimeRange          TimeRange      // filtrar por período
    TopMessagesLimit   int            // padrão: 10
    TimelineBucketSize time.Duration  // padrão: 1h
    DetectAnomalies    bool           // padrão: true
}
```

**Padrões sensatos:**
```go
opts := DefaultAggregateOptions()
// TopMessagesLimit = 10
// TimelineBucketSize = 1 hour
// DetectAnomalies = true
```

## Integração

### Com Parser
```go
// 1. Parsear
parser := parser.NewPythonLoggingParser()
entries, _ := parser.Parse(content)

// 2. Agregar
aggregator := aggregator.NewLogAggregator()
result, _ := aggregator.Aggregate(entries, opts)
```

### Com Discovery
```go
// 1. Descobrir
discoverer := discovery.NewFileDiscoverer()
files, _ := discoverer.Discover("/var/log", discovery.DefaultDiscoverOptions())

// 2. Parsear todos
allEntries := []parser.LogEntry{}
for _, file := range files {
    // ... parse each file
    allEntries = append(allEntries, entries...)
}

// 3. Agregar
result, _ := aggregator.Aggregate(allEntries, opts)
```

## Testes

```bash
# Rodar todos os testes
go test -v

# Ver cobertura
go test -cover

# Gerar relatório HTML
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Documentação Completa

Veja [AGGREGATOR_GUIDE.md](../../../docs/AGGREGATOR_GUIDE.md) para:
- Guia detalhado de cada análise
- Casos de uso completos
- Detecção de anomalias explicada
- Troubleshooting
- Boas práticas
- Extensibilidade

## Estrutura de Dados

### AggregateResult
```go
TotalEntries       int                // total
ByLevel            map[LogLevel]int   // por severidade
BySource           map[string]int     // por arquivo
Timeline           []TimePoint        // temporal
TopMessages        []MessageFrequency // frequentes
Anomalies          []Anomaly          // detectadas
EffectiveTimeRange TimeRange          // período real
Summary            string             // resumo
```

### TimePoint
```go
Timestamp time.Time              // bucket
Count     int                    // total
ByLevel   map[LogLevel]int       // distribuição
```

### MessageFrequency
```go
Message   string       // texto
Count     int          // frequência
Level     LogLevel     // severidade
FirstSeen time.Time    // primeira vez
LastSeen  time.Time    // última vez
Sources   []string     // arquivos
```

### Anomaly
```go
Type           AnomalyType    // tipo
Description    string         // descrição
Severity       float64        // 0.0 a 1.0
Timestamp      time.Time      // quando
RelatedEntries []LogEntry     // relacionados
```