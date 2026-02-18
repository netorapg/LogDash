# Discovery System Guide

Este guia explica como o sistema de descoberta de logs do LogDash funciona e como utilizá-lo.

## Visão Geral

O Discovery é responsável por vasculhar o sistema de arquivos e identificar arquivos de log. Ele retorna metadados sobre os arquivos encontrados (path, tamanho, última modificação) sem carregar o conteúdo completo na memória.

### Fluxo de Descoberta

```
Root Path → Descoberta Recursiva → Filtros → []LogFile
                ↓
         ┌──────┴──────┐
         │   Opções    │
         ├─────────────┤
         │ MaxDepth    │
         │ Patterns    │
         │ Excludes    │
         │ Size limits │
         └─────────────┘
```

## Componentes Principais

### Discoverer Interface

Define o contrato para descoberta de arquivos:

```go
type Discoverer interface {
    Discover(rootPath string, opts DiscoverOptions) ([]LogFile, error)
}
```

**Comportamento:**
- Retorna `[]LogFile{}` (vazio) quando não encontrar arquivos, nunca `nil`
- Retorna erro apenas para problemas críticos (path não existe, sem permissão no root)
- Continua a descoberta mesmo com erros em subdiretórios específicos

### LogFile

Representa um arquivo de log encontrado:

```go
type LogFile struct {
    Path           string    // caminho absoluto do arquivo
    Size           int64     // tamanho em bytes
    LastModified   time.Time // última modificação
    DetectedFormat string    // formato detectado ("python-logging", "json", "generic", etc)
}
```

**Campos:**
- **Path**: Sempre absoluto para evitar ambiguidade
- **Size**: Útil para filtros e estimativa de tempo de processamento
- **LastModified**: Permite priorizar logs recentes e cache invalidation
- **DetectedFormat**: Pré-identificação do formato para seleção de parser

### DiscoverOptions

Configura o comportamento da descoberta:

```go
type DiscoverOptions struct {
    MaxDepth       int      // profundidade máxima de recursão
    FilePatterns   []string // padrões glob (*.log, *.txt)
    ExcludePaths   []string // diretórios/arquivos a ignorar
    FollowSymlinks bool     // seguir links simbólicos
    MinSize        int64    // tamanho mínimo em bytes
    MaxSize        int64    // tamanho máximo em bytes
    SampleSize     int      // bytes a ler para detectar formato
}
```

#### MaxDepth (Profundidade)

Controla quantos níveis de subdiretórios explorar:

- **0**: Apenas o diretório root (não entra em subdiretórios)
- **1**: Root + 1 nível de subdiretórios
- **2**: Root + 2 níveis de subdiretórios
- **>10**: Use com cautela em sistemas grandes

**Exemplos:**
```
Estrutura:
/project/
  app.log          ← depth 0
  logs/
    error.log      ← depth 1
    archive/
      old.log      ← depth 2

MaxDepth=0 → encontra: app.log
MaxDepth=1 → encontra: app.log, error.log
MaxDepth=2 → encontra: app.log, error.log, old.log
```

#### FilePatterns (Padrões de Arquivo)

Lista de padrões glob para filtrar arquivos:

```go
opts.FilePatterns = []string{"*.log", "*.txt", "*.out"}
```

**Sintaxe glob suportada:**
- `*` - qualquer sequência de caracteres
- `?` - qualquer caractere único
- `[abc]` - qualquer caractere no conjunto
- `*.log` - todos os arquivos .log

**Padrões padrão:** `["*.log", "*.txt", "*.out"]`

**Importante:** Patterns aplicam-se ao **nome do arquivo**, não ao path completo.

#### ExcludePaths (Exclusões)

Diretórios ou arquivos a ignorar completamente:

```go
opts.ExcludePaths = []string{".git", "node_modules", "vendor"}
```

**Exclusões padrão:**
- `.git` - repositório Git
- `node_modules` - dependências Node.js
- `vendor` - dependências Go/PHP
- `__pycache__` - cache Python
- `.venv` - ambiente virtual Python

**Como funciona:**
- Verifica nome exato do diretório/arquivo
- Verifica se o path completo contém o padrão
- Aplicado antes de entrar em subdiretórios (eficiente)

