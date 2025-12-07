<template>
  <div class="flex h-screen bg-gray-900 text-white">
    <!-- Sidebar -->
    <aside class="w-45 bg-gray-800 border-r border-gray-700 flex flex-col">
      <div class="p-6 border-b border-gray-700">
        <h1 class="text-2xl font-bold">Firefighter</h1>
        <p class="text-sm text-gray-400 mt-1">IPS/IDS Dashboard</p>
      </div>
      
      <nav class="flex-1 p-4 space-y-2">
        <router-link 
          to="/" 
          class="flex items-center gap-3 px-4 py-3 rounded-lg transition-colors"
          :class="$route.path === '/' ? 'bg-blue-600' : 'hover:bg-gray-700'"
        >
          <span class="text-xl"></span>
          <span class="font-medium">Dashboard</span>
        </router-link>

        <router-link 
          to="/analytics" 
          class="flex items-center gap-3 px-4 py-3 rounded-lg transition-colors"
          :class="$route.path === '/analytics' ? 'bg-blue-600' : 'hover:bg-gray-700'"
        >
          <span class="text-xl"></span>
          <span class="font-medium">Analytics</span>
        </router-link>

        <router-link 
          to="/whitelist" 
          class="flex items-center gap-3 px-4 py-3 rounded-lg transition-colors"
          :class="$route.path === '/whitelist' ? 'bg-blue-600' : 'hover:bg-gray-700'"
        >
          <span class="text-xl"></span>
          <span class="font-medium">Whitelist</span>
        </router-link>
      </nav>
      
      <!-- Connection Status -->
      <div class="p-4 border-t border-gray-700">
        <div class="flex items-center gap-2">
          <div :class="connected ? 'bg-green-500' : 'bg-red-500'" class="w-2 h-2 rounded-full"></div>
          <span class="text-sm text-gray-400">
            {{ connected ? 'Connected' : 'Disconnected' }}
          </span>
        </div>
      </div>
    </aside>
    
    <!-- Main content -->
    <main class="flex-1 overflow-y-auto">
      <router-view :connected="connected" :alerts="alerts" />
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const connected = ref(false)
const alerts = ref([])
let ws = null

onMounted(() => {
  connectWebSocket()
})

onUnmounted(() => {
  if (ws) ws.close()
})

function connectWebSocket() {
  const WS_URL = `ws://${window.location.host}/ws`
  ws = new WebSocket(WS_URL)
  
  ws.onopen = () => {
    connected.value = true
  }
  
  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      alerts.value.unshift({
        timestamp: data.timestamp,
        ip: data.ip,
        reason: data.reason,
        type: data.type
      })
      
      if (alerts.value.length > 100) {
        alerts.value.pop()
      }
    } catch (error) {
      console.error('WebSocket parse error:', error)
    }
  }
  
  ws.onclose = () => {
    connected.value = false
    setTimeout(connectWebSocket, 5000)
  }
}
</script>
