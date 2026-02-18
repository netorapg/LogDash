# Service Package

Camada de aplicação que orquestra Discovery + Parser + Aggregator em um fluxo completo end-to-end.

## Visão Rápida

```go
service := service.NewLogDashService()
opts := service.DefaultAnalyzeOptions()
opts.RootPath = "/var/log"

result, err := service.AnalyzeLogs(opts)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Analyzed %d files, found %d entries\n", 
    result.FilesProcessed, result.EntriesParsed)
fmt.Println(result.Summary)
```

## O que o Service faz

O `LogDashService` executa um pipeline completo:

```
1. Discovery  → Encontra arquivos de log
2. Selection  → Escolhe parser apropriado
3. Parsing    → Extrai LogEntry de cada arquivo
4. Aggregation → Gera estatísticas e insights
5. Return     → AnalysisResult completo
```

## Features

### 🔍 Discovery Automático
- Encontra logs recursivamente no `RootPath`
- Aplica filtros configuráveis (extensões, profundidade, etc)

### 🎯 Parser Selection Inteligente
- Usa detecção de formato do Discovery
- Calcula score de confiança de cada parser
- Seleciona automaticamente o melhor parser

### ⚡ Processamento Otimizado
- Prioriza arquivos mais recentes (`PrioritizeRecent`)
- Limita quantidade de arquivos (`MaxFilesToProcess`)
- Continua mesmo se alguns arquivos falharem

### 📊 Análise Completa
- Estatísticas agregadas (total, por nível, por fonte)
- Timeline temporal
- Top mensagens
- Detecção de anomalias
- Resumo automático

### 📋 Rastreabilidade
- `FileResults`: status de cada arquivo processado
- Lista de erros com contexto
- Tempo de processamento por arquivo

## Uso

### Análise Básica

```go
service := service.NewLogDashService()
opts := service.DefaultAnalyzeOptions()
opts.RootPath = "/var/log"

result, _ := service.AnalyzeLogs(opts)

fmt.Printf("Files: %d discovered, %d processed\n",
    result.FilesDiscovered, result.FilesProcessed)
fmt.Printf("Entries: %d parsed\n", result.EntriesParsed)
fmt.Printf("Errors: %d, Warnings: %d\n",
    result.ByLevel[parser.LevelError],
    result.ByLevel[parser.LevelWarning])
```

### Últimas 24 Horas

```go
opts := service.DefaultAnalyzeOptions()
opts.RootPath = "/var/log"
opts.AggregateOpts.TimeRange = aggregator.TimeRange{
    Start: time.Now().Add(-24 * time.Hour),
    End:   time.Now(),
}

result, _ := service.AnalyzeLogs(opts)
```

### Processar Apenas Arquivos Recentes

```go
opts := service.DefaultAnalyzeOptions()
opts.RootPath = "/app/logs"
opts.MaxFilesToProcess = 5      // apenas 5 arquivos
opts.PrioritizeRecent = true    // os mais recentes

result, _ := service.AnalyzeLogs(opts)
```

### Análise Focada

```go
opts := service.DefaultAnalyzeOptions()
opts.RootPath = "/var/log"

// Discovery: apenas .log, até 3 níveis de profundidade
opts.DiscoverOpts.FilePatterns = []string{"*.log"}
opts.DiscoverOpts.MaxDepth = 3

// Aggregation: timeline de 5 minutos, top 20 mensagens
opts.AggregateOpts.TimelineBucketSize = 5 * time.Minute
opts.AggregateOpts.TopMessagesLimit = 20

result, _ := service.AnalyzeLogs(opts)
```

### Verificar Arquivos Processados

```go
result, _ := service.AnalyzeLogs(opts)

fmt.Println("Files processed:")
for _, fr := range result.FileResults {
    status := "✓"
    if fr.Error != nil {
        status = "✗"
    }
    fmt.Printf("%s %s: %d entries (%dms) [%s]\n",
        status, filepath.Base(fr.FilePath),
        fr.EntriesFound, fr.ProcessingTime, fr.ParserUsed)
}
```

### Error Handling

```go
result, err := service.AnalyzeLogs(opts)
if err != nil {
    // Erro crítico (path não existe, etc)
    log.Fatal(err)
}

// Erros recuperáveis (arquivo corrompido, etc)
if len(result.Errors) > 0 {
    fmt.Println("Warnings:")
    for _, e := range result.Errors {
        fmt.Printf("- [%s] %s: %v\n", e.Stage, e.FilePath, e.Error)
    }
}

// Processar mesmo com erros parciais
fmt.Printf("Successfully processed %d/%d files\n",
    result.FilesProcessed, result.FilesDiscovered)
```

## AnalysisResult

Estrutura completa retornada:

```go
type AnalysisResult struct {
    // Estatísticas gerais
    FilesDiscovered int
    FilesProcessed  int
    EntriesParsed   int
    
    // Resultado do Aggregator
    *aggregator.AggregateResult  // TotalEntries, ByLevel, BySource, 
                                  // Timeline, TopMessages, Anomalies, Summary
    
    // Detalhes por arquivo
    FileResults []FileResult
    
    // Erros encontrados
    Errors []ProcessingError
}
```

### FileResult

