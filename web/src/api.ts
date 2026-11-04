export interface Vehicle { make: string; model: string; year: number | null }
export interface LeadInput {
  first_name: string; last_name: string; email: string; finance_amount: number
  salary: number | null; state: string; vehicle: Vehicle | null
}
export interface Lead extends LeadInput { id: number; created_at: string }
export interface ImportResult { imported: number; rejected: { line: number; errors: Record<string, string> }[] }
export class ApiError extends Error {
  constructor(message: string, public fields: Record<string, string> = {}, public result?: ImportResult) { super(message) }
}
export const states = ['ACT', 'NSW', 'NT', 'QLD', 'SA', 'TAS', 'VIC', 'WA']
export const money = (value: number) => new Intl.NumberFormat('en-AU', {
  style: 'currency', currency: 'AUD', maximumFractionDigits: 2,
}).format(value)
export async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(path, options)
  const data = await response.json()
  if (!response.ok) throw new ApiError(data.error || 'Please correct the highlighted fields.', data.errors, data.result)
  return data as T
}
