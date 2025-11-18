import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import Whitelist from '../views/Whitelist.vue'
import '../style.css'
const routes = [
  { path: '/', name: 'Dashboard', component: Dashboard },
  { path: '/whitelist', name: 'Whitelist', component: Whitelist }
]

export default createRouter({
  history: createWebHistory(),
  routes
})
