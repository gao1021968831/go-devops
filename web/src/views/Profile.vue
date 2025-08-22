<template>
  <div class="profile-container">
    <div class="profile-header">
      <h1>个人资料</h1>
      <p class="header-desc">管理您的个人信息和账户设置</p>
    </div>

    <div class="profile-content">
      <el-row :gutter="24">
        <!-- 基本信息卡片 -->
        <el-col :span="12">
          <el-card class="profile-card" shadow="hover">
            <template #header>
              <div class="card-header">
                <el-icon><User /></el-icon>
                <span>基本信息</span>
              </div>
            </template>
            
            <el-form 
              :model="userInfo" 
              :rules="userRules" 
              ref="userFormRef"
              label-width="80px"
              class="profile-form"
            >
              <el-form-item label="用户名" prop="username">
                <el-input 
                  v-model="userInfo.username" 
                  :disabled="!editMode"
                  placeholder="请输入用户名"
                />
              </el-form-item>
              
              <el-form-item label="邮箱" prop="email">
                <el-input 
                  v-model="userInfo.email" 
                  :disabled="!editMode"
                  placeholder="请输入邮箱"
                />
              </el-form-item>
              
              <el-form-item label="角色">
                <el-tag :type="getRoleType(userInfo.role)">
                  {{ getRoleText(userInfo.role) }}
                </el-tag>
              </el-form-item>
              
              <el-form-item label="注册时间">
                <span class="info-text">{{ formatDate(userInfo.created_at) }}</span>
              </el-form-item>
            </el-form>

            <div class="card-actions">
              <el-button 
                v-if="!editMode" 
                type="primary" 
                @click="enableEdit"
                :icon="Edit"
              >
                编辑信息
              </el-button>
              <template v-else>
                <el-button 
                  type="primary" 
                  @click="saveUserInfo"
                  :loading="saving"
                  :icon="Check"
                >
                  保存
                </el-button>
                <el-button @click="cancelEdit" :icon="Close">
                  取消
                </el-button>
              </template>
            </div>
          </el-card>
        </el-col>

        <!-- 修改密码卡片 -->
        <el-col :span="12">
          <el-card class="profile-card" shadow="hover">
            <template #header>
              <div class="card-header">
                <el-icon><Lock /></el-icon>
                <span>修改密码</span>
              </div>
            </template>
            
            <el-form 
              :model="passwordForm" 
              :rules="passwordRules" 
              ref="passwordFormRef"
              label-width="80px"
              class="profile-form"
            >
              <el-form-item label="旧密码" prop="oldPassword">
                <el-input 
                  v-model="passwordForm.oldPassword" 
                  type="password"
                  placeholder="请输入当前密码"
                  show-password
                />
              </el-form-item>
              
              <el-form-item label="新密码" prop="newPassword">
                <el-input 
                  v-model="passwordForm.newPassword" 
                  type="password"
                  placeholder="请输入新密码"
                  show-password
                />
              </el-form-item>
              
              <el-form-item label="确认密码" prop="confirmPassword">
                <el-input 
                  v-model="passwordForm.confirmPassword" 
                  type="password"
                  placeholder="请再次输入新密码"
                  show-password
                />
              </el-form-item>
            </el-form>

            <div class="card-actions">
              <el-button 
                type="primary" 
                @click="changePassword"
                :loading="changingPassword"
                :icon="Key"
              >
                修改密码
              </el-button>
              <el-button @click="resetPasswordForm" :icon="Refresh">
                重置
              </el-button>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 账户统计信息 -->
      <el-row :gutter="24" class="stats-row">
        <el-col :span="24">
          <el-card class="stats-card" shadow="hover">
            <template #header>
              <div class="card-header">
                <el-icon><DataAnalysis /></el-icon>
                <span>账户统计</span>
              </div>
            </template>
            
            <el-row :gutter="16">
              <el-col :span="6">
                <div class="stat-item">
                  <div class="stat-value">{{ stats.script_count }}</div>
                  <div class="stat-label">创建的脚本</div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="stat-item">
                  <div class="stat-value">{{ stats.execution_count }}</div>
                  <div class="stat-label">执行次数</div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="stat-item">
                  <div class="stat-value">{{ stats.last_login_days }}</div>
                  <div class="stat-label">天前登录</div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="stat-item">
                  <div class="stat-value">{{ formatDate(userInfo.updated_at, 'date') }}</div>
                  <div class="stat-label">最后更新</div>
                </div>
              </el-col>
            </el-row>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  User, 
  Lock, 
  Edit, 
  Check, 
  Close, 
  Key, 
  Refresh, 
  DataAnalysis 
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import api from '@/utils/api'

const userStore = useUserStore()

// 响应式数据
const editMode = ref(false)
const saving = ref(false)
const changingPassword = ref(false)

const userInfo = reactive({
  username: '',
  email: '',
  role: '',
  created_at: '',
  updated_at: ''
})

const originalUserInfo = reactive({})

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const stats = reactive({
  script_count: 0,
  execution_count: 0,
  last_login_days: 0
})

// 表单引用
const userFormRef = ref()
const passwordFormRef = ref()

// 表单验证规则
const userRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在3到20个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ]
}

