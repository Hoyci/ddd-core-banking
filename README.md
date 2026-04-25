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
cmd/onboarding/main.go               # Ponto de entrada — liga as dependências e sobe os serviços
internal/onboarding/
  domain/                            # Núcleo — zero dependências externas
    entity/                          # Aggregate root: Client
    event/                           # Eventos de domínio
    repository/                      # Interfaces de repositório
    errors.go                        # Erros sentinela do domínio
  application/usecases/              # Casos de uso: CreateClient, ApproveClient, RejectClient
  infrastructure/
    outbox/                          # Worker do padrão Outbox
    persistence/postgres/            # Implementações PostgreSQL
  interfaces/http/                   # Handlers HTTP
pkg/
  events/                            # Interface DomainEvent (Shared Kernel)
  outbox/                            # Struct OutboxRecord
  valueobjects/                      # Document (CPF/CNPJ), Email, Address, ID
```

---

## Conceitos DDD no Código

### 1. Bounded Context

Um **Bounded Context** é uma fronteira explícita dentro da qual um modelo de domínio específico se aplica. Dentro dessa fronteira, os termos têm significado preciso e consistente.

Neste projeto, o bounded context é **Onboarding**. Todo o código que pertence a esse contexto vive em `internal/onboarding/`. Se no futuro o sistema crescer (ex: `payments`, `loans`), cada um teria seu próprio bounded context com seus próprios modelos — mesmo que compartilhem conceitos com nomes parecidos.

---

### 2. Aggregate Root

Um **Aggregate** é um cluster de objetos de domínio tratados como uma unidade. O **Aggregate Root** é o único ponto de entrada para modificar o estado desse cluster — nada de fora pode modificar os objetos internos diretamente.

Neste projeto, `Client` é o aggregate root do contexto de Onboarding:

```go
// internal/onboarding/domain/entity/client.go

type Client struct {
    id        string
    document  valueobjects.Document
    fullName  string
    email     valueobjects.Email
    phone     string
    address   valueobjects.Address
    status    OnboardingStatus
    createdAt time.Time

    events []events.DomainEvent  // eventos acumulados, não visíveis de fora
}
```

Todos os campos são **privados**. O único jeito de criar um `Client` é pela função `CreateClient()`, que valida as regras de negócio:

```go
func CreateClient(input CreateClientInput) (*Client, error) {
    if input.FullName == "" {
        return nil, domain.ErrFullNameRequired
    }

    doc, err := valueobjects.NewDocument(input.Document)
    if err != nil {
        return nil, fmt.Errorf("%w: %w", domain.ErrInvalidDocument, err)
    }

    // ... mais validações ...

    client := &Client{
        id:     valueobjects.GenerateID(),
        status: OnboardingStatusPending,
        // ...
    }

    return client, nil
}
```

As transições de estado (`ApproveClient`, `RejectClient`) também passam pelo aggregate, garantindo que as regras de negócio sempre sejam respeitadas:

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

**Por que isso importa?** Sem aggregate root, qualquer parte do sistema poderia mudar `status` diretamente, bypassando as regras. Com ele, é impossível aprovar um cliente que não está `PENDING` — a regra está codificada, não documentada.

---

### 3. Value Objects

**Value Objects** representam conceitos do domínio que são definidos pelos seus **valores**, não por identidade. Dois value objects com os mesmos valores são idênticos. Eles são **imutáveis** e carregam validação dentro de si.

Neste projeto, `Document` encapsula CPF e CNPJ com validação real dos dígitos verificadores:

```go
// pkg/valueobjects/document.go

type Document struct {
    number   string
    category DocumentCategory  // "CPF" ou "CNPJ"
}

func NewDocument(value string) (Document, error) {
    digits := onlyDigits.ReplaceAllString(value, "")

    switch len(digits) {
    case 11:
        if err := validateCPF(digits); err != nil {
            return Document{}, err
        }
        return Document{number: digits, category: CPF}, nil
    case 14:
        if err := validateCNPJ(digits); err != nil {
            return Document{}, err
        }
        return Document{number: digits, category: CNPJ}, nil
    default:
        return Document{}, domain.ErrInvalidDocument
    }
}

func (d Document) Number() string             { return d.number }
func (d Document) Category() DocumentCategory { return d.category }
```

Outros value objects neste projeto: `Email` (valida formato e normaliza para minúsculas) e `Address` (valida CEP, UF com 2 letras, campos obrigatórios).

**Por que isso importa?** Sem value objects, um `document` seria uma `string` qualquer — e a validação estaria espalhada pelo código, duplicada, ou ausente. Com value objects, um `Document` que chegou ao aggregate root é **garantidamente válido**.

---

### 4. Eventos de Domínio

**Domain Events** registram algo significativo que aconteceu no domínio. Eles são fatos imutáveis expressados na linguagem do negócio.

Neste projeto, cada transição de estado do `Client` emite um evento:

```go
// pkg/events/events.go

