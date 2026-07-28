# Castilho Health OS — Roadmap

Repo: https://github.com/AugustoCastilhoDev/castilho_health_os
Última atualização: 2026-07-28 (tela de Financeiro implementada, incl. CRUD de regras de repasse)

## Stack

Backend: Go (Fiber) + GORM/Postgres + Redis (provisionado, ainda não usado) + Docker Compose.
Frontend: React + TypeScript + Vite + Tailwind CSS v4 + react-router-dom (`frontend/`).
Nginx + Certbot prontos para a VPS, ainda não deployados (deploy fica por último, por decisão do usuário).

## O que já existe

### Domínio e persistência
- **Modelos** (`internal/domain/models`): `Tenant`, `User`, `Patient`, `Appointment` (+ `AppointmentStatusLog`), `FinancialRule`, `FinancialTransaction`.
- **Modelos novos, ainda sem migration/repositório/service/handler** (ver seção "PEP / Documentos / Receitas" abaixo): `MedicalRecord`, `PatientDocument`, `DocumentTemplate`, `MemedPrescriptionLog`.
- **Migrations** (`migrations/`, formato golang-migrate): 6 pares up/down cobrindo o schema atual, com `CHECK` constraints espelhando os enums do Go.
- **Repositórios** (`internal/repository`): um por agregado, todos com `tenant_id` explícito em toda query (nunca confia só no campo da struct). `AppointmentRepository.TransitionStatus` roda em transação com `SELECT FOR UPDATE`, valida a máquina de estados e grava o audit log atomicamente.

### Auth e API
- **JWT + bcrypt + RBAC** (`internal/auth`, `internal/api/middleware`): claims carregam `tenant_id`/`role`; `RequireAuth` + `RequireRole` protegem as rotas.
- **Services** (`internal/service`): Auth (login com erro genérico contra enumeração), Tenant (onboarding atômico clínica+admin + `Get` para o `/api/tenant`), User (CRUD + troca/reset de senha), Patient, Appointment, Financial, Settlement.
- **HTTP API** (`internal/api`, Fiber): `/auth/register`, `/auth/login`, `/api/tenant`, `/api/me`, `/api/users/*`, `/api/patients/*`, `/api/appointments/*` (+ transições de status, + `/settle` manual), `/api/financial-rules/*`, `/api/financial-transactions/*`.
- **CORS** (`internal/config` + `cmd/api/main.go`): `CORS_ALLOWED_ORIGINS` (default `http://localhost:5173`) — necessário pro frontend chamar a API cross-origin em dev.
- **Settlement financeiro automático** (`internal/service/settlement_service.go`): gera o `PROFESSIONAL_PAYOUT` sozinho assim que as duas condições de gate são satisfeitas — `Appointment.Status == COMPLETED` e existe um `FinancialTransaction` do tipo `PATIENT_PAYMENT` com `Status == PAID` para aquele agendamento. Disparado best-effort a partir de dois pontos: `AppointmentHandler.Transition` (quando a transição resulta em `COMPLETED`) e `FinancialHandler.MarkPaid`. Idempotente via `FindPayoutBySource`; se nenhuma regra financeira aplicável existir ainda, fica pendente e pode ser reprocessado manualmente via `POST /api/appointments/:id/settle` (role `finance`).

### Frontend (`frontend/`)
- **Design system**: paleta "Clean Tech" (slate/sky/emerald/rose) aliada como tokens semânticos (`bg-brand-*`/`text-brand-*`) em `src/index.css` (`@theme` do Tailwind v4).
- **Auth real** (`src/lib/auth/AuthContext.tsx`): login contra `/auth/login`, JWT decodificado no cliente só para exibição (nunca confiado — a API sempre revalida), sessão persistida em `localStorage`, `ProtectedRoute` guardando o app.
- **Telas**: Login, Dashboard (contadores do dia), Agenda Semanal (grade CSS Grid, cores por status, criação de agendamento e transições de status via modal), Pacientes (busca + cadastro/edição completos), **Financeiro** (Lançamentos + Regras de Repasse, ver abaixo). Todas com dados reais da API, não mockados.
- **Lacunas conhecidas da integração**: não existe endpoint "todos os agendamentos de hoje da clínica" (só por profissional) — resolvido no frontend com um seletor de profissional (`useProfessionalScope`) quando o usuário logado não é `DOCTOR`/`DENTIST`. Não existe endpoint de faturamento agregado — o Dashboard soma as transações de cada agendamento individualmente (aceitável na escala atual).
- **Financeiro** (`src/pages/FinancialPage.tsx`): duas abas.
  - **Lançamentos**: lista paginada de `FinancialTransaction` (recebimentos e repasses) com filtro por situação/tipo, botão "Registrar Pagamento" (cria `PATIENT_PAYMENT` via busca de paciente + valor + forma de pagamento + convênio opcional) e "Marcar como pago" por linha (só aparece para quem já é `TENANT_ADMIN`/`FINANCE` — o backend também restringe via `finance` middleware). Dois cards de resumo (A Receber / A Repassar pendentes) somam até 100 linhas pendentes de cada tipo — mesmo tipo de limitação aceita do Dashboard, não há endpoint de soma agregada.
  - **Regras de Repasse**: seletor de profissional (reaproveita `useProfessionalScope`) + tabela de `FinancialRule` com CRUD completo pela tela — "Nova Regra" e "Editar" (`RuleFormModal.tsx`: tipo percentual/fixo, escopo procedimento/convênio, dedução de taxa, prioridade) e "Ativar/Desativar" por linha (visível só para `TENANT_ADMIN`/`FINANCE`, backend restringe via `finance` middleware).
  - Backend ganhou `GET /api/financial-transactions` (novo — antes só existia por agendamento): `FinancialTransactionRepository.ListByTenant` com filtro por `Type`/`Status` + paginação (`FinancialService.ListTransactions` default 20/máx 100 por página), coberto por teste de repositório.
  - Backend ganhou `PUT /api/financial-rules/:id` (novo — o service já tinha `UpdateRule`, mas sem handler/rota): validação (`validateRule`, compartilhada com `CreateRule`) e semântica PUT de substituição completa — `is_active` omitido é zerado, então o frontend sempre envia a regra inteira.
