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
- 💻 Oferece interface web e TUI

## Quick Start
```bash
# Download
curl -L https://github.com/seu-usuario/logdashboard/releases/latest/download/logdashboard -o logdashboard
chmod +x logdashboard

# Execute na raiz do seu projeto
./logdash

# Acesse http://localhost:8080
# Ou pressione 'T' no terminal para TUI
```

## Features

- ✅ Parsing automático de formatos comuns (Log4j, syslog, JSON, etc)
- ✅ Foco inteligente nas últimas 24h por padrão
- ✅ Filtros por severidade, período, texto
- ✅ Detecção de padrões e anomalias
- ✅ Zero configuração necessária
- ✅ Cache inteligente para análises subsequentes

## Roadmap

- [ ] Suporte a mais formatos de log
- [ ] Exportação de relatórios
- [ ] Sistema de plugins para parsers customizados
- [ ] Monitoramento em tempo real (fase 2)

## Contribuindo

Contribuições são bem-vindas! Veja [CONTRIBUTING.md](CONTRIBUTING.md) para detalhes.

## Licença

MIT License - veja [LICENSE](LICENSE) para detalhes.