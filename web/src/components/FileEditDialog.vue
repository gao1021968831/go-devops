<template>
  <el-dialog
    v-model="visible"
    title="编辑文件信息"
    width="500px"
    :before-close="handleClose"
  >
    <div class="edit-container" v-if="file">
      <el-form :model="editForm" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="文件名" prop="name">
          <el-input v-model="editForm.name" placeholder="请输入文件名" />
        </el-form-item>
        
        <el-form-item label="文件分类" prop="category">
          <el-select v-model="editForm.category" placeholder="选择分类" style="width: 100%">
            <el-option label="通用文件" value="general" />
            <el-option label="脚本文件" value="scripts" />
            <el-option label="配置文件" value="configs" />
            <el-option label="文档文件" value="documents" />
            <el-option label="其他" value="others" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="访问权限">
          <el-switch
            v-model="editForm.isPublic"
            active-text="公开"
            inactive-text="私有"
          />
        </el-form-item>
        
        <el-form-item label="文件描述">
          <el-input
            v-model="editForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入文件描述（可选）"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" @click="saveChanges" :loading="saving">
          {{ saving ? '保存中...' : '保存' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'

// Props & Emits
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  file: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:modelValue', 'updated'])

// 响应式数据
const formRef = ref()
const saving = ref(false)

const editForm = reactive({
  name: '',
  category: '',
  description: '',
  isPublic: false
})

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '请输入文件名', trigger: 'blur' },
    { min: 1, max: 255, message: '文件名长度应在1-255字符之间', trigger: 'blur' }
  ],
  category: [
    { required: true, message: '请选择文件分类', trigger: 'change' }
  ]
}

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

// 保存更改
const saveChanges = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
  } catch (error) {
    return
  }

  saving.value = true
  try {
    const data = {
      name: editForm.name,
      category: editForm.category,
      description: editForm.description,
      is_public: editForm.isPublic
    }

    await api.put(`/api/v1/files/${props.file.id}`, data)
    
    ElMessage.success('文件信息更新成功')
    emit('updated')
    handleClose()
  } catch (error) {
    ElMessage.error('文件信息更新失败')
  } finally {
    saving.value = false
  }
}

// 关闭对话框
const handleClose = () => {
  if (saving.value) {
    ElMessage.warning('正在保存中，请稍候...')
    return
  }

  visible.value = false
}

// 监听文件变化，初始化表单
watch(() => props.file, (newFile) => {
  if (newFile) {
    editForm.name = newFile.original_name || ''
    editForm.category = newFile.category || 'general'
    editForm.description = newFile.description || ''
    editForm.isPublic = newFile.is_public || false
  }
}, { immediate: true })

// 监听对话框显示状态
watch(visible, (newVal) => {
  if (newVal && props.file) {
    editForm.name = props.file.original_name || ''
    editForm.category = props.file.category || 'general'
    editForm.description = props.file.description || ''
    editForm.isPublic = props.file.is_public || false
  }
})
</script>

<style scoped>
.edit-container {
  padding: 20px 0;
}

.dialog-footer {
  text-align: right;
}
</style>
