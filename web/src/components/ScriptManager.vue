<template>
  <div class="script-manager">
    <!-- 搜索和过滤 -->
    <div class="search-section">
      <div class="search-left">
        <el-input
          v-model="searchText"
          placeholder="搜索脚本名称或描述..."
          class="search-input"
          clearable
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-select v-model="typeFilter" placeholder="类型筛选" class="type-filter">
          <el-option label="全部类型" value="" />
          <el-option label="Shell" value="shell" />
          <el-option label="Python2" value="python2" />
          <el-option label="Python3" value="python3" />
        </el-select>
      </div>
      <div class="search-right">
        <el-button @click="refreshScripts" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          新建脚本
        </el-button>
      </div>
    </div>

    <!-- 脚本列表 -->
    <div v-if="filteredScripts.length === 0" class="empty-state">
      <el-empty description="暂无脚本数据">
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          创建第一个脚本
        </el-button>
      </el-empty>
    </div>
    
    <el-table v-else :data="filteredScripts" style="width: 100%" stripe>
      <el-table-column prop="name" label="脚本名称" min-width="150">
        <template #default="{ row }">
          <div class="script-name">
            <strong>{{ row.name }}</strong>
            <div class="script-description">{{ row.description || '无描述' }}</div>
          </div>
        </template>
      </el-table-column>
      
      <el-table-column prop="type" label="类型" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="getTypeColor(row.type)" size="small">
            {{ row.type.toUpperCase() }}
          </el-tag>
        </template>
      </el-table-column>
      
      <el-table-column label="内容预览" min-width="200">
        <template #default="{ row }">
          <div class="code-preview">{{ getCodePreview(row.content) }}</div>
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
      
      <el-table-column label="操作" width="280" align="center" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="viewScript(row)">
            <el-icon><View /></el-icon>
            查看
          </el-button>
          <el-button size="small" type="primary" @click="createJob(row)">
            <el-icon><Clock /></el-icon>
            创建作业
          </el-button>
          <el-dropdown trigger="click" @command="(command) => handleCommand(command, row)">
            <el-button size="small">
              <el-icon><More /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="edit">
                  <el-icon><Edit /></el-icon>
                  编辑脚本
                </el-dropdown-item>
                <el-dropdown-item command="duplicate">
                  <el-icon><CopyDocument /></el-icon>
                  复制脚本
                </el-dropdown-item>
                <el-dropdown-item command="history">
                  <el-icon><List /></el-icon>
                  执行历史
                </el-dropdown-item>
                <el-dropdown-item divided command="delete" class="danger-item">
                  <el-icon><Delete /></el-icon>
                  删除脚本
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-table-column>
    </el-table>

    <!-- 脚本表单对话框 -->
    <ScriptForm
      v-model:visible="showFormDialog"
      :script="editingScript"
      @saved="handleScriptSaved"
    />

    <!-- 脚本详情对话框 -->
    <el-dialog
      v-model="showViewDialog"
      title="脚本详情"
      width="800px"
    >
      <div v-if="viewingScript" class="script-detail">
        <div class="detail-header">
          <h3>{{ viewingScript.name }}</h3>
          <el-tag :type="getTypeColor(viewingScript.type)">
            {{ viewingScript.type.toUpperCase() }}
          </el-tag>
        </div>
        <p class="description">{{ viewingScript.description || '无描述' }}</p>
        <div class="meta-info">
          <span>创建者: {{ viewingScript.user?.username }}</span>
          <span>创建时间: {{ formatDate(viewingScript.created_at) }}</span>
        </div>
        <div class="code-content">
          <pre>{{ viewingScript.content }}</pre>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import api from '@/utils/api'
import ScriptForm from './ScriptForm.vue'
import {
  Plus,
  Search,
  View,
  Edit,
  Delete,
  User,
  Clock,
  More,
  CopyDocument,
  List,
  Refresh
} from '@element-plus/icons-vue'

const router = useRouter()
const emit = defineEmits(['create-job'])

// 响应式数据
const scripts = ref([])
const searchText = ref('')
const typeFilter = ref('')
const loading = ref(false)
const showFormDialog = ref(false)
const showViewDialog = ref(false)
const editingScript = ref(null)
const viewingScript = ref(null)