type DomainEvent interface {
    EventName() string
    OccurredAt() time.Time
}
```

```go
// internal/onboarding/domain/event/client-created.go

type ClientCreated struct {
    ClientID  string
    Document  string
    FullName  string
    Email     string
    CreatedAt time.Time
}

func (e ClientCreated) EventName() string     { return "Onboarding.ClientCreated" }
func (e ClientCreated) OccurredAt() time.Time { return e.CreatedAt }
```

Os eventos são **acumulados no aggregate** durante a operação e extraídos depois pelo repositório via `PullEvents()`:

```go
func (c *Client) PullEvents() []events.DomainEvent {
    evts := c.events
    c.events = nil
    return evts
}
```

**Por que isso importa?** Eventos de domínio desacoplam o que aconteceu do que deve ser feito em reação. Outros sistemas (notificações, auditoria, relatórios) podem reagir aos eventos sem que o domínio saiba que eles existem.

---

### 5. Repository Pattern

O **Repository** abstrai o acesso a dados. O domínio define uma **interface** — ele sabe o que precisa, mas não sabe como é armazenado. A implementação concreta fica na camada de infraestrutura.

A interface vive no domínio:

```go
// internal/onboarding/domain/repository/client_repository.go

type ClientRepository interface {
    Save(client *entity.Client, events []events.DomainEvent) error
    FindByID(id string) (*entity.Client, error)
    FindByEmail(email string) (*entity.Client, error)
}
```

A implementação PostgreSQL fica na infraestrutura e é invisível para o domínio. O `Save` persiste o cliente **e** os eventos de domínio na mesma transação:

```go
// internal/onboarding/infrastructure/persistence/postgres/client_repository.go

func (r *ClientRepository) Save(client *entity.Client, domainEvents []events.DomainEvent) error {
    tx, err := r.db.Begin(context.Background())
    // ... upsert do cliente ...
    // ... insert dos eventos no outbox ...
    _, err = tx.Exec(ctx, `SELECT pg_notify('outbox_event', '')`)
    return tx.Commit(context.Background())
}
```

**Por que isso importa?** Os casos de uso dependem da interface, não da implementação. Trocar PostgreSQL por outro banco de dados — ou usar um repositório em memória nos testes — não requer mudança nenhuma no domínio.

---

### 6. Use Cases (Camada de Aplicação)

**Use Cases** orquestram o fluxo de uma operação do sistema. Eles não contêm regras de negócio — delegam para o domínio. O padrão é sempre: **carregar → aplicar lógica → salvar**.

```go
// internal/onboarding/application/usecases/create_client.go

type CreateClientUseCase struct {
    repo repository.ClientRepository
}

func (uc *CreateClientUseCase) Execute(input entity.CreateClientInput) error {
    // 1. Verificar pré-condições (regra da aplicação: email único)
    existing, err := uc.repo.FindByEmail(input.Email)
    if err != nil && !errors.Is(err, domain.ErrNotFound) {
        return fmt.Errorf("finding client by email: %w", err)
    }
    if existing != nil {
        return domain.ErrEmailAlreadyInUse
    }

    // 2. Criar o aggregate (regras de negócio ficam dentro de CreateClient)
    client, err := entity.CreateClient(input)
    if err != nil {
        return err
    }

    // 3. Salvar com os eventos
    return uc.repo.Save(client, client.PullEvents())
}
```

O caso de uso de aprovação segue o mesmo padrão — carregar, aplicar, salvar:

```go
// internal/onboarding/application/usecases/approve_client.go

func (uc *ApproveClientUseCase) Execute(input ApproveClientInput) error {
    client, err := uc.repo.FindByID(input.ClientID)  // carregar
    if err != nil {
        return fmt.Errorf("finding client by id: %w", err)
    }

    if err := client.ApproveClient(); err != nil {    // aplicar lógica de domínio
        return fmt.Errorf("approving client: %w", err)
    }

    return uc.repo.Save(client, client.PullEvents())  // salvar com evento
}
```

**Por que isso importa?** Separar casos de uso do domínio torna cada camada testável de forma isolada. O domínio pode ser testado sem banco de dados; os casos de uso podem ser testados com repositório em memória; os handlers HTTP podem ser testados com use cases mockados.

---

### 7. Erros Sentinela

Todos os erros de domínio são variáveis declaradas em um único lugar, verificáveis com `errors.Is()`. Isso preserva a identidade do erro mesmo quando ele é encapsulado com `fmt.Errorf("%w", ...)`.

```go
// internal/onboarding/domain/errors.go

