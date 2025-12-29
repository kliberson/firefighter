const API_URL = window.location.origin

export default {
  // Stats
  async getStats() {
    const res = await fetch(`${API_URL}/api/stats`)
    return res.json()
  },

  async getAlertBuckets(days = 1) {
    const res = await fetch(`${API_URL}/api/stats/alerts/buckets?days=${days}`)
    return res.json()
  },

  async getBlockBuckets(days = 1) {
    const res = await fetch(`${API_URL}/api/stats/blocks/buckets?days=${days}`)
    return res.json()
  },

  async getTopIPs(limit = 5) {
    const res = await fetch(`${API_URL}/api/stats/top_ips?limit=${limit}`)
    return res.json()
  },

  async getCategories(days = 1) {
    const res = await fetch(`${API_URL}/api/stats/categories?days=${days}`)
    return res.json()
  },

  async getAlertsByIP(ip) {
    const res = await fetch(`${API_URL}/api/stats/alerts/by_ip?ip=${encodeURIComponent(ip)}`)
    return res.json()
  },

  async getBlockedByIP(ip) {
    const res = await fetch(`${API_URL}/api/blocked/by_ip?ip=${encodeURIComponent(ip)}`)
    return res.json()
  },

  // Blocked IPs
  async getBlockedIPs() {
    const res = await fetch(`${API_URL}/api/blocked`)
    return res.json()
  },

  async unblockIP(ip) {
    return fetch(`${API_URL}/api/unblock/${ip}`, { method: 'POST' })
  },

  // Whitelist
  async getWhitelist() {
    const res = await fetch(`${API_URL}/api/whitelist`)
    return res.json()
  },

  async addToWhitelist(ip, description) {
    return fetch(`${API_URL}/api/whitelist/${ip}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ip, description })
    })
  },

  async getActivity(search = '', typeFilter = '', limit = 100) {
    const params = new URLSearchParams()
    if (search) params.append('search', search)
    if (typeFilter) params.append('type', typeFilter) // ⬅️ DODAJ TO
    params.append('limit', limit)
    
    const res = await fetch(`${API_URL}/api/activity?${params}`)
    return res.json()
  },

  async removeFromWhitelist(ip) {
    return fetch(`${API_URL}/api/whitelist/${ip}`, { method: 'DELETE' })
  }
}
