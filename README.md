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
- **智能删除保护** - 检测主机关联关系，防止误删

### 📝 脚本管理
- **多语言支持** - Shell、Python2/3、PowerShell脚本
- **在线编辑** - 代码高亮、语法检查、格式化
- **版本管理** - 脚本历史记录和版本对比
- **快速执行** - 一键执行脚本到指定主机
- **模板支持** - 内置常用脚本模板

### ⚙️ 作业执行
- **批量执行** - 同时在多台主机执行脚本
- **实时日志** - WebSocket实时推送执行日志
- **执行记录** - 完整的执行历史和结果查看
- **重新执行** - 支持快速重新执行失败的作业
- **执行统计** - 执行耗时、成功率等指标统计

### 👥 用户管理
- **用户认证** - JWT Token认证机制
- **角色权限** - 管理员和普通用户角色
- **个人资料** - 用户信息管理和密码修改
- **活动记录** - 完整的用户操作审计和活动追踪

### 📊 数据可视化
- **仪表盘** - 系统概览和关键指标
- **图表展示** - 作业趋势、主机状态分布
- **实时数据** - 动态更新的统计信息
- **活动监控** - 用户操作活动的实时展示和历史查询

## 技术栈

### 后端
- Go 1.21+
- Gin Web框架
- GORM ORM
- 多数据库支持（SQLite/MySQL）
- JWT认证
- WebSocket实时通信
- YAML配置管理
- 定时任务调度
- 用户活动记录系统
- 完善的错误处理和日志记录

### 前端
- Vue 3 + Composition API
- Element Plus UI组件库
- Vue Router 4
- Pinia状态管理
- ECharts图表
- Vite构建工具

## 🚀 快速开始

### 环境要求
- **Go** 1.21+
- **Node.js** 16+
- **npm** 或 yarn
- **数据库** SQLite 3.x 或 MySQL 5.7+（可选）
- **操作系统** Linux/macOS/Windows

### 克隆项目
```bash
git clone https://github.com/your-username/go-devops.git
cd go-devops
```

> 📝 **注意**: 请替换为您的实际仓库地址

### 后端启动

1. **安装依赖**
```bash
go mod tidy
```

2. **配置数据库**
编辑 `config/config.yaml` 文件，选择数据库类型：
```yaml
database:
  type: "sqlite"  # 或 "mysql"
  sqlite:
    file: "devops.db"
  mysql:
    host: "localhost"
    port: 3306
    username: "devops"
    password: "password"
    database: "devops"
```

3. **启动服务**
```bash
go run main.go
```
后端服务将在 http://localhost:8080 启动

> ℹ️ **提示**: 首次启动会自动创建数据库表和默认管理员账户

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
前端开发服务器将在 http://localhost:3000 启动

> 🌐 **访问**: 打开浏览器访问 http://localhost:3000 即可使用系统

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
│   ├── services/            # 业务服务层
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
├── config/                  # 配置文件目录
│   └── config.yaml         # 主配置文件
├── logs/                    # 日志文件
├── templates/               # 模板文件
├── go.mod
├── go.sum
├── main.go                  # 应用入口
└── devops.db               # SQLite数据库文件（默认）
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
- `GET /api/v1/dashboard/recent-activities` - 获取最近活动记录
- `GET /api/v1/dashboard/activities` - 获取全部活动记录（支持分页和筛选）
- `GET /api/v1/dashboard/job-trend` - 获取作业趋势
- `GET /api/v1/dashboard/host-status` - 获取主机状态分布

### 🔧 系统信息
- `GET /api/v1/system/info` - 获取系统信息
- `GET /api/v1/system/health` - 系统健康检查

### 🏗️ 拓扑管理
- `GET /api/v1/topology/tree` - 获取拓扑树
- `GET /api/v1/topology/businesses` - 获取业务列表
- `POST /api/v1/topology/businesses` - 创建业务
- `GET /api/v1/topology/environments` - 获取环境列表
- `GET /api/v1/topology/clusters` - 获取集群列表

## ⚙️ 配置说明

### 配置文件
系统使用 YAML 格式的配置文件，位置：`config/config.yaml`

```yaml
# 应用配置
app:
  name: "DevOps管理平台"
  version: "1.0.0"
  environment: "development"  # development, production, test
  port: 8080

# 数据库配置
database:
  type: "sqlite"  # sqlite, mysql
  # SQLite 配置
  sqlite:
    file: "devops.db"
  # MySQL 配置
  mysql:
    host: "localhost"
    port: 3306
    username: "devops"
    password: "password"
    database: "devops"
    charset: "utf8mb4"
    parse_time: true
    loc: "Local"

# JWT配置
jwt:
  secret: "your-secret-key-change-in-production"
  expire_hours: 24

# 日志配置
logging:
  level: "info"  # debug, info, warn, error, fatal
  to_file: true
  file_path: "logs/"

# 定时任务配置
scheduler:
  enabled: true
  host_check_interval: "5m"  # 主机状态检查间隔
```

