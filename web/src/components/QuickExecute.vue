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
      
      <el-form-item label="脚本类型" prop="scriptType">
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
        <div class="code-editor-container">
          <div class="editor-toolbar">
            <el-button-group size="small">
              <el-button @click="formatCode">
                <el-icon><Tools /></el-icon>
                格式化
              </el-button>
              <el-button @click="insertTemplate">
                <el-icon><Document /></el-icon>
                插入模板
              </el-button>
              <el-button @click="clearContent">
                <el-icon><Delete /></el-icon>
                清空
              </el-button>
            </el-button-group>
            <div class="editor-info">
              <span>行数: {{ lineCount }}</span>
              <span>字符: {{ form.scriptContent.length }}</span>
            </div>
          </div>
          <el-input
            v-model="form.scriptContent"
            type="textarea"
            :rows="12"
            placeholder="请输入脚本内容..."
            class="script-textarea"
            @input="updateLineCount"
          />
        </div>
      </el-form-item>
      
      <el-form-item label="目标主机" prop="hostIds">
        <HostSelector v-model="form.hostIds" />
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
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import api from '@/utils/api'
import HostSelector from './HostSelector.vue'
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
  Connection
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

const form = ref({
  name: '',
  scriptType: 'shell',
  scriptContent: '',
  hostIds: [],
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

const formatCode = () => {
  if (!form.value.scriptContent.trim()) {
    ElMessage.warning('请先输入脚本内容')
    return
  }
  
  // 简单的代码格式化
  const lines = form.value.scriptContent.split('\n')
  const formattedLines = lines.map(line => line.trim()).filter(line => line)
  form.value.scriptContent = formattedLines.join('\n')
  updateLineCount()
  ElMessage.success('代码格式化完成')
}

const insertTemplate = () => {
  const templates = {
    shell: `#!/bin/bash
echo "Hello BlueKing"
echo "当前时间: $(date)"
echo "当前用户: $(whoami)"
echo "系统信息: $(uname -a)"`,
    python2: `#!/usr/bin/env python2
# -*- coding: utf-8 -*-
print "Hello BlueKing"
import datetime
print "当前时间: " + str(datetime.datetime.now())
import os
print "当前用户: " + os.getenv('USER', 'unknown')`,
    python3: `#!/usr/bin/env python3
# -*- coding: utf-8 -*-
print("Hello BlueKing")
import datetime
print(f"当前时间: {datetime.datetime.now()}")
import os
print(f"当前用户: {os.getenv('USER', 'unknown')}")`
  }
  
  const template = templates[form.value.scriptType] || templates.shell
  form.value.scriptContent = template
  updateLineCount()
  ElMessage.success('模板插入成功')
}

const clearContent = () => {
  ElMessageBox.confirm('确定要清空脚本内容吗？', '确认清空', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    form.value.scriptContent = ''
    updateLineCount()
    ElMessage.success('内容已清空')
  }).catch(() => {})
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
    scriptType: 'shell',
    scriptContent: '',
    hostIds: [],
    description: ''
  }
  lineCount.value = 1
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
      scriptType: newData.scriptType || 'shell',
      scriptContent: newData.scriptContent || '',
      hostIds: newData.hostIds || [],
      description: newData.description || ''
    }
    updateLineCount()
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
</style>
