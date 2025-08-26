<template>
  <el-dialog
    v-model="visible"
    title="上传文件"
    width="600px"
    :before-close="handleClose"
  >
    <div class="upload-container">
      <!-- 上传区域 -->
      <el-upload
        ref="uploadRef"
        class="upload-dragger"
        drag
        :action="uploadUrl"
        :headers="uploadHeaders"
        :data="uploadData"
        :multiple="true"
        :auto-upload="false"
        :on-change="handleFileChange"
        :on-success="handleUploadSuccess"
        :on-error="handleUploadError"
        :on-progress="handleUploadProgress"
        :before-upload="beforeUpload"
        :file-list="fileList"
        :limit="5"
        :on-exceed="handleExceed"
      >
        <el-icon class="el-icon--upload"><upload-filled /></el-icon>
        <div class="el-upload__text">
          将文件拖到此处，或<em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            支持多文件上传，单个文件不超过100MB，最多同时上传5个文件
          </div>
        </template>
      </el-upload>

      <!-- 文件配置 -->
      <div class="file-config" v-if="fileList.length > 0">
        <el-divider content-position="left">文件配置</el-divider>
        
        <el-form :model="uploadForm" label-width="80px">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="文件分类">
                <el-select v-model="uploadForm.category" placeholder="选择分类">
                  <el-option label="通用文件" value="general" />
                  <el-option label="脚本文件" value="scripts" />
                  <el-option label="配置文件" value="configs" />
                  <el-option label="文档文件" value="documents" />
                  <el-option label="其他" value="others" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="访问权限">
                <el-switch
                  v-model="uploadForm.isPublic"
                  active-text="公开"
                  inactive-text="私有"
                />
              </el-form-item>
            </el-col>
          </el-row>
          
          <el-form-item label="文件描述">
            <el-input
              v-model="uploadForm.description"
              type="textarea"
              :rows="3"
              placeholder="请输入文件描述（可选）"
              maxlength="500"
              show-word-limit
            />
          </el-form-item>
        </el-form>
      </div>

      <!-- 文件列表 -->
      <div class="file-list" v-if="fileList.length > 0">
        <el-divider content-position="left">待上传文件</el-divider>
        
        <div class="file-item" v-for="file in fileList" :key="file.uid">
          <div class="file-info">
            <el-icon class="file-icon" :color="getFileIconColor(file.raw?.type)">
              <component :is="getFileIcon(file.raw?.type)" />
            </el-icon>
            <div class="file-details">
              <div class="file-name">{{ file.name }}</div>
              <div class="file-meta">
                {{ formatFileSize(file.size) }} • {{ file.raw?.type || '未知类型' }}
              </div>
              <div class="file-status" v-if="file.status">
                <el-tag 
                  :type="getStatusType(file.status)" 
                  size="small"
                >
                  {{ getStatusText(file.status) }}
                </el-tag>
              </div>
            </div>
          </div>
          
          <!-- 上传进度 -->
          <div class="file-progress" v-if="file.status === 'uploading'">
            <el-progress 
              :percentage="file.percentage || 0" 
              :stroke-width="6"
              :show-text="false"
            />
          </div>
          
          <!-- 操作按钮 -->
          <div class="file-actions">
            <el-button 
              size="small" 
              type="danger" 
              @click="removeFile(file)"
              :disabled="file.status === 'uploading'"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button 
          type="primary" 
          @click="startUpload"
          :loading="uploading"
          :disabled="fileList.length === 0"
        >
          {{ uploading ? '上传中...' : '开始上传' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { UploadFilled, Delete, Document, Picture, VideoPlay, Files } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { formatFileSize } from '@/utils/format.js'

const userStore = useUserStore()

// Props & Emits
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'uploaded'])

// 响应式数据
const uploadRef = ref()
const uploading = ref(false)
const fileList = ref([])