### 环境变量覆盖
以下环境变量可以覆盖配置文件设置：
- `ENVIRONMENT` - 运行环境
- `PORT` - 服务端口
- `DATABASE_URL` - SQLite数据库文件路径
- `JWT_SECRET` - JWT密钥
- `LOG_LEVEL` - 日志级别
- `LOG_TO_FILE` - 是否输出日志到文件

### 数据库配置

#### SQLite（默认）
```yaml
database:
  type: "sqlite"
  sqlite:
    file: "devops.db"
```

#### MySQL
```yaml
database:
  type: "mysql"
  mysql:
    host: "localhost"
    port: 3306
    username: "devops"
    password: "password"
    database: "devops"
    charset: "utf8mb4"
    parse_time: true
    loc: "Local"
```

使用 MySQL 前需要先创建数据库：
```sql
CREATE DATABASE devops CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

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
3. 选择脚本类型（Shell/Python2/Python3/PowerShell）
4. 编写脚本内容，支持代码高亮和格式化
5. 保存脚本

### 3. 执行作业
1. 进入"脚本管理"页面，选择要执行的脚本
2. 点击"执行"按钮
3. 选择目标主机（支持多选）
4. 点击"开始执行"
5. 实时查看执行日志和结果

### 4. 快速执行脚本
1. 进入"脚本管理"页面
2. 点击"快速执行"按钮
3. 在线编写脚本内容
4. 选择目标主机
5. 立即执行并查看结果

### 5. 查看执行记录
1. 进入"执行记录"页面
2. 查看所有执行历史
3. 点击记录可查看详细日志
4. 支持重新执行失败的作业
5. 可以复制输出内容和错误信息

### 6. 监控用户活动
1. 进入"仪表盘"页面查看最近活动
2. 点击"查看全部"查看完整活动记录
3. 支持按类型、时间范围、关键词筛选
4. 实时追踪用户操作和系统事件

### 7. 主机拓扑管理
1. 进入"主机拓扑"页面查看拓扑结构
2. 按业务-环境-集群三级结构组织主机
3. 支持拖拽调整主机位置
4. 直观展示主机间的关联关系

## 🔍 故障排除

### 常见问题

#### 1. 后端启动失败
- 检查Go版本是否为1.21+
- 确认8080端口未被占用
- 检查配置文件格式是否正确
- 查看日志文件获取详细错误信息

#### 2. 前端无法访问
- 确认前端服务已启动
- 检查3000端口是否可访问
- 验证后端API是否正常响应
- 清除浏览器缓存重试

#### 3. SSH连接失败
- 验证主机IP和端口是否正确
- 检查SSH认证信息（用户名/密码/密钥）
- 确认目标主机SSH服务已启动
- 检查网络连通性

#### 4. 脚本执行失败
- 检查脚本语法是否正确
- 验证目标主机环境（Python版本等）
- 确认执行权限是否足够
- 查看详细错误日志

### 日志查看
```bash
# 查看应用日志
tail -f logs/app-$(date +%Y-%m-%d).log

# 查看系统运行状态
ps aux | grep devops
```

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

### 代码规范
- Go代码遵循 `gofmt` 格式
- Vue代码使用 ESLint 规范
- 提交信息使用英文，格式清晰
- 添加必要的注释和文档

## 🆕 更新日志

### v1.0.0 (2024-08-22)
- ✨ 完整的用户活动记录系统
- 🔧 主机删除保护机制
- 🎨 优化前端UI和用户体验
- 🐛 修复执行详情页面重新执行功能
- 📊 完善仪表盘数据展示
- 🔐 增强错误处理和安全性

## 📞 支持

如果您在使用过程中遇到问题，可以通过以下方式获取帮助：

- 📖 查看文档和使用指南
- 🐛 提交 Issue 报告问题
- 💬 参与讨论和交流
- 📧 联系维护团队

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者和用户！

特别感谢以下技术和项目：
- [Go](https://golang.org/) - 强大的后端语言
- [Vue.js](https://vuejs.org/) - 优秀的前端框架
- [Element Plus](https://element-plus.org/) - 精美的UI组件库
- [GORM](https://gorm.io/) - 优雅的ORM框架

---

如果您觉得这个项目有用，请给我们一个 ⭐️ ！

**让DevOps管理更简单、更高效！** 🚀
