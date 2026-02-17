# Guia de Contribuição

Obrigado pelo interesse em contribuir com o LogDash! Este documento fornece diretrizes para contribuir com o projeto de forma eficaz.

## Código de Conduta

Este projeto adota um código de conduta que esperamos que todos os participantes sigam. Seja respeitoso, inclusivo e construtivo em todas as interações.

## Como Contribuir

### Reportando Bugs

Antes de criar um issue:
- Verifique se o bug já foi reportado nas [issues existentes](https://github.com/seu-usuario/logdashboard/issues)
- Se não encontrar nada similar, crie uma nova issue incluindo:
  - Descrição clara do problema
  - Passos para reproduzir
  - Comportamento esperado vs comportamento atual
  - Versão do LogDashboard, sistema operacional, e Go version
  - Logs relevantes (se aplicável)

### Sugerindo Melhorias

Adoramos receber sugestões! Para propor uma nova feature:
- Abra uma issue descrevendo a funcionalidade
- Explique o problema que ela resolve
- Descreva como você imagina que funcionaria
- Aguarde feedback antes de começar a implementar

### Submetendo Pull Requests

1. **Fork o repositório** e crie sua branch a partir da `main`:
```bash
   git checkout -b feature/minha-contribuicao
```

2. **Faça suas alterações** seguindo os padrões do projeto (veja abaixo)

3. **Adicione testes** para qualquer código novo ou alterado

4. **Certifique-se de que tudo passa**:
```bash
   go test ./...
   go vet ./...
   gofmt -s -w .
```

5. **Commit suas mudanças** com mensagens claras:
```bash
   git commit -m "feat: adiciona suporte para parsing de logs Nginx"
```

6. **Push para sua fork** e abra um Pull Request

7. **Descreva suas mudanças** no PR:
   - O que foi alterado e por quê
   - Relate a issue correspondente (se houver)
   - Adicione screenshots se for mudança visual

## Padrões de Código

### Estilo Go

Seguimos as convenções padrão da comunidade Go:
- Use `gofmt` para formatação automática
- Use `go vet` para detectar problemas comuns
- Siga as [Effective Go guidelines](https://golang.org/doc/effective_go.html)
- Nomes de pacotes devem ser curtos, claros e em lowercase
- Exports apenas o que é necessário (capitalize apenas o público)

### Estrutura de Código
```go
// Bom: interfaces definem contratos
type Parser interface {
    Parse(content []byte) ([]LogEntry, error)
}

// Bom: implementações concretas são privadas
type log4jParser struct {
    // ...
}

// Bom: funções construtoras retornam interfaces
func NewLog4jParser() Parser {
    return &log4jParser{}
}
```

### Comentários

- Documente todas as funções e tipos exportados
- Use comentários para explicar "por quê", não "o quê"
- Evite comentários óbvios
```go
// Bom
// ParseWithTimeout aborts parsing if it exceeds the specified duration
// to prevent blocking on malformed files.
func ParseWithTimeout(content []byte, timeout time.Duration) ([]LogEntry, error)

// Ruim
// Parse parses the content
func Parse(content []byte) ([]LogEntry, error)
```

### Testes

- Todo código novo precisa de testes
- Use table-driven tests quando possível
- Teste casos de sucesso E falha
- Nomes de testes devem descrever o cenário: `TestParseLog4j_InvalidTimestamp`
```go
func TestParser_Parse(t *testing.T) {
    tests := []struct {
        name    string
        input   []byte
        want    []LogEntry
        wantErr bool
    }{
        {
            name:  "valid log entry",
            input: []byte("[2024-01-15 10:30:00] ERROR - Failed"),
            want:  []LogEntry{{Level: "ERROR", Message: "Failed"}},
            wantErr: false,
        },
        // mais casos...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // teste...
        })
    }
}
```

### Commits

Usamos [Conventional Commits](https://www.conventionalcommits.org/) para mensagens de commit:

- `feat:` nova funcionalidade
- `fix:` correção de bug
- `docs:` mudanças em documentação
- `test:` adição ou correção de testes
- `refactor:` refatoração que não altera comportamento
- `perf:` melhorias de performance
- `chore:` atualizações de build, dependências, etc

Exemplos:
```
feat: adiciona parser para logs JSON estruturados
fix: corrige detecção de timezone em logs syslog
docs: atualiza exemplos de uso no README
test: adiciona testes para aggregator com múltiplos arquivos
```

## Arquitetura do Projeto

Antes de contribuir com código, leia o [ARCHITECTURE.md](ARCHITECTURE.md) para entender:
- Estrutura de camadas do projeto
- Como as interfaces se comunicam
- Padrões de design utilizados
- Decisões arquiteturais importantes

## Fluxo de Desenvolvimento

1. Discuta mudanças grandes em uma issue primeiro
2. Mantenha PRs focados — uma feature/fix por PR
3. Responda a feedback em code reviews construtivamente
4. Seja paciente — revisões podem levar alguns dias

## Precisa de Ajuda?

- Abra uma issue com a label `question`
- Veja issues com label `good first issue` para começar
- Leia a documentação em `/docs`

## Licença

Ao contribuir, você concorda que suas contribuições serão licenciadas sob a mesma licença do projeto (MIT).

---

Obrigado por fazer o LogDashboard melhor! 🚀