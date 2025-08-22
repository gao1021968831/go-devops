<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon hosts">
          <el-icon><Monitor /></el-icon>
        </div>
        <div class="stat-content">
          <h3>{{ stats.totalHosts }}</h3>
          <p>主机总数</p>
          <span class="stat-trend">在线: {{ stats.onlineHosts }}</span>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon jobs">
          <el-icon><Operation /></el-icon>
        </div>
        <div class="stat-content">
          <h3>{{ stats.totalJobs }}</h3>
          <p>作业总数</p>
          <span class="stat-trend">运行中: {{ stats.runningJobs }}</span>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon scripts">
          <el-icon><Document /></el-icon>
        </div>
        <div class="stat-content">
          <h3>{{ stats.totalScripts }}</h3>
          <p>脚本总数</p>
        </div>
      </div>
      
    </div>

    <!-- 图表区域 -->
    <div class="charts-grid">
      <div class="chart-card">
        <div class="card-header">
          <h3>主机状态分布</h3>
        </div>
        <div class="chart-container">
          <v-chart :option="hostStatusChart" style="height: 300px;" />
        </div>
      </div>
      
      <div class="chart-card">
        <div class="card-header">
          <h3>作业执行趋势</h3>
        </div>
        <div class="chart-container">
          <v-chart :option="jobTrendChart" style="height: 300px;" />
        </div>
      </div>
    </div>

    <!-- 最近活动 -->
    <div class="activity-section">
      <div class="card">
        <div class="card-header">
          <h3>最近活动</h3>
          <div class="header-actions">
            <el-button type="text" size="small" @click="showAllActivities">
              查看全部
            </el-button>
            <el-button type="primary" size="small" @click="refreshData">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>
        <el-timeline>
          <el-timeline-item
            v-for="activity in recentActivities"
            :key="activity.id"
            :timestamp="activity.timestamp"
            :type="activity.type"
          >
            {{ activity.message }}
          </el-timeline-item>
        </el-timeline>
        <div v-if="recentActivities.length === 0" class="empty-state">
          <el-empty description="暂无活动记录" />
        </div>
      </div>
    </div>
  </div>

  <!-- 全部活动对话框 -->
  <el-dialog
    v-model="activitiesDialogVisible"
    title="全部活动"
    width="900px"
    top="5vh"
    :before-close="handleCloseActivitiesDialog"
    class="activities-dialog-wrapper"
  >
    <div class="activities-dialog">
      <!-- 筛选器 -->
      <div class="filter-bar">
        <el-row :gutter="16">
          <el-col :span="8">
            <el-select v-model="activityFilter.type" placeholder="活动类型" clearable @change="loadAllActivities">
              <el-option label="全部" value="" />
              <el-option label="成功" value="success" />
              <el-option label="错误" value="error" />
              <el-option label="警告" value="warning" />
              <el-option label="信息" value="info" />
            </el-select>
          </el-col>
          <el-col :span="8">
            <el-date-picker
              v-model="activityFilter.dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
              @change="loadAllActivities"
            />
          </el-col>
          <el-col :span="8">
            <el-input
              v-model="activityFilter.keyword"
              placeholder="搜索关键词"
              clearable
              @input="debounceSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </el-col>
        </el-row>
      </div>

      <!-- 活动列表 -->
      <div class="activities-list" v-loading="activitiesLoading">
        <el-timeline>
          <el-timeline-item
            v-for="activity in allActivities"
            :key="activity.id"
            :timestamp="activity.timestamp"
            :type="activity.type"
          >
            <div class="activity-content">
              <div class="activity-message">{{ activity.message }}</div>
              <div class="activity-meta">
                <span class="activity-user" v-if="activity.user">用户: {{ activity.user }}</span>
                <span class="activity-resource" v-if="activity.resource">资源: {{ activity.resource }}</span>
              </div>
            </div>
          </el-timeline-item>
        </el-timeline>
        
        <div v-if="allActivities.length === 0 && !activitiesLoading" class="empty-state">
          <el-empty description="暂无活动记录">
            <template #image>
              <div class="empty-icon">
                <el-icon size="64" color="#d9d9d9"><Document /></el-icon>
              </div>
            </template>
            <template #description>
              <p class="empty-description">暂无活动记录</p>
              <p class="empty-tip">系统中还没有任何作业执行记录</p>
            </template>
          </el-empty>
        </div>
      </div>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="activityPagination.page"
          v-model:page-size="activityPagination.size"
          :page-sizes="[10, 20, 50, 100]"
          :total="activityPagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadAllActivities"
          @current-change="loadAllActivities"
        />
      </div>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart, LineChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import api from '@/utils/api'
import { ElMessage } from 'element-plus'
import { debounce } from 'lodash-es'
import { Document } from '@element-plus/icons-vue'

use([
  CanvasRenderer,
  PieChart,
  LineChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
])

const stats = ref({
  totalHosts: 0,
  onlineHosts: 0,
  totalJobs: 0,
  runningJobs: 0,
  totalScripts: 0
})

