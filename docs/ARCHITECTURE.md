# Arquitetura do LogDash

Este documento descreve a arquitetura técnica do LogDash, as decisões de design e como os componentes interagem entre si.

## Visão Geral

O LogDash segue uma arquitetura em camadas com separação clara de responsabilidades. O core é independente de interface, permitindo múltiplas formas de apresentação (web e TUI) sem duplicação de lógica.
```
┌─────────────────────────────────────────────────┐
│          Camada de Apresentação                 │
│  ┌──────────────┐      ┌──────────────┐        │
│  │  Web UI      │      │     TUI      │        │
│  │ (htmx+Alpine)│      │ (bubbletea)  │        │
│  └──────┬───────┘      └──────┬───────┘        │
│         │                     │                 │
└─────────┼─────────────────────┼─────────────────┘
          │                     │
┌─────────┴─────────────────────┴─────────────────┐
│           Camada de Aplicação                   │
│  ┌──────────────────────────────────────┐      │
│  │       Application Service             │      │
│  │  (orquestra operações de alto nível)  │      │
│  └──────────────┬───────────────────────┘      │
└─────────────────┼──────────────────────────────┘
                  │
┌─────────────────┴──────────────────────────────┐
│              Camada de Core                     │
│  ┌─────────┐  ┌─────────┐  ┌──────────┐       │
│  │Discovery│  │ Parser  │  │Aggregator│       │
│  └────┬────┘  └────┬────┘  └─────┬────┘       │
│       │            │              │             │
│  ┌────┴────────────┴──────────────┴────┐       │
│  │          Indexer (Cache)             │       │
│  └──────────────────────────────────────┘       │
└─────────────────────────────────────────────────┘
```

## Camadas e Responsabilidades

### 1. Camada de Core

O core contém toda a lógica de negócio e é completamente independente de interface. Componentes principais:

#### Discovery (Descoberta de Arquivos)
**Responsabilidade**: Encontrar arquivos de log no sistema de arquivos.

**Interface**:
```go
type Discoverer interface {
    Discover(rootPath string, opts DiscoverOptions) ([]LogFile, error)
}

type LogFile struct {
    Path          string
    Size          int64
    LastModified  time.Time
    DetectedFormat string
}

type DiscoverOptions struct {
    MaxDepth      int
    FilePatterns  []string
    ExcludePaths  []string
    FollowSymlinks bool
}
```

**Comportamento**:
- Percorre recursivamente a partir de `rootPath`
- Identifica arquivos de log por extensão (`.log`, `.txt`, etc) e conteúdo
- Respeita limites de profundidade e exclusões
- Retorna metadados básicos sem ler conteúdo completo
- Não faz parsing, apenas catalogação

**Decisões de Design**:
- Usa goroutines para percorrer múltiplos diretórios em paralelo
- Implementa backpressure para não sobrecarregar memória com milhares de arquivos
- Detecta formato por amostragem (primeiras linhas) para otimizar

#### Parser (Análise de Logs)
**Responsabilidade**: Converter conteúdo bruto de logs em estruturas padronizadas.

**Interface**:
```go
type Parser interface {
    Parse(content []byte) ([]LogEntry, error)
    SupportsFormat(sample []byte) (confidence float64)
    Name() string
}

type LogEntry struct {
    Timestamp   time.Time
    Level       LogLevel
    Message     string
    Source      string      // arquivo de origem
    LineNumber  int
    Raw         string      // linha original
    Metadata    map[string]interface{}
}

type LogLevel int

const (
    LevelUnknown LogLevel = iota
    LevelTrace
    LevelDebug
    LevelInfo
    LevelWarning
    LevelError
    LevelFatal
)
```

**Parsers Implementados**:
- `Log4jParser` - formatos Log4j/Logback
- `SyslogParser` - RFC 3164/5424
- `JSONParser` - logs estruturados em JSON
- `GenericParser` - fallback para formatos desconhecidos

**Comportamento**:
- Cada parser retorna um score de confiança (0.0 a 1.0)
- Parser com maior confiança é selecionado
- Parsing parcial é aceitável - campos não reconhecidos vão para `Raw`
- Parsers são stateless e thread-safe

**Decisões de Design**:
- Arquitetura de plugins: fácil adicionar novos parsers
- Parsing incremental: processa em chunks para não carregar arquivo inteiro
- Timezone handling: tenta detectar ou assume UTC
- Campos opcionais via `Metadata` permitem parsers especializados sem quebrar interface

#### Aggregator (Agregação e Análise)
**Responsabilidade**: Processar múltiplos LogEntry e gerar insights.

