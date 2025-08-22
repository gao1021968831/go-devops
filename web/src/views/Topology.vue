<template>
  <div class="topology-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">
          <el-icon><Share /></el-icon>
          主机拓扑
        </h2>
        <p class="page-subtitle">管理业务拓扑结构，组织主机资源</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="showCreateDialog('business')">
          <el-icon><Plus /></el-icon>
          创建业务
        </el-button>
        <el-button @click="refreshTopology" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <!-- 拓扑视图 -->
    <div class="topology-content">
      <div class="topology-sidebar">
        <!-- 拓扑树 -->
        <div class="topology-tree-container">
          <div class="tree-header">
            <h3>拓扑结构</h3>
            <div class="tree-actions">
              <el-button size="small" text @click="expandAll">
                <el-icon><FolderOpened /></el-icon>
                展开全部
              </el-button>
              <el-button size="small" text @click="collapseAll">
                <el-icon><Folder /></el-icon>
                收起全部
              </el-button>
            </div>
          </div>
          
          <el-tree
            ref="topologyTreeRef"
            :data="topologyTree"
            :props="treeProps"
            node-key="unique_id"
            :default-expand-all="false"
            :expand-on-click-node="false"
            @node-click="handleNodeClick"
            class="topology-tree"
          >
            <template #default="{ node, data }">
              <div class="tree-node" :class="`node-${data.type}`">
                <div class="node-icon">
                  <el-icon v-if="data.type === 'business'"><OfficeBuilding /></el-icon>
                  <el-icon v-else-if="data.type === 'environment'"><Collection /></el-icon>
                  <el-icon v-else-if="data.type === 'cluster'"><Grid /></el-icon>
                  <el-icon v-else-if="data.type === 'host'"><Monitor /></el-icon>
                </div>
                <div class="node-content">
                  <span class="node-name">{{ data.name }}</span>
                  <span v-if="data.stats" class="node-stats">
                    ({{ data.stats.online_hosts }}/{{ data.stats.total_hosts }})
                  </span>
                  <el-tag 
                    v-if="data.type === 'host' && data.host_info"
                    :type="getHostStatusType(data.host_info.status)"
                    size="small"
                    class="host-status-tag"
                  >
                    {{ getHostStatusText(data.host_info.status) }}
                  </el-tag>
                </div>
                <div class="node-actions">
                  <el-dropdown @command="handleNodeAction" trigger="click">
                    <el-button size="small" text>
                      <el-icon><More /></el-icon>
                    </el-button>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item 
                          v-if="data.type === 'business'" 
                          :command="{action: 'create', type: 'environment', parent: data}"
                        >
                          创建环境
                        </el-dropdown-item>
                        <el-dropdown-item 
                          v-if="data.type === 'environment'" 
                          :command="{action: 'create', type: 'cluster', parent: data}"
                        >
                          创建集群
                        </el-dropdown-item>
                        <el-dropdown-item 
                          v-if="data.type === 'cluster'" 
                          :command="{action: 'assign', type: 'host', parent: data}"
                        >
                          分配主机
                        </el-dropdown-item>
                        <el-dropdown-item 
                          :command="{action: 'edit', data: data}"
                        >
                          编辑
                        </el-dropdown-item>
                        <el-dropdown-item 
                          :command="{action: 'delete', data: data}"
                          divided
                        >
                          删除
                        </el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                </div>
              </div>
            </template>
          </el-tree>
        </div>
      </div>

      <!-- 详情面板 -->
      <div class="topology-detail">
        <div v-if="selectedNode" class="detail-container">
          <!-- 节点信息 -->
          <div class="detail-header">
            <div class="detail-title">
              <el-icon v-if="selectedNode.type === 'business'"><OfficeBuilding /></el-icon>
              <el-icon v-else-if="selectedNode.type === 'environment'"><Collection /></el-icon>
              <el-icon v-else-if="selectedNode.type === 'cluster'"><Grid /></el-icon>
              <el-icon v-else-if="selectedNode.type === 'host'"><Monitor /></el-icon>
              <span>{{ selectedNode.name }}</span>
              <el-tag :type="getNodeTypeColor(selectedNode.type)">
                {{ getNodeTypeText(selectedNode.type) }}
              </el-tag>
            </div>
          </div>

          <!-- 统计信息 -->
          <div v-if="selectedNode.stats" class="stats-grid">
            <div class="stat-card">
              <div class="stat-icon total">
                <el-icon><Monitor /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-number">{{ selectedNode.stats.total_hosts }}</div>
                <div class="stat-label">总主机数</div>
              </div>
            </div>
            <div class="stat-card">
              <div class="stat-icon online">
                <el-icon><CircleCheck /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-number">{{ selectedNode.stats.online_hosts }}</div>
                <div class="stat-label">在线主机</div>
              </div>
            </div>
            <div class="stat-card">
              <div class="stat-icon offline">
                <el-icon><CircleClose /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-number">{{ selectedNode.stats.offline_hosts }}</div>
                <div class="stat-label">离线主机</div>
              </div>
            </div>
          </div>

          <!-- 主机信息详情 -->
          <div v-if="selectedNode.type === 'host' && selectedNode.host_info" class="host-detail">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="主机名">{{ selectedNode.host_info.name }}</el-descriptions-item>
              <el-descriptions-item label="IP地址">{{ selectedNode.host_info.ip }}</el-descriptions-item>
              <el-descriptions-item label="端口">{{ selectedNode.host_info.port }}</el-descriptions-item>
              <el-descriptions-item label="操作系统">{{ selectedNode.host_info.os || '未知' }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="getHostStatusType(selectedNode.host_info.status)">
                  {{ getHostStatusText(selectedNode.host_info.status) }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="认证方式">{{ selectedNode.host_info.auth_type }}</el-descriptions-item>
              <el-descriptions-item label="描述" :span="2">
                {{ selectedNode.host_info.description || '无' }}
              </el-descriptions-item>
            </el-descriptions>
          </div>

          <!-- 子节点列表 -->
          <div v-if="selectedNode.children && selectedNode.children.length > 0" class="children-list">
            <h4>子节点 ({{ selectedNode.children.length }})</h4>
            <div class="children-grid">
              <div 
                v-for="child in selectedNode.children" 
                :key="child.id"
                class="child-card"
                @click="selectNode(child)"
              >
                <div class="child-icon">
                  <el-icon v-if="child.type === 'environment'"><Collection /></el-icon>
                  <el-icon v-else-if="child.type === 'cluster'"><Grid /></el-icon>
                  <el-icon v-else-if="child.type === 'host'"><Monitor /></el-icon>
                </div>
                <div class="child-info">
                  <div class="child-name">{{ child.name }}</div>
                  <div class="child-type">{{ getNodeTypeText(child.type) }}</div>
                </div>
                <div v-if="child.stats" class="child-stats">
                  {{ child.stats.online_hosts }}/{{ child.stats.total_hosts }}
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 空状态 -->
        <div v-else class="empty-detail">
          <el-empty description="选择左侧节点查看详情" />
        </div>
      </div>
    </div>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="showDialog"
      :title="dialogTitle"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入名称" />
        </el-form-item>
        <el-form-item v-if="dialogType === 'environment'" label="所属业务" prop="business_id">
          <el-select v-model="formData.business_id" placeholder="请选择业务" style="width: 100%">
            <el-option
              v-for="business in businesses"
              :key="business.id"
              :label="business.name"
              :value="business.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="dialogType === 'cluster'" label="所属环境" prop="environment_id">
          <el-select v-model="formData.environment_id" placeholder="请选择环境" style="width: 100%">
            <el-option
              v-for="environment in environments"
              :key="environment.id"
              :label="environment.name"
              :value="environment.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="dialogType === 'business'" label="负责人" prop="owner">
          <el-input v-model="formData.owner" placeholder="请输入负责人" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入描述"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 主机分配对话框 -->
    <el-dialog
      v-model="showAssignDialog"
      title="分配主机到集群"
      width="800px"
      :close-on-click-modal="false"
    >
      <div class="assign-container">
        <div class="unassigned-hosts">
          <h4>未分配主机</h4>
          <div class="host-list">
            <div
              v-for="host in unassignedHosts"
              :key="host.id"
              class="host-item"
              :class="{ selected: selectedHosts.includes(host.id) }"
              @click="toggleHostSelection(host.id)"
            >
              <div class="host-info">
                <el-icon><Monitor /></el-icon>
                <span class="host-name">{{ host.name }}</span>
                <span class="host-ip">{{ host.ip }}</span>
              </div>
              <el-tag :type="getHostStatusType(host.status)" size="small">
                {{ getHostStatusText(host.status) }}
              </el-tag>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="showAssignDialog = false">取消</el-button>
        <el-button 
          type="primary" 
          @click="handleAssignHosts" 
          :loading="assigning"
          :disabled="selectedHosts.length === 0"
        >
          分配选中主机 ({{ selectedHosts.length }})
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Share, Plus, Refresh, FolderOpened, Folder, More, OfficeBuilding,
  Collection, Grid, Monitor, CircleCheck, CircleClose
} from '@element-plus/icons-vue'
import api from '@/utils/api'

// 响应式数据
const loading = ref(false)
const topologyTree = ref([])
const selectedNode = ref(null)
const businesses = ref([])
const environments = ref([])
const unassignedHosts = ref([])
const selectedHosts = ref([])

// 对话框状态
const showDialog = ref(false)
const showAssignDialog = ref(false)
const dialogType = ref('')
const dialogMode = ref('create')
const submitting = ref(false)
const assigning = ref(false)
const currentParent = ref(null)
const currentCluster = ref(null)

// 表单数据
const formData = reactive({
  name: '',
  code: '',
  description: '',
  owner: '',
  business_id: null,
  environment_id: null
})

// 树组件引用
const topologyTreeRef = ref(null)
const formRef = ref(null)

// 树配置
const treeProps = {
  children: 'children',
  label: 'name'
}

// 表单验证规则
const formRules = computed(() => {
  const baseRules = {
    name: [{ required: true, message: '请输入名称', trigger: 'blur' }]
  }
  
  // 所有类型的编码都由后端自动生成，不需要验证
  
  if (dialogType.value === 'environment') {
    baseRules.business_id = [{ required: true, message: '请选择业务', trigger: 'change' }]
  }
  
  if (dialogType.value === 'cluster') {
    baseRules.environment_id = [{ required: true, message: '请选择环境', trigger: 'change' }]
  }
  
  return baseRules
})

// 计算属性
const dialogTitle = computed(() => {
  const typeMap = {
    business: '业务',
    environment: '环境',
    cluster: '集群'
  }
  const modeMap = {
    create: '创建',
    edit: '编辑'
  }
  return `${modeMap[dialogMode.value]}${typeMap[dialogType.value]}`
})

// 方法
const refreshTopology = async () => {
  loading.value = true
  try {
    const response = await api.get('/api/v1/topology/tree')
    topologyTree.value = response.data.data
  } catch (error) {
    ElMessage.error('获取拓扑数据失败')
  } finally {
    loading.value = false
  }
}

const loadBusinesses = async () => {
  try {
    const response = await api.get('/api/v1/topology/businesses')
    businesses.value = response.data.data
  } catch (error) {
    ElMessage.error('获取业务列表失败')
  }
}

const loadEnvironments = async (businessId = null) => {
  try {
    const params = businessId ? { business_id: businessId } : {}
    const response = await api.get('/api/v1/topology/environments', { params })
    environments.value = response.data.data
  } catch (error) {
    ElMessage.error('获取环境列表失败')
  }
}

const loadUnassignedHosts = async () => {
  try {
    const response = await api.get('/api/v1/topology/hosts/unassigned')
    unassignedHosts.value = response.data.data
  } catch (error) {
    ElMessage.error('获取未分配主机失败')
  }
}

const expandAll = () => {
  const tree = topologyTreeRef.value
  if (!tree || !topologyTree.value.length) return
  
  try {
    // 递归获取所有节点的key（包括所有节点，不仅仅是有子节点的）
    const getAllNodeKeys = (nodes) => {
      let keys = []
      nodes.forEach(node => {
        keys.push(node.unique_id) // 使用unique_id作为节点标识
        if (node.children && node.children.length > 0) {
          keys = keys.concat(getAllNodeKeys(node.children)) // 递归添加子节点
        }
      })
      return keys
    }
    
    const allKeys = getAllNodeKeys(topologyTree.value)
    console.log('展开节点keys:', allKeys)
    
    // 使用nextTick确保DOM更新后再操作
    nextTick(() => {
      allKeys.forEach(key => {
        const node = tree.getNode(key)
        if (node && node.childNodes && node.childNodes.length > 0 && !node.expanded) {
          node.expand()
        }
      })
    })
  } catch (error) {
    console.error('展开全部失败:', error)
  }
}

const handleNodeClick = (data) => {
  selectedNode.value = data
}

const collapseAll = () => {
  const tree = topologyTreeRef.value
  if (!tree || !topologyTree.value.length) return
  
  try {
    // 递归获取所有节点的key（使用unique_id作为节点标识）
    const getAllNodeKeys = (nodes) => {
      let keys = []
      nodes.forEach(node => {
        keys.push(node.unique_id) // 使用unique_id作为节点标识
        if (node.children && node.children.length > 0) {
          keys = keys.concat(getAllNodeKeys(node.children)) // 递归添加子节点
        }
      })
      return keys
    }
    
    const allKeys = getAllNodeKeys(topologyTree.value)
    console.log('收起节点keys:', allKeys)
    
    // 使用nextTick确保DOM更新后再操作
    nextTick(() => {
      allKeys.forEach(key => {
        const node = tree.getNode(key)
        if (node && node.childNodes && node.childNodes.length > 0 && node.expanded) {
          node.collapse()
        }
      })
    })
  } catch (error) {
    console.error('收起全部失败:', error)
  }
}

const selectNode = (node) => {
  selectedNode.value = node
}

const handleNodeAction = async (command) => {
  const { action, type, parent, data } = command

  switch (action) {
    case 'create':
      showCreateDialog(type, parent)
      break
    case 'edit':
      showEditDialog(data)
      break
    case 'delete':
      await handleDelete(data)
      break
    case 'assign':
      showAssignHostDialog(parent)
      break
  }
}

const showCreateDialog = (type, parent = null) => {
  dialogType.value = type
  dialogMode.value = 'create'
  currentParent.value = parent
  resetForm()
  
  if (type === 'environment' && parent) {
    formData.business_id = parent.id
  } else if (type === 'cluster' && parent) {
    formData.environment_id = parent.id
  }
  
  showDialog.value = true
}

const showEditDialog = (data) => {
  dialogType.value = data.type
  dialogMode.value = 'edit'
  
  formData.name = data.name
  formData.code = data.code
  formData.description = data.description || ''
  formData.owner = data.owner || ''
  
  showDialog.value = true
}

const showAssignHostDialog = (cluster) => {
  currentCluster.value = cluster
  selectedHosts.value = []
  loadUnassignedHosts()
  showAssignDialog.value = true
}

const resetForm = () => {
  Object.keys(formData).forEach(key => {
    formData[key] = key.includes('_id') ? null : ''
  })
}

const handleSubmit = async () => {
  const form = formRef.value
  if (!form) return

  try {
    await form.validate()
    submitting.value = true

    let url = `/api/v1/topology/${dialogType.value}s`
    if (dialogType.value === 'business') {
      url = '/api/v1/topology/businesses'
    }
    
    // 根据类型过滤提交数据
    let data = {
      name: formData.name,
      description: formData.description
    }
    
    if (dialogType.value === 'business') {
      data.owner = formData.owner
    } else if (dialogType.value === 'environment') {
      data.business_id = formData.business_id
    } else if (dialogType.value === 'cluster') {
      data.environment_id = formData.environment_id
    }
    // 所有类型的编码都由后端自动生成

    let response
    if (dialogMode.value === 'edit') {
      url += `/${selectedNode.value.id}`
      response = await api.put(url, data)
    } else {
      response = await api.post(url, data)
    }
    
    ElMessage.success(`${dialogTitle.value}成功`)
    showDialog.value = false
    await refreshTopology()
  } catch (error) {
    console.error('提交失败:', error)
    ElMessage.error(`${dialogTitle.value}失败: ${error.response?.data?.error || error.message}`)
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (data) => {
  try {
    const actionText = data.type === 'host' ? '从集群中移除' : '删除'
    const confirmTitle = data.type === 'host' ? '移除确认' : '删除确认'
    await ElMessageBox.confirm(
      `确定要${actionText}${getNodeTypeText(data.type)} "${data.name}" 吗？`,
      confirmTitle,
      { type: 'warning' }
    )

    let deleteUrl = `/api/v1/topology/${data.type}s/${data.id}`
    if (data.type === 'business') {
      deleteUrl = `/api/v1/topology/businesses/${data.id}`
    } else if (data.type === 'host') {
      // 主机是从集群中移除，不是删除主机本身
      deleteUrl = `/api/v1/topology/hosts/${data.id}/remove`
    }
    await api.delete(deleteUrl)
    const successText = data.type === 'host' ? '移除成功' : '删除成功'
    ElMessage.success(successText)
    await refreshTopology()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error(`删除失败: ${error.response?.data?.error || error.message}`)
    }
  }
}

const toggleHostSelection = (hostId) => {
  const index = selectedHosts.value.indexOf(hostId)
  if (index > -1) {
    selectedHosts.value.splice(index, 1)
  } else {
    selectedHosts.value.push(hostId)
  }
}

const handleAssignHosts = async () => {
  if (selectedHosts.value.length === 0) return

  assigning.value = true
  try {
    for (const hostId of selectedHosts.value) {
      await api.post('/api/v1/topology/hosts/assign', {
        host_id: hostId,
        cluster_id: currentCluster.value.id
      })
    }
    
    ElMessage.success(`成功分配 ${selectedHosts.value.length} 台主机`)
    showAssignDialog.value = false
    await refreshTopology()
  } catch (error) {
    ElMessage.error('分配主机失败')
  } finally {
    assigning.value = false
  }
}

// 工具方法
const getNodeTypeText = (type) => {
  const typeMap = {
    business: '业务',
    environment: '环境',
    cluster: '集群',
    host: '主机'
  }
  return typeMap[type] || type
}

const getNodeTypeColor = (type) => {
  const colorMap = {
    business: 'primary',
    environment: 'success',
    cluster: 'warning',
    host: 'info'
  }
  return colorMap[type] || 'info'
}

const getHostStatusType = (status) => {
  const statusMap = {
    online: 'success',
    offline: 'danger',
    unknown: 'info'
  }
  return statusMap[status] || 'info'
}

const getHostStatusText = (status) => {
  const statusMap = {
    online: '在线',
    offline: '离线',
    unknown: '未知'
  }
  return statusMap[status] || status
}

// 生命周期
onMounted(() => {
  refreshTopology()
  loadBusinesses()
  loadEnvironments()
})
</script>

<style scoped>
.topology-container {
  padding: 24px;
  background: #f5f7fa;
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  background: white;
  padding: 24px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #1f2937;
}

.page-subtitle {
  margin: 0;
  color: #6b7280;
  font-size: 14px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.topology-content {
  display: flex;
  gap: 24px;
  height: calc(100vh - 200px);
}

.topology-sidebar {
  width: 400px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.topology-tree-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.tree-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e5e7eb;
}

.tree-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.tree-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.topology-tree {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}

.tree-node {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  border-radius: 6px;
  transition: background-color 0.2s;
}

.tree-node:hover {
  background: #f3f4f6;
}

.node-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 4px;
}

.node-business .node-icon {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.node-environment .node-icon {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: white;
}

.node-cluster .node-icon {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  color: white;
}

.node-host .node-icon {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
  color: white;
}

.node-content {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
}

.node-name {
  font-weight: 500;
  color: #1f2937;
}

.node-stats {
  font-size: 12px;
  color: #6b7280;
}

.host-status-tag {
  margin-left: auto;
}

.node-actions {
  opacity: 0;
  transition: opacity 0.2s;
}

.tree-node:hover .node-actions {
  opacity: 1;
}

.topology-detail {
  flex: 1;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.detail-container {
  height: 100%;
  overflow-y: auto;
  padding: 24px;
}

.detail-header {
  margin-bottom: 24px;
}

.detail-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 20px;
  font-weight: 600;
  color: #1f2937;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: #f8fafc;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.stat-icon.total {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.online {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
}

.stat-icon.offline {
  background: linear-gradient(135deg, #fc466b 0%, #3f5efb 100%);
}

.stat-info {
  flex: 1;
}

.stat-number {
  font-size: 24px;
  font-weight: 700;
  color: #1f2937;
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: #6b7280;
  margin-top: 4px;
}

.host-detail {
  margin-bottom: 24px;
}

.children-list h4 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.children-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
}

.child-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.child-card:hover {
  border-color: #3b82f6;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
}

.child-icon {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
}

.child-info {
  flex: 1;
}

.child-name {
  font-weight: 500;
  color: #1f2937;
  margin-bottom: 2px;
}

.child-type {
  font-size: 12px;
  color: #6b7280;
}

.child-stats {
  font-size: 12px;
  color: #059669;
  font-weight: 500;
}

.empty-detail {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.assign-container {
  max-height: 400px;
  overflow-y: auto;
}

.unassigned-hosts h4 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.host-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.host-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.host-item:hover {
  border-color: #3b82f6;
  background: #f8fafc;
}

.host-item.selected {
  border-color: #3b82f6;
  background: #eff6ff;
}

.host-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.host-name {
  font-weight: 500;
  color: #1f2937;
}

.host-ip {
  color: #6b7280;
  font-size: 14px;
}
</style>
