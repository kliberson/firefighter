<template>
  <div class="p-8 space-y-6">
    <div class="flex justify-between items-center">
      <h1 class="text-3xl font-semibold">Activity Log</h1>
      <button @click="fetchActivity" class="bg-blue-600 px-6 py-2 rounded hover:bg-blue-700">
        Refresh
      </button>
    </div>

    <!-- Search Bar -->
    <div class="bg-gray-800 border border-gray-700 rounded-xl p-6">
        <div class="flex gap-3 ">
            <select 
                v-model="typeFilter" 
                @change="fetchActivity"
                class="bg-gray-700 px-4 py-2 rounded-lg focus:ring-2 focus:ring-blue-500"
                >
                <option value="">All Types</option>
                <option value="alert">Alerts</option>
                <option value="block">Blocks</option>
                <option value="unblock">Unblocks</option>
                <option value="whitelist_add">Whitelist Add</option>
                <option value="whitelist_remove">Whitelist Remove</option>
            </select>

            <input
                v-model="searchQuery"
                placeholder="Search by IP, message, or keyword..."
                class="flex-1 bg-gray-700 px-4 py-2 rounded-lg focus:ring-2 focus:ring-blue-500"
                @keyup.enter="fetchActivity"
                >
            <button
                @click="fetchActivity"
                class="bg-blue-600 px-5 py-2 rounded hover:bg-blue-700"
                >
                Search
            </button>
            <button
                v-if="searchQuery || typeFilter"
                @click="clearSearch"
                class="bg-gray-600 px-5 py-2 rounded hover:bg-gray-700"
            >
                Clear
            </button>
        </div>
    </div>

    <!-- Activity Table with Fixed Height -->
    <div class="bg-gray-800 rounded-xl p-6 border border-gray-700 rounded-xl p-6 ">
      <div v-if="activity.length" class="overflow-x-auto">
        <div class="max-h-[600px] overflow-y-auto">
          <table class="w-full text-sm">
            <thead class="sticky top-0 bg-gray-800">
              <tr class="border-b border-gray-700">
                <th class="p-3 text-left text-gray-400">Time</th>
                <th class="p-3 text-left text-gray-400">Type</th>
                <th class="p-3 text-left text-gray-400">IP Address</th>
                <th class="p-3 text-left text-gray-400">Details</th>
                <th class="p-3 text-left text-gray-400">Extra</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="entry in activity"
                :key="`${entry.type}-${entry.timestamp}-${entry.ip}`"
                class="border-b border-gray-700 hover:bg-gray-700/50"
              >
                <td class="p-3 text-gray-400">{{ formatTimestamp(entry.timestamp) }}</td>
                <td class="p-3">
                  <span
                    :class="getTypeClass(entry.type)"
                    class="px-2 py-1 rounded text-xs font-semibold uppercase"
                  >
                    {{ getTypeLabel(entry.type) }}
                  </span>
                </td>
                <td class="p-3 font-mono text-red-400">{{ entry.ip }}</td>
                <td class="p-3 text-gray-200 truncate max-w-md">{{ entry.details }}</td>
                <td class="p-3 font-mono text-blue-400">{{ entry.extra }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div v-else class="text-gray-500 text-sm text-center py-8">
        <p v-if="searchQuery">No results found for "{{ searchQuery }}"</p>
        <p v-else>No activity recorded yet.</p>
      </div>
    </div>
  </div>
</template>


<script setup>
import { ref, onMounted } from 'vue'
import api from '@/services/api'
import { formatTimestamp } from '@/utils/dateHelpers'

const activity = ref([])
const searchQuery = ref('')
const typeFilter = ref('')

async function fetchActivity() {
  try {
    const data = await api.getActivity(searchQuery.value, typeFilter.value) 
    activity.value = data.activity || []
  } catch (error) {
    console.error('Failed to fetch activity:', error)
  }
}

function clearSearch() {
  searchQuery.value = ''
  typeFilter.value = ''
  fetchActivity()
}

function getTypeClass(type) {
  switch(type) {
    case 'alert': return 'bg-yellow-500/20 text-yellow-300'
    case 'block': return 'bg-red-500/20 text-red-300'
    case 'unblock': return 'bg-green-500/20 text-green-300'
    case 'whitelist_add': return 'bg-blue-500/20 text-blue-300'
    case 'whitelist_remove': return 'bg-gray-500/20 text-gray-300'
    default: return 'bg-gray-500/20 text-gray-300'
  }
}

function getTypeLabel(type) {
  switch(type) {
    case 'alert': return 'Alert'
    case 'block': return 'Block'
    case 'unblock': return 'Unblock'
    case 'whitelist_add': return 'Whitelist+'
    case 'whitelist_remove': return 'Whitelist-'
    default: return type
  }
}

onMounted(() => {
  fetchActivity()
})
</script>
