# Castilho Health OS — Roadmap

Repo: https://github.com/AugustoCastilhoDev/castilho_health_os
Última atualização: 2026-07-28 (PEP: upload de documentos via Cloudflare R2 real, conta já configurada)

## Stack

Backend: Go (Fiber) + GORM/Postgres + Redis (provisionado, ainda não usado) + Docker Compose.
Frontend: React + TypeScript + Vite + Tailwind CSS v4 + react-router-dom (`frontend/`).
Nginx + Certbot prontos para a VPS, ainda não deployados (deploy fica por último, por decisão do usuário).

## O que já existe

### Domínio e persistência
- **Modelos** (`internal/domain/models`): `Tenant`, `User`, `Patient`, `Appointment` (+ `AppointmentStatusLog`), `FinancialRule`, `FinancialTransaction`, `MedicalRecord`, `DocumentTemplate`, `PatientDocument`.
- **Modelo que ainda existe só como struct Go, sem migration/repositório/service/handler** (bloqueado por parceria externa — ver seção "PEP / Documentos / Receitas" abaixo): `MemedPrescriptionLog` (precisa de parceria comercial com a Memed).
- **Migrations** (`migrations/`, formato golang-migrate): 8 pares up/down cobrindo o schema atual, com `CHECK` constraints espelhando os enums do Go.
- **Repositórios** (`internal/repository`): um por agregado, todos com `tenant_id` explícito em toda query (nunca confia só no campo da struct). `AppointmentRepository.TransitionStatus` roda em transação com `SELECT FOR UPDATE`, valida a máquina de estados e grava o audit log atomicamente.

### Auth e API
- **JWT + bcrypt + RBAC** (`internal/auth`, `internal/api/middleware`): claims carregam `tenant_id`/`role`; `RequireAuth` + `RequireRole` protegem as rotas.
- **Services** (`internal/service`): Auth (login com erro genérico contra enumeração), Tenant (onboarding atômico clínica+admin + `Get` para o `/api/tenant`), User (CRUD + troca/reset de senha), Patient, Appointment, Financial, Settlement, MedicalRecord (CRUD com trava de `IsLocked`), DocumentTemplate (CRUD + `Generate`), PatientDocument (presigned upload/download via R2, ver abaixo).
- **HTTP API** (`internal/api`, Fiber): `/auth/register`, `/auth/login`, `/api/tenant`, `/api/me`, `/api/users/*`, `/api/patients/*`, `/api/appointments/*` (+ transições de status, + `/settle` manual), `/api/financial-rules/*`, `/api/financial-transactions/*`, `/api/medical-records/*` (+ `/:id/lock`), `/api/document-templates/*` (+ `/:id/generate`, retorna PDF binário), `/api/patients/:patientID/documents/*` (upload em duas etapas) + `/api/patient-documents/:id/*` (download/exclusão).
- **CORS** (`internal/config` + `cmd/api/main.go`): `CORS_ALLOWED_ORIGINS` (default `http://localhost:5173`) — necessário pro frontend chamar a API cross-origin em dev. Também expõe `Content-Disposition` (`ExposeHeaders`) para o frontend conseguir ler o nome do arquivo gerado.
- **Cloudflare R2** (`internal/config` + `cmd/api/main.go`): `R2_ACCOUNT_ID`/`R2_ACCESS_KEY_ID`/`R2_SECRET_ACCESS_KEY`/`R2_BUCKET_NAME` — opcionais (`Config.IsR2Configured()`); se ausentes, o app sobe normalmente e só o upload/download de documento fica indisponível (`503`, `service.ErrStorageNotConfigured`). Configurados de verdade em 2026-07-28 (conta e bucket reais do usuário).
- **Settlement financeiro automático** (`internal/service/settlement_service.go`): gera o `PROFESSIONAL_PAYOUT` sozinho assim que as duas condições de gate são satisfeitas — `Appointment.Status == COMPLETED` e existe um `FinancialTransaction` do tipo `PATIENT_PAYMENT` com `Status == PAID` para aquele agendamento. Disparado best-effort a partir de dois pontos: `AppointmentHandler.Transition` (quando a transição resulta em `COMPLETED`) e `FinancialHandler.MarkPaid`. Idempotente via `FindPayoutBySource`; se nenhuma regra financeira aplicável existir ainda, fica pendente e pode ser reprocessado manualmente via `POST /api/appointments/:id/settle` (role `finance`).