// 计算属性
const filteredScripts = computed(() => {
  return scripts.value.filter(script => {
    const matchesSearch = !searchText.value || 
      script.name.toLowerCase().includes(searchText.value.toLowerCase()) ||
      (script.description && script.description.toLowerCase().includes(searchText.value.toLowerCase()))
    
    const matchesType = !typeFilter.value || script.type === typeFilter.value
    
    return matchesSearch && matchesType
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

const getCodePreview = (content) => {
  if (!content) return ''
  const lines = content.split('\n')
  return lines.slice(0, 3).join('\n') + (lines.length > 3 ? '\n...' : '')
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleString('zh-CN')
}

// API 方法
const loadScripts = async () => {
  loading.value = true
  try {
    const response = await api.get('/api/v1/scripts')
    scripts.value = response.data
  } catch (error) {
    ElMessage.error('加载脚本列表失败')
  } finally {
    loading.value = false
  }
}

const refreshScripts = () => {
  loadScripts()
}

// 脚本操作方法
const showCreateDialog = () => {
  editingScript.value = null
  showFormDialog.value = true
}

const viewScript = (script) => {
  viewingScript.value = script
  showViewDialog.value = true
}

const editScript = (script) => {
  editingScript.value = script
  showFormDialog.value = true
}

const deleteScript = async (script) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除脚本 "${script.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await api.delete(`/api/v1/scripts/${script.id}`)
    ElMessage.success('脚本删除成功')
    loadScripts()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除脚本失败')
    }
  }
}

const duplicateScript = async (script) => {
  try {
    const newScript = {
      name: `${script.name} - 副本`,
      type: script.type,
      description: script.description,
      content: script.content
    }
    
    await api.post('/api/v1/scripts', newScript)
    ElMessage.success('脚本复制成功')
    loadScripts()
  } catch (error) {
    ElMessage.error('复制脚本失败')
  }
}

const createJob = (script) => {
  emit('create-job', script)
}

const viewScriptExecutions = (script) => {
  router.push({
    path: '/executions',
    query: { script_id: script.id, script_name: script.name }
  })
}

const handleScriptSaved = () => {
  loadScripts()
}

// 处理下拉菜单命令
const handleCommand = (command, script) => {
  switch (command) {
    case 'edit':
      editScript(script)
      break
    case 'duplicate':
      duplicateScript(script)
      break
    case 'history':
      viewScriptExecutions(script)
      break
    case 'delete':
      deleteScript(script)
      break
  }
}

// 生命周期
onMounted(() => {
  loadScripts()
})
</script>

<style scoped>
.script-manager {
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

.scripts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
}

.script-card {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 20px;
  transition: all 0.3s ease;
  cursor: pointer;
}

.script-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.script-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.script-info h3 {
  margin: 0 0 4px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.script-info p {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.script-content {
  margin: 12px 0;
}

.code-preview {
  background: #f5f7fa;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 8px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  color: #606266;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 60px;
  overflow: hidden;
  position: relative;
}

.code-preview::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 15px;
  background: linear-gradient(transparent, #f5f7fa);
}

.script-name strong {
  color: #303133;
  font-size: 14px;
  font-weight: 600;
}

.script-description {
  color: #909399;
  font-size: 12px;
  margin-top: 4px;
  line-height: 1.4;
}

.script-meta {
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

.script-actions {
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

.script-detail .detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.script-detail .detail-header h3 {
  margin: 0;
  font-size: 18px;
}

.script-detail .description {
  color: #606266;
  margin-bottom: 16px;
}

.script-detail .meta-info {
  display: flex;
  gap: 24px;
  margin-bottom: 20px;
  font-size: 14px;
  color: #909399;
}

.script-detail .code-content {
  background: #f5f7fa;
  padding: 16px;
  border-radius: 4px;
  max-height: 400px;
  overflow: auto;
}

.script-detail .code-content pre {
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
  white-space: pre-wrap;
}

.danger-item {
  color: #f56c6c !important;
}
</style>
