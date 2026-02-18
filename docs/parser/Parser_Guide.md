# Parser System Guide

Este guia explica como o sistema de parsers do LogDash funciona e como criar novos parsers personalizados.

## Visão Geral

O sistema de parsers é responsável por converter logs em formatos variados (texto bruto, Log4j, syslog, JSON, etc) em uma estrutura padronizada `LogEntry`, permitindo análise uniforme.

### Arquitetura

```
Arquivo de Log → Parser Selection → Parsing → []LogEntry
                      ↑
                Score de Confiança
```

1. **Parser Selection**: Múltiplos parsers competem analisando uma amostra do arquivo
2. **Scoring**: Cada parser retorna um score (0.0-1.0) de confiança
3. **Melhor Parser**: O parser com maior score é selecionado
4. **Parsing**: O parser escolhido processa o arquivo completo

## Estruturas Principais

### LogEntry

Estrutura normalizada que representa uma entrada de log:

```go
type LogEntry struct {
    Timestamp  time.Time              // quando aconteceu (zero se não detectado)
    Level      LogLevel               // severidade do log
    Message    string                 // mensagem principal
    Source     string                 // arquivo de origem (preenchido externamente)
    LineNumber int                    // número da linha no arquivo
    Raw        string                 // linha original sem processamento
    Metadata   map[string]interface{} // dados adicionais específicos do formato
}
```

**Campos Obrigatórios vs Opcionais:**
- **Sempre preencher**: `Raw`, `LineNumber`, `Message`
- **Tentar preencher**: `Level`, `Timestamp`
- **Opcional**: `Metadata` (para informações específicas do formato)
- **Não preencher**: `Source` (será preenchido pela camada superior)

### LogLevel

Níveis de severidade padronizados:

```go
const (
    LevelUnknown  // não foi possível determinar
    LevelTrace    // mais detalhado
    LevelDebug    // informações de desenvolvimento
    LevelInfo     // eventos normais
    LevelWarning  // potencialmente problemático
    LevelError    // erro não-fatal
    LevelFatal    // erro crítico/fatal
)
```

## Interface Parser

Todo parser deve implementar três métodos:

```go
type Parser interface {
    Parse(content []byte) ([]LogEntry, error)
    SupportsFormat(sample []byte) float64
    Name() string
}
```

### 1. Parse

Processa o conteúdo completo e retorna slice de `LogEntry`.

**Regras:**
- Retornar `[]LogEntry{}` (vazio) quando não houver entradas, nunca `nil`
- Retornar erro apenas para falhas críticas (arquivo corrompido, etc)
- Parsing parcial não é erro - extraia o que conseguir
- Processar linha por linha é a abordagem comum
- Pular linhas vazias automaticamente

**Exemplo básico:**

```go
func (p *myParser) Parse(content []byte) ([]LogEntry, error) {
    var entries []LogEntry
    lines := strings.Split(string(content), "\n")
    
    for i, line := range lines {
        if len(strings.TrimSpace(line)) == 0 {
            continue // pula linhas vazias
        }
        
        entry := LogEntry{
            LineNumber: i + 1,
            Raw:        line,
            Message:    extractMessage(line),
            Level:      extractLevel(line),
            Timestamp:  extractTimestamp(line),
            Metadata:   make(map[string]interface{}),
        }
        
        entries = append(entries, entry)
    }
    
    return entries, nil
}
```

### 2. SupportsFormat

Analisa uma amostra e retorna score de confiança (0.0 a 1.0).

**Diretrizes de Scoring:**
- **0.0-0.3**: Baixa - formato não reconhecido ou muito genérico
- **0.4-0.7**: Média - algumas características identificadas
- **0.8-1.0**: Alta - formato claramente identificado

**Estratégias:**
- Procurar por padrões únicos do formato (ex: `[INFO]` para Log4j)
- Validar estrutura (ex: JSON válido)
- Contar ocorrências de marcadores característicos
- Quanto mais específico o parser, maior deve ser o score quando reconhecer

**Exemplo:**

```go
func (p *log4jParser) SupportsFormat(sample []byte) float64 {
    text := string(sample)
    lines := strings.Split(text, "\n")
    
    matches := 0
    for _, line := range lines {
        // Padrão típico Log4j: [YYYY-MM-DD HH:MM:SS] LEVEL
        if regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] [A-Z]+`).MatchString(line) {
            matches++
        }
    }
    
    // Se >50% das linhas batem o padrão, alta confiança
    if len(lines) > 0 && float64(matches)/float64(len(lines)) > 0.5 {
        return 0.9
    }
    
    return 0.2
}
```

### 3. Name

Retorna identificador único do parser.

```go
func (p *myParser) Name() string {
    return "my-format"
}
```

**Convenções:**
- Lowercase com hífens
- Descritivo mas conciso
- Exemplos: "log4j", "syslog", "json", "generic"

## Criando um Novo Parser

### Passo 1: Criar Arquivo

Crie `internal/core/parser/myformat.go`:

```go
package parser

