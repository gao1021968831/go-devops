<template>
  <el-dialog
    v-model="dialogVisible"
    title="快速脚本执行"
    width="1000px"
    :close-on-click-modal="false"
    class="quick-execute-dialog"
    @closed="handleClosed"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item label="执行名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入执行任务名称" />
      </el-form-item>
      
      <el-form-item label="执行方式" prop="executeMode">
        <el-radio-group v-model="form.executeMode" @change="handleExecuteModeChange">
          <el-radio value="new">新建脚本</el-radio>
          <el-radio value="existing">选择已有脚本</el-radio>
        </el-radio-group>
      </el-form-item>
      
      <el-form-item v-if="form.executeMode === 'existing'" label="选择脚本" prop="selectedScriptId">
        <el-select 
          v-model="form.selectedScriptId" 
          placeholder="请选择脚本" 
          style="width: 100%"
          @change="handleScriptSelect"
          filterable
        >
          <el-option
            v-for="script in availableScripts"
            :key="script.id"
            :label="script.name"
            :value="script.id"
          >
            <div class="script-option">
              <div class="script-name">{{ script.name }}</div>
              <div class="script-desc">{{ script.description || '无描述' }} - {{ script.type }}</div>
            </div>
          </el-option>
        </el-select>
      </el-form-item>
      
      <el-form-item v-if="form.executeMode === 'new'" label="脚本类型" prop="scriptType">
        <el-select v-model="form.scriptType" placeholder="请选择脚本类型" style="width: 200px">
          <el-option
            v-for="type in scriptTypes"
            :key="type.value"
            :label="type.label"
            :value="type.value"
          >
            <div class="script-type-option">
              <el-icon><component :is="type.icon" /></el-icon>
              <span>{{ type.label }}</span>
            </div>
          </el-option>
        </el-select>
        <span class="type-desc">{{ getTypeDescription(form.scriptType) }}</span>
      </el-form-item>
      
      <el-form-item label="脚本内容" prop="scriptContent">
        <div class="script-content-section">
          <div v-if="form.executeMode === 'existing' && form.selectedScriptId" class="script-actions">
            <el-button size="small" type="primary" @click="enableScriptEdit" v-if="!scriptEditable">
              <el-icon><Edit /></el-icon>
              编辑脚本
            </el-button>
            <el-button size="small" @click="cancelScriptEdit" v-if="scriptEditable">
              <el-icon><Close /></el-icon>
              取消编辑
            </el-button>
            <el-button size="small" type="success" @click="saveScriptEdit" v-if="scriptEditable">
              <el-icon><Check /></el-icon>
              保存修改
            </el-button>
          </div>
          <CodeEditor
            v-model="form.scriptContent"
            :language="getEditorLanguage(getCurrentScriptType())"
            :height="350"
            :readonly="form.executeMode === 'existing' && !scriptEditable"
            @language-change="handleLanguageChange"
          />
        </div>
      </el-form-item>
      
      <el-form-item label="目标主机" prop="hostIds">
        <HostSelector v-model="form.hostIds" />
      </el-form-item>
      
      <el-form-item label="输入文件">
        <div class="input-files-section">
          <div class="selected-files" v-if="form.inputFiles.length > 0">
            <div 
              v-for="file in form.inputFiles" 
              :key="file.id" 
              class="file-tag"
            >
              <el-icon><Document /></el-icon>
              <span class="file-name">{{ file.name }}</span>
              <span class="file-size">({{ formatFileSize(file.size) }})</span>
              <el-icon 
                class="remove-file" 
                @click="removeInputFile(file.id)"
              >
                <Close />
              </el-icon>
            </div>
          </div>
          <el-button 
            type="primary" 
            plain 
            size="small" 
            @click="handleSelectFiles"
          >
            <el-icon><Files /></el-icon>
            选择输入文件
          </el-button>
          <span class="file-help-text">可选择多个文件作为脚本执行的输入参数</span>
        </div>
      </el-form-item>
      
      <el-form-item label="备注" prop="description">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="2"
          placeholder="请输入执行备注信息（可选）"
        />
      </el-form-item>
    </el-form>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button size="large" @click="handleCancel">取消</el-button>
        <el-button 
          size="large" 
          type="primary" 
          @click="handleExecute" 
          :loading="executing"
          :disabled="form.hostIds.length === 0 || !form.scriptContent.trim()"
        >
          <el-icon><VideoPlay /></el-icon>
          {{ executing ? '执行中...' : '立即执行' }}
        </el-button>
      </div>
    </template>
    
    <!-- 文件选择器 -->
    <FileSelector
      ref="fileSelectorRef"
      v-model="form.inputFiles"
      :multiple="true"
      title="选择输入文件"
    />
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import api from '@/utils/api'
import HostSelector from './HostSelector.vue'
import CodeEditor from './CodeEditor.vue'
import FileSelector from './FileSelector.vue'
import {
  Tools,
  Document,
  Delete,
  VideoPlay,
  Monitor,
  Edit,
  DataLine,
  Setting,
  Files,
  Connection,
  Close,
  Check
} from '@element-plus/icons-vue'

