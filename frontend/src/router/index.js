import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import Whitelist from '../views/Whitelist.vue'
import '../style.css'
const routes = [
  { path: '/', name: 'Dashboard', component: Dashboard },
  { path: '/whitelist', name: 'Whitelist', component: Whitelist },
  { path: '/analytics', name: 'Analytics', component: () => import('../views/Analytics.vue')
}

]

export default createRouter({
  history: createWebHistory(),
  routes
})
