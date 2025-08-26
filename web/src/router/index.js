import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/layout/index.vue'),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘', icon: 'DataBoard' }
      },
      {
        path: 'topology',
        name: 'Topology',
        component: () => import('@/views/Topology.vue'),
        meta: { title: '主机拓扑', icon: 'Share' }
      },
      {
        path: 'hosts',
        name: 'Hosts',
        component: () => import('@/views/Hosts.vue'),
        meta: { title: '主机管理', icon: 'Monitor' }
      },
      {
        path: 'scripts',
        name: 'Scripts',
        component: () => import('@/views/Scripts.vue'),
        meta: { title: '脚本管理', icon: 'Document' }
      },
      {
        path: 'executions',
        name: 'Executions',
        component: () => import('@/views/Executions.vue'),
        meta: { title: '执行记录', icon: 'Operation' }
      },
      {
        path: 'executions/:id',
        name: 'ExecutionDetail',
        component: () => import('@/views/ExecutionDetail.vue'),
        meta: { title: '执行详情', hidden: true }
      },
      {
        path: 'files',
        name: 'Files',
        component: () => import('@/views/Files.vue'),
        meta: { title: '文件管理', icon: 'Folder' }
      },
      {
        path: 'file-distributions',
        name: 'FileDistributions',
        component: () => import('@/views/FileDistributions.vue'),
        meta: { title: '分发记录', icon: 'Share' }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/Users.vue'),
        meta: { title: '用户管理', icon: 'User', requiresAdmin: true }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/Profile.vue'),
        meta: { title: '个人资料', icon: 'UserFilled', hidden: true }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  if (to.meta.requiresAuth !== false && !userStore.token) {
    next('/login')
  } else if (to.path === '/login' && userStore.token) {
    next('/')
  } else if (to.meta.requiresAdmin && userStore.user?.role !== 'admin') {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
