<template>
  <div class="p-8">
    <h1 class="text-4xl font-bold mb-8">Firefighter Dashboard</h1>
    
    <!-- Stats Row -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
      <div class="bg-gray-800 rounded-lg p-6 border-l-4 border-red-500">
        <p class="text-gray-400 text-sm">Blocked IPs</p>
        <p class="text-4xl font-bold mt-2">{{ blockedIPs.length }}</p>
      </div>
      
      <div class="bg-gray-800 rounded-lg p-6 border-l-4 border-yellow-500">
        <p class="text-gray-400 text-sm">Live Alerts</p>
        <p class="text-4xl font-bold mt-2">{{ alerts.length }}</p>
      </div>
      
      <div class="bg-gray-800 rounded-lg p-6 border-l-4 border-green-500">
        <p class="text-gray-400 text-sm">
          <span v-if="connected" class="text-green-400">● Online</span>
          <span v-else class="text-red-400">● Offline</span>
        </p>
        <p class="text-xl font-bold mt-2">WebSocket Status</p>
      </div>
    </div>
    
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
              <th class="text-left p-3 text-gray-400 font-medium">Timestamp</th>
              <th class="text-left p-3 text-gray-400 font-medium">IP Address</th>
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
              <td class="p-3 text-yellow-300 text-sm">{{ item.reason }}</td>
              <td class="p-3">
                <button 
                  @click="unblockIP(item.ip)"
                  class="bg-red-600 px-4 py-1 rounded text-sm hover:bg-red-700 transition"
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
    
    <!-- Live Alerts Feed -->
    <div class="bg-gray-800 rounded-lg p-6">
      <h2 class="text-2xl font-bold mb-6">Live Alerts Feed</h2>
      <div class="space-y-3 max-h-96 overflow-y-auto">
        <div 
          v-for="(alert, index) in alerts.slice(0, 20)" 
          :key="index"
          class="bg-gray-700 p-4 rounded hover:bg-gray-600 transition"
        >
          <div class="flex justify-between items-center mb-2">
            <span class="font-mono text-red-400">{{ alert.ip }}</span>
            <span class="text-gray-400 text-xs">{{ formatTime(alert.timestamp) }}</span>
          </div>
          <div class="text-yellow-300 text-sm">{{ alert.reason }}</div>
        </div>
        <div v-if="alerts.length === 0" class="text-gray-500 text-center py-8">
          Waiting for alerts...
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const API_URL = window.location.origin
const blockedIPs = ref([])

const props = defineProps({
  connected: Boolean,
  alerts: Array
})

onMounted(() => {
  fetchBlockedIPs()
})

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

function formatTime(timestamp) {
  return new Date(timestamp).toLocaleTimeString()
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