### Frontend (`frontend/`)
- **Design system**: paleta "Clean Tech" (slate/sky/emerald/rose) aliada como tokens semânticos (`bg-brand-*`/`text-brand-*`) em `src/index.css` (`@theme` do Tailwind v4).
- **Auth real** (`src/lib/auth/AuthContext.tsx`): login contra `/auth/login`, JWT decodificado no cliente só para exibição (nunca confiado — a API sempre revalida), sessão persistida em `localStorage`, `ProtectedRoute` guardando o app.
- **Telas**: Login, Dashboard (contadores do dia), Agenda Semanal (grade CSS Grid, cores por status, criação de agendamento e transições de status via modal), Pacientes (busca + cadastro/edição completos + **Prontuário/PEP**, ver abaixo), **Financeiro** (Lançamentos + Regras de Repasse), **Documentos** (modelos de documentos, ver abaixo). Todas com dados reais da API, não mockados.
- **Lacunas conhecidas da integração**: não existe endpoint "todos os agendamentos de hoje da clínica" (só por profissional) — resolvido no frontend com um seletor de profissional (`useProfessionalScope`) quando o usuário logado não é `DOCTOR`/`DENTIST`. Não existe endpoint de faturamento agregado — o Dashboard soma as transações de cada agendamento individualmente (aceitável na escala atual).
- **Financeiro** (`src/pages/FinancialPage.tsx`): duas abas.
  - **Lançamentos**: lista paginada de `FinancialTransaction` (recebimentos e repasses) com filtro por situação/tipo, botão "Registrar Pagamento" (cria `PATIENT_PAYMENT` via busca de paciente + valor + forma de pagamento + convênio opcional) e "Marcar como pago" por linha (só aparece para quem já é `TENANT_ADMIN`/`FINANCE` — o backend também restringe via `finance` middleware). Dois cards de resumo (A Receber / A Repassar pendentes) somam até 100 linhas pendentes de cada tipo — mesmo tipo de limitação aceita do Dashboard, não há endpoint de soma agregada.
  - **Regras de Repasse**: seletor de profissional (reaproveita `useProfessionalScope`) + tabela de `FinancialRule` com CRUD completo pela tela — "Nova Regra" e "Editar" (`RuleFormModal.tsx`: tipo percentual/fixo, escopo procedimento/convênio, dedução de taxa, prioridade) e "Ativar/Desativar" por linha (visível só para `TENANT_ADMIN`/`FINANCE`, backend restringe via `finance` middleware).
  - Backend ganhou `GET /api/financial-transactions` (novo — antes só existia por agendamento): `FinancialTransactionRepository.ListByTenant` com filtro por `Type`/`Status` + paginação (`FinancialService.ListTransactions` default 20/máx 100 por página), coberto por teste de repositório.
  - Backend ganhou `PUT /api/financial-rules/:id` (novo — o service já tinha `UpdateRule`, mas sem handler/rota): validação (`validateRule`, compartilhada com `CreateRule`) e semântica PUT de substituição completa — `is_active` omitido é zerado, então o frontend sempre envia a regra inteira.
