<template>
  <div class="hosts-page">
    <div class="page-header">
      <div class="header-left">
        <div class="page-title">
          <el-icon class="title-icon"><Monitor /></el-icon>
          <h2>主机管理</h2>
        </div>
        <p class="page-subtitle">管理和监控服务器主机资源</p>
      </div>
      <div class="header-actions">
        <el-button type="warning" @click="checkAllHostsStatus" :loading="checkingStatus">
          <el-icon><Refresh /></el-icon>
          批量检查
        </el-button>
        <el-dropdown @command="handleBatchAction">
          <el-button type="success">
            <el-icon><Upload /></el-icon>
            批量导入
            <el-icon class="el-icon--right"><arrow-down /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="csv-upload">CSV文件导入</el-dropdown-item>
              <el-dropdown-item command="download-template">下载CSV模板</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button type="primary" @click="showAddDialog = true" class="create-btn">
          <el-icon><Plus /></el-icon>
          添加主机
        </el-button>
      </div>
    </div>

    <!-- 搜索和过滤 -->
    <div class="search-section">
      <div class="search-left">
        <el-input
          v-model="searchText"
          placeholder="搜索主机名、IP地址或描述..."
          class="search-input"
          clearable
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-select v-model="statusFilter" placeholder="状态筛选" class="status-filter">
          <el-option label="全部状态" value="" />
          <el-option label="在线" value="online" />
          <el-option label="离线" value="offline" />
          <el-option label="未知" value="unknown" />
        </el-select>
        <el-select v-model="osFilter" placeholder="系统筛选" class="os-filter">
          <el-option label="全部系统" value="" />
          <el-option label="Linux" value="Linux" />
          <el-option label="Windows" value="Windows" />
          <el-option label="macOS" value="macOS" />
          <el-option label="其他" value="Other" />
        </el-select>
      </div>
      <div class="search-right">
        <el-button @click="loadHosts" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <!-- 主机列表 -->
    <div class="hosts-container">
      <div v-if="filteredHosts.length === 0" class="empty-state">
        <el-empty description="暂无主机数据">
          <el-button type="primary" @click="showAddDialog = true">
            <el-icon><Plus /></el-icon>
            添加第一台主机
          </el-button>
        </el-empty>
      </div>
      
      <el-table v-else :data="filteredHosts" class="hosts-table" v-loading="loading">
        <el-table-column type="selection" width="55" />
        
        <el-table-column label="主机信息" min-width="60">
          <template #default="{ row }">
            <div class="host-info-cell">
              <div class="host-avatar">
                <el-icon><Monitor /></el-icon>
              </div>
              <div class="host-details">
                <div class="host-name clickable" @click="viewHost(row)">{{ row.name }}</div>
                <div class="host-ip">{{ row.ip }}:{{ row.port }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" effect="light">
              <el-icon v-if="row.status === 'online'"><CircleCheck /></el-icon>
              <el-icon v-else-if="row.status === 'offline'"><CircleClose /></el-icon>
              <el-icon v-else><Warning /></el-icon>
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="操作系统" width="300" align="center">
          <template #default="{ row }">
            <div class="os-cell">
              <el-icon v-if="row.os === 'Linux'"><Platform /></el-icon>
              <el-icon v-else-if="row.os === 'Windows'"><Monitor /></el-icon>
              <el-icon v-else-if="row.os === 'macOS'"><Iphone /></el-icon>
              <el-icon v-else><QuestionFilled /></el-icon>
              <span>{{ row.os || '未知' }}</span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="描述" min-width="50" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="description-text">{{ row.description || '无描述' }}</span>
          </template>
        </el-table-column>
        
        <el-table-column label="标签" width="150">
          <template #default="{ row }">
            <div class="tags-cell" v-if="row.tags">
              <el-tag
                v-for="tag in row.tags.split(',').slice(0, 2)"
                :key="tag"
                size="small"
                effect="plain"
                class="tag-item"
              >
                {{ tag.trim() }}
              </el-tag>
              <el-tag v-if="row.tags.split(',').length > 2" size="small" type="info" effect="plain">
                +{{ row.tags.split(',').length - 2 }}
              </el-tag>
            </div>
            <span v-else class="no-tags">无标签</span>
          </template>
        </el-table-column>
        
        <el-table-column label="创建时间" width="180" align="center">
          <template #default="{ row }">
            <div class="time-cell">
              <el-icon><Calendar /></el-icon>
              <span>{{ formatDate(row.created_at) }}</span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="300" align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" @click="checkHostStatus(row)" type="warning" link>
                <el-icon><Refresh /></el-icon>
                检查
              </el-button>
              <el-button size="small" @click="editHost(row)" type="primary" link>
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button size="small" @click="deleteHost(row)" type="danger" link>
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 添加/编辑主机对话框 -->
    <el-dialog
      v-model="showAddDialog"
      :title="editingHost ? '编辑主机' : '添加主机'"
      width="700px"
      class="host-form-dialog"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div class="dialog-header">
        <div class="dialog-icon">
          <el-icon><Monitor /></el-icon>
        </div>
        <div class="dialog-title-info">
          <h3>{{ editingHost ? '编辑主机信息' : '添加新主机' }}</h3>
          <p>{{ editingHost ? '修改主机配置和认证信息' : '配置主机基本信息和SSH认证' }}</p>
        </div>
      </div>

      <el-form
        ref="hostFormRef"
        :model="hostForm"
        :rules="hostRules"
        label-width="120px"
        class="host-form"
      >
        <!-- 基本信息 -->
        <div class="form-section">
          <div class="section-title">
            <el-icon><Monitor /></el-icon>
            <span>基本信息</span>
            <div class="section-divider"></div>
          </div>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="主机名称" prop="name">
                <el-input 
                  v-model="hostForm.name" 
                  placeholder="如：web-server-01"
                  clearable
                >
                  <template #prefix>
                    <el-icon><Monitor /></el-icon>
                  </template>
                </el-input>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="IP地址" prop="ip">
                <el-input 
                  v-model="hostForm.ip" 
                  placeholder="如：192.168.1.100"
                  clearable
                >
                  <template #prefix>
                    <el-icon><Location /></el-icon>
                  </template>
                </el-input>
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="SSH端口" prop="port">
                <el-input-number 
                  v-model="hostForm.port" 
                  :min="1" 
                  :max="65535" 
                  placeholder="默认22"
                  style="width: 100%"
                  controls-position="right"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="操作系统">
                <el-select 
                  v-model="hostForm.os" 
                  placeholder="选择操作系统" 
                  style="width: 100%"
                  clearable
                >
                  <el-option label="Linux" value="Linux">
                    <div class="option-item">
                      <el-icon><Platform /></el-icon>
                      <span>Linux</span>
                    </div>
                  </el-option>
                  <el-option label="Windows" value="Windows">
                    <div class="option-item">
                      <el-icon><Monitor /></el-icon>
                      <span>Windows</span>
                    </div>
                  </el-option>
                  <el-option label="macOS" value="macOS">
                    <div class="option-item">
                      <el-icon><Iphone /></el-icon>
                      <span>macOS</span>
                    </div>
                  </el-option>
                  <el-option label="其他" value="Other">
                    <div class="option-item">
                      <el-icon><QuestionFilled /></el-icon>
                      <span>其他</span>
                    </div>
                  </el-option>
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>
        </div>

        <!-- SSH认证配置 -->
        <div class="form-section">
          <div class="section-title">
            <el-icon><Key /></el-icon>
            <span>SSH认证配置</span>
            <div class="section-divider"></div>
          </div>
          <el-alert
            title="认证提示"
            type="info"
            :closable="false"
            show-icon
            class="auth-alert"
          >
            <template #default>
              <div class="alert-content">
                <p>• SSH认证信息用于远程连接和管理主机</p>
                <p>• 请确保用户名和密码正确，避免连接失败</p>
                <p v-if="editingHost">• 密码留空将保持原有密码不变</p>
              </div>
            </template>
          </el-alert>
          
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="SSH用户名" prop="username">
                <el-input 
                  v-model="hostForm.username" 
                  placeholder="如：root, admin, ubuntu"
                  clearable
                >
                  <template #prefix>
                    <el-icon><User /></el-icon>
                  </template>
                </el-input>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="SSH密码" prop="password">
                <el-input 
                  v-model="hostForm.password" 
                  type="password" 
                  :placeholder="editingHost ? '留空保持不变' : '请输入SSH密码'"
                  show-password
                  clearable
                >
                  <template #prefix>
                    <el-icon><Lock /></el-icon>
                  </template>
                </el-input>
              </el-form-item>
            </el-col>
          </el-row>
        </div>
        
        <!-- 附加信息 -->
        <div class="form-section">
          <div class="section-title">
            <el-icon><Document /></el-icon>
            <span>附加信息</span>
            <div class="section-divider"></div>
          </div>
          <el-form-item label="主机描述">
            <el-input
              v-model="hostForm.description"
              type="textarea"
              :rows="3"
              placeholder="请输入主机描述信息，如：生产环境Web服务器"
              show-word-limit
              maxlength="200"
              resize="none"
            />
          </el-form-item>
          <el-form-item label="主机标签" prop="tags">
            <el-input
              v-model="hostForm.tags"
              placeholder="多个标签用逗号分隔，如：web,production,nginx"
              clearable
              maxlength="100"
              show-word-limit
            >
              <template #prefix>
                <el-icon><Discount /></el-icon>
              </template>
            </el-input>
            <template #extra>
              <div class="form-tip">
                <el-icon><InfoFilled /></el-icon>
                <span>标签用于分类和筛选主机，建议使用简短的关键词，多个标签用逗号分隔</span>
              </div>
            </template>
          </el-form-item>
        </div>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="cancelForm" size="large">
            <el-icon><Close /></el-icon>
            取消
          </el-button>
          <el-button 
            type="primary" 
            @click="saveHost" 
            :loading="saving" 
            size="large"
            class="save-btn"
          >
            <el-icon v-if="!saving">
              <component :is="editingHost ? Edit : Plus" />
            </el-icon>
            {{ editingHost ? '更新主机' : '添加主机' }}
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- CSV批量导入对话框 -->
    <el-dialog
      v-model="showCSVDialog"
      title="CSV批量导入主机"
      width="600px"
      :close-on-click-modal="false"
    >
      <div class="csv-import-section">
        <el-alert
          title="导入说明"
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 20px"
        >
          <template #default>
            <p>1. 请使用提供的CSV模板格式</p>
            <p>2. 必填字段：主机名(name)、IP地址(ip)</p>
            <p>3. 支持的认证类型：password、key</p>
            <p>4. 标签请用分号(;)分隔</p>
          </template>
        </el-alert>

        <el-upload
          ref="csvUploadRef"
          class="csv-upload"
          drag
          :auto-upload="false"
          :limit="1"
          accept=".csv"
          :on-change="handleCSVFileChange"
          :on-exceed="handleCSVFileExceed"
        >
          <el-icon class="el-icon--upload"><upload-filled /></el-icon>
          <div class="el-upload__text">
            将CSV文件拖拽到此处，或<em>点击上传</em>
          </div>
          <template #tip>
            <div class="el-upload__tip">
              只能上传CSV文件，且不超过10MB
            </div>
          </template>
        </el-upload>

        <div v-if="csvFile" class="file-info">
          <el-icon><Document /></el-icon>
          <span>{{ csvFile.name }}</span>
          <span class="file-size">({{ formatFileSize(csvFile.size) }})</span>
        </div>

        <!-- 导入进度 -->
        <div v-if="importing" class="import-progress">
          <el-progress :percentage="importProgress" :status="importStatus" />
          <p class="progress-text">{{ importProgressText }}</p>
        </div>

        <!-- 导入结果 -->
        <div v-if="importResult" class="import-result">
          <el-alert
            :title="`导入完成：成功 ${importResult.success} 台，失败 ${importResult.failed} 台`"
            :type="importResult.failed > 0 ? 'warning' : 'success'"
            :closable="false"
            show-icon
          />
          
          <div v-if="importResult.failed_hosts && importResult.failed_hosts.length > 0" class="failed-hosts">
            <h4>失败详情：</h4>
            <el-table :data="importResult.failed_hosts" size="small" max-height="200">
              <el-table-column prop="index" label="行号" width="80" />
              <el-table-column prop="host.name" label="主机名" width="120" />
              <el-table-column prop="host.ip" label="IP地址" width="120" />
              <el-table-column prop="error" label="错误原因" />
            </el-table>
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="closeCSVDialog">取消</el-button>
        <el-button 
          type="primary" 
          @click="startCSVImport" 
          :loading="importing"
          :disabled="!csvFile"
        >
          开始导入
        </el-button>
      </template>
    </el-dialog>


    <!-- 主机详情对话框 -->
    <el-dialog
      v-model="showDetailDialog"
      title="主机详情"
      width="800px"
      class="host-detail-dialog"
      :close-on-click-modal="false"
    >
      <div v-if="currentHost" class="host-detail-content">
        <!-- 主机概览卡片 -->
        <div class="host-overview-card">
          <div class="host-header">
            <div class="host-avatar-large">
              <el-icon><Monitor /></el-icon>
            </div>
            <div class="host-info">
              <h2 class="host-title">{{ currentHost.name }}</h2>
              <p class="host-address">{{ currentHost.ip }}:{{ currentHost.port }}</p>
              <el-tag :type="getStatusType(currentHost.status)" effect="light" size="large" class="status-tag">
                <el-icon v-if="currentHost.status === 'online'"><CircleCheck /></el-icon>
                <el-icon v-else-if="currentHost.status === 'offline'"><CircleClose /></el-icon>
                <el-icon v-else><Warning /></el-icon>
                {{ getStatusText(currentHost.status) }}
              </el-tag>
            </div>
            <div class="host-actions">
              <el-button type="warning" @click="checkHostStatus(currentHost)" :loading="checkingStatus">
                <el-icon><Refresh /></el-icon>
                检查状态
              </el-button>
            </div>
          </div>
        </div>

        <!-- 主机基本信息 -->
        <div class="detail-section">
          <div class="section-header">
            <el-icon><Monitor /></el-icon>
            <h3>基本信息</h3>
          </div>
          <!-- 拓扑信息卡片 -->
          <div v-if="currentHost.topology" class="topology-section">
            <div class="topology-card">
              <h4>
                <el-icon><DataBoard /></el-icon>
                拓扑位置
              </h4>
              <div class="topology-path">
                <div class="topology-item">
                  <el-icon><OfficeBuilding /></el-icon>
                  <span class="topology-label">业务线：</span>
                  <span class="topology-value">{{ currentHost.topology.business }}</span>
                </div>
                <div class="topology-arrow">→</div>
                <div class="topology-item">
                  <el-icon><Folder /></el-icon>
                  <span class="topology-label">环境：</span>
                  <span class="topology-value">{{ currentHost.topology.environment }}</span>
                </div>
                <div class="topology-arrow">→</div>
                <div class="topology-item">
                  <el-icon><Monitor /></el-icon>
                  <span class="topology-label">集群：</span>
                  <span class="topology-value">{{ currentHost.topology.cluster }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 基本信息卡片 -->
          <div class="detail-cards">
            <div class="info-card">
              <div class="card-icon">
                <el-icon><Platform /></el-icon>
              </div>
              <div class="card-content">
                <div class="card-label">操作系统</div>
                <div class="card-value">{{ currentHost.os || '未知' }}</div>
              </div>
            </div>
            
            <div class="info-card">
              <div class="card-icon">
                <el-icon><Location /></el-icon>
              </div>
              <div class="card-content">
                <div class="card-label">IP地址</div>
                <div class="card-value ip-value">{{ currentHost.ip }}</div>
              </div>
            </div>
            
            <div class="info-card">
              <div class="card-icon">
                <el-icon><Connection /></el-icon>
              </div>
              <div class="card-content">
                <div class="card-label">端口</div>
                <div class="card-value">{{ currentHost.port }}</div>
              </div>
            </div>
            
            <div class="info-card">
              <div class="card-icon">
                <el-icon><User /></el-icon>
              </div>
              <div class="card-content">
                <div class="card-label">用户名</div>
                <div class="card-value">{{ currentHost.username || '未设置' }}</div>
              </div>
            </div>
            
            <div class="info-card">
              <div class="card-icon">
                <el-icon><Key /></el-icon>
              </div>
              <div class="card-content">
                <div class="card-label">认证方式</div>
                <div class="card-value">
                  {{ currentHost.auth_type === 'password' ? '密码认证' : 'SSH密钥认证' }}
                </div>
              </div>
            </div>
            
            <div class="info-card">
              <div class="card-icon">
                <el-icon><Calendar /></el-icon>
              </div>
              <div class="card-content">
                <div class="card-label">创建时间</div>
                <div class="card-value">
                  {{ new Date(currentHost.created_at).toLocaleString('zh-CN') }}
                </div>
              </div>
            </div>
          </div>
          <div class="info-card" v-if="currentHost.updated_at !== currentHost.created_at">
            <div class="card-icon">
              <el-icon><Edit /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-label">更新时间</div>
              <div class="card-value">
                {{ new Date(currentHost.updated_at).toLocaleString('zh-CN') }}
                <div class="card-value">{{ formatDate(currentHost.updated_at) }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- SSH认证信息 -->
        <div class="detail-section">
          <div class="section-header">
            <el-icon><Key /></el-icon>
            <h3>SSH认证信息</h3>
          </div>
          <div class="detail-cards">
            <div class="info-card">
              <div class="card-icon">
                <el-icon><Key /></el-icon>
              </div>
              <div class="card-content">
                <div class="card-label">认证类型</div>
                <div class="card-value">
                  <el-tag type="info" effect="plain">
                    {{ currentHost.auth_type === 'password' ? '密码认证' : 'SSH密钥认证' }}
                  </el-tag>
                </div>
              </div>
            </div>
            <div class="info-card">
              <div class="card-icon">
                <el-icon><User /></el-icon>
              </div>
              <div class="card-content">
                <div class="card-label">SSH用户名</div>
                <div class="card-value username-value">{{ currentHost.username || '未设置' }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- 描述和标签 -->
        <div class="detail-section">
          <div class="section-header">
            <el-icon><Document /></el-icon>
            <h3>附加信息</h3>
          </div>
          <div class="description-section">
            <div class="description-card">
              <h4>主机描述</h4>
              <div class="description-content">
                {{ currentHost.description || '暂无描述信息' }}
              </div>
            </div>
            <div class="tags-card">
              <h4>主机标签</h4>
              <div class="tags-container">
                <el-tag
                  v-if="currentHost.tags"
                  v-for="tag in currentHost.tags.split(',')"
                  :key="tag"
                  effect="plain"
                  class="tag-item"
                  size="large"
                >
                  <el-icon><Discount /></el-icon>
                  {{ tag.trim() }}
                </el-tag>
                <div v-else class="no-tags">
                  <el-icon><Warning /></el-icon>
                  暂无标签
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 操作记录 -->
        <div class="detail-section">
          <div class="section-header">
            <el-icon><Clock /></el-icon>
            <h3>最近操作</h3>
          </div>
          <div class="operation-timeline">
            <div class="timeline-item">
              <div class="timeline-dot success"></div>
              <div class="timeline-content">
                <div class="operation-title">主机创建</div>
                <div class="operation-time">{{ formatDate(currentHost.created_at) }}</div>
              </div>
            </div>
            <div class="timeline-item" v-if="currentHost.updated_at !== currentHost.created_at">
              <div class="timeline-dot info"></div>
              <div class="timeline-content">
                <div class="operation-title">信息更新</div>
                <div class="operation-time">{{ formatDate(currentHost.updated_at) }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <template #footer>
        <div class="detail-footer">
          <el-button @click="showDetailDialog = false" size="large">
            <el-icon><Close /></el-icon>
            关闭
          </el-button>
          <el-button type="primary" @click="editHost(currentHost)" size="large">
            <el-icon><Edit /></el-icon>
            编辑主机
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Plus, Edit, Delete, Search, Refresh, Upload, Download, ArrowDown,
  Monitor, View, CircleCheck, CircleClose, Warning, Platform, Iphone,
  QuestionFilled, Calendar, Key, Document, Clock, Close, User, Location,
  Connection, Discount, OfficeBuilding, Folder, DataBoard, InfoFilled, Lock
} from '@element-plus/icons-vue'
import api from '@/utils/api'

const hosts = ref([])
const searchText = ref('')
const statusFilter = ref('')
const osFilter = ref('')
const loading = ref(false)
const showAddDialog = ref(false)
const showCSVDialog = ref(false)
const showDetailDialog = ref(false)
const saving = ref(false)
const editingHost = ref(null)
const currentHost = ref(null)
const hostForm = ref({
  name: '',
  ip: '',
  port: 22,
  username: '',
  auth_type: 'password',
  password: '',
  private_key: '',
  passphrase: '',
  os: '',
  description: '',
  tags: ''
})
const hostFormRef = ref(null)
const csvFile = ref(null)
const csvFileRef = ref(null)
const importing = ref(false)
const checkingStatus = ref(false)
const importProgressText = ref('')
const importStatus = ref('')
const csvUploadRef = ref(null)

const hostRules = {
  name: [
    { required: true, message: '请输入主机名称', trigger: 'blur' },
    { min: 2, max: 50, message: '主机名称长度在 2 到 50 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9\-_一-龥]+$/, message: '主机名称只能包含字母、数字、中文、连字符和下划线', trigger: 'blur' }
  ],
  ip: [
    { required: true, message: '请输入IP地址', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        const ipRegex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/
        if (!ipRegex.test(value)) {
          callback(new Error('请输入正确的IP地址格式，如：192.168.1.100'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  port: [
    { required: true, message: '请输入SSH端口号', trigger: 'blur' },
    { type: 'number', min: 1, max: 65535, message: '端口号范围为 1-65535', trigger: 'blur' }
  ],
  username: [
    { required: true, message: '请输入SSH用户名', trigger: 'blur' },
    { min: 1, max: 32, message: '用户名长度在 1 到 32 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_\-]+$/, message: '用户名只能包含字母、数字、下划线和连字符', trigger: 'blur' }
  ],
  password: [
    { 
      validator: (rule, value, callback) => {
        if (!value && !editingHost.value) {
          callback(new Error('请输入SSH密码'))
        } else if (value && value.length < 1) {
          callback(new Error('密码长度至少1个字符'))
        } else {
          callback()
        }
      }, 
      trigger: 'blur' 
    }
  ]
}

const filteredHosts = computed(() => {
  return hosts.value.filter(host => {
    const matchesSearch = !searchText.value || 
      host.name.toLowerCase().includes(searchText.value.toLowerCase()) ||
      host.ip.includes(searchText.value) ||
      (host.description && host.description.toLowerCase().includes(searchText.value.toLowerCase()))
    
    const matchesStatus = !statusFilter.value || host.status === statusFilter.value
    const matchesOS = !osFilter.value || host.os === osFilter.value
    
    return matchesSearch && matchesStatus && matchesOS
  })
})

const getStatusType = (status) => {
  const statusMap = {
    online: 'success',
    offline: 'danger',
    unknown: 'warning'
  }
  return statusMap[status] || 'info'
}

const getStatusText = (status) => {
  const statusMap = {
    online: '在线',
    offline: '离线',
    unknown: '未知'
  }
  return statusMap[status] || '未知'
}

const loadHosts = async () => {
  loading.value = true
  try {
    const response = await api.get('/api/v1/hosts')
    hosts.value = response.data
  } catch (error) {
    console.error('加载主机列表失败:', error)
    ElMessage.error(`加载主机列表失败: ${error.response?.data?.error || error.message}`)
  } finally {
    loading.value = false
  }
}

const viewHost = async (host) => {
  try {
    loading.value = true
    const response = await api.get(`/api/v1/hosts/${host.id}`)
    currentHost.value = response.data
    showDetailDialog.value = true
  } catch (error) {
    console.error('获取主机详情失败:', error)
    ElMessage.error('获取主机详情失败')
  } finally {
    loading.value = false
  }
}

// 格式化日期
const formatDate = (dateString) => {
  if (!dateString) return '未知'
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const editHost = (host) => {
  editingHost.value = host
  hostForm.value = {
    name: host.name,
    ip: host.ip,
    port: host.port,
    username: host.username,
    auth_type: host.auth_type,
    password: '',
    private_key: host.private_key || '',
    passphrase: host.passphrase || '',
    os: host.os,
    description: host.description,
    tags: host.tags
  }
  showDetailDialog.value = false
  showAddDialog.value = true
}

const deleteHost = async (host) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除主机 "${host.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await api.delete(`/api/v1/hosts/${host.id}`)
    ElMessage.success('主机删除成功')
    loadHosts()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除主机失败:', error)
      
      // 处理不同类型的错误响应
      if (error.response?.status === 409) {
        // 冲突错误，主机有关联数据
        const errorData = error.response.data
        
        if (errorData.topology) {
          // 主机在拓扑中有关联
          ElMessageBox.alert(
            `${errorData.message}\n\n业务线：${errorData.topology.business}\n环境：${errorData.topology.environment}\n集群：${errorData.topology.cluster}\n\n${errorData.suggestion}`,
            '无法删除主机',
            {
              confirmButtonText: '我知道了',
              type: 'warning',
              dangerouslyUseHTMLString: false
            }
          )
        } else {
          // 其他关联数据错误
          ElMessageBox.alert(
            `${errorData.message}\n\n${errorData.suggestion || ''}`,
            errorData.error || '删除失败',
            {
              confirmButtonText: '我知道了',
              type: 'warning',
              dangerouslyUseHTMLString: false
            }
          )
        }
      } else if (error.response?.status === 404) {
        ElMessage.error('主机不存在或已被删除')
        loadHosts() // 刷新列表
      } else {
        // 其他错误
        const errorMsg = error.response?.data?.error || error.response?.data?.message || error.message
        ElMessage.error(`删除主机失败: ${errorMsg}`)
      }
    }
  }
}

const checkHostStatus = async (host) => {
  try {
    const response = await api.post(`/api/v1/hosts/${host.id}/check`)
    const { status, message } = response.data
    
    if (status === 'online') {
      ElMessage.success(`主机 ${host.name} 检查完成：${message || '连接正常'}`)
    } else if (status === 'offline') {
      ElMessage.warning(`主机 ${host.name} 检查完成：${message || '连接失败'}`)
    } else {
      ElMessage.info(`主机 ${host.name} 检查完成：${message || '状态未知'}`)
    }
    
    loadHosts()
  } catch (error) {
    console.error('检查主机状态失败:', error)
    ElMessage.error(`检查主机 ${host.name} 状态失败: ${error.response?.data?.error || error.message}`)
  }
}

const checkAllHostsStatus = async () => {
  checkingStatus.value = true
  try {
    const response = await api.post('/api/v1/hosts/check-all')
    const { results, message } = response.data
    
    if (!results || results.length === 0) {
      ElMessage.warning('没有主机需要检查')
      return
    }
    
    let onlineCount = 0
    let offlineCount = 0
    let unknownCount = 0
    
    results.forEach(result => {
      if (result.status === 'online') onlineCount++
      else if (result.status === 'offline') offlineCount++
      else unknownCount++
    })
    
    // 显示详细的检查结果
    const totalCount = results.length
    const successMessage = `批量检查完成！共检查 ${totalCount} 台主机：\n` +
      `• 在线: ${onlineCount} 台\n` +
      `• 离线: ${offlineCount} 台\n` +
      `• 未知: ${unknownCount} 台`
    
    if (offlineCount > 0 || unknownCount > 0) {
      ElMessage.warning(successMessage)
    } else {
      ElMessage.success(successMessage)
    }
    
    // 刷新主机列表
    loadHosts()
  } catch (error) {
    console.error('批量检查主机状态失败:', error)
    ElMessage.error(`批量检查主机状态失败: ${error.response?.data?.error || error.message}`)
  } finally {
    checkingStatus.value = false
  }
}

const saveHost = async () => {
  if (!hostFormRef.value) return
  
  try {
    const valid = await hostFormRef.value.validate()
    if (!valid) {
      ElMessage.warning('请检查表单填写是否正确')
      return
    }
  } catch (error) {
    ElMessage.warning('请检查表单填写是否正确')
    return
  }

  // 验证IP地址是否已存在（仅新增时）
  if (!editingHost.value) {
    const existingHost = hosts.value.find(host => host.ip === hostForm.value.ip)
    if (existingHost) {
      ElMessage.error(`IP地址 ${hostForm.value.ip} 已存在，请使用其他IP地址`)
      return
    }
  }

  saving.value = true
  try {
    const formData = { ...hostForm.value }
    
    // 处理标签格式
    if (formData.tags) {
      formData.tags = formData.tags.split(',').map(tag => tag.trim()).filter(tag => tag).join(',')
    }
    
    if (editingHost.value) {
      // 编辑模式：如果密码为空，则不更新密码
      if (!formData.password) {
        delete formData.password
      }
      await api.put(`/api/v1/hosts/${editingHost.value.id}`, formData)
      ElMessage.success('主机信息更新成功')
    } else {
      await api.post('/api/v1/hosts', formData)
      ElMessage.success('主机添加成功')
    }
    
    showAddDialog.value = false
    resetForm()
    loadHosts()
  } catch (error) {
    console.error('保存主机失败:', error)
    const action = editingHost.value ? '更新' : '添加'
    const errorMsg = error.response?.data?.error || error.message
    ElMessage.error(`${action}主机失败: ${errorMsg}`)
  } finally {
    saving.value = false
  }
}

const cancelForm = () => {
  ElMessageBox.confirm(
    '确定要取消吗？未保存的更改将会丢失。',
    '确认取消',
    {
      confirmButtonText: '确定',
      cancelButtonText: '继续编辑',
      type: 'warning'
    }
  ).then(() => {
    showAddDialog.value = false
    resetForm()
  }).catch(() => {
    // 用户选择继续编辑，不做任何操作
  })
}

const resetForm = () => {
  editingHost.value = null
  hostForm.value = {
    name: '',
    ip: '',
    port: 22,
    username: '',
    auth_type: 'password',
    password: '',
    private_key: '',
    passphrase: '',
    os: '',
    description: '',
    tags: ''
  }
}

// 批量操作处理
const handleBatchAction = (command) => {
  switch (command) {
    case 'csv-upload':
      showCSVDialog.value = true
      break
    case 'download-template':
      downloadCSVTemplate()
      break
  }
}

// 下载CSV模板
const downloadCSVTemplate = async () => {
  try {
    const response = await api.get('/api/v1/hosts/csv-template', {
      responseType: 'blob'
    })
    
    const blob = new Blob([response.data], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = 'host_import_template.csv'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    
    ElMessage.success('CSV模板下载成功')
  } catch (error) {
    ElMessage.error('下载CSV模板失败')
  }
}

// CSV文件选择处理
const handleCSVFileChange = (file) => {
  const rawFile = file.raw
  
  // 验证文件类型
  if (!rawFile.name.toLowerCase().endsWith('.csv')) {
    ElMessage.error('请选择CSV格式的文件')
    return false
  }
  
  // 验证文件大小 (10MB)
  if (rawFile.size > 10 * 1024 * 1024) {
    ElMessage.error('文件大小不能超过10MB')
    return false
  }
  
  csvFile.value = rawFile
  importResult.value = null
  ElMessage.success(`文件 ${rawFile.name} 选择成功`)
}

// 文件数量超限处理
const handleCSVFileExceed = () => {
  ElMessage.warning('只能选择一个CSV文件')
}

// 格式化文件大小
const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 开始CSV导入
const startCSVImport = async () => {
  if (!csvFile.value) {
    ElMessage.error('请先选择CSV文件')
    return
  }
  
  importing.value = true
  importProgress.value = 0
  importStatus.value = ''
  importProgressText.value = '正在上传文件...'
  importResult.value = null
  
  try {
    const formData = new FormData()
    formData.append('file', csvFile.value)
    
    // 模拟进度更新
    const progressInterval = setInterval(() => {
      if (importProgress.value < 90) {
        importProgress.value += Math.random() * 10
        if (importProgress.value < 30) {
          importProgressText.value = '正在解析CSV文件...'
        } else if (importProgress.value < 60) {
          importProgressText.value = '正在验证主机信息...'
        } else {
          importProgressText.value = '正在导入主机数据...'
        }
      }
    }, 200)
    
    const response = await api.post('/api/v1/hosts/batch/import-csv', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    
    clearInterval(progressInterval)
    importProgress.value = 100
    importStatus.value = response.data.failed > 0 ? 'warning' : 'success'
    importProgressText.value = '导入完成'
    importResult.value = response.data
    
    ElMessage.success(`CSV导入完成：成功 ${response.data.success} 台，失败 ${response.data.failed} 台`)
    
    // 刷新主机列表
    if (response.data.success > 0) {
      loadHosts()
    }
    
  } catch (error) {
    clearInterval(progressInterval)
    importProgress.value = 100
    importStatus.value = 'exception'
    importProgressText.value = '导入失败'
    
    const errorMsg = error.response?.data?.error || '导入失败'
    ElMessage.error(errorMsg)
  } finally {
    importing.value = false
  }
}

// 关闭CSV导入对话框
const closeCSVDialog = () => {
  showCSVDialog.value = false
  csvFile.value = null
  importing.value = false
  importProgress.value = 0
  importStatus.value = ''
  importProgressText.value = ''
  importResult.value = null
  
  // 清空上传组件
  if (csvUploadRef.value) {
    csvUploadRef.value.clearFiles()
  }
}

onMounted(() => {
  loadHosts()
})
</script>

<style scoped>
.hosts-page {
  padding: 24px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  padding: 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  color: white;
  box-shadow: 0 8px 32px rgba(102, 126, 234, 0.3);
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
  color: #ffffff;
}

.page-header h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: white;
}

.page-subtitle {
  margin: 0;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
}

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.create-btn {
  background: linear-gradient(135deg, #ff6b6b 0%, #ee5a24 100%);
  border: none;
  box-shadow: 0 4px 15px rgba(255, 107, 107, 0.3);
}

.create-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(255, 107, 107, 0.4);
}

.search-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.search-left {
  display: flex;
  gap: 16px;
  align-items: center;
  flex: 1;
}

.search-input {
  width: 320px;
}

.status-filter, .os-filter {
  width: 140px;
}

.search-right {
  display: flex;
  gap: 12px;
}

.hosts-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

.empty-state {
  padding: 60px 20px;
  text-align: center;
}

.hosts-table {
  width: 100%;
}

.host-info-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.host-avatar {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 18px;
}

.host-details {
  flex: 1;
}

.host-name {
  font-weight: 600;
  color: #303133;
  font-size: 14px;
  margin-bottom: 4px;
}

.host-name.clickable {
  color: #409eff;
  cursor: pointer;
  transition: color 0.3s;
}

.host-name.clickable:hover {
  color: #66b1ff;
  text-decoration: underline;
}

.host-ip {
  color: #909399;
  font-size: 12px;
  font-family: 'Courier New', monospace;
}

.os-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #606266;
}

.description-text {
  color: #606266;
  font-size: 13px;
}

.tags-cell {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.tag-item {
  font-size: 11px;
}

.no-tags {
  color: #c0c4cc;
  font-size: 12px;
  font-style: italic;
}

.time-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #909399;
  font-size: 12px;
}

.action-buttons {
  display: flex;
  gap: 4px;
  justify-content: center;
  flex-wrap: wrap;
  align-items: center;
}

.action-buttons .el-button {
  min-width: 50px;
  padding: 4px 8px;
}

/* 主机表单对话框样式 */
.host-form-dialog {
  border-radius: 16px;
}

.host-form-dialog .el-dialog__header {
  padding: 0;
  border-bottom: none;
}

.host-form-dialog .el-dialog__body {
  padding: 0 24px 24px 24px;
}

.dialog-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 24px 24px 20px 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 16px 16px 0 0;
  margin: -20px -24px 24px -24px;
}

.dialog-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  backdrop-filter: blur(10px);
}

.dialog-title-info h3 {
  margin: 0 0 4px 0;
  font-size: 20px;
  font-weight: 600;
}

.dialog-title-info p {
  margin: 0;
  font-size: 14px;
  opacity: 0.9;
}

.host-form {
  padding: 0;
}

.form-section {
  margin-bottom: 24px;
  padding: 24px;
  background: #fafbfc;
  border-radius: 12px;
  border: 1px solid #e4e7ed;
  transition: all 0.3s ease;
}

.form-section:hover {
  border-color: #c6e2ff;
  background: #f0f9ff;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0 0 20px 0;
  font-size: 16px;
  font-weight: 600;
  color: #409eff;
  position: relative;
}

.section-title span {
  flex-shrink: 0;
}

.section-divider {
  flex: 1;
  height: 1px;
  background: linear-gradient(to right, #409eff, transparent);
  margin-left: 12px;
}

.form-tip {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
  line-height: 1.4;
  padding: 8px 12px;
  background: #f0f9ff;
  border-radius: 6px;
  border-left: 3px solid #409eff;
}

.option-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
}

.el-input-number .el-input__inner {
  text-align: left;
}

.el-select-dropdown__item {
  padding: 8px 20px;
}

.el-form-item__label {
  font-weight: 500;
  color: #606266;
}

.el-input__prefix {
  color: #909399;
}

.el-textarea__inner {
  font-family: inherit;
}

.auth-type-group {
  display: flex;
  gap: 20px;
}

.auth-radio {
  margin-right: 0;
  padding: 12px 16px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  transition: all 0.3s;
  flex: 1;
}

.auth-radio:hover {
  border-color: #409eff;
  background-color: #f0f9ff;
}

.auth-radio.is-checked {
  border-color: #409eff;
  background-color: #f0f9ff;
}

.radio-content {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.auth-alert {
  margin-bottom: 20px;
  border-radius: 8px;
}

.alert-content {
  line-height: 1.6;
}

.alert-content p {
  margin: 0 0 4px 0;
}

.alert-content p:last-child {
  margin-bottom: 0;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
  margin-top: 20px;
}

.save-btn {
  background: linear-gradient(135deg, #67c23a 0%, #85ce61 100%);
  border: none;
  box-shadow: 0 4px 12px rgba(103, 194, 58, 0.3);
  transition: all 0.3s ease;
}

.save-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(103, 194, 58, 0.4);
}

.save-btn:active {
  transform: translateY(0);
}

/* 主机详情对话框样式 */
.host-detail-dialog {
  border-radius: 16px;
}

.host-detail-content {
  max-height: 70vh;
  overflow-y: auto;
  padding: 0 8px;
}

/* 主机概览卡片 */
.host-overview-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  padding: 24px;
  margin-bottom: 24px;
  color: white;
}

.host-header {
  display: flex;
  align-items: center;
  gap: 20px;
}

.host-avatar-large {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: white;
  backdrop-filter: blur(10px);
}

.host-info {
  flex: 1;
}

.host-title {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: white;
}

.host-address {
  margin: 0 0 12px 0;
  font-size: 16px;
  color: rgba(255, 255, 255, 0.8);
  font-family: 'Courier New', monospace;
}

.status-tag {
  background: rgba(255, 255, 255, 0.2) !important;
  border: 1px solid rgba(255, 255, 255, 0.3) !important;
  color: white !important;
  backdrop-filter: blur(10px);
}

.host-actions {
  display: flex;
  gap: 12px;
}

/* 拓扑信息样式 */
.topology-section {
  margin-bottom: 24px;
}

.topology-card {
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.topology-card h4 {
  margin: 0 0 16px 0;
  font-size: 16px;
  color: #409eff;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.topology-path {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.topology-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e9ecef;
}

.topology-label {
  font-size: 12px;
  color: #909399;
  font-weight: 500;
}

.topology-value {
  font-size: 14px;
  color: #303133;
  font-weight: 600;
}

.topology-arrow {
  color: #409eff;
  font-weight: bold;
  font-size: 16px;
}

/* 详情卡片样式 */
.detail-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 16px;
}

.info-card {
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.3s;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.info-card:hover {
  border-color: #409eff;
  box-shadow: 0 4px 16px rgba(64, 158, 255, 0.1);
  transform: translateY(-2px);
}

.card-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(135deg, #409eff 0%, #66b1ff 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 20px;
  flex-shrink: 0;
}

.card-content {
  flex: 1;
}

.card-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
  font-weight: 500;
}

.card-value {
  font-size: 16px;
  color: #303133;
  font-weight: 600;
  word-break: break-all;
}

/* 描述和标签区域 */
.description-section {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.description-card, .tags-card {
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.description-card h4, .tags-card h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #409eff;
  font-weight: 600;
}

.description-content {
  color: #606266;
  line-height: 1.6;
  font-size: 14px;
  min-height: 40px;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  min-height: 40px;
  align-items: flex-start;
}

.tag-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.no-tags {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #c0c4cc;
  font-size: 14px;
}

.detail-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 12px;
}

.detail-section {
  margin-bottom: 24px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 2px solid #f0f2f5;
}

.section-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-item.full-width {
  grid-column: 1 / -1;
}

.detail-item .label {
  font-size: 12px;
  color: #909399;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.detail-item .value {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

.ip-value {
  font-family: 'Courier New', monospace;
  background: #f0f0f0;
  padding: 4px 8px;
  border-radius: 4px;
  display: inline-block;
}

.username-value {
  font-family: 'Courier New', monospace;
  color: #667eea;
}

.description-value {
  color: #606266;
  line-height: 1.5;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 4px;
}

.operation-timeline {
  position: relative;
  padding-left: 20px;
}

.timeline-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
  position: relative;
}

.timeline-item::before {
  content: '';
  position: absolute;
  left: -12px;
  top: 8px;
  bottom: -8px;
  width: 2px;
  background: #e4e7ed;
}

.timeline-item:last-child::before {
  display: none;
}

.timeline-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-left: -16px;
  margin-top: 4px;
  position: relative;
  z-index: 1;
}

.timeline-dot.success {
  background: #67c23a;
}

.timeline-dot.info {
  background: #409eff;
}

.timeline-content {
  flex: 1;
}

.operation-title {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
  margin-bottom: 2px;
}

.operation-time {
  font-size: 12px;
  color: #909399;
}

.detail-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid #ebeef5;
}

/* 表单样式保持原有 */
.dialog-footer {
  text-align: right;
}

.form-row {
  display: flex;
  gap: 20px;
}

.form-row .el-form-item {
  flex: 1;
}

.csv-import-tips {
  background-color: #f4f4f5;
  border: 1px solid #e9e9eb;
  border-radius: 4px;
  padding: 12px;
  margin-bottom: 16px;
}

.csv-import-tips h4 {
  margin: 0 0 8px 0;
  color: #303133;
}

.csv-import-tips ul {
  margin: 0;
  padding-left: 20px;
  color: #606266;
}

.csv-import-tips li {
  margin-bottom: 4px;
}

.import-progress {
  margin: 16px 0;
}

.import-result {
  margin-top: 16px;
  padding: 12px;
  border-radius: 4px;
}

.import-result.success {
  background-color: #f0f9ff;
  border: 1px solid #67c23a;
  color: #67c23a;
}

.import-result.error {
  background-color: #fef0f0;
  border: 1px solid #f56c6c;
  color: #f56c6c;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .hosts-page {
    padding: 16px;
  }
  
  .page-header {
    flex-direction: column;
    gap: 16px;
  }
  
  .search-left {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  
  .search-input {
    width: 100%;
  }
  
  .detail-grid {
    grid-template-columns: 1fr;
  }
  
  .host-form-dialog {
    width: 95% !important;
    margin: 5vh auto;
  }
  
  .dialog-header {
    padding: 20px 16px 16px 16px;
    margin: -20px -16px 20px -16px;
  }
  
  .host-form {
    padding: 0;
  }
  
  .form-section {
    padding: 16px;
    margin-bottom: 16px;
  }
  
  .el-row {
    margin: 0;
  }
  
  .el-col {
    padding: 0 0 16px 0;
  }
  
  .dialog-footer {
    flex-direction: column;
    gap: 8px;
  }
  
  .dialog-footer .el-button {
    width: 100%;
    margin: 0;
  }
}

.csv-upload .el-icon--upload {
  font-size: 48px;
  color: #c0c4cc;
  margin-bottom: 16px;
}

.csv-upload .el-upload__text {
  color: #606266;
  font-size: 14px;
  line-height: 1.4;
}

.csv-upload .el-upload__text em {
  color: #409eff;
  font-style: normal;
}
</style>
