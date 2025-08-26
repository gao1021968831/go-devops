<template>
  <el-dialog
    v-model="visible"
    title="文件详情"
    width="600px"
    :before-close="handleClose"
  >
    <div class="detail-container" v-if="file">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="文件名">
          <div class="file-info">
            <el-icon class="file-icon" :color="getFileIconColor(file.mime_type)">
              <component :is="getFileIcon(file.mime_type)" />
            </el-icon>
            <span>{{ file.original_name }}</span>
          </div>
        </el-descriptions-item>
        
        <el-descriptions-item label="文件大小">
          {{ formatFileSize(file.size) }}
        </el-descriptions-item>
        
        <el-descriptions-item label="文件类型">
          {{ file.mime_type }}
        </el-descriptions-item>
        
        <el-descriptions-item label="MD5哈希">
          <el-text class="hash-text">{{ file.md5_hash }}</el-text>
        </el-descriptions-item>
        
        <el-descriptions-item label="文件分类">
          <el-tag :type="getCategoryType(file.category)">
            {{ getCategoryName(file.category) }}
          </el-tag>
        </el-descriptions-item>
        
        <el-descriptions-item label="访问权限">
          <el-tag :type="file.is_public ? 'success' : 'info'" size="small">
            {{ file.is_public ? '公开' : '私有' }}
          </el-tag>
        </el-descriptions-item>
        
        <el-descriptions-item label="上传者">
          <div class="user-info">
            <el-avatar :size="24" :src="file.user?.avatar">
              {{ file.user?.username?.charAt(0) }}
            </el-avatar>
            <span>{{ file.user?.username }}</span>
          </div>
        </el-descriptions-item>
        
        <el-descriptions-item label="下载次数">
          <el-tag type="info" size="small">{{ file.download_count }}</el-tag>
        </el-descriptions-item>
        
        <el-descriptions-item label="上传时间">
          {{ formatDateTime(file.created_at) }}
        </el-descriptions-item>
        
        <el-descriptions-item label="更新时间">
          {{ formatDateTime(file.updated_at) }}
        </el-descriptions-item>
        
        <el-descriptions-item label="文件描述" v-if="file.description">
          <div class="description">{{ file.description }}</div>
        </el-descriptions-item>
      </el-descriptions>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">关闭</el-button>
        <el-button type="primary" @click="downloadFile" v-if="file">
          <el-icon><Download /></el-icon>
          下载文件
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Document, Download, User, Calendar, Picture, VideoPlay, Files } from '@element-plus/icons-vue'
import { formatFileSize, formatDateTime } from '@/utils/format.js'
import api from '@/utils/api'

// Props & Emits
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  file: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:modelValue'])

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

// 下载文件
const downloadFile = async () => {
  if (!props.file) return

  try {
    const response = await api.get(`/api/v1/files/${props.file.id}/download`, {
      responseType: 'blob'
    })
    
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.download = props.file.original_name
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    
    ElMessage.success('文件下载成功')
  } catch (error) {
    ElMessage.error('文件下载失败')
  }
}

// 获取文件图标
const getFileIcon = (mimeType) => {
  if (mimeType.startsWith('image/')) return Picture
  if (mimeType.startsWith('video/')) return VideoPlay
  if (mimeType.includes('text/') || mimeType.includes('script')) return Document
  return Files
}

// 获取文件图标颜色
const getFileIconColor = (mimeType) => {
  if (mimeType.startsWith('image/')) return '#67C23A'
  if (mimeType.startsWith('video/')) return '#E6A23C'
  if (mimeType.startsWith('audio/')) return '#F56C6C'
  if (mimeType.includes('text/') || mimeType.includes('script')) return '#409EFF'
  return '#909399'
}

// 获取分类类型
const getCategoryType = (category) => {
  const types = {
    general: 'primary',
    scripts: 'success',
    configs: 'warning',
    documents: 'info',
    others: 'danger'
  }
  return types[category] || 'primary'
}

// 获取分类名称
const getCategoryName = (category) => {
  const names = {
    general: '通用',
    scripts: '脚本',
    configs: '配置',
    documents: '文档',
    others: '其他'
  }
  return names[category] || category
}

// 关闭对话框
const handleClose = () => {
  visible.value = false
}
</script>

<style scoped>
.detail-container {
  max-height: 60vh;
  overflow-y: auto;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-icon {
  font-size: 20px;
}

.hash-text {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
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

.description {
  background: #f8f9fa;
  padding: 8px 12px;
  border-radius: 4px;
  border-left: 3px solid #409EFF;
  color: #606266;
}

.dialog-footer {
  text-align: right;
}

:deep(.el-descriptions__label) {
  font-weight: 500;
}
</style>
