<template>
  <div class="files-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">
          <el-icon><Folder /></el-icon>
          文件管理
        </h2>
        <p class="page-description">管理和分发系统文件</p>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="showUploadDialog = true">
          <el-icon><Upload /></el-icon>
          上传文件
        </el-button>
      </div>
    </div>

    <!-- 搜索和筛选 -->
    <div class="filter-section">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-input
            v-model="searchForm.name"
            placeholder="搜索文件名"
            clearable
            @keyup.enter="loadFiles"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :span="4">
          <el-select v-model="searchForm.category" placeholder="文件分类" clearable>
            <el-option label="全部分类" value="" />
            <el-option label="通用文件" value="general" />
            <el-option label="脚本文件" value="scripts" />
            <el-option label="配置文件" value="configs" />
            <el-option label="文档文件" value="documents" />
            <el-option label="其他" value="others" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="searchForm.isPublic" placeholder="访问权限" clearable>
            <el-option label="全部文件" value="" />
            <el-option label="公开文件" value="true" />
            <el-option label="私有文件" value="false" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-button type="primary" @click="loadFiles">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
        </el-col>
      </el-row>
    </div>

    <!-- 文件列表 -->
    <div class="files-table">
      <el-table
        v-loading="loading"
        :data="files"
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column label="文件名" min-width="200">
          <template #default="{ row }">
            <div class="file-info">
              <el-icon class="file-icon" :color="getFileIconColor(row.mime_type)">
                <component :is="getFileIcon(row.mime_type)" />
              </el-icon>
              <div class="file-details">
                <div class="file-name">{{ row.original_name }}</div>
                <div class="file-meta">{{ formatFileSize(row.size) }} • {{ row.mime_type }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="分类" width="100">
          <template #default="{ row }">
            <el-tag :type="getCategoryType(row.category)">
              {{ getCategoryName(row.category) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="权限" width="80">
          <template #default="{ row }">
            <el-tag :type="row.is_public ? 'success' : 'info'" size="small">
              {{ row.is_public ? '公开' : '私有' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="上传者" width="120">
          <template #default="{ row }">
            <div class="user-info">
              <el-avatar :size="24" :src="row.user?.avatar">
                {{ row.user?.username?.charAt(0) }}
              </el-avatar>
              <span class="username">{{ row.user?.username }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="下载次数" width="100" align="center">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.download_count }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="上传时间" width="160">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" @click="downloadFile(row)" title="下载">
                <el-icon><Download /></el-icon>
              </el-button>
              <el-button size="small" @click="showFileDetail(row)" title="详情">
                <el-icon><View /></el-icon>
              </el-button>
              <el-button 
                size="small" 
                type="primary" 
                @click="showDistributeDialog(row)"
                v-if="canOperate(row)"
                title="分发"
              >
                <el-icon><Share /></el-icon>
              </el-button>
              <el-button 
                size="small" 
                @click="editFile(row)"
                v-if="canOperate(row)"
                title="编辑"
              >
                <el-icon><Edit /></el-icon>
              </el-button>
              <el-button 
                size="small" 
                type="danger" 
                @click="deleteFile(row)"
                v-if="canOperate(row)"
                title="删除"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
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
          @size-change="loadFiles"
          @current-change="loadFiles"
        />
      </div>
    </div>

    <!-- 批量操作 -->
    <div class="batch-actions" v-if="selectedFiles.length > 0">
      <el-card>
        <div class="batch-info">
          <span>已选择 {{ selectedFiles.length }} 个文件</span>
          <div class="batch-buttons">
            <el-button @click="batchDownload">批量下载</el-button>
            <el-button type="primary" @click="batchDistribute">批量分发</el-button>
            <el-button type="danger" @click="batchDelete">批量删除</el-button>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 文件上传对话框 -->
    <FileUploadDialog
      v-model="showUploadDialog"
      @uploaded="handleFileUploaded"
    />

    <!-- 文件分发对话框 -->
    <FileDistributeDialog
      v-model="showDistributeDialogVisible"
      :file="currentFile"
      @distributed="handleFileDistributed"
    />

    <!-- 文件详情对话框 -->
    <FileDetailDialog
      v-model="showDetailDialog"
      :file="currentFile"
    />

    <!-- 文件编辑对话框 -->
    <FileEditDialog
      v-model="showEditDialog"
      :file="currentFile"
      @updated="handleFileUpdated"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Folder, UploadFilled, Search, Download, View, Share, Edit, Delete, Document, Picture, VideoPlay, Files 
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { formatFileSize, formatDateTime, formatRelativeTime } from '@/utils/format.js'
import api from '@/utils/api'
import FileUploadDialog from '@/components/FileUploadDialog.vue'
import FileDistributeDialog from '@/components/FileDistributeDialog.vue'
import FileDetailDialog from '@/components/FileDetailDialog.vue'
import FileEditDialog from '@/components/FileEditDialog.vue'

const userStore = useUserStore()

// 响应式数据
const loading = ref(false)
const files = ref([])
const selectedFiles = ref([])
const showUploadDialog = ref(false)
const showDistributeDialogVisible = ref(false)
const showDetailDialog = ref(false)
const showEditDialog = ref(false)
const currentFile = ref(null)

// 搜索表单
const searchForm = reactive({
  name: '',
  category: '',
  isPublic: ''
})

// 分页信息
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 计算属性
const canOperate = computed(() => (file) => {
  return userStore.user?.role === 'admin' || file.uploaded_by === userStore.user?.id
})

// 加载文件列表
const loadFiles = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size,
      ...searchForm
    }
    
    const response = await api.get('/api/v1/files', { params })
    files.value = response.data.data
    pagination.total = response.data.total
  } catch (error) {
    ElMessage.error('加载文件列表失败')
  } finally {
    loading.value = false
  }
}

// 文件选择变化
const handleSelectionChange = (selection) => {
  selectedFiles.value = selection
}

// 下载文件
const downloadFile = async (file) => {
  try {
    const response = await api.get(`/api/v1/files/${file.id}/download`, {
      responseType: 'blob'
    })
    
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.download = file.original_name
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    
    ElMessage.success('文件下载成功')
  } catch (error) {
    ElMessage.error('文件下载失败')
  }
}

// 显示文件详情
const showFileDetail = (file) => {
  currentFile.value = file
  showDetailDialog.value = true
}

// 显示分发对话框
const showDistributeDialog = (file) => {
  currentFile.value = file
  showDistributeDialogVisible.value = true
}

// 编辑文件
const editFile = (file) => {
  currentFile.value = file
  showEditDialog.value = true
}

// 删除文件
const deleteFile = async (file) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除文件 "${file.original_name}" 吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    await api.delete(`/api/v1/files/${file.id}`)
    ElMessage.success('文件删除成功')
    loadFiles()
  } catch (error) {
    if (error !== 'cancel') {
      // 处理409冲突错误（文件已分发）
      if (error.response && error.response.status === 409) {
        const errorData = error.response.data
        ElMessageBox.alert(
          errorData.message || `文件 "${file.original_name}" 已被分发到主机，无法直接删除。请先删除相关的分发记录，或联系管理员处理。`,
          '无法删除文件',
          {
            confirmButtonText: '我知道了',
            type: 'warning',
            dangerouslyUseHTMLString: true
          }
        )
      } else {
        // 其他错误
        const errorMessage = error.response?.data?.error || '文件删除失败'
        ElMessage.error(errorMessage)
      }
    }
  }
}

