import { createRouter, createWebHashHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/home/index.vue'),
    meta: { title: '首页' },
  },
  {
    path: '/share/:code',
    name: 'ShareView',
    component: () => import('@/views/share/View.vue'),
    meta: { title: '分享详情' },
  },
  {
    path: '/user/login',
    name: 'Login',
    component: () => import('@/views/user/Login.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/user/register',
    name: 'Register',
    component: () => import('@/views/user/Register.vue'),
    meta: { title: '注册' },
  },
  {
    path: '/user/dashboard',
    name: 'UserDashboard',
    component: () => import('@/views/user/Dashboard.vue'),
    meta: { title: '用户中心', requiresAuth: true },
  },
  {
    path: '/admin/login',
    name: 'AdminLogin',
    component: () => import('@/views/admin/Login.vue'),
    meta: { title: '管理员登录' },
  },
  {
    path: '/admin',
    name: 'Admin',
    component: () => import('@/views/admin/index.vue'),
    meta: { title: '管理后台', requiresAuth: true, requiresAdmin: true },
    children: [
      {
        path: '',
        name: 'AdminDashboard',
        component: () => import('@/views/admin/Dashboard.vue'),
        meta: { title: '仪表盘' },
      },
      {
        path: 'dashboard',
        name: 'AdminDashboardAlias',
        component: () => import('@/views/admin/Dashboard.vue'),
        meta: { title: '仪表盘' },
      },
      {
        path: 'files',
        name: 'AdminFiles',
        component: () => import('@/views/admin/Files.vue'),
        meta: { title: '文件管理' },
      },
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('@/views/admin/Users.vue'),
        meta: { title: '用户管理' },
      },
      {
        path: 'config',
        name: 'AdminConfig',
        component: () => import('@/views/admin/Config.vue'),
        meta: { title: '系统配置' },
      },
      {
        path: 'storage',
        name: 'AdminStorage',
        component: () => import('@/views/admin/Storage.vue'),
        meta: { title: '存储管理' },
      },
      {
        path: 'logs',
        name: 'AdminLogs',
        component: () => import('@/views/admin/TransferLogs.vue'),
        meta: { title: '传输日志' },
      },
      {
        path: 'maintenance',
        name: 'AdminMaintenance',
        component: () => import('@/views/admin/Maintenance.vue'),
        meta: { title: '维护工具' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHashHistory(),  // 使用 hash 模式，避免与后端 API 路由冲突
  routes,
})

// 路由守卫
router.beforeEach((to, _from, next) => {
  // 设置页面标题
  document.title = (to.meta.title as string) || 'FileCodeBox'

  // 检查是否需要登录
  if (to.meta.requiresAuth) {
    const token = localStorage.getItem('token')
    if (!token) {
      // 如果是管理后台，跳转到管理员登录页面
      if (to.path.startsWith('/admin')) {
        next('/admin/login')
      } else {
        next('/user/login')
      }
      return
    }

    // 检查是否需要管理员权限
    if (to.meta.requiresAdmin) {
      const userRole = localStorage.getItem('userRole')
      if (userRole !== 'admin') {
        next('/admin/login')
        return
      }
    }
  }

  next()
})

export default router