#### FollowSymlinks

Controla se deve seguir links simbólicos:

```go
opts.FollowSymlinks = false // padrão (mais seguro)
```

**Recomendação:** Manter `false` para evitar:
- Loops infinitos (symlink circular)
- Acessar áreas não intencionadas do sistema
- Problemas de permissão

**Quando usar `true`:**
- Logs armazenados via symlinks
- Ambiente controlado sem risco de loops

#### Size Filters (MinSize / MaxSize)

Filtram arquivos por tamanho:

```go
opts.MinSize = 1024      // apenas arquivos > 1KB
opts.MaxSize = 10485760  // apenas arquivos < 10MB
```

**Valores:**
- `0` = sem limite (padrão)
- Valores em bytes

**Casos de uso:**
- `MinSize`: Ignorar arquivos vazios ou muito pequenos
- `MaxSize`: Evitar processar logs gigantes que podem travar

#### SampleSize

Quantos bytes ler para detectar formato:

```go
opts.SampleSize = 1024 // padrão: 1KB
```

**Impacto:**
- Maior = detecção mais precisa, mas mais I/O
- Menor = mais rápido, mas pode errar formato
- Recomendado: 512-2048 bytes

## Implementação: FileDiscoverer

A implementação padrão que usa `os.ReadDir` e `filepath.Walk`:

```go
discoverer := NewFileDiscoverer()
```

### Características

**✅ Eficiente:**
- Não lê conteúdo completo dos arquivos
- Para a recursão ao atingir MaxDepth
- Pula diretórios excluídos imediatamente

**✅ Robusto:**
- Continua mesmo com erros de permissão
- Trata graciosamente paths inválidos
- Funciona com path sendo arquivo ou diretório

**✅ Seguro:**
- Não segue symlinks por padrão
- Respeita limites de tamanho
- Não modifica nada no filesystem

## Uso Básico

### Descoberta Simples

```go
package main

import (
    "fmt"
    "log"
    "github.com/seu-usuario/logdash/internal/core/discovery"
)

func main() {
    // Criar discoverer
    discoverer := discovery.NewFileDiscoverer()
    
    // Usar opções padrão
    opts := discovery.DefaultDiscoverOptions()
    
    // Descobrir logs
    files, err := discoverer.Discover("/var/log", opts)
    if err != nil {
        log.Fatal(err)
    }
    
    // Processar resultados
    for _, file := range files {
        fmt.Printf("%s (%d bytes, %s)\n", 
            file.Path, file.Size, file.DetectedFormat)
    }
}
```

### Descoberta Customizada

```go
// Criar opções personalizadas
opts := discovery.DiscoverOptions{
    MaxDepth:       3,                    // até 3 níveis
    FilePatterns:   []string{"*.log"},    // apenas .log
    ExcludePaths:   []string{"archive"},  // ignorar archive/
    FollowSymlinks: false,
    MinSize:        1024,                 // > 1KB
    MaxSize:        104857600,            // < 100MB
    SampleSize:     2048,                 // 2KB para detecção
}

files, err := discoverer.Discover("/home/user/project", opts)
```

### Descoberta com Filtros Progressivos

```go
// Primeira passada: todos os logs recentes
opts1 := discovery.DefaultDiscoverOptions()
opts1.MaxDepth = 1
allRecent, _ := discoverer.Discover("/app/logs", opts1)

// Segunda passada: apenas erros grandes (para análise profunda)
opts2 := discovery.DefaultDiscoverOptions()
opts2.FilePatterns = []string{"*error*.log"}
opts2.MinSize = 10240 // > 10KB
errors, _ := discoverer.Discover("/app/logs", opts2)
```

### Descoberta de Arquivo Único

```go
// Pode apontar para um arquivo específico
files, err := discoverer.Discover("/var/log/app.log", opts)
// Retorna slice com 1 elemento se o arquivo passa nos filtros
```

## Detecção de Formato

O Discovery faz pré-detecção do formato lendo uma amostra:

### Formatos Detectados