// 批量下载
const batchDownload = () => {
  selectedFiles.value.forEach(file => {
    downloadFile(file)
  })
}

// 批量分发
const batchDistribute = () => {
  if (selectedFiles.value.length === 1) {
    showDistributeDialog(selectedFiles.value[0])
  } else {
    ElMessage.info('批量分发功能开发中，请先选择单个文件进行分发')
  }
}

// 批量删除
const batchDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedFiles.value.length} 个文件吗？此操作不可恢复。`,
      '确认批量删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    const promises = selectedFiles.value.map(file => 
      api.delete(`/api/v1/files/${file.id}`)
    )
    
    await Promise.all(promises)
    ElMessage.success('批量删除成功')
    selectedFiles.value = []
    loadFiles()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

// 获取文件图标
const getFileIcon = (mimeType) => {
  if (mimeType.startsWith('image/')) return Picture
  if (mimeType.startsWith('video/')) return VideoPlay
  if (mimeType.startsWith('audio/')) return Headphone
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

// 事件处理
const handleFileUploaded = () => {
  showUploadDialog.value = false
  loadFiles()
}

const handleFileDistributed = () => {
  showDistributeDialogVisible.value = false
  ElMessage.success('文件分发任务创建成功')
}

const handleFileUpdated = () => {
  showEditDialog.value = false
  loadFiles()
}

// 初始化
onMounted(() => {
  loadFiles()
})
</script>

<style scoped>
.files-container {
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

.files-table {
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

.batch-actions {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 1000;
}

.batch-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.batch-buttons {
  display: flex;
  gap: 8px;
}

:deep(.el-table) {
  background: transparent;
}

:deep(.el-table__header) {
  background: #f8f9fa;
}

:deep(.el-table tr) {
  background: transparent;
}

:deep(.el-table--enable-row-hover .el-table__body tr:hover > td) {
  background: rgba(64, 158, 255, 0.1);
}

.action-buttons {
  display: flex;
  gap: 4px;
  justify-content: flex-start;
  align-items: center;
  flex-wrap: nowrap;
}

.action-buttons .el-button {
  min-width: 32px;
  padding: 5px 8px;
}
</style>
