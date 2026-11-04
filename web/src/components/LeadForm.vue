<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ApiError, request, states, type Lead, type LeadInput } from '../api'
const emit = defineEmits<{ saved: [lead: Lead]; cancel: [] }>()
const blank = () => ({ first_name: '', last_name: '', email: '', finance_amount: '', salary: '', state: '', make: '', model: '', year: '' })
const form = reactive(blank())
const errors = ref<Record<string, string>>({})
const message = ref('')
const busy = ref(false)
const fields = [
  { key: 'first_name', label: 'First name', type: 'text', autocomplete: 'given-name' },
  { key: 'last_name', label: 'Last name', type: 'text', autocomplete: 'family-name' },
  { key: 'email', label: 'Email address', type: 'email', autocomplete: 'email' },
  { key: 'finance_amount', label: 'Finance amount (AUD)', type: 'number', autocomplete: 'off' },
  { key: 'salary', label: 'Annual salary (AUD, optional)', type: 'number', autocomplete: 'off' },
] as const
async function save() {
  busy.value = true; errors.value = {}; message.value = ''
  const input: LeadInput = {
    first_name: form.first_name, last_name: form.last_name, email: form.email,
    finance_amount: Number(form.finance_amount), salary: form.salary === '' ? null : Number(form.salary), state: form.state,
    vehicle: form.make || form.model || form.year ? { make: form.make, model: form.model, year: form.year === '' ? null : Number(form.year) } : null,
  }
  try {
    const lead = await request<Lead>('/api/leads', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(input) })
    emit('saved', lead); Object.assign(form, blank())
  } catch (error) {
    if (error instanceof ApiError) { errors.value = error.fields; message.value = error.message }
    else message.value = 'Could not reach the server. Please try again.'
  } finally { busy.value = false }
}
</script>
<template>
  <form class="space-y-5" @submit.prevent="save">
    <p class="text-sm text-slate-500">Contact and finance details are required. Salary and vehicle details are optional.</p>
    <p v-if="message" role="alert" class="rounded-lg bg-red-50 p-3 text-sm text-red-800">{{ message }}</p>
    <div class="grid gap-4 sm:grid-cols-2">
      <div v-for="field in fields" :key="field.key" :class="{ 'sm:col-span-2': field.key === 'email' }">
        <label :for="field.key">{{ field.label }}</label>
        <input :id="field.key" v-model="form[field.key]" :type="field.type" :autocomplete="field.autocomplete" :required="field.key !== 'salary'"
          :step="field.type === 'number' ? 'any' : undefined" :min="field.type === 'number' ? '0' : undefined"
          :aria-invalid="!!errors[field.key]" :aria-describedby="errors[field.key] ? `${field.key}-error` : undefined" />
        <p v-if="errors[field.key]" :id="`${field.key}-error`" class="field-error">{{ errors[field.key] }}</p>
      </div>
      <div><label for="state">State or territory</label>
        <select id="state" v-model="form.state" required :aria-invalid="!!errors.state" aria-describedby="state-error"><option disabled value="">Select a state</option><option v-for="state in states" :key="state">{{ state }}</option></select>
        <p v-if="errors.state" id="state-error" class="field-error">{{ errors.state }}</p>
      </div>
    </div>
    <fieldset class="rounded-xl border border-slate-200 p-4">
      <legend class="px-2 text-sm font-semibold">Vehicle <span class="font-normal text-slate-500">· optional</span></legend>
      <div class="grid gap-4 sm:grid-cols-3"><div v-for="key in (['make', 'model', 'year'] as const)" :key="key">
        <label :for="key" class="capitalize">{{ key }}</label>
        <input :id="key" v-model="form[key]" :type="key === 'year' ? 'number' : 'text'" :min="key === 'year' ? 1990 : undefined"
          :max="key === 'year' ? new Date().getFullYear() + 1 : undefined" :required="key === 'make' && !!(form.model || form.year)"
          :aria-invalid="!!errors[`vehicle.${key}`]" :aria-describedby="`${key}-error`" />
        <p v-if="errors[`vehicle.${key}`]" :id="`${key}-error`" class="field-error">{{ errors[`vehicle.${key}`] }}</p>
      </div></div>
    </fieldset>
    <div class="flex justify-end gap-3 border-t border-slate-100 pt-5"><button type="button" class="secondary" :disabled="busy" @click="emit('cancel')">Cancel</button><button type="submit" class="primary" :disabled="busy">{{ busy ? 'Saving…' : 'Save lead' }}</button></div>
  </form>
</template>

