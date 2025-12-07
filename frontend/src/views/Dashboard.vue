<template>
  <div class="p-8">
    <h1 class="text-3xl font-semibold mb-6">Dashboard</h1>
    
    <!-- Compact Stats Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
      <div class="bg-gradient-to-r from-red-500/20 to-red-600/30 border border-red-500/30 p-4 rounded-xl">
        <div class="text-2xl font-bold text-red-400 mb-1">{{ stats.total_alerts }}</div>
        <div class="text-sm text-gray-400">Total Alerts</div>
      </div>
      <div class="bg-gradient-to-r from-orange-500/20 to-orange-600/30 border border-orange-500/30 p-4 rounded-xl">
        <div class="text-2xl font-bold text-orange-400 mb-1">{{ stats.total_blocked }}</div>
        <div class="text-sm text-gray-400">Blocked IPs</div>
      </div>
      <div class="bg-gradient-to-r from-blue-500/20 to-blue-600/30 border border-blue-500/30 p-4 rounded-xl">
        <div class="text-2xl font-bold text-blue-400 mb-1">{{ stats.unique_ips }}</div>
        <div class="text-sm text-gray-400">Unique IPs</div>
      </div>
      <div class="bg-gradient-to-r from-purple-500/20 to-purple-600/30 border border-purple-500/30 p-4 rounded-xl">
        <div class="text-2xl font-bold text-purple-400 mb-1">{{ stats.total_alerts > 0 ? Math.round(stats.total_blocked / stats.total_alerts * 100) : 0 }}%</div>
        <div class="text-sm text-gray-400">Block Rate</div>
      </div>
    </div>

    <!-- Alerts Last 24h Chart -->
    <div class="bg-gray-800 rounded-lg p-6 mb-4">
      <h3 class="text-xl font-semibold mb-4">Alerts Last 24h</h3>
      <div class="h-32">
        <LineChart 
          v-if="miniChartData && miniChartData.labels" 
          :data="miniChartData" 
          :options="miniChartOptions" 
        />
        <div v-else class="text-gray-500 text-sm">Loading...</div>
      </div>
    </div>

    <!-- Live Alerts Feed -->
    <div class="bg-gray-800 rounded-lg p-6 mb-6">
    <h2 class="text-xl font-semibold mb-3">Live Alerts</h2>
    <div class="space-y-2 max-h-60 overflow-y-auto"> <!-- niższe, ciaśniejsze -->
      <div 
        v-for="(alert, index) in liveAlerts.slice(0, 20)" 
        :key="index"
        class="bg-gray-700/70 px-3 py-2 rounded flex items-center justify-between text-xs"
      >
        <div class="flex flex-col">
          <span class="font-mono text-yellow-400">{{ alert.ip }}</span>
          <span class="text-gray-300 truncate max-w-xs">{{ alert.reason }}</span>
          <div class="flex gap-2 mt-1 text-[10px] text-gray-400">
            <span v-if="alert.sid">SID: {{ alert.sid }}</span>
            <span v-if="alert.protocol">{{ alert.protocol }}</span>
            <span v-if="alert.dst_port">Port: {{ alert.dst_port }}</span>
            <span v-if="alert.category">{{ alert.category }}</span>
          </div>
        </div>
        <div class="flex flex-col items-end gap-1">
          <span 
            v-if="alert.severity" 
            :class="getSeverityBadge(alert.severity)"
            class="px-2 py-0.5 rounded text-[10px] font-bold"
          >
            {{ getSeverityText(alert.severity) }}
          </span>
          <span class="text-gray-400 text-[10px]">{{ formatTime(alert.timestamp) }}</span>
        </div>
      </div>
      <div v-if="liveAlerts.length === 0" class="text-gray-500 text-center py-4 text-sm">
        Waiting for alerts...
      </div>
    </div>
  </div>

    <!-- Main Content: Blocked IPs -->
    <div class="bg-gray-800 rounded-lg p-6 mb-8">
      <div class="flex justify-between items-center mb-6">
        <h2 class="text-2xl font-semibold">Blocked IP Addresses</h2>
        <button 
          @click="fetchBlockedIPs" 
          class="bg-blue-600 px-4 py-2 rounded hover:bg-blue-700 transition"
        >
          Refresh
        </button>
      </div>
      
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead>
            <tr class="border-b border-gray-700">
              <th class="text-left p-3 text-gray-400 font-medium">Time</th>
              <th class="text-left p-3 text-gray-400 font-medium">IP Address</th>
              <th class="text-left p-3 text-gray-400 font-medium">Score</th>
              <th class="text-left p-3 text-gray-400 font-medium">Alerts</th>
              <th class="text-left p-3 text-gray-400 font-medium">Reason</th>
              <th class="text-left p-3 text-gray-400 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr 
              v-for="item in blockedIPs" 
              :key="item.ip"
              class="border-b border-gray-700 hover:bg-gray-700/50 transition"
            >
              <td class="p-3 text-gray-400 text-sm">{{ formatTimestamp(item.timestamp) }}</td>
              <td class="p-3 font-mono text-red-400">{{ item.ip }}</td>
              <td class="p-3">
                <span 
                  :class="getSeverityClass(item.score)" 
                  class="px-3 py-1 rounded text-sm font-bold"
                >
                  {{ item.score || 'N/A' }}
                </span>
              </td>
              <td class="p-3 text-gray-300">{{ item.alert_count || 'N/A' }}</td>
              <td class="p-3 text-yellow-300 text-sm">{{ item.reason }}</td>
              <td class="p-3">
                <button 
                  @click="showDetails(item)"
                  class="bg-blue-600 px-3 py-1 rounded text-xs hover:bg-blue-700 mr-2"
                >
                  Details
                </button>
                <button 
                  @click="unblockIP(item.ip)"
                  class="bg-red-600 px-3 py-1 rounded text-xs hover:bg-red-700"
                >
                  Unblock
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="blockedIPs.length === 0" class="text-gray-500 text-center py-8">
          No blocked IPs
        </div>
      </div>
    </div>
    
    <!-- Details Modal -->
    <div v-if="selectedBlock" class="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50">
      <div class="bg-gray-800 rounded-lg p-6 w-2/3 max-w-3xl max-h-[80vh] overflow-y-auto">
        <h2 class="text-2xl font-bold mb-4">Block Details: {{ selectedBlock.ip }}</h2>
        
        <div class="grid grid-cols-2 gap-4 mb-4">
          <div class="bg-gray-700 p-4 rounded">
            <p class="text-xs text-gray-400">Final Score</p>
            <p class="text-3xl font-bold text-red-400">{{ selectedBlock.score }}</p>
          </div>
          <div class="bg-gray-700 p-4 rounded">
            <p class="text-xs text-gray-400">Alert Count</p>
            <p class="text-3xl font-bold text-orange-400">{{ selectedBlock.alert_count }}</p>
          </div>
          <div class="bg-gray-700 p-4 rounded">
            <p class="text-xs text-gray-400">Severity Score</p>
            <p class="text-2xl font-bold text-yellow-400">{{ selectedBlock.severity_score }}</p>
          </div>
          <div class="bg-gray-700 p-4 rounded">
            <p class="text-xs text-gray-400">Unique Ports</p>
            <p class="text-2xl font-bold text-purple-400">{{ selectedBlock.unique_ports }}</p>
          </div>
          <div class="bg-gray-700 p-4 rounded">
            <p class="text-xs text-gray-400">Protocols</p>
            <p class="text-2xl font-bold text-blue-400">{{ selectedBlock.unique_protos }}</p>
          </div>
          <div class="bg-gray-700 p-4 rounded">
            <p class="text-xs text-gray-400">Flow Sessions</p>
            <p class="text-2xl font-bold text-green-400">{{ selectedBlock.unique_flows }}</p>
          </div>
        </div>
        
        <div class="bg-gray-900 p-4 rounded mb-4">
          <p class="text-sm text-gray-400 mb-2">Attack Categories:</p>
          <p class="text-green-400 font-mono text-sm">{{ selectedBlock.categories || 'N/A' }}</p>
        </div>
        
        <div class="bg-gray-900 p-4 rounded mb-4">
          <p class="text-sm text-gray-400 mb-2">Full Details:</p>
          <pre class="text-xs text-green-400 overflow-x-auto">{{ selectedBlock.details || 'N/A' }}</pre>
        </div>
        
        <button 
          @click="selectedBlock = null"
          class="w-full bg-gray-600 px-4 py-2 rounded hover:bg-gray-700"
        >
          Close
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { Line as LineChart } from 'vue-chartjs'
import { Chart, LineElement, PointElement, LinearScale, CategoryScale, Tooltip, Legend, Filler } from 'chart.js'

