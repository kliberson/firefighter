<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold">Whitelist Management</h1>
      <button 
        @click="showAddModal = true"
        class="bg-green-600 px-4 py-2 rounded hover:bg-green-700"
      >
        + Add IP
      </button>
    </div>
    
    <!-- Add IP Modal -->
    <div v-if="showAddModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
      <div class="bg-gray-800 rounded-lg p-6 w-96">
        <h2 class="text-xl font-bold mb-4">Add to Whitelist</h2>
        
        <input 
          v-model="newIP" 
          type="text" 
          placeholder="192.168.1.100"
          class="w-full bg-gray-700 px-4 py-2 rounded mb-4"
        />
        
        <textarea 
          v-model="newDescription" 
          placeholder="Description (optional)"
          class="w-full bg-gray-700 px-4 py-2 rounded mb-4 h-24"
        />
        
        <div class="flex gap-2">
          <button 
            @click="addToWhitelist"
            class="flex-1 bg-green-600 px-4 py-2 rounded hover:bg-green-700"
          >
            Add
          </button>
          <button 
            @click="showAddModal = false"
            class="flex-1 bg-gray-600 px-4 py-2 rounded hover:bg-gray-700"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
    
    <!-- Whitelist Table -->
    <div class="bg-gray-800 rounded-lg p-6">
      <table class="w-full">
        <thead>
          <tr class="border-b border-gray-700">
            <th class="text-left p-2">IP Address</th>
            <th class="text-left p-2">Description</th>
            <th class="text-left p-2">Added</th>
            <th class="text-left p-2">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr 
            v-for="item in whitelistedIPs" 
            :key="item.ip"
            class="border-b border-gray-700"
          >
            <td class="p-2">{{ item.ip }}</td>
            <td class="p-2 text-gray-400">{{ item.description || 'N/A' }}</td>
            <td class="p-2 text-gray-400">{{ formatTimestamp(item.added_at) }}</td>
            <td class="p-2">
              <button 
                @click="removeFromWhitelist(item.ip)"
                class="bg-red-600 px-3 py-1 rounded text-sm hover:bg-red-700"
              >
                Remove
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const API_URL = window.location.origin
const whitelistedIPs = ref([])
const showAddModal = ref(false)
const newIP = ref('')
const newDescription = ref('')

onMounted(() => {
  fetchWhitelist()
})

async function fetchWhitelist() {
  try {
    const res = await fetch(`${API_URL}/api/whitelist`)
    const data = await res.json()
    whitelistedIPs.value = data.whitelisted_ips || []
  } catch (error) {
    console.error('Failed to fetch whitelist:', error)
  }
}

async function addToWhitelist() {
  try {
    const res = await fetch(`${API_URL}/api/whitelist/${newIP.value}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ 
        ip: newIP.value,
        description: newDescription.value 
      })
    })
    
    if (res.ok) {
      showAddModal.value = false
      newIP.value = ''
      newDescription.value = ''
      fetchWhitelist()
    }
  } catch (error) {
    console.error('Failed to add to whitelist:', error)
  }
}

async function removeFromWhitelist(ip) {
  if (!confirm(`Remove ${ip} from whitelist?`)) return
  
  try {
    const res = await fetch(`${API_URL}/api/whitelist/${ip}`, {
      method: 'DELETE'
    })
    
    if (res.ok) {
      fetchWhitelist()
    }
  } catch (error) {
    console.error('Failed to remove from whitelist:', error)
  }
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
 