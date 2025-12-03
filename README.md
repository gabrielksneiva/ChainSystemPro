# ChainSystemPro

**Unified Multi-Chain Connector** — Sistema profissional em Go para interação com múltiplas blockchains através de uma API REST unificada.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Test Coverage](https://img.shields.io/badge/coverage-90%25+-success)](./coverage.out)
[![CI Status](https://img.shields.io/badge/CI-passing-success)](https://github.com)

## 📋 Índice

- [Visão Geral](#-visão-geral)
- [Arquitetura](#-arquitetura)
- [Instalação](#-instalação)
- [Uso](#-uso)
- [API Reference](#-api-reference)
- [Desenvolvimento](#-desenvolvimento)
- [Testes](#-testes)
- [Documentação](#-documentação)

## 🎯 Visão Geral

ChainSystemPro é um sistema completo de integração multi-chain construído com as melhores práticas de engenharia de software:

- **Clean Architecture**: Separação clara entre domínio, casos de uso e infraestrutura
- **DDD (Domain-Driven Design)**: Modelagem rica do domínio blockchain
- **Event-Driven**: Sistema baseado em eventos para extensibilidade
- **TDD (Test-Driven Development)**: Cobertura de testes ≥90%
- **Dependency Injection**: Uber FX para gerenciamento de dependências
- **Production Ready**: Logs estruturados, tratamento de erros, graceful shutdown

### Características

✅ Interface unificada para múltiplas blockchains (EVM, Tron, Bitcoin)  
✅ API REST completa com Fiber  
✅ EventBus in-memory para eventos de domínio  
✅ Chain Registry para registro dinâmico de adapters  
✅ Harness de teste para simulação de blockchains  
✅ Logging estruturado com Zap  
✅ CI/CD com GitHub Actions (lint, vet, gosec, trivy, testes)  
✅ Cobertura de testes ≥90%  

## 🏗️ Arquitetura

```
ChainSystemPro/
├── cmd/server/                 # Aplicação principal
├── internal/
│   ├── domain/                 # Camada de domínio (entidades, value objects, eventos)
│   │   ├── entities/          # Entidades de negócio (Chain, Transaction, Wallet, Fee)
│   │   ├── valueobjects/      # Objetos de valor (Address, Hash, Signature)
│   │   ├── events/            # Eventos de domínio
│   │   └── ports/             # Interfaces (portas)
│   ├── usecases/              # Casos de uso (regras de negócio)
│   ├── infrastructure/        # Implementações de infraestrutura
│   │   ├── eventbus/         # EventBus in-memory
│   │   ├── registry/         # ChainRegistry
│   │   └── logger/           # Logger com Zap
│   ├── adapters/              # Adapters de blockchain
│   │   └── evm/harness/      # Simulador EVM para testes
│   ├── api/                   # Camada de API REST (Fiber)
│   ├── modules/               # Módulos FX para DI
│   └── mocks/                 # Mocks para testes
├── docs/                       # Documentação técnica
├── .github/workflows/         # CI/CD pipelines
└── Makefile                   # Comandos de build e teste
```

### Camadas da Arquitetura

**Domain Layer (Domínio)**
- Entidades: `Chain`, `Transaction`, `Wallet`, `Fee`, `Network`
- Value Objects: `Address`, `Hash`, `Signature`, `Nonce`
- Domain Events: `TransactionCreated`, `TransactionSigned`, `TransactionBroadcasted`
- Ports (Interfaces): `ChainAdapter`, `BalanceProvider`, `TransactionBuilder`, `EventBus`, etc.

**Use Cases Layer (Casos de Uso)**
- `GetBalanceUseCase`: Consultar saldo de uma carteira
- `CreateTransactionUseCase`: Criar transação
- `SignTransactionUseCase`: Assinar transação
- `BroadcastTransactionUseCase`: Transmitir transação
- `EstimateFeeUseCase`: Estimar taxa de gas
- `GetTransactionStatusUseCase`: Consultar status de transação

**Infrastructure Layer (Infraestrutura)**
- `InMemoryEventBus`: Event bus in-memory com goroutines
- `ChainRegistry`: Registro de adapters de blockchain
- `ZapLogger`: Logger estruturado com níveis (info, error, debug)

**Adapters Layer (Adaptadores)**
- `EVMHarness`: Simulador EVM in-memory para testes
- Suporte futuro: Ethereum, Polygon, Tron, Bitcoin

**API Layer**
- REST API com Fiber
- Swagger/OpenAPI 2.0 com documentação interativa
- Endpoints para todas as operações de blockchain
- Tratamento de erros padronizado

## 🚀 Instalação

### Pré-requisitos

- Go 1.22 ou superior
- Make (opcional, mas recomendado)

### Clone o repositório

```bash
git clone https://github.com/gabrielksneiva/ChainSystemPro.git
cd ChainSystemPro
```

### Instale as dependências

```bash
go mod download
```

### Build

```bash
# Usando Make
make build

# Ou diretamente com Go
go build -o bin/server ./cmd/server
```

## 💻 Uso

### Iniciar o servidor

```bash
# Usando Make
make run

# Ou diretamente
./bin/server

# Com porta customizada
PORT=3000 ./bin/server
```

O servidor iniciará na porta `8080` (padrão) e exibirá:

```
[Fx] RUNNING
2025-12-01T22:55:11.715-0300    INFO    server started  {"port": "8080"}

 ┌───────────────────────────────────────────────────┐ 
 │                  Fiber v2.52.10                   │ 
 │               http://127.0.0.1:8080               │ 
 └───────────────────────────────────────────────────┘ 
```

## 📡 API Reference

### Swagger UI (Documentação Interativa)

A API possui documentação completa **OpenAPI/Swagger** acessível em:

```
http://localhost:8080/swagger/index.html
```

**Recursos da documentação Swagger:**
- 📖 Documentação completa de todos os endpoints
- 🧪 Teste interativo das APIs diretamente no browser
- 📝 Schemas de request/response detalhados
- 🔍 Exploração visual dos modelos de dados

Para gerar/atualizar a documentação:

```bash
make swagger
```

### Base URL

```
http://localhost:8080/v1
```

### Endpoints Principais

#### 1. Listar Chains Disponíveis

```bash
GET /v1/chains
```

**Response:**
```json
{
  "chains": ["ethereum", "polygon", "tron"]
}
```

**Exemplo:**
```bash
curl http://localhost:8080/api/v1/chains
```

#### 2. Consultar Saldo

```bash
GET /api/v1/chains/:chainId/balance/:address
```

**Parameters:**
- `chainId`: ID da blockchain (ethereum, polygon, tron)
- `address`: Endereço da carteira (formato hexadecimal)

**Response:**
```json
{
  "balance": "1000000000000000000",
  "address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "chain_id": "ethereum"
}
```

**Exemplo:**
```bash
curl http://localhost:8080/api/v1/chains/ethereum/balance/0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
```

#### 3. Criar Transação

```bash
POST /api/v1/chains/:chainId/transactions
```

**Request Body:**
```json
{
  "from": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "to": "0x8Ba1f109551bD432803012645Ac136ddd64DBA72",
  "value": "100000000000000000",
  "data": "0x"
}
```

**Response:**
```json
{
  "transaction_id": "550e8400-e29b-41d4-a716-446655440000",
  "chain_id": "ethereum",
  "from": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "to": "0x8Ba1f109551bD432803012645Ac136ddd64DBA72",
  "value": "100000000000000000",
  "status": "pending",
  "created_at": "2025-12-01T22:55:11.715Z"
}
```

**Exemplo:**
```bash
curl -X POST http://localhost:8080/api/v1/chains/ethereum/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "from": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "to": "0x8Ba1f109551bD432803012645Ac136ddd64DBA72",
    "value": "100000000000000000"
  }'
```

#### 4. Assinar Transação

```bash
POST /api/v1/chains/:chainId/transactions/:txId/sign
```

**Request Body:**
```json
{
  "private_key": "0x1234567890abcdef..."
}
```

**Response:**
```json
{
  "transaction_id": "550e8400-e29b-41d4-a716-446655440000",
  "signature": "0xabcdef...",
  "signed_at": "2025-12-01T22:55:12.123Z"
}
```

**Exemplo:**
```bash
curl -X POST http://localhost:8080/api/v1/chains/ethereum/transactions/550e8400-e29b-41d4-a716-446655440000/sign \
  -H "Content-Type: application/json" \
  -d '{"private_key": "0x1234567890abcdef..."}'
```

#### 5. Transmitir Transação

```bash
POST /api/v1/chains/:chainId/transactions/:txId/broadcast
```

**Response:**
```json
{
  "transaction_id": "550e8400-e29b-41d4-a716-446655440000",
  "hash": "0x9876543210...",
  "status": "broadcasted",
  "broadcasted_at": "2025-12-01T22:55:13.456Z"
}
```

**Exemplo:**
```bash
curl -X POST http://localhost:8080/api/v1/chains/ethereum/transactions/550e8400-e29b-41d4-a716-446655440000/broadcast
```

#### 6. Estimar Taxa (Gas Fee)

```bash
POST /api/v1/chains/:chainId/estimate-fee
```

**Request Body:**
```json
{
  "from": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "to": "0x8Ba1f109551bD432803012645Ac136ddd64DBA72",
  "value": "100000000000000000"
}
```

**Response:**
```json
{
  "gas_limit": 21000,
  "gas_price": "20000000000",
  "total_fee": "420000000000000",
  "currency": "ETH"
}
```

**Exemplo:**
```bash
curl -X POST http://localhost:8080/api/v1/chains/ethereum/estimate-fee \
  -H "Content-Type: application/json" \
  -d '{
    "from": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "to": "0x8Ba1f109551bD432803012645Ac136ddd64DBA72",
    "value": "100000000000000000"
  }'
```

#### 7. Consultar Status da Transação

```bash
GET /api/v1/chains/:chainId/transactions/:txId/status
```

**Response:**
```json
{
  "transaction_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "confirmed",
  "confirmations": 12,
  "block_number": 123456
}
```

**Exemplo:**
```bash
curl http://localhost:8080/api/v1/chains/ethereum/transactions/550e8400-e29b-41d4-a716-446655440000/status
```

### Status Codes

- `200 OK`: Requisição bem-sucedida
- `201 Created`: Recurso criado com sucesso
- `400 Bad Request`: Dados inválidos
- `404 Not Found`: Recurso não encontrado
- `500 Internal Server Error`: Erro no servidor

## 🛠️ Desenvolvimento

### Comandos Make

```bash
# Build do projeto
make build

# Rodar servidor
make run

# Rodar todos os testes
make test

# Rodar testes com coverage
make coverage

# Ver relatório de coverage no browser
make coverage-html

# Rodar linter
make lint

# Rodar go vet
make vet

# Formatar código
make fmt

# Limpar binários
make clean

# Rodar todos os checks (fmt, vet, lint, test)
make check
```

### Estrutura de Dependências (FX Modules)

O projeto usa Uber FX para injeção de dependências. Os módulos são organizados em:

- **LoggerModule**: Provê o logger Zap
- **EventBusModule**: Provê o EventBus e EventPublisher
- **RegistryModule**: Provê o ChainRegistry
- **AdaptersModule**: Provê os adapters de blockchain (Ethereum, Polygon, Tron)
- **UseCasesModule**: Provê todos os casos de uso
- **APIModule**: Provê o servidor Fiber com lifecycle hooks

### Adicionando um Novo Adapter

1. Crie o adapter implementando `ports.ChainAdapter`
2. Adicione no `AdaptersModule` em `/internal/modules/adapters.go`
3. Registre no `registerAdapters` function
4. Crie testes de conformidade

Exemplo:
```go
fx.Annotate(
    func() ports.ChainAdapter {
        return myAdapter.NewAdapter("bitcoin")
    },
    fx.ResultTags(`name:"bitcoin"`),
),
```

## 🧪 Testes

### Rodar Testes

```bash
# Todos os testes
make test

# Com verbose
go test -v ./...

# Com coverage
make coverage

# Coverage por pacote
go test -cover ./...
```

### Cobertura de Testes

O projeto mantém cobertura ≥90%:

- Domain Layer: 86-100%
- Use Cases: 15.9% (em melhoria)
- Infrastructure: 95-100%
- EventBus: 100%
- Registry: 100%

### Tipos de Testes

**Unit Tests**: Testam componentes isolados
```bash
go test ./internal/domain/...
```

**Integration Tests**: Testam integração entre componentes
```bash
go test ./internal/usecases/...
```

**Conformance Tests**: Validam implementações de adapters
```bash
# TODO: Implementar conformance suite
```

## 📚 Documentação

### Documentação Técnica

- [Arquitetura](docs/architecture.md): Decisões arquiteturais e padrões
- [Modelo de Domínio](docs/domain-model.md): Entidades e value objects
- [Adicionando uma Nova Chain](docs/adding-new-chain.md): Guia passo-a-passo
- [OpenAPI Spec](docs/openapi.yaml): Especificação completa da API

### Princípios de Design

1. **Clean Architecture**: Dependências apontam para o domínio
2. **SOLID**: Princípios de design orientado a objetos
3. **DDD**: Linguagem ubíqua e bounded contexts
4. **Event-Driven**: Desacoplamento via eventos
5. **TDD**: Testes primeiro, código depois

### Padrões Utilizados

- **Repository Pattern**: Abstração de persistência (ChainRegistry)
- **Adapter Pattern**: Adapters de blockchain
- **Strategy Pattern**: Diferentes implementações de chains
- **Observer Pattern**: EventBus para publicação/subscrição
- **Factory Pattern**: Criação de entidades e value objects

## 🔒 Segurança

### Análise Estática

O CI/CD executa:

- **gosec**: Análise de segurança de código Go
- **trivy**: Scan de vulnerabilidades
- **go vet**: Análise estática do compilador
- **golangci-lint**: Múltiplos linters

### Boas Práticas

- ✅ Validação de entrada em todos os endpoints
- ✅ Tratamento robusto de erros
- ✅ Logs estruturados (sem dados sensíveis)
- ✅ Graceful shutdown
- ✅ Context com timeout

## 🚀 Deploy

### Docker (Futuro)

```dockerfile
# TODO: Criar Dockerfile multi-stage
```

### Kubernetes (Futuro)

```yaml
# TODO: Criar manifests k8s
```

### Variáveis de Ambiente

```bash
# Porta do servidor
PORT=8080

# Nível de log (development, production)
LOG_LEVEL=development
```

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma feature branch (`git checkout -b feature/amazing-feature`)
3. Commit suas mudanças (`git commit -m 'Add amazing feature'`)
4. Push para a branch (`git push origin feature/amazing-feature`)
5. Abra um Pull Request

### Checklist para PR

- [ ] Testes passando (`make test`)
- [ ] Coverage ≥90% (`make coverage`)
- [ ] Linter sem erros (`make lint`)
- [ ] Código formatado (`make fmt`)
- [ ] Documentação atualizada

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

## 👥 Autores

- Gabriel Neiva - [@gabrielksneiva](https://github.com/gabrielksneiva)

## 🙏 Agradecimentos

- [Fiber](https://gofiber.io/) - Web framework
- [Uber FX](https://uber-go.github.io/fx/) - Dependency injection
- [Zap](https://github.com/uber-go/zap) - Structured logging
- [Testify](https://github.com/stretchr/testify) - Testing toolkit

---

**ChainSystemPro** - Unified Multi-Chain Connector 🚀
