export function formatTimestamp(unixTimestamp) {
  if (!unixTimestamp || unixTimestamp === 0) return 'N/A'
  const date = new Date(unixTimestamp * 1000)
  if (isNaN(date.getTime())) return 'Invalid Date'
  
  return date.toLocaleString('en-US', {
    month: '2-digit',
    day: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

export function formatTimeAgo(timestamp) {
  if (!timestamp) return 'N/A'
  const date = new Date(timestamp * 1000)
  const now = Date.now()
  const diff = Math.floor((now - date.getTime()) / 1000)
  
  if (diff < 60) return 'just now'
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return date.toLocaleTimeString()
}