const recentActivities = ref([])
const allActivities = ref([])
const activitiesDialogVisible = ref(false)
const activitiesLoading = ref(false)

// 活动筛选器
const activityFilter = ref({
  type: '',
  dateRange: null,
  keyword: ''
})

// 分页信息
const activityPagination = ref({
  page: 1,
  size: 20,
  total: 0
})

// 将活动类型映射为ElTimelineItem支持的类型
const getTimelineItemType = (type) => {
  const typeMap = {
    'error': 'danger',
    'success': 'success', 
    'warning': 'warning',
    'info': 'info',
    'primary': 'primary'
  }
  return typeMap[type] || 'info'
}

const hostStatusChart = ref({
  tooltip: {
    trigger: 'item'
  },
  legend: {
    bottom: '0%',
    left: 'center'
  },
  series: [
    {
      name: '主机状态',
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 10,
        borderColor: '#fff',
        borderWidth: 2
      },
      label: {
        show: false,
        position: 'center'
      },
      emphasis: {
        label: {
          show: true,
          fontSize: 20,
          fontWeight: 'bold'
        }
      },
      labelLine: {
        show: false
      },
      data: [
        { value: 0, name: '在线', itemStyle: { color: '#52c41a' } },
        { value: 0, name: '离线', itemStyle: { color: '#ff4d4f' } },
        { value: 0, name: '未知', itemStyle: { color: '#faad14' } }
      ]
    }
  ]
})

const jobTrendChart = ref({
  tooltip: {
    trigger: 'axis'
  },
  legend: {
    data: ['成功', '失败']
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true
  },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: []
  },
  yAxis: {
    type: 'value'
  },
  series: [
    {
      name: '成功',
      type: 'line',
      stack: 'Total',
      itemStyle: { color: '#52c41a' },
      data: []
    },
    {
      name: '失败',
      type: 'line',
      stack: 'Total',
      itemStyle: { color: '#ff4d4f' },
      data: []
    }
  ]
})

const loadDashboardData = async () => {
  // 加载仪表盘统计数据
  try {
    const statsResponse = await api.get('/api/v1/dashboard/stats')
    const dashboardStats = statsResponse.data
    stats.value.totalHosts = dashboardStats.total_hosts || 0
    stats.value.onlineHosts = dashboardStats.online_hosts || 0
    stats.value.totalJobs = dashboardStats.total_jobs || 0
    stats.value.runningJobs = dashboardStats.running_jobs || 0
    stats.value.totalScripts = dashboardStats.total_scripts || 0
  } catch (error) {
    console.error('加载统计数据失败:', error)
    ElMessage.warning('加载统计数据失败')
  }

  // 加载主机状态分布
  try {
    const hostStatusResponse = await api.get('/api/v1/dashboard/host-status')
    const hostStatusData = hostStatusResponse.data || []
    hostStatusChart.value.series[0].data = hostStatusData.map(item => ({
      value: item.count,
      name: item.status === 'online' ? '在线' : item.status === 'offline' ? '离线' : '未知'
    }))
  } catch (error) {
    console.error('加载主机状态分布失败:', error)
    ElMessage.warning('加载主机状态分布失败')
  }

  // 获取最近活动
  try {
    const activitiesResponse = await api.get('/api/v1/dashboard/recent-activities?limit=10')
    const activitiesData = activitiesResponse.data || []
    recentActivities.value = activitiesData.map(activity => ({
      id: activity.id,
      message: activity.message,
      type: getTimelineItemType(activity.type),
      timestamp: activity.timestamp
    }))
  } catch (error) {
    console.error('加载最近活动失败:', error)
    ElMessage.warning('加载最近活动失败')
    recentActivities.value = []
  }

  // 加载作业趋势数据
  try {
    const trendResponse = await api.get('/api/v1/dashboard/job-trend?days=7')
    const trendData = trendResponse.data || []
    
    jobTrendChart.value.xAxis.data = trendData.map(item => item.date)
    jobTrendChart.value.series[0].data = trendData.map(item => item.success)
    jobTrendChart.value.series[1].data = trendData.map(item => item.failed)
  } catch (error) {
    console.error('加载作业趋势数据失败:', error)
    ElMessage.warning('加载作业趋势数据失败')
  }
}

const refreshData = () => {
  loadDashboardData()
}

// 显示全部活动
const showAllActivities = () => {
  activitiesDialogVisible.value = true
  loadAllActivities()
}

