#!/bin/bash

# 系统健康检查脚本
# 检查CPU、内存、磁盘、网络等系统资源使用情况

echo "========================================="
echo "系统健康检查报告 - $(date)"
echo "========================================="

# 系统基本信息
echo "📊 系统信息:"
echo "主机名: $(hostname)"
echo "操作系统: $(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)"
echo "内核版本: $(uname -r)"
echo "系统运行时间: $(uptime -p)"
echo ""

# CPU使用率
echo "💻 CPU使用情况:"
cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)
echo "CPU使用率: ${cpu_usage}%"
if (( $(echo "$cpu_usage > 80" | bc -l) )); then
    echo "⚠️  警告: CPU使用率过高!"
fi
echo ""

# 内存使用情况
echo "🧠 内存使用情况:"
free -h | grep -E "Mem|Swap"
mem_usage=$(free | grep Mem | awk '{printf("%.1f"), $3/$2 * 100.0}')
echo "内存使用率: ${mem_usage}%"
if (( $(echo "$mem_usage > 85" | bc -l) )); then
    echo "⚠️  警告: 内存使用率过高!"
fi
echo ""

# 磁盘使用情况
echo "💾 磁盘使用情况:"
df -h | grep -vE '^Filesystem|tmpfs|cdrom'
echo ""
echo "磁盘使用率超过80%的分区:"
df -h | awk '$5 > 80 {print "⚠️  " $0}'
echo ""

# 网络连接
echo "🌐 网络连接状态:"
echo "活跃连接数: $(netstat -an | grep ESTABLISHED | wc -l)"
echo "监听端口数: $(netstat -ln | grep LISTEN | wc -l)"
echo ""

# 系统负载
echo "⚖️  系统负载:"
uptime | awk -F'load average:' '{print "负载平均值:" $2}'
echo ""

# 进程信息
echo "🔄 进程信息:"
echo "总进程数: $(ps aux | wc -l)"
echo "运行中进程: $(ps aux | awk '$8 ~ /^R/ {count++} END {print count+0}')"
echo "僵尸进程: $(ps aux | awk '$8 ~ /^Z/ {count++} END {print count+0}')"
echo ""

# 最占用资源的进程
echo "📈 资源占用TOP5进程:"
echo "CPU占用TOP5:"
ps aux --sort=-%cpu | head -6 | tail -5 | awk '{printf "%-20s %s%%\n", $11, $3}'
echo ""
echo "内存占用TOP5:"
ps aux --sort=-%mem | head -6 | tail -5 | awk '{printf "%-20s %s%%\n", $11, $4}'
echo ""

# 检查重要服务状态
echo "🔧 重要服务状态:"
services=("sshd" "nginx" "apache2" "mysql" "docker")
for service in "${services[@]}"; do
    if systemctl is-active --quiet $service 2>/dev/null; then
        echo "✅ $service: 运行中"
    elif systemctl list-unit-files | grep -q "^$service.service"; then
        echo "❌ $service: 已停止"
    fi
done
echo ""

# 最近的系统日志错误
echo "📋 最近的系统错误 (最近10条):"
journalctl -p err -n 10 --no-pager 2>/dev/null || echo "无法访问系统日志"
echo ""

echo "========================================="
echo "检查完成 - $(date)"
echo "========================================="