| Formato | Características | Exemplo |
|---------|----------------|---------|
| `python-logging` | `[LEVEL]` + timestamp com vírgula | `[INFO] 2024-01-15 10:30:00,181` |
| `json` | Começa com `{` | `{"level":"info","msg":"test"}` |
| `structured` | Contém keywords de level | `2024-01-15 ERROR Connection failed` |
| `generic` | Qualquer outro | Logs sem padrão claro |

### Heurísticas de Detecção

```go
// python-logging: [LEVEL] + timestamp com vírgula
if strings.Contains(sample, "[INFO]") && strings.Contains(sample, ",") {
    return "python-logging"
}

// JSON: começa com {
if strings.HasPrefix(strings.TrimSpace(sample), "{") {
    return "json"
}

// Structured: tem keywords de severidade
if strings.Contains(sample, "ERROR") || strings.Contains(sample, "WARN") {
    return "structured"
}

// Fallback
return "generic"
```

**Nota:** Detecção é **heurística**, não garantida. Os parsers fazem detecção mais robusta depois.

## Casos de Uso Comuns

### 1. Descoberta Rápida (Últimos Logs)

```go
opts := discovery.DefaultDiscoverOptions()
opts.MaxDepth = 1  // apenas root + 1 nível
opts.MaxSize = 10485760  // < 10MB

files, _ := discoverer.Discover("/var/log", opts)
```

### 2. Descoberta Profunda (Projeto Completo)

```go
opts := discovery.DefaultDiscoverOptions()
opts.MaxDepth = 10  // explorar tudo
opts.MinSize = 0    // incluir arquivos vazios
opts.ExcludePaths = []string{
    ".git", "node_modules", "vendor", 
    "dist", "build", "__pycache__",
}

files, _ := discoverer.Discover("/home/user/project", opts)
```

### 3. Descoberta de Logs de Produção

```go
opts := discovery.DiscoverOptions{
    MaxDepth:     5,
    FilePatterns: []string{"*.log", "*.log.*"}, // inclui rotacionados
    ExcludePaths: []string{"debug", "test"},
    MinSize:      100,  // > 100 bytes
    MaxSize:      1073741824,  // < 1GB
}

files, _ := discoverer.Discover("/var/log/production", opts)
```

### 4. Descoberta de Logs Específicos

```go
opts := discovery.DefaultDiscoverOptions()
opts.FilePatterns = []string{"error*.log", "*exception*.log"}

errors, _ := discoverer.Discover("/app", opts)
```

## Performance

### Benchmarks Estimados

| Cenário | Arquivos | Profundidade | Tempo |
|---------|----------|--------------|-------|
| Pequeno | ~100 | 3 | < 10ms |
| Médio | ~1,000 | 5 | < 100ms |
| Grande | ~10,000 | 7 | < 1s |
| Enorme | ~100,000 | 10 | < 10s |

**Fatores que afetam:**
- Velocidade do disco (SSD vs HDD)
- Número de diretórios excluídos
- Profundidade máxima
- Tamanho da amostra para detecção

### Otimizações

**✅ Use MaxDepth apropriado:**
```go
// ❌ Ruim: explora tudo desnecessariamente
opts.MaxDepth = 0  // sem limite

// ✅ Bom: profundidade sensata
opts.MaxDepth = 5
```

**✅ Exclua diretórios grandes:**
```go
opts.ExcludePaths = []string{
    "node_modules",  // pode ter 10,000+ arquivos
    ".git",
    "dist",
}
```

**✅ Use filtros de tamanho:**
```go
// Ignorar arquivos muito pequenos ou muito grandes
opts.MinSize = 100
opts.MaxSize = 104857600  // 100MB
```

## Tratamento de Erros

### Erros Que Retornam Error

```go
files, err := discoverer.Discover(path, opts)
if err != nil {
    // Path não existe
    // Sem permissão para ler o diretório root
    // Path é inválido
}
```

### Erros Silenciosos (Continuam)

- **Permissão negada em subdiretório:** Pula e continua
- **Symlink quebrado:** Pula e continua
- **Arquivo sendo escrito:** Pula e continua

### Exemplo de Tratamento Robusto