**Interface**:
```go
type Aggregator interface {
    Aggregate(entries []LogEntry, opts AggregateOptions) (*AggregateResult, error)
}

type AggregateOptions struct {
    TimeRange    TimeRange
    GroupBy      []GroupByField
    Filters      []Filter
}

type AggregateResult struct {
    TotalEntries    int
    ByLevel         map[LogLevel]int
    BySource        map[string]int
    Timeline        []TimePoint
    TopMessages     []MessageFrequency
    Anomalies       []Anomaly
}

type TimePoint struct {
    Timestamp time.Time
    ByLevel   map[LogLevel]int
}

type Anomaly struct {
    Type        AnomalyType
    Description string
    Entries     []LogEntry
    Severity    float64
}
```

**Comportamento**:
- Gera estatísticas sobre distribuição de logs
- Identifica padrões temporais (picos, silêncios)
- Detecta anomalias (erros repetidos, mensagens incomuns)
- Otimizado para grandes volumes (processa em streaming)

**Decisões de Design**:
- Usa algoritmos aproximados para top-k (heavy hitters)
- Detecção de anomalias baseada em heurísticas simples (fase 1)
- Resultados são imutáveis e cacheáveis

#### Indexer (Sistema de Cache)
**Responsabilidade**: Persistir resultados processados para análises subsequentes.

**Interface**:
```go
type Indexer interface {
    Store(key string, data interface{}) error
    Load(key string, dest interface{}) error
    Invalidate(pattern string) error
    Clear() error
}

type IndexKey struct {
    FilePath     string
    FileModTime  time.Time
    Operation    string
}
```

**Comportamento**:
- Salva índices em disco (formato JSON ou gob)
- Valida se cache está atualizado comparando timestamps
- Suporta invalidação seletiva
- Localização configurável (temp dir ou pen drive)

**Decisões de Design**:
- Cache opcional - ferramenta funciona sem ele
- Chaves baseadas em hash de (path + modtime) para detectar mudanças
- Compressão para reduzir tamanho
- TTL configurável para evitar cache obsoleto

### 2. Camada de Aplicação

**Application Service** orquestra os componentes do core:
```go
type LogDashService struct {
    discoverer Discoverer
    parsers    []Parser
    aggregator Aggregator
    indexer    Indexer
}

func (s *LogDashService) AnalyzeLogs(rootPath string, opts AnalyzeOptions) (*AnalysisResult, error) {
    // 1. Descobrir arquivos
    files := s.discoverer.Discover(rootPath, opts.DiscoverOpts)
    
    // 2. Filtrar por data (últimas 24h por padrão)
    files = s.filterByTimeRange(files, opts.TimeRange)
    
    // 3. Parsear cada arquivo (em paralelo)
    entries := s.parseFiles(files)
    
    // 4. Agregar resultados
    result := s.aggregator.Aggregate(entries, opts.AggregateOpts)
    
    // 5. Cachear para próxima execução
    s.indexer.Store(cacheKey(rootPath), result)
    
    return result, nil
}
```

**Responsabilidades**:
- Coordenar fluxo de descoberta → parsing → agregação
- Gerenciar concorrência e paralelismo
- Implementar circuit breakers para arquivos problemáticos
- Fornecer API unificada para interfaces

### 3. Camada de Apresentação

#### Web UI
**Stack**: Go (servidor) + htmx + Alpine.js + CSS

**Estrutura**:
```
internal/web/
├── server/
│   └── server.go          # HTTP server setup
├── handlers/
│   ├── dashboard.go       # página principal
│   ├── api.go            # endpoints JSON para htmx
│   └── filters.go        # lógica de filtros
└── templates/
    ├── layout.html
    ├── dashboard.html
    └── components/
        ├── timeline.html
        ├── filters.html
        └── log-table.html
```

**Fluxo**:
1. Servidor Go renderiza HTML inicial com template engine
2. htmx faz requisições para atualizar partes da página
3. Alpine.js gerencia estado local (filtros abertos/fechados, etc)
4. Servidor retorna HTML parcial que htmx injeta

**Endpoints**:
- `GET /` - dashboard principal
- `GET /api/logs` - logs filtrados (retorna HTML table)
- `GET /api/stats` - estatísticas (retorna HTML cards)
- `GET /api/timeline` - dados da timeline (retorna HTML chart)

#### TUI
**Stack**: Go + bubbletea (Elm Architecture)

**Componentes**:
```go
type Model struct {
    logs        []LogEntry
    stats       *AggregateResult
    viewport    viewport.Model
    filters     FilterState
    focusedPane Pane
}

type Msg interface{}

func (m Model) Update(msg Msg) (Model, tea.Cmd)
func (m Model) View() string
```

