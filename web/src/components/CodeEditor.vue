<template>
  <div class="code-editor-container" :class="{ 'fullscreen': isFullscreen }">
    <div class="editor-toolbar">
      <div class="toolbar-left">
        <el-select 
          v-model="currentLanguage" 
          size="small" 
          style="width: 120px"
          @change="handleLanguageChange"
        >
          <el-option label="Shell" value="shell" />
          <el-option label="Python2" value="python2" />
          <el-option label="Python3" value="python3" />
        </el-select>
        
        <el-select 
          v-model="currentTheme" 
          size="small" 
          style="width: 100px; margin-left: 8px"
          @change="handleThemeChange"
        >
          <el-option label="浅色" value="light" />
          <el-option label="深色" value="dark" />
        </el-select>
      </div>
      
      <div class="toolbar-right">
        <el-button size="small" @click="formatCode">
          <el-icon><Tools /></el-icon>
          格式化
        </el-button>
        <el-button size="small" @click="insertTemplate">
          <el-icon><Document /></el-icon>
          模板
        </el-button>
        <el-button size="small" @click="toggleFullscreen">
          <el-icon><FullScreen /></el-icon>
          {{ isFullscreen ? '退出全屏' : '全屏' }}
        </el-button>
      </div>
    </div>
    
    <div class="editor-content">
      <textarea
        ref="textareaRef"
        v-model="content"
        :placeholder="placeholder"
        :readonly="readonly"
        :class="[
          'code-textarea',
          `theme-${currentTheme}`,
          `lang-${currentLanguage}`,
          { 'readonly': readonly }
        ]"
        :style="{ height: editorHeight }"
        @input="handleInput"
        @keydown="handleKeydown"
        @scroll="handleScroll"
      />
      <div class="line-numbers" v-if="showLineNumbers" :style="{ height: editorHeight }">
        <div 
          v-for="line in lineCount" 
          :key="line" 
          class="line-number"
          :class="{ 'active': line === currentLine }"
        >
          {{ line }}
        </div>
      </div>
    </div>
    
    <div class="editor-status">
      <span class="status-item">行: {{ currentLine }}</span>
      <span class="status-item">列: {{ currentColumn }}</span>
      <span class="status-item">长度: {{ contentLength }}</span>
      <span class="status-item">{{ currentLanguage.toUpperCase() }}</span>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch, computed, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Tools, Document, FullScreen } from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  language: {
    type: String,
    default: 'shell'
  },
  theme: {
    type: String,
    default: 'light'
  },
  height: {
    type: [String, Number],
    default: 400
  },
  readonly: {
    type: Boolean,
    default: false
  },
  placeholder: {
    type: String,
    default: '请输入代码...'
  }
})

const emit = defineEmits(['update:modelValue', 'change', 'languageChange'])

const textareaRef = ref(null)
const currentLanguage = ref(props.language)
const currentTheme = ref(props.theme)
const isFullscreen = ref(false)
const showLineNumbers = ref(true)
const currentLine = ref(1)
const currentColumn = ref(1)

const content = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const contentLength = computed(() => props.modelValue.length)
const lineCount = computed(() => props.modelValue.split('\n').length)
const editorHeight = computed(() => {
  if (typeof props.height === 'number') {
    return `${props.height}px`
  }
  return props.height
})

// 脚本模板
const templates = {
  shell: `#!/bin/bash
# Shell脚本模板
echo "Hello World"
echo "当前时间: $(date)"
echo "当前用户: $(whoami)"
echo "系统信息: $(uname -a)"`,
  
  python2: `#!/usr/bin/env python2
# -*- coding: utf-8 -*-
"""
Python2脚本模板
"""
import os
import sys
import datetime

def main():
    print "Hello World"
    print "当前时间: " + str(datetime.datetime.now())
    print "当前用户: " + os.getenv('USER', 'unknown')
    print "Python版本: " + sys.version

if __name__ == "__main__":
    main()`,
    
  python3: `#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Python3脚本模板
"""
import os
import sys
from datetime import datetime

def main():
    print("Hello World")
    print(f"当前时间: {datetime.now()}")
    print(f"当前用户: {os.getenv('USER', 'unknown')}")
    print(f"Python版本: {sys.version}")

if __name__ == "__main__":
    main()`
}

// 处理输入
const handleInput = (event) => {
  const textarea = event.target
  updateCursorPosition(textarea)
  emit('change', content.value)
}

// 处理键盘事件
const handleKeydown = (event) => {
  const textarea = event.target
  
  // Tab键处理
  if (event.key === 'Tab') {
    event.preventDefault()
    const start = textarea.selectionStart
    const end = textarea.selectionEnd
    const value = textarea.value
    
    // 插入两个空格作为缩进
    const newValue = value.substring(0, start) + '  ' + value.substring(end)
    content.value = newValue
    
    nextTick(() => {
      textarea.selectionStart = textarea.selectionEnd = start + 2
    })
  }
  
  // ESC键退出全屏
  if (event.key === 'Escape' && isFullscreen.value) {
    isFullscreen.value = false
  }
  
  nextTick(() => {
    updateCursorPosition(textarea)
  })
}