const props = defineProps({
  visible: Boolean,
  prefillData: Object
})

const emit = defineEmits(['update:visible', 'executed'])

const router = useRouter()

// 响应式数据
const formRef = ref()
const executing = ref(false)
const availableScripts = ref([])
const scriptEditable = ref(false)
const originalScriptContent = ref('')
const fileSelectorRef = ref()
const availableFiles = ref([])

const form = ref({
  name: '',
  executeMode: 'new',
  selectedScriptId: '',
  scriptType: 'shell',
  scriptContent: '',
  hostIds: [],
  inputFiles: [],
  description: ''
})

const scriptTypes = [
  { value: 'shell', label: 'Shell', icon: 'Monitor', desc: 'Linux/Unix Shell脚本' },
  { value: 'python2', label: 'Python2', icon: 'Edit', desc: 'Python 2.x 脚本' },
  { value: 'python3', label: 'Python3', icon: 'Edit', desc: 'Python 3.x 脚本' }
]

const rules = {
  name: [
    { required: true, message: '请输入执行名称', trigger: 'blur' },
    { min: 2, max: 100, message: '执行名称长度在 2 到 100 个字符', trigger: 'blur' }
  ],
  scriptType: [
    { required: true, message: '请选择脚本类型', trigger: 'change' }
  ],
  scriptContent: [
    { required: true, message: '请输入脚本内容', trigger: 'blur' },
    { min: 1, message: '脚本内容不能为空', trigger: 'blur' }
  ],
  hostIds: [
    { 
      type: 'array', 
      required: true, 
      message: '请选择目标主机', 
      trigger: 'change',
      validator: (rule, value, callback) => {
        if (!value || value.length === 0) {
          callback(new Error('请至少选择一台主机'))
        } else {
          callback()
        }
      }
    }
  ]
}

// 计算属性
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
})

const lineCount = ref(1)

// 方法
const updateLineCount = () => {
  lineCount.value = form.value.scriptContent.split('\n').length
}

const getTypeDescription = (type) => {
  const scriptType = scriptTypes.find(t => t.value === type)
  return scriptType ? scriptType.desc : ''
}

const getEditorLanguage = (scriptType) => {
  const languageMap = {
    'shell': 'shell',
    'python2': 'python2',
    'python3': 'python3'
  }
  return languageMap[scriptType] || 'shell'
}

const handleLanguageChange = (newLanguage) => {
  // 根据编辑器语言更新脚本类型
  const typeMap = {
    'shell': 'shell',
    'python2': 'python2',
    'python3': 'python3'
  }
  const newType = typeMap[newLanguage]
  if (newType && scriptTypes.find(t => t.value === newType)) {
    form.value.scriptType = newType
  }
}

// 加载可用脚本列表
const loadAvailableScripts = async () => {
  try {
    const response = await api.get('/api/v1/scripts')
    availableScripts.value = response.data || []
  } catch (error) {
    console.error('加载脚本列表失败:', error)
  }
}

// 处理执行方式变化
const handleExecuteModeChange = (mode) => {
  if (mode === 'existing') {
    loadAvailableScripts()
    form.value.scriptContent = ''
    form.value.selectedScriptId = ''
  } else {
    form.value.selectedScriptId = ''
    form.value.scriptContent = ''
  }
  scriptEditable.value = false
}

// 处理脚本选择
const handleScriptSelect = async (scriptId) => {
  if (!scriptId) return
  
  try {
    const response = await api.get(`/api/v1/scripts/${scriptId}`)
    const script = response.data
    form.value.scriptContent = script.content
    form.value.scriptType = script.type
    originalScriptContent.value = script.content
    scriptEditable.value = false
    
    // 自动设置执行名称
    if (!form.value.name) {
      form.value.name = `执行-${script.name}`
    }
  } catch (error) {
    ElMessage.error('加载脚本内容失败')
  }
}

// 获取当前脚本类型
const getCurrentScriptType = () => {
  if (form.value.executeMode === 'existing' && form.value.selectedScriptId) {
    const script = availableScripts.value.find(s => s.id === form.value.selectedScriptId)
    return script?.type || 'shell'
  }
  return form.value.scriptType
}

