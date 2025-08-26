<template>
  <el-dialog
    v-model="visible"
    title="文件分发"
    width="800px"
    :before-close="handleClose"
  >
    <div class="distribute-container" v-if="file">
      <!-- 文件信息 -->
      <div class="file-info-section">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-icon><Document /></el-icon>
              <span>分发文件信息</span>
            </div>
          </template>
          
          <div class="file-info">
            <div class="file-item">
              <el-icon class="file-icon" :color="getFileIconColor(file.mime_type)">
                <component :is="getFileIcon(file.mime_type)" />
              </el-icon>
              <div class="file-details">
                <div class="file-name">{{ file.original_name }}</div>
                <div class="file-meta">
                  {{ formatFileSize(file.size) }} • {{ file.mime_type }} • {{ getCategoryName(file.category) }}
                </div>
                <div class="file-description" v-if="file.description">
                  {{ file.description }}
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </div>

      <!-- 分发配置 -->
      <div class="distribute-config">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-icon><Setting /></el-icon>
              <span>分发配置</span>
            </div>
          </template>

          <el-form :model="distributeForm" :rules="rules" ref="formRef" label-width="100px">
            <el-form-item label="目标路径" prop="targetPath" required>
              <el-input
                v-model="distributeForm.targetPath"
                placeholder="请输入目标主机上的文件路径，如: /tmp/script.sh"
                clearable
              >
                <template #prepend>路径</template>
              </el-input>
              <div class="form-tip">
                文件将保存到目标主机的此路径下
              </div>
            </el-form-item>

            <el-form-item label="任务描述">
              <el-input
                v-model="distributeForm.description"
                type="textarea"
                :rows="3"
                placeholder="请输入分发任务描述（可选）"
                maxlength="200"
                show-word-limit
              />
            </el-form-item>
          </el-form>
        </el-card>
      </div>

      <!-- 主机选择 -->
      <div class="host-selection">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-icon><Monitor /></el-icon>
              <span>目标主机选择</span>
              <div class="header-actions">
                <el-button size="small" @click="selectAllHosts">全选</el-button>
                <el-button size="small" @click="clearSelection">清空</el-button>
                <el-button size="small" @click="loadHosts">
                  <el-icon><Refresh /></el-icon>
                </el-button>
              </div>
            </div>
          </template>

          <!-- 主机筛选 -->
          <div class="host-filter">
            <el-row :gutter="16">
              <el-col :span="8">
                <el-input
                  v-model="hostFilter.name"
                  placeholder="搜索主机名或IP"
                  clearable
                  @input="filterHosts"
                >
                  <template #prefix>
                    <el-icon><Search /></el-icon>
                  </template>
                </el-input>
              </el-col>
              <el-col :span="6">
                <el-select v-model="hostFilter.status" placeholder="主机状态" clearable @change="filterHosts">
                  <el-option label="全部状态" value="" />
                  <el-option label="在线" value="online" />
                  <el-option label="离线" value="offline" />
                </el-select>
              </el-col>
              <el-col :span="6">
                <el-select v-model="hostFilter.environment" placeholder="环境" clearable @change="filterHosts">
                  <el-option label="全部环境" value="" />
                  <el-option label="生产环境" value="production" />
                  <el-option label="测试环境" value="test" />
                  <el-option label="开发环境" value="development" />
                </el-select>
              </el-col>
            </el-row>
          </div>

          <!-- 主机列表 -->
          <div class="host-list" v-loading="loadingHosts">
            <div class="selection-info" v-if="selectedHosts.length > 0">
              <el-tag type="primary">已选择 {{ selectedHosts.length }} 台主机</el-tag>
            </div>

            <div class="host-grid">
              <div
                v-for="host in filteredHosts"
                :key="host.id"
                class="host-card"
                :class="{ 'selected': selectedHosts.includes(host.id) }"
                @click="toggleHostSelection(host.id)"
              >
                <div class="host-header">
                  <div class="host-status">
                    <div 
                      class="status-dot"
                      :style="{ backgroundColor: host.status === 'online' ? '#67C23A' : '#F56C6C' }"
                    ></div>
                  </div>
                  <div class="host-selection">
                    <el-checkbox 
                      :model-value="selectedHosts.includes(host.id)"
                      @change="toggleHostSelection(host.id)"
                    />
                  </div>
                </div>
                
                <div class="host-info">
                  <div class="host-name">{{ host.name }}</div>
                  <div class="host-ip">{{ host.ip }}:{{ host.port }}</div>
                  <div class="host-os">{{ host.os }}</div>
                </div>

                <div class="host-meta">
                  <el-tag size="small" :type="getEnvironmentType(host.environment)">
                    {{ getEnvironmentName(host.environment) }}
                  </el-tag>
                </div>
              </div>
            </div>

            <el-empty v-if="filteredHosts.length === 0" description="没有找到匹配的主机" />
          </div>
        </el-card>
      </div>
    </div>

    <!-- 分发进度 -->
    <div class="distribution-progress" v-if="distributionProgress.show">
      <el-card>
        <template #header>
          <div class="card-header">
            <el-icon><Share /></el-icon>
            <span>分发进度</span>
          </div>
        </template>
        
        <div class="progress-info">
          <div class="progress-stats">
            <span>总进度: {{ distributionProgress.completed }}/{{ distributionProgress.total }}</span>
            <span class="progress-percent">{{ Math.round((distributionProgress.completed / distributionProgress.total) * 100) }}%</span>
          </div>
          
          <el-progress 
            :percentage="Math.round((distributionProgress.completed / distributionProgress.total) * 100)"
            :status="distributionProgress.status"
            :stroke-width="8"
          />
          
          <div class="progress-details" v-if="distributionProgress.details.length > 0">
            <div class="detail-item" v-for="detail in distributionProgress.details" :key="detail.host_id">
              <div class="host-info">
                <el-icon><Monitor /></el-icon>
                <span>{{ detail.host_name }} ({{ detail.host_ip }})</span>
              </div>
              <div class="status-info">
                <el-tag 
                  :type="getStatusType(detail.status)" 
                  size="small"
                >
                  {{ getStatusText(detail.status) }}
                </el-tag>
                <span class="time-info" v-if="detail.end_time">
                  {{ formatDateTime(detail.end_time) }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </el-card>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose" :disabled="distributing">取消</el-button>
        <el-button 
          type="primary" 
          @click="startDistribute"
          :loading="distributing"
          :disabled="selectedHosts.length === 0 || distributing"
          v-if="!distributionProgress.show"
        >
          {{ distributing ? '创建任务中...' : `开始分发 (${selectedHosts.length})` }}
        </el-button>
        <el-button 
          type="success" 
          @click="handleClose"
          v-if="distributionProgress.show && distributionProgress.status === 'success'"
        >
          完成
        </el-button>
        <el-button 
          type="warning" 
          @click="handleClose"
          v-if="distributionProgress.show && distributionProgress.status === 'exception'"
        >
          关闭
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Document, Setting, Monitor, Search, Refresh, Share, Warning, Picture, VideoPlay, Files
} from '@element-plus/icons-vue'
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

