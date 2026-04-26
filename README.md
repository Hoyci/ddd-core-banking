# DDD Core Banking

Este repositório é um projeto onde estou praticando os conceitos de **Domain-Driven Design (DDD)** aplicados a um sistema bancário em Go pensado para ser executado em produção.

O código aqui não é um produto. Mas cada decisão de design foi tomada para tornar o projeto escalável para produção utilizando os conceitos de DDD.

---

## O que é Domain-Driven Design?

DDD é uma abordagem para desenvolvimento de software que coloca o **domínio do negócio** no centro das decisões técnicas. Em vez de modelar o sistema em torno do banco de dados ou da API, modelamos em torno das regras e da linguagem do negócio.

Os principais pilares são:

- **Linguagem Ubíqua**: desenvolvedores e especialistas do negócio usam o mesmo vocabulário
- **Bounded Context**: fronteiras explícitas que isolam diferentes partes do domínio

---

## Estrutura do Projeto

```
cmd/
  onboarding/main.go     # HTTP :8080 + outbox worker + RabbitMQ publisher
  account/main.go        # Consumer RabbitMQ — cria contas e atualiza saldos
  payment/main.go        # HTTP :8081 + outbox worker + RabbitMQ publisher
  notification/main.go   # Consumer RabbitMQ — envia e-mails de rejeição

internal/
  onboarding/            # Bounded context: cadastro e aprovação de clientes
    domain/
      entity/            # Aggregate root: Client
      event/             # ClientCreated, ClientApproved, ClientRejected
      repository/        # Interfaces ClientRepository, OutboxRepository
      errors.go          # Erros sentinela do domínio
    application/usecases/  # CreateClient, ApproveClient, RejectClient
    infrastructure/
      outbox/            # Worker LISTEN/NOTIFY
      persistence/postgres/
    interfaces/http/     # Handlers HTTP

  account/               # Bounded context: contas bancárias e saldos
    domain/
      entity/            # Aggregate root: Account
      errors/
      repository/        # AccountRepository
      service/           # AccountNumberGenerator
    application/usecases/  # CreateAccount, DebitAccount, TransferBalance
    infrastructure/
      persistence/postgres/
    interfaces/eventhandler/  # ClientApproved, InvoicePaymentProcessed, TransferProcessed

  payment/               # Bounded context: pagamentos
    domain/
      entity/            # InvoicePayment, Transfer (com Entry)
      event/             # InvoicePaymentProcessed, TransferProcessed
      repository/        # PaymentRepository, OutboxRepository
      errors/
    application/usecases/  # PayInvoice, TransferFunds
    infrastructure/
      corebanking/       # Interface Client + StubClient
      outbox/            # Worker LISTEN/NOTIFY
      persistence/postgres/
    interfaces/http/     # Handlers HTTP

  notification/          # Bounded context: notificações
    application/usecases/  # NotifyRejection
    infrastructure/email/  # Interface Sender + StubSender
    interfaces/eventhandler/  # ClientRejected

pkg/
  events/        # Interface DomainEvent (Shared Kernel)
  messaging/     # RabbitMQ: Publisher (topic exchange) e Consumer (DLQ)
  outbox/        # Struct OutboxRecord
  http/          # Struct ApiResponse compartilhada
  valueobjects/  # Document (CPF/CNPJ), Email, Address, AccountNumber, ID
```

---

## Fluxo Completo

