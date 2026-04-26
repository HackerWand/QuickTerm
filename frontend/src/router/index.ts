import { createRouter, createWebHashHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useWorkspaceStore } from '../stores/workspace'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'WorkspaceSelector',
    component: () => import('../views/WorkspaceSelector.vue')
  },
  {
    path: '/workspace/:id',
    name: 'MainView',
    component: () => import('../views/MainView.vue')
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const workspaceStore = useWorkspaceStore()
  
  if (to.name === 'MainView' && !workspaceStore.currentWorkspace) {
    next({ name: 'WorkspaceSelector' })
  } else {
    next()
  }
})

export default router
