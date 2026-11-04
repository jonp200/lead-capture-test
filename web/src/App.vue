<script setup lang="ts">
import LeadForm from './components/LeadForm.vue'
import LeadsView from './components/LeadsView.vue'
</script>

<template>
  <div class="min-h-screen bg-slate-50 text-slate-800">
    <header class="bg-blue-900 px-8 py-4 text-white">
      <h1 class="text-2xl font-semibold">Leads</h1>
    </header>

    <main class="grid gap-8 p-8">
      <section>
        <h2 class="mb-4 text-xl font-semibold">Add a lead</h2>
        <LeadForm />
      </section>

      <section>
        <h2 class="mb-4 text-xl font-semibold">Leads</h2>
        <LeadsView />
      </section>
    </main>
  </div>
</template>
