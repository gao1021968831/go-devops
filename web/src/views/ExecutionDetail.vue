<template>
  <div class="execution-detail-page">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="goBack" text>
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h2>执行详情</h2>
      </div>
      <div class="header-actions">
        <div class="auto-refresh-control">
          <el-switch
            v-model="autoRefresh"
            active-text="自动刷新"
            inactive-text=""
            @change="toggleAutoRefresh"
          />
          <span v-if="autoRefresh && countdown > 0" class="countdown">
            {{ countdown }}s后刷新
          </span>
        </div>
        <el-button @click="refreshDetail">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button 
          v-if="execution && execution.is_quick_exec"
          type="success" 
          @click="redoQuickExecution"
        >
          <el-icon><Refresh /></el-icon>
          重新执行
        </el-button>
        <el-button 
          v-if="execution && !execution.is_quick_exec && execution.job_id"
          type="success" 
          @click="redoJobExecution"
        >
          <el-icon><Refresh /></el-icon>
          重新执行
        </el-button>
      </div>
    </div>

    <div v-if="execution" class="execution-content">
      <!-- 基本信息 -->
      <el-card class="info-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>基本信息</span>
            <el-tag :type="getStatusType(execution.status)" size="large">
              {{ getStatusText(execution.status) }}
            </el-tag>
          </div>
        </template>
        
        <div class="info-grid">
          <div class="info-item">
            <label>执行ID:</label>
            <span>{{ execution.id }}</span>
          </div>
          <div class="info-item">
            <label>作业名称:</label>
            <span>{{ execution.job_name || execution.job?.name || '未知作业' }}</span>
          </div>
          <div class="info-item">
            <label>脚本名称:</label>
            <span>{{ execution.script_name || execution.job?.script?.name || '未知脚本' }}</span>
          </div>
          <div class="info-item">
            <label>脚本类型:</label>
            <el-tag :type="getScriptTypeColor(execution.script_type || execution.job?.script?.type)" size="small">
              {{ (execution.script_type || execution.job?.script?.type || '').toUpperCase() }}
            </el-tag>
          </div>
          <div class="info-item">
            <label>目标主机:</label>
            <span>{{ execution.host?.name || '未知主机' }} ({{ execution.host?.ip || '' }})</span>
          </div>
          <div class="info-item">
            <label>执行人:</label>
            <span>{{ execution.executed_user?.username || '未知用户' }}</span>
          </div>
          <div class="info-item">
            <label>开始时间:</label>
            <span>{{ formatDate(execution.start_time) }}</span>
          </div>
          <div class="info-item">
            <label>结束时间:</label>
            <span>{{ execution.end_time ? formatDate(execution.end_time) : '运行中' }}</span>
          </div>
          <div class="info-item">
            <label>执行耗时:</label>
            <span>{{ getDuration(execution) }}</span>
          </div>
        </div>
      </el-card>

      <!-- 脚本内容 -->
      <el-card class="script-card" shadow="never">
        <template #header>
          <span>脚本内容</span>
        </template>
        <div class="script-content">
          <pre>{{ execution.script_content || execution.job?.script?.content || '无脚本内容' }}</pre>
        </div>
      </el-card>

      <!-- 执行输出 -->
      <el-card class="output-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>执行输出</span>
            <div class="output-actions">
              <el-button size="small" @click="copyOutput" v-if="execution.output">
                <el-icon><DocumentCopy /></el-icon>
                复制输出
              </el-button>
              <el-button 
                size="small" 
                type="primary" 
                @click="showSaveResultDialog = true"
                v-if="execution.output || execution.error"
              >
                <el-icon><Download /></el-icon>
                保存为文件
              </el-button>
            </div>
          </div>
        </template>
        <div class="output-content">
          <pre v-if="execution.output" class="output-text">{{ execution.output }}</pre>
          <div v-else class="no-output">暂无输出内容</div>
        </div>
      </el-card>

      <!-- 错误信息 -->
      <el-card v-if="execution.error" class="error-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>错误信息</span>
            <el-button size="small" @click="copyError">
              <el-icon><DocumentCopy /></el-icon>
              复制错误
            </el-button>
          </div>
        </template>
        <div class="error-content">
          <pre class="error-text">{{ execution.error }}</pre>
        </div>
      </el-card>

      <!-- 同批次其他执行记录 -->
      <el-card v-if="relatedExecutions.length > 1" class="related-card" shadow="never">
        <template #header>
          <span>同批次执行记录</span>
        </template>
        <el-table :data="relatedExecutions" size="small">
          <el-table-column prop="id" label="执行ID" width="80" />
          <el-table-column label="主机" width="200">
            <template #default="{ row }">
              <span>{{ row.host?.name || '未知主机' }} ({{ row.host?.ip || '' }})</span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)" size="small">
                {{ getStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="耗时" width="100">
            <template #default="{ row }">
              {{ getDuration(row) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <el-button
                v-if="row.id !== execution.id"
                size="small"
                type="primary"
                @click="switchExecution(row.id)"
              >
                查看
              </el-button>
              <el-tag v-else type="info" size="small">当前</el-tag>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <div v-else-if="!loading" class="not-found">
      <el-empty description="执行记录不存在" />
    </div>

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="8" animated />
    </div>

    <!-- 保存结果为文件对话框 -->
    <el-dialog
      v-model="showSaveResultDialog"
      title="保存执行结果为文件"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form :model="saveResultForm" label-width="100px">
        <el-form-item label="保存内容">
          <el-checkbox-group v-model="saveResultForm.saveTypes">
            <el-checkbox 
              label="output" 
              :disabled="!execution.output"
            >
              执行输出 {{ execution.output ? `(${execution.output.length} 字符)` : '(无内容)' }}
            </el-checkbox>
            <el-checkbox 
              label="error" 
              :disabled="!execution.error"
            >
              错误日志 {{ execution.error ? `(${execution.error.length} 字符)` : '(无内容)' }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        
        <el-form-item label="文件分类">
          <el-select v-model="saveResultForm.outputCategory" placeholder="选择文件分类">
            <el-option label="脚本输出" value="script_output" />
            <el-option label="日志文件" value="log" />
            <el-option label="报告文件" value="report" />
            <el-option label="通用文件" value="general" />
          </el-select>
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showSaveResultDialog = false">取消</el-button>
          <el-button 
            type="primary" 
            @click="saveExecutionResult"
            :loading="savingResult"
            :disabled="saveResultForm.saveTypes.length === 0"
          >
            保存文件
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Refresh, DocumentCopy, Download } from '@element-plus/icons-vue'
import api from '@/utils/api'

const route = useRoute()
const router = useRouter()

const execution = ref(null)
const relatedExecutions = ref([])
const loading = ref(false)
const showSaveResultDialog = ref(false)
const savingResult = ref(false)

// 自动刷新相关
const autoRefresh = ref(false)
const countdown = ref(0)
const refreshInterval = ref(null)
const countdownInterval = ref(null)
const REFRESH_SECONDS = 5

const saveResultForm = ref({
  saveTypes: [],
  outputCategory: 'script_output'
})

const getStatusType = (status) => {
  const typeMap = {
    running: 'warning',
    completed: 'success',
    failed: 'danger'
  }
  return typeMap[status] || 'info'
}

const getStatusText = (status) => {
  const textMap = {
    running: '运行中',
    completed: '已完成',
    failed: '失败'
  }
  return textMap[status] || status
}

const getScriptTypeColor = (type) => {
  const colorMap = {
    shell: 'success',
    python2: 'warning',
    python3: 'info'
  }
  return colorMap[type] || 'primary'
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleString('zh-CN')
}

const getDuration = (exec) => {
  if (!exec.start_time) return '-'
  
  const startTime = new Date(exec.start_time)
  const endTime = exec.end_time ? new Date(exec.end_time) : new Date()
  
  const duration = Math.floor((endTime - startTime) / 1000) // 秒
  
  if (duration < 60) {
    return `${duration}秒`
  } else if (duration < 3600) {
    return `${Math.floor(duration / 60)}分${duration % 60}秒`
  } else {
    const hours = Math.floor(duration / 3600)
    const minutes = Math.floor((duration % 3600) / 60)
    return `${hours}时${minutes}分`
  }
}

const loadExecutionDetail = async (executionId) => {
  loading.value = true
  try {
    const response = await api.get(`/api/v1/executions/${executionId}`)
    execution.value = response.data
    
    // 加载同批次执行记录
    if (execution.value.job_id) {
      const relatedResponse = await api.get(`/api/v1/jobs/${execution.value.job_id}/executions`)
      relatedExecutions.value = relatedResponse.data
    }
    
    // 如果执行状态是运行中，默认开启自动刷新
    if (execution.value.status === 'running' && !autoRefresh.value) {
      autoRefresh.value = true
      startAutoRefresh()
    }
  } catch (error) {
    ElMessage.error('加载执行详情失败')
  } finally {
    loading.value = false
  }
}

const refreshDetail = () => {
  loadExecutionDetail(route.params.id)
}

// 自动刷新功能
const startAutoRefresh = () => {
  if (refreshInterval.value) return
  
  countdown.value = REFRESH_SECONDS
  
  // 启动倒计时
  countdownInterval.value = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      refreshDetail()
      countdown.value = REFRESH_SECONDS
    }
  }, 1000)
}

const stopAutoRefresh = () => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
    refreshInterval.value = null
  }
  if (countdownInterval.value) {
    clearInterval(countdownInterval.value)
    countdownInterval.value = null
  }
  countdown.value = 0
}

