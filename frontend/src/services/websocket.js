import { ref, onMounted, onUnmounted } from 'vue'

export function useWebSocket() {
  const connected = ref(false)
  const alerts = ref([])
  let ws = null

  function connect() {
    const WS_URL = `ws://${window.location.host}/ws`
    ws = new WebSocket(WS_URL)
    
    ws.onopen = () => {
      connected.value = true
    }
    
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        alerts.value.unshift(data)
        
        if (alerts.value.length > 100) {
          alerts.value.pop()
        }
      } catch (error) {
        console.error('WebSocket parse error:', error)
      }
    }
    
    ws.onclose = () => {
      connected.value = false
      setTimeout(connect, 5000)
    }
  }

  function disconnect() {
    if (ws) ws.close()
  }

  onMounted(() => {
    connect()
  })

  onUnmounted(() => {
    disconnect()
  })

  return {
    connected,
    alerts
  }
}
