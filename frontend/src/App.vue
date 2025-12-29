<script setup>
import { watch } from 'vue'
import notificationService from '@/services/notificationService'
import { useWebSocket } from '@/services/websocket'

const { connected, alerts } = useWebSocket()

watch(alerts, (newAlerts) => {
  newAlerts.forEach(event => {
    switch(event.type) {
      case 'block':
        notificationService.blocked(
          event.ip,
          event.score,
          event.reason,
          event.alert_count
        )
        break
        
      case 'unblock':
        notificationService.unblocked(event.ip)
        break
    }
  })
}, { deep: true })
</script>

<template>
  <div class="flex h-screen bg-gray-900 text-white">
    <aside class="w-45 bg-gray-800 border-r border-gray-700 flex flex-col">
      <div class="p-6 border-b border-gray-700">
        <h1 class="text-2xl font-bold">Firefighter</h1>
        <p class="text-sm text-gray-400 mt-1">Admin Panel</p>
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
          to="/activity" 
          class="flex items-center gap-3 px-4 py-3 rounded-lg transition-colors"
          :class="$route.path === '/activity' ? 'bg-blue-600' : 'hover:bg-gray-700'"
        >
          <span class="text-xl"></span>
          <span class="font-medium">Activity</span>
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
    </aside>

    <main class="flex-1 overflow-y-auto">
      <router-view :connected="connected" :alerts="alerts" />
    </main>
  </div>
</template>