const uploadForm = reactive({
  category: 'general',
  description: '',
  isPublic: false
})

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const uploadUrl = computed(() => {
  return `${import.meta.env.VITE_API_BASE_URL || ''}/api/v1/files/upload`
})

const uploadHeaders = computed(() => ({
  'Authorization': `Bearer ${userStore.token}`
}))

const uploadData = computed(() => ({
  category: uploadForm.category,
  description: uploadForm.description,
  is_public: uploadForm.isPublic
}))

// 文件处理方法
const handleFileChange = (file, files) => {
  fileList.value = files
}

const handleExceed = (files, fileList) => {
  ElMessage.warning(`最多只能上传5个文件，当前选择了${files.length}个文件，共${files.length + fileList.length}个`)
}

const beforeUpload = (file) => {
  // 文件大小限制 100MB
  const maxSize = 100 * 1024 * 1024
  if (file.size > maxSize) {
    ElMessage.error(`文件 ${file.name} 大小超过100MB限制`)
    return false
  }
  
  return true
}

const removeFile = (file) => {
  const index = fileList.value.findIndex(item => item.uid === file.uid)
  if (index > -1) {
    fileList.value.splice(index, 1)
  }
}

// 上传处理
const startUpload = () => {
  if (fileList.value.length === 0) {
    ElMessage.warning('请先选择要上传的文件')
    return
  }
  
  uploading.value = true
  uploadRef.value.submit()
}

const handleUploadProgress = (event, file) => {
  file.percentage = Math.round(event.percent)
}

const handleUploadSuccess = (response, file) => {
  file.status = 'success'
  ElMessage.success(`文件 ${file.name} 上传成功`)
  
  // 检查是否所有文件都上传完成
  const allCompleted = fileList.value.every(f => 
    f.status === 'success' || f.status === 'fail'
  )
  
  if (allCompleted) {
    uploading.value = false
    const successCount = fileList.value.filter(f => f.status === 'success').length
    
    if (successCount > 0) {
      ElMessage.success(`成功上传 ${successCount} 个文件`)
      emit('uploaded')
      handleClose()
    }
  }
}

const handleUploadError = (error, file) => {
  file.status = 'fail'
  
  // 解析错误信息
  let errorMessage = `文件 ${file.name} 上传失败`
  let messageType = 'error'
  
  try {
    // Element Plus upload 组件的错误对象结构
    let responseData = null
    
    // 尝试多种方式解析响应数据
    if (error && error.response) {
      responseData = error.response
    } else if (error && typeof error === 'string') {
      try {
        responseData = JSON.parse(error)
      } catch (e) {
        // 如果不是JSON字符串，可能是XMLHttpRequest的responseText
        responseData = error
      }
    }
    
    // 检查HTTP状态码
    if (error && (error.status === 409 || (error.response && error.response.status === 409))) {
      // 文件已存在
      errorMessage = `文件 ${file.name} 已存在，请检查是否重复上传`
      messageType = 'warning'
    } else if (responseData) {
      // 尝试解析具体错误信息
      let parsedResponse = responseData
      if (typeof responseData === 'string') {
        try {
          parsedResponse = JSON.parse(responseData)
        } catch (e) {
          // 解析失败，使用原始字符串
        }
      }
      
      if (parsedResponse && parsedResponse.error) {
        if (parsedResponse.error === '文件已存在') {
          // 使用后端返回的具体消息，区分是否为用户重复上传
          if (parsedResponse.message && parsedResponse.message.includes('您已上传过相同的文件')) {
            errorMessage = `${parsedResponse.message}`
          } else {
            errorMessage = `文件 ${file.name} 已存在，请检查是否重复上传`
          }
          messageType = 'warning'
        } else {
          errorMessage = `文件 ${file.name} 上传失败: ${parsedResponse.error}`
        }
      }
    }
    
    // 显示相应的消息
    if (messageType === 'warning') {
      ElMessage.warning(errorMessage)
    } else {
      ElMessage.error(errorMessage)
    }
    
  } catch (e) {
    // 如果解析失败，使用默认错误信息
    console.error('解析上传错误失败:', e, error)
    ElMessage.error(errorMessage)
  }
  
  // 检查是否所有文件都处理完成
  const allCompleted = fileList.value.every(f => 
    f.status === 'success' || f.status === 'fail'
  )
  
  if (allCompleted) {
    uploading.value = false
  }
}

