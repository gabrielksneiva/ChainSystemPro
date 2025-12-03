# Arquivos Removidos/Substituídos - ChainSystemPro Refactoring

**Data:** 2025-12-03  
**Razão:** Refatoração completa para implementar framework blockchain com TDD rigoroso, coverage 90%+, e implementação completa (sem TODOs/stubs).

## Arquivos Mantidos (Com Modificações)

- `go.mod` - Atualizado com novas dependências (btcd, btcutil, btcwallet, etc.)
- `Makefile` - Expandido com targets de test/coverage/lint
- `README.md` - Atualizado com nova documentação
- `.github/workflows/` - CI expandido para coverage 90%+

## Arquivos Removidos

### Implementações Incompletas
- `internal/adapters/bitcoin/adapter.go` - Reimplementado com TDD
- `internal/adapters/bitcoin/adapter_test.go` - Reimplementado com coverage completo
- `internal/adapters/bitcoin/harness/` - Novo harness determinístico
- `internal/adapters/evm/` - Será reimplementado na fase ETH

### Domain Layer
- `internal/domain/entities/entities.go` - Refatorado para suportar UTXO model (Bitcoin)
- `internal/domain/valueobjects/valueobjects.go` - Expandido com tipos Bitcoin-específicos
- `internal/domain/ports/ports.go` - Redesenhado para suportar múltiplos modelos (UTXO/Account)

### API Layer
- `internal/api/server.go` - Reimplementado com Fiber e OpenAPI completo
- `docs/swagger.json` - Substituído por OpenAPI 3.0 YAML

## Nova Estrutura

```
ChainSystemPro/
├── cmd/
│   └── server/                 # Entry point
├── pkg/                        # Bibliotecas públicas reutilizáveis
│   ├── crypto/                # Cryptografia (HD wallets, signing)
│   ├── encoding/              # Encoding utils (base58, bech32)
│   └── rpc/                   # RPC clients genéricos
├── internal/
│   ├── domain/
│   │   ├── bitcoin/           # Bitcoin-specific domain
│   │   ├── ethereum/          # Ethereum (fase 2)
│   │   ├── tron/              # Tron (fase 3)
│   │   └── solana/            # Solana (fase 4)
│   ├── application/           # Use cases
│   ├── infrastructure/        # Implementações concretas
│   └── interfaces/            # API REST (Fiber)
├── test/                      # Testes E2E e harnesses
└── docs/
    └── openapi.yaml           # OpenAPI 3.0 spec completa
```

## Justificativa

A estrutura anterior tinha:
- Implementações parciais com TODOs
- Coverage insuficiente (~60-70%)
- Falta de suporte adequado para UTXO model (Bitcoin)
- Abstrações genéricas demais que não representavam bem as diferenças entre chains

Nova abordagem:
- **Chain-first**: Implementar 100% para uma chain antes de generalizar
- **TDD rigoroso**: Testes antes de implementação
- **Coverage 90%+**: Garantido por CI
- **Sem abstrações prematuras**: Extrair padrões após implementar BTC e ETH