// 处理滚动
const handleScroll = () => {
  // 同步行号滚动
}

// 更新光标位置
const updateCursorPosition = (textarea) => {
  const cursorPos = textarea.selectionStart
  const textBeforeCursor = textarea.value.substring(0, cursorPos)
  const lines = textBeforeCursor.split('\n')
  
  currentLine.value = lines.length
  currentColumn.value = lines[lines.length - 1].length + 1
}

// 处理语言变化
const handleLanguageChange = (newLanguage) => {
  emit('languageChange', newLanguage)
}

// 处理主题变化
const handleThemeChange = (newTheme) => {
  // 主题变化通过CSS类处理
}

// 格式化代码
const formatCode = () => {
  if (!content.value.trim()) {
    ElMessage.warning('请先输入代码内容')
    return
  }
  
  // 简单的代码格式化
  let formattedContent = content.value
  
  // 移除多余的空行
  formattedContent = formattedContent.replace(/\n\s*\n\s*\n/g, '\n\n')
  
  // 统一缩进（将tab转换为2个空格）
  formattedContent = formattedContent.replace(/\t/g, '  ')
  
  content.value = formattedContent
  ElMessage.success('代码格式化完成')
}

// 插入模板
const insertTemplate = () => {
  const template = templates[currentLanguage.value]
  if (template) {
    content.value = template
    ElMessage.success('模板插入成功')
  } else {
    ElMessage.warning('当前语言暂无模板')
  }
}

// 切换全屏
const toggleFullscreen = () => {
  isFullscreen.value = !isFullscreen.value
}

// 全局键盘事件处理
const handleGlobalKeydown = (event) => {
  if (event.key === 'Escape' && isFullscreen.value) {
    event.preventDefault()
    event.stopPropagation()
    isFullscreen.value = false
  }
}

// 监听props变化
watch(() => props.language, (newLanguage) => {
  currentLanguage.value = newLanguage
})

watch(() => props.theme, (newTheme) => {
  currentTheme.value = newTheme
})

onMounted(() => {
  if (textareaRef.value) {
    updateCursorPosition(textareaRef.value)
  }
  // 添加全局键盘事件监听
  document.addEventListener('keydown', handleGlobalKeydown)
})

onUnmounted(() => {
  // 移除全局键盘事件监听
  document.removeEventListener('keydown', handleGlobalKeydown)
})

// 暴露方法
defineExpose({
  getValue: () => content.value,
  setValue: (value) => { content.value = value },
  focus: () => textareaRef.value?.focus(),
  format: formatCode,
  insertTemplate
})
</script>

<style scoped>
.code-editor-container {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  overflow: hidden;
  background: #fff;
}

.editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
}

.toolbar-left {
  display: flex;
  align-items: center;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.editor-content {
  position: relative;
  display: flex;
}

.editor-content.fullscreen {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
  background: #fff;
}

.code-textarea {
  width: 100%;
  border: none;
  outline: none;
  resize: none;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  padding: 12px 12px 12px 60px;
  background: transparent;
  color: #303133;
  tab-size: 2;
}

.code-textarea.theme-dark {
  background: #1e1e1e;
  color: #d4d4d4;
}

.code-textarea.theme-light {
  background: #ffffff;
  color: #303133;
}

.code-textarea.readonly {
  background: #f5f7fa;
  color: #909399;
  cursor: not-allowed;
}

.line-numbers {
  position: absolute;
  left: 0;
  top: 0;
  width: 50px;
  background: #f8f9fa;
  border-right: 1px solid #e4e7ed;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
  color: #909399;
  user-select: none;
  overflow: hidden;
  padding-top: 12px;
}

.theme-dark .line-numbers {
  background: #252526;
  border-right-color: #3e3e42;
  color: #858585;
}

.line-number {
  height: 21px;
  padding: 0 8px;
  text-align: right;
  font-size: 12px;
  line-height: 21px;
}

.line-number.active {
  background: rgba(64, 158, 255, 0.1);
  color: #409eff;
}

.editor-status {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  padding: 4px 12px;
  background: #f5f7fa;
  border-top: 1px solid #e4e7ed;
  font-size: 12px;
  color: #909399;
}

.status-item {
  margin-left: 16px;
}

.status-item:first-child {
  margin-left: 0;
}

/* 全屏样式 */
.code-editor-container.fullscreen {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
  background: #fff;
  border-radius: 0;
}

.code-editor-container.fullscreen .editor-toolbar {
  position: relative;
  z-index: 10001;
}

.code-editor-container.fullscreen .editor-content {
  height: calc(100vh - 100px);
}

.code-editor-container.fullscreen .editor-status {
  position: relative;
  z-index: 10001;
}

/* 语法高亮样式预留 */

/* 滚动条样式 */
.code-textarea::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.code-textarea::-webkit-scrollbar-track {
  background: #f1f1f1;
}

.code-textarea::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
}

.code-textarea::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}
</style>
