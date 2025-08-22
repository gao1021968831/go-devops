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
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'

const route = useRoute()
const router = useRouter()

const execution = ref(null)
const relatedExecutions = ref([])
const loading = ref(false)

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
  } catch (error) {
    ElMessage.error('加载执行详情失败')
  } finally {
    loading.value = false
  }
}

const refreshDetail = () => {
  loadExecutionDetail(route.params.id)
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

// 监听路由参数变化
watch(() => route.params.id, (newId) => {
  if (newId) {
    loadExecutionDetail(newId)
  }
}, { immediate: true })

onMounted(() => {
  if (route.params.id) {
    loadExecutionDetail(route.params.id)
  }
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
  gap: 12px;
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
