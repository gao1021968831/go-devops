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
        <el-dropdown trigger="click" :disabled="selectedScripts.length === 0">
          <el-button :disabled="selectedScripts.length === 0">
            <el-icon><Operation /></el-icon>
            批量操作 <el-icon class="el-icon--right"><arrow-down /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="exportSelected">
                <el-icon><Download /></el-icon>
                导出选中脚本
              </el-dropdown-item>
              <el-dropdown-item @click="deleteSelected" class="danger-item">
                <el-icon><Delete /></el-icon>
                删除选中脚本
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button type="success" @click="showImportDialog">
          <el-icon><Upload /></el-icon>
          批量导入
        </el-button>
        <el-button type="primary" @click="exportAll">导出全部脚本</el-button>
        <el-button type="success" @click="downloadTemplate">下载导入模板</el-button>
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
    
    <el-table v-else :data="filteredScripts" style="width: 100%" stripe @selection-change="handleSelectionChange">
      <el-table-column type="selection" width="55" />
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

    <!-- 批量导入对话框 -->
    <el-dialog
      v-model="showImportDialogFlag"
      title="批量导入脚本"
      width="600px"
    >
      <div class="import-section">
        <el-alert
          title="导入说明"
          type="info"
          :closable="false"
          show-icon
        >
          <p>1. 请使用CSV格式文件，包含以下列：脚本名称、描述、类型、脚本内容</p>
          <p>2. 脚本类型支持：shell、python2、python3</p>
          <p>3. 脚本名称不能重复，重复的脚本将跳过导入</p>
          <p>4. 建议先下载模板文件，按照格式填写数据</p>
        </el-alert>
        
        <div class="upload-section">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            accept=".csv"
            :on-change="handleFileChange"
            :file-list="fileList"
            drag
          >
            <el-icon class="el-icon--upload"><upload-filled /></el-icon>
            <div class="el-upload__text">
              将CSV文件拖到此处，或<em>点击上传</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">
                只能上传CSV文件，且不超过10MB
              </div>
            </template>
          </el-upload>
        </div>

        <div v-if="importResult" class="import-result">
          <el-alert
            :title="importResult.message"
            :type="importResult.success_count > 0 ? 'success' : 'error'"
            show-icon
          >
            <p>总计: {{ importResult.total_count }} 个</p>
            <p>成功: {{ importResult.success_count }} 个</p>
            <p>失败: {{ importResult.error_count }} 个</p>
            <div v-if="importResult.errors && importResult.errors.length > 0">
              <p>错误详情:</p>
              <ul>
                <li v-for="error in importResult.errors" :key="error">{{ error }}</li>
              </ul>
            </div>
          </el-alert>
        </div>
      </div>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showImportDialogFlag = false">取消</el-button>
          <el-button type="primary" @click="downloadTemplate">下载模板</el-button>
          <el-button type="success" @click="performImport" :loading="importing" :disabled="!selectedFile">
            开始导入
          </el-button>
        </span>
      </template>
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
  Refresh,
  Operation,
  Download,
  Upload,
  UploadFilled,
  ArrowDown
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
const selectedScripts = ref([])
const showImportDialogFlag = ref(false)
const importing = ref(false)
const selectedFile = ref(null)
const fileList = ref([])
const importResult = ref(null)

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
    scripts.value = response.data.scripts || []
  } catch (error) {
    ElMessage.error('加载脚本列表失败')
  } finally {
    loading.value = false
  }
}

const refreshScripts = () => {
  loadScripts()
}

// 批量操作相关方法
const handleSelectionChange = (selection) => {
  selectedScripts.value = selection
}

const handleBatchCommand = (command) => {
  if (command === 'delete') {
    batchDeleteScripts()
  } else if (command === 'export') {
    exportSelectedScripts()
  }
}

const batchDeleteScripts = async () => {
  if (selectedScripts.value.length === 0) {
    ElMessage.warning('请先选择要删除的脚本')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedScripts.value.length} 个脚本吗？`,
      '批量删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const ids = selectedScripts.value.map(script => script.id)
    const response = await api.post('/api/v1/scripts/batch/delete', { ids })
    
    ElMessage.success(response.data.message)
    selectedScripts.value = []
    loadScripts()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

const exportSelectedScripts = async () => {
  if (selectedScripts.value.length === 0) {
    ElMessage.warning('请先选择要导出的脚本')
    return
  }

  try {
    const ids = selectedScripts.value.map(script => script.id)
    const response = await api.get('/api/v1/scripts/export', {
      params: { ids: JSON.stringify(ids) },
      responseType: 'blob'
    })
    
    // 创建下载链接
    const blob = new Blob([response.data], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `scripts_export_${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    
    ElMessage.success(`成功导出 ${selectedScripts.value.length} 个脚本`)
  } catch (error) {
    ElMessage.error('导出脚本失败')
  }
}

const exportAllScripts = async () => {
  try {
    const response = await api.get('/api/v1/scripts/export', {
      responseType: 'blob'
    })
    
    // 创建下载链接
    const blob = new Blob([response.data], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `scripts_export_all_${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    
    ElMessage.success('成功导出所有脚本')
  } catch (error) {
    ElMessage.error('导出脚本失败')
  }
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

// 导入相关方法
const showImportDialog = () => {
  showImportDialogFlag.value = true
  importResult.value = null
  selectedFile.value = null
  fileList.value = []
}

const handleFileChange = (file) => {
  selectedFile.value = file.raw
  fileList.value = [file]
}

const downloadTemplate = async () => {
  try {
    const response = await api.get('/api/v1/scripts/import/template', {
      responseType: 'blob'
    })
    
    // 创建下载链接
    const blob = new Blob([response.data], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = 'scripts_import_template.csv'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    
    ElMessage.success('模板下载成功')
  } catch (error) {
    ElMessage.error('下载模板失败')
  }
}

const performImport = async () => {
  if (!selectedFile.value) {
    ElMessage.warning('请先选择要导入的CSV文件')
    return
  }

  importing.value = true
  importResult.value = null

  try {
    const formData = new FormData()
    formData.append('file', selectedFile.value)

    const response = await api.post('/api/v1/scripts/batch/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })

    importResult.value = response.data
    ElMessage.success(response.data.message)
    
    // 如果有成功导入的脚本，刷新列表
    if (response.data.success_count > 0) {
      loadScripts()
    }
  } catch (error) {
    ElMessage.error('导入失败：' + (error.response?.data?.error || '未知错误'))
  } finally {
    importing.value = false
  }
}

// 修复方法名映射
const exportSelected = () => exportSelectedScripts()
const deleteSelected = () => batchDeleteScripts()
const exportAll = () => exportAllScripts()

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
