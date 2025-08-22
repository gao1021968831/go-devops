<template>
  <div class="job-manager">
    <!-- 搜索和过滤 -->
    <div class="search-section">
      <div class="search-left">
        <el-input
          v-model="searchText"
          placeholder="搜索作业名称..."
          class="search-input"
          clearable
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-select v-model="statusFilter" placeholder="状态筛选" class="type-filter">
          <el-option label="全部状态" value="" />
          <el-option label="手动执行" value="manual" />
          <el-option label="定时执行" value="scheduled" />
        </el-select>
      </div>
      <div class="search-right">
        <el-button @click="refreshJobs" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <!-- 作业列表 -->
    <div v-if="filteredJobs.length === 0" class="empty-state">
      <el-empty description="暂无作业数据">
        <el-text>通过脚本管理页面创建作业</el-text>
      </el-empty>
    </div>

    <el-table v-else :data="filteredJobs" style="width: 100%" stripe>
      <el-table-column prop="name" label="作业名称" min-width="150">
        <template #default="{ row }">
          <div class="job-name">
            <strong>{{ row.name }}</strong>
            <div class="job-description">{{ row.description || '无描述' }}</div>
          </div>
        </template>
      </el-table-column>
      
      <el-table-column label="执行类型" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="row.execute_type === 'scheduled' ? 'warning' : 'info'" size="small">
            {{ row.execute_type === 'scheduled' ? '定时执行' : '手动执行' }}
          </el-tag>
        </template>
      </el-table-column>
      
      <el-table-column label="关联脚本" min-width="150">
        <template #default="{ row }">
          <div class="script-info">
            <el-tag :type="getTypeColor(row.script?.type)" size="small">
              {{ (row.script?.type || '').toUpperCase() }}
            </el-tag>
            <span class="script-name">{{ row.script?.name || '脚本已删除' }}</span>
          </div>
        </template>
      </el-table-column>
      
      <el-table-column label="目标主机" width="100" align="center">
        <template #default="{ row }">
          <span>{{ getHostCount(row.host_ids) }} 台</span>
        </template>
      </el-table-column>
      
      <el-table-column label="计划时间" width="160" align="center">
        <template #default="{ row }">
          <span v-if="row.scheduled_time">{{ formatDateTime(row.scheduled_time) }}</span>
          <span v-else class="text-muted">-</span>
        </template>
      </el-table-column>
      
      <el-table-column prop="user.username" label="创建者" width="100" align="center">
        <template #default="{ row }">
          <span>{{ row.user?.username || '未知' }}</span>
        </template>
      </el-table-column>
      
      <el-table-column prop="created_at" label="创建时间" width="160" align="center">
        <template #default="{ row }">
          <span>{{ formatDate(row.created_at) }}</span>
        </template>
      </el-table-column>
      
      <el-table-column label="操作" width="260" align="center" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="success" @click="executeJob(row)">
            <el-icon><VideoPlay /></el-icon>
            执行
          </el-button>
          <el-button size="small" type="info" @click="viewJobExecutions(row)">
            <el-icon><List /></el-icon>
            历史
          </el-button>
          <el-dropdown trigger="click" @command="(command) => handleJobCommand(command, row)">
            <el-button size="small">
              <el-icon><More /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="edit">
                  <el-icon><Edit /></el-icon>
                  编辑作业
                </el-dropdown-item>
                <el-dropdown-item command="duplicate">
                  <el-icon><CopyDocument /></el-icon>
                  复制作业
                </el-dropdown-item>
                <el-dropdown-item divided command="delete" class="danger-item">
                  <el-icon><Delete /></el-icon>
                  删除作业
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-table-column>
    </el-table>

    <!-- 作业表单对话框 -->
    <JobForm
      v-model:visible="showFormDialog"
      :job="editingJob"
      :script="selectedScript"
      @saved="handleJobSaved"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import api from '@/utils/api'
import JobForm from './JobForm.vue'
import {
  Search,
  Refresh,
  Monitor,
  Clock,
  User,
  VideoPlay,
  List,
  More,
  Edit,
  CopyDocument,
  Delete
} from '@element-plus/icons-vue'

const router = useRouter()

// 响应式数据
const jobs = ref([])
const searchText = ref('')
const statusFilter = ref('')
const loading = ref(false)
const showFormDialog = ref(false)
const editingJob = ref(null)
const selectedScript = ref(null)

// 计算属性
const filteredJobs = computed(() => {
  // 确保jobs.value是数组
  if (!Array.isArray(jobs.value)) {
    return []
  }
  
  return jobs.value.filter(job => {
    const matchesSearch = !searchText.value || 
      job.name.toLowerCase().includes(searchText.value.toLowerCase())
    
    const matchesStatus = !statusFilter.value || job.execute_type === statusFilter.value
    
    return matchesSearch && matchesStatus
  })
})

