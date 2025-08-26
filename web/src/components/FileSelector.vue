<template>
  <div class="file-selector">
    <div class="file-selector-header">
      <h4>{{ title }}</h4>
    </div>

    <div class="selected-files" v-if="selectedFiles.length > 0">
      <div class="file-item" v-for="file in selectedFiles" :key="file.id">
        <div class="file-info">
          <el-icon class="file-icon">
            <Document />
          </el-icon>
          <div class="file-details">
            <div class="file-name">{{ file.original_name }}</div>
            <div class="file-meta">
              <span class="file-size">{{ formatFileSize(file.size) }}</span>
              <span class="file-category">{{ getCategoryLabel(file.category) }}</span>
              <span class="file-date">{{ formatDate(file.created_at) }}</span>
            </div>
          </div>
        </div>
        <el-button 
          type="danger" 
          size="small" 
          text 
          @click="removeFile(file.id)"
          :icon="Close"
        >
          移除
        </el-button>
      </div>
    </div>

    <div class="empty-state" v-else>
      <el-empty 
        :image-size="80" 
        description="暂未选择文件"
      />
    </div>

    <!-- 文件选择对话框 -->
    <el-dialog
      v-model="showFileDialog"
      title="选择文件"
      width="800px"
      :close-on-click-modal="false"
    >
      <div class="file-dialog-content">
        <!-- 搜索和筛选 -->
        <div class="file-filters">
          <el-input
            v-model="searchQuery"
            placeholder="搜索文件名..."
            :prefix-icon="Search"
            clearable
            @input="handleSearch"
            style="width: 300px; margin-right: 16px;"
          />
          <el-select
            v-model="selectedCategory"
            placeholder="选择分类"
            clearable
            @change="handleCategoryChange"
            style="width: 150px;"
          >
            <el-option label="全部" value="" />
            <el-option label="脚本" value="script" />
            <el-option label="配置" value="config" />
            <el-option label="软件包" value="package" />
            <el-option label="通用" value="general" />
          </el-select>
        </div>

        <!-- 文件列表 -->
        <div class="file-list">
          <el-table
            :data="filteredFiles"
            v-loading="loading"
            @selection-change="handleSelectionChange"
            max-height="400"
          >
            <el-table-column type="selection" width="55" />
            <el-table-column prop="original_name" label="文件名" min-width="200">
              <template #default="{ row }">
                <div class="file-name-cell">
                  <el-icon class="file-icon">
                    <Document />
                  </el-icon>
                  <span>{{ row.original_name }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="size" label="大小" width="100">
              <template #default="{ row }">
                {{ formatFileSize(row.size) }}
              </template>
            </el-table-column>
            <el-table-column prop="category" label="分类" width="100">
              <template #default="{ row }">
                <el-tag :type="getCategoryType(row.category)" size="small">
                  {{ getCategoryLabel(row.category) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="上传时间" width="150">
              <template #default="{ row }">
                {{ formatDate(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
          </el-table>
        </div>

        <!-- 分页 -->
        <div class="pagination-container">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="total"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
          />
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showFileDialog = false">取消</el-button>
          <el-button type="primary" @click="confirmSelection">
            确定选择 ({{ tempSelectedFiles.length }})
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Close, Search, Document } from '@element-plus/icons-vue'
import api from '@/utils/api'

// Props
const props = defineProps({
  title: {
    type: String,
    default: '输入文件'
  },
  modelValue: {
    type: Array,
    default: () => []
  },
  multiple: {
    type: Boolean,
    default: true
  },
  categoryFilter: {
    type: Array,
    default: () => []
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'change'])

// 暴露方法给父组件
const openFileDialog = () => {
  showFileDialog.value = true
  loadFiles()
}

defineExpose({
  openFileDialog
})

// 响应式数据
const showFileDialog = ref(false)
const loading = ref(false)
const searchQuery = ref('')
const selectedCategory = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

const files = ref([])
const selectedFiles = ref([])
const tempSelectedFiles = ref([])

// 计算属性
const filteredFiles = computed(() => {
  let result = files.value
  
  if (searchQuery.value) {
    result = result.filter(file => 
      file.original_name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      file.description.toLowerCase().includes(searchQuery.value.toLowerCase())
    )
  }
  
  if (selectedCategory.value) {
    result = result.filter(file => file.category === selectedCategory.value)
  }
  
  if (props.categoryFilter.length > 0) {
    result = result.filter(file => props.categoryFilter.includes(file.category))
  }
  
  return result
})

// 监听器
watch(() => props.modelValue, (newValue) => {
  selectedFiles.value = newValue || []
}, { immediate: true })

watch(selectedFiles, (newValue) => {
  emit('update:modelValue', newValue)
  emit('change', newValue)
}, { deep: true })

// 方法
const loadFiles = async () => {
  loading.value = true
  try {
    const response = await api.get('/api/v1/files', {
      params: {
        page: currentPage.value,
        page_size: pageSize.value,
        category: selectedCategory.value || undefined,
        search: searchQuery.value || undefined
      }
    })
    
    files.value = response.data.data || []
    total.value = response.data.total || 0
  } catch (error) {
    console.error('加载文件列表失败:', error)
    ElMessage.error('加载文件列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  loadFiles()
}

const handleCategoryChange = () => {
  currentPage.value = 1
  loadFiles()
}

const handleSizeChange = () => {
  currentPage.value = 1
  loadFiles()
}

const handleCurrentChange = () => {
  loadFiles()
}

const handleSelectionChange = (selection) => {
  tempSelectedFiles.value = selection
}

const confirmSelection = () => {
  if (!props.multiple && tempSelectedFiles.value.length > 1) {
    ElMessage.warning('只能选择一个文件')
    return
  }
  
  // 合并已选择的文件和新选择的文件
  const existingIds = selectedFiles.value.map(f => f.id)
  const newFiles = tempSelectedFiles.value.filter(f => !existingIds.includes(f.id))
  
  selectedFiles.value = [...selectedFiles.value, ...newFiles]
  showFileDialog.value = false
  tempSelectedFiles.value = []
}

const removeFile = (fileId) => {
  selectedFiles.value = selectedFiles.value.filter(f => f.id !== fileId)
}

const formatFileSize = (size) => {
  if (!size) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let index = 0
  let fileSize = size
  
  while (fileSize >= 1024 && index < units.length - 1) {
    fileSize /= 1024
    index++
  }
  
  return `${fileSize.toFixed(1)} ${units[index]}`
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN')
}

const getCategoryLabel = (category) => {
  const labels = {
    script: '脚本',
    config: '配置',
    package: '软件包',
    general: '通用',
    script_output: '脚本输出',
    error_log: '错误日志'
  }
  return labels[category] || category
}

const getCategoryType = (category) => {
  const types = {
    script: 'primary',
    config: 'success',
    package: 'warning',
    general: 'info',
    script_output: 'primary',
    error_log: 'danger'
  }
  return types[category] || 'info'
}

// 生命周期
onMounted(() => {
  loadFiles()
})
</script>

<style scoped>
.file-selector {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
  background: #fff;
}

.file-selector-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.file-selector-header h4 {
  margin: 0;
  color: #303133;
  font-size: 16px;
  font-weight: 600;
}

.selected-files {
  max-height: 300px;
  overflow-y: auto;
}

.file-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  margin-bottom: 8px;
  background: #fafafa;
  transition: all 0.3s;
}

.file-item:hover {
  background: #f0f9ff;
  border-color: #409eff;
}

.file-info {
  display: flex;
  align-items: center;
  flex: 1;
}

.file-icon {
  color: #409eff;
  margin-right: 12px;
  font-size: 20px;
}

.file-details {
  flex: 1;
}

.file-name {
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.file-meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #909399;
}

.empty-state {
  padding: 40px 0;
}

.file-dialog-content {
  max-height: 600px;
}

.file-filters {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e4e7ed;
}

.file-list {
  margin-bottom: 16px;
}

.file-name-cell {
  display: flex;
  align-items: center;
}

.file-name-cell .file-icon {
  margin-right: 8px;
  font-size: 16px;
}

.pagination-container {
  display: flex;
  justify-content: center;
  padding-top: 16px;
  border-top: 1px solid #e4e7ed;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
