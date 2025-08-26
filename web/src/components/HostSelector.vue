<template>
  <div class="host-selector">
    <!-- 选择模式切换 -->
    <div class="selection-mode-tabs">
      <el-radio-group v-model="selectionMode" class="mode-tabs">
        <el-radio-button label="list">
          <el-icon><List /></el-icon>
          列表选择
        </el-radio-button>
        <el-radio-button label="topology">
          <el-icon><Share /></el-icon>
          拓扑选择
        </el-radio-button>
      </el-radio-group>
    </div>

    <!-- 快速操作工具栏 -->
    <div class="host-quick-actions">
      <div class="quick-action-title">
        <el-icon><Setting /></el-icon>
        <span>快速选择</span>
      </div>
      <div class="quick-buttons">
        <el-button size="small" @click="selectAll">
          <el-icon><Check /></el-icon>
          全选
        </el-button>
        <el-button size="small" @click="clearSelection">
          <el-icon><Close /></el-icon>
          清空
        </el-button>
        <el-button size="small" @click="selectOnline">
          <el-icon><Connection /></el-icon>
          仅在线
        </el-button>
      </div>
    </div>

    <!-- IP 粘贴添加 -->
    <div class="ip-paste-section">
      <el-collapse>
        <el-collapse-item title="批量添加IP地址" name="ip-paste">
          <el-input
            v-model="ipPasteText"
            type="textarea"
            :rows="3"
            placeholder="可以粘贴IP地址列表，支持多种格式：