const emit = defineEmits(['update:modelValue', 'distributed'])

// 响应式数据
const formRef = ref()
const loading = ref(false)
const distributing = ref(false)
const loadingHosts = ref(false)
const hosts = ref([])
const selectedHosts = ref([])

const distributeForm = reactive({
  targetPath: '',
  description: ''
})

// 分发进度数据
const distributionProgress = reactive({
  show: false,
  total: 0,
  completed: 0,
  status: '', // '', 'success', 'exception'
  distributionId: null,
  details: []
})

const hostFilter = reactive({
  name: '',
  status: '',
  environment: ''
})

// 表单验证规则
const rules = {
  targetPath: [
    { required: true, message: '请输入目标路径', trigger: 'blur' },
    { min: 1, max: 500, message: '路径长度应在1-500字符之间', trigger: 'blur' }
  ]
}

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const filteredHosts = computed(() => {
  let result = hosts.value

  if (hostFilter.name) {
    const keyword = hostFilter.name.toLowerCase()
    result = result.filter(host => 
      host.name.toLowerCase().includes(keyword) ||
      host.ip.toLowerCase().includes(keyword)
    )
  }

  if (hostFilter.status) {
    result = result.filter(host => host.status === hostFilter.status)
  }

  if (hostFilter.environment) {
    result = result.filter(host => host.environment === hostFilter.environment)
  }

  return result
})

