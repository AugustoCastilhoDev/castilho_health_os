import { useEffect, useRef, useState, type FormEvent } from 'react'
import { TextField } from '../components/common/TextField'
import { useAuth } from '../lib/auth/AuthContext'
import { ApiError } from '../lib/api/client'
import type { TenantDTO } from '../lib/api/types'

interface ProfileFormState {
  name: string
  document: string
  email: string
  phone: string
  addressStreet: string
  addressCity: string
  addressState: string
  addressZip: string
}

function toFormState(tenant: TenantDTO): ProfileFormState {
  return {
    name: tenant.name,
    document: tenant.document,
    email: tenant.email,
    phone: tenant.phone ?? '',
    addressStreet: tenant.address_street ?? '',
    addressCity: tenant.address_city ?? '',
    addressState: tenant.address_state ?? '',
    addressZip: tenant.address_zip ?? '',
  }
}

export function ConfiguracoesPage() {
  const { apiFetch, tenant, refreshTenant } = useAuth()
  const [form, setForm] = useState<ProfileFormState | null>(null)
  const [savingProfile, setSavingProfile] = useState(false)
  const [profileError, setProfileError] = useState<string | null>(null)
  const [profileSaved, setProfileSaved] = useState(false)

  const [logoUrl, setLogoUrl] = useState<string | null>(null)
  const [logoLoading, setLogoLoading] = useState(true)
  const [uploadingLogo, setUploadingLogo] = useState(false)
  const [logoError, setLogoError] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (tenant) setForm(toFormState(tenant))
  }, [tenant])

  useEffect(() => {
    let cancelled = false
    setLogoLoading(true)
    apiFetch<{ url: string }>('/api/tenant/logo/download-url')
      .then((data) => {
        if (!cancelled) setLogoUrl(data.url || null)
      })
      .catch(() => {
        if (!cancelled) setLogoUrl(null)
      })
      .finally(() => {
        if (!cancelled) setLogoLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [apiFetch])

  function set<K extends keyof ProfileFormState>(key: K, value: string) {
    setForm((prev) => (prev ? { ...prev, [key]: value } : prev))
    setProfileSaved(false)
  }

  async function handleSaveProfile(e: FormEvent) {
    e.preventDefault()
    if (!form) return
    setProfileError(null)
    setProfileSaved(false)
    setSavingProfile(true)
    try {
      await apiFetch<TenantDTO>('/api/tenant', {
        method: 'PUT',
        body: JSON.stringify({
          name: form.name,
          document: form.document,
          email: form.email,
          phone: form.phone,
          address_street: form.addressStreet || null,
          address_city: form.addressCity || null,
          address_state: form.addressState || null,
          address_zip: form.addressZip || null,
        }),
      })
      await refreshTenant()
      setProfileSaved(true)
    } catch (err) {
      setProfileError(err instanceof ApiError ? err.message : 'Não foi possível salvar os dados da clínica.')
    } finally {
      setSavingProfile(false)
    }
  }

  async function handleUploadLogo() {
    const file = fileInputRef.current?.files?.[0]
    if (!file) {
      setLogoError('Selecione uma imagem.')
      return
    }
    setLogoError(null)
    setUploadingLogo(true)
    try {
      const { upload_url, file_key } = await apiFetch<{ upload_url: string; file_key: string }>(
        '/api/tenant/logo/upload-url',
        { method: 'POST', body: JSON.stringify({ file_name: file.name, content_type: file.type || 'image/png' }) },
      )
      const uploadRes = await fetch(upload_url, {
        method: 'PUT',
        headers: { 'Content-Type': file.type || 'image/png' },
        body: file,
      })
      if (!uploadRes.ok) {
        throw new ApiError(uploadRes.status, 'Falha ao enviar o logotipo para o armazenamento.')
      }
      await apiFetch<TenantDTO>('/api/tenant/logo', {
        method: 'POST',
        body: JSON.stringify({ file_key }),
      })
      const { url } = await apiFetch<{ url: string }>('/api/tenant/logo/download-url')
      setLogoUrl(url || null)
      if (fileInputRef.current) fileInputRef.current.value = ''
    } catch (err) {
      setLogoError(err instanceof ApiError ? err.message : 'Não foi possível enviar o logotipo.')
    } finally {
      setUploadingLogo(false)
    }
  }

  async function handleRemoveLogo() {
    setLogoError(null)
    setUploadingLogo(true)
    try {
      await apiFetch('/api/tenant/logo', { method: 'DELETE' })
      setLogoUrl(null)
    } catch (err) {
      setLogoError(err instanceof ApiError ? err.message : 'Não foi possível remover o logotipo.')
    } finally {
      setUploadingLogo(false)
    }
  }

  if (!form) {
    return <main className="p-6 text-sm text-brand-text-muted">Carregando…</main>
  }

  return (
    <>
      <header className="border-b border-slate-200 bg-brand-surface px-8 py-5">
        <p className="text-sm text-brand-text-muted">Clínica</p>
        <h1 className="text-xl font-semibold text-brand-text">Configurações</h1>
      </header>

      <main className="max-w-3xl space-y-6 p-6">
        <div className="rounded-xl bg-brand-surface p-6 shadow-sm ring-1 ring-slate-200">
          <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-brand-text-muted">
            Logotipo (papel timbrado)
          </h2>
          <p className="mb-4 text-sm text-brand-text-muted">
            Aparece no cabeçalho dos documentos gerados (atestados, laudos, declarações) quando o modelo tiver essa
            opção marcada.
          </p>
          <div className="flex items-center gap-6">
            <div className="flex h-24 w-24 shrink-0 items-center justify-center rounded-lg border border-dashed border-slate-300 bg-slate-50">
              {logoLoading ? (
                <span className="text-xs text-brand-text-muted">Carregando…</span>
              ) : logoUrl ? (
                <img src={logoUrl} alt="Logotipo da clínica" className="max-h-full max-w-full object-contain" />
              ) : (
                <span className="px-2 text-center text-xs text-brand-text-muted">Sem logotipo</span>
              )}
            </div>
            <div className="flex-1 space-y-2">
              <input
                ref={fileInputRef}
                type="file"
                accept="image/png,image/jpeg"
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm file:mr-3 file:rounded-md file:border-0 file:bg-slate-100 file:px-3 file:py-1.5 file:text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
              />
              <div className="flex gap-2">
                <button
                  type="button"
                  onClick={handleUploadLogo}
                  disabled={uploadingLogo}
                  className="rounded-lg bg-brand-action px-4 py-2 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {uploadingLogo ? 'Enviando…' : 'Enviar logotipo'}
                </button>
                {logoUrl && (
                  <button
                    type="button"
                    onClick={handleRemoveLogo}
                    disabled={uploadingLogo}
                    className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-brand-text hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    Remover
                  </button>
                )}
              </div>
              {logoError && <p className="text-sm text-brand-alert-text">{logoError}</p>}
            </div>
          </div>
        </div>

        <form
          onSubmit={handleSaveProfile}
          className="space-y-4 rounded-xl bg-brand-surface p-6 shadow-sm ring-1 ring-slate-200"
        >
          <h2 className="text-sm font-semibold uppercase tracking-wide text-brand-text-muted">Dados da clínica</h2>

          <TextField label="Nome" value={form.name} onChange={(v) => set('name', v)} required />
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <TextField label="CNPJ/CPF" value={form.document} onChange={(v) => set('document', v)} required />
            <TextField label="Telefone" value={form.phone} onChange={(v) => set('phone', v)} />
          </div>
          <TextField label="E-mail" type="email" value={form.email} onChange={(v) => set('email', v)} required />

          <div>
            <p className="mb-2 text-sm font-medium text-brand-text">Endereço</p>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-4">
              <TextField label="CEP" value={form.addressZip} onChange={(v) => set('addressZip', v)} />
              <div className="sm:col-span-2">
                <TextField label="Rua" value={form.addressStreet} onChange={(v) => set('addressStreet', v)} />
              </div>
              <TextField label="Cidade" value={form.addressCity} onChange={(v) => set('addressCity', v)} />
            </div>
            <div className="mt-4 max-w-[120px]">
              <TextField label="UF" value={form.addressState} onChange={(v) => set('addressState', v)} />
            </div>
          </div>

          {profileError && <p className="text-sm text-brand-alert-text">{profileError}</p>}
          {profileSaved && <p className="text-sm text-brand-success-text">Dados salvos com sucesso.</p>}

          <button
            type="submit"
            disabled={savingProfile}
            className="rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
          >
            {savingProfile ? 'Salvando…' : 'Salvar Alterações'}
          </button>
        </form>
      </main>
    </>
  )
}
