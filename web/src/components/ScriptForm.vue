<template>
  <el-dialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑脚本' : '新建脚本'"
    width="800px"
    :close-on-click-modal="false"
    @closed="handleClosed"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="80px"
    >
      <el-form-item label="脚本名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入脚本名称" />
      </el-form-item>
      
      <el-form-item label="脚本类型" prop="type">
        <el-select v-model="form.type" placeholder="选择脚本类型" style="width: 100%">
          <el-option label="Shell" value="shell">
            <div class="option-item">
              <span class="option-label">Shell</span>
              <span class="option-desc">适用于 Linux/Unix 系统</span>
            </div>
          </el-option>
          <el-option label="Python2" value="python2">
            <div class="option-item">
              <span class="option-label">Python2</span>
              <span class="option-desc">Python 2.x 脚本</span>
            </div>
          </el-option>
          <el-option label="Python3" value="python3">
            <div class="option-item">
              <span class="option-label">Python3</span>
              <span class="option-desc">Python 3.x 脚本</span>
            </div>
          </el-option>
        </el-select>
      </el-form-item>
      
      <el-form-item label="描述">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="2"
          placeholder="请输入脚本描述（可选）"
        />
      </el-form-item>
      
      <el-form-item label="脚本内容" prop="content">
        <div class="code-editor-container">
          <div class="editor-toolbar">
            <el-button-group size="small">
              <el-button @click="formatCode">
                <el-icon><Tools /></el-icon>
                格式化
              </el-button>
              <el-button @click="insertTemplate">
                <el-icon><Document /></el-icon>
                模板
              </el-button>
            </el-button-group>
            <div class="editor-info">
              <span>行数: {{ lineCount }}</span>
              <span>字符: {{ form.content.length }}</span>
            </div>
          </div>
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="15"
            placeholder="请输入脚本内容"
            class="code-editor"
            @input="updateLineCount"
          />
        </div>
      </el-form-item>
    </el-form>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleCancel">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">
          <el-icon><Check /></el-icon>
          {{ isEdit ? '更新' : '保存' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'
import {
  Check,
  Tools,
  Document
} from '@element-plus/icons-vue'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
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
const lineCount = ref(1)

const form = ref({
  name: '',
  type: 'shell',
  description: '',
  content: ''
})

const rules = {
  name: [
    { required: true, message: '请输入脚本名称', trigger: 'blur' },
    { min: 2, max: 50, message: '脚本名称长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择脚本类型', trigger: 'change' }
  ],
  content: [
    { required: true, message: '请输入脚本内容', trigger: 'blur' },
    { min: 10, message: '脚本内容至少需要 10 个字符', trigger: 'blur' }
  ]
}

// 计算属性
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
})

const isEdit = computed(() => !!props.script)

// 脚本模板
const scriptTemplates = {
  shell: `#!/bin/bash
# Shell 脚本模板
# 描述: 

set -e  # 遇到错误时退出

echo "开始执行脚本..."

# 在这里添加你的脚本逻辑

echo "脚本执行完成"`,
  
  python2: `#!/usr/bin/env python2
# -*- coding: utf-8 -*-
"""
Python2 脚本模板
描述: 
"""

import sys
import os

def main():
    """主函数"""
    print "开始执行脚本..."
    
    # 在这里添加你的脚本逻辑
    
    print "脚本执行完成"

if __name__ == "__main__":
    main()`,
  
  python3: `#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Python3 脚本模板
描述: 
"""

import sys
import os

def main():
    """主函数"""
    print("开始执行脚本...")
    
    # 在这里添加你的脚本逻辑
    
    print("脚本执行完成")

if __name__ == "__main__":
    main()`
}

// 方法
const updateLineCount = () => {
  lineCount.value = form.value.content.split('\n').length
}

const formatCode = () => {
  // 简单的代码格式化
  let content = form.value.content
  
  // 移除多余的空行
  content = content.replace(/\n\s*\n\s*\n/g, '\n\n')
  
  // 统一缩进（将 tab 转换为 2 个空格）
  content = content.replace(/\t/g, '  ')
  
  form.value.content = content
  updateLineCount()
  ElMessage.success('代码格式化完成')
}

const insertTemplate = () => {
  const template = templates[form.value.type]
  if (template) {
    form.value.content = template
    updateLineCount()
    ElMessage.success('模板插入完成')
  }
}

const resetForm = () => {
  form.value = {
    name: '',
    type: 'shell',
    description: '',
    content: ''
  }
  lineCount.value = 1
}

const loadScript = () => {
  if (props.script) {
    form.value = {
      name: props.script.name,
      type: props.script.type,
      description: props.script.description || '',
      content: props.script.content
    }
    nextTick(() => {
      updateLineCount()
    })
  } else {
    resetForm()
  }
}

const handleSave = async () => {
  if (!formRef.value) return
  
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  saving.value = true
  try {
    if (isEdit.value) {
      await api.put(`/api/v1/scripts/${props.script.id}`, form.value)
      ElMessage.success('脚本更新成功')
    } else {
      await api.post('/api/v1/scripts', form.value)
      ElMessage.success('脚本创建成功')
    }
    
    emit('saved')
    dialogVisible.value = false
  } catch (error) {
    ElMessage.error(isEdit.value ? '更新脚本失败' : '创建脚本失败')
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
    loadScript()
  }
})

watch(() => form.value.type, (newType) => {
  // 当类型改变时，如果内容为空，可以提示用户插入模板
  if (!form.value.content.trim()) {
    nextTick(() => {
      ElMessage.info('可以点击"模板"按钮插入代码模板')
    })
  }
})
</script>

<style scoped>
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.option-item {
  display: flex;
  flex-direction: column;
}

.option-label {
  font-weight: 600;
  color: #303133;
}

.option-desc {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

.code-editor-container {
  border: 1px solid #dcdfe6;
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

.code-editor {
  border: none !important;
  border-radius: 0 !important;
}

.code-editor :deep(.el-textarea__inner) {
  border: none !important;
  border-radius: 0 !important;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  resize: none;
}

.code-editor :deep(.el-textarea__inner):focus {
  box-shadow: none !important;
}
</style>