```
┌─────────────────────────────────────────────────────────────────────┐
│                         ONBOARDING :8080                            │
│                                                                     │
│  POST /clients → Client{PENDING} + ClientCreated event              │
│  PATCH /clients/{id}/approve → Client{APPROVED} + ClientApproved    │
│  PATCH /clients/{id}/reject  → Client{REJECTED} + ClientRejected    │
│                    │                       │                        │
│              [Outbox Worker]         [Outbox Worker]                │
└──────────────────────────────────────────────────────────────────── ┘
                      │                       │
                      ▼                       ▼
             [RabbitMQ: core-banking exchange]
                      │                       │
          ┌───────────┘           ┌───────────┘
          ▼                       ▼
┌──────────────────┐    ┌──────────────────────┐
│  ACCOUNT worker  │    │ NOTIFICATION worker   │
│                  │    │                       │
│ ClientApproved   │    │ ClientRejected        │
│  → CreateAccount │    │  → envia e-mail       │
│                  │    │    "tente em 90 dias" │
│ InvoicePayment   │    └──────────────────────┘
│  Processed       │
│  → DebitAccount  │
│                  │
│ TransferProcessed│
│ → TransferBalance│
└──────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                          PAYMENT :8081                              │
│                                                                     │
│  POST /payments/invoice                                             │
│    1. Verifica saldo e bloqueio da conta (leitura local)            │
│    2. Chama Core Banking (síncrono) ← fonte da verdade             │
│    3. Persiste InvoicePayment + Entry{DEBIT} + outbox (1 tx)        │
│    4. Retorna comprovante imediatamente                             │
│                                                                     │
│  POST /payments/transfers                                           │
│    1. Verifica saldo/bloqueio do remetente e destinatário           │
│    2. Chama Core Banking (síncrono)                                 │
│    3. Persiste Transfer + 2× Entry{DEBIT/CREDIT} + outbox (1 tx)   │
│    4. Retorna comprovante imediatamente                             │
│                    │                                                │
│              [Outbox Worker]                                        │
└─────────────────────────────────────────────────────────────────────┘
                      │
                      ▼
             [RabbitMQ: core-banking exchange]
                      │
                      ▼
┌──────────────────┐
│  ACCOUNT worker  │
│                  │
│ → atualiza saldo │  (eventual consistency — saldo no app é
│   de forma async │   read model; Core Banking é autoritativo)
└──────────────────┘
```

---

## Decisões de Arquitetura

### Core Banking síncrono + saldo eventual

Quando o usuário faz um pagamento ou transferência, a resposta (comprovante) é **síncrona**: o módulo de Payment chama o Core Banking via HTTP e aguarda a confirmação antes de responder ao cliente. Isso garante que o comprovante só é emitido após confirmação real.

A atualização do saldo no banco de dados local é **assíncrona via eventos** — o módulo Account consome os eventos de Payment e atualiza o campo `balance`. Essa eventual consistência é aceitável porque:

- O Core Banking é a fonte da verdade para o saldo real
- Novas transações sempre passam pelo Core Banking (que valida o saldo real)
- O saldo exibido no app pode ter um delay mínimo, mas nunca causa prejuízo

Se o Core Banking estiver indisponível, o módulo Payment retorna `503 Service Unavailable` — nenhuma transação é aceita sem confirmação, preservando a integridade financeira.

### Outbox Pattern

Cada módulo que publica eventos (`onboarding`, `payment`) usa o padrão Outbox: os eventos são persistidos na mesma transação do banco de dados que salva o estado do negócio. Um worker separado lê a tabela de outbox e publica no RabbitMQ.

Isso elimina o risco de perda de eventos em caso de falha entre o commit do banco e a publicação no broker.

```
[Use Case] → [Repository.Save()]
                    │
                    ▼ (mesma transação)
          ┌───────────────────────┐
          │  INSERT business data │
          │  INSERT outbox rows   │
          │  pg_notify(channel)   │
          └───────────────────────┘
                    │ (LISTEN/NOTIFY)
                    ▼
          [Outbox Worker] → [RabbitMQ] → [Consumers]
```

### Tabela `entries` genérica

Todos os lançamentos financeiros — independente do tipo de transação — são registrados na tabela `entries` com `source_type` e `source_id`:

| source_type       | entries geradas                          |
|-------------------|------------------------------------------|
| `INVOICE_PAYMENT` | 1× DEBIT no pagador                      |
| `TRANSFER`        | 1× DEBIT no remetente + 1× CREDIT no destinatário |

Adicionar PIX, TED, ou qualquer outro tipo futuro não requer alteração na tabela.

### Dead Letter Queue

O Consumer RabbitMQ configura automaticamente uma DLQ para cada fila. Mensagens que falham no processamento são movidas para a DLQ em vez de serem recolocadas na fila principal, evitando loops infinitos de retry.

---

## Conceitos DDD no Código

### Bounded Context

Um **Bounded Context** é uma fronteira explícita dentro da qual um modelo de domínio específico se aplica. Este projeto possui quatro bounded contexts:

- **Onboarding**: cadastro e aprovação de clientes
- **Core Banking**: criação e gestão de contas bancárias e saldos
- **Payment**: pagamentos de invoice e transferências entre contas
- **Notification**: envio de e-mails transacionais