- **Estoque**: ainda não tem tela — aparece na sidebar marcado "em breve".

### Infra
- **Docker**: `Dockerfile` multi-stage (build Go 1.25 → runtime Alpine, roda como `nobody`). `docker-compose.yml` com `db`/`redis`/`migrate`/`app` (dev local) e `nginx`/`certbot` atrás de um profile `production` (não sobem em dev).
- **CI** (`.github/workflows/ci.yml`): GitHub Actions rodando gofmt/vet/build/test (com Postgres de serviço + migrations aplicadas) e um build do Dockerfile, a cada push/PR na `main`. **Verde.**
- **Testes**: 54 automatizados (backend) — 41 de serviço (mocks, sem banco) + 13 de repositório (integração, banco real). Fluxos de frontend (login, agenda, CRUD de paciente, settlement) validados manualmente ponta a ponta com Playwright headless contra a API real a cada rodada, mas não há suíte automatizada de frontend ainda.

## PEP / Documentos / Receitas — arquitetura definida, implementação pendente

Pedido em 2026-07-27: gestão de Documentos, Receitas e Evolução Clínica. Três decisões tecnológicas oficiais do projeto:

1. **Arquivos externos (exames, laudos de terceiros)**: Cloudflare R2 (S3-compatible), acesso via **presigned URLs** geradas no Go (`internal/storage/r2.go`, usando `aws-sdk-go-v2`) — o servidor nunca vê os bytes do arquivo, só autoriza upload/download por um tempo curto. `PatientDocument` guarda só metadados (`FileKey`, nome, tamanho, mimetype).
2. **Receitas digitais/controladas**: SDK da Memed integrado no **frontend** (o widget deles fala direto com a Memed); o Go só registra um log de auditoria (`MemedPrescriptionLog`) com o ID externo da receita — nunca é o sistema de registro do conteúdo prescrito.
3. **Prontuário/evolução + PDFs locais** (atestados, laudos, declarações): editor de rich-text **interno** no frontend salvando em `MedicalRecord.Content` (JSON/HTML opaco para o backend), e geração de PDF local em Go (`gofpdf` ou `maroto`) a partir de `DocumentTemplate` (texto com tags `{{...}}`) quando for atestado/declaração/laudo avulso.

**O que já foi feito** (código real no repo, `internal/domain/models/`):
- `medical_record.go` — `MedicalRecord` (Patient + Professional + Tenant, `AppointmentID` opcional, `Type` enum MEDICA/ODONTOLOGICA/PSICOLOGICA/PSIQUIATRICA, `Content text`, `IsLocked` + `LockedAt`/`LockedByID` para auditoria de imutabilidade).
- `patient_document.go` — `PatientDocument` (Patient + Tenant + `UploadedByID`, `FileKey`/`FileName`/`FileSize`/`MimeType`).
- `document_template.go` — `DocumentTemplate` (tenant-level, sem Patient/User, `Type` ATESTADO/DECLARACAO/LAUDO/OUTRO, `Content` com tags).
- `memed_prescription_log.go` — `MemedPrescriptionLog` (Patient + Professional + Tenant, `MemedPrescriptionID` único, `Status`, `IssuedAt`).
- `internal/storage/r2.go` — `R2Client` com `PresignUpload`/`PresignDownload` (aws-sdk-go-v2, endpoint `https://{account_id}.r2.cloudflarestorage.com`, path-style, região `"auto"`). Dependências já em `go.mod`/`go.sum`, build e `docker build --no-cache` verificados.