192.168.1.1
192.168.1.2,192.168.1.3
192.168.1.4;192.168.1.5"
            class="ip-paste-input"
          />
          <div class="ip-paste-actions">
            <el-button size="small" type="primary" @click="parseAndSelectIPs" :disabled="!ipPasteText.trim()">
              <el-icon><Plus /></el-icon>
              解析并添加
            </el-button>
            <el-button size="small" @click="clearIPPaste">
              <el-icon><Delete /></el-icon>
              清空
            </el-button>
          </div>
        </el-collapse-item>
      </el-collapse>
    </div>

    <!-- 列表模式 -->
    <div v-if="selectionMode === 'list'" class="host-selection-container">
      <div class="host-search">
        <el-input
          v-model="searchText"
          placeholder="搜索主机名称或IP..."
          clearable
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
      
      <div class="host-list">
        <div 
          v-for="host in filteredHosts" 
          :key="host.id"
          class="host-item"
          :class="{ 
            'selected': selectedHostIds.includes(host.id),
            'offline': host.status === 'offline'
          }"
          @click="toggleHostSelection(host.id)"
        >
          <div class="host-checkbox">
            <el-checkbox :model-value="selectedHostIds.includes(host.id)" />
          </div>
          <div class="host-info">
            <div class="host-main">
              <span class="host-name">{{ host.name }}</span>
              <span class="host-ip">{{ host.ip }}</span>
            </div>
            <div class="host-meta">
              <el-tag 
                :type="host.status === 'online' ? 'success' : 'danger'" 
                size="small"
              >
                {{ host.status === 'online' ? '在线' : '离线' }}
              </el-tag>
              <span class="host-os">{{ host.os || '未知系统' }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 拓扑模式 -->
    <div v-else-if="selectionMode === 'topology'" class="topology-selection">
      <div v-if="topologyLoading" class="topology-loading">
        <el-skeleton :rows="5" animated />
      </div>
      <div v-else-if="topologyTree.length === 0" class="empty-topology">
        <el-empty description="暂无拓扑数据">
          <el-button type="primary" @click="loadTopologyTree">
            <el-icon><Refresh /></el-icon>
            重新加载
          </el-button>
        </el-empty>
      </div>
      <div v-else class="topology-tree-container">
        <el-tree
          ref="topologyTreeRef"
          :data="topologyTree"
          :props="topologyTreeProps"
          node-key="unique_id"
          show-checkbox
          :check-strictly="false"
          :default-expand-all="true"
          @check="handleTopologyCheck"
          class="topology-tree"
          :indent="20"
        >
          <template #default="{ data }">
            <div class="custom-tree-node" :class="`node-${data.type}`">
              <div class="node-left">
                <div class="node-icon">
                  <el-icon v-if="data.type === 'business'"><OfficeBuilding /></el-icon>
                  <el-icon v-else-if="data.type === 'environment'"><Collection /></el-icon>
                  <el-icon v-else-if="data.type === 'cluster'"><Grid /></el-icon>
                  <el-icon v-else-if="data.type === 'host'"><Monitor /></el-icon>
                </div>
                <div class="node-content">
                  <div class="node-title">
                    <span class="node-name">{{ data.name }}</span>
                    <el-tag v-if="data.type === 'host' && data.host_info" 
                            :type="data.host_info.status === 'online' ? 'success' : 'danger'" 
                            size="small">
                      {{ data.host_info.status === 'online' ? '在线' : '离线' }}
                    </el-tag>
                  </div>
                  <div v-if="data.type === 'host' && data.host_info" class="node-subtitle">
                    {{ data.host_info.ip }} | {{ data.host_info.os || '未知系统' }}
                  </div>
                </div>
              </div>
            </div>
          </template>
        </el-tree>
      </div>
    </div>

    <!-- 选择摘要 -->
    <div class="host-selection-summary" v-if="selectedHostIds.length > 0">
      <el-alert
        :title="`已选择 ${selectedHostIds.length} 台主机`"
        type="info"
        show-icon
        :closable="false"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'
import {
  List,
  Share,
  Setting,
  Check,
  Close,
  Connection,
  Plus,
  Delete,
  Search,
  Refresh,
  OfficeBuilding,
  Collection,
  Grid,
  Monitor
} from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:modelValue'])

// 响应式数据
const selectionMode = ref('list')
const searchText = ref('')
const ipPasteText = ref('')
const hosts = ref([])
const topologyTree = ref([])
const topologyTreeRef = ref()
const topologyLoading = ref(false)

const topologyTreeProps = {
  label: 'name',
  children: 'children'
}

// 计算属性
const selectedHostIds = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const filteredHosts = computed(() => {
  if (!searchText.value) return hosts.value
  
  const search = searchText.value.toLowerCase()
  return hosts.value.filter(host => 
    host.name.toLowerCase().includes(search) ||
    host.ip.toLowerCase().includes(search)
  )
})

// 方法
const loadHosts = async () => {
  try {
    const response = await api.get('/api/v1/hosts')
    hosts.value = response.data.data || []
  } catch (error) {
    console.error('加载主机列表失败:', error)
    // 根据HTTP状态码显示具体错误信息
    if (error.response) {
      const { status, data } = error.response
      switch (status) {
        case 401:
          ElMessage.error('未授权，请重新登录')
          break
        case 403:
          ElMessage.error('没有权限查看主机列表')
          break
        case 500:
          ElMessage.error('服务器内部错误，加载主机列表失败')
          break
        default:
          ElMessage.error(data?.error || `加载主机列表失败 (状态码: ${status})`)
      }
    } else {
      ElMessage.error('网络连接失败，无法加载主机列表')
    }
  }
}

const loadTopologyTree = async () => {
  topologyLoading.value = true
  try {
    const response = await api.get('/api/v1/topology/tree')
    // 后端返回格式是 {data: topologyTree}，需要访问 response.data.data
    const treeData = response.data?.data || response.data || []
    topologyTree.value = Array.isArray(treeData) ? treeData : []
  } catch (error) {
    console.error('加载拓扑树失败:', error)
    // 根据HTTP状态码显示具体错误信息
    if (error.response) {
      const { status, data } = error.response
      switch (status) {
        case 401:
          ElMessage.error('未授权，请重新登录')
          break
        case 403:
          ElMessage.error('没有权限查看拓扑树')
          break
        case 500:
          ElMessage.error('服务器内部错误，加载拓扑树失败')
          break
        default:
          ElMessage.error(data?.error || `加载拓扑树失败 (状态码: ${status})`)
      }
    } else {
      ElMessage.error('网络连接失败，无法加载拓扑树')
    }
    topologyTree.value = []
  } finally {
    topologyLoading.value = false
  }
}

const toggleHostSelection = (hostId) => {
  const index = selectedHostIds.value.indexOf(hostId)
  if (index > -1) {
    selectedHostIds.value = selectedHostIds.value.filter(id => id !== hostId)
  } else {
    selectedHostIds.value = [...selectedHostIds.value, hostId]
  }
}

const selectAll = () => {
  selectedHostIds.value = hosts.value.map(host => host.id)
  ElMessage.success(`已选择 ${hosts.value.length} 台主机`)
}

const clearSelection = () => {
  selectedHostIds.value = []
  if (topologyTreeRef.value) {
    topologyTreeRef.value.setCheckedKeys([])
  }
  ElMessage.success('已清空选择')
}

const selectOnline = () => {
  const onlineHosts = hosts.value.filter(host => host.status === 'online')
  selectedHostIds.value = onlineHosts.map(host => host.id)
  ElMessage.success(`已选择 ${onlineHosts.length} 台在线主机`)
}

const parseAndSelectIPs = () => {
  if (!ipPasteText.value.trim()) return
  
  // 解析IP地址（支持换行、逗号、分号分隔）
  const ips = ipPasteText.value
    .split(/[\n,;]/)
    .map(ip => ip.trim())
    .filter(ip => ip && /^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$/.test(ip))
  
  if (ips.length === 0) {
    ElMessage.warning('未找到有效的IP地址')
    return
  }
  
  // 根据IP匹配主机
  const matchedHosts = hosts.value.filter(host => ips.includes(host.ip))
  const newSelectedIds = [...new Set([...selectedHostIds.value, ...matchedHosts.map(h => h.id)])]
  
  selectedHostIds.value = newSelectedIds
  ElMessage.success(`成功添加 ${matchedHosts.length} 台主机`)
  
  // 清空输入框
  ipPasteText.value = ''
}

const clearIPPaste = () => {
  ipPasteText.value = ''
}

const handleTopologyCheck = (data, { checkedKeys }) => {
  // 从拓扑树中提取主机ID
  const hostIds = []
  
  const extractHostIds = (nodes) => {
    nodes.forEach(node => {
      if (node.type === 'host' && node.host_info) {
        hostIds.push(node.host_info.id)
      }
      if (node.children) {
        extractHostIds(node.children)
      }
    })
  }
  
  // 获取所有选中的节点数据
  const checkedNodes = topologyTreeRef.value.getCheckedNodes()
  extractHostIds(checkedNodes)
  
  selectedHostIds.value = [...new Set(hostIds)]
}

// 生命周期
onMounted(() => {
  loadHosts()
  loadTopologyTree()
})

// 监听器
watch(() => selectionMode.value, (newMode) => {
  if (newMode === 'topology' && topologyTree.value.length === 0) {
    loadTopologyTree()
  }
})
</script>

<style scoped>
.host-selector {
  width: 100%;
}

.selection-mode-tabs {
  margin-bottom: 16px;
}

.mode-tabs {
  width: 100%;
}

.host-quick-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
  margin-bottom: 16px;
}