Cada contexto tem seus próprios aggregates, erros de domínio e repositórios. A comunicação entre eles acontece via **eventos de domínio publicados no RabbitMQ**, nunca por chamadas diretas.

### Aggregate Root

Um **Aggregate** é um cluster de objetos de domínio tratados como uma unidade. O **Aggregate Root** é o único ponto de entrada para modificar o estado desse cluster.

**`Client`** (Onboarding):

```go
func (c *Client) ApproveClient() error {
    if c.status != OnboardingStatusPending {
        return domain.ErrClientNotPending  // regra de negócio aplicada aqui
    }
    c.status = OnboardingStatusApproved
    c.events = append(c.events, event.ClientApproved{ /* ... */ })
    return nil
}
```

**`InvoicePayment`** e **`Transfer`** (Payment): validam os dados na construção e são imutáveis após criados.

### Value Objects

**Value Objects** representam conceitos do domínio definidos pelos seus **valores**, não por identidade. São imutáveis e carregam validação dentro de si.

- `Document`: valida CPF e CNPJ com algoritmo de dígito verificador
- `Email`: valida formato e normaliza para minúsculas
- `Address`: valida CEP, UF com 2 letras, campos obrigatórios
- `AccountNumber`: dígito verificador calculado via algoritmo de Luhn

### Erros Sentinela

Todos os erros de domínio são variáveis declaradas em um único lugar, verificáveis com `errors.Is()`:

```go
// Handler HTTP mapeia erros para status codes sem conhecer detalhes internos
case errors.Is(err, domain.ErrEmailAlreadyInUse):
    w.WriteHeader(http.StatusConflict)           // 409
case errors.Is(err, payerrors.ErrInsufficientFunds):
    w.WriteHeader(http.StatusUnprocessableEntity) // 422
case errors.Is(err, payerrors.ErrCoreBankingUnavailable):
    w.WriteHeader(http.StatusServiceUnavailable)  // 503
```

### Repository Pattern

O domínio define interfaces de repositório — ele sabe **o que** precisa, não **como** é armazenado. A implementação concreta fica na infraestrutura e é invisível para o domínio.

---

## Como Rodar

**Pré-requisitos:** Go 1.22+, PostgreSQL e RabbitMQ.

### Com Tilt (recomendado — hot reload)

```bash
tilt up
```

O Tiltfile sobe a infra via docker-compose e reinicia os serviços automaticamente ao salvar arquivos.

### Manualmente

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/corebanking?sslmode=disable"
export RABBITMQ_URL="amqp://guest:guest@localhost:5672/"

# Infra
docker-compose up -d

# Migrations
psql $DATABASE_URL -f migrations/001_initial.sql
psql $DATABASE_URL -f migrations/002_create_accounts.sql
psql $DATABASE_URL -f migrations/003_create_payment_tables.sql

# Serviços (terminais separados)
go run ./cmd/onboarding/...
go run ./cmd/account/...
go run ./cmd/payment/...
go run ./cmd/notification/...
```

**Comandos úteis:**

```bash
go build ./cmd/... ./internal/... ./pkg/...
go test ./...
go test -race ./...
go fmt ./...
go vet ./...
```

**Endpoints:**

| Serviço      | Método  | Rota                              | Descrição                        |
|--------------|---------|-----------------------------------|----------------------------------|
| Onboarding   | `POST`  | `/clients`                        | Cria cliente (status: PENDING)   |
| Onboarding   | `PATCH` | `/clients/{id}/approve`           | Aprova cliente → cria conta      |
| Onboarding   | `PATCH` | `/clients/{id}/reject`            | Rejeita cliente → envia e-mail   |
| Payment      | `POST`  | `/payments/invoice`               | Paga invoice (boleto)           |
| Payment      | `POST`  | `/payments/transfers`             | Transferência entre contas       |

---

## Referências

- [Domain-Driven Design — Eric Evans](https://www.domainlanguage.com/ddd/)
- [Implementing Domain-Driven Design — Vaughn Vernon](https://vaughnvernon.com/?page_id=168)
- [DDD, Hexagonal, Onion, Clean, CQRS — How I put it all together](https://herbertograca.com/2017/11/16/explicit-architecture-01-ddd-hexagonal-onion-clean-cqrs-how-i-put-it-all-together/)