// 方法
const loadHosts = async () => {
  loadingHosts.value = true
  try {
    const response = await api.get('/api/v1/hosts', {
      params: { size: 1000 } // 获取所有主机
    })
    hosts.value = response.data.data || []
    
    if (hosts.value.length === 0) {
      ElMessage.warning('当前没有可用的主机，请联系管理员添加主机')
    }
  } catch (error) {
    console.error('加载主机列表失败:', error)
    if (error.response) {
      const { status, data } = error.response
      if (status === 403) {
        ElMessage.error('权限不足，无法获取主机列表')
      } else if (status === 401) {
        ElMessage.error('认证失败，请重新登录')
      } else {
        ElMessage.error(`加载主机列表失败: ${data?.error || '未知错误'}`)
      }
    } else {
      ElMessage.error('网络连接失败，无法加载主机列表')
    }
  } finally {
    loadingHosts.value = false
  }
}

const filterHosts = () => {
  // 触发计算属性重新计算
}

const toggleHostSelection = (hostId) => {
  const index = selectedHosts.value.indexOf(hostId)
  if (index > -1) {
    selectedHosts.value.splice(index, 1)
  } else {
    selectedHosts.value.push(hostId)
  }
}

const selectAllHosts = () => {
  selectedHosts.value = filteredHosts.value.map(host => host.id)
}

const clearSelection = () => {
  selectedHosts.value = []
}

// 获取状态类型
const getStatusType = (status) => {
  switch (status) {
    case 'success':
      return 'success'
    case 'failed':
      return 'danger'
    case 'running':
      return 'warning'
    default:
      return 'info'
  }
}

// 获取状态文本
const getStatusText = (status) => {
  switch (status) {
    case 'success':
      return '成功'
    case 'failed':
      return '失败'
    case 'running':
      return '执行中'
    case 'pending':
      return '等待中'
    default:
      return '未知'
  }
}

// 轮询分发进度
const pollDistributionProgress = async (distributionId) => {
  try {
    const response = await api.get(`/api/v1/file-distributions/${distributionId}`)
    const { distribution, details } = response.data
    
    // 更新进度数据
    distributionProgress.details = details || []
    distributionProgress.completed = distributionProgress.details.filter(d => d.status === 'completed' || d.status === 'failed').length
    
    // 检查是否全部完成
    const allCompleted = distributionProgress.details.every(d => d.status === 'completed' || d.status === 'failed')
    
    if (allCompleted) {
      const hasFailure = distributionProgress.details.some(d => d.status === 'failed')
      distributionProgress.status = hasFailure ? 'exception' : 'success'
      
      if (hasFailure) {
        ElMessage.warning('部分主机分发失败，请查看详情')
      } else {
        ElMessage.success('文件分发完成')
      }
      
      emit('distributed')
      return true // 停止轮询
    }
    
    return false // 继续轮询
  } catch (error) {
    console.error('获取分发进度失败:', error)
    distributionProgress.status = 'exception'
    
    // 根据HTTP状态码显示具体错误信息
    if (error.response) {
      const { status, data } = error.response
      switch (status) {
        case 403:
          ElMessage.error('没有权限查看此分发记录')
          break
        case 404:
          ElMessage.error('分发记录不存在')
          break
        case 500:
          ElMessage.error('服务器内部错误，获取分发进度失败')
          break
        default:
          ElMessage.error(data?.error || `获取分发进度失败 (状态码: ${status})`)
      }
    } else {
      ElMessage.error('网络连接失败，无法获取分发进度')
    }
    
    return true // 停止轮询
  }
}

// 开始轮询
const startPolling = (distributionId) => {
  const poll = async () => {
    const shouldStop = await pollDistributionProgress(distributionId)
    if (!shouldStop && distributionProgress.show) {
      setTimeout(poll, 2000) // 每2秒轮询一次
    }
  }
  
  // 延迟1秒后开始轮询，给后端时间处理
  setTimeout(poll, 1000)
}

