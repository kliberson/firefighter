import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import Whitelist from '../views/Whitelist.vue'
import Analytics from '../views/Analytics.vue'
import Activity from '../views/Activity.vue'
import '../style.css'
const routes = [
  { path: '/', name: 'Dashboard', component: Dashboard },
  { path: '/activity', name: 'Activity', component: Activity },
  { path: '/analytics', name: 'Analytics', component: Analytics},
  { path: '/whitelist', name: 'Whitelist', component: Whitelist }
]

export default createRouter({
  history: createWebHistory(),
  routes
})