- **Prontuário / PEP** (`PatientRecordPage.tsx` + `components/prontuario/`): a linha do tempo do paciente mistura `Appointment` e `MedicalRecord` (evoluções clínicas) num único histórico ordenado por data, cada tipo com seu próprio visual. "Novo Registro"/"Editar" (`MedicalRecordFormModal.tsx`, textarea simples — nenhum editor rich-text foi escolhido ainda, ver dívida técnica) e "Finalizar registro" (trava `IsLocked`, depois disso o registro fica só-leitura) só aparecem para `DOCTOR`/`DENTIST` (backend também restringe via middleware `health`). "Gerar Documento" (`GenerateDocumentModal.tsx`) lista os modelos ativos, extrai via regex as tags `{{...}}` do modelo escolhido além das 3 automáticas (`patient_name`/`professional_name`/`date`, sempre resolvidas no backend, nunca confiadas do cliente) e baixa o PDF gerado.
  - **Documentos anexados** (`PatientDocumentsPanel.tsx` + `UploadDocumentModal.tsx`): upload real em duas etapas — (1) `POST .../documents/upload-url` devolve uma URL presignada, (2) o navegador dá `PUT` direto no R2 (os bytes nunca passam pelo nosso servidor), (3) `POST .../documents` grava só o metadado. "Baixar" busca uma URL presignada de leitura e abre em nova aba (arquivos sem `Content-Disposition: attachment` — como `.txt`/imagem — abrem inline, é comportamento normal do navegador, não bug). "Excluir" é `TENANT_ADMIN`-only, remove o metadado e, best-effort, o objeto no R2.
- **Documentos** (`DocumentTemplatesPage.tsx`): CRUD de `DocumentTemplate` (criar/editar/ativar-desativar), restrito a `TENANT_ADMIN` na tela e no backend (`admin` middleware) — listar fica aberto a qualquer papel autenticado, pois é o que alimenta o "Gerar Documento" do Prontuário.
- **Estoque**: ainda não tem tela — aparece na sidebar marcado "em breve".

### Infra
- **Docker**: `Dockerfile` multi-stage (build Go 1.25 → runtime Alpine, roda como `nobody`). `docker-compose.yml` com `db`/`redis`/`migrate`/`app` (dev local) e `nginx`/`certbot` atrás de um profile `production` (não sobem em dev).
- **CI** (`.github/workflows/ci.yml`): GitHub Actions rodando gofmt/vet/build/test (com Postgres de serviço + migrations aplicadas) e um build do Dockerfile, a cada push/PR na `main`. **Verde.**
- **Testes**: 64 automatizados (backend) — 45 de serviço (mocks, sem banco) + 19 de repositório (integração, banco real). Fluxos de frontend (login, agenda, CRUD de paciente, settlement, financeiro, PEP, upload de documento contra o R2 real) validados manualmente ponta a ponta com Playwright headless a cada rodada, mas não há suíte automatizada de frontend ainda.

## PEP / Documentos / Receitas

Pedido em 2026-07-27: gestão de Documentos, Receitas e Evolução Clínica. Três decisões tecnológicas oficiais do projeto:

1. **Arquivos externos (exames, laudos de terceiros)**: Cloudflare R2 (S3-compatible), acesso via **presigned URLs** geradas no Go (`internal/storage/r2.go`, usando `aws-sdk-go-v2`) — o servidor nunca vê os bytes do arquivo, só autoriza upload/download por um tempo curto. `PatientDocument` guarda só metadados (`FileKey`, nome, tamanho, mimetype).
2. **Receitas digitais/controladas**: SDK da Memed integrado no **frontend** (o widget deles fala direto com a Memed); o Go só registra um log de auditoria (`MemedPrescriptionLog`) com o ID externo da receita — nunca é o sistema de registro do conteúdo prescrito.
3. **Prontuário/evolução + PDFs locais** (atestados, laudos, declarações): geração de PDF local em Go a partir de `DocumentTemplate` (texto com tags `{{...}}`) quando for atestado/declaração/laudo avulso.

