interface TextFieldProps {
  label: string
  type?: string
  value: string
  onChange: (value: string) => void
  placeholder?: string
  required?: boolean
}

export function TextField({ label, type = 'text', value, onChange, placeholder, required }: TextFieldProps) {
  return (
    <label className="block">
      <span className="mb-1 block text-sm font-medium text-brand-text">{label}</span>
      <input
        type={type}
        required={required}
        value={value}
        placeholder={placeholder}
        onChange={(e) => onChange(e.target.value)}
        className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-brand-text focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
      />
    </label>
  )
}
