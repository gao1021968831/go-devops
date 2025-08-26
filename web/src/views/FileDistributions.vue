<template>
  <div class="distributions-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">
          <el-icon><Share /></el-icon>
          文件分发记录
        </h2>
        <p class="page-description">查看和管理文件分发任务</p>
      </div>
      <div class="header-right">
        <el-button @click="$router.push('/api/v1/files')">
          <el-icon><Back /></el-icon>
          返回文件管理
        </el-button>
      </div>
    </div>

    <!-- 筛选区域 -->
    <div class="filter-section">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-select v-model="searchForm.status" placeholder="分发状态" clearable>
            <el-option label="全部状态" value="" />
            <el-option label="等待中" value="pending" />
            <el-option label="运行中" value="running" />
            <el-option label="已完成" value="completed" />
            <el-option label="失败" value="failed" />
            <el-option label="部分成功" value="partial" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-button type="primary" @click="loadDistributions">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
        </el-col>
        <el-col :span="4" :offset="10">
          <el-button @click="loadDistributions">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </el-col>
      </el-row>
    </div>

    <!-- 分发记录列表 -->
    <div class="distributions-table">
      <el-table
        v-loading="loading"
        :data="distributions"
        style="width: 100%"
        @row-click="showDistributionDetail"
      >
        <el-table-column label="文件名" min-width="200">
          <template #default="{ row }">
            <div class="file-info">
              <el-icon class="file-icon" :color="getFileIconColor(row.file?.mime_type)">
                <component :is="getFileIcon(row.file?.mime_type)" />
              </el-icon>
              <div class="file-details">
                <div class="file-name">{{ row.file?.original_name }}</div>
                <div class="file-meta">{{ formatFileSize(row.file?.size) }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="目标路径" min-width="180">
          <template #default="{ row }">
            <el-text class="target-path">{{ row.target_path }}</el-text>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" :effect="getStatusEffect(row.status)">
              <el-icon v-if="row.status === 'running'" class="is-loading">
                <Loading />
              </el-icon>
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="进度" width="120">
          <template #default="{ row }">
            <div class="progress-container">
              <el-progress 
                :percentage="row.progress || 0" 
                :stroke-width="8"
                :show-text="false"
                :color="getProgressColor(row.status)"
              />
              <span class="progress-text">{{ row.progress || 0 }}%</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="创建者" width="120">
          <template #default="{ row }">
            <div class="user-info">
              <el-avatar :size="24" :src="row.user?.avatar">
                {{ row.user?.username?.charAt(0) }}
              </el-avatar>
              <span class="username">{{ row.user?.username }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column label="耗时" width="100">
          <template #default="{ row }">
            <span v-if="row.start_time && row.end_time">
              {{ formatDuration(row.start_time, row.end_time) }}
            </span>
            <span v-else-if="row.start_time">
              {{ formatDuration(row.start_time, new Date()) }}
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click.stop="showDistributionDetail(row)">
              <el-icon><View /></el-icon>
              详情
            </el-button>
            <el-button 
              size="small" 
              type="danger" 
              @click.stop="deleteDistribution(row)"
              v-if="canDeleteDistribution(row)"
            >
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadDistributions"
          @current-change="loadDistributions"
        />
      </div>
    </div>

    <!-- 分发详情对话框 -->
    <DistributionDetailDialog
      v-model="showDetailDialog"
      :distribution="currentDistribution"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Share, Back, Search, Refresh, View, Delete, Loading, Picture, VideoPlay, Files, Document
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { formatDateTime, formatFileSize } from '@/utils/format.js'
import api from '@/utils/api'
import DistributionDetailDialog from '@/components/DistributionDetailDialog.vue'

const userStore = useUserStore()

// 响应式数据
const loading = ref(false)
const distributions = ref([])
const showDetailDialog = ref(false)
const currentDistribution = ref(null)

// 搜索表单
const searchForm = reactive({
  status: ''
})

// 分页信息
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 加载分发记录列表
const loadDistributions = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size,
      ...searchForm
    }
    
    const response = await api.get('/api/v1/file-distributions', { params })
    distributions.value = response.data.data
    pagination.total = response.data.total
  } catch (error) {
    ElMessage.error('加载分发记录失败')
  } finally {
    loading.value = false
  }
}

// 显示分发详情
const showDistributionDetail = (distribution) => {
  currentDistribution.value = distribution
  showDetailDialog.value = true
}

// 获取文件图标
const getFileIcon = (mimeType) => {
  if (!mimeType) return Files
  if (mimeType.startsWith('image/')) return Picture
  if (mimeType.startsWith('video/')) return VideoPlay
  if (mimeType.includes('text/') || mimeType.includes('script')) return Document
  return Files
}