Chart.register(LineElement, PointElement, LinearScale, CategoryScale, Tooltip, Legend, Filler)

const API_URL = window.location.origin
const blockedIPs = ref([])
const liveAlerts = ref([])
const selectedBlock = ref(null)
const stats = ref({ total_alerts: 0, total_blocked: 0, unique_ips: 0 })
const miniChartData = ref({ labels: [], datasets: [] })

const miniChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false }
  },
  scales: {
    x: {
      grid: { color: 'rgba(75,85,99,0.2)' },
      ticks: {
        color: '#9CA3AF',
        maxRotation: 0,
        minRotation: 0,
        autoSkip: true,
        callback: (value) => {
          const raw = miniChartData.value.labels[value] || ''
          const parts = String(raw).split(' ')
          return parts.length > 1 ? parts[1] : raw
        }
      }
    }
  },
  elements: {
    point: { radius: 0 }
  }
}

const props = defineProps({
  connected: Boolean,
  alerts: Array
})

onMounted(() => {
  fetchBlockedIPs()
  fetchStats()
  fetchMiniChart()
})

async function fetchStats() {
  try {
    const res = await fetch(`${API_URL}/api/stats`)
    stats.value = await res.json()
  } catch (error) {
    console.error('Failed to fetch stats:', error)
  }
}

