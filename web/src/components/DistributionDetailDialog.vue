<template>
  <el-dialog
    v-model="visible"
    title="分发详情"
    width="900px"
    :before-close="handleClose"
  >
    <div class="detail-container" v-if="distribution">
      <!-- 分发任务信息 -->
      <div class="task-info">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-icon><View /></el-icon>
              <span>任务信息</span>
            </div>
          </template>

          <el-descriptions :column="2" border>
            <el-descriptions-item label="文件名">
              <div class="file-info">
                <el-icon class="file-icon" :color="getFileIconColor(distribution.file?.mime_type)">
                  <component :is="getFileIcon(distribution.file?.mime_type)" />
                </el-icon>
                <span>{{ distribution.file?.original_name }}</span>
              </div>
            </el-descriptions-item>
            <el-descriptions-item label="文件大小">
              {{ formatFileSize(distribution.file?.size) }}
            </el-descriptions-item>
            <el-descriptions-item label="目标路径">
              <el-text class="target-path">{{ distribution.target_path }}</el-text>
            </el-descriptions-item>
            <el-descriptions-item label="任务状态">
              <el-tag :type="getStatusType(distribution.status)" :effect="getStatusEffect(distribution.status)">
                <el-icon v-if="distribution.status === 'running'" class="is-loading">
                  <Loading />
                </el-icon>
                {{ getStatusText(distribution.status) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="创建者">
              <div class="user-info">
                <el-avatar :size="24" :src="distribution.user?.avatar">
                  {{ distribution.user?.username?.charAt(0) }}
                </el-avatar>
                <span>{{ distribution.user?.username }}</span>
              </div>
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ formatDateTime(distribution.created_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="开始时间">
              {{ distribution.start_time ? formatDateTime(distribution.start_time) : '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="结束时间">
              {{ distribution.end_time ? formatDateTime(distribution.end_time) : '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="总体进度" :span="2">
              <div class="progress-container">
                <el-progress 
                  :percentage="distribution.progress || 0" 
                  :stroke-width="12"
                  :color="getProgressColor(distribution.status)"
                />
              </div>
            </el-descriptions-item>
            <el-descriptions-item label="任务描述" :span="2" v-if="distribution.description">
              <div class="description">{{ distribution.description }}</div>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </div>

      <!-- 主机分发详情 -->
      <div class="host-details">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-icon><Monitor /></el-icon>
              <span>主机分发详情</span>
              <div class="header-actions">
                <el-button size="small" @click="loadDetails">
                  <el-icon><Refresh /></el-icon>
                  刷新
                </el-button>
              </div>
            </div>
          </template>

          <!-- 统计信息 -->
          <div class="stats-section" v-if="details.length > 0">
            <el-row :gutter="16">
              <el-col :span="6">
                <el-statistic title="总主机数" :value="details.length" />
              </el-col>
              <el-col :span="6">
                <el-statistic title="成功" :value="getSuccessCount()" />
              </el-col>
              <el-col :span="6">
                <el-statistic title="失败" :value="getFailedCount()" />
              </el-col>
              <el-col :span="6">
                <el-statistic title="运行中" :value="getRunningCount()" />
              </el-col>
            </el-row>
          </div>

          <!-- 主机列表 -->
          <div class="host-list" v-loading="loadingDetails">
            <el-table :data="details" style="width: 100%">
              <el-table-column label="主机" min-width="200">
                <template #default="{ row }">
                  <div class="host-info">
                    <div class="host-status">
                      <div 
                        class="status-dot"
                        :style="{ backgroundColor: row.host?.status === 'online' ? '#67C23A' : '#F56C6C' }"
                      ></div>
                    </div>
                    <div class="host-details">
                      <div class="host-name">{{ row.host?.name }}</div>
                      <div class="host-ip">{{ row.host?.ip }}:{{ row.host?.port }}</div>
                    </div>
                  </div>
                </template>
              </el-table-column>

              <el-table-column label="状态" width="120">
                <template #default="{ row }">
                  <el-tag :type="getStatusType(row.status)" size="small">
                    <el-icon v-if="row.status === 'running'" class="is-loading">
                      <Loading />
                    </el-icon>
                    {{ getStatusText(row.status) }}
                  </el-tag>
                </template>
              </el-table-column>

              <el-table-column label="开始时间" width="160">
                <template #default="{ row }">
                  {{ row.start_time ? formatDateTime(row.start_time) : '-' }}
                </template>
              </el-table-column>

              <el-table-column label="结束时间" width="160">
                <template #default="{ row }">
                  {{ row.end_time ? formatDateTime(row.end_time) : '-' }}
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

              <el-table-column label="结果" min-width="200">
                <template #default="{ row }">
                  <div v-if="row.status === 'completed' && row.output">
                    <el-text type="success" size="small">{{ row.output }}</el-text>
                  </div>
                  <div v-else-if="row.status === 'failed' && row.error">
                    <el-text type="danger" size="small">{{ row.error }}</el-text>
                  </div>
                  <div v-else-if="row.status === 'running'">
                    <el-text type="warning" size="small">正在执行...</el-text>
                  </div>
                  <div v-else>
                    <el-text type="info" size="small">等待执行</el-text>
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-card>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">关闭</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  View, Monitor, Refresh, Loading, Picture, VideoPlay, Files, Document
} from '@element-plus/icons-vue'
import { formatDateTime, formatFileSize } from '@/utils/format.js'
import api from '@/utils/api'

// Props & Emits
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  distribution: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:modelValue'])

// 响应式数据
const loadingDetails = ref(false)
const details = ref([])

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

// 加载分发详情
const loadDetails = async () => {
  if (!props.distribution?.id) return

  loadingDetails.value = true
  try {
    const response = await api.get(`/api/v1/file-distributions/${props.distribution.id}`)
    details.value = response.data.details || []
  } catch (error) {
    ElMessage.error('加载分发详情失败')
  } finally {
    loadingDetails.value = false
  }
}

// 统计方法
const getSuccessCount = () => {
  return details.value.filter(d => d.status === 'completed').length
}

const getFailedCount = () => {
  return details.value.filter(d => d.status === 'failed').length
}

const getRunningCount = () => {
  return details.value.filter(d => d.status === 'running').length
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

// 关闭对话框
const handleClose = () => {
  visible.value = false
}

// 监听对话框显示状态
watch(visible, (newVal) => {
  if (newVal && props.distribution) {
    loadDetails()
  }
})

// 监听分发对象变化
watch(() => props.distribution, (newDistribution) => {
  if (newDistribution && visible.value) {
    loadDetails()
  }
})
</script>

<style scoped>
.detail-container {
  max-height: 70vh;
  overflow-y: auto;
}

.task-info {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.header-actions {
  margin-left: auto;
  display: flex;
  gap: 8px;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-icon {
  font-size: 16px;
}

.target-path {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  color: #606266;
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 4px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.progress-container {
  width: 100%;
}

.description {
  background: #f8f9fa;
  padding: 8px 12px;
  border-radius: 4px;
  border-left: 3px solid #409EFF;
  color: #606266;
}

.stats-section {
  margin-bottom: 20px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
}

.host-details {
  margin-bottom: 20px;
}

.host-list {
  margin-top: 16px;
}

.host-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.host-status {
  display: flex;
  align-items: center;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 8px;
}

.host-details {
  flex: 1;
}

.host-name {
  font-weight: 500;
  color: #2c3e50;
  margin-bottom: 4px;
}

.host-ip {
  font-size: 13px;
  color: #7f8c8d;
}

.dialog-footer {
  text-align: right;
}

:deep(.el-card__header) {
  padding: 16px 20px;
  background: #f8f9fa;
}

:deep(.el-card__body) {
  padding: 20px;
}

:deep(.el-descriptions__label) {
  font-weight: 500;
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
