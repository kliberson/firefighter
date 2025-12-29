<script setup>
import { ref, onMounted, watch } from 'vue'
import { Line as LineChart } from 'vue-chartjs'
import { Chart, LineElement, PointElement, LinearScale, CategoryScale, Tooltip, Legend, Filler, BarElement } from 'chart.js'
import api from '@/services/api'
import { formatTimestamp, formatTimeAgo } from '@/utils/dateHelpers'
import { createMiniChartOptions } from '@/utils/chartOptions'

Chart.register(LineElement, PointElement, LinearScale, CategoryScale, Tooltip, Legend, Filler, BarElement)

const blockedIPs = ref([])
const liveAlerts = ref([])
const selectedBlock = ref(null)
const stats = ref({ total_alerts: 0, total_blocked: 0, unique_ips: 0 })
const miniChartData = ref({ labels: [], datasets: [] })
const miniChartOptions = createMiniChartOptions(miniChartData)

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
    stats.value = await api.getStats()
  } catch (error) {
    console.error('Failed to fetch stats:', error)
  }
}

async function fetchMiniChart() {
  try {
    const [alertsJson, blocksJson] = await Promise.all([
      api.getAlertBuckets(1),
      api.getBlockBuckets(1)
    ])

    const labels = (alertsJson.data || []).map(d => d.bucket)
    const alertsData = (alertsJson.data || []).map(d => d.count)

    const blocksMap = new Map(
      (blocksJson.data || []).map(d => [d.bucket, d.count])
    )
    const blocksData = labels.map(ts => blocksMap.get(ts) || 0)

    miniChartData.value = {
      labels,
      datasets: [
        {
          type: 'line',
          label: 'Alerts',
          data: alertsData,
          borderColor: '#3B82F6',
          backgroundColor: 'rgba(59,130,246,0.12)',
          fill: true,
          tension: 0.3,
          yAxisID: 'y'
        },
        {
          type: 'bar',
          label: 'Blocks',
          data: blocksData,
          backgroundColor: 'rgba(185,28,28,0.8)',
          borderColor: 'rgba(239,68,68,0.9)',
          borderWidth: 1.5,
          borderRadius: 3,
          barPercentage: 0.5,
          categoryPercentage: 0.8,
          yAxisID: 'y1'
        }
      ]
    }
  } catch (error) {
    console.error('Failed to fetch mini chart:', error)
  }
}

watch(() => props.alerts, (newAlerts) => {
  if (!newAlerts || newAlerts.length === 0) return
  
  newAlerts.forEach(event => {
    if (event.type === 'alert') {
      liveAlerts.value = [event, ...liveAlerts.value].slice(0, 50)
    } 
    else if (event.type === 'block') {
      blockedIPs.value = blockedIPs.value.filter(item => item.ip !== event.ip)
      
      blockedIPs.value.unshift({
        ip: event.ip,
        reason: event.reason,
        score: parseInt(event.score, 10) || 0,
        alert_count: parseInt(event.alert_count, 10) || 0,
        severity_score: parseInt(event.severity_score, 10) || 0,
        unique_ports: parseInt(event.unique_ports, 10) || 0,
        unique_protos: parseInt(event.unique_protos, 10) || 0,
        unique_flows: parseInt(event.unique_flows, 10) || 0,
        categories: event.categories || '',
        details: event.details || '',
        timestamp: event.timestamp,
      })
    } 
    else if (event.type === 'unblock') {
      blockedIPs.value = blockedIPs.value.filter(item => item.ip !== event.ip)
    }
  })
}, { immediate: true, deep: true })

async function fetchBlockedIPs() {
  try {
    const data = await api.getBlockedIPs()
    
    const freshIPs = data.blocked_ips || []
    const existingIPs = blockedIPs.value.map(b => b.ip)
    const newIPs = freshIPs.filter(b => !existingIPs.includes(b.ip))
    
    blockedIPs.value = [...blockedIPs.value, ...newIPs]
  } catch (error) {
    console.error('Failed to fetch blocked IPs:', error)
  }
}

