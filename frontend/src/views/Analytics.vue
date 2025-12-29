<template>
  <div class="p-8 space-y-8">
    <div class="flex justify-between items-center">
      <h1 class="text-3xl font-semibold">Analytics</h1>
      <div class="flex gap-4">
        <select v-model="timeRange" @change="fetchData" class="bg-gray-800 px-4 py-2 rounded">
          <option value="1">24h</option>
          <option value="7">7 dni</option>
          <option value="30">30 dni</option>
        </select>
        <button @click="fetchData" class="bg-blue-600 px-6 py-2 rounded hover:bg-blue-700">
          Refresh
        </button>
      </div>
    </div>

    <!-- Wykresy 2×2 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Alerts Over Time -->
      <div class="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-xl p-6 flex flex-col">
        <h3 class="text-xl font-semibold mb-4 flex items-center gap-2">
          Alerts Over Time
          <span class="text-xs bg-blue-500/20 text-blue-300 px-2 py-1 rounded-full">{{ timeRange }}d</span>
        </h3>
        <div class="flex-1">
          <LineChart
            v-if="chartData.hourly && chartData.hourly.labels"
            :data="chartData.hourly"
            :options="hourlyOptions"
          />
        </div>
      </div>

      <!-- Top Attacking IPs -->
      <div class="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-xl p-6 flex flex-col">
        <h3 class="text-xl font-semibold mb-4">Top Attacking IPs</h3>
        
        <div class="flex-1 space-y-3">
          <div 
            v-for="(item, index) in topIPsData" 
            :key="item.ip"
            class="relative"
          >
            <div class="absolute inset-0 bg-red-500/10 rounded" 
                :style="{ width: (item.count / topIPsData[0].count * 100) + '%' }">
            </div>
            <div class="relative flex items-center justify-between p-3">
              <div class="flex items-center gap-3">
                <span class="text-gray-500 font-mono text-sm w-4">{{ index + 1 }}</span>
                <span class="font-mono text-sm text-red-400">{{ item.ip }}</span>
              </div>
              <span class="text-lg font-bold text-red-300">{{ item.count }}</span>
            </div>
          </div>

          <div v-if="!topIPsData || topIPsData.length === 0" class="text-gray-500 text-sm">
            No data
          </div>
        </div>
      </div>

      <!-- Attack Categories -->
      <div class="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-xl p-6 flex flex-col">
        <h3 class="text-xl font-semibold mb-4">Attack Categories</h3>
        <div class="flex-1">
          <BarChart
            v-if="chartData.categories && chartData.categories.labels"
            :data="chartData.categories"
            :options="categoriesOptions"
          />
        </div>
      </div>

      <!-- Blocks Timeline -->
      <div class="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-xl p-6 flex flex-col">
        <h3 class="text-xl font-semibold mb-4">Blocks Timeline</h3>
        <div class="flex-1">
          <LineChart
            v-if="chartData.blocks && chartData.blocks.labels"
            :data="chartData.blocks"
            :options="blocksOptions"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Line as LineChart, Bar as BarChart } from 'vue-chartjs'
import { Chart, LineElement, PointElement, LinearScale, CategoryScale, Tooltip, Legend, Filler, BarElement } from 'chart.js'
import api from '@/services/api'
import { formatTimestamp } from '@/utils/dateHelpers'
import { baseLineChartOptions, barChartOptions } from '@/utils/chartOptions'

Chart.register(LineElement, PointElement, LinearScale, CategoryScale, Tooltip, Legend, Filler, BarElement)

const chartData = ref({
  hourly: { labels: [], datasets: [] },
  topIPs: { labels: [], datasets: [] },
  categories: { labels: [], datasets: [] },
  blocks: { labels: [], datasets: [] }
})

const topIPsData = ref([])
const timeRange = ref('1')
const ipSearch = ref('')
const ipHistory = ref([])

const hourlyOptions = baseLineChartOptions
const blocksOptions = { ...baseLineChartOptions, tension: 0.4 }
const categoriesOptions = barChartOptions

async function fetchData() {
  try {
    const [alertBucketsJson, topJson, catsJson, blockBucketsJson] = await Promise.all([
      api.getAlertBuckets(timeRange.value),
      api.getTopIPs(5),
      api.getCategories(timeRange.value),
      api.getBlockBuckets(timeRange.value)
    ])

    chartData.value.hourly = prepareBucketLineData(alertBucketsJson.data || [], 'Alerts')
    topIPsData.value = (topJson.data || []).slice(0, 5)
    chartData.value.categories = prepareBarData(catsJson.data || [])
    chartData.value.blocks = prepareBucketLineData(blockBucketsJson.data || [], 'Blocks')
  } catch (error) {
    console.error('Failed to fetch analytics:', error)
  }
}

function prepareBucketLineData(data, label) {
  return {
    labels: data.map(d => d.bucket),
    datasets: [{
      label,
      data: data.map(d => d.count),
      borderColor: label === 'Blocks' ? '#F97316' : '#3B82F6',
      backgroundColor: label === 'Blocks'
        ? 'rgba(249,115,22,0.15)'
        : 'rgba(59,130,246,0.1)',
      fill: true,
      tension: 0.3
    }]
  }
}

function prepareBarData(data) {
  return {
    labels: data.map(d => {
      let name = d.name.replace(/^ET\s+/, '')
      return name.length > 25 ? name.slice(0, 22) + '...' : name
    }),
    datasets: [{
      label: 'Attacks',
      data: data.map(d => d.count),
      backgroundColor: '#EF4444',
      borderRadius: 6
    }]
  }
}

async function runIpSearch() {
  const ip = ipSearch.value.trim()
  if (!ip) {
    ipHistory.value = []
    return
  }

  try {
    const [alertsJson, blocksJson] = await Promise.all([
      api.getAlertsByIP(ip),
      api.getBlockedByIP(ip)
    ])

    const alertsRows = (alertsJson.alerts || []).map(a => ({
      id: `a-${a.id}`,
      type: 'alert',
      ip: a.ip,
      time: formatTimestamp(a.timestamp),
      sid: a.sid,
      score: null,
      message: a.message || ''
    }))

    const blockRows = (blocksJson.blocked_ips || []).map(b => ({
      id: `b-${b.ip}-${b.timestamp}`,
      type: 'block',
      ip: b.ip,
      time: formatTimestamp(b.timestamp),
      sid: null,
      score: b.score,
      message: b.reason || ''
    }))

    ipHistory.value = [...alertsRows, ...blockRows].sort((a, b) => {
      return new Date(b.time) - new Date(a.time)
    })

  } catch (error) {
    console.error('IP search failed:', error)
    ipHistory.value = []
  }
}

onMounted(() => {
  fetchData()
})
</script>

