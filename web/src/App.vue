<script setup>
import { computed, onMounted, ref } from 'vue'

const loading = ref(true)
const error = ref('')
const generatedAt = ref('')
const groups = ref([])
const activeGroup = ref('')
const keyword = ref('')

const visibleGroups = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  if (!q) return groups.value
  return groups.value
    .map((g) => ({
      ...g,
      calendars: g.calendars.filter((c) => {
        const text = `${c.name} ${c.description || ''} ${c.provider}`.toLowerCase()
        return text.includes(q)
      })
    }))
    .filter((g) => g.calendars.length > 0)
})

const currentGroup = computed(() => {
  if (!visibleGroups.value.length) return null
  const matched = visibleGroups.value.find((g) => g.key === activeGroup.value)
  return matched || visibleGroups.value[0]
})

const totalCalendars = computed(() => groups.value.reduce((sum, g) => sum + g.calendars.length, 0))

function resolveIcsUrl(path) {
  if (/^https?:\/\//i.test(path)) return path
  const clean = String(path || '').replace(/^\/+/, '')
  return `${window.location.origin}${window.location.pathname.replace(/\/[^/]*$/, '/')}${clean}`
}

async function copyText(text) {
  try {
    await navigator.clipboard.writeText(text)
    window.alert('订阅链接已复制')
  } catch {
    window.alert('复制失败，请手动复制')
  }
}

onMounted(async () => {
  try {
    const resp = await fetch('./data/subscriptions.json')
    if (!resp.ok) {
      throw new Error(`请求失败: ${resp.status}`)
    }
    const payload = await resp.json()
    generatedAt.value = payload.generatedAt
    groups.value = Array.isArray(payload.groups) ? payload.groups : []
    activeGroup.value = groups.value[0]?.key || ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <main class="page">
    <section class="hero">
      <h1>ICS 订阅中心</h1>
      <p>统一数据源生成 JSON + ICS，按分组快速筛选并一键订阅。</p>
      <div class="meta">
        <span>日历总数：{{ totalCalendars }}</span>
        <span v-if="generatedAt">更新时间：{{ new Date(generatedAt).toLocaleString() }}</span>
      </div>
    </section>

    <section class="panel">
      <input
        v-model="keyword"
        class="search"
        type="search"
        placeholder="搜索日历名称 / 描述 / provider"
      />

      <div v-if="loading" class="state">数据加载中...</div>
      <div v-else-if="error" class="state error">{{ error }}</div>
      <template v-else>
        <div class="tabs" v-if="visibleGroups.length">
          <button
            v-for="group in visibleGroups"
            :key="group.key"
            class="tab"
            :class="{ active: currentGroup?.key === group.key }"
            @click="activeGroup = group.key"
          >
            {{ group.name }}
            <small>{{ group.calendars.length }}</small>
          </button>
        </div>

        <div v-if="!currentGroup" class="state">没有匹配的数据</div>

        <div v-else class="cards">
          <article class="card" v-for="cal in currentGroup.calendars" :key="cal.id">
            <h3>{{ cal.name }}</h3>
            <p class="desc">{{ cal.description || '无描述' }}</p>
            <div class="row"><span>Provider</span><strong>{{ cal.provider }}</strong></div>
            <div class="row"><span>事件数</span><strong>{{ cal.eventCount }}</strong></div>
            <div class="actions">
              <a :href="cal.icsPath" target="_blank" rel="noreferrer">下载 ICS</a>
              <button type="button" @click="copyText(resolveIcsUrl(cal.icsPath))">复制订阅链接</button>
            </div>
          </article>
        </div>
      </template>
    </section>
  </main>
</template>