async function unblockIP(ip) {
  if (!confirm(`Unblock ${ip}?`)) return
  
  try {
    const res = await api.unblockIP(ip)
    if (res.ok) {
      blockedIPs.value = blockedIPs.value.filter(item => item.ip !== ip)
    }
  } catch (error) {
    console.error('Failed to unblock IP:', error)
  }
}

function showDetails(block) {
  selectedBlock.value = block
}

function getSeverityClass(score) {
  if (score === null || score === undefined) return 'bg-gray-600'
  if (score >= 50) return 'bg-red-600'
  if (score >= 30) return 'bg-orange-500'
  return 'bg-yellow-500'
}

function getSeverityBadge(severity) {
  const sev = parseInt(severity, 10)
  if (sev === 1) return 'bg-red-600'
  if (sev === 2) return 'bg-orange-500'
  return 'bg-yellow-500'
}

function getSeverityText(severity) {
  const sev = parseInt(severity, 10)
  if (sev === 1) return 'HIGH'
  if (sev === 2) return 'MEDIUM'
  return 'LOW'
}
</script>

<template>
  <div class="p-8 space-y-6">
    <!-- Stats Cards -->
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

    <!-- Mini Chart (24h Activity) -->
    <div class="bg-gray-800 rounded-xl p-6 border border-gray-700">
      <h3 class="text-xl font-semibold mb-4">Activity (Last 24h)</h3>
      <div class="h-48">
        <LineChart
          v-if="miniChartData.labels && miniChartData.labels.length > 0"
          :data="miniChartData"
          :options="miniChartOptions"
        />
        <div v-else class="h-full flex items-center justify-center text-gray-500">
          No data available
        </div>
      </div>
    </div>

    <!-- Live Alerts & Blocked IPs Grid -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Live Alerts -->
      <div class="bg-gray-800 rounded-xl p-6 border border-gray-700">
        <h3 class="text-xl font-semibold mb-4">Live Alerts</h3>
        <div class="space-y-2 max-h-96 overflow-y-auto">
          <div 
            v-for="(alert, index) in liveAlerts.slice(0, 20)" 
            :key="index"
            class="flex items-center justify-between p-3 bg-gray-700/50 rounded-lg hover:bg-gray-700 transition-colors"
          >
            <div class="flex items-center gap-3 flex-1 min-w-0">
              <span 
                class="px-2 py-1 rounded text-xs font-semibold flex-shrink-0"
                :class="getSeverityBadge(alert.severity)"
              >
                {{ getSeverityText(alert.severity) }}
              </span>
              <span class="font-mono text-sm text-red-400 flex-shrink-0">{{ alert.ip }}</span>
              <span class="text-sm text-gray-300 truncate">{{ alert.reason }}</span>
            </div>
            <span class="text-xs text-gray-500 flex-shrink-0 ml-2">
              {{ formatTimeAgo(alert.timestamp) }}
            </span>
          </div>
          
          <div v-if="liveAlerts.length === 0" class="text-center py-8 text-gray-500">
            No alerts yet
          </div>
        </div>
      </div>

      <!-- Blocked IPs -->
      <div class="bg-gray-800 rounded-xl p-6 border border-gray-700">
        <h3 class="text-xl font-semibold mb-4">Blocked IPs</h3>
        <div class="space-y-2 max-h-96 overflow-y-auto">
          <div 
            v-for="block in blockedIPs" 
            :key="block.ip"
            class="flex items-center justify-between p-3 bg-gray-700/50 rounded-lg hover:bg-gray-700 transition-colors"
          >
            <div class="flex items-center gap-3 flex-1 min-w-0">
              <span 
                class="w-2 h-2 rounded-full flex-shrink-0"
                :class="getSeverityClass(block.score)"
              ></span>
              <span class="font-mono text-sm text-red-400 flex-shrink-0">{{ block.ip }}</span>
              <span class="text-sm text-gray-300 truncate">{{ block.reason }}</span>
            </div>
            <div class="flex items-center gap-2 flex-shrink-0 ml-2">
              <span class="text-xs bg-red-500/20 text-red-300 px-2 py-1 rounded">
                {{ block.score }}
              </span>
              <button 
                @click="showDetails(block)"
                class="text-xs text-blue-400 hover:text-blue-300"
              >
                Details
              </button>
              <button 
                @click="unblockIP(block.ip)"
                class="text-xs text-green-400 hover:text-green-300"
              >
                Unblock
              </button>
            </div>
          </div>

          <div v-if="blockedIPs.length === 0" class="text-center py-8 text-gray-500">
            No blocked IPs
          </div>
        </div>
      </div>
    </div>

    <!-- Details Modal -->
    <div 
      v-if="selectedBlock" 
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
      @click="selectedBlock = null"
    >
      <div 
        class="bg-gray-800 rounded-xl p-6 max-w-2xl w-full mx-4 border border-gray-700"
        @click.stop
      >
        <div class="flex justify-between items-start mb-4">
          <h3 class="text-xl font-semibold">Block Details</h3>
          <button 
            @click="selectedBlock = null"
            class="text-gray-400 hover:text-white text-2xl"
          >
            ×
          </button>
        </div>

        <div class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <p class="text-gray-400 text-sm">IP Address</p>
              <p class="text-lg font-mono text-red-400">{{ selectedBlock.ip }}</p>
            </div>
            <div>
              <p class="text-gray-400 text-sm">Score</p>
              <p class="text-lg font-bold" :class="getSeverityClass(selectedBlock.score)">
                {{ selectedBlock.score }}
              </p>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <p class="text-gray-400 text-sm">Alert Count</p>
              <p class="text-lg">{{ selectedBlock.alert_count }}</p>
            </div>
            <div>
              <p class="text-gray-400 text-sm">Severity Score</p>
              <p class="text-lg">{{ selectedBlock.severity_score }}</p>
            </div>
          </div>

          <div class="grid grid-cols-3 gap-4">
            <div>
              <p class="text-gray-400 text-sm">Unique Ports</p>
              <p class="text-lg">{{ selectedBlock.unique_ports }}</p>
            </div>
            <div>
              <p class="text-gray-400 text-sm">Unique Protocols</p>
              <p class="text-lg">{{ selectedBlock.unique_protos }}</p>
            </div>
            <div>
              <p class="text-gray-400 text-sm">Unique Flows</p>
              <p class="text-lg">{{ selectedBlock.unique_flows }}</p>
            </div>
          </div>

          <div>
            <p class="text-gray-400 text-sm mb-1">Reason</p>
            <p class="text-sm bg-gray-700 p-3 rounded">{{ selectedBlock.reason }}</p>
          </div>

          <div>
            <p class="text-gray-400 text-sm mb-1">Categories</p>
            <p class="text-sm bg-gray-700 p-3 rounded font-mono">{{ selectedBlock.categories }}</p>
          </div>

          <div>
            <p class="text-gray-400 text-sm mb-1">Details</p>
            <p class="text-xs bg-gray-700 p-3 rounded font-mono text-gray-300">
              {{ selectedBlock.details }}
            </p>
          </div>

          <div>
            <p class="text-gray-400 text-sm">Blocked At</p>
            <p class="text-sm">{{ formatTimestamp(selectedBlock.timestamp) }}</p>
          </div>
        </div>

        <div class="mt-6 flex gap-3">
          <button 
            @click="unblockIP(selectedBlock.ip); selectedBlock = null"
            class="flex-1 bg-green-600 px-4 py-2 rounded hover:bg-green-700"
          >
            Unblock IP
          </button>
          <button 
            @click="selectedBlock = null"
            class="flex-1 bg-gray-600 px-4 py-2 rounded hover:bg-gray-700"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

