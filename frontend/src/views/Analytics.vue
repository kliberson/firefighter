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

    <!-- Tabela z filtrowaniem -->
    <div class="bg-gray-800 rounded-xl p-6">
      <div class="flex flex-col sm:flex-row gap-4 mb-6">
        <input 
          v-model="searchIP" 
          placeholder="Filter by IP..." 
          class="flex-1 bg-gray-700 px-4 py-2 rounded-lg focus:ring-2 focus:ring-blue-500"
        >
        <select v-model="filterSID" class="bg-gray-700 px-4 py-2 rounded-lg">
          <option value="">All SIDs</option>
          <option v-for="sid in uniqueSIDs" :key="sid" :value="sid">{{ sid }}</option>
        </select>
      </div>
      
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead>
            <tr class="border-b border-gray-700">
              <th class="p-4 text-left text-gray-400">Time</th>
              <th class="p-4 text-left text-gray-400">IP</th>
              <th class="p-4 text-left text-gray-400">SID</th>
              <th class="p-4 text-left text-gray-400">Message</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="alert in filteredAlerts" :key="alert.id" class="border-b border-gray-700 hover:bg-gray-700/50">
              <td class="p-4">{{ formatTime(alert.timestamp) }}</td>
              <td class="p-4 font-mono text-red-400">{{ alert.ip }}</td>
              <td class="p-4 font-mono text-blue-400">{{ alert.sid }}</td>
              <td class="p-4 text-sm max-w-md truncate">{{ alert.message }}</td>
            </tr>
            <tr v-if="filteredAlerts.length === 0">
              <td colspan="4" class="p-4 text-center text-gray-500 text-sm">
                No alerts for selected filters.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { Line as LineChart, Bar as BarChart } from 'vue-chartjs' 
import { Chart, LineElement, PointElement, LinearScale, CategoryScale, Tooltip, Legend, Filler, BarElement } from 'chart.js'  // ← USUŃ ArcElement

Chart.register(LineElement, PointElement, LinearScale, CategoryScale, Tooltip, Legend, Filler, BarElement)  // ← USUŃ ArcElement

const API_URL = window.location.origin
const stats = ref({ total_alerts: 0, total_blocked: 0, unique_ips: 0 })

const chartData = ref({
  hourly:     { labels: [], datasets: [] },
  topIPs:    { labels: [], datasets: [] },
  categories: { labels: [], datasets: [] },
  blocks:     { labels: [], datasets: [] }
})

const topIPsData = ref([])  // ← DODAJ to

const timeRange = ref('7')
const searchIP = ref('')
const filterSID = ref('')
const alerts = ref([])
const uniqueSIDs = ref([])

const filteredAlerts = computed(() => {
  return alerts.value
    .filter(alert => 
      alert.ip.includes(searchIP.value) &&
      (filterSID.value === '' || alert.sid == filterSID.value)
    )
    .slice(0, 100)
})

// Chart configs
const hourlyOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false }
  },
  scales: {
    x: { grid: { color: 'rgba(75,85,99,0.2)' } },
    y: { 
      beginAtZero: true,
      grid: { color: 'rgba(75,85,99,0.2)' },
      ticks: { color: '#9CA3AF' }
    }
  },
  elements: {
    point: { radius: 0 }
  }
}

const blocksOptions = { ...hourlyOptions, tension: 0.4 }

async function fetchData() {
  try {
    const [statsRes, alertBucketsRes, topRes, catsRes, blockBucketsRes] = await Promise.all([
      fetch(`${API_URL}/api/stats`),
      fetch(`${API_URL}/api/stats/alerts/buckets?days=${timeRange.value}`),
      fetch(`${API_URL}/api/stats/top_ips?limit=5`),  // ← ZMIEŃ na 5 (mniej)
      fetch(`${API_URL}/api/stats/categories?days=${timeRange.value}`),
      fetch(`${API_URL}/api/stats/blocks/buckets?days=${timeRange.value}`)
    ])

    const statsJson        = await statsRes.json()
    const alertBucketsJson = await alertBucketsRes.json()
    const topJson          = await topRes.json()
    const catsJson         = await catsRes.json()
    const blockBucketsJson = await blockBucketsRes.json()

    stats.value = statsJson

    chartData.value.hourly     = prepareBucketLineData(alertBucketsJson.data || [], 'Alerts')
    topIPsData.value           = (topJson.data || []).slice(0, 5)  
    chartData.value.categories = prepareBarData(catsJson.data || [])
    chartData.value.blocks     = prepareBucketLineData(blockBucketsJson.data || [], 'Blocks')
  } catch (error) {
    console.error('Failed to fetch analytics:', error)
  }
}

async function fetchAlerts() {
  try {
    const res = await fetch(`${API_URL}/api/stats/recent_alerts?limit=200`)
    const data = await res.json()
    alerts.value = data.alerts || []
    updateSIDs()
  } catch (error) {
    console.error('Failed to fetch recent alerts:', error)
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

const categoriesOptions = {
  indexAxis: 'y',
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false }
  },
  scales: {
    x: { 
      beginAtZero: true,
      grid: { color: 'rgba(75,85,99,0.2)' },
      ticks: { color: '#9CA3AF' }
    },
    y: { 
      grid: { display: false },
      ticks: { color: '#9CA3AF', font: { size: 11 } }
    }
  }
}

function updateSIDs() {
  const sids = [...new Set(alerts.value.map(a => a.sid))]
  uniqueSIDs.value = sids.sort((a, b) => b - a)
}

function formatTime(timestamp) {
  if (!timestamp) return 'N/A'
  return new Date(timestamp * 1000).toLocaleString()
}

onMounted(() => {
  fetchData()
  fetchAlerts()
})
</script>