const startDistribute = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
  } catch (error) {
    return
  }

  if (selectedHosts.value.length === 0) {
    ElMessage.warning('请至少选择一台主机')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要将文件 "${props.file.original_name}" 分发到 ${selectedHosts.value.length} 台主机吗？`,
      '确认分发',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info',
      }
    )

    distributing.value = true

    const data = {
      file_id: props.file.id,
      host_ids: selectedHosts.value,
      target_path: distributeForm.targetPath,
      description: distributeForm.description
    }

    const response = await api.post(`/api/v1/files/${props.file.id}/distribute`, data)
    const distributionId = response.data.distribution.id
    
    // 初始化进度数据
    distributionProgress.show = true
    distributionProgress.total = selectedHosts.value.length
    distributionProgress.completed = 0
    distributionProgress.status = ''
    distributionProgress.distributionId = distributionId
    distributionProgress.details = []
    
    ElMessage.success('文件分发任务创建成功，开始执行...')
    
    // 开始轮询进度
    startPolling(distributionId)
    
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('创建分发任务失败')
    }
  } finally {
    distributing.value = false
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

// 获取环境类型
const getEnvironmentType = (environment) => {
  const types = {
    production: 'danger',
    test: 'warning',
    development: 'success'
  }
  return types[environment] || 'info'
}

// 获取环境名称
const getEnvironmentName = (environment) => {
  const names = {
    production: '生产',
    test: '测试',
    development: '开发'
  }
  return names[environment] || environment
}

// 关闭对话框
const handleClose = () => {
  if (distributing.value) {
    ElMessage.warning('分发任务正在创建中，请稍候...')
    return
  }

  // 重置数据
  selectedHosts.value = []
  distributeForm.targetPath = ''
  distributeForm.description = ''
  hostFilter.name = ''
  hostFilter.status = ''
  hostFilter.environment = ''
  
  // 重置进度数据
  distributionProgress.show = false
  distributionProgress.total = 0
  distributionProgress.completed = 0
  distributionProgress.status = ''
  distributionProgress.distributionId = null
  distributionProgress.details = []

  visible.value = false
}

// 监听对话框显示状态
watch(visible, (newVal) => {
  if (newVal && props.file) {
    loadHosts()
    // 根据文件类型设置默认路径
    if (props.file.category === 'scripts') {
      distributeForm.targetPath = `/tmp/${props.file.original_name}`
    } else {
      distributeForm.targetPath = `/tmp/${props.file.original_name}`
    }
  }
})

// 监听文件变化
watch(() => props.file, (newFile) => {
  if (newFile && visible.value) {
    // 根据文件类型设置默认路径
    if (newFile.category === 'scripts') {
      distributeForm.targetPath = `/tmp/${newFile.original_name}`
    } else {
      distributeForm.targetPath = `/tmp/${newFile.original_name}`
    }
  }
})
</script>

<style scoped>
.distribute-container {
  max-height: 70vh;
  overflow-y: auto;
}

.file-info-section {
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
  padding: 16px 0;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 16px;
}

.file-icon {
  font-size: 32px;
}

.file-details {
  flex: 1;
}

.file-name {
  font-size: 16px;
  font-weight: 500;
  color: #2c3e50;
  margin-bottom: 8px;
}

.file-meta {
  font-size: 13px;
  color: #7f8c8d;
  margin-bottom: 8px;
}

.file-description {
  font-size: 14px;
  color: #606266;
  background: #f8f9fa;
  padding: 8px 12px;
  border-radius: 4px;
  border-left: 3px solid #409EFF;
}

.distribute-config {
  margin-bottom: 20px;
}

.distribution-progress {
  margin-top: 20px;
}

.progress-info {
  padding: 10px 0;
}

.progress-stats {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  font-size: 14px;
}

.progress-percent {
  font-weight: bold;
  color: #409EFF;
}

.progress-details {
  margin-top: 20px;
  max-height: 200px;
  overflow-y: auto;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  margin-bottom: 8px;
  background: #f8f9fa;
  border-radius: 6px;
  border-left: 3px solid #e9ecef;
}

.detail-item:last-child {
  margin-bottom: 0;
}

.host-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.status-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.time-info {
  font-size: 12px;
  color: #666;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.host-selection {
  margin-bottom: 20px;
}

.host-filter {
  margin-bottom: 16px;
}

.selection-info {
  margin-bottom: 16px;
  text-align: center;
}

.host-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
  max-height: 400px;
  overflow-y: auto;
}

.host-card {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.3s;
  background: #fff;
}

.host-card:hover {
  border-color: #409EFF;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);
}

.host-card.selected {
  border-color: #409EFF;
  background: #f0f9ff;
}

.host-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.host-status {
  display: flex;
  align-items: center;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}

.host-info {
  margin-bottom: 12px;
}

.host-name {
  font-weight: 500;
  color: #2c3e50;
  margin-bottom: 4px;
}

.host-ip {
  font-size: 13px;
  color: #7f8c8d;
  margin-bottom: 4px;
}

.host-os {
  font-size: 12px;
  color: #909399;
}

.host-meta {
  display: flex;
  justify-content: flex-end;
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
</style>
