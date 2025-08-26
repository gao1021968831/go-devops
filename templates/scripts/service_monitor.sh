#!/bin/bash

# 服务监控脚本
# 监控重要服务的运行状态，自动重启异常服务

echo "========================================="
echo "服务监控脚本 - $(date)"
echo "========================================="

# 配置服务列表 (可通过环境变量覆盖)
SERVICES=${SERVICES:-"nginx,mysql,redis,docker,sshd"}
AUTO_RESTART=${AUTO_RESTART:-"true"}
NOTIFICATION_EMAIL=${NOTIFICATION_EMAIL:-""}

# 将服务字符串转换为数组
IFS=',' read -ra SERVICE_ARRAY <<< "$SERVICES"

echo "🔍 监控服务列表: ${SERVICES}"
echo "🔄 自动重启: ${AUTO_RESTART}"
echo ""

# 发送通知函数
send_notification() {
    local service=$1
    local status=$2
    local action=$3
    
    local message="[$(hostname)] 服务监控报告
时间: $(date)
服务: $service
状态: $status
操作: $action"
    
    if [ -n "$NOTIFICATION_EMAIL" ]; then
        echo "$message" | mail -s "服务监控警报: $service" "$NOTIFICATION_EMAIL" 2>/dev/null
    fi
    
    # 记录到系统日志
    logger "ServiceMonitor: $service - $status - $action"
}

# 检查服务状态函数
check_service() {
    local service=$1
    
    echo "🔍 检查服务: $service"
    
    # 检查服务是否存在
    if ! systemctl list-unit-files | grep -q "^${service}.service"; then
        echo "  ⚠️  服务不存在或未安装"
        return 3
    fi
    
    # 检查服务状态
    if systemctl is-active --quiet "$service"; then
        echo "  ✅ 运行中"
        
        # 检查服务是否启用
        if ! systemctl is-enabled --quiet "$service"; then
            echo "  ⚠️  服务未设置为开机自启"
        fi
        
        return 0
    else
        echo "  ❌ 已停止"
        
        # 尝试自动重启
        if [ "$AUTO_RESTART" = "true" ]; then
            echo "  🔄 尝试重启服务..."
            if systemctl start "$service"; then
                sleep 3
                if systemctl is-active --quiet "$service"; then
                    echo "  ✅ 重启成功"
                    send_notification "$service" "已停止" "自动重启成功"
                    return 1
                else
                    echo "  ❌ 重启失败"
                    send_notification "$service" "已停止" "自动重启失败"
                    return 2
                fi
            else
                echo "  ❌ 重启失败"
                send_notification "$service" "已停止" "自动重启失败"
                return 2
            fi
        else
            send_notification "$service" "已停止" "需要手动处理"
            return 2
        fi
    fi
}

# 统计变量
total_services=0
running_count=0
restarted_count=0
failed_count=0
missing_count=0

# 检查每个服务
for service in "${SERVICE_ARRAY[@]}"; do
    # 去除空格
    service=$(echo "$service" | xargs)
    if [ -n "$service" ]; then
        total_services=$((total_services + 1))
        
        check_service "$service"
        result=$?
        
        case $result in
            0) running_count=$((running_count + 1)) ;;
            1) restarted_count=$((restarted_count + 1)) ;;
            2) failed_count=$((failed_count + 1)) ;;
            3) missing_count=$((missing_count + 1)) ;;
        esac
        
        echo ""
    fi
done

# 检查系统负载
echo "📊 系统负载检查:"
load_avg=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
cpu_count=$(nproc)
load_threshold=$(echo "$cpu_count * 0.8" | bc)

echo "  当前负载: $load_avg"
echo "  CPU核心数: $cpu_count"
echo "  负载阈值: $load_threshold"

if (( $(echo "$load_avg > $load_threshold" | bc -l) )); then
    echo "  ⚠️  系统负载过高!"
    send_notification "系统负载" "过高 ($load_avg)" "需要检查"
else
    echo "  ✅ 系统负载正常"
fi
echo ""

# 检查内存使用
echo "💾 内存使用检查:"
mem_usage=$(free | grep Mem | awk '{printf("%.1f"), $3/$2 * 100.0}')
mem_threshold=85

echo "  内存使用率: ${mem_usage}%"
echo "  警告阈值: ${mem_threshold}%"

if (( $(echo "$mem_usage > $mem_threshold" | bc -l) )); then
    echo "  ⚠️  内存使用率过高!"
    send_notification "内存使用" "过高 (${mem_usage}%)" "需要检查"
else
    echo "  ✅ 内存使用正常"
fi
echo ""

# 检查磁盘空间
echo "💿 磁盘空间检查:"
disk_warning=false
df -h | grep -vE '^Filesystem|tmpfs|cdrom' | awk '{print $5 " " $1}' | while read output; do
    usage=$(echo $output | awk '{print $1}' | sed 's/%//g')
    partition=$(echo $output | awk '{print $2}')
    
    if [ $usage -ge 85 ]; then
        echo "  ⚠️  $partition: ${usage}% (过高)"
        send_notification "磁盘空间" "$partition 使用率过高 (${usage}%)" "需要清理"
        disk_warning=true
    else
        echo "  ✅ $partition: ${usage}%"
    fi
done
echo ""

# 显示统计结果
echo "========================================="
echo "📊 监控统计:"
echo "总服务数: $total_services"
echo "🟢 正常运行: $running_count"
echo "🔄 自动重启: $restarted_count"
echo "❌ 重启失败: $failed_count"
echo "⚠️  服务缺失: $missing_count"
echo ""

# 生成建议
if [ $failed_count -gt 0 ]; then
    echo "🚨 警告: 有 $failed_count 个服务无法启动，需要手动检查!"
elif [ $restarted_count -gt 0 ]; then
    echo "ℹ️  信息: 有 $restarted_count 个服务已自动重启。"
else
    echo "✅ 所有监控服务运行正常。"
fi

echo ""
echo "🔍 服务监控完成 - $(date)"
echo "========================================="

# 返回适当的退出码
if [ $failed_count -gt 0 ]; then
    exit 2  # 有服务启动失败
elif [ $restarted_count -gt 0 ]; then
    exit 1  # 有服务被重启
else
    exit 0  # 所有服务正常
fi
