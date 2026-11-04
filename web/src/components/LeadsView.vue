<script setup lang="ts">
import { computed, ref } from 'vue'
import { money, states, type Lead } from '../api'
const props = defineProps<{ leads: Lead[] }>()
const state = ref(''); const query = ref(''); const sort = ref('newest')
const visible = computed(() => props.leads.filter(lead => (!state.value || lead.state === state.value) &&
  `${lead.first_name} ${lead.last_name} ${lead.email}`.toLowerCase().includes(query.value.toLowerCase()))
  .sort((a, b) => sort.value === 'name' ? `${a.first_name} ${a.last_name}`.localeCompare(`${b.first_name} ${b.last_name}`) : b.id - a.id))
</script>
<template>
  <div class="flex flex-col gap-3 border-b border-slate-200 p-5 sm:flex-row sm:items-end">
    <div class="flex-1"><label for="search">Search leads</label><input id="search" v-model="query" type="search" placeholder="Search name or email…" /></div>
    <div><label for="filter">State</label><select id="filter" v-model="state"><option value="">All states</option><option v-for="item in states" :key="item">{{ item }}</option></select></div>
    <div><label for="sort">Sort by</label><select id="sort" v-model="sort"><option value="newest">Newest first</option><option value="name">Name A–Z</option></select></div>
  </div>
  <div class="overflow-x-auto"><table class="w-full text-left text-sm">
    <caption class="sr-only">Saved leads with contact, finance and vehicle details</caption>
    <thead class="bg-slate-50 text-xs uppercase tracking-wider text-slate-500"><tr><th scope="col">Contact</th><th scope="col">State</th><th scope="col">Finance amount</th><th scope="col">Annual salary</th><th scope="col">Vehicle</th></tr></thead>
    <tbody class="divide-y divide-slate-100"><tr v-for="lead in visible" :key="lead.id" class="hover:bg-slate-50/70">
      <td><div class="font-semibold text-slate-900">{{ lead.first_name }} {{ lead.last_name }}</div><div class="mt-1 text-slate-500">{{ lead.email }}</div></td>
      <td><span class="rounded-md bg-slate-100 px-2 py-1 text-xs font-semibold text-slate-600">{{ lead.state }}</span></td>
      <td class="whitespace-nowrap font-medium tabular-nums">{{ money(lead.finance_amount) }}</td>
      <td class="whitespace-nowrap text-slate-500 tabular-nums">{{ lead.salary === null ? '—' : money(lead.salary) }}</td>
      <td class="text-slate-600">{{ lead.vehicle ? [lead.vehicle.make, lead.vehicle.model, lead.vehicle.year].filter(Boolean).join(' ') : '—' }}</td>
    </tr><tr v-if="!visible.length"><td colspan="5" class="py-16 text-center text-slate-500">{{ leads.length ? 'No leads match your filters.' : 'No leads yet. Import a CSV or add your first lead.' }}</td></tr></tbody>
  </table></div>
  <p class="border-t border-slate-100 px-5 py-4 text-xs text-slate-500">Showing {{ visible.length }} of {{ leads.length }} leads · Amounts in AUD</p>
</template>
