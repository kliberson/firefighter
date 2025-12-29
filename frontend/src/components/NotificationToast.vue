<template>
  <div class="flex items-start gap-3">
    <!-- Icon -->
    <div class="flex-shrink-0">
      <div 
        class="w-10 h-10 rounded-full flex items-center justify-center"
        :class="iconBgClass"
      >
        <component :is="iconComponent" class="w-6 h-6" :class="iconColorClass" />
      </div>
    </div>
    
    <!-- Content -->
    <div class="flex-1 pt-1">
      <h4 class="text-sm font-semibold text-white mb-1">{{ title }}</h4>
      <p class="text-xs text-gray-300">{{ message }}</p>
      
      <!-- IP Details (optional) -->
      <div v-if="ip" class="mt-2 p-2 bg-black/30 rounded text-xs font-mono">
        <div class="flex items-center justify-between">
          <span class="text-red-400">{{ ip }}</span>
          <span v-if="score" class="text-yellow-400">Score: {{ score }}</span>
        </div>
        <div v-if="reason" class="text-gray-400 mt-1 truncate">
          {{ reason }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

// SVG Icons as components
import { 
  ShieldExclamationIcon, 
  ExclamationTriangleIcon, 
  InformationCircleIcon,
  CheckCircleIcon 
} from '@heroicons/vue/24/solid'

const props = defineProps({
  type: {
    type: String,
    default: 'info', // 'block', 'alert', 'unblock', 'info', 'error', 'success'
    validator: (value) => ['block', 'alert', 'unblock', 'info', 'error', 'success'].includes(value)
  },
  title: {
    type: String,
    required: true
  },
  message: {
    type: String,
    default: ''
  },
  ip: {
    type: String,
    default: null
  },
  score: {
    type: [Number, String],
    default: null
  },
  reason: {
    type: String,
    default: null
  }
})

const iconComponent = computed(() => {
  const icons = {
    block: ShieldExclamationIcon,
    alert: ExclamationTriangleIcon,
    unblock: CheckCircleIcon,
    success: CheckCircleIcon,
    error: ShieldExclamationIcon,
    info: InformationCircleIcon
  }
  return icons[props.type] || InformationCircleIcon
})

const iconBgClass = computed(() => {
  const classes = {
    block: 'bg-red-500/20',
    alert: 'bg-yellow-500/20',
    unblock: 'bg-green-500/20',
    success: 'bg-green-500/20',
    error: 'bg-red-500/20',
    info: 'bg-blue-500/20'
  }
  return classes[props.type] || 'bg-gray-500/20'
})

const iconColorClass = computed(() => {
  const classes = {
    block: 'text-red-400',
    alert: 'text-yellow-400',
    unblock: 'text-green-400',
    success: 'text-green-400',
    error: 'text-red-400',
    info: 'text-blue-400'
  }
  return classes[props.type] || 'text-gray-400'
})
</script>