**O que falta, de propósito ainda não feito** (para não deixar código pela metade antes de decidir prioridade):
- Migrations para as 4 tabelas novas.
- Repositórios/services/handlers/rotas (CRUD de prontuário com trava de `IsLocked` no service, upload/download de documento, CRUD de templates, registro do log da Memed).
- Conta Cloudflare R2 de verdade (bucket + token de API) — usuário vai pedir passo a passo de configuração quando for a hora.
- Parceria/credenciais Memed (integração comercial, não só técnica — precisa de cadastro na Memed).
- Editor de rich-text no frontend (nenhuma biblioteca escolhida ainda) e a tela de geração de PDF.

## Decisões de arquitetura para lembrar

- **Multi-tenant**: `tenant_id` sempre vem do JWT (claims), nunca do payload do cliente.
- **Dinheiro**: sempre `int64` em centavos, nunca `float64`.
- **PUT = substituição completa**: campos omitidos são zerados — exceto em `User`, onde `PasswordHash` é protegido explicitamente (load-then-mutate no service) porque não faz parte do DTO de update.
- **RBAC em duas camadas**: a maioria das rotas usa `RequireRole` na própria rota; a restrição de quem pode criar um `PROFESSIONAL_PAYOUT` mora no service, porque depende do campo `type` do corpo da requisição, que o middleware de rota não enxerga.
- **Agendamento**: transição de status só acontece via `TransitionStatus`, que é a única operação que grava em `appointments` e `appointment_status_logs` — nunca two-step.
- **Arquivos nunca passam pelo nosso servidor**: R2 é acessado só via presigned URL, gerada sob demanda, com expiração curta — vale tanto para upload quanto para download.
- **Backend nunca é dono do conteúdo da receita médica**: a Memed é o sistema de registro da prescrição; o Go só audita (quem, quando, qual ID).

## Dívida técnica consciente (decisões, não esquecimentos)

- Sem hard-delete de usuário — só desativação (`is_active=false`), pra não quebrar histórico de agendamento/repasse.
- Sem proteção contra o último `TENANT_ADMIN` se autodesativar/rebaixar (ficaria um tenant sem admin).
- Quem pode disparar qual transição específica de agendamento (ex: só o profissional atribuído inicia `IN_PROGRESS`) não é restrito por role — só a máquina de estados protege contra pulos ilegais.
- Sem testes de handler HTTP nem suíte automatizada de frontend (ambas as camadas validadas manualmente/via Playwright a cada rodada).
- Redis provisionado no compose mas nada usa ainda (sem sessão/cache).
- Quem pode disparar o `/settle` manual é `finance`/`admin`, mas não há alerta/fila visível para "agendamentos completos sem repasse pendente".
- Settlement paga sempre 100% do valor calculado de uma vez (`PENDING`); não há parcelamento nem estorno automático se o `PATIENT_PAYMENT` for cancelado depois que o payout já foi gerado.
- Dashboard/Agenda fazem N+1 de lookups (nome de paciente/profissional, transações por agendamento) — aceitável na escala atual, primeiro candidato a otimizar se a base crescer.

## Próximos passos (recomendação de ordem — decidir na volta)

Avaliado em 2026-07-27, concluído em 2026-07-28: a prioridade era **Financeiro (tela) antes de PEP/Documentos completo** — feito, incluindo o CRUD de regras que tinha ficado pendente. Próxima decisão em aberto entre PEP/Documentos completo (maior frente, com dependências externas) e Odontograma (módulo novo do zero, sem bloqueio externo).

1. ~~Settlement financeiro automático~~ — feito.
2. ~~Frontend (Dashboard, Agenda, Pacientes)~~ — feito.
3. ~~Tela de Financeiro~~ — feito (2026-07-28): lançamentos (listar/filtrar/paginar, registrar pagamento, marcar como pago) + regras de repasse (CRUD completo: criar, editar, ativar/desativar).
4. **PEP completo** — migrations + repos/services/handlers para os 4 modelos novos, upload/download via R2, editor rich-text, geração de PDF local, integração Memed. Arquitetura já definida (ver seção acima); maior de todas as frentes.
5. **Odontograma interativo** — módulo exclusivo odonto.
6. **Integração WhatsApp** — confirmação automática 24h antes, processando resposta do paciente.
7. **Estoque** — nenhum modelo ainda; começar do zero.
8. **Deploy na VPS Hostinger** — Nginx+Certbot já configurados (`nginx/conf.d/`), falta executar de fato quando o domínio estiver apontado.
9. **Uso real do Redis** — sessões e/ou cache de agenda.

## Como retomar

```bash
cd "Castilho Health OS"
docker compose up -d db redis        # sobe Postgres + Redis
docker compose run --rm migrate      # aplica migrations pendentes, se houver
go run ./cmd/api                     # roda a API localmente (porta 8080)

# Frontend:
cd frontend && npm install && npm run dev   # http://localhost:5173

# Testes do backend:
set -a; source .env; set +a          # exporta env vars pro go test achar o Postgres
go test ./...
```