// 启用脚本编辑
const enableScriptEdit = () => {
  scriptEditable.value = true
  originalScriptContent.value = form.value.scriptContent
}

// 取消脚本编辑
const cancelScriptEdit = () => {
  scriptEditable.value = false
  form.value.scriptContent = originalScriptContent.value
}

// 保存脚本编辑
const saveScriptEdit = async () => {
  try {
    await api.put(`/api/v1/scripts/${form.value.selectedScriptId}`, {
      content: form.value.scriptContent
    })
    originalScriptContent.value = form.value.scriptContent
    scriptEditable.value = false
    ElMessage.success('脚本内容已保存')
  } catch (error) {
    ElMessage.error('保存脚本失败')
  }
}

// 选择输入文件
const handleSelectFiles = () => {
  if (fileSelectorRef.value) {
    fileSelectorRef.value.openFileDialog()
  }
}

// 移除输入文件
const removeInputFile = (fileId) => {
  form.value.inputFiles = form.value.inputFiles.filter(file => file.id !== fileId)
}

// 格式化文件大小
const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}


const handleExecute = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    
    executing.value = true
    
    const response = await api.post('/api/v1/scripts/quick-execute', {
      name: form.value.name,
      script_content: form.value.scriptContent,
      script_type: form.value.scriptType,
      host_ids: form.value.hostIds,
      input_file_ids: form.value.inputFiles.map(f => f.id),
      description: form.value.description
    })
    
    ElMessage.success('快速脚本执行已启动')
    emit('executed', response.data.executions)
    dialogVisible.value = false
    
    // 跳转到执行记录详情页面
    if (response.data.executions && response.data.executions.length > 0) {
      const executionId = response.data.executions[0].id
      router.push(`/executions/${executionId}`)
    }
    
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '执行失败')
  } finally {
    executing.value = false
  }
}

const handleCancel = () => {
  dialogVisible.value = false
}

const handleClosed = () => {
  // 重置表单
  if (formRef.value) {
    formRef.value.resetFields()
  }
  form.value = {
    name: '',
    executeMode: 'new',
    selectedScriptId: '',
    scriptType: 'shell',
    scriptContent: '',
    hostIds: [],
    inputFiles: [],
    description: ''
  }
  scriptEditable.value = false
  originalScriptContent.value = ''
  executing.value = false
}

// 监听器
watch(() => form.value.scriptContent, updateLineCount)

watch(() => form.value.scriptType, (newType) => {
  if (!form.value.name) {
    form.value.name = `快速执行-${scriptTypes.find(t => t.value === newType)?.label || newType}`
  }
})

// 监听预填充数据
watch(() => props.prefillData, (newData) => {
  if (newData && props.visible) {
    form.value = {
      name: newData.name || '',
      executeMode: newData.executeMode || 'new',
      selectedScriptId: newData.selectedScriptId || '',
      scriptType: newData.scriptType || 'shell',
      scriptContent: newData.scriptContent || '',
      hostIds: newData.hostIds || [],
      inputFiles: newData.inputFiles || [],
      description: newData.description || ''
    }
  }
}, { immediate: true })
</script>

<style scoped>
.quick-execute-dialog :deep(.el-dialog__body) {
  padding: 20px 30px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.script-type-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.type-desc {
  margin-left: 12px;
  color: #909399;
  font-size: 12px;
}

.script-option {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.script-name {
  font-weight: 600;
  color: #303133;
}

.script-desc {
  font-size: 12px;
  color: #909399;
}

.script-content-section {
  position: relative;
}

.script-actions {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  padding: 8px;
  background: #f5f7fa;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
}

.code-editor-container {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  overflow: hidden;
}

.editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
}

.editor-info {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #909399;
}

.script-textarea {
  border: none;
}

.script-textarea :deep(.el-textarea__inner) {
  border: none;
  border-radius: 0;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  resize: none;
}

.script-textarea :deep(.el-textarea__inner):focus {
  box-shadow: none;
}

.input-files-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.selected-files {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.file-tag {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: #f0f9ff;
  border: 1px solid #bae6fd;
  border-radius: 16px;
  font-size: 13px;
  color: #0369a1;
}

.file-name {
  font-weight: 500;
}

.file-size {
  color: #64748b;
  font-size: 12px;
}

.remove-file {
  cursor: pointer;
  color: #ef4444;
  font-size: 14px;
  margin-left: 4px;
}

.remove-file:hover {
  color: #dc2626;
}

.file-help-text {
  font-size: 12px;
  color: #64748b;
  margin-left: 8px;
}
</style>
