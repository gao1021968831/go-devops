#!/bin/bash

# 安全审计脚本
# 检查系统安全配置和潜在的安全风险

echo "========================================="
echo "系统安全审计 - $(date)"
echo "========================================="

audit_issues=0

echo "🔒 开始安全审计检查..."
echo ""

# 检查用户账户安全
echo "👤 用户账户安全检查:"

# 检查空密码账户
empty_passwd=$(awk -F: '($2 == "") {print $1}' /etc/shadow 2>/dev/null)
if [ -n "$empty_passwd" ]; then
    echo "  ❌ 发现空密码账户: $empty_passwd"
    audit_issues=$((audit_issues + 1))
else
    echo "  ✅ 没有空密码账户"
fi

# 检查UID为0的账户
root_accounts=$(awk -F: '($3 == 0) {print $1}' /etc/passwd)
root_count=$(echo "$root_accounts" | wc -l)
if [ $root_count -gt 1 ]; then
    echo "  ⚠️  发现多个UID为0的账户: $root_accounts"
    audit_issues=$((audit_issues + 1))
else
    echo "  ✅ 只有root账户UID为0"
fi

# 检查sudo权限
sudo_users=$(grep -E '^[^#]*ALL=\(ALL\)' /etc/sudoers /etc/sudoers.d/* 2>/dev/null | wc -l)
echo "  ℹ️  sudo权限用户数: $sudo_users"
echo ""

# 检查SSH安全配置
echo "🔑 SSH安全配置检查:"
if [ -f /etc/ssh/sshd_config ]; then
    # 检查root登录
    root_login=$(grep -E "^PermitRootLogin" /etc/ssh/sshd_config | awk '{print $2}')
    if [ "$root_login" = "yes" ]; then
        echo "  ❌ SSH允许root直接登录"
        audit_issues=$((audit_issues + 1))
    else
        echo "  ✅ SSH禁止root直接登录"
    fi
    
    # 检查密码认证
    passwd_auth=$(grep -E "^PasswordAuthentication" /etc/ssh/sshd_config | awk '{print $2}')
    if [ "$passwd_auth" = "yes" ]; then
        echo "  ⚠️  SSH允许密码认证"
    else
        echo "  ✅ SSH禁用密码认证"
    fi
    
    # 检查协议版本
    protocol=$(grep -E "^Protocol" /etc/ssh/sshd_config | awk '{print $2}')
    if [ "$protocol" = "1" ]; then
        echo "  ❌ SSH使用不安全的协议版本1"
        audit_issues=$((audit_issues + 1))
    else
        echo "  ✅ SSH协议版本安全"
    fi
else
    echo "  ⚠️  SSH配置文件不存在"
fi
echo ""

# 检查文件权限
echo "📁 关键文件权限检查:"

# 检查/etc/passwd权限
passwd_perm=$(stat -c %a /etc/passwd 2>/dev/null)
if [ "$passwd_perm" != "644" ]; then
    echo "  ❌ /etc/passwd权限异常: $passwd_perm (应为644)"
    audit_issues=$((audit_issues + 1))
else
    echo "  ✅ /etc/passwd权限正常"
fi

# 检查/etc/shadow权限
shadow_perm=$(stat -c %a /etc/shadow 2>/dev/null)
if [ "$shadow_perm" != "640" ] && [ "$shadow_perm" != "600" ]; then
    echo "  ❌ /etc/shadow权限异常: $shadow_perm (应为640或600)"
    audit_issues=$((audit_issues + 1))
else
    echo "  ✅ /etc/shadow权限正常"
fi

# 检查世界可写文件
world_writable=$(find / -type f -perm -002 2>/dev/null | head -10)
if [ -n "$world_writable" ]; then
    echo "  ⚠️  发现世界可写文件 (显示前10个):"
    echo "$world_writable" | sed 's/^/    /'
    audit_issues=$((audit_issues + 1))
else
    echo "  ✅ 没有发现世界可写文件"
fi
echo ""

# 检查网络安全
echo "🌐 网络安全检查:"

# 检查开放端口
open_ports=$(netstat -tuln | grep LISTEN | wc -l)
echo "  ℹ️  监听端口数: $open_ports"

# 检查危险端口
dangerous_ports=("23" "513" "514" "515" "111")
for port in "${dangerous_ports[@]}"; do
    if netstat -tuln | grep -q ":$port "; then
        echo "  ❌ 发现危险端口开放: $port"
        audit_issues=$((audit_issues + 1))
    fi
done

# 检查防火墙状态
if command -v ufw >/dev/null 2>&1; then
    ufw_status=$(ufw status | grep "Status:" | awk '{print $2}')
    if [ "$ufw_status" = "active" ]; then
        echo "  ✅ UFW防火墙已启用"
    else
        echo "  ⚠️  UFW防火墙未启用"
    fi
elif command -v firewall-cmd >/dev/null 2>&1; then
    if systemctl is-active --quiet firewalld; then
        echo "  ✅ firewalld防火墙已启用"
    else
        echo "  ⚠️  firewalld防火墙未启用"
    fi
else
    echo "  ⚠️  未检测到防火墙"
fi
echo ""

# 检查系统更新
echo "🔄 系统更新检查:"
if command -v apt >/dev/null 2>&1; then
    updates=$(apt list --upgradable 2>/dev/null | grep -c upgradable)
    if [ $updates -gt 0 ]; then
        echo "  ⚠️  有 $updates 个软件包需要更新"
    else
        echo "  ✅ 系统已是最新版本"
    fi
elif command -v yum >/dev/null 2>&1; then
    updates=$(yum check-update 2>/dev/null | grep -c "^[a-zA-Z]")
    if [ $updates -gt 0 ]; then
        echo "  ⚠️  有 $updates 个软件包需要更新"
    else
        echo "  ✅ 系统已是最新版本"
    fi
fi
echo ""

# 检查恶意软件和异常进程
echo "🦠 恶意软件检查:"

# 检查可疑进程
suspicious_processes=$(ps aux | grep -E "(nc|netcat|ncat)" | grep -v grep)
if [ -n "$suspicious_processes" ]; then
    echo "  ⚠️  发现可疑网络工具进程:"
    echo "$suspicious_processes" | sed 's/^/    /'
fi

# 检查异常网络连接
suspicious_connections=$(netstat -an | grep ESTABLISHED | awk '{print $5}' | grep -E "^[0-9]" | sort | uniq -c | sort -nr | head -5)
if [ -n "$suspicious_connections" ]; then
    echo "  ℹ️  连接最多的外部IP (TOP5):"
    echo "$suspicious_connections" | sed 's/^/    /'
fi
echo ""

# 检查日志异常
echo "📋 日志安全检查:"

# 检查登录失败
failed_logins=$(grep "Failed password" /var/log/auth.log 2>/dev/null | tail -10 | wc -l)
if [ $failed_logins -gt 0 ]; then
    echo "  ⚠️  最近10条记录中有 $failed_logins 次登录失败"
fi

# 检查sudo使用
sudo_usage=$(grep "sudo:" /var/log/auth.log 2>/dev/null | tail -5 | wc -l)
if [ $sudo_usage -gt 0 ]; then
    echo "  ℹ️  最近5条sudo使用记录"
fi
echo ""

# 生成安全建议
echo "========================================="
echo "🛡️  安全建议:"

if [ $audit_issues -eq 0 ]; then
    echo "✅ 未发现严重安全问题"
else
    echo "❌ 发现 $audit_issues 个安全问题需要处理"
fi

echo ""
echo "💡 通用安全建议:"
echo "  • 定期更新系统和软件包"
echo "  • 使用强密码策略"
echo "  • 启用防火墙"
echo "  • 定期备份重要数据"
echo "  • 监控系统日志"
echo "  • 限制不必要的服务"
echo "  • 使用SSH密钥认证"
echo ""

echo "🔒 安全审计完成 - $(date)"
echo "========================================="

# 返回适当的退出码
if [ $audit_issues -gt 5 ]; then
    exit 2  # 严重安全问题
elif [ $audit_issues -gt 0 ]; then
    exit 1  # 有安全问题
else
    exit 0  # 安全状态良好
fi
