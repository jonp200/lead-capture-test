<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import LeadForm from './components/LeadForm.vue'
import LeadsView from './components/LeadsView.vue'
import { ApiError, money, request, type ImportResult, type Lead } from './api'
const leads = ref<Lead[]>([])
const loading = ref(true)
const loadError = ref('')
const notice = ref('')
const showForm = ref(false)
const dialog = ref<HTMLDialogElement>()
const uploading = ref(false)
const report = ref<ImportResult>()
const uploadError = ref('')
const total = computed(() => leads.value.reduce((sum, lead) => sum + lead.finance_amount, 0))
async function load() {
  loading.value = true; loadError.value = ''
  try { leads.value = await request<Lead[]>('/api/leads') }
  catch { loadError.value = 'Could not load leads. Check the server connection and try again.' }
  finally { loading.value = false }
}
async function openForm() { showForm.value = true; await nextTick(); dialog.value?.showModal() }
function closeForm() { dialog.value?.close(); showForm.value = false }
function saved(lead: Lead) { leads.value.unshift(lead); notice.value = 'Lead saved successfully.'; closeForm() }
async function upload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  uploadError.value = ''; report.value = undefined; notice.value = ''
  if (file.size > 10 * 1024 * 1024) { uploadError.value = 'Choose a CSV smaller than 10 MB.'; input.value = ''; return }
  uploading.value = true
  try { report.value = await request<ImportResult>('/api/leads/import', { method: 'POST', headers: { 'Content-Type': 'text/csv' }, body: file }) }
  catch (error) {
    uploadError.value = error instanceof ApiError ? error.message : 'Could not upload the file. Please try again.'
    if (error instanceof ApiError) report.value = error.result
  } finally { uploading.value = false; input.value = ''; await load() }
}
onMounted(load)
</script>
<template>
  <div class="min-h-screen bg-slate-50 text-slate-800">
    <header class="border-b border-slate-200 bg-white">
      <div class="mx-auto flex max-w-7xl items-center gap-3 px-5 py-5 sm:px-8">
        <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-teal-800 text-lg font-bold text-white" aria-hidden="true">N</div>
        <div><p class="font-bold tracking-tight text-slate-900">Novated Lease Australia</p><p class="text-xs text-slate-500">Partner lead workspace</p></div>
        <span class="ml-auto hidden rounded-full bg-teal-50 px-3 py-1 text-xs font-medium text-teal-800 sm:block">Staff portal</span>
      </div>
    </header>
    <main class="mx-auto max-w-7xl space-y-6 px-5 py-10 sm:px-8">
      <div class="flex flex-wrap items-center justify-between gap-5">
        <div><p class="mb-2 text-xs font-semibold uppercase tracking-widest text-teal-700">Lead management</p><h1 class="text-3xl font-bold tracking-tight text-slate-950">Your next conversations.</h1><p class="mt-2 text-sm text-slate-500">Bring partner leads together and keep new opportunities moving.</p></div>
        <button class="primary" @click="openForm">＋ Add a lead</button>
      </div>
      <p v-if="notice" role="status" class="rounded-xl border border-teal-200 bg-teal-50 p-4 text-sm text-teal-900">{{ notice }}</p>
      <div class="grid gap-4 sm:grid-cols-3">
        <div class="card p-5"><p class="stat-label">Total leads</p><p class="stat-value">{{ loading ? '…' : leads.length }}</p></div>
        <div class="card p-5"><p class="stat-label">Finance requested</p><p class="stat-value">{{ loading ? '…' : money(total) }}</p></div>
        <div class="card p-5"><p class="stat-label">States & territories</p><p class="stat-value">{{ loading ? '…' : new Set(leads.map(lead => lead.state)).size }} <span class="text-sm font-normal text-slate-400">/ 8</span></p></div>
      </div>
      <section class="card p-5 sm:p-6" aria-labelledby="import-title">
        <div class="flex flex-wrap items-center justify-between gap-4"><div><h2 id="import-title" class="font-semibold text-slate-900">Import partner leads</h2><p class="mt-1 text-sm text-slate-500">Upload a CSV up to 10 MB. Valid rows are saved; issues are listed for correction.</p></div>
          <label class="secondary relative cursor-pointer focus-within:ring-2 focus-within:ring-teal-700" :class="{ 'opacity-50': uploading }">{{ uploading ? 'Importing…' : '↑ Upload CSV' }}<input class="absolute inset-0 cursor-pointer opacity-0" type="file" accept=".csv,text/csv" aria-label="Upload leads CSV" :disabled="uploading" @change="upload" /></label>
        </div>
        <p v-if="uploadError" role="alert" class="mt-4 text-sm text-red-800">{{ uploadError }}</p>
        <div v-if="report" class="mt-5 rounded-lg border border-slate-200 p-4" role="status">
          <p class="text-sm font-semibold">{{ report.imported }} imported · {{ report.rejected.length }} need review</p>
          <details v-if="report.rejected.length" class="mt-3 text-sm"><summary class="cursor-pointer font-medium text-amber-800">Review skipped rows</summary>
            <p class="mt-3 text-slate-500">Correct these rows in the source file and upload them again. Existing emails are skipped.</p>
            <ul class="mt-3 max-h-64 space-y-2 overflow-auto"><li v-for="row in report.rejected" :key="row.line"><strong>Line {{ row.line }}:</strong> {{ Object.entries(row.errors).map(([field, message]) => `${field}: ${message}`).join('; ') }}</li></ul>
          </details>
        </div>
      </section>
      <section class="card overflow-hidden" aria-labelledby="leads-title">
        <div class="flex items-center justify-between border-b border-slate-200 px-5 py-5"><h2 id="leads-title" class="font-semibold text-slate-900">All leads</h2><button class="text-sm font-medium text-teal-700 disabled:opacity-50" :disabled="loading" @click="load">Refresh</button></div>
        <p v-if="loading" role="status" class="p-12 text-center text-slate-500">Loading leads…</p>
        <div v-else-if="loadError" role="alert" class="p-8 text-center"><p class="text-red-800">{{ loadError }}</p><button class="secondary mt-4" @click="load">Try again</button></div>
        <LeadsView v-else :leads="leads" />
      </section>
      <p class="text-center text-xs text-slate-400">Partner lead workspace · Handle customer information with care</p>
    </main>
    <dialog v-if="showForm" ref="dialog" class="m-auto max-h-[90vh] w-[min(94vw,640px)] overflow-y-auto rounded-2xl bg-white p-6 shadow-xl backdrop:bg-slate-950/40 sm:p-8" aria-labelledby="form-title" @cancel="closeForm">
      <div class="mb-5 flex items-center justify-between"><h2 id="form-title" class="text-xl font-bold">Add a lead</h2><button aria-label="Close add lead form" class="text-2xl text-slate-500" @click="closeForm">×</button></div>
      <LeadForm @saved="saved" @cancel="closeForm" />
    </dialog>
  </div>
</template>

