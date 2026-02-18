# Parser Package

Pacote responsável por processar logs em diferentes formatos e convertê-los em estruturas padronizadas `LogEntry`.

## Parsers Disponíveis

### GenericParser
Parser fallback que funciona com qualquer formato de log.

**Características:**
- Detecta níveis por palavras-chave (ERROR, WARN, INFO, etc)
- Extrai timestamps de formatos comuns
- Lida com múltiplos line endings (Unix, Windows)
- Score baixo-médio (0.2-0.3) para dar prioridade a parsers especializados

**Uso:**
```go
parser := NewGenericParser()
entries, err := parser.Parse(content)
```

**Exemplo de entrada:**
```
2024-01-15 10:30:00 ERROR Connection failed
2024-01-15 10:30:01 WARN Retrying...
INFO: Application started
```

## Implementando Novos Parsers

Veja o guia completo em [`docs/PARSER_GUIDE.md`](../../../docs/PARSER_GUIDE.md).

## Testes

```bash
# Rodar todos os testes
go test -v

# Ver cobertura
go test -cover

# Gerar relatório HTML de cobertura
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Arquitetura

```
Parser Interface
    ↓
├── GenericParser (implementado)
├── Log4jParser (em breve)
├── SyslogParser (em breve)
└── JSONParser (em breve)
```

Todos os parsers implementam a mesma interface, permitindo seleção automática baseada em score de confiança.