// 工具函数
const getTypeColor = (type) => {
  const colorMap = {
    shell: 'success',
    python2: 'warning',
    python3: 'info'
  }
  return colorMap[type] || 'primary'
}

const getHostCount = (hostIds) => {
  if (!hostIds) return 0
  try {
    return JSON.parse(hostIds).length
  } catch {
    return 0
  }
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleString('zh-CN')
}

const formatDateTime = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// API 方法
const loadJobs = async () => {
  loading.value = true
  try {
    const response = await api.get('/api/v1/jobs')
    // 后端返回的数据结构是 {data: jobs, total, page, size}
    jobs.value = Array.isArray(response.data.data) ? response.data.data : []
  } catch (error) {
    ElMessage.error('加载作业列表失败')
    jobs.value = [] // 出错时也要确保是数组
  } finally {
    loading.value = false
  }
}

const refreshJobs = () => {
  loadJobs()
}

const refreshExecutions = () => {
  // 刷新执行记录，如果当前在执行记录页面
  loadJobs()
}

// 作业操作方法
const executeJob = async (job) => {
  try {
    await ElMessageBox.confirm(
      `确定要执行作业 "${job.name}" 吗？`,
      '确认执行',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info'
      }
    )
    
    const response = await api.post(`/api/v1/jobs/${job.id}/execute`)
    ElMessage.success('作业执行请求已提交')
    
    // 跳转到执行详情页面
    if (response.data.executions && response.data.executions.length > 0) {
      const executionId = response.data.executions[0].id
      router.push(`/executions/${executionId}`)
    } else {
      // 如果没有返回执行记录，跳转到执行记录页面
      router.push({
        path: '/executions',
        query: { job_id: job.id, job_name: job.name }
      })
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('执行作业失败')
    }
  }
}

const viewJobExecutions = (job) => {
  router.push({
    path: '/executions',
    query: { job_id: job.id, job_name: job.name }
  })
}

const handleJobCommand = (command, job) => {
  switch (command) {
    case 'edit':
      editJob(job)
      break
    case 'duplicate':
      duplicateJob(job)
      break
    case 'delete':
      deleteJob(job)
      break
  }
}

const editJob = (job) => {
  editingJob.value = job
  selectedScript.value = job.script
  showFormDialog.value = true
}

const duplicateJob = async (job) => {
  try {
    const newJob = {
      name: `${job.name} - 副本`,
      script_id: job.script_id,
      host_ids: job.host_ids,
      execute_type: job.execute_type,
      scheduled_time: job.scheduled_time,
      description: job.description
    }
    
    await api.post('/api/v1/jobs', newJob)
    ElMessage.success('作业复制成功')
    loadJobs()
  } catch (error) {
    ElMessage.error('复制作业失败')
  }
}

const deleteJob = async (job) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除作业 "${job.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await api.delete(`/api/v1/jobs/${job.id}`)
    ElMessage.success('作业删除成功')
    loadJobs()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除作业失败')
    }
  }
}

const handleJobSaved = () => {
  loadJobs()
}

// 生命周期
onMounted(() => {
  loadJobs()
})

// 暴露方法给父组件
defineExpose({
  createJobForScript: (script) => {
    editingJob.value = null
    selectedScript.value = script
    showFormDialog.value = true
  },
  refreshExecutions
})
</script>

<style scoped>
.job-manager {
  height: 100%;
}

.search-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  gap: 16px;
}

.search-left {
  display: flex;
  gap: 12px;
  flex: 1;
}

.search-input {
  width: 300px;
}

.type-filter {
  width: 150px;
}

.search-right {
  display: flex;
  gap: 12px;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}

.job-name strong {
  color: #303133;
  font-size: 14px;
  font-weight: 600;
}

.job-description {
  color: #909399;
  font-size: 12px;
  margin-top: 4px;
  line-height: 1.4;
}

.script-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.script-name {
  color: #606266;
  font-size: 13px;
}

.text-muted {
  color: #909399;
}

.job-card {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 20px;
  transition: all 0.3s ease;
}

.job-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.job-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.job-info h3 {
  margin: 0 0 4px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.job-info p {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.job-content {
  margin: 12px 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.job-script {
  display: flex;
  align-items: center;
  gap: 8px;
}

.script-name {
  font-size: 14px;
  color: #606266;
}

.job-hosts,
.job-schedule {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: #909399;
}

.job-meta {
  display: flex;
  gap: 16px;
  margin: 12px 0;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #909399;
}

.job-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #f0f2f5;
}

.action-group {
  flex: 1;
}

.more-actions {
  margin-left: 12px;
}

.danger-item {
  color: #f56c6c !important;
}
</style>
