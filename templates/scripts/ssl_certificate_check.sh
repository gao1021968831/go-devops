#!/bin/bash

# SSL证书检查脚本
# 检查网站SSL证书的有效性和过期时间

echo "========================================="
echo "SSL证书检查脚本 - $(date)"
echo "========================================="

# 默认域名列表 (可通过参数传入)
DOMAINS=${1:-"example.com,www.example.com,api.example.com"}
WARNING_DAYS=30  # 证书过期前30天开始警告

# 将域名字符串转换为数组
IFS=',' read -ra DOMAIN_ARRAY <<< "$DOMAINS"

echo "🔐 开始检查SSL证书..."
echo "⚠️  过期警告阈值: $WARNING_DAYS 天"
echo ""

check_ssl_certificate() {
    local domain=$1
    local port=${2:-443}
    
    echo "🌐 检查域名: $domain:$port"
    
    # 获取证书信息
    cert_info=$(echo | timeout 10 openssl s_client -servername "$domain" -connect "$domain:$port" 2>/dev/null | openssl x509 -noout -dates -subject -issuer 2>/dev/null)
    
    if [ $? -ne 0 ] || [ -z "$cert_info" ]; then
        echo "  ❌ 无法获取证书信息"
        return 1
    fi
    
    # 解析证书信息
    not_before=$(echo "$cert_info" | grep "notBefore" | cut -d= -f2)
    not_after=$(echo "$cert_info" | grep "notAfter" | cut -d= -f2)
    subject=$(echo "$cert_info" | grep "subject" | cut -d= -f2-)
    issuer=$(echo "$cert_info" | grep "issuer" | cut -d= -f2-)
    
    # 转换日期格式
    expiry_date=$(date -d "$not_after" +%Y-%m-%d 2>/dev/null || date -j -f "%b %d %H:%M:%S %Y %Z" "$not_after" +%Y-%m-%d 2>/dev/null)
    expiry_timestamp=$(date -d "$not_after" +%s 2>/dev/null || date -j -f "%b %d %H:%M:%S %Y %Z" "$not_after" +%s 2>/dev/null)
    current_timestamp=$(date +%s)
    
    # 计算剩余天数
    days_left=$(( (expiry_timestamp - current_timestamp) / 86400 ))
    
    echo "  📋 证书主体: $subject"
    echo "  🏢 颁发机构: $issuer"
    echo "  📅 有效期至: $expiry_date"
    echo "  ⏰ 剩余天数: $days_left 天"
    
    # 状态判断
    if [ $days_left -lt 0 ]; then
        echo "  🔴 状态: 已过期"
        return 2
    elif [ $days_left -lt $WARNING_DAYS ]; then
        echo "  🟡 状态: 即将过期 (警告)"
        return 1
    else
        echo "  🟢 状态: 正常"
        return 0
    fi
}

# 统计变量
total_domains=0
normal_count=0
warning_count=0
expired_count=0
error_count=0

# 检查每个域名
for domain in "${DOMAIN_ARRAY[@]}"; do
    # 去除空格
    domain=$(echo "$domain" | xargs)
    if [ -n "$domain" ]; then
        total_domains=$((total_domains + 1))
        
        check_ssl_certificate "$domain"
        result=$?
        
        case $result in
            0) normal_count=$((normal_count + 1)) ;;
            1) warning_count=$((warning_count + 1)) ;;
            2) expired_count=$((expired_count + 1)) ;;
            *) error_count=$((error_count + 1)) ;;
        esac
        
        echo ""
    fi
done

# 显示统计结果
echo "========================================="
echo "📊 检查统计:"
echo "总域名数: $total_domains"
echo "🟢 正常: $normal_count"
echo "🟡 警告: $warning_count"
echo "🔴 过期: $expired_count"
echo "❌ 错误: $error_count"
echo ""

# 生成建议
if [ $expired_count -gt 0 ]; then
    echo "🚨 紧急: 有 $expired_count 个域名的证书已过期，需要立即更新!"
elif [ $warning_count -gt 0 ]; then
    echo "⚠️  注意: 有 $warning_count 个域名的证书即将过期，建议尽快更新。"
else
    echo "✅ 所有证书状态正常。"
fi

echo ""
echo "🔐 SSL证书检查完成 - $(date)"
echo "========================================="

# 返回适当的退出码
if [ $expired_count -gt 0 ]; then
    exit 2  # 有过期证书
elif [ $warning_count -gt 0 ]; then
    exit 1  # 有即将过期的证书
else
    exit 0  # 所有证书正常
fi
