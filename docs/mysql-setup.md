# MySQL 数据库设置指南

## 1. 连接到 MySQL

```bash
# 使用 root 用户连接
mysql -u root -p
```

## 2. 创建数据库

```sql
-- 创建数据库
CREATE DATABASE devops CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 查看数据库
SHOW DATABASES;
```

## 3. 创建用户并授权

### 方法一：创建用户并授权（推荐）

```sql
-- 创建用户（允许本地连接）
CREATE USER 'devops'@'localhost' IDENTIFIED BY '123456';

-- 创建用户（允许任意IP连接，生产环境不推荐）
CREATE USER 'devops'@'%' IDENTIFIED BY '123456';

-- 授予数据库权限
GRANT ALL PRIVILEGES ON devops.* TO 'devops'@'localhost';

-- 刷新权限
FLUSH PRIVILEGES;
```

### 方法二：一步创建用户并授权

```sql
-- 直接授权（会自动创建用户）
GRANT ALL PRIVILEGES ON devops.* TO 'devops'@'localhost' IDENTIFIED BY '123456';
FLUSH PRIVILEGES;
```

## 4. 验证设置

```sql
-- 查看用户
SELECT User, Host FROM mysql.user WHERE User = 'devops';

-- 查看权限
SHOW GRANTS FOR 'devops'@'localhost';

-- 使用新用户连接测试
-- 退出当前连接
EXIT;
```

```bash
# 使用新用户连接
mysql -u devops -p devops
```

## 5. 生产环境安全设置

### 创建更安全的用户权限

```sql
-- 创建用户
CREATE USER 'devops'@'localhost' IDENTIFIED BY 'StrongPassword123!';

-- 只授予必要权限
GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, INDEX, ALTER ON devops.* TO 'devops'@'localhost';

-- 刷新权限
FLUSH PRIVILEGES;
```

### 修改用户密码

```sql
-- MySQL 5.7+
ALTER USER 'devops'@'localhost' IDENTIFIED BY 'NewPassword123!';

-- 或者使用
SET PASSWORD FOR 'devops'@'localhost' = PASSWORD('NewPassword123!');

FLUSH PRIVILEGES;
```

## 6. 删除用户和数据库（如需要）

```sql
-- 删除用户
DROP USER 'devops'@'localhost';

-- 删除数据库
DROP DATABASE devops;
```

## 7. 配置文件对应设置

创建完数据库和用户后，在 `config/config.yaml` 中配置：

```yaml
database:
  type: "mysql"
  mysql:
    host: "localhost"
    port: 3306
    username: "devops"
    password: "123456"
    database: "devops"
    charset: "utf8mb4"
    parse_time: true
    loc: "Local"
```

## 8. 常见问题解决

### 连接被拒绝
```sql
-- 检查用户是否存在
SELECT User, Host FROM mysql.user WHERE User = 'devops';

-- 检查权限
SHOW GRANTS FOR 'devops'@'localhost';
```

### 字符集问题
```sql
-- 查看数据库字符集
SELECT SCHEMA_NAME, DEFAULT_CHARACTER_SET_NAME, DEFAULT_COLLATION_NAME 
FROM information_schema.SCHEMATA 
WHERE SCHEMA_NAME = 'devops';
```

### 修改现有数据库字符集
```sql
ALTER DATABASE devops CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## 9. 完整的一键设置脚本

```sql
-- 一键设置脚本
CREATE DATABASE IF NOT EXISTS devops CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'devops'@'localhost' IDENTIFIED BY 'DevOps123!';
GRANT ALL PRIVILEGES ON devops.* TO 'devops'@'localhost';
FLUSH PRIVILEGES;

-- 验证
SELECT 'Database created successfully' as Status;
SHOW GRANTS FOR 'devops'@'localhost';
```

## 10. Docker MySQL 快速启动

如果你想使用 Docker 运行 MySQL：

```bash
# 启动 MySQL 容器
docker run --name mysql-devops \
  -e MYSQL_ROOT_PASSWORD=rootpassword \
  -e MYSQL_DATABASE=devops \
  -e MYSQL_USER=devops \
  -e MYSQL_PASSWORD=DevOps123! \
  -p 3306:3306 \
  -d mysql:8.0

# 连接到容器中的 MySQL
docker exec -it mysql-devops mysql -u devops -p devops
```
