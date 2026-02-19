# LogDash TUI (Terminal User Interface)

Interface interativa de terminal usando Bubbletea para navegação visual e análise de logs.

## Features

- ✨ **Interface Visual**: Bonita e intuitiva no terminal
- 🎨 **Cores e Estilos**: Usando Lipgloss para styling
- 📑 **Tabs Navegáveis**: Overview, Top Messages, Timeline, Anomalies
- ⚡ **Refresh em Tempo Real**: Pressione 'r' para reanalisar
- 🎯 **Foco no que Importa**: Visualização clara de erros e anomalias

## Uso

```bash
# Iniciar TUI
logdash tui

# Com path específico
logdash tui -path /var/log

# Filtrar últimas 24h
logdash tui -path /var/log -last 24h

# Com limites
logdash tui -path /var/log -max-files 10
```

## Navegação

### Controles

- **Tab / Shift+Tab**: Navegar entre abas
- **R**: Refresh (reanalisar)
- **Q / Ctrl+C**: Sair

### Abas Disponíveis

#### 1. Overview
Visão geral da análise:
- Summary textual
- Estatísticas de arquivos
- Distribuição por nível (Error, Warning, Info, etc)
- Cores indicando severidade

#### 2. Top Messages
Mensagens mais frequentes:
- Top 15 mensagens
- Contador de ocorrências
- Nível de severidade com cores
- Primeiro e último timestamp

#### 3. Timeline
Distribuição temporal:
- Gráfico de barras no tempo
- Até 20 pontos temporais
- Destaque para erros
- Escala visual proporcional

#### 4. Anomalies
Anomalias detectadas:
- Tipo de anomalia
- Descrição detalhada
- Severity score
- Timestamp da ocorrência
- Ícones indicando gravidade (🚨 crítico, ⚠️ médio, ℹ️ baixo)

## Interface

```
┌─────────────────────────────────────────────────────────────────┐
│ LogDash TUI                                                     │
│ Path: /var/log                                                  │
│                                                                 │
│ ╭─────────╮ ╭─────────────╮ ╭─────────╮ ╭───────────╮        │
│ │Overview │ │Top Messages │ │Timeline │ │ Anomalies │        │
│ ╰─────────╯ ╰─────────────╯ ╰─────────╯ ╰───────────╯        │
│                                                                 │
│ Summary                                                         │
│   Analyzed 1543 log entries, 47 errors, 23 warnings           │
│                                                                 │
│ Files                                                          │
│   Discovered: 15                                              │
│   Processed:  12                                              │
│   Entries:    1543                                            │
│                                                                 │
│ By Level                                                       │
│   ❌ Error:   47                                              │
│   ⚠️  Warning: 23                                              │
│   ℹ️  Info:    1473                                            │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│ tab/shift+tab: navigate • r: refresh • q: quit                │
│ Status: Analysis complete                                      │
└─────────────────────────────────────────────────────────────────┘
```

## Cores e Ícones

### Níveis de Log

- 💀 **Fatal** - Branco/Vermelho
- ❌ **Error** - Vermelho
- ⚠️  **Warning** - Amarelo/Laranja
- ℹ️  **Info** - Azul claro
- 🐛 **Debug** - Cinza
- 🔍 **Trace** - Cinza escuro

### Anomalias

- 🚨 **Severity ≥0.8** - Crítico (vermelho)
- ⚠️  **Severity ≥0.5** - Médio (amarelo)
- ℹ️  **Severity <0.5** - Baixo (azul)

## Opções

Mesmas opções do comando `analyze`:

- `-path` - Caminho raiz para buscar logs
- `-max-files` - Máximo de arquivos a processar
- `-depth` - Profundidade máxima de busca
- `-patterns` - Padrões de arquivo (comma-separated)
- `-last` - Período a analisar (24h, 7d, 30m)
- `-no-anomalies` - Desabilita detecção de anomalias

## Exemplos

### Health Check Interativo

```bash
logdash tui -path /var/log -last 1h -max-files 5
```

### Investigação Visual

```bash
logdash tui -path /app/logs -last 24h
```

### Análise Focada

```bash
logdash tui -path /var/log -patterns '*.log' -depth 3
```

## Performance

- **Análise Inicial**: Executada em background
- **Refresh**: Re-executa análise (pressione 'r')
- **Navegação**: Instantânea entre tabs
- **Memória**: Mantém resultado em memória para navegação rápida

## Troubleshooting

### Cores não aparecem

Verifique se seu terminal suporta cores ANSI (a maioria dos terminais modernos suporta).

### Interface quebrada

- Redimensione o terminal (mínimo recomendado: 80x24)
- Use um terminal moderno (iTerm2, Alacritty, Windows Terminal)

### Slow refresh

- Use `-no-anomalies` para análise mais rápida
- Limite arquivos com `-max-files`
- Reduza profundidade com `-depth`

## Dependências

- `github.com/charmbracelet/bubbletea` - Framework TUI (Elm Architecture)
- `github.com/charmbracelet/lipgloss` - Terminal styling

## Comparação: CLI vs TUI

### Use CLI (`analyze`) quando:
- Precisa de output para scripts
- Quer resultado rápido e direto
- Vai processar output programaticamente
- Está em ambiente sem interação (CI/CD)

### Use TUI (`tui`) quando:
- Quer navegar visualmente pelos dados
- Precisa investigar interativamente
- Quer refresh sem reexecutar comando
- Interface visual ajuda na análise

## Próximos Passos

- [ ] Scrolling para listas longas
- [ ] Busca/filtro em tempo real
- [ ] Exportar view atual
- [ ] Mais customização de cores
- [ ] Gráficos mais avançados

## Veja Também

- [CLI Documentation](../cmd/logdash/README.md)
- [Service Documentation](../../internal/service/README.md)
- [Bubbletea Examples](https://github.com/charmbracelet/bubbletea/tree/master/examples)