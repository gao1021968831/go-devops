<template>
  <div class="scripts-page">
    <el-card class="page-card">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <h2>脚本管理</h2>
            <p>管理和执行系统脚本，支持多种脚本类型</p>
          </div>
          <div class="header-actions">
            <el-button type="primary" @click="showQuickExecute = true" size="large">
              <el-icon><VideoPlay /></el-icon>
              快速执行
            </el-button>
          </div>
        </div>
      </template>

      <el-tabs v-model="activeTab" class="scripts-tabs">
        <el-tab-pane label="脚本管理" name="scripts">
          <ScriptManager ref="scriptManagerRef" @create-job="handleCreateJob" />
        </el-tab-pane>
        <el-tab-pane label="作业管理" name="jobs">
          <JobManager ref="jobManagerRef" />
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 快速执行对话框 -->
    <QuickExecute 
      v-model:visible="showQuickExecute" 
      :prefill-data="quickExecutePrefillData"
      @executed="handleQuickExecuted"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import ScriptManager from '@/components/ScriptManager.vue'
import JobManager from '@/components/JobManager.vue'
import QuickExecute from '@/components/QuickExecute.vue'
import { VideoPlay } from '@element-plus/icons-vue'

const route = useRoute()

// 响应式数据
const activeTab = ref('scripts')
const jobManagerRef = ref()
const showQuickExecute = ref(false)
const quickExecutePrefillData = ref(null)

// 处理从脚本管理组件创建作业的事件
const handleCreateJob = (script) => {
  // 切换到作业管理标签页
  activeTab.value = 'jobs'
  
  // 调用作业管理组件的创建方法
  if (jobManagerRef.value) {
    jobManagerRef.value.createJobForScript(script)
  }
}

// 处理快速执行完成事件
const handleQuickExecuted = (executions) => {
  ElMessage.success(`快速执行已启动，共创建 ${executions.length} 个执行任务`)
  
  // 切换到作业管理标签页查看执行结果
  activeTab.value = 'jobs'
  
  // 刷新作业管理列表
  if (jobManagerRef.value) {
    jobManagerRef.value.refreshExecutions()
  }
}

// 处理路由参数，检查是否需要自动打开快速执行对话框
onMounted(() => {
  if (route.query.action === 'quick-execute' && route.query.prefill) {
    try {
      quickExecutePrefillData.value = JSON.parse(route.query.prefill)
      showQuickExecute.value = true
    } catch (error) {
      console.error('解析预填充数据失败:', error)
    }
  }
})
</script>

<style scoped>
.scripts-page {
  padding: 24px;
  background: #f5f7fa;
  min-height: 100vh;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.header-left h2 {
  margin: 0 0 8px 0;
  color: #303133;
  font-size: 24px;
  font-weight: 600;
}

.header-left p {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.header-actions {
  flex-shrink: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.header-left {
  flex: 1;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.title-icon {
  font-size: 28px;
  color: #409eff;
}

.page-title h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.page-subtitle {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.main-tabs {
  min-height: 600px;
}

.main-tabs :deep(.el-tabs__content) {
  padding: 20px 0;
}

.main-tabs :deep(.el-tab-pane) {
  min-height: 500px;
}
</style>