// 获取文件图标
const getFileIcon = (mimeType) => {
  if (!mimeType) return Files
  if (mimeType.startsWith('image/')) return Picture
  if (mimeType.startsWith('video/')) return VideoPlay
  if (mimeType.startsWith('audio/')) return Headphone
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
    ready: 'info',
    uploading: 'warning',
    success: 'success',
    fail: 'danger'
  }
  return types[status] || 'info'
}

// 获取状态文本
const getStatusText = (status) => {
  const texts = {
    ready: '准备上传',
    uploading: '上传中',
    success: '上传成功',
    fail: '上传失败'
  }
  return texts[status] || '未知状态'
}

// 关闭对话框
const handleClose = () => {
  if (uploading.value) {
    ElMessage.warning('文件正在上传中，请稍候...')
    return
  }
  
  // 重置数据
  fileList.value = []
  uploadForm.category = 'general'
  uploadForm.description = ''
  uploadForm.isPublic = false
  uploading.value = false
  
  visible.value = false
}

// 监听对话框显示状态
watch(visible, (newVal) => {
  if (!newVal) {
    // 对话框关闭时清理数据
    setTimeout(() => {
      fileList.value = []
      uploadForm.category = 'general'
      uploadForm.description = ''
      uploadForm.isPublic = false
      uploading.value = false
    }, 300)
  }
})
</script>

<style scoped>
.upload-container {
  padding: 20px 0;
}

.upload-dragger {
  margin-bottom: 20px;
}

:deep(.el-upload-dragger) {
  width: 100%;
  height: 180px;
  border: 2px dashed #d9d9d9;
  border-radius: 8px;
  background: #fafafa;
  transition: all 0.3s;
}

:deep(.el-upload-dragger:hover) {
  border-color: #409EFF;
  background: #f0f9ff;
}

:deep(.el-upload-dragger .el-icon--upload) {
  font-size: 48px;
  color: #c0c4cc;
  margin: 40px 0 16px;
}

:deep(.el-upload__text) {
  color: #606266;
  font-size: 14px;
  text-align: center;
}

:deep(.el-upload__text em) {
  color: #409EFF;
  font-style: normal;
}

:deep(.el-upload__tip) {
  font-size: 12px;
  color: #909399;
  margin-top: 7px;
  text-align: center;
}

.file-config {
  margin: 20px 0;
}

.file-list {
  margin-top: 20px;
}

.file-item {
  display: flex;
  align-items: center;
  padding: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  margin-bottom: 8px;
  background: #fafafa;
  transition: all 0.3s;
}

.file-item:hover {
  background: #f0f9ff;
  border-color: #409EFF;
}

.file-info {
  display: flex;
  align-items: center;
  flex: 1;
  gap: 12px;
}

.file-icon {
  font-size: 24px;
}

.file-details {
  flex: 1;
}

.file-name {
  font-weight: 500;
  color: #2c3e50;
  margin-bottom: 4px;
  font-size: 14px;
}

.file-meta {
  font-size: 12px;
  color: #7f8c8d;
  margin-bottom: 4px;
}

.file-status {
  margin-top: 4px;
}

.file-progress {
  flex: 0 0 200px;
  margin: 0 16px;
}

.file-actions {
  flex: 0 0 auto;
}

.dialog-footer {
  text-align: right;
}

:deep(.el-divider__text) {
  font-weight: 500;
  color: #2c3e50;
}
</style>
