# LogDash Web UI

Interface web moderna para análise de logs usando htmx + Alpine.js + Tailwind CSS.

## Features

- ✨ **Dashboard Visual**: Cards com estatísticas principais
- 📊 **Gráfico Interativo**: Timeline com Chart.js
- 🎨 **Design Moderno**: Tailwind CSS dark mode
- ⚡ **Performance**: htmx para atualizações rápidas
- 🔄 **Refresh**: Botão para reanalisar logs
- 📑 **5 Tabs**: Overview, Errors, Warnings, Timeline, Anomalies

## Uso

```bash
# Iniciar servidor
logdash web

# Com path específico
logdash web -path /var/log

# Porta customizada
logdash web -path /var/log -port 3000
```

Depois abra http://localhost:8080 no navegador!

## Tecnologias

- **htmx**: Interatividade sem JavaScript pesado
- **Alpine.js**: Reatividade leve (como Vue mas menor)
- **Tailwind CSS**: Styling utility-first
- **Chart.js**: Gráficos bonitos
- **Go net/http**: Servidor HTTP nativo

## Screenshots

### Dashboard
- Cards com métricas principais
- Total entries, errors, warnings, anomalies

### Tabs
1. **Overview**: Summary + estatísticas gerais
2. **Errors**: Lista de todos os erros com contador
3. **Warnings**: Lista de todos os warnings com contador
4. **Timeline**: Gráfico temporal interativo
5. **Anomalies**: Anomalias detectadas com severity

## API Endpoints

```
GET  /                   - Página principal
POST /api/refresh        - Reanalisar logs
GET  /api/stats          - Estatísticas gerais
GET  /api/errors         - Lista de erros
GET  /api/warnings       - Lista de warnings
GET  /api/timeline       - Dados da timeline
GET  /api/anomalies      - Anomalias detectadas
```

## Desenvolvimento

O servidor carrega templates de `internal/web/templates/`.

Para modificar o frontend:
1. Editar `internal/web/templates/index.html`
2. Rebuild: `go build -o logdash ./cmd/logdash`
3. Executar: `./logdash web`
4. Refresh no navegador

## Próximas Features

- [ ] Busca/filtro em tempo real
- [ ] Export para PDF/CSV
- [ ] Dark/Light mode toggle
- [ ] Auto-refresh (polling)
- [ ] WebSocket para updates em tempo real
- [ ] Filtros avançados (por nível, período, source)
- [ ] Detalhes expandíveis de cada mensagem

## Veja Também

- [CLI Documentation](../cmd/logdash/README.md)
- [TUI Documentation](../tui/README.md)
- [Service Documentation](../service/README.md)