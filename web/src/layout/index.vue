<template>
  <div class="layout-container">
    <!-- 顶部导航栏 -->
    <div class="layout-header">
      <div class="header-left">
        <h2 style="color: white; margin: 0;">{{ systemStore.appName }}</h2>
      </div>
      <div class="header-right">
        <el-dropdown @command="handleCommand">
          <span class="user-info">
            <el-avatar :size="32" style="margin-right: 8px;">
              {{ userStore.user?.username?.charAt(0).toUpperCase() }}
            </el-avatar>
            {{ userStore.user?.username }}
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人资料</el-dropdown-item>
              <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <!-- 主体内容 -->
    <div class="layout-content">
      <!-- 侧边栏 -->
      <div class="layout-sidebar">
        <el-menu
          :default-active="$route.path"
          router
          class="sidebar-menu"
        >
          <el-menu-item
            v-for="route in menuRoutes"
            :key="route.path"
            :index="route.path"
          >
            <el-icon><component :is="route.meta.icon" /></el-icon>
            <span>{{ route.meta.title }}</span>
          </el-menu-item>
        </el-menu>
      </div>

      <!-- 主要内容区域 -->
      <div class="layout-main">
        <router-view />
        
        <!-- 页脚版本信息 -->
        <div class="layout-footer">
          <span class="footer-text">
            {{ systemStore.appName }} v{{ systemStore.appVersion }}
            <el-tag v-if="!systemStore.isProduction" type="warning" size="small" style="margin-left: 8px;">
              {{ systemStore.appEnvironment }}
            </el-tag>
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useSystemStore } from '@/stores/system'
import { ElMessageBox } from 'element-plus'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const systemStore = useSystemStore()

// 初始化时加载系统信息
onMounted(() => {
  systemStore.loadSystemInfo()
})

// 菜单路由
const menuRoutes = computed(() => {
  const routes = router.getRoutes()
    .find(r => r.name === 'Layout')
    ?.children?.filter(child => {
      // 过滤掉需要管理员权限但当前用户不是管理员的路由
      if (child.meta?.requiresAdmin && userStore.user?.role !== 'admin') {
        return false
      }
      // 过滤掉隐藏的路由
      if (child.meta?.hidden) {
        return false
      }
      return child.meta?.title
    }) || []
  
  return routes.map(r => ({
    path: `/${r.path}`,
    meta: r.meta
  }))
})

const handleCommand = async (command) => {
  switch (command) {
    case 'profile':
      router.push('/profile')
      break
    case 'logout':
      try {
        await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        userStore.logout()
        router.push('/login')
      } catch {
        // 用户取消
      }
      break
  }
}
</script>

<style scoped>
.header-left {
  display: flex;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  color: white;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.user-info:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.layout-main {
  display: flex;
  flex-direction: column;
  min-height: calc(100vh - 60px);
}

.sidebar-menu {
  border: none;
  height: 100%;
}

.sidebar-menu .el-menu-item {
  height: 50px;
  line-height: 50px;
}

.sidebar-menu .el-menu-item.is-active {
  background-color: #e6f7ff;
  color: #1890ff;
  border-right: 3px solid #1890ff;
}

.layout-footer {
  margin-top: auto;
  padding: 16px 24px;
  text-align: center;
  border-top: 1px solid #f0f0f0;
  background-color: #fafafa;
}

.footer-text {
  font-size: 12px;
  color: #909399;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}
</style>
