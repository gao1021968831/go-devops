<template>
  <el-dialog
    v-model="dialogVisible"
    title="创建作业"
    width="900px"
    :close-on-click-modal="false"
    class="job-dialog"
    @closed="handleClosed"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item label="作业名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入作业名称" />
      </el-form-item>
      
      <el-form-item label="脚本信息">
        <div class="script-info-display">
          <div class="script-icon">
            <el-icon><Document /></el-icon>
          </div>
          <div class="script-details">
            <div class="script-title">
              <span class="script-name">{{ script?.name }}</span>
              <el-tag :type="getTypeColor(script?.type)" size="small">
                {{ (script?.type || '').toUpperCase() }}
              </el-tag>
              <el-button 
                size="small" 
                type="primary" 
                link 
                @click="showScriptContent = !showScriptContent"
              >
                <el-icon><View /></el-icon>
                {{ showScriptContent ? '隐藏内容' : '查看内容' }}
              </el-button>
            </div>
            <div class="script-desc" v-if="script?.description">
              {{ script.description }}
            </div>
            <div v-if="showScriptContent && script?.content" class="script-content-preview">
              <div class="content-header">
                <span>脚本内容预览</span>
                <el-tag size="small" type="info">{{ script.content.split('\n').length }} 行</el-tag>
              </div>
              <pre class="script-code">{{ script.content }}</pre>
            </div>
          </div>
        </div>
      </el-form-item>
      
      <el-form-item label="目标主机" prop="hostIds">
        <HostSelector v-model="form.hostIds" />
      </el-form-item>
      
      <el-form-item label="执行方式" prop="executeType">
        <el-select 
          v-model="form.executeType" 
          placeholder="请选择执行方式"
          style="width: 100%"
          size="large"
        >
          <el-option
            label="手动执行"
            value="manual"
          >
            <div class="select-option-content">
              <div class="option-icon">
                <el-icon><VideoPlay /></el-icon>
              </div>
              <div class="option-text">
                <div class="option-title">手动执行</div>
                <div class="option-desc">立即创建作业，需要手动触发执行</div>
              </div>
            </div>
          </el-option>
          <el-option
            label="定时执行"
            value="scheduled"
          >
            <div class="select-option-content">
              <div class="option-icon">
                <el-icon><Clock /></el-icon>
              </div>
              <div class="option-text">
                <div class="option-title">定时执行</div>
                <div class="option-desc">设置执行时间，系统自动执行</div>
              </div>
            </div>
          </el-option>
        </el-select>
      </el-form-item>
      
      <el-form-item v-if="form.executeType === 'scheduled'" label="执行时间" prop="scheduledTime">
        <el-date-picker
          v-model="form.scheduledTime"
          type="datetime"
          placeholder="选择执行时间"
          style="width: 100%"
          :disabled-date="disabledDate"
          format="YYYY-MM-DD HH:mm:ss"
          value-format="YYYY-MM-DD HH:mm:ss"
        />
        <div class="time-help-text">
          <el-icon><InfoFilled /></el-icon>
          <span>请选择未来的时间点进行定时执行</span>
        </div>
      </el-form-item>
      
      <el-form-item label="备注" prop="description">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="3"
          placeholder="请输入作业备注信息（可选）"
        />
      </el-form-item>
    </el-form>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button size="large" @click="handleCancel">取消</el-button>
        <el-button 
          type="primary" 
          size="large" 
          @click="handleSave" 
          :loading="saving"
          :disabled="form.hostIds.length === 0"
        >
          <el-icon><Plus /></el-icon>
          {{ isEdit ? '更新作业' : '创建作业' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'
import HostSelector from './HostSelector.vue'
import {
  Document,
  VideoPlay,
  Clock,
  InfoFilled,
  Plus,
  View
} from '@element-plus/icons-vue'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  job: {
    type: Object,
    default: null
  },
  script: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:visible', 'saved'])

// 响应式数据
const formRef = ref()
const saving = ref(false)
const showScriptContent = ref(false)

const form = ref({
  name: '',
  hostIds: [],
  executeType: 'manual',
  scheduledTime: null,
  description: ''
})

const rules = {
  name: [
    { required: true, message: '请输入作业名称', trigger: 'blur' },
    { min: 2, max: 100, message: '作业名称长度在 2 到 100 个字符', trigger: 'blur' }
  ],
  hostIds: [
    { 
      required: true, 
      validator: (rule, value, callback) => {
        if (!value || value.length === 0) {
          callback(new Error('请选择目标主机'))
        } else {
          callback()
        }
      }, 
      trigger: 'change' 
    }
  ],
  scheduledTime: [
    { 
      required: true, 
      validator: (rule, value, callback) => {
        if (form.value.executeType === 'scheduled' && !value) {
          callback(new Error('请选择执行时间'))
        } else {
          callback()
        }
      }, 
      trigger: 'change' 
    }
  ]
}

// 计算属性
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
})

