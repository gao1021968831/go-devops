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
          <el-button type="primary" size="small" @click="refreshData">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
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
      </div>
    </div>
  </div>
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
  try {
    // 加载仪表盘统计数据
    const statsResponse = await api.get('/api/v1/dashboard/stats')
    const dashboardStats = statsResponse.data
    stats.value.totalHosts = dashboardStats.total_hosts
    stats.value.onlineHosts = dashboardStats.online_hosts
    stats.value.totalJobs = dashboardStats.total_jobs
    stats.value.runningJobs = dashboardStats.running_jobs
    stats.value.totalScripts = dashboardStats.total_scripts

    // 加载主机状态分布
    const hostStatusResponse = await api.get('/api/v1/dashboard/host-status')
    const hostStatusData = hostStatusResponse.data
    hostStatusChart.value.series[0].data = hostStatusData.map(item => ({
      value: item.count,
      name: item.status === 'online' ? '在线' : item.status === 'offline' ? '离线' : '未知'
    }))

    // 加载最近活动数据
    const activitiesResponse = await api.get('/api/v1/dashboard/activities?limit=10')
    recentActivities.value = activitiesResponse.data.map(activity => ({
      id: activity.id,
      message: activity.message,
      type: getTimelineItemType(activity.type),
      timestamp: activity.timestamp
    }))

    // 加载作业趋势数据
    const trendResponse = await api.get('/api/v1/dashboard/job-trend?days=7')
    const trendData = trendResponse.data
    
    jobTrendChart.value.xAxis.data = trendData.map(item => item.date)
    jobTrendChart.value.series[0].data = trendData.map(item => item.success)
    jobTrendChart.value.series[1].data = trendData.map(item => item.failed)

  } catch (error) {
    console.error('加载仪表盘数据失败:', error)
    ElMessage.error('加载仪表盘数据失败')
  }
}

const refreshData = () => {
  loadDashboardData()
}

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
</style>