.quick-action-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #606266;
}

.quick-buttons {
  display: flex;
  gap: 8px;
}

.ip-paste-section {
  margin-bottom: 16px;
}

.ip-paste-input {
  margin-bottom: 12px;
}

.ip-paste-actions {
  display: flex;
  gap: 8px;
}

.host-search {
  margin-bottom: 16px;
}

.host-selection-container {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  max-height: 400px;
  overflow: hidden;
}

.host-list {
  max-height: 400px;
  overflow-y: auto;
}

.host-item {
  display: flex;
  align-items: center;
  padding: 12px;
  border-bottom: 1px solid #f0f2f5;
  cursor: pointer;
  transition: all 0.2s;
}

.host-item:hover {
  background: #f5f7fa;
}

.host-item.selected {
  background: #e8f4fd;
  border-color: #409eff;
}

.host-item.offline {
  opacity: 0.6;
}

.host-checkbox {
  margin-right: 12px;
}

.host-info {
  flex: 1;
}

.host-main {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.host-name {
  font-weight: 600;
  color: #303133;
}

.host-ip {
  color: #909399;
  font-size: 14px;
}

.host-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.host-os {
  color: #909399;
  font-size: 12px;
}

.topology-selection {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  min-height: 300px;
}

.topology-loading,
.empty-topology {
  padding: 40px;
  text-align: center;
}

.topology-tree-container {
  padding: 16px;
  max-height: 400px;
  overflow-y: auto;
}

.topology-tree :deep(.el-tree-node__content) {
  height: auto !important;
  min-height: 32px;
  padding: 4px 0;
}

.topology-tree :deep(.el-tree-node__label) {
  width: 100%;
  overflow: visible;
}

.custom-tree-node {
  display: flex;
  align-items: flex-start;
  width: 100%;
  min-height: 32px;
  padding: 4px 0;
}

.node-left {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  width: 100%;
}

.node-icon {
  color: #909399;
  margin-top: 2px;
  flex-shrink: 0;
}

.node-content {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.node-title {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 2px;
}

.node-name {
  font-weight: 500;
  color: #303133;
  word-break: break-word;
  flex: 1;
  min-width: 0;
}

.node-subtitle {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
  word-break: break-all;
  line-height: 1.4;
}

.node-business .node-icon {
  color: #e6a23c;
}

.node-environment .node-icon {
  color: #67c23a;
}

.node-cluster .node-icon {
  color: #409eff;
}

.node-host .node-icon {
  color: #909399;
}

.host-selection-summary {
  margin-top: 16px;
}
</style>
