<template>
  <div class="p-8">
    <h1 class="text-4xl font-bold mb-8">Firefighter Dashboard</h1>
    
    <!-- Main Content: Blocked IPs -->
    <div class="bg-gray-800 rounded-lg p-6 mb-8">
      <div class="flex justify-between items-center mb-6">
        <h2 class="text-2xl font-bold">Blocked IP Addresses</h2>
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
    
    <!-- Live Alerts Feed -->
    <div class="bg-gray-800 rounded-lg p-6">
      <h2 class="text-2xl font-bold mb-6">Live Alerts</h2>
      <div class="space-y-3 max-h-96 overflow-y-auto">
        <div 
          v-for="(alert, index) in liveAlerts.slice(0, 20)" 
          :key="index"
          class="bg-gray-700 p-3 rounded hover:bg-gray-600 transition"
        >
          <div class="flex justify-between items-center mb-2">
            <span class="font-mono text-yellow-400">{{ alert.ip }}</span>
            <span 
              v-if="alert.severity" 
              :class="getSeverityBadge(alert.severity)"
              class="px-2 py-1 rounded text-xs font-bold"
            >
              {{ getSeverityText(alert.severity) }}
            </span>
            <span class="text-gray-400 text-xs">{{ formatTime(alert.timestamp) }}</span>
          </div>
          
          <span class="text-gray-300 text-sm mb-2">{{ alert.reason }}</span>
          
          <div v-if="alert.sid || alert.protocol" class="flex gap-3 text-xs mt-2">
            <span v-if="alert.sid" class="text-gray-500">SID: {{ alert.sid }}</span>
            <span v-if="alert.protocol" class="text-blue-400">{{ alert.protocol }}</span>
            <span v-if="alert.dst_port" class="text-purple-400">Port: {{ alert.dst_port }}</span>
            <span v-if="alert.category" class="text-orange-400">{{ alert.category }}</span>
          </div>
        </div>
        <div v-if="liveAlerts.length === 0" class="text-gray-500 text-center py-8">
          Waiting for alerts...
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'

const API_URL = window.location.origin
const blockedIPs = ref([])
const liveAlerts = ref([])
const selectedBlock = ref(null)

const props = defineProps({
  connected: Boolean,
  alerts: Array
})

onMounted(() => {
  fetchBlockedIPs()
})

watch(() => props.alerts, (newAlerts) => {
  if (!newAlerts || newAlerts.length === 0) return
  
  const alerts = []
  const blocks = []
  
  newAlerts.forEach(event => {
    if (event.type === 'alert') {
      alerts.push(event)
    } else if (event.type === 'block') {
      blocks.push(event)
    }
  })
  
  liveAlerts.value = alerts.slice(0, 50)
  
 blocks.forEach(block => {
  const exists = blockedIPs.value.find(ip => ip.ip === block.ip)
  if (!exists) {
    blockedIPs.value.unshift({
      ip: block.ip,
      reason: block.reason,
      score: block.score,
      alert_count: block.alert_count || 0,          // ← TUTAJ
      severity_score: block.severity_score || 0,    // ← I TUTAJ
      unique_ports: block.unique_ports || 0,
      unique_protos: block.unique_protos || 0,
      unique_flows: block.unique_flows || 0,
      categories: block.categories || '',
      details: block.details,
      timestamp: block.timestamp
    })
  }
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
