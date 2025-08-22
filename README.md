# DevOps管理平台

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Vue Version](https://img.shields.io/badge/Vue-3.0+-green.svg)](https://vuejs.org)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

基于Go和Vue.js的现代化DevOps管理平台，提供主机管理、脚本执行、作业调度、实时监控等功能。专为中小型团队设计，简单易用，功能完善。

## ✨ 功能特性

### 🖥️ 主机管理
- **主机信息管理** - 支持SSH密码和密钥认证
- **实时状态监控** - 定时检查主机连接状态
- **批量操作** - CSV导入导出、批量状态检查
- **主机拓扑** - 业务-环境-集群三级拓扑结构

### 📝 脚本管理
- **多语言支持** - Shell、Python、PowerShell脚本
- **在线编辑** - 代码高亮、语法检查
- **版本管理** - 脚本历史记录和版本对比
- **快速执行** - 一键执行脚本到指定主机

### ⚙️ 作业执行
- **批量执行** - 同时在多台主机执行脚本
- **实时日志** - WebSocket实时推送执行日志
- **执行记录** - 完整的执行历史和结果查看
- **重新执行** - 支持快速重新执行失败的作业

### 👥 用户管理
- **用户认证** - JWT Token认证机制
- **角色权限** - 管理员和普通用户角色
- **个人资料** - 用户信息管理和密码修改
- **操作日志** - 完整的用户操作审计

### 📊 数据可视化
- **仪表盘** - 系统概览和关键指标
- **图表展示** - 作业趋势、主机状态分布
- **实时数据** - 动态更新的统计信息

## 技术栈

### 后端
- Go 1.21+
- Gin Web框架
- GORM ORM
- SQLite数据库
- JWT认证
- WebSocket实时通信

### 前端
- Vue 3 + Composition API
- Element Plus UI组件库
- Vue Router 4
- Pinia状态管理
- ECharts图表
- Vite构建工具

## 🚀 快速开始

### 环境要求
- Go 1.21+
- Node.js 16+
- npm或yarn

### 克隆项目
```bash
git clone https://github.com/your-username/go-devops.git
cd go-devops
```

### 后端启动

1. **安装依赖**
```bash
go mod tidy
```

2. **启动服务**
```bash
go run main.go
```
后端服务将在 http://localhost:8080 启动

### 前端启动

1. **安装依赖**
```bash
cd web
npm install
```

2. **开发模式**
```bash
npm run dev
```
前端开发服务器将在 http://localhost:5173 启动

3. **生产构建**
```bash
npm run build
```

### 🐳 Docker部署

1. **构建镜像**
```bash
docker build -t go-devops .
```

2. **运行容器**
```bash
docker run -d -p 8080:8080 -v $(pwd)/data:/app/data go-devops
```

### 📦 生产部署

1. **构建前端**
```bash
cd web
npm run build
```

2. **构建后端**
```bash
go build -o devops-server main.go
```

3. **运行服务**
```bash
./devops-server
```

## 📁 项目结构

```
go-devops/
├── internal/                 # 后端核心代码
│   ├── api/                 # API路由定义
│   ├── config/              # 配置管理
│   ├── database/            # 数据库初始化和迁移
│   ├── executor/            # 脚本执行引擎
│   ├── handlers/            # HTTP请求处理器
│   ├── logger/              # 日志系统
│   ├── middleware/          # 中间件
│   ├── models/              # 数据模型
│   └── ssh/                 # SSH连接管理
├── web/                     # 前端Vue应用
│   ├── src/
│   │   ├── components/      # 可复用组件
│   │   ├── layout/          # 布局组件
│   │   ├── router/          # 路由配置
│   │   ├── stores/          # Pinia状态管理
│   │   ├── utils/           # 工具函数
│   │   └── views/           # 页面组件
│   ├── package.json
│   └── vite.config.js
├── logs/                    # 日志文件
├── templates/               # 模板文件
├── go.mod
├── go.sum
├── main.go                  # 应用入口
└── devops.db               # SQLite数据库文件
```

## 📚 API文档

### 🔐 用户认证
- `POST /api/v1/login` - 用户登录
- `POST /api/v1/register` - 用户注册
- `GET /api/v1/profile` - 获取用户信息
- `PUT /api/v1/profile` - 更新用户信息
- `PUT /api/v1/profile/password` - 修改密码
- `GET /api/v1/profile/stats` - 获取用户统计

### 🖥️ 主机管理
- `GET /api/v1/hosts` - 获取主机列表
- `POST /api/v1/hosts` - 创建主机
- `GET /api/v1/hosts/:id` - 获取主机详情
- `PUT /api/v1/hosts/:id` - 更新主机
- `DELETE /api/v1/hosts/:id` - 删除主机
- `POST /api/v1/hosts/:id/check` - 检查主机状态
- `POST /api/v1/hosts/:id/test-ssh` - 测试SSH连接
- `POST /api/v1/hosts/batch/import-csv` - CSV批量导入

### 📝 脚本管理
- `GET /api/v1/scripts` - 获取脚本列表
- `POST /api/v1/scripts` - 创建脚本
- `GET /api/v1/scripts/:id` - 获取脚本详情
- `PUT /api/v1/scripts/:id` - 更新脚本
- `DELETE /api/v1/scripts/:id` - 删除脚本
- `POST /api/v1/scripts/quick-execute` - 快速执行脚本

### ⚙️ 作业管理
- `GET /api/v1/jobs` - 获取作业列表
- `POST /api/v1/jobs` - 创建作业
- `POST /api/v1/jobs/:id/execute` - 执行作业
- `GET /api/v1/jobs/:id/executions` - 获取执行记录
- `GET /api/v1/executions` - 获取所有执行记录
- `GET /api/v1/executions/:id` - 获取执行详情

### 📊 仪表盘
- `GET /api/v1/dashboard/stats` - 获取统计数据
- `GET /api/v1/dashboard/activities` - 获取活动记录
- `GET /api/v1/dashboard/job-trend` - 获取作业趋势
- `GET /api/v1/dashboard/host-status` - 获取主机状态分布

### 🏗️ 拓扑管理
- `GET /api/v1/topology/tree` - 获取拓扑树
- `GET /api/v1/topology/businesses` - 获取业务列表
- `POST /api/v1/topology/businesses` - 创建业务
- `GET /api/v1/topology/environments` - 获取环境列表
- `GET /api/v1/topology/clusters` - 获取集群列表

## ⚙️ 配置说明

### 环境变量
- `ENVIRONMENT` - 运行环境 (development/production)
- `PORT` - 服务端口 (默认: 8080)
- `DATABASE_URL` - 数据库文件路径 (默认: devops.db)
- `JWT_SECRET` - JWT密钥 (默认: your-secret-key)

### 配置文件
系统支持通过配置文件进行设置，配置文件位置：`config/config.yaml`

## 👤 默认账户

系统启动后会自动创建管理员账户：
- **用户名**: admin
- **密码**: admin123
- **角色**: 管理员

> ⚠️ **安全提示**: 首次登录后请立即修改默认密码！

## 🔧 使用指南

### 1. 添加主机
1. 登录系统后，进入"主机管理"页面
2. 点击"添加主机"按钮
3. 填写主机信息（IP、端口、认证方式等）
4. 测试SSH连接确保配置正确
5. 保存主机信息

### 2. 创建脚本
1. 进入"脚本管理"页面
2. 点击"新建脚本"
3. 选择脚本类型（Shell/Python/PowerShell）
4. 编写脚本内容
5. 保存脚本

### 3. 执行作业
1. 进入"脚本管理"页面，选择要执行的脚本
2. 点击"执行"按钮
3. 选择目标主机
4. 点击"开始执行"
5. 实时查看执行日志和结果

### 4. 查看执行记录
1. 进入"执行记录"页面
2. 查看所有执行历史
3. 点击记录可查看详细日志
4. 支持重新执行失败的作业

## 🤝 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者！

---

如果您觉得这个项目有用，请给我们一个 ⭐️ ！