**Layout**:
```
┌─────────────────────────────────────────────────┐
│ LogDash - /path/to/project           [q] Quit  │
├────────────┬────────────────────────────────────┤
│  Stats     │  Timeline                          │
│            │  ████████░░░░                      │
│ Errors: 42 │  ░░████████░░                      │
│ Warns:  18 │  ░░░░████████                      │
│ Info:  203 │                                    │
├────────────┴────────────────────────────────────┤
│ Filters: [✓] Error [✓] Warn [ ] Info           │
├─────────────────────────────────────────────────┤
│ 2024-01-15 10:30:00 ERROR Connection failed    │
│ 2024-01-15 10:30:02 ERROR Timeout after 30s    │
│ 2024-01-15 10:31:15 WARN  Retrying operation   │
│ ...                                             │
└─────────────────────────────────────────────────┘
```

## Fluxo de Dados

### Inicialização
```
1. main() → cria LogDashService com dependências
2. Lê flags CLI (--path, --port, --tui)
3. Se --tui: inicia TUI
   Senão: inicia Web Server + mostra URL
```

### Análise de Logs (primeira execução)
```
1. Usuário acessa /
2. Handler chama service.AnalyzeLogs(path)
3. Service:
   a. Discovery encontra arquivos
   b. Filtra últimas 24h
   c. Para cada arquivo (paralelo):
      - Seleciona parser
      - Parse em chunks
   d. Aggregator processa entries
   e. Salva resultado em cache
4. Template renderiza dashboard
```

### Análise Subsequente (com cache)
```
1. service.AnalyzeLogs(path) verifica cache
2. Se válido: retorna resultado cacheado
3. Se inválido: reprocessa apenas arquivos novos/modificados
4. Merge com cache existente
```

### Filtragem Interativa (htmx)
```
1. Usuário clica filtro "apenas Errors"
2. htmx: GET /api/logs?level=error
3. Handler filtra entries em memória
4. Retorna HTML table com resultados
5. htmx substitui <div id="logs-table">
```

## Decisões Arquiteturais

### Por que Interfaces?
- Facilita testes (mocks)
- Permite trocar implementações (ex: parser otimizado)
- Suporta plugins futuros

### Por que Stateless Parsers?
- Thread-safe por design
- Simplifica paralelização
- Facilita caching

### Por que Cache Opcional?
- Ferramenta funciona mesmo sem disco
- Útil em análises pontuais
- Essencial para projetos grandes

### Por que Processar em Paralelo?
- Aproveita múltiplos cores
- Reduz tempo total drasticamente
- Sistema de logs é embaraçosamente paralelo

### Por que Priorizar Últimas 24h?
- Maioria dos casos de uso são recentes
- Entrega valor rápido
- Evita timeout em projetos enormes

## Testabilidade

Cada camada é testável isoladamente:
```go
// Core - testes unitários puros
func TestLog4jParser_Parse(t *testing.T) {
    parser := NewLog4jParser()
    // testa parsing sem I/O
}

// Application - testes com mocks
func TestLogDashService_AnalyzeLogs(t *testing.T) {
    mockDiscoverer := &MockDiscoverer{...}
    service := NewLogDashService(mockDiscoverer, ...)
    // testa orquestração
}

// Web - testes de integração
func TestDashboardHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/", nil)
    rr := httptest.NewRecorder()
    handler.ServeHTTP(rr, req)
    // verifica resposta HTTP
}
```

## Extensibilidade

### Adicionar Novo Parser
1. Implementar interface `Parser`
2. Registrar em `parsers` slice
3. Adicionar testes

### Adicionar Nova Interface
1. Criar em `internal/newui/`
2. Consumir `LogDashService`
3. Não precisa tocar no core

### Adicionar Detecção de Anomalias
1. Estender `Aggregator`
2. Adicionar campo em `AggregateResult`
3. UI consome automaticamente

## Performance

### Targets
- Descoberta: ~1000 arquivos/segundo
- Parsing: ~100MB/s por core
- Agregação: ~1M entries em <2s

### Otimizações Planejadas
- [ ] Memory pooling para reduzir GC
- [ ] Parser SIMD para formatos comuns
- [ ] Compressão de cache mais eficiente

## Segurança

- **Sem execução de código**: apenas lê arquivos
- **Sem network**: tudo local (exceto localhost)
- **Sem escrita no projeto**: apenas leitura
- **Cache isolado**: não mistura projetos

## Próximos Passos Arquiteturais

- [ ] Sistema de plugins para parsers customizados
- [ ] Streaming para logs em tempo real (fase 2)
- [ ] Exportação para formatos externos (CSV, JSON)
- [ ] Correlação entre múltiplos serviços

---

Esta arquitetura prioriza simplicidade, testabilidade e extensibilidade, mantendo o core focado e as interfaces desacopladas.