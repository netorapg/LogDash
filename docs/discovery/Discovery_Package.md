# Discovery Package

Pacote responsável por descobrir arquivos de log no sistema de arquivos.

## Visão Rápida

```go
discoverer := discovery.NewFileDiscoverer()
opts := discovery.DefaultDiscoverOptions()

files, err := discoverer.Discover("/var/log", opts)
if err != nil {
    log.Fatal(err)
}

for _, file := range files {
    fmt.Printf("%s - %s (%d bytes)\n", 
        file.Path, file.DetectedFormat, file.Size)
}
```

## Features

- ✅ Descoberta recursiva com controle de profundidade
- ✅ Filtros por padrão de arquivo (glob patterns)
- ✅ Exclusão de diretórios (node_modules, .git, etc)
- ✅ Filtros por tamanho de arquivo
- ✅ Detecção básica de formato (python-logging, json, structured, generic)
- ✅ Tratamento gracioso de erros (permissões, symlinks)
- ✅ Funciona com diretórios e arquivos únicos

## Opções Principais

### MaxDepth
Controla profundidade de recursão:
- `0` = apenas root (não entra em subdiretórios)
- `1` = root + 1 nível
- `5` = recomendado para projetos
- `10` = padrão (cuidado com sistemas grandes)

### FilePatterns
Padrões glob para filtrar:
```go
opts.FilePatterns = []string{"*.log", "*.txt", "*.out"}
```

### ExcludePaths
Diretórios a ignorar:
```go
opts.ExcludePaths = []string{".git", "node_modules", "vendor"}
```

### Size Filters
```go
opts.MinSize = 1024        // apenas > 1KB
opts.MaxSize = 104857600   // apenas < 100MB
```

## Exemplos

### Descoberta Básica
```go
discoverer := NewFileDiscoverer()
opts := DefaultDiscoverOptions()
files, _ := discoverer.Discover("/path/to/project", opts)
```

### Descoberta Customizada
```go
opts := DiscoverOptions{
    MaxDepth:     3,
    FilePatterns: []string{"*.log"},
    ExcludePaths: []string{"archive", "backup"},
    MinSize:      100,
}
files, _ := discoverer.Discover("/app/logs", opts)
```

### Filtrar por Data (Pós-Descoberta)
```go
files, _ := discoverer.Discover("/var/log", opts)

// Filtrar apenas logs das últimas 24h
recent := []LogFile{}
cutoff := time.Now().Add(-24 * time.Hour)

for _, file := range files {
    if file.LastModified.After(cutoff) {
        recent = append(recent, file)
    }
}
```

## Performance

- **< 10ms**: ~100 arquivos, depth 3
- **< 100ms**: ~1,000 arquivos, depth 5
- **< 1s**: ~10,000 arquivos, depth 7

**Dicas:**
- Use `MaxDepth` apropriado
- Exclua `node_modules`, `.git`
- Use filtros de tamanho quando possível

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

Veja [DISCOVERY_GUIDE.md](../../../docs/DISCOVERY_GUIDE.md) para:
- Guia detalhado de uso
- Casos de uso comuns
- Troubleshooting
- Boas práticas
- Integração com parsers

## Implementação

- `NewFileDiscoverer()` - Implementação padrão usando `os.ReadDir`
- Thread-safe (stateless)
- Não modifica filesystem
- Continua mesmo com erros em subdiretórios