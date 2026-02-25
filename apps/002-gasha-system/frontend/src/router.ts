import { createRouter, createWebHistory } from 'vue-router'
import UserView from './views/UserView.vue'
import AdminView from './views/AdminView.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/user' },
    { path: '/user', component: UserView },
    { path: '/admin', component: AdminView }
  ]
})
