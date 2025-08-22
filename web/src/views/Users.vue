<template>
  <div class="users-page">
    <div class="page-header">
      <h2>用户管理</h2>
    </div>

    <!-- 搜索和过滤 -->
    <div class="search-section">
      <el-input
        v-model="searchText"
        placeholder="搜索用户名或邮箱"
        style="width: 300px"
        clearable
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-select v-model="roleFilter" placeholder="角色筛选" style="width: 120px">
        <el-option label="全部" value="" />
        <el-option label="管理员" value="admin" />
        <el-option label="用户" value="user" />
      </el-select>
    </div>

    <!-- 用户统计 -->
    <div class="user-stats">
      <div class="stat-item">
        <div class="stat-number">{{ users.length }}</div>
        <div class="stat-label">总用户数</div>
      </div>
      <div class="stat-item">
        <div class="stat-number">{{ getAdminCount() }}</div>
        <div class="stat-label">管理员</div>
      </div>
      <div class="stat-item">
        <div class="stat-number">{{ getRegularUserCount() }}</div>
        <div class="stat-label">普通用户</div>
      </div>
    </div>

    <!-- 用户表格 -->
    <div class="users-table">
      <el-table :data="filteredUsers" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="email" label="邮箱" width="200" />
        <el-table-column prop="role" label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="getRoleType(row.role)">
              {{ getRoleText(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button size="small" @click="editUserRole(row)">
              <el-icon><Edit /></el-icon>
              修改角色
            </el-button>
            <el-button 
              size="small" 
              type="danger" 
              @click="deleteUser(row)"
              :disabled="row.id === currentUser?.id"
            >
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 修改角色对话框 -->
    <el-dialog
      v-model="showRoleDialog"
      title="修改用户角色"
      width="400px"
    >
      <div v-if="editingUser" class="role-edit">
        <p>用户: <strong>{{ editingUser.username }}</strong></p>
        <p>当前角色: 
          <el-tag :type="getRoleType(editingUser.role)">
            {{ getRoleText(editingUser.role) }}
          </el-tag>
        </p>
        <el-form :model="roleForm" label-width="80px">
          <el-form-item label="新角色">
            <el-select v-model="roleForm.role" placeholder="选择角色" style="width: 100%">
              <el-option label="管理员" value="admin" />
              <el-option label="用户" value="user" />
            </el-select>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="showRoleDialog = false">取消</el-button>
        <el-button type="primary" @click="saveUserRole" :loading="saving">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import api from '@/utils/api'

const userStore = useUserStore()
const users = ref([])
const searchText = ref('')
const roleFilter = ref('')
const showRoleDialog = ref(false)
const editingUser = ref(null)
const saving = ref(false)

const roleForm = ref({
  role: ''
})

const currentUser = computed(() => userStore.user)

const filteredUsers = computed(() => {
  return users.value.filter(user => {
    const matchesSearch = !searchText.value || 
      user.username.toLowerCase().includes(searchText.value.toLowerCase()) ||
      user.email.toLowerCase().includes(searchText.value.toLowerCase())
    
    const matchesRole = !roleFilter.value || user.role === roleFilter.value
    
    return matchesSearch && matchesRole
  })
})

const getAdminCount = () => {
  return users.value.filter(user => user.role === 'admin').length
}

const getRegularUserCount = () => {
  return users.value.filter(user => user.role === 'user').length
}

const getRoleType = (role) => {
  const typeMap = {
    admin: 'danger',
    user: 'primary'
  }
  return typeMap[role] || 'info'
}

const getRoleText = (role) => {
  const textMap = {
    admin: '管理员',
    user: '用户'
  }
  return textMap[role] || '未知'
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleString('zh-CN')
}

const loadUsers = async () => {
  try {
    const response = await api.get('/api/v1/admin/users')
    users.value = response.data
  } catch (error) {
    ElMessage.error('加载用户列表失败')
  }
}

const editUserRole = (user) => {
  editingUser.value = user
  roleForm.value.role = user.role
  showRoleDialog.value = true
}

const saveUserRole = async () => {
  if (!editingUser.value) return

  saving.value = true
  try {
    await api.put(`/api/v1/admin/users/${editingUser.value.id}/role`, {
      role: roleForm.value.role
    })
    ElMessage.success('用户角色更新成功')
    showRoleDialog.value = false
    loadUsers()
  } catch (error) {
    ElMessage.error('更新用户角色失败')
  } finally {
    saving.value = false
  }
}

const deleteUser = async (user) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${user.username}" 吗？此操作不可恢复！`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await api.delete(`/api/v1/admin/users/${user.id}`)
    ElMessage.success('用户删除成功')
    loadUsers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除用户失败')
    }
  }
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.users-page {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h2 {
  margin: 0;
  color: #2c3e50;
  font-size: 24px;
  font-weight: 600;
}

.search-section {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
  align-items: center;
}

.user-stats {
  display: flex;
  gap: 20px;
  margin-bottom: 24px;
}

.stat-item {
  background: white;
  border-radius: 8px;
  padding: 16px 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  text-align: center;
  min-width: 100px;
  border-left: 4px solid #1890ff;
}

.stat-number {
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #7f8c8d;
}

.users-table {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.role-edit {
  font-size: 14px;
}

.role-edit p {
  margin-bottom: 12px;
}

.role-edit strong {
  color: #2c3e50;
}
</style>
