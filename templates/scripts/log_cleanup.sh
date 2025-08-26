#!/bin/bash

# 日志清理脚本
# 清理系统中的旧日志文件，释放磁盘空间

echo "========================================="
echo "日志清理脚本 - $(date)"
echo "========================================="

# 配置参数
DAYS_TO_KEEP=7  # 保留最近7天的日志
LOG_DIRS=(
    "/var/log"
    "/var/log/nginx"
    "/var/log/apache2"
    "/var/log/mysql"
    "/tmp"
)

# 显示清理前的磁盘使用情况
echo "🔍 清理前磁盘使用情况:"
df -h /var/log
echo ""

# 计算清理前的总大小
total_before=0
for dir in "${LOG_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        size=$(du -sb "$dir" 2>/dev/null | cut -f1)
        total_before=$((total_before + size))
    fi
done

echo "📊 开始清理日志文件 (保留最近 ${DAYS_TO_KEEP} 天)..."
echo ""

cleaned_files=0
freed_space=0

# 清理各个目录的日志文件
for dir in "${LOG_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        echo "🗂️  处理目录: $dir"
        
        # 查找并删除旧的日志文件
        while IFS= read -r -d '' file; do
            if [ -f "$file" ]; then
                file_size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo 0)
                rm -f "$file"
                if [ $? -eq 0 ]; then
                    echo "  ✅ 删除: $(basename "$file") ($(numfmt --to=iec $file_size))"
                    cleaned_files=$((cleaned_files + 1))
                    freed_space=$((freed_space + file_size))
                fi
            fi
        done < <(find "$dir" -name "*.log.*" -type f -mtime +$DAYS_TO_KEEP -print0 2>/dev/null)
        
        # 清理压缩的日志文件
        while IFS= read -r -d '' file; do
            if [ -f "$file" ]; then
                file_size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo 0)
                rm -f "$file"
                if [ $? -eq 0 ]; then
                    echo "  ✅ 删除: $(basename "$file") ($(numfmt --to=iec $file_size))"
                    cleaned_files=$((cleaned_files + 1))
                    freed_space=$((freed_space + file_size))
                fi
            fi
        done < <(find "$dir" -name "*.gz" -type f -mtime +$DAYS_TO_KEEP -print0 2>/dev/null)
        
        echo ""
    fi
done

# 清理系统日志 (journal)
echo "🗂️  清理系统日志 (journalctl)..."
if command -v journalctl >/dev/null 2>&1; then
    journal_before=$(journalctl --disk-usage 2>/dev/null | grep -o '[0-9.]*[KMGT]B' | head -1)
    journalctl --vacuum-time=${DAYS_TO_KEEP}d >/dev/null 2>&1
    journal_after=$(journalctl --disk-usage 2>/dev/null | grep -o '[0-9.]*[KMGT]B' | head -1)
    echo "  ✅ 系统日志: $journal_before → $journal_after"
fi
echo ""

# 清理临时文件
echo "🗂️  清理临时文件..."
temp_cleaned=0
if [ -d "/tmp" ]; then
    while IFS= read -r -d '' file; do
        if [ -f "$file" ]; then
            file_size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo 0)
            rm -f "$file"
            if [ $? -eq 0 ]; then
                temp_cleaned=$((temp_cleaned + 1))
                freed_space=$((freed_space + file_size))
            fi
        fi
    done < <(find /tmp -name "*.tmp" -o -name "*.temp" -type f -mtime +1 -print0 2>/dev/null)
    echo "  ✅ 清理临时文件: $temp_cleaned 个"
fi
echo ""

# 清理旧的核心转储文件
echo "🗂️  清理核心转储文件..."
core_cleaned=0
while IFS= read -r -d '' file; do
    if [ -f "$file" ]; then
        file_size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo 0)
        rm -f "$file"
        if [ $? -eq 0 ]; then
            core_cleaned=$((core_cleaned + 1))
            freed_space=$((freed_space + file_size))
        fi
    fi
done < <(find /var/crash /var/core /tmp -name "core.*" -type f -mtime +$DAYS_TO_KEEP -print0 2>/dev/null)
echo "  ✅ 清理核心转储文件: $core_cleaned 个"
echo ""

# 显示清理结果
echo "========================================="
echo "📈 清理统计:"
echo "清理文件数量: $cleaned_files"
echo "释放空间: $(numfmt --to=iec $freed_space)"
echo ""

# 显示清理后的磁盘使用情况
echo "🔍 清理后磁盘使用情况:"
df -h /var/log
echo ""

echo "✅ 日志清理完成 - $(date)"
echo "========================================="