// 获取文件图标颜色
const getFileIconColor = (mimeType) => {
  if (!mimeType) return '#909399'
  if (mimeType.startsWith('image/')) return '#67C23A'
  if (mimeType.startsWith('video/')) return '#E6A23C'
  if (mimeType.startsWith('audio/')) return '#F56C6C'
  if (mimeType.includes('text/') || mimeType.includes('script')) return '#409EFF'
  return '#909399'
}

// 获取状态类型
const getStatusType = (status) => {
  const types = {
    pending: 'info',
    running: 'warning',
    completed: 'success',
    failed: 'danger',
    partial: 'warning'
  }
  return types[status] || 'info'
}

// 获取状态效果
const getStatusEffect = (status) => {
  return status === 'running' ? 'plain' : 'dark'
}

// 获取状态文本
const getStatusText = (status) => {
  const texts = {
    pending: '等待中',
    running: '运行中',
    completed: '已完成',
    failed: '失败',
    partial: '部分成功'
  }
  return texts[status] || status
}

// 获取进度条颜色
const getProgressColor = (status) => {
  const colors = {
    pending: '#909399',
    running: '#E6A23C',
    completed: '#67C23A',
    failed: '#F56C6C',
    partial: '#E6A23C'
  }
  return colors[status] || '#409EFF'
}

// 格式化持续时间
const formatDuration = (startTime, endTime) => {
  const start = new Date(startTime)
  const end = new Date(endTime)
  const duration = Math.floor((end - start) / 1000) // 秒

  if (duration < 60) {
    return `${duration}秒`
  } else if (duration < 3600) {
    const minutes = Math.floor(duration / 60)
    const seconds = duration % 60
    return `${minutes}分${seconds}秒`
  } else {
    const hours = Math.floor(duration / 3600)
    const minutes = Math.floor((duration % 3600) / 60)
    return `${hours}时${minutes}分`
  }
}

// 检查是否可以删除分发记录
const canDeleteDistribution = (distribution) => {
  const currentUser = userStore.user
  // 管理员或创建者可以删除
  return currentUser.role === 'admin' || distribution.created_by === currentUser.id
}

// 删除分发记录
const deleteDistribution = async (distribution) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除分发记录吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    await api.delete(`/api/v1/file-distributions/${distribution.id}`)
    ElMessage.success('分发记录删除成功')
    loadDistributions()
  } catch (error) {
    if (error !== 'cancel') {
      const errorMessage = error.response?.data?.error || '分发记录删除失败'
      ElMessage.error(errorMessage)
    }
  }
}

// 初始化
onMounted(() => {
  loadDistributions()
  
  // 设置自动刷新（仅针对运行中的任务）
  const refreshInterval = setInterval(() => {
    const hasRunningTasks = distributions.value.some(d => d.status === 'running')
    if (hasRunningTasks) {
      loadDistributions()
    }
  }, 5000) // 每5秒刷新一次

  // 组件卸载时清理定时器
  // onUnmounted(() => {
  //   clearInterval(refreshInterval)
  // })
})
</script>

<style scoped>
.distributions-container {
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 24px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 12px;
  backdrop-filter: blur(10px);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.header-left .page-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}

.header-left .page-description {
  margin: 0;
  color: #7f8c8d;
  font-size: 14px;
}

.filter-section {
  margin-bottom: 20px;
  padding: 20px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 12px;
  backdrop-filter: blur(10px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

.distributions-table {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 12px;
  backdrop-filter: blur(10px);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.file-icon {
  font-size: 20px;
}

.file-details {
  flex: 1;
}

.file-name {
  font-weight: 500;
  color: #2c3e50;
  margin-bottom: 4px;
}

.file-meta {
  font-size: 12px;
  color: #7f8c8d;
}

.target-path {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  color: #606266;
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 4px;
}

.progress-container {
  display: flex;
  align-items: center;
  gap: 8px;
}

.progress-text {
  font-size: 12px;
  color: #606266;
  min-width: 35px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.username {
  font-size: 13px;
  color: #2c3e50;
}

.pagination-container {
  padding: 20px;
  text-align: center;
}

:deep(.el-table) {
  background: transparent;
}

:deep(.el-table__header) {
  background: #f8f9fa;
}

:deep(.el-table tr) {
  background: transparent;
  cursor: pointer;
}

:deep(.el-table--enable-row-hover .el-table__body tr:hover > td) {
  background: rgba(64, 158, 255, 0.1);
}

:deep(.el-tag .el-icon.is-loading) {
  animation: rotating 2s linear infinite;
}

@keyframes rotating {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}
</style>