const validateConfirmPassword = (rule, value, callback) => {
  if (value !== passwordForm.newPassword) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const passwordRules = {
  oldPassword: [
    { required: true, message: '请输入当前密码', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

// 方法
const loadUserProfile = async () => {
  try {
    const response = await api.get('/api/v1/profile')
    Object.assign(userInfo, response.data)
    Object.assign(originalUserInfo, response.data)
  } catch (error) {
    ElMessage.error('获取用户信息失败')
  }
}

const loadUserStats = async () => {
  try {
    const response = await api.get('/api/v1/profile/stats')
    Object.assign(stats, response.data)
  } catch (error) {
    console.error('获取统计信息失败:', error)
    ElMessage.error('获取统计信息失败')
  }
}

const enableEdit = () => {
  editMode.value = true
}

const cancelEdit = () => {
  editMode.value = false
  Object.assign(userInfo, originalUserInfo)
}

const saveUserInfo = async () => {
  if (!userFormRef.value) return
  
  try {
    await userFormRef.value.validate()
    saving.value = true
    
    const updateData = {
      username: userInfo.username,
      email: userInfo.email
    }
    
    await api.put('/api/v1/profile', updateData)
    
    Object.assign(originalUserInfo, userInfo)
    editMode.value = false
    
    // 更新用户store中的信息
    userStore.updateUserInfo(userInfo)
    
    ElMessage.success('个人信息更新成功')
  } catch (error) {
    if (error.response?.data?.error) {
      ElMessage.error(error.response.data.error)
    } else {
      ElMessage.error('更新个人信息失败')
    }
  } finally {
    saving.value = false
  }
}

const changePassword = async () => {
  if (!passwordFormRef.value) return
  
  try {
    await passwordFormRef.value.validate()
    
    await ElMessageBox.confirm(
      '确定要修改密码吗？修改后需要重新登录。',
      '确认修改',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    changingPassword.value = true
    
    await api.put('/api/v1/profile/password', {
      old_password: passwordForm.oldPassword,
      new_password: passwordForm.newPassword
    })
    
    ElMessage.success('密码修改成功，即将跳转到登录页面')
    
    // 清空密码表单
    resetPasswordForm()
    
    // 立即清除用户状态并跳转到登录页
    setTimeout(() => {
      userStore.logout()
      // 强制刷新页面到登录页面
      window.location.href = '/login'
    }, 1000)
    
  } catch (error) {
    if (error === 'cancel') {
      return
    }
    if (error.response?.data?.error) {
      ElMessage.error(error.response.data.error)
    } else {
      ElMessage.error('修改密码失败')
    }
  } finally {
    changingPassword.value = false
  }
}

const resetPasswordForm = () => {
  passwordForm.oldPassword = ''
  passwordForm.newPassword = ''
  passwordForm.confirmPassword = ''
  passwordFormRef.value?.clearValidate()
}

const getRoleType = (role) => {
  const roleTypes = {
    admin: 'danger',
    user: 'primary'
  }
  return roleTypes[role] || 'info'
}

const getRoleText = (role) => {
  const roleTexts = {
    admin: '管理员',
    user: '普通用户'
  }
  return roleTexts[role] || '未知'
}

const formatDate = (dateString, type = 'datetime') => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  if (type === 'date') {
    return date.toLocaleDateString('zh-CN')
  }
  return date.toLocaleString('zh-CN')
}

// 生命周期
onMounted(() => {
  loadUserProfile().then(() => {
    loadUserStats()
  })
})
</script>

<style scoped>
.profile-container {
  padding: 24px;
  background-color: #f5f7fa;
  min-height: calc(100vh - 60px);
}

.profile-header {
  margin-bottom: 24px;
  text-align: center;
}

.profile-header h1 {
  margin: 0 0 8px 0;
  font-size: 28px;
  font-weight: 600;
  color: #303133;
}

.header-desc {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.profile-content {
  max-width: 1200px;
  margin: 0 auto;
}

.profile-card {
  margin-bottom: 24px;
  border-radius: 12px;
  overflow: hidden;
}

.profile-card :deep(.el-card__header) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 16px 20px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 16px;
}

.profile-form {
  padding: 8px 0;
}

.profile-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
}

.info-text {
  color: #909399;
  font-size: 14px;
}

.card-actions {
  display: flex;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid #ebeef5;
  margin-top: 16px;
}

.stats-row {
  margin-top: 24px;
}

.stats-card {
  border-radius: 12px;
  overflow: hidden;
}

.stats-card :deep(.el-card__header) {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  color: white;
  padding: 16px 20px;
}

.stat-item {
  text-align: center;
  padding: 16px;
  border-radius: 8px;
  background: #f8f9fa;
  transition: all 0.3s ease;
}

.stat-item:hover {
  background: #e9ecef;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #409eff;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #909399;
  font-weight: 500;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .profile-container {
    padding: 16px;
  }
  
  .profile-content :deep(.el-col) {
    margin-bottom: 16px;
  }
  
  .card-actions {
    flex-direction: column;
  }
  
  .card-actions .el-button {
    width: 100%;
  }
}

/* 动画效果 */
.profile-card {
  transition: all 0.3s ease;
}

.profile-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
}

.el-button {
  transition: all 0.3s ease;
}

.el-button:hover {
  transform: translateY(-1px);
}
</style>