var (
    ErrNotFound                = errors.New("not found")
    ErrEmailAlreadyInUse       = errors.New("email already in use")
    ErrInvalidDocument         = errors.New("invalid document")
    ErrClientNotPending        = errors.New("client is not in pending status")
    ErrRejectionReasonRequired = errors.New("rejection reason is required")
    // ...
)
```

O handler HTTP usa esses erros para mapear para status codes corretos:

```go
if err := h.createClient.Execute(input); err != nil {
    switch {
    case errors.Is(err, domain.ErrEmailAlreadyInUse):
        w.WriteHeader(http.StatusConflict)           // 409
    case errors.Is(err, domain.ErrInvalidDocument),
         errors.Is(err, domain.ErrFullNameRequired):
        w.WriteHeader(http.StatusUnprocessableEntity) // 422
    case errors.Is(err, domain.ErrNotFound):
        w.WriteHeader(http.StatusNotFound)            // 404
    default:
        w.WriteHeader(http.StatusInternalServerError) // 500
    }
}
```

**Por que isso importa?** Comparar strings de erro é frágil. Erros sentinela com `errors.Is()` permitem que a camada de apresentação tome decisões precisas sem conhecer os detalhes internos de cada camada.

---

### 8. Outbox Pattern

O **Outbox Pattern** resolve o problema de consistência entre salvar dados e publicar eventos. Se o sistema salvasse no banco e depois publicasse no broker de mensagens, uma falha entre as duas operações causaria inconsistência.

A solução: salvar o cliente e os eventos **na mesma transação do banco de dados**. Um worker separado lê a tabela `outbox` e publica os eventos de forma assíncrona.

```
[HTTP Handler]
     |
     v
[Use Case] → [Repository.Save()]
                    |
                    v
          ┌─────────────────────┐
          │   TRANSAÇÃO ÚNICA   │
          │  INSERT client      │
          │  INSERT outbox rows │
          │  pg_notify(...)     │
          └─────────────────────┘
                    |
                    v (LISTEN/NOTIFY)
          [Outbox Worker]
                    |
                    v
          [MessagePublisher] → Kafka / RabbitMQ / etc.
```

O worker bloqueia aguardando notificações do PostgreSQL, sem polling:

```go
// internal/onboarding/infrastructure/outbox/worker.go

func (w *Worker) Start(ctx context.Context) error {
    w.conn.Exec(ctx, "LISTEN outbox_event")

    for {
        _, err := w.conn.WaitForNotification(ctx)  // bloqueia até pg_notify
        if err != nil {
            if ctx.Err() != nil {
                return nil  // shutdown gracioso
            }
            return err
        }

        w.processAll()  // publica e marca como processado
    }
}
```

**Por que isso importa?** Sem o outbox pattern, um crash depois do `INSERT` mas antes do `Publish` perderia o evento para sempre. Com ele, o pior caso é o evento ser processado duas vezes (at-least-once delivery) — o que é muito mais fácil de tratar do que perda de dados.

---

## Fluxo do Onboarding

```
POST /clients
  └─ CreateClient() → status: PENDING, emite ClientCreated
          |
          ├─ PATCH /clients/{id}/approve
          │    └─ ApproveClient() → status: APPROVED, emite ClientApproved
          │
          └─ PATCH /clients/{id}/reject
               └─ RejectClient(reason) → status: REJECTED, emite ClientRejected
```

Apenas clientes com status `PENDING` podem ser aprovados ou rejeitados. Qualquer tentativa com outro status retorna `ErrClientNotPending`.

---

## Como Rodar

**Pré-requisitos:** Go 1.22+ e PostgreSQL

```bash
# Variável de ambiente necessária
export DATABASE_URL="postgres://user:password@localhost:5432/corebanking?sslmode=disable"

# Aplicar migrations
psql $DATABASE_URL -f migrations/001_initial.sql

# Rodar a aplicação
go run ./cmd/onboarding/...
```

**Comandos úteis:**

```bash
# Build
go build ./cmd/onboarding/... ./internal/... ./pkg/...

# Testes
go test ./...

# Testes com race detector
go test -race ./...

# Formatar e verificar
go fmt ./...
go vet ./...
```

**Endpoints disponíveis:**

```
POST   /clients                        Cria um novo cliente (status: PENDING)
PATCH  /clients/{clientID}/approve     Aprova um cliente pendente
PATCH  /clients/{clientID}/reject      Rejeita um cliente pendente
```

---

## Referências

- [Domain-Driven Design — Eric Evans](https://www.domainlanguage.com/ddd/)
- [Implementing Domain-Driven Design — Vaughn Vernon](https://vaughnvernon.com/?page_id=168)
- [DDD, Hexagonal, Onion, Clean, CQRS — How I put it all together](https://herbertograca.com/2017/11/16/explicit-architecture-01-ddd-hexagonal-onion-clean-cqrs-how-i-put-it-all-together/)
