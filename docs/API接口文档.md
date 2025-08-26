# DevOps管理平台 - 后端API接口文档

## 项目概述

DevOps管理平台是一个基于Go语言开发的运维管理系统，提供主机管理、脚本管理、作业调度、文件管理、拓扑管理等功能。

### 技术栈
- **后端框架**: Gin (Go Web框架)
- **数据库**: MySQL/SQLite (支持双数据库)
- **ORM**: GORM
- **认证**: JWT
- **日志**: Logrus
- **SSH**: golang.org/x/crypto/ssh

### 基础信息
- **基础URL**: `http://localhost:8080/api/v1`
- **认证方式**: JWT Bearer Token
- **数据格式**: JSON
- **字符编码**: UTF-8

---

## 认证说明

除了公开接口外，所有API都需要在请求头中携带JWT Token：

```http
Authorization: Bearer <your_jwt_token>
```

---

## 1. 认证管理 (Authentication)

### 1.1 用户登录
- **接口**: `POST /login`
- **描述**: 用户登录获取JWT Token
- **权限**: 公开接口

**请求参数**:
```json
{
  "username": "admin",
  "password": "123456"
}
```

**响应示例**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "role": "admin",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 1.2 用户注册
- **接口**: `POST /register`
- **描述**: 新用户注册
- **权限**: 公开接口

**请求参数**:
```json
{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "123456"
}
```

### 1.3 获取用户信息
- **接口**: `GET /profile`
- **描述**: 获取当前用户信息
- **权限**: 需要认证

### 1.4 更新用户信息
- **接口**: `PUT /profile`
- **描述**: 更新当前用户信息
- **权限**: 需要认证

### 1.5 修改密码
- **接口**: `PUT /profile/password`
- **描述**: 修改当前用户密码
- **权限**: 需要认证

**请求参数**:
```json
{
  "old_password": "oldpass",
  "new_password": "newpass"
}
```

### 1.6 获取用户统计
- **接口**: `GET /profile/stats`
- **描述**: 获取用户相关统计信息
- **权限**: 需要认证

**响应示例**:
```json
{
  "script_count": 5,
  "execution_count": 20,
  "last_login_days": 1
}
```

---

## 2. 主机管理 (Host Management)

### 2.1 获取主机列表
- **接口**: `GET /hosts`
- **描述**: 获取所有主机列表
- **权限**: 需要认证

