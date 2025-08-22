<template>
  <div class="execution-history">
    <div class="history-header">
      <div class="search-section">
        <el-input
          v-model="searchQuery"
          placeholder="搜索执行历史..."
          clearable
          @input="handleSearch"
          style="width: 300px"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        
        <el-select
          v-model="statusFilter"
          placeholder="执行状态"
          clearable
          style="width: 150px; margin-left: 12px"
          @change="handleSearch"
        >
          <el-option label="全部" value="" />
          <el-option label="运行中" value="running" />
          <el-option label="已完成" value="completed" />
          <el-option label="失败" value="failed" />
        </el-select>
        
        <el-select
          v-model="typeFilter"
          placeholder="脚本类型"
          clearable
          style="width: 150px; margin-left: 12px"
          @change="handleSearch"
        >
          <el-option label="全部" value="" />
          <el-option label="Shell" value="shell" />
          <el-option label="Python2" value="python2" />
          <el-option label="Python3" value="python3" />
        </el-select>
        
        <el-button @click="refreshHistory" style="margin-left: 12px">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <div class="history-content" v-loading="loading">
      <div v-if="executions.length === 0" class="empty-state">
        <el-empty description="暂无执行历史" />
      </div>
      
      <div v-else class="execution-list">
        <div
          v-for="execution in executions"
          :key="execution.id"
          class="execution-item"
          :class="{ 'quick-exec': execution.is_quick_exec }"
        >
          <div class="execution-header">
            <div class="execution-info">
              <div class="execution-title">
                <el-tag v-if="execution.is_quick_exec" type="success" size="small">快速执行</el-tag>
                <span class="name">{{ execution.script_name || '未知脚本' }}</span>
                <el-tag :type="getStatusType(execution.status)" size="small">
                  {{ getStatusText(execution.status) }}
                </el-tag>
              </div>
              
              <div class="execution-meta">
                <span class="meta-item">
                  <el-icon><Monitor /></el-icon>
                  {{ execution.host_name }}
                </span>
                <span class="meta-item">
                  <el-icon><Document /></el-icon>
                  {{ execution.script_type }}
                </span>
                <span class="meta-item">
                  <el-icon><Clock /></el-icon>
                  {{ formatTime(execution.created_at) }}
                </span>
                <span v-if="execution.duration" class="meta-item">
                  <el-icon><Timer /></el-icon>
                  {{ formatDuration(execution.duration) }}
                </span>
              </div>
            </div>
            
            <div class="execution-actions">
              <el-button-group size="small">
                <el-button @click="viewDetails(execution)">
                  <el-icon><View /></el-icon>
                  详情
                </el-button>
                <el-button 
                  v-if="execution.script_content" 
                  @click="redoExecution(execution)"
                  :disabled="execution.status === 'running'"
                >
                  <el-icon><RefreshRight /></el-icon>
                  重做
                </el-button>
                <el-dropdown @command="handleCommand" trigger="click">
                  <el-button>
                    <el-icon><More /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item :command="`logs-${execution.id}`">
                        <el-icon><Document /></el-icon>
                        查看日志
                      </el-dropdown-item>
                      <el-dropdown-item 
                        v-if="execution.script_content" 
                        :command="`copy-${execution.id}`"
                      >
                        <el-icon><CopyDocument /></el-icon>
                        复制脚本
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </el-button-group>
            </div>
          </div>
          
          <div v-if="execution.script_content" class="script-preview">
            <div class="script-header">
              <span>脚本内容预览</span>
              <el-button 
                text 
                size="small" 
                @click="toggleScriptExpand(execution.id)"
              >
                {{ expandedScripts.has(execution.id) ? '收起' : '展开' }}
              </el-button>
            </div>
            <div 
              v-show="expandedScripts.has(execution.id)" 
              class="script-content"
            >
              <pre>{{ execution.script_content }}</pre>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div class="pagination-section" v-if="total > 0">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- 执行详情对话框 -->
    <el-dialog
      v-model="showDetailsDialog"
      title="执行详情"
      width="800px"
      :close-on-click-modal="false"
    >
      <div v-if="selectedExecution" class="execution-details">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="执行ID">
            {{ selectedExecution.id }}
          </el-descriptions-item>
          <el-descriptions-item label="执行状态">
            <el-tag :type="getStatusType(selectedExecution.status)">
              {{ getStatusText(selectedExecution.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="目标主机">
            {{ selectedExecution.host_name }}
          </el-descriptions-item>
          <el-descriptions-item label="脚本类型">
            {{ selectedExecution.script_type }}
          </el-descriptions-item>
          <el-descriptions-item label="开始时间">
            {{ formatTime(selectedExecution.created_at) }}
          </el-descriptions-item>
          <el-descriptions-item label="结束时间">
            {{ selectedExecution.finished_at ? formatTime(selectedExecution.finished_at) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="执行用户">
            {{ selectedExecution.executed_by || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="执行时长">
            {{ selectedExecution.duration ? formatDuration(selectedExecution.duration) : '-' }}
          </el-descriptions-item>
        </el-descriptions>
        
        <div class="output-section" style="margin-top: 20px">
          <h4>执行输出</h4>
          <el-input
            v-model="selectedExecution.output"
            type="textarea"
            :rows="8"
            readonly
            placeholder="暂无输出"
            class="output-textarea"
          />
        </div>
        
        <div v-if="selectedExecution.error" class="error-section" style="margin-top: 20px">
          <h4>错误信息</h4>
          <el-input
            v-model="selectedExecution.error"
            type="textarea"
            :rows="4"
            readonly
            class="error-textarea"
          />
        </div>
      </div>
      
      <template #footer>
        <el-button @click="showDetailsDialog = false">关闭</el-button>
        <el-button 
          v-if="selectedExecution?.script_content" 
          type="primary" 
          @click="redoExecution(selectedExecution)"
        >
          重做执行
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/utils/api'
import {
  Search,
  Refresh,
  Monitor,
  Document,
  Clock,
  Timer,
  View,
  RefreshRight,
  More,
  CopyDocument
} from '@element-plus/icons-vue'

const emit = defineEmits(['redo-execution'])

// 响应式数据
const loading = ref(false)
const executions = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

const searchQuery = ref('')
const statusFilter = ref('')
const typeFilter = ref('')

const showDetailsDialog = ref(false)
const selectedExecution = ref(null)
const expandedScripts = ref(new Set())

// 计算属性
const filteredExecutions = computed(() => {
  return executions.value.filter(execution => {
    const matchesSearch = !searchQuery.value || 
      execution.script_name?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      execution.host_name?.toLowerCase().includes(searchQuery.value.toLowerCase())
    
    const matchesStatus = !statusFilter.value || execution.status === statusFilter.value
    const matchesType = !typeFilter.value || execution.script_type === typeFilter.value
    
    return matchesSearch && matchesStatus && matchesType
  })
})

// 方法
const loadExecutions = async () => {
  loading.value = true
  try {
    const response = await api.get('/api/v1/executions', {
      params: {
        page: currentPage.value,
        size: pageSize.value,
        search: searchQuery.value,
        status: statusFilter.value,
        script_type: typeFilter.value
      }
    })
    executions.value = response.data.data || []
    total.value = response.data.total || 0
  } catch (error) {
    ElMessage.error('加载执行历史失败')
  } finally {
    loading.value = false
  }
}

const refreshHistory = () => {
  loadExecutions()
}

const handleSearch = () => {
  currentPage.value = 1
  loadExecutions()
}

const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  loadExecutions()
}

const handleCurrentChange = (page) => {
  currentPage.value = page
  loadExecutions()
}

const getStatusType = (status) => {
  const types = {
    running: 'warning',
    completed: 'success',
    failed: 'danger'
  }
  return types[status] || 'info'
}

const getStatusText = (status) => {
  const texts = {
    running: '运行中',
    completed: '已完成',
    failed: '失败'
  }
  return texts[status] || status
}

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const formatDuration = (seconds) => {
  if (!seconds) return '-'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60
  
  if (hours > 0) {
    return `${hours}时${minutes}分${secs}秒`
  } else if (minutes > 0) {
    return `${minutes}分${secs}秒`
  } else {
    return `${secs}秒`
  }
}

const viewDetails = (execution) => {
  selectedExecution.value = execution
  showDetailsDialog.value = true
}

const redoExecution = async (execution) => {
  try {
    await ElMessageBox.confirm(
      `确定要重新执行脚本 "${execution.script_name}" 吗？`,
      '确认重做',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info'
      }
    )
    
    emit('redo-execution', {
      name: `重做-${execution.script_name}`,
      scriptType: execution.script_type,
      scriptContent: execution.script_content,
      hostIds: [execution.host_id],
      description: `重做执行 #${execution.id}`
    })
    
    showDetailsDialog.value = false
    
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('重做执行失败')
    }
  }
}

const toggleScriptExpand = (executionId) => {
  if (expandedScripts.value.has(executionId)) {
    expandedScripts.value.delete(executionId)
  } else {
    expandedScripts.value.add(executionId)
  }
}

const handleCommand = (command) => {
  const [action, id] = command.split('-')
  const execution = executions.value.find(e => e.id == id)
  
  if (!execution) return
  
  switch (action) {
    case 'logs':
      viewDetails(execution)
      break
    case 'copy':
      navigator.clipboard.writeText(execution.script_content)
      ElMessage.success('脚本内容已复制到剪贴板')
      break
  }
}

// 生命周期
onMounted(() => {
  loadExecutions()
})

// 暴露方法
defineExpose({
  refreshHistory,
  loadExecutions
})
</script>

<style scoped>
.execution-history {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.history-header {
  margin-bottom: 20px;
}

.search-section {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.history-content {
  flex: 1;
  overflow-y: auto;
}

.execution-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.execution-item {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
  background: #fff;
  transition: all 0.3s;
}

.execution-item:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.execution-item.quick-exec {
  border-left: 4px solid #67c23a;
}

.execution-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.execution-info {
  flex: 1;
}

.execution-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.execution-title .name {
  font-weight: 600;
  font-size: 16px;
  color: #303133;
}

.execution-meta {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #909399;
}

.execution-actions {
  flex-shrink: 0;
}

.script-preview {
  margin-top: 16px;
  border-top: 1px solid #f0f2f5;
  padding-top: 16px;
}

.script-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 14px;
  font-weight: 600;
  color: #606266;
}

.script-content {
  background: #f5f7fa;
  border-radius: 4px;
  padding: 12px;
  max-height: 200px;
  overflow-y: auto;
}

.script-content pre {
  margin: 0;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
}

.pagination-section {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

.execution-details .output-textarea :deep(.el-textarea__inner) {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  background: #f5f7fa;
}

.execution-details .error-textarea :deep(.el-textarea__inner) {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  background: #fef0f0;
  color: #f56c6c;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 300px;
}
</style>