type myFormatParser struct{}

func NewMyFormatParser() Parser {
    return &myFormatParser{}
}

func (p *myFormatParser) Name() string {
    return "my-format"
}

func (p *myFormatParser) SupportsFormat(sample []byte) float64 {
    // Implementar detecção
    return 0.0
}

func (p *myFormatParser) Parse(content []byte) ([]LogEntry, error) {
    // Implementar parsing
    return []LogEntry{}, nil
}
```

### Passo 2: Criar Testes (TDD!)

Crie `internal/core/parser/myformat_test.go`:

```go
package parser

import "testing"

func TestMyFormatParser_Parse(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        wantCount int
        wantErr   bool
    }{
        {
            name:      "valid log line",
            input:     "your sample log here",
            wantCount: 1,
            wantErr:   false,
        },
        // mais casos...
    }
    
    parser := NewMyFormatParser()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            entries, err := parser.Parse([]byte(tt.input))
            
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if len(entries) != tt.wantCount {
                t.Errorf("got %v entries, want %v", len(entries), tt.wantCount)
            }
        })
    }
}
```

### Passo 3: Implementar

1. Escreva os testes primeiro (TDD)
2. Implemente o código para passar nos testes
3. Rode `go test -v` para verificar
4. Rode `go test -cover` para ver cobertura

### Passo 4: Registrar (futuro)

No futuro, quando tivermos o sistema de seleção de parser, você registrará seu parser:

```go
// Em algum lugar central
parsers := []Parser{
    NewGenericParser(),
    NewLog4jParser(),
    NewMyFormatParser(), // seu novo parser
}
```

## Boas Práticas

### ✅ FAZER

- Escrever testes antes do código (TDD)
- Ser permissivo - extrair o que conseguir mesmo que o formato esteja "errado"
- Documentar padrões esperados com exemplos
- Usar funções auxiliares privadas para lógica complexa
- Testar edge cases (linhas vazias, muito longas, caracteres especiais)
- Manter parsers stateless (sem estado interno)

### ❌ EVITAR

- Retornar erro para parsing parcial
- Assumir encoding específico
- Modificar o conteúdo original (usar `Raw` field)
- Depender de ordem específica de campos
- Fazer I/O ou rede dentro do parser
- Manter estado entre chamadas

## Padrões Comuns

### Detectando Timestamps

```go
formats := []string{
    "2006-01-02 15:04:05",
    "2006-01-02T15:04:05Z",
    "Jan 02 15:04:05",
}

for _, format := range formats {
    if t, err := time.Parse(format, text); err == nil {
        return t
    }
}
```

### Detectando Níveis

```go
upperLine := strings.ToUpper(line)

switch {
case strings.Contains(upperLine, "FATAL"):
    return LevelFatal
case strings.Contains(upperLine, "ERROR"):
    return LevelError
case strings.Contains(upperLine, "WARN"):
    return LevelWarning
// etc...
}
```

### Usando Regex com Cautela

```go
// OK: regex simples e específica
pattern := regexp.MustCompile(`^\[(\d{4}-\d{2}-\d{2})\]`)

// EVITAR: regex complexa que pode ser lenta
pattern := regexp.MustCompile(`^.*\[(.*?)\].*\[(.*?)\].*$`)
```

### Populando Metadata

```go
entry.Metadata = map[string]interface{}{
    "thread_id": threadID,
    "class_name": className,
    "method": methodName,
}
```

## Exemplos de Referência

Veja os parsers já implementados:

- **GenericParser** (`generic.go`) - Exemplo de parser simples e permissivo
- **Log4jParser** (`log4j.go`) - Exemplo com regex e múltiplos formatos (em breve)
- **JSONParser** (`json.go`) - Exemplo com unmarshaling estruturado (em breve)

## Debugging

### Testando Manualmente

```go
parser := NewMyFormatParser()
sample := []byte("your log content here")

// Testar detecção
score := parser.SupportsFormat(sample)
fmt.Printf("Confidence: %.2f\n", score)

// Testar parsing
entries, err := parser.Parse(sample)
fmt.Printf("Parsed %d entries\n", len(entries))
for _, e := range entries {
    fmt.Printf("%s [%s] %s\n", e.Timestamp, e.Level, e.Message)
}
```

### Verificando Cobertura

```bash
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Contribuindo

Ao criar um novo parser:

1. Abra issue descrevendo o formato que quer suportar
2. Crie branch `feature/parser-formatname`
3. Siga TDD - testes primeiro!
4. Garanta cobertura >80%
5. Adicione exemplo no README se for formato comum
6. Abra Pull Request com descrição clara

---

**Dúvidas?** Abra uma issue com label `question` ou consulte os exemplos existentes.