```go
files, err := discoverer.Discover(rootPath, opts)
if err != nil {
    log.Printf("Failed to discover: %v", err)
    return
}

if len(files) == 0 {
    log.Println("No log files found matching criteria")
    return
}

log.Printf("Found %d log files", len(files))
for _, file := range files {
    log.Printf("  - %s (%d bytes)", file.Path, file.Size)
}
```

## Integração com Parsers

O Discovery prepara o terreno para os Parsers:

```go
// 1. Descobrir arquivos
discoverer := discovery.NewFileDiscoverer()
files, _ := discoverer.Discover("/var/log", opts)

// 2. Para cada arquivo, usar parser apropriado
for _, file := range files {
    // DetectedFormat sugere qual parser usar
    var parser parser.Parser
    
    switch file.DetectedFormat {
    case "python-logging":
        parser = parser.NewPythonLoggingParser()
    case "json":
        parser = parser.NewJSONParser()
    default:
        parser = parser.NewGenericParser()
    }
    
    // 3. Ler e parsear conteúdo
    content, _ := os.ReadFile(file.Path)
    entries, _ := parser.Parse(content)
    
    // 4. Processar entries...
}
```

## Testabilidade

### Estrutura de Teste

```go
func TestMyFeature(t *testing.T) {
    // Criar estrutura de teste
    tmpDir := t.TempDir()
    
    os.WriteFile(filepath.Join(tmpDir, "app.log"), []byte("log"), 0644)
    os.MkdirAll(filepath.Join(tmpDir, "logs"), 0755)
    os.WriteFile(filepath.Join(tmpDir, "logs/error.log"), []byte("error"), 0644)
    
    // Testar
    discoverer := discovery.NewFileDiscoverer()
    files, err := discoverer.Discover(tmpDir, opts)
    
    // Validar
    assert.NoError(t, err)
    assert.Len(t, files, 2)
}
```

## Boas Práticas

### ✅ FAZER

- Usar `DefaultDiscoverOptions()` como ponto de partida
- Definir `MaxDepth` apropriado para o contexto
- Excluir diretórios conhecidos que não têm logs
- Tratar resultado vazio como válido (não erro)
- Usar `MinSize` para ignorar arquivos vazios

### ❌ EVITAR

- `MaxDepth = 0` (sem limite) em sistemas grandes
- Seguir symlinks sem necessidade (`FollowSymlinks = true`)
- Processar todos os arquivos sem filtros
- Assumir que `DetectedFormat` é sempre correto
- Ignorar erros sem logging

## Troubleshooting

### "No files found" mas existem logs

**Possíveis causas:**
1. `FilePatterns` não inclui a extensão dos arquivos
2. `MaxDepth` muito baixo
3. Arquivos em diretórios excluídos
4. Filtros de tamanho muito restritivos

**Solução:**
```go
// Debug: remover filtros temporariamente
opts := discovery.DiscoverOptions{
    MaxDepth:     10,
    FilePatterns: []string{"*"}, // todos os arquivos
    ExcludePaths: []string{},    // sem exclusões
}
```

### Discovery muito lento

**Causas:**
1. `MaxDepth` muito alto
2. Muitos arquivos sendo verificados
3. `SampleSize` muito grande

**Solução:**
```go
opts.MaxDepth = 3        // reduzir profundidade
opts.SampleSize = 512    // reduzir amostra
opts.ExcludePaths = append(opts.ExcludePaths, "node_modules")
```

### "Permission denied"

**Causa:** Sem permissão para ler diretório

**Solução:**
- Rodar com permissões adequadas
- Excluir diretórios problemáticos
- Verificar ownership dos arquivos

## Próximos Passos

Após descobrir os arquivos:
1. **Filtrar por data** (usando `LastModified`)
2. **Priorizar** (logs mais recentes ou maiores primeiro)
3. **Parsear** (usando parsers apropriados)
4. **Agregar** (combinar resultados)

Veja também:
- [Parser Guide](PARSER_GUIDE.md) - Como processar os arquivos encontrados
- [Architecture](../ARCHITECTURE.md) - Visão geral do sistema

---

**Dúvidas?** Abra uma issue com label `question` ou consulte os exemplos de teste.