**响应示例**:
```json
[
  {
    "id": 1,
    "name": "Web服务器1",
    "ip": "192.168.1.100",
    "port": 22,
    "os": "Ubuntu 20.04",
    "status": "online",
    "description": "前端Web服务器",
    "tags": "web,frontend",
    "auth_type": "password",
    "username": "root",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

### 2.2 创建主机
- **接口**: `POST /hosts`
- **描述**: 创建新主机
- **权限**: 需要认证

**请求参数**:
```json
{
  "name": "新服务器",
  "ip": "192.168.1.101",
  "port": 22,
  "os": "CentOS 7",
  "description": "数据库服务器",
  "tags": "database,mysql",
  "auth_type": "password",
  "username": "root",
  "password": "password123"
}
```

### 2.3 获取单个主机
- **接口**: `GET /hosts/:id`
- **描述**: 获取指定主机详细信息
- **权限**: 需要认证

### 2.4 更新主机
- **接口**: `PUT /hosts/:id`
- **描述**: 更新主机信息
- **权限**: 需要认证

### 2.5 删除主机
- **接口**: `DELETE /hosts/:id`
- **描述**: 删除指定主机
- **权限**: 需要认证

### 2.6 检查主机状态
- **接口**: `POST /hosts/:id/check`
- **描述**: 检查指定主机连接状态
- **权限**: 需要认证

**响应示例**:
```json
{
  "host_id": 1,
  "status": "online",
  "message": "连接测试成功，延迟: 15ms"
}
```

### 2.7 测试SSH连接
- **接口**: `POST /hosts/:id/test-ssh`
- **描述**: 测试主机SSH连接
- **权限**: 需要认证

### 2.8 批量检查主机状态
- **接口**: `POST /hosts/check-all`
- **描述**: 批量检查所有主机状态
- **权限**: 需要认证

### 2.9 更新主机认证信息
- **接口**: `PUT /hosts/:id/auth`
- **描述**: 更新主机SSH认证信息
- **权限**: 需要认证

**请求参数**:
```json
{
  "auth_type": "key",
  "private_key": "-----BEGIN RSA PRIVATE KEY-----\n...",
  "passphrase": "keypassword"
}
```

### 2.10 批量导入主机
- **接口**: `POST /hosts/batch/import`
- **描述**: 批量导入主机
- **权限**: 需要认证

**请求参数**:
```json
{
  "hosts": [
    {
      "name": "服务器1",
      "ip": "192.168.1.100",
      "port": 22,
      "os": "Ubuntu",
      "auth_type": "password",
      "username": "root",
      "password": "password"
    }
  ]
}
```

### 2.11 CSV批量导入主机
- **接口**: `POST /hosts/batch/import-csv`
- **描述**: 通过CSV文件批量导入主机
- **权限**: 需要认证
- **请求类型**: multipart/form-data

### 2.12 下载CSV模板
- **接口**: `GET /hosts/csv-template`
- **描述**: 下载主机导入CSV模板
- **权限**: 需要认证

### 2.13 批量主机操作
- **接口**: `POST /hosts/batch/operation`
- **描述**: 批量执行主机操作
- **权限**: 需要认证

**请求参数**:
```json
{
  "host_ids": [1, 2, 3],
  "operation": "test_connection"
}
```

### 2.14 获取定时检查配置
- **接口**: `GET /hosts/schedule/config`
- **描述**: 获取主机定时检查配置
- **权限**: 需要认证

### 2.15 更新定时检查配置
- **接口**: `PUT /hosts/schedule/config`
- **描述**: 更新主机定时检查配置
- **权限**: 需要认证

---

## 3. 脚本管理 (Script Management)

### 3.1 获取脚本列表
- **接口**: `GET /scripts`
- **描述**: 获取所有脚本列表
- **权限**: 需要认证

**响应示例**:
```json
[
  {
    "id": 1,
    "name": "系统信息检查",
    "content": "#!/bin/bash\nuname -a\ndf -h",
    "type": "shell",
    "description": "获取系统基本信息",
    "created_by": 1,
    "user": {
      "id": 1,
      "username": "admin"
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

### 3.2 创建脚本
- **接口**: `POST /scripts`
- **描述**: 创建新脚本
- **权限**: 需要认证

**请求参数**:
```json
{
  "name": "磁盘清理脚本",
  "content": "#!/bin/bash\nfind /tmp -type f -atime +7 -delete",
  "type": "shell",
  "description": "清理7天前的临时文件"
}
```

### 3.3 获取单个脚本
- **接口**: `GET /scripts/:id`
- **描述**: 获取指定脚本详细信息
- **权限**: 需要认证

### 3.4 更新脚本
- **接口**: `PUT /scripts/:id`
- **描述**: 更新脚本信息
- **权限**: 需要认证（仅创建者或管理员）

### 3.5 删除脚本
- **接口**: `DELETE /scripts/:id`
- **描述**: 删除指定脚本
- **权限**: 需要认证（仅创建者或管理员）

---

## 4. 作业管理 (Job Management)

### 4.1 获取作业列表
- **接口**: `GET /jobs`
- **描述**: 获取所有作业列表
- **权限**: 需要认证

### 4.2 创建作业
- **接口**: `POST /jobs`
- **描述**: 创建新作业
- **权限**: 需要认证

**请求参数**:
```json
{
  "name": "系统巡检作业",
  "script_id": 1,
  "host_ids": "[1,2,3]"
}
```

### 4.3 获取单个作业
- **接口**: `GET /jobs/:id`
- **描述**: 获取指定作业详细信息
- **权限**: 需要认证

### 4.4 更新作业
- **接口**: `PUT /jobs/:id`
- **描述**: 更新作业信息
- **权限**: 需要认证

### 4.5 删除作业
- **接口**: `DELETE /jobs/:id`
- **描述**: 删除指定作业
- **权限**: 需要认证

### 4.6 执行作业
- **接口**: `POST /jobs/:id/execute`
- **描述**: 执行指定作业
- **权限**: 需要认证

### 4.7 获取作业执行记录
- **接口**: `GET /jobs/:id/executions`
- **描述**: 获取指定作业的执行记录
- **权限**: 需要认证

### 4.8 快速执行脚本
- **接口**: `POST /scripts/quick-execute`
- **描述**: 快速执行脚本（不创建作业）
- **权限**: 需要认证

**请求参数**:
```json
{
  "script_id": 1,
  "host_ids": [1, 2, 3],
  "name": "临时执行任务"
}
```

---

## 5. 执行记录管理 (Execution Management)

### 5.1 获取所有执行记录
- **接口**: `GET /executions`
- **描述**: 获取所有作业执行记录
- **权限**: 需要认证

**响应示例**:
```json
[
  {
    "id": 1,
    "job_id": 1,
    "host_id": 1,
    "status": "completed",
    "output": "Linux server 5.4.0-42-generic #46-Ubuntu",
    "error": "",
    "start_time": "2024-01-01T10:00:00Z",
    "end_time": "2024-01-01T10:00:05Z",
    "executed_by": 1,
    "job_name": "系统巡检",
    "script_name": "系统信息检查",
    "is_quick_exec": false,
    "created_at": "2024-01-01T10:00:00Z"
  }
]
```

### 5.2 获取执行记录详情
- **接口**: `GET /executions/:id`
- **描述**: 获取指定执行记录详细信息
- **权限**: 需要认证

---

## 6. 仪表盘统计 (Dashboard)

### 6.1 获取仪表盘统计数据
- **接口**: `GET /dashboard/stats`
- **描述**: 获取仪表盘统计信息
- **权限**: 需要认证

**响应示例**:
```json
{
  "total_hosts": 10,
  "online_hosts": 8,
  "total_jobs": 5,
  "running_jobs": 2,
  "total_scripts": 15
}
```

### 6.2 获取最近活动
- **接口**: `GET /dashboard/recent-activities`
- **描述**: 获取最近的用户活动记录
- **权限**: 需要认证
- **查询参数**: `limit` (默认10)

### 6.3 获取所有活动
- **接口**: `GET /dashboard/activities`
- **描述**: 获取所有活动记录（支持分页和筛选）
- **权限**: 需要认证

**查询参数**:
- `page`: 页码（默认1）
- `size`: 每页数量（默认20）
- `type`: 活动类型筛选（success/error/info）
- `start_date`: 开始日期
- `end_date`: 结束日期
- `keyword`: 关键词搜索

### 6.4 获取作业趋势
- **接口**: `GET /dashboard/job-trend`
- **描述**: 获取作业执行趋势数据
- **权限**: 需要认证
- **查询参数**: `days` (默认7天)

**响应示例**:
```json
[
  {
    "date": "01-15",
    "success": 10,
    "failed": 2
  },
  {
    "date": "01-16",
    "success": 8,
    "failed": 1
  }
]
```

### 6.5 获取主机状态分布
- **接口**: `GET /dashboard/host-status`
- **描述**: 获取主机状态分布统计
- **权限**: 需要认证

**响应示例**:
```json
[
  {
    "status": "online",
    "count": 8
  },
  {
    "status": "offline",
    "count": 2
  }
]
```

---

## 7. 拓扑管理 (Topology Management)

### 7.1 获取拓扑树
- **接口**: `GET /topology/tree`
- **描述**: 获取完整的拓扑结构树
- **权限**: 需要认证

**响应示例**:
```json
{
  "data": [
    {
      "id": 1,
      "unique_id": "business-1",
      "name": "电商业务",
      "code": "ECOM001",
      "type": "business",
      "children": [
        {
          "id": 1,
          "unique_id": "environment-1",
          "name": "生产环境",
          "code": "PROD001",
          "type": "environment",
          "parent_id": 1,
          "children": [
            {
              "id": 1,
              "unique_id": "cluster-1",
              "name": "Web集群",
              "code": "WEB001",
              "type": "cluster",
              "parent_id": 1,
              "stats": {
                "total_hosts": 3,
                "online_hosts": 2,
                "offline_hosts": 1
              },
              "children": [
                {
                  "id": 1,
                  "unique_id": "host-1",
                  "name": "Web服务器1",
                  "type": "host",
                  "parent_id": 1,
                  "host_info": {
                    "id": 1,
                    "name": "Web服务器1",
                    "ip": "192.168.1.100",
                    "status": "online"
                  }
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
```

### 7.2 业务管理

#### 7.2.1 获取业务列表
- **接口**: `GET /topology/businesses`
- **描述**: 获取所有业务列表
- **权限**: 需要认证

#### 7.2.2 创建业务
- **接口**: `POST /topology/businesses`
- **描述**: 创建新业务
- **权限**: 需要认证

**请求参数**:
```json
{
  "name": "电商业务",
  "description": "在线电商平台",
  "owner": "张三"
}
```

#### 7.2.3 更新业务
- **接口**: `PUT /topology/businesses/:id`
- **描述**: 更新业务信息
- **权限**: 需要认证

#### 7.2.4 删除业务
- **接口**: `DELETE /topology/businesses/:id`
- **描述**: 删除业务（需要先删除下属环境）
- **权限**: 需要认证

### 7.3 环境管理

#### 7.3.1 获取环境列表
- **接口**: `GET /topology/environments`
- **描述**: 获取环境列表
- **权限**: 需要认证
- **查询参数**: `business_id` (可选，筛选指定业务下的环境)

#### 7.3.2 创建环境
- **接口**: `POST /topology/environments`
- **描述**: 创建新环境
- **权限**: 需要认证

**请求参数**:
```json
{
  "name": "生产环境",
  "business_id": 1,
  "description": "生产环境服务器"
}
```

#### 7.3.3 更新环境
- **接口**: `PUT /topology/environments/:id`
- **描述**: 更新环境信息
- **权限**: 需要认证

#### 7.3.4 删除环境
- **接口**: `DELETE /topology/environments/:id`
- **描述**: 删除环境（需要先删除下属集群）
- **权限**: 需要认证

### 7.4 集群管理

#### 7.4.1 获取集群列表
- **接口**: `GET /topology/clusters`
- **描述**: 获取集群列表
- **权限**: 需要认证
- **查询参数**: `environment_id` (可选，筛选指定环境下的集群)

#### 7.4.2 创建集群
- **接口**: `POST /topology/clusters`
- **描述**: 创建新集群
- **权限**: 需要认证

**请求参数**:
```json
{
  "name": "Web集群",
  "environment_id": 1,
  "description": "前端Web服务集群"
}
```

#### 7.4.3 更新集群
- **接口**: `PUT /topology/clusters/:id`
- **描述**: 更新集群信息
- **权限**: 需要认证

#### 7.4.4 删除集群
- **接口**: `DELETE /topology/clusters/:id`
- **描述**: 删除集群（需要先移除关联主机）
- **权限**: 需要认证

### 7.5 主机拓扑管理

#### 7.5.1 分配主机到集群
- **接口**: `POST /topology/hosts/assign`
- **描述**: 将主机分配到指定集群
- **权限**: 需要认证

**请求参数**:
```json
{
  "host_id": 1,
  "cluster_id": 1
}
```

#### 7.5.2 从集群移除主机
- **接口**: `DELETE /topology/hosts/:hostId/remove`
- **描述**: 从集群中移除主机
- **权限**: 需要认证

#### 7.5.3 获取未分配主机
- **接口**: `GET /topology/hosts/unassigned`
- **描述**: 获取未分配到任何集群的主机列表
- **权限**: 需要认证

#### 7.5.4 获取集群下的主机
- **接口**: `GET /topology/clusters/:clusterId/hosts`
- **描述**: 获取指定集群下的所有主机
- **权限**: 需要认证

---

## 8. 系统管理 (System Management)

### 8.1 获取系统信息
- **接口**: `GET /system/info`
- **描述**: 获取系统基本信息
- **权限**: 公开接口

### 8.2 健康检查
- **接口**: `GET /system/health`
- **描述**: 系统健康检查
- **权限**: 公开接口

---

## 9. 管理员接口 (Admin Only)

### 9.1 获取用户列表
- **接口**: `GET /admin/users`
- **描述**: 获取所有用户列表
- **权限**: 管理员

### 9.2 删除用户
- **接口**: `DELETE /admin/users/:id`
- **描述**: 删除指定用户
- **权限**: 管理员

### 9.3 更新用户角色
- **接口**: `PUT /admin/users/:id/role`
- **描述**: 更新用户角色
- **权限**: 管理员

**请求参数**:
```json
{
  "role": "admin"
}
```

---

## 错误码说明

| HTTP状态码 | 说明 |
|-----------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未认证或Token无效 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 500 | 服务器内部错误 |

## 通用错误响应格式

```json
{
  "error": "错误描述信息"
}
```

## 通用成功响应格式

```json
{
  "message": "操作成功",
  "data": {}
}
```

---

## 数据模型说明

### 用户模型 (User)
```json
{
  "id": 1,
  "username": "admin",
  "email": "admin@example.com",
  "role": "admin",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 主机模型 (Host)
```json
{
  "id": 1,
  "name": "Web服务器1",
  "ip": "192.168.1.100",
  "port": 22,
  "os": "Ubuntu 20.04",
  "status": "online",
  "description": "前端Web服务器",
  "tags": "web,frontend",
  "auth_type": "password",
  "username": "root",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 脚本模型 (Script)
```json
{
  "id": 1,
  "name": "系统信息检查",
  "content": "#!/bin/bash\nuname -a",
  "type": "shell",
  "description": "获取系统基本信息",
  "created_by": 1,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 作业模型 (Job)
```json
{
  "id": 1,
  "name": "系统巡检作业",
  "script_id": 1,
  "host_ids": "[1,2,3]",
  "status": "pending",
  "is_temporary": false,
  "created_by": 1,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 执行记录模型 (JobExecution)
```json
{
  "id": 1,
  "job_id": 1,
  "host_id": 1,
  "status": "completed",
  "output": "执行输出内容",
  "error": "",
  "start_time": "2024-01-01T10:00:00Z",
  "end_time": "2024-01-01T10:00:05Z",
  "executed_by": 1,
  "job_name": "系统巡检",
  "script_name": "系统信息检查",
  "script_content": "#!/bin/bash\nuname -a",
  "script_type": "shell",
  "is_quick_exec": false,
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

---

## 8. 文件管理 (File Management)

### 8.1 文件上传
- **接口**: `POST /files/upload`
- **描述**: 上传文件到系统
- **权限**: 需要认证
- **请求类型**: multipart/form-data

**请求参数**:
```
file: 文件对象 (必填)
category: 文件分类 (可选，默认为general)
description: 文件描述 (可选)
is_public: 是否公开 (可选，默认为false)
```

**响应示例**:
```json
{
  "message": "文件上传成功",
  "file": {
    "id": 1,
    "name": "20240822_180207_script.sh",
    "original_name": "script.sh",
    "path": "uploads/general/20240822_180207_script.sh",
    "size": 1024,
    "mime_type": "text/plain",
    "md5_hash": "d41d8cd98f00b204e9800998ecf8427e",
    "category": "general",
    "description": "系统脚本",
    "is_public": false,
    "uploaded_by": 1,
    "download_count": 0,
    "created_at": "2024-08-22T18:02:07Z",
    "updated_at": "2024-08-22T18:02:07Z"
  }
}
```

### 8.2 文件下载
- **接口**: `GET /files/{id}/download`
- **描述**: 下载指定文件
- **权限**: 需要认证，文件所有者或管理员或公开文件

**响应**: 文件流

### 8.3 获取文件列表
- **接口**: `GET /files`
- **描述**: 获取文件列表
- **权限**: 需要认证

**查询参数**:
- `page`: 页码 (默认1)
- `size`: 每页数量 (默认20，最大100)
- `category`: 文件分类过滤
- `name`: 文件名搜索

**响应示例**:
```json
{
  "data": [
    {
      "id": 1,
      "name": "20240822_180207_script.sh",
      "original_name": "script.sh",
      "size": 1024,
      "category": "general",
      "is_public": false,
      "download_count": 5,
      "created_at": "2024-08-22T18:02:07Z"
    }
  ],
  "total": 1,
  "page": 1,
  "size": 20,
  "total_pages": 1
}
```

### 8.4 获取文件详情
- **接口**: `GET /files/{id}`
- **描述**: 获取指定文件详细信息
- **权限**: 需要认证，文件所有者或管理员或公开文件

### 8.5 更新文件信息
- **接口**: `PUT /files/{id}`
- **描述**: 更新文件元信息
- **权限**: 需要认证，文件所有者或管理员

**请求参数**:
```json
{
  "category": "scripts",
  "description": "更新后的描述",
  "is_public": true
}
```

### 8.6 删除文件
- **接口**: `DELETE /files/{id}`
- **描述**: 删除指定文件
- **权限**: 需要认证，文件所有者或管理员

### 8.7 文件分发
- **接口**: `POST /files/{id}/distribute`
- **描述**: 将文件分发到指定主机
- **权限**: 需要认证，文件所有者或管理员

**请求参数**:
```json
{
  "host_ids": [1, 2, 3],
  "target_path": "/tmp/script.sh",
  "description": "脚本分发任务"
}
```

**响应示例**:
```json
{
  "message": "文件分发任务创建成功",
  "distribution": {
    "id": 1,
    "file_id": 1,
    "target_path": "/tmp/script.sh",
    "description": "脚本分发任务",
    "status": "pending",
    "progress": 0,
    "created_by": 1,
    "created_at": "2024-08-22T18:02:07Z"
  }
}
```

### 8.8 获取分发记录
- **接口**: `GET /file-distributions`
- **描述**: 获取文件分发记录列表
- **权限**: 需要认证

**查询参数**:
- `page`: 页码 (默认1)
- `size`: 每页数量 (默认20，最大100)
- `status`: 状态过滤 (pending/running/completed/failed/partial)

### 8.9 获取分发详情
- **接口**: `GET /file-distributions/{id}`
- **描述**: 获取指定分发任务的详细信息
- **权限**: 需要认证，任务创建者或管理员

**响应示例**:
```json
{
  "distribution": {
    "id": 1,
    "file_id": 1,
    "target_path": "/tmp/script.sh",
    "status": "completed",
    "progress": 100,
    "start_time": "2024-08-22T18:02:07Z",
    "end_time": "2024-08-22T18:02:30Z"
  },
  "details": [
    {
      "id": 1,
      "distribution_id": 1,
      "host_id": 1,
      "status": "completed",
      "output": "文件传输成功",
      "start_time": "2024-08-22T18:02:07Z",
      "end_time": "2024-08-22T18:02:15Z",
      "host": {
        "id": 1,
        "name": "Web服务器",
        "ip": "192.168.1.100"
      }
    }
  ]
}
```

### 8.10 删除分发记录
- **接口**: `DELETE /file-distributions/{id}`
- **描述**: 删除指定的文件分发记录及其详情
- **权限**: 需要认证，任务创建者或管理员

**响应示例**:
```json
{
  "message": "分发记录删除成功"
}
```

**错误响应**:
- `400`: 无效的分发记录ID
- `403`: 没有权限删除此分发记录
- `404`: 分发记录不存在
- `500`: 删除分发记录失败

---

## 使用示例

### 1. 用户登录获取Token
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "123456"
  }'
```

### 2. 创建主机
```bash
curl -X POST http://localhost:8080/api/v1/hosts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your_token>" \
  -d '{
    "name": "Web服务器",
    "ip": "192.168.1.100",
    "port": 22,
    "os": "Ubuntu 20.04",
    "auth_type": "password",
    "username": "root",
    "password": "password123"
  }'
```

### 3. 执行脚本
```bash
curl -X POST http://localhost:8080/api/v1/scripts/quick-execute \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your_token>" \
  -d '{
    "script_id": 1,
    "host_ids": [1, 2],
    "name": "系统检查"
  }'
```

---

## 更新日志

### v0.2 (2024-08-22)
- 新增文件管理功能
- 支持文件上传、下载、分类管理
- 实现文件分发到远程主机功能
- 增强SSH文件传输能力
- 完善权限控制和活动日志记录

### v0.1 (2024-01-01)
- 初始版本发布
- 实现基础的用户认证、主机管理、脚本管理功能
- 支持作业调度和执行记录管理
- 实现拓扑管理和仪表盘统计功能

---

## 8. 文件管理 (File Management)

### 8.1 获取文件列表
- **接口**: `GET /files`
- **描述**: 获取文件列表，支持分页和筛选
- **权限**: 需要认证

**查询参数**:
- `page`: 页码，默认1
- `page_size`: 每页数量，默认20
- `name`: 文件名筛选
- `category`: 分类筛选
- `isPublic`: 公开状态筛选

**响应示例**:
```json
{
  "data": [
    {
      "id": 1,
      "name": "1756171401824216300_config.txt",
      "original_name": "config.txt",
      "path": "uploads/general/1756171401824216300_config.txt",
      "size": 1024,
      "mime_type": "text/plain",
      "md5_hash": "d41d8cd98f00b204e9800998ecf8427e",
      "category": "general",
      "description": "配置文件",
      "is_public": true,
      "uploaded_by": 1,
      "download_count": 5,
      "created_at": "2025-08-26T10:00:00Z",
      "updated_at": "2025-08-26T10:00:00Z"
    }
  ],
  "total": 50,
  "page": 1,
  "size": 20,
  "total_pages": 3
}
```

### 8.2 上传文件
- **接口**: `POST /files/upload`
- **描述**: 上传文件到系统
- **权限**: 需要认证
- **Content-Type**: `multipart/form-data`

**请求参数**:
- `file`: 文件数据 (form-data)
- `category`: 文件分类，默认"general"
- `description`: 文件描述
- `is_public`: 是否公开，默认false

### 8.3 获取文件详情
- **接口**: `GET /files/:id`
- **描述**: 获取指定文件的详细信息
- **权限**: 需要认证

### 8.4 下载文件
- **接口**: `GET /files/:id/download`
- **描述**: 下载指定文件
- **权限**: 需要认证

### 8.5 更新文件信息
- **接口**: `PUT /files/:id`
- **描述**: 更新文件的元信息（不包括文件内容）
- **权限**: 需要认证

**请求参数**:
```json
{
  "description": "更新后的描述",
  "is_public": true,
  "category": "config"
}
```

### 8.6 删除文件
- **接口**: `DELETE /files/:id`
- **描述**: 删除指定文件
- **权限**: 需要认证

### 8.7 分发文件到主机
- **接口**: `POST /files/:id/distribute`
- **描述**: 将文件分发到指定主机
- **权限**: 需要认证

**请求参数**:
```json
{
  "host_ids": [1, 2, 3],
  "target_path": "/tmp/",
  "description": "配置文件分发"
}
```

### 8.8 获取分发记录
- **接口**: `GET /file-distributions`
- **描述**: 获取文件分发记录列表
- **权限**: 需要认证

**查询参数**:
- `page`: 页码，默认1
- `size`: 每页数量，默认20
- `status`: 状态筛选 (pending, running, completed, failed)

### 8.9 获取分发详情
- **接口**: `GET /file-distributions/:id`
- **描述**: 获取指定分发任务的详细信息
- **权限**: 需要认证

### 8.10 删除分发记录
- **接口**: `DELETE /file-distributions/:id`
- **描述**: 删除指定分发记录
- **权限**: 需要认证

## 9. 脚本文件关联功能

### 9.1 保存执行结果为文件
- **接口**: `POST /job-executions/save-result`
- **描述**: 将作业执行结果保存为文件
- **权限**: 需要认证

**请求参数**:
```json
{
  "execution_id": 123,
  "save_output": true,
  "save_error": true,
  "output_category": "script_output"
}
```

**响应示例**:
```json
{
  "message": "执行结果已保存为文件",
  "execution": {
    "id": 123,
    "output_file_id": 456,
    "error_file_id": 457
  }
}
```

### 9.2 作业文件参数支持
作业创建和更新时支持文件参数：

**请求参数**:
```json
{
  "name": "数据处理作业",
  "script_id": 1,
  "host_ids": "[1,2,3]",
  "input_file_ids": "[10,11,12]",
  "save_output": true,
  "save_error": false,
  "output_category": "script_output",
  "description": "处理数据文件"
}
```

---

## 版本更新日志

### v0.3 (2025-08-26)
- 新增完整的文件管理功能
- 支持脚本文件关联和参数传递
- 实现执行结果文件保存功能
- 增强文件分发和状态追踪
- 完善文件权限控制和MD5去重

### v0.2 (2024-08-22)
- 新增文件管理功能
- 支持文件上传、下载、分类管理
- 实现文件分发到远程主机功能
- 增强SSH文件传输能力
- 完善权限控制和活动日志记录

### v0.1 (2024-01-01)
- 初始版本发布
- 实现基础的用户认证、主机管理、脚本管理功能
- 支持作业调度和执行记录管理
- 实现拓扑管理和仪表盘统计功能

---

*文档最后更新时间: 2025-08-26*