const isEdit = computed(() => !!props.job)

// 工具函数
const getTypeColor = (type) => {
  const colorMap = {
    shell: 'success',
    python2: 'warning',
    python3: 'info'
  }
  return colorMap[type] || 'primary'
}

const disabledDate = (time) => {
  return time.getTime() < Date.now() - 24 * 60 * 60 * 1000
}

// 方法
const resetForm = () => {
  form.value = {
    name: '',
    hostIds: [],
    executeType: 'manual',
    scheduledTime: null,
    description: ''
  }
}

const loadJob = () => {
  if (props.job) {
    // 编辑模式
    form.value = {
      name: props.job.name,
      hostIds: props.job.host_ids ? JSON.parse(props.job.host_ids) : [],
      executeType: props.job.execute_type,
      scheduledTime: props.job.scheduled_time,
      description: props.job.description || ''
    }
  } else if (props.script) {
    // 创建模式
    form.value = {
      name: `执行脚本: ${props.script.name}`,
      hostIds: [],
      executeType: 'manual',
      scheduledTime: null,
      description: ''
    }
  } else {
    resetForm()
  }
}

const handleSave = async () => {
  if (!formRef.value) return
  
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  // 检查定时执行时间
  if (form.value.executeType === 'scheduled' && !form.value.scheduledTime) {
    ElMessage.warning('请选择执行时间')
    return
  }

  saving.value = true
  try {
    const jobData = {
      name: form.value.name,
      script_id: props.script?.id || props.job?.script_id,
      host_ids: JSON.stringify(form.value.hostIds),
      execute_type: form.value.executeType,
      scheduled_time: form.value.scheduledTime,
      description: form.value.description
    }
    
    if (isEdit.value) {
      await api.put(`/api/v1/jobs/${props.job.id}`, jobData)
      ElMessage.success('作业更新成功')
    } else {
      await api.post('/api/v1/jobs', jobData)
      ElMessage.success('作业创建成功')
    }
    
    emit('saved')
    dialogVisible.value = false
  } catch (error) {
    ElMessage.error(isEdit.value ? '更新作业失败' : '创建作业失败')
  } finally {
    saving.value = false
  }
}

const handleCancel = () => {
  dialogVisible.value = false
}

const handleClosed = () => {
  // 重置表单验证状态
  if (formRef.value) {
    formRef.value.resetFields()
  }
}

// 监听器
watch(() => props.visible, (visible) => {
  if (visible) {
    nextTick(() => {
      loadJob()
    })
  }
})

watch(() => form.value.executeType, (newType) => {
  if (newType === 'manual') {
    form.value.scheduledTime = null
  }
})
</script>

<style scoped>
.job-dialog :deep(.el-dialog__body) {
  padding: 20px 30px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.script-info-display {
  display: flex;
  align-items: center;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}

.script-icon {
  margin-right: 12px;
  color: #409eff;
  font-size: 24px;
}

.script-details {
  flex: 1;
}

.script-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.script-name {
  font-weight: 600;
  color: #303133;
  font-size: 16px;
}

.script-desc {
  color: #909399;
  font-size: 14px;
}

.execute-type-group {
  width: 100%;
}

.execute-radio {
  width: 100%;
  margin-bottom: 16px;
  padding: 16px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  transition: all 0.3s;
}

.execute-radio:hover {
  border-color: #409eff;
  background: #f0f9ff;
}

.execute-radio.is-checked {
  border-color: #409eff;
  background: #f0f9ff;
}

.execute-radio :deep(.el-radio__input) {
  margin-right: 12px;
}

.execute-radio :deep(.el-radio__label) {
  width: 100%;
  padding: 0;
}

.radio-content {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  width: 100%;
}

.radio-icon {
  color: #409eff;
  font-size: 20px;
  margin-top: 2px;
}

.radio-text {
  flex: 1;
}

.radio-title {
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

.radio-desc {
  color: #909399;
  font-size: 14px;
  line-height: 1.4;
}

.time-help-text {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}

.script-content-preview {
  margin-top: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  overflow: hidden;
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #fafafa;
  border-bottom: 1px solid #e4e7ed;
  font-size: 13px;
  font-weight: 500;
  color: #606266;
}

.script-code {
  margin: 0;
  padding: 12px;
  background: #f8f9fa;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  color: #2c3e50;
  max-height: 200px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.select-option-content {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
  width: 100%;
}

:deep(.el-select-dropdown__item) {
  height: auto !important;
  padding: 12px 20px !important;
  line-height: normal !important;
}

.option-icon {
  color: #409eff;
  font-size: 18px;
  flex-shrink: 0;
}

.option-text {
  flex: 1;
}

.option-title {
  font-weight: 600;
  color: #303133;
  margin-bottom: 2px;
}

.option-desc {
  color: #909399;
  font-size: 12px;
  line-height: 1.4;
}
</style>
