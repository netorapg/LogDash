# LogDash CLI

Interface de linha de comando para o LogDash - análise rápida e inteligente de logs.

## Instalação

### Build from Source

```bash
# Na raiz do projeto
go build -o logdash ./cmd/logdash

# Ou instalar globalmente
go install ./cmd/logdash

# Testar
./logdash version
```

### Build para Distribuição

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o logdash-linux ./cmd/logdash

# macOS
GOOS=darwin GOARCH=amd64 go build -o logdash-macos ./cmd/logdash

# Windows
GOOS=windows GOARCH=amd64 go build -o logdash.exe ./cmd/logdash
```

## Uso

### Análise Básica

```bash
# Analisar diretório atual
logdash analyze

# Analisar diretório específico
logdash analyze -path /var/log

# Analisar com output verbose
logdash analyze -path /var/log -verbose
```

### Filtros Temporais

```bash
# Últimas 24 horas
logdash analyze -path /var/log -last 24h

# Últimos 7 dias
logdash analyze -path /var/log -last 7d

# Últimos 30 minutos
logdash analyze -path /app/logs -last 30m
```

### Limites e Filtros

```bash
# Processar apenas 5 arquivos mais recentes
logdash analyze -path /var/log -max-files 5

# Buscar até 3 níveis de profundidade
logdash analyze -path /var/log -depth 3

# Apenas arquivos .log
logdash analyze -path /var/log -patterns '*.log'

# Múltiplos patterns
logdash analyze -path /var/log -patterns '*.log,*.txt,*.out'
```

### Customizar Output

```bash
# Mostrar top 20 mensagens
logdash analyze -path /var/log -top 20

# Desabilitar detecção de anomalias (mais rápido)
logdash analyze -path /var/log -no-anomalies

# Análise rápida com limites
logdash analyze -path /var/log -max-files 10 -no-anomalies -top 5
```

### Exemplos Práticos

```bash
# Health check rápido
logdash analyze -path /var/log -last 1h -max-files 5

# Investigação profunda
logdash analyze -path /app/logs -last 24h -verbose -top 20

# Análise focada em erros recentes
logdash analyze -path /var/log -last 6h -patterns '*.log'

# Troubleshooting específico
logdash analyze -path /home/user/project/logs -depth 5 -verbose
```

## Comandos

### analyze

Analisa arquivos de log e gera relatório.

**Opções:**

- `-path` (string) - Caminho raiz para buscar logs (padrão: diretório atual)
- `-max-files` (int) - Máximo de arquivos a processar (padrão: sem limite)
- `-depth` (int) - Profundidade máxima de busca (padrão: 10)
- `-patterns` (string) - Padrões de arquivo separados por vírgula (padrão: '*.log,*.txt')
- `-last` (string) - Período a analisar (ex: '24h', '7d', '30m')
- `-top` (int) - Número de top mensagens (padrão: 10)
- `-no-anomalies` - Desabilita detecção de anomalias
- `-verbose` - Output detalhado

### version

Mostra versão do LogDash.

```bash
logdash version
```

### help

Mostra ajuda.

```bash
logdash help
```

## Output

O CLI produz um relatório formatado com:

### 📝 Summary
Resumo geral da análise

### 📁 Files
- Arquivos descobertos
- Arquivos processados
- Total de entries

### 📈 By Level
Distribuição por severidade:
- 💀 Fatal
- ❌ Error
- ⚠️  Warning
- ℹ️  Info
- 🐛 Debug
- 🔍 Trace

### 🔝 Top Messages
Mensagens mais frequentes com:
- Ícone de severidade
- Nível
- Mensagem (truncada)
- Contador de ocorrências

### ⚠️  Anomalies
Anomalias detectadas:
- 🚨 Severity alta (≥0.8)
- ⚠️  Severity média (≥0.5)
- ℹ️  Severity baixa (<0.5)

### 📄 File Details (verbose)
Por arquivo:
- ✓/✗ Status
- Nome do arquivo
- Entries encontradas
- Tempo de processamento
- Parser usado

### ❌ Errors
Erros encontrados durante processamento

## Exemplo de Output

```
🔍 Analyzing logs in: /var/log

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 ANALYSIS RESULTS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 Summary:
   Analyzed 1543 log entries, 47 errors, 23 warnings, spanning 8h15m.

📁 Files:
   Discovered: 15
   Processed:  12
   Entries:    1543

📈 By Level:
   ❌ Error:    47
   ⚠️  Warning:  23
   ℹ️  Info:     1473

🔝 Top Messages:
   1. ❌ [ERROR] Connection timeout (x15)
   2. ⚠️  [WARNING] Slow query detected (x8)
   3. ℹ️  [INFO] Request processed successfully (x1420)

⚠️  Anomalies Detected:
   🚨 [error_burst] 10 errors in 10m0s
   ⚠️  [repeated_error] Error message repeated 15 times: Connectio...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## Performance

### Benchmarks Típicos

| Arquivos | Entries | Tempo |
|----------|---------|-------|
| 10 | ~10K | < 1s |
| 50 | ~50K | < 3s |
| 100 | ~100K | < 10s |

### Otimizações

Para análises mais rápidas:

```bash
# Limitar arquivos
logdash analyze -max-files 10

# Desabilitar anomalias
logdash analyze -no-anomalies

# Reduzir top messages
logdash analyze -top 5

# Combinar tudo
logdash analyze -max-files 10 -no-anomalies -top 5
```

## Integração

### Scripts

```bash
#!/bin/bash
# health-check.sh

OUTPUT=$(logdash analyze -path /var/log -last 1h -max-files 5)

if echo "$OUTPUT" | grep -q "CRITICAL"; then
    echo "🔴 System unhealthy"
    exit 1
elif echo "$OUTPUT" | grep -q "WARNING"; then
    echo "🟡 System degraded"
    exit 0
else
    echo "🟢 System healthy"
    exit 0
fi
```

### Cron Jobs

```bash
# Relatório diário às 00:00
0 0 * * * /usr/local/bin/logdash analyze -path /var/log -last 24h > /tmp/daily-report.txt

# Health check a cada hora
0 * * * * /usr/local/bin/logdash analyze -path /var/log -last 1h -max-files 5 -no-anomalies
```

### Pipeline CI/CD

```yaml
# .github/workflows/log-analysis.yml
name: Log Analysis

on:
  schedule:
    - cron: '0 0 * * *'

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Analyze logs
        run: |
          ./logdash analyze -path ./logs -verbose
```

## Troubleshooting

### "No log files found"

```bash
# Verificar path
ls -la /var/log

# Tentar outros patterns
logdash analyze -path /var/log -patterns '*'

# Aumentar depth
logdash analyze -path /var/log -depth 15
```

### "Permission denied"

```bash
# Rodar com sudo
sudo logdash analyze -path /var/log

# Ou mudar permissões
chmod +r /var/log/*.log
```

### Performance lenta

```bash
# Limitar processamento
logdash analyze -max-files 20 -no-anomalies

# Reduzir depth
logdash analyze -depth 3
```

## Próximos Passos

- [ ] Adicionar formato de output JSON (`-format json`)
- [ ] Export para arquivo (`-output report.txt`)
- [ ] Filtros por nível (`-levels error,warning`)
- [ ] Watch mode (`-watch`)
- [ ] Configuração via arquivo (`.logdash.yml`)

## Veja Também

- [Service Documentation](../../internal/service/README.md)
- [Parser Guide](../../docs/PARSER_GUIDE.md)
- [Aggregator Guide](../../docs/AGGREGATOR_GUIDE.md)