```go
type FileResult struct {
    FilePath       string
    EntriesFound   int
    ParserUsed     string
    ProcessingTime int64  // millisegundos
    Error          error
}
```

### ProcessingError

```go
type ProcessingError struct {
    FilePath    string
    Stage       string  // "discovery", "parsing", "aggregation"
    Error       error
    Recoverable bool
}
```

## Opções Configuráveis

```go
type AnalyzeOptions struct {
    RootPath          string
    DiscoverOpts      discovery.DiscoverOptions
    AggregateOpts     aggregator.AggregateOptions
    MaxFilesToProcess int
    PrioritizeRecent  bool
}
```

### Padrões

```go
DefaultAnalyzeOptions() = {
    DiscoverOpts: {
        MaxDepth: 10,
        FilePatterns: ["*.log", "*.txt", "*.out"],
        ExcludePaths: [".git", "node_modules", "vendor"],
    },
    AggregateOpts: {
        TopMessagesLimit: 10,
        TimelineBucketSize: 1h,
        DetectAnomalies: true,
    },
    MaxFilesToProcess: 0,  // sem limite
    PrioritizeRecent: true,
}
```

## Performance

### Benchmarks Típicos

| Arquivos | Entries Total | Tempo |
|----------|---------------|-------|
| 10 | ~10K | < 100ms |
| 50 | ~50K | < 500ms |
| 100 | ~100K | < 2s |

**Fatores que afetam:**
- Tamanho dos arquivos
- Complexidade dos parsers
- Detecção de anomalias
- Número de mensagens únicas

### Otimizações

```go
// Para análises rápidas
opts.MaxFilesToProcess = 10
opts.AggregateOpts.DetectAnomalies = false
opts.AggregateOpts.TopMessagesLimit = 5

// Para análise profunda
opts.MaxFilesToProcess = 0  // todos
opts.AggregateOpts.DetectAnomalies = true
opts.DiscoverOpts.MaxDepth = 15
```

## Casos de Uso

### 1. Health Check Rápido

```go
func quickHealthCheck(logPath string) string {
    service := service.NewLogDashService()
    opts := service.DefaultAnalyzeOptions()
    opts.RootPath = logPath
    opts.MaxFilesToProcess = 5
    opts.AggregateOpts.TimeRange = aggregator.TimeRange{
        Start: time.Now().Add(-1 * time.Hour),
    }
    
    result, _ := service.AnalyzeLogs(opts)
    
    errors := result.ByLevel[parser.LevelError]
    if errors > 10 {
        return "🔴 CRITICAL"
    } else if errors > 5 {
        return "🟡 WARNING"
    }
    return "🟢 HEALTHY"
}
```

### 2. Relatório Diário

```go
func dailyReport(logPath string) {
    service := service.NewLogDashService()
    opts := service.DefaultAnalyzeOptions()
    opts.RootPath = logPath
    opts.AggregateOpts.TimeRange = aggregator.TimeRange{
        Start: time.Now().Truncate(24 * time.Hour),
        End:   time.Now(),
    }
    
    result, _ := service.AnalyzeLogs(opts)
    
    fmt.Printf("Daily Report - %s\n", time.Now().Format("2006-01-02"))
    fmt.Println(result.Summary)
    fmt.Printf("\nTop Issues:\n")
    for i, msg := range result.TopMessages[:5] {
        if msg.Level >= parser.LevelWarning {
            fmt.Printf("%d. [%s] %s (x%d)\n",
                i+1, msg.Level, msg.Message, msg.Count)
        }
    }
}
```

### 3. Troubleshooting Específico

```go
func investigateIssue(logPath string, keyword string) {
    service := service.NewLogDashService()
    opts := service.DefaultAnalyzeOptions()
    opts.RootPath = logPath
    
    result, _ := service.AnalyzeLogs(opts)
    
    fmt.Printf("Searching for '%s'...\n", keyword)
    for _, msg := range result.TopMessages {
        if strings.Contains(strings.ToLower(msg.Message), 
                           strings.ToLower(keyword)) {
            fmt.Printf("[%s] %s\n", msg.Level, msg.Message)
            fmt.Printf("  Occurrences: %d\n", msg.Count)
            fmt.Printf("  First seen: %s\n", msg.FirstSeen)
            fmt.Printf("  Last seen: %s\n", msg.LastSeen)
            fmt.Printf("  Sources: %v\n", msg.Sources)
        }
    }
}
```

## Testes

```bash
go test -v
go test -cover
```

Cobertura: ~73% (integração de múltiplos componentes)

## Integração

Este service é usado pelas interfaces:
- **CLI**: Comando direto `logdash analyze /var/log`
- **Web UI**: Backend para dashboard
- **TUI**: Motor de análise do terminal

## Próximos Passos

Com o service pronto, você pode:
1. Criar CLI simples para testar
2. Implementar Web UI (htmx + Alpine.js)
3. Implementar TUI (bubbletea)
4. Adicionar cache (Indexer)

Veja também:
- [Parser Guide](../../docs/PARSER_GUIDE.md)
- [Discovery Guide](../../docs/DISCOVERY_GUIDE.md)
- [Aggregator Guide](../../docs/AGGREGATOR_GUIDE.md)
- [Architecture](../../docs/ARCHITECTURE.md)