**Implementado em 2026-07-28, primeira rodada** (escopo reduzido: só a parte 3, sem dependência externa):
- `medical_record.go` / migration / `MedicalRecordRepository` / `MedicalRecordService` / `MedicalRecordHandler` — CRUD completo do prontuário. `Content` é texto simples (não HTML/JSON de rich-text — ver dívida técnica). `Update` só afeta linha com `is_locked = false` (`WHERE` atômico no repositório, não um check-then-act no service); `Lock` é idempotente. Rotas restritas a `DOCTOR`/`DENTIST` (grupo `health`), exceto leitura.
- `document_template.go` / migration / `DocumentTemplateRepository` / `DocumentTemplateService` / `DocumentTemplateHandler` — CRUD de modelos (admin-only) + `Generate`: resolve `{{tag}}` via regex (`resolvePlaceholders`, tags sem valor ficam literais no texto — nunca são silenciosamente apagadas) e renderiza PDF com `internal/pdf` (pacote novo, usa `github.com/go-pdf/fpdf` com `UnicodeTranslatorFromDescriptor` para acentuação em português funcionar nas fontes core). `patient_name`/`professional_name`/`date` são sempre resolvidos no servidor a partir de `patient_id`/`professional_id` da requisição — nunca aceitos como `vars` do cliente, pra um documento gerado não poder ser adulterado para nomear o paciente/profissional errado.
- CORS ganhou `ExposeHeaders: "Content-Disposition"` — sem isso o frontend não conseguia ler o nome do arquivo do PDF gerado (bug pego durante a verificação com Playwright: o download caía sempre como "documento.pdf").
- Frontend: timeline do Prontuário mistura `Appointment` + `MedicalRecord`; modais de criar/editar registro e de gerar documento (extrai as tags do modelo escolhido e monta um campo por tag automaticamente); tela `Documentos` para o CRUD dos modelos.

**Implementado em 2026-07-28, segunda rodada** (parte 1 — upload/download via R2 — assim que a conta Cloudflare ficou pronta):
- Conta R2 real criada pelo usuário: bucket `castilho-health-os-docs`, token de API escopado só a esse bucket (`Object Read & Write`), CORS do bucket liberado para `http://localhost:5173` (`PUT`/`GET`) — sem isso o upload direto do navegador falha com erro de CORS (peguei isso ao vivo na primeira tentativa de verificação).
- `Config.R2AccountID/R2AccessKeyID/R2SecretAccessKey/R2BucketName` (todos opcionais) + `Config.IsR2Configured()`: se as 4 variáveis não estiverem setadas, o app sobe normalmente (CI, clone novo) e só o upload/download de documento fica indisponível — isso é deliberado, pra não quebrar ambientes sem conta R2.
- `internal/storage/r2.go` ganhou `DeleteObject` (faltava — só tinha presign de upload/download).
- `PatientDocumentRepository`/`Service`/`Handler`: upload em duas etapas — `CreateUploadURL` monta o `file_key` (`tenants/{tenant_id}/patients/{patient_id}/{uuid}-{filename}`) e devolve uma URL presignada de `PUT`; o navegador sobe os bytes direto pro R2; só então `Create` grava o metadado no Postgres. `DownloadURL` resolve o `file_key` já salvo (nunca aceita do cliente) numa URL presignada de leitura. `Delete` remove o metadado primeiro e faz limpeza best-effort do objeto no R2 (falha na limpeza é só logada, não impede o delete de "funcionar" do ponto de vista de quem pediu). `ObjectStorage` é uma interface pequena definida no próprio `service` (não importa o SDK da AWS), então os testes usam um fake em vez de precisar de credenciais reais.
- Verificado ao vivo contra o bucket real via Playwright: upload → aparece na lista → download abre a URL presignada real do R2 com o conteúdo exato do arquivo enviado → exclusão remove tanto o metadado quanto o objeto no bucket.

