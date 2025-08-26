#!/bin/bash

# 数据库备份脚本
# 支持MySQL和PostgreSQL数据库备份

echo "========================================="
echo "数据库备份脚本 - $(date)"
echo "========================================="

# 配置参数
BACKUP_DIR="/backup/database"
RETENTION_DAYS=30
DATE=$(date +%Y%m%d_%H%M%S)

# 数据库配置 (可通过环境变量覆盖)
DB_TYPE=${DB_TYPE:-"mysql"}  # mysql 或 postgresql
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"3306"}
DB_USER=${DB_USER:-"backup_user"}
DB_PASSWORD=${DB_PASSWORD:-""}
DB_NAME=${DB_NAME:-""}

# 创建备份目录
mkdir -p "$BACKUP_DIR"
if [ $? -ne 0 ]; then
    echo "❌ 无法创建备份目录: $BACKUP_DIR"
    exit 1
fi

echo "📁 备份目录: $BACKUP_DIR"
echo "🗄️  数据库类型: $DB_TYPE"
echo "🖥️  数据库主机: $DB_HOST:$DB_PORT"
echo ""

# 检查磁盘空间
available_space=$(df "$BACKUP_DIR" | tail -1 | awk '{print $4}')
if [ "$available_space" -lt 1048576 ]; then  # 小于1GB
    echo "⚠️  警告: 备份目录可用空间不足1GB"
fi

backup_success=0
total_databases=0

# MySQL备份函数
backup_mysql() {
    local db_name=$1
    local backup_file="$BACKUP_DIR/mysql_${db_name}_${DATE}.sql"
    
    echo "🔄 备份MySQL数据库: $db_name"
    
    if [ -n "$DB_PASSWORD" ]; then
        mysqldump -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" \
                  --single-transaction --routines --triggers "$db_name" > "$backup_file"
    else
        mysqldump -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" \
                  --single-transaction --routines --triggers "$db_name" > "$backup_file"
    fi
    
    if [ $? -eq 0 ] && [ -s "$backup_file" ]; then
        # 压缩备份文件
        gzip "$backup_file"
        backup_file="${backup_file}.gz"
        
        file_size=$(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null)
        echo "  ✅ 备份成功: $(basename "$backup_file") ($(numfmt --to=iec $file_size))"
        backup_success=$((backup_success + 1))
    else
        echo "  ❌ 备份失败: $db_name"
        rm -f "$backup_file" "${backup_file}.gz"
    fi
}

# PostgreSQL备份函数
backup_postgresql() {
    local db_name=$1
    local backup_file="$BACKUP_DIR/postgresql_${db_name}_${DATE}.sql"
    
    echo "🔄 备份PostgreSQL数据库: $db_name"
    
    export PGPASSWORD="$DB_PASSWORD"
    pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$db_name" \
            --no-password --verbose > "$backup_file" 2>/dev/null
    
    if [ $? -eq 0 ] && [ -s "$backup_file" ]; then
        # 压缩备份文件
        gzip "$backup_file"
        backup_file="${backup_file}.gz"
        
        file_size=$(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null)
        echo "  ✅ 备份成功: $(basename "$backup_file") ($(numfmt --to=iec $file_size))"
        backup_success=$((backup_success + 1))
    else
        echo "  ❌ 备份失败: $db_name"
        rm -f "$backup_file" "${backup_file}.gz"
    fi
    unset PGPASSWORD
}

# 获取数据库列表并备份
if [ "$DB_TYPE" = "mysql" ]; then
    echo "🔍 获取MySQL数据库列表..."
    
    if [ -n "$DB_NAME" ]; then
        # 备份指定数据库
        databases=("$DB_NAME")
    else
        # 获取所有数据库
        if [ -n "$DB_PASSWORD" ]; then
            databases=($(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" \
                        -e "SHOW DATABASES;" | grep -Ev "^(Database|information_schema|performance_schema|mysql|sys)$"))
        else
            databases=($(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" \
                        -e "SHOW DATABASES;" | grep -Ev "^(Database|information_schema|performance_schema|mysql|sys)$"))
        fi
    fi
    
    total_databases=${#databases[@]}
    echo "📊 找到 $total_databases 个数据库"
    echo ""
    
    for db in "${databases[@]}"; do
        backup_mysql "$db"
    done
    
elif [ "$DB_TYPE" = "postgresql" ]; then
    echo "🔍 获取PostgreSQL数据库列表..."
    
    if [ -n "$DB_NAME" ]; then
        # 备份指定数据库
        databases=("$DB_NAME")
    else
        # 获取所有数据库
        export PGPASSWORD="$DB_PASSWORD"
        databases=($(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -t -c "SELECT datname FROM pg_database WHERE datistemplate = false;" | grep -v "^$" | tr -d ' '))
        unset PGPASSWORD
    fi
    
    total_databases=${#databases[@]}
    echo "📊 找到 $total_databases 个数据库"
    echo ""
    
    for db in "${databases[@]}"; do
        backup_postgresql "$db"
    done
else
    echo "❌ 不支持的数据库类型: $DB_TYPE"
    exit 1
fi

echo ""

# 清理旧备份
echo "🧹 清理 $RETENTION_DAYS 天前的旧备份..."
old_backups=$(find "$BACKUP_DIR" -name "*.sql.gz" -mtime +$RETENTION_DAYS -type f 2>/dev/null)
if [ -n "$old_backups" ]; then
    echo "$old_backups" | while read -r file; do
        rm -f "$file"
        echo "  🗑️  删除: $(basename "$file")"
    done
else
    echo "  ℹ️  没有需要清理的旧备份"
fi

echo ""

# 显示备份统计
echo "========================================="
echo "📈 备份统计:"
echo "总数据库数: $total_databases"
echo "成功备份数: $backup_success"
echo "失败备份数: $((total_databases - backup_success))"

if [ $backup_success -eq $total_databases ]; then
    echo "✅ 所有数据库备份成功!"
else
    echo "⚠️  部分数据库备份失败!"
fi

echo ""
echo "📁 备份文件位置: $BACKUP_DIR"
echo "🕒 备份完成时间: $(date)"
echo "========================================="

# 返回适当的退出码
if [ $backup_success -eq $total_databases ]; then
    exit 0
else
    exit 1
fi
