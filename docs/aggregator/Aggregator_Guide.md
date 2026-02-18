# Aggregator System Guide

Guia completo do sistema de agregação do LogDash.

## O que é o Aggregator?

Transforma `[]LogEntry` em insights acionáveis através de múltiplas análises.

## Principais Features

- **Estatísticas**: Total, por nível, por source
- **Timeline**: Distribuição temporal com buckets
- **Top Messages**: Mensagens mais frequentes
- **Anomalias**: Error burst, spikes, repeated errors
- **Summary**: Resumo textual automático

## Uso Básico

```go
aggregator := aggregator.NewLogAggregator()
result, _ := aggregator.Aggregate(entries, aggregator.DefaultAggregateOptions())

fmt.Println(result.Summary)
fmt.Printf("Errors: %d\n", result.ByLevel[parser.LevelError])
```

## Documentação Completa

Veja `internal/core/aggregator/README.md` para guia detalhado com exemplos.