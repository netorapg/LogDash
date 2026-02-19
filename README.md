# LogDash

> Um painel visual e interativo para investigação de logs, projetado para devs que precisam entender o que aconteceu no sistema.

## O Problema

Sistemas complexos geram logs em múltiplas pastas, formatos variados, e volumes crescentes. Investigar problemas significa navegar por dezenas de arquivos, usar grep repetidamente, e tentar reconstruir mentalmente a sequência de eventos. Warnings ficam invisíveis no meio do ruído.

## A Solução

LogDash é uma ferramenta portátil que:
- 🔍 Descobre automaticamente logs em qualquer projeto
- 📊 Apresenta uma visão agregada e visual em segundos
- 🎯 Destaca padrões, tendências e warnings ignorados
- 🚀 Roda sem instalação — basta executar
- 💻 Interface CLI poderosa e intuitiva

## Quick Start

### Instalação

```bash
# Clone o repositório
git clone https://github.com/netorapg/LogDash.git
cd LogDash

# Build
go build -o logdash ./cmd/logdash

# Executar
./logdash analyze -path /var/log
```

### Uso Básico

```bash
# Analisar diretório atual
logdash analyze

# Últimas 24 horas
logdash analyze -path /var/log -last 24h

# Com limites
logdash analyze -path /var/log -max-files 10 -verbose
```

## Features

### 🔍 Discovery Inteligente
- ✅ Busca recursiva automática
- ✅ Filtros por padrão (*.log, *.txt, etc)
- ✅ Exclusão de diretórios (.git, node_modules)
- ✅ Controle de profundidade

### 🎯 Parsing Automático
- ✅ Detecta formatos automaticamente
- ✅ Suporta Python logging, JSON, logs genéricos
- ✅ Seleção inteligente de parser por confidence score
- ✅ Fallback para formato desconhecido

### 📊 Análise Avançada
- ✅ Estatísticas por nível (Error, Warning, Info, etc)
- ✅ Timeline temporal com buckets configuráveis
- ✅ Top mensagens mais frequentes
- ✅ Detecção de anomalias (error bursts, spikes)
- ✅ Resumo automático

### ⚡ Performance
- ✅ Processa 100K+ entries em segundos
- ✅ Processamento paralelo
- ✅ Filtros configuráveis para otimização

## Arquitetura

```
┌─────────────────────────────────────────────┐
│              CLI Interface                   │
└────────────────┬────────────────────────────┘
                 │
┌────────────────┴────────────────────────────┐
│         Application Service                  │
│  (Orquestra Discovery → Parser → Aggregator) │
└────────────────┬────────────────────────────┘
                 │
    ┌────────────┼────────────┐
    │            │            │
┌───▼───┐   ┌───▼────┐   ┌──▼────────┐
│Discovery│   │Parser  │   │Aggregator │
└────────┘   └────────┘   └───────────┘
```

### Componentes

- **Discovery**: Encontra arquivos de log no filesystem
- **Parser**: Extrai LogEntry de diferentes formatos
- **Aggregator**: Gera estatísticas, timeline e detecta anomalias  
- **Service**: Orquestra tudo em pipeline end-to-end
- **CLI**: Interface de linha de comando

## Exemplos

### Health Check Rápido

```bash
logdash analyze -path /var/log -last 1h -max-files 5
```

### Investigação Profunda

```bash
logdash analyze -path /app/logs -last 24h -verbose -top 20
```

### Análise Focada

```bash
logdash analyze -path /var/log -patterns '*.log' -depth 3
```

## Roadmap

- [x] Parser para logs genéricos
- [x] Parser para Python logging
- [x] Discovery recursivo com filtros
- [x] Agregação e estatísticas
- [x] Detecção de anomalias
- [x] CLI funcional
- [ ] Interface Web (htmx + Alpine.js)
- [ ] Interface TUI (bubbletea)
- [ ] Cache (Indexer) para performance
- [ ] Mais parsers (Log4j, syslog, nginx)
- [ ] Export para JSON/CSV
- [ ] Watch mode em tempo real

## Documentação

- [CLI Guide](./cmd/logdash/README.md) - Guia completo da interface CLI
- [Parser Guide](./docs/PARSER_GUIDE.md) - Como criar novos parsers
- [Discovery Guide](./docs/DISCOVERY_GUIDE.md) - Sistema de descoberta
- [Aggregator Guide](./docs/AGGREGATOR_GUIDE.md) - Análise e agregação
- [Architecture](./docs/ARCHITECTURE.md) - Visão geral técnica

## Contribuindo

Contribuições são bem-vindas! Veja [CONTRIBUTING.md](CONTRIBUTING.md) para detalhes.

### Desenvolvimento

```bash
# Rodar testes
go test ./...

# Rodar testes com cobertura
go test -cover ./...

# Build
go build -o logdash ./cmd/logdash
```

## Licença

MIT License - veja [LICENSE](LICENSE) para detalhes.

---
