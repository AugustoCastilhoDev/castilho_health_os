# Castilho Health OS — Roadmap

Repo: https://github.com/AugustoCastilhoDev/castilho_health_os
Última atualização: 2026-07-27 (settlement automático)

## Stack

Go (Fiber) + GORM/Postgres + Redis (provisionado, ainda não usado) + Docker Compose.
Nginx + Certbot prontos para a VPS, ainda não deployados (deploy fica por último, por decisão do usuário).

## O que já existe

### Domínio e persistência
- **Modelos** (`internal/domain/models`): `Tenant`, `User`, `Patient`, `Appointment` (+ `AppointmentStatusLog`), `FinancialRule`, `FinancialTransaction`.
- **Migrations** (`migrations/`, formato golang-migrate): 5 pares up/down cobrindo todo o schema, com `CHECK` constraints espelhando os enums do Go.
- **Repositórios** (`internal/repository`): um por agregado, todos com `tenant_id` explícito em toda query (nunca confia só no campo da struct). `AppointmentRepository.TransitionStatus` roda em transação com `SELECT FOR UPDATE`, valida a máquina de estados e grava o audit log atomicamente.

### Auth e API
- **JWT + bcrypt + RBAC** (`internal/auth`, `internal/api/middleware`): claims carregam `tenant_id`/`role`; `RequireAuth` + `RequireRole` protegem as rotas.
- **Services** (`internal/service`): Auth (login com erro genérico contra enumeração), Tenant (onboarding atômico clínica+admin), User (CRUD + troca/reset de senha), Patient, Appointment, Financial.
- **HTTP API** (`internal/api`, Fiber): `/auth/register`, `/auth/login`, `/api/me`, `/api/users/*`, `/api/patients/*`, `/api/appointments/*` (+ transições de status, + `/settle` manual), `/api/financial-rules/*`, `/api/financial-transactions/*`.
- **Settlement financeiro automático** (`internal/service/settlement_service.go`): gera o `PROFESSIONAL_PAYOUT` sozinho assim que as duas condições de gate são satisfeitas — `Appointment.Status == COMPLETED` e existe um `FinancialTransaction` do tipo `PATIENT_PAYMENT` com `Status == PAID` para aquele agendamento. Disparado best-effort (nunca falha a resposta HTTP da ação que o originou) a partir de dois pontos: `AppointmentHandler.Transition` (quando a transição resulta em `COMPLETED`) e `FinancialHandler.MarkPaid`. Idempotente via `FindPayoutBySource`; se nenhuma regra financeira aplicável existir ainda, fica pendente e pode ser reprocessado manualmente via `POST /api/appointments/:id/settle` (role `finance`). `FinancialTransaction` ganhou `ProcedureCode`/`InsurancePlan` (migration `000006`) para que `FinancialRule.FindApplicable` resolva regras não-wildcard automaticamente.

### Infra
- **Docker**: `Dockerfile` multi-stage (build Go 1.25 → runtime Alpine, roda como `nobody`). `docker-compose.yml` com `db`/`redis`/`migrate`/`app` (dev local) e `nginx`/`certbot` atrás de um profile `production` (não sobem em dev).
- **CI** (`.github/workflows/ci.yml`): GitHub Actions rodando gofmt/vet/build/test (com Postgres de serviço + migrations aplicadas) e um build do Dockerfile, a cada push/PR na `main`. **Verde.**
- **Testes**: 54 automatizados — 41 de serviço (mocks, sem banco, inclui 7 de `SettlementService`) + 13 de repositório (integração, banco real). `go test ./internal/service/...` roda sem setup; os de repositório precisam de `docker compose up -d db` + env vars exportadas (`.env` não é lido automaticamente pelo `go test` por causa do working directory por pacote). O fluxo completo do settlement (registro → agendamento → pagamento → transições → mark-paid → repasse gerado sozinho → retry idempotente) foi validado manualmente ponta a ponta via smoke script contra a API rodando.

## Decisões de arquitetura para lembrar

- **Multi-tenant**: `tenant_id` sempre vem do JWT (claims), nunca do payload do cliente.
- **Dinheiro**: sempre `int64` em centavos, nunca `float64`.
- **PUT = substituição completa**: campos omitidos são zerados — exceto em `User`, onde `PasswordHash` é protegido explicitamente (load-then-mutate no service) porque não faz parte do DTO de update.
- **RBAC em duas camadas**: a maioria das rotas usa `RequireRole` na própria rota; a restrição de quem pode criar um `PROFESSIONAL_PAYOUT` mora no service, porque depende do campo `type` do corpo da requisição, que o middleware de rota não enxerga.
- **Agendamento**: transição de status só acontece via `TransitionStatus`, que é a única operação que grava em `appointments` e `appointment_status_logs` — nunca two-step.

## Dívida técnica consciente (decisões, não esquecimentos)

- Sem hard-delete de usuário — só desativação (`is_active=false`), pra não quebrar histórico de agendamento/repasse.
- Sem proteção contra o último `TENANT_ADMIN` se autodesativar/rebaixar (ficaria um tenant sem admin).
- Quem pode disparar qual transição específica de agendamento (ex: só o profissional atribuído inicia `IN_PROGRESS`) não é restrito por role — só a máquina de estados protege contra pulos ilegais.
- Sem testes de handler HTTP (camada mais fina; validada manualmente via curl em várias rodadas).
- Redis provisionado no compose mas nada usa ainda (sem sessão/cache).
- Quem pode disparar o `/settle` manual é `finance`/`admin`, mas não há alerta/fila visível para "agendamentos completos sem repasse pendente" — hoje é preciso saber que falta configurar uma regra e chamar o endpoint por conta própria.
- Settlement paga sempre 100% do valor calculado de uma vez (`PENDING`); não há parcelamento nem estorno automático se o `PATIENT_PAYMENT` for cancelado depois que o payout já foi gerado.

## Próximos passos (sem ordem fixada — decidir na volta)

1. ~~Settlement financeiro automático~~ — feito (ver acima).
2. **Prontuário Eletrônico (PEP)** — próximo módulo de negócio central, auditável.
3. **Odontograma interativo** — módulo exclusivo odonto.
4. **Integração WhatsApp** — confirmação automática 24h antes, processando resposta do paciente.
5. **Prescrição digital + TISS/TUSS** — módulo exclusivo médico (Memed/Nexodata).
6. **Frontend** — ainda não existe nada; hoje é só API.
7. **Deploy na VPS Hostinger** — Nginx+Certbot já configurados (`nginx/conf.d/`), falta executar de fato quando o domínio estiver apontado.
8. **Uso real do Redis** — sessões e/ou cache de agenda.

## Como retomar

```bash
cd "Castilho Health OS"
docker compose up -d db redis        # sobe Postgres + Redis
docker compose run --rm migrate      # aplica migrations pendentes, se houver
go run ./cmd/api                     # roda a API localmente (porta 8080)

# Testes:
set -a; source .env; set +a          # exporta env vars pro go test achar o Postgres
go test ./...
```