// 加载全部活动
const loadAllActivities = async () => {
  try {
    activitiesLoading.value = true
    
    const params = {
      page: activityPagination.value.page,
      size: activityPagination.value.size
    }
    
    if (activityFilter.value.type) {
      params.type = activityFilter.value.type
    }
    
    if (activityFilter.value.dateRange && activityFilter.value.dateRange.length === 2) {
      params.start_date = activityFilter.value.dateRange[0]
      params.end_date = activityFilter.value.dateRange[1]
    }
    
    if (activityFilter.value.keyword) {
      params.keyword = activityFilter.value.keyword
    }
    
    const response = await api.get('/api/v1/dashboard/activities', { params })
    const data = response.data || { data: [], total: 0 }
    
    // 安全处理活动数据
    const items = data.data || []
    allActivities.value = items.map(activity => ({
      id: activity.id,
      message: activity.message || '未知活动',
      type: getTimelineItemType(activity.type),
      timestamp: activity.timestamp,
      user: activity.username || '',
      resource: activity.resource || ''
    }))
    
    activityPagination.value.total = data.total || 0
    
  } catch (error) {
    console.error('加载活动数据失败:', error)
    // 当没有数据时，不显示错误消息，而是显示空状态
    allActivities.value = []
    activityPagination.value.total = 0
    
    // 只有在真正的网络错误或服务器错误时才显示错误消息
    if (error.response && error.response.status >= 500) {
      ElMessage.error('服务器错误，请稍后重试')
    } else if (!error.response) {
      ElMessage.error('网络连接失败，请检查网络')
    }
  } finally {
    activitiesLoading.value = false
  }
}

// 关闭活动对话框
const handleCloseActivitiesDialog = () => {
  activitiesDialogVisible.value = false
  // 重置筛选条件
  activityFilter.value = {
    type: '',
    dateRange: null,
    keyword: ''
  }
  activityPagination.value.page = 1
}

// 防抖搜索
const debounceSearch = debounce(() => {
  activityPagination.value.page = 1
  loadAllActivities()
}, 500)

onMounted(() => {
  loadDashboardData()
})
</script>

<style scoped>
.dashboard {
  padding: 0;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.stat-card {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  display: flex;
  align-items: center;
  transition: transform 0.2s, box-shadow 0.2s;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  font-size: 24px;
  color: white;
}

.stat-icon.hosts {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.jobs {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.scripts {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}


.stat-content h3 {
  font-size: 28px;
  font-weight: 600;
  color: #2c3e50;
  margin: 0 0 4px 0;
}

.stat-content p {
  color: #7f8c8d;
  margin: 0 0 8px 0;
  font-size: 14px;
}

.stat-trend {
  font-size: 12px;
  color: #52c41a;
  font-weight: 500;
}

.stat-trend.critical {
  color: #ff4d4f;
}

.charts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.chart-card {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.chart-container {
  margin-top: 16px;
}

.activity-section {
  margin-bottom: 20px;
}

.card {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.card-header h3 {
  margin: 0;
  color: #2c3e50;
  font-size: 18px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.empty-state {
  text-align: center;
  padding: 40px 0;
}

/* 活动对话框样式 */
.activities-dialog-wrapper {
  .el-dialog__header {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 20px 24px;
    margin: 0;
    border-radius: 8px 8px 0 0;
  }
  
  .el-dialog__title {
    color: white;
    font-weight: 600;
    font-size: 18px;
  }
  
  .el-dialog__headerbtn .el-dialog__close {
    color: white;
    font-size: 18px;
  }
  
  .el-dialog__body {
    padding: 0;
  }
}

.activities-dialog {
  max-height: 70vh;
  display: flex;
  flex-direction: column;
}

.filter-bar {
  margin: 0;
  padding: 20px 24px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  border-bottom: 1px solid #e8eaec;
  
  .el-select,
  .el-date-editor,
  .el-input {
    width: 100%;
  }
}

.activities-list {
  flex: 1;
  max-height: 450px;
  overflow-y: auto;
  padding: 20px 24px;
  
  .el-timeline {
    padding-left: 0;
  }
  
  .el-timeline-item__wrapper {
    padding-left: 28px;
  }
  
  .el-timeline-item__tail {
    left: 4px;
  }
  
  .el-timeline-item__node {
    left: -4px;
  }
}

.activity-content {
  width: 100%;
  padding: 12px 16px;
  background: #fafbfc;
  border-radius: 8px;
  border-left: 4px solid #e8eaec;
  margin-bottom: 8px;
  transition: all 0.3s ease;
  
  &:hover {
    background: #f0f2f5;
    transform: translateX(2px);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }
}

.activity-message {
  font-size: 14px;
  color: #2c3e50;
  margin-bottom: 8px;
  line-height: 1.5;
  font-weight: 500;
}

.activity-meta {
  font-size: 12px;
  color: #7f8c8d;
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.activity-user,
.activity-resource {
  background: #e8f4fd;
  color: #1890ff;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 500;
  border: 1px solid #d4edda;
}

.activity-user {
  background: #f0f9ff;
  color: #0369a1;
  border-color: #bae6fd;
}

.activity-resource {
  background: #f0fdf4;
  color: #166534;
  border-color: #bbf7d0;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  
  .empty-icon {
    margin-bottom: 16px;
  }
  
  .empty-description {
    font-size: 16px;
    color: #606266;
    margin: 0 0 8px 0;
    font-weight: 500;
  }
  
  .empty-tip {
    font-size: 14px;
    color: #909399;
    margin: 0;
  }
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  padding: 16px 24px;
  background: #fafbfc;
  border-top: 1px solid #e8eaec;
  border-radius: 0 0 8px 8px;
  
  .el-pagination {
    .el-pagination__total,
    .el-pagination__jump {
      color: #606266;
    }
  }
}
</style>