**O que ficou de fora, de propósito** (dependem de parceria comercial, não de configuração técnica):
- `MemedPrescriptionLog` — sem migration/repo/service/handler ainda; falta parceria/credenciais Memed (integração comercial, não só técnica).
- Editor de rich-text no frontend para `MedicalRecord.Content` — continua uma `<textarea>` simples (nenhuma biblioteca escolhida ainda); trocar por um editor de verdade é compatível com o schema atual (`Content` continua sendo só `text`), não bloqueia nada.

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
- Quem pode editar/travar um `MedicalRecord` não é restrito a "só quem escreveu" — qualquer `DOCTOR`/`DENTIST` do tenant pode editar/finalizar o registro de outro profissional, mesmo padrão de "sem restrição por ownership" já aceito nas transições de agendamento.
- `MedicalRecord.Content` é texto simples, não o HTML/JSON de um editor rich-text (que ainda não foi escolhido) — formatação (negrito, listas) não existe ainda na evolução clínica.
- Deletar um `PatientDocument` remove o objeto do R2 só best-effort (log, não erro) — se a chamada de delete no R2 falhar, o metadado já foi removido e o objeto fica órfão no bucket (invisível pro app, mas ainda ocupando espaço).
- CORS do bucket R2 é configurado manualmente no painel da Cloudflare (não pelo app) — ao apontar pra um domínio de produção, alguém precisa lembrar de adicionar esse domínio em `AllowedOrigins` lá, senão o upload quebra silenciosamente só em prod.

## Próximos passos (recomendação de ordem — decidir na volta)

Avaliado em 2026-07-27, concluído em 2026-07-28: a prioridade era **Financeiro (tela) antes de PEP/Documentos completo** — feito, incluindo o CRUD de regras que tinha ficado pendente. Em seguida, o PEP foi implementado em duas rodadas no mesmo dia: primeiro o escopo sem dependência externa (prontuário + templates/PDF), depois o upload/download via R2 assim que a conta Cloudflare ficou pronta. Só falta a Memed pra fechar o PEP por completo. Próxima decisão em aberto entre: Memed (depende de parceria comercial, fora do nosso controle), Odontograma (módulo novo do zero, sem bloqueio externo), ou um editor de rich-text de verdade para o prontuário.

1. ~~Settlement financeiro automático~~ — feito.
2. ~~Frontend (Dashboard, Agenda, Pacientes)~~ — feito.
3. ~~Tela de Financeiro~~ — feito (2026-07-28): lançamentos (listar/filtrar/paginar, registrar pagamento, marcar como pago) + regras de repasse (CRUD completo: criar, editar, ativar/desativar).
4. ~~PEP — prontuário, templates/PDF~~ — feito (2026-07-28): CRUD de prontuário com trava de finalização, CRUD de modelos de documento, geração de PDF local com tags dinâmicas.
5. ~~PEP — upload/download de documento via R2~~ — feito (2026-07-28): conta Cloudflare real configurada, upload em duas etapas (presigned URL + PUT direto do navegador), download via URL presignada, exclusão com limpeza best-effort do objeto. Verificado ao vivo contra o bucket real.
6. **Memed** — falta parceria/credenciais comerciais (não é um problema técnico); `MemedPrescriptionLog` já modelado, sem migration/repo/service/handler ainda.
7. **Editor de rich-text para o prontuário** — hoje é uma `<textarea>` simples; nenhuma biblioteca escolhida ainda.
8. **Odontograma interativo** — módulo exclusivo odonto.
9. **Integração WhatsApp** — confirmação automática 24h antes, processando resposta do paciente.
10. **Estoque** — nenhum modelo ainda; começar do zero.
11. **Deploy na VPS Hostinger** — Nginx+Certbot já configurados (`nginx/conf.d/`), falta executar de fato quando o domínio estiver apontado.
12. **Uso real do Redis** — sessões e/ou cache de agenda.

## Como retomar

As credenciais R2 (`R2_ACCOUNT_ID`/`R2_ACCESS_KEY_ID`/`R2_SECRET_ACCESS_KEY`/`R2_BUCKET_NAME`) já estão no `.env` local (gitignored, nunca commitado) — não precisa reconfigurar nada pra rodar o upload de documentos.

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
