<template>
  <div class="executions-page">
    <div class="page-header">
      <h2>执行记录</h2>
      <div class="header-actions">
        <el-button @click="refreshExecutions">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <!-- 搜索和过滤 -->
    <div class="search-section">
      <el-input
        v-model="searchText"
        placeholder="搜索作业名称或脚本名称"
        style="width: 300px"
        clearable
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-select v-model="statusFilter" placeholder="状态筛选" style="width: 120px">
        <el-option label="全部" value="" />
        <el-option label="运行中" value="running" />
        <el-option label="已完成" value="completed" />
        <el-option label="失败" value="failed" />
      </el-select>
      <el-select v-model="scriptTypeFilter" placeholder="脚本类型" style="width: 120px">
        <el-option label="全部类型" value="" />
        <el-option label="Shell" value="shell" />
        <el-option label="Python2" value="python2" />
        <el-option label="Python3" value="python3" />
      </el-select>
      <el-checkbox v-model="showQuickExecOnly">仅显示快速执行</el-checkbox>
      <el-date-picker
        v-model="dateRange"
        type="datetimerange"
        range-separator="至"
        start-placeholder="开始时间"
        end-placeholder="结束时间"
        style="width: 350px"
        @change="handleDateChange"
      />
    </div>

    <!-- 执行记录表格 -->
    <div class="table-container">
      <el-table
        :data="filteredExecutions"
        v-loading="loading"
        stripe
        style="width: 100%"
        @row-click="viewExecutionDetail"
      >
        <el-table-column prop="id" label="执行ID" width="80" />
        <el-table-column label="作业信息" min-width="200">
          <template #default="{ row }">
            <div class="job-info">
              <div class="job-name">
                {{ row.job_name || row.job?.name || '未知作业' }}
                <el-tag v-if="row.is_quick_exec" size="small" type="warning">快速执行</el-tag>
              </div>
              <div class="script-name">{{ row.script_name || row.job?.script?.name || '未知脚本' }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="脚本类型" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="getScriptTypeColor(row.script_type)">
              {{ getScriptTypeText(row.script_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="主机" width="180">
          <template #default="{ row }">
            <div class="host-info">
              <div class="host-name">{{ row.host?.name || '未知主机' }}</div>
              <div class="host-ip">{{ row.host?.ip || '' }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="执行人" width="120">
          <template #default="{ row }">
            {{ row.executed_user?.username || '未知用户' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="开始时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.start_time) }}
          </template>
        </el-table-column>
        <el-table-column label="结束时间" width="180">
          <template #default="{ row }">
            {{ row.end_time ? formatDate(row.end_time) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="耗时" width="100">
          <template #default="{ row }">
            {{ getDuration(row) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button
                size="small"
                type="primary"
                @click.stop="viewExecutionDetail(row)"
              >
                查看详情
              </el-button>
              <el-button
                v-if="row.is_quick_exec"
                size="small"
                type="success"
                @click.stop="redoExecution(row)"
              >
                重新执行
              </el-button>
              <el-button
                v-if="!row.is_quick_exec && row.job_id"
                size="small"
                type="success"
                @click.stop="redoJobExecution(row)"
              >
                重新执行
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 分页 -->
    <div class="pagination-container">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="totalExecutions"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- 快速执行对话框 -->
    <QuickExecute 
      v-model:visible="showQuickExecute" 
      :prefill-data="quickExecutePrefillData"
      @executed="handleQuickExecuted"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'
import QuickExecute from '@/components/QuickExecute.vue'

const router = useRouter()
const route = useRoute()

const executions = ref([])
const loading = ref(false)
const searchText = ref('')
const statusFilter = ref('')
const scriptTypeFilter = ref('')
const showQuickExecOnly = ref(false)
const dateRange = ref([])
const currentPage = ref(1)
const pageSize = ref(20)
const totalExecutions = ref(0)
const showQuickExecute = ref(false)
const quickExecutePrefillData = ref(null)

const filteredExecutions = computed(() => {
  let filtered = executions.value

  // 搜索过滤
  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    filtered = filtered.filter(execution => 
      (execution.job_name || execution.job?.name || '').toLowerCase().includes(search) ||
      (execution.script_name || execution.job?.script?.name || '').toLowerCase().includes(search)
    )
  }

  // 状态过滤
  if (statusFilter.value) {
    filtered = filtered.filter(execution => execution.status === statusFilter.value)
  }

  // 脚本类型过滤
  if (scriptTypeFilter.value) {
    filtered = filtered.filter(execution => execution.script_type === scriptTypeFilter.value)
  }

  // 快速执行过滤
  if (showQuickExecOnly.value) {
    filtered = filtered.filter(execution => execution.is_quick_exec)
  }

  // 日期过滤
  if (dateRange.value && dateRange.value.length === 2) {
    const [startDate, endDate] = dateRange.value
    filtered = filtered.filter(execution => {
      const executionDate = new Date(execution.start_time)
      return executionDate >= startDate && executionDate <= endDate
    })
  }

  return filtered
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

const getScriptTypeText = (type) => {
  const typeMap = {
    shell: 'Shell',
    python2: 'Python2',
    python3: 'Python3'
  }
  return typeMap[type] || type
}

const getScriptTypeColor = (type) => {
  const colorMap = {
    shell: 'primary',
    python2: 'success',
    python3: 'warning'
  }
  return colorMap[type] || 'info'
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleString('zh-CN')
}

const getDuration = (execution) => {
  if (!execution.start_time) return '-'
  
  const startTime = new Date(execution.start_time)
  const endTime = execution.end_time ? new Date(execution.end_time) : new Date()
  
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

const loadExecutions = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      size: pageSize.value
    }
    
    // 如果有脚本筛选参数，添加到请求中
    const scriptId = route.query.script_id
    if (scriptId) {
      params.script_id = scriptId
    }
    
    const response = await api.get('/api/v1/executions', { params })
    
    // 后端返回的数据结构是 {data: executions, total, page, size}
    executions.value = response.data.data || []
    totalExecutions.value = response.data.total || 0
  } catch (error) {
    ElMessage.error('加载执行记录失败')
  } finally {
    loading.value = false
  }
}

const refreshExecutions = () => {
  loadExecutions()
}

const viewExecutionDetail = (execution) => {
  router.push(`/executions/${execution.id}`)
}

const redoExecution = (execution) => {
  // 构造预填充数据
  quickExecutePrefillData.value = {
    name: execution.script_name || execution.job_name || '重新执行',
    scriptType: execution.script_type || 'shell',
    scriptContent: execution.script_content || '',
    hostIds: execution.host_id ? [execution.host_id] : [],
    description: `重新执行 - ${execution.script_name || execution.job_name || ''}`
  }
  
  // 打开快速执行对话框
  showQuickExecute.value = true
  
  ElMessage.info('已加载历史执行数据，请确认后执行')
}

const redoJobExecution = async (execution) => {
  try {
    const response = await api.post(`/api/v1/jobs/${execution.job_id}/execute`)
    
    ElMessage.success('作业重新执行已启动')
    
    // 跳转到执行详情页面
    if (response.data.executions && response.data.executions.length > 0) {
      const executionId = response.data.executions[0].id
      router.push(`/executions/${executionId}`)
    } else {
      // 刷新执行记录列表
      loadExecutions()
    }
  } catch (error) {
    ElMessage.error('作业重新执行失败: ' + (error.response?.data?.error || error.message))
  }
}

const handleQuickExecuted = (executions) => {
  ElMessage.success(`快速执行已启动，共创建 ${executions.length} 个执行任务`)
  
  // 刷新执行记录列表
  loadExecutions()
}

const handleDateChange = () => {
  // 日期变化时重新加载数据
  loadExecutions()
}

const handleSizeChange = (newSize) => {
  pageSize.value = newSize
  currentPage.value = 1
  loadExecutions()
}

const handleCurrentChange = (newPage) => {
  currentPage.value = newPage
  loadExecutions()
}

// 监听搜索和过滤条件变化
watch([searchText, statusFilter, scriptTypeFilter, showQuickExecOnly], () => {
  currentPage.value = 1
})

onMounted(() => {
  loadExecutions()
})
</script>

<style scoped>
.executions-page {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h2 {
  margin: 0;
  color: #2c3e50;
  font-size: 24px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.search-section {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
  align-items: center;
  flex-wrap: wrap;
}

.table-container {
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.job-info .job-name {
  font-weight: 600;
  color: #2c3e50;
  margin-bottom: 4px;
}

.job-info .script-name {
  font-size: 12px;
  color: #7f8c8d;
}

.host-info .host-name {
  font-weight: 500;
  color: #2c3e50;
  margin-bottom: 2px;
}

.host-info .host-ip {
  font-size: 12px;
  color: #7f8c8d;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

:deep(.el-table tbody tr) {
  cursor: pointer;
}

:deep(.el-table tbody tr:hover) {
  background-color: #f5f7fa;
}
</style>