async function fetchMiniChart() {
  try {
    const res = await fetch(`${API_URL}/api/stats/alerts/buckets?days=1`)
    const data = await res.json()
    
    miniChartData.value = {
      labels: (data.data || []).map(d => d.bucket),
      datasets: [{
        label: 'Alerts',
        data: (data.data || []).map(d => d.count),
        borderColor: '#3B82F6',
        backgroundColor: 'rgba(59, 130, 246, 0.1)',
        fill: true,
        tension: 0.3
      }]
    }
  } catch (error) {
    console.error('Failed to fetch mini chart:', error)
  }
}

watch(() => props.alerts, (newAlerts) => {
  if (!newAlerts || newAlerts.length === 0) return
  
  const alerts = []
  const blocks = []
  
  newAlerts.forEach(event => {
    // Backend wysyła "alert", "block", "unblock" w polu type
    if (event.type === 'alert') {
      alerts.push(event)
    } else if (event.type === 'block') {
      blocks.push(event)
    } else if (event.type === 'unblock') {
      // Obsługa live unblock (opcjonalnie - usuwanie z listy)
      blockedIPs.value = blockedIPs.value.filter(item => item.ip !== event.ip)
    }
  })
  
  // Aktualizacja Live Alerts
  if (alerts.length > 0) {
    // Dodajemy nowe na początek i przycinamy do 50
    liveAlerts.value = [...alerts, ...liveAlerts.value].slice(0, 50)
  }
  
  blocks.forEach(block => {
    // Usuń stary wpis (jeśli istnieje)
    blockedIPs.value = blockedIPs.value.filter(item => item.ip !== block.ip)
    
    // Dodaj nowy na początku
    blockedIPs.value.unshift({
      ip: block.ip,
      reason: block.reason,
      score: block.score,
      alert_count: block.alert_count,
      severity_score: block.severity_score,
      unique_ports: block.unique_ports,
      unique_protos: block.unique_protos,
      unique_flows: block.unique_flows,
      categories: block.categories,
      details: block.details,
      timestamp: block.timestamp,
  })
})
}, { immediate: true, deep: true })

async function fetchBlockedIPs() {
  try {
    const res = await fetch(`${API_URL}/api/blocked`)
    const data = await res.json()
    blockedIPs.value = data.blocked_ips || []
  } catch (error) {
    console.error('Failed to fetch blocked IPs:', error)
  }
}

async function unblockIP(ip) {
  if (!confirm(`Unblock ${ip}?`)) return
  
  try {
    const res = await fetch(`${API_URL}/api/unblock/${ip}`, {
      method: 'POST'
    })
    
    if (res.ok) {
      fetchBlockedIPs()
    }
  } catch (error) {
    console.error('Failed to unblock IP:', error)
  }
}

function showDetails(block) {
  selectedBlock.value = block
}

function getSeverityClass(score) {
  if (!score) return 'bg-gray-600'
  if (score >= 50) return 'bg-red-600'
  if (score >= 30) return 'bg-orange-500'
  return 'bg-yellow-500'
}

function getSeverityBadge(severity) {
  if (severity === 1) return 'bg-red-600'
  if (severity === 2) return 'bg-orange-500'
  return 'bg-yellow-500'
}

function getSeverityText(severity) {
  if (severity === 1) return 'HIGH'
  if (severity === 2) return 'MEDIUM'
  return 'LOW'
}

function formatTime(timestamp) {
  if (!timestamp) return 'N/A'
  const date = new Date(timestamp * 1000)
  const now = Date.now()
  const diff = Math.floor((now - date.getTime()) / 1000)
  
  if (diff < 60) return 'just now'
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return date.toLocaleTimeString()
}

function formatTimestamp(unixTimestamp) {
  if (!unixTimestamp) return 'N/A'
  const date = new Date(unixTimestamp * 1000)
  return date.toLocaleString('en-US', {
    month: '2-digit',
    day: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}
</script>

<style scoped>
.chart-container {
  position: relative;
  height: 250px;
  width: 100%;
}
</style>