const toggleAutoRefresh = (enabled) => {
  if (enabled) {
    // 如果执行状态是运行中，启动自动刷新
    if (execution.value?.status === 'running') {
      startAutoRefresh()
    } else {
      // 如果执行已完成，提示用户
      ElMessage.info('当前执行已完成，无需自动刷新')
      autoRefresh.value = false
    }
  } else {
    stopAutoRefresh()
  }
}

const goBack = () => {
  router.back()
}

const switchExecution = (executionId) => {
  router.push(`/executions/${executionId}`)
}

const copyOutput = async () => {
  try {
    await navigator.clipboard.writeText(execution.value.output)
    ElMessage.success('输出内容已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

const copyError = async () => {
  try {
    await navigator.clipboard.writeText(execution.value.error)
    ElMessage.success('错误信息已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

const saveExecutionResult = async () => {
  if (saveResultForm.value.saveTypes.length === 0) {
    ElMessage.warning('请选择要保存的内容')
    return
  }

  savingResult.value = true
  try {
    const requestData = {
      execution_id: execution.value.id,
      save_output: saveResultForm.value.saveTypes.includes('output'),
      save_error: saveResultForm.value.saveTypes.includes('error'),
      output_category: saveResultForm.value.outputCategory
    }

    await api.post('/api/v1/job-executions/save-result', requestData)
    ElMessage.success('执行结果已保存为文件')
    showSaveResultDialog.value = false
    
    // 重新加载执行详情以获取文件链接
    await loadExecutionDetail(execution.value.id)
  } catch (error) {
    ElMessage.error('保存文件失败: ' + (error.response?.data?.error || error.message))
  } finally {
    savingResult.value = false
  }
}

// 重置保存表单
const resetSaveForm = () => {
  saveResultForm.value = {
    saveTypes: [],
    outputCategory: 'script_output'
  }
}

const redoQuickExecution = () => {
  // 跳转到脚本页面并触发快速执行对话框
  const prefillData = {
    name: execution.value.script_name || execution.value.job_name || '重新执行',
    scriptType: execution.value.script_type || 'shell',
    scriptContent: execution.value.script_content || '',
    hostIds: execution.value.host_id ? [execution.value.host_id] : [],
    description: `重新执行 - ${execution.value.script_name || execution.value.job_name || '快速执行脚本'}`
  }
  
  // 跳转到脚本页面，并通过 query 参数传递预填充数据
  router.push({
    path: '/scripts',
    query: {
      action: 'quick-execute',
      prefill: JSON.stringify(prefillData)
    }
  })
}

const redoJobExecution = async () => {
  try {
    const response = await api.post(`/api/v1/jobs/${execution.value.job_id}/execute`)
    
    ElMessage.success('作业重新执行已启动')
    
    // 跳转到新的执行详情页面
    if (response.data.executions && response.data.executions.length > 0) {
      const executionId = response.data.executions[0].id
      router.push(`/executions/${executionId}`)
    }
  } catch (error) {
    ElMessage.error('作业重新执行失败: ' + (error.response?.data?.error || error.message))
  }
}

// 监听执行状态变化，自动管理刷新
watch(() => execution.value?.status, (newStatus, oldStatus) => {
  if (autoRefresh.value) {
    if (newStatus === 'running') {
      // 如果变为运行中，确保自动刷新启动
      if (!countdownInterval.value) {
        startAutoRefresh()
      }
    } else if (oldStatus === 'running' && (newStatus === 'completed' || newStatus === 'failed')) {
      // 如果从运行中变为完成或失败，停止自动刷新
      stopAutoRefresh()
      autoRefresh.value = false
      ElMessage.success(`执行${newStatus === 'completed' ? '完成' : '失败'}，已停止自动刷新`)
    }
  }
})

// 监听路由参数变化
watch(() => route.params.id, (newId) => {
  if (newId) {
    // 切换执行记录时停止自动刷新
    stopAutoRefresh()
    autoRefresh.value = false
    loadExecutionDetail(newId)
  }
}, { immediate: true })

// 监听对话框显示状态，重置表单
watch(showSaveResultDialog, (visible) => {
  if (visible) {
    resetSaveForm()
    // 根据执行结果自动选择保存类型
    if (execution.value?.output) {
      saveResultForm.value.saveTypes.push('output')
    }
    if (execution.value?.error) {
      saveResultForm.value.saveTypes.push('error')
    }
  }
})

onMounted(() => {
  if (route.params.id) {
    loadExecutionDetail(route.params.id)
  }
})

// 组件卸载时清理定时器
onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.execution-detail-page {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-left h2 {
  margin: 0;
  color: #2c3e50;
  font-size: 24px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.auto-refresh-control {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 6px;
}

.countdown {
  font-size: 12px;
  color: #409eff;
  font-weight: 500;
  min-width: 60px;
}

.execution-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.info-card, .script-card, .output-card, .error-card, .related-card {
  border: 1px solid #f0f0f0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 16px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-item label {
  font-weight: 600;
  color: #606266;
  min-width: 80px;
}

.script-content {
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 6px;
  padding: 16px;
  max-height: 300px;
  overflow-y: auto;
}

.script-content pre {
  margin: 0;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  color: #495057;
  white-space: pre-wrap;
  word-break: break-all;
}

.output-content {
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 6px;
  padding: 16px;
  max-height: 400px;
  overflow-y: auto;
}

.output-text {
  margin: 0;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  color: #495057;
  white-space: pre-wrap;
  word-break: break-all;
}

.no-output {
  color: #909399;
  text-align: center;
  padding: 20px;
}

.error-content {
  background: #fef0f0;
  border: 1px solid #fbc4c4;
  border-radius: 6px;
  padding: 16px;
  max-height: 300px;
  overflow-y: auto;
}

.error-text {
  margin: 0;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  color: #f56c6c;
  white-space: pre-wrap;
  word-break: break-all;
}

.output-actions {
  display: flex;
  gap: 8px;
}

.not-found {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 300px;
}

.loading-container {
  padding: 24px;
}
</style>
