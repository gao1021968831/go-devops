#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
网络诊断脚本
检查网络连通性、DNS解析、端口状态等
"""

import socket
import subprocess
import time
import sys
import json
from datetime import datetime
from concurrent.futures import ThreadPoolExecutor, as_completed

def print_header():
    print("=" * 50)
    print(f"网络诊断脚本 - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print("=" * 50)

def check_ping(host, count=3):
    """检查ping连通性"""
    try:
        result = subprocess.run(
            ['ping', '-c', str(count), host],
            capture_output=True,
            text=True,
            timeout=10
        )
        
        if result.returncode == 0:
            # 解析ping结果
            lines = result.stdout.split('\n')
            stats_line = [line for line in lines if 'packet loss' in line]
            if stats_line:
                loss = stats_line[0].split(',')[2].strip().split('%')[0]
                return {'status': 'success', 'loss': f"{loss}%"}
        
        return {'status': 'failed', 'error': 'ping失败'}
    except Exception as e:
        return {'status': 'error', 'error': str(e)}

def check_dns(domain, dns_server='8.8.8.8'):
    """检查DNS解析"""
    try:
        # 使用nslookup检查DNS解析
        result = subprocess.run(
            ['nslookup', domain, dns_server],
            capture_output=True,
            text=True,
            timeout=5
        )
        
        if result.returncode == 0 and 'Address:' in result.stdout:
            addresses = []
            for line in result.stdout.split('\n'):
                if 'Address:' in line and '#' not in line:
                    addr = line.split('Address:')[1].strip()
                    if addr:
                        addresses.append(addr)
            return {'status': 'success', 'addresses': addresses}
        
        return {'status': 'failed', 'error': 'DNS解析失败'}
    except Exception as e:
        return {'status': 'error', 'error': str(e)}

def check_port(host, port, timeout=3):
    """检查端口连通性"""
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(timeout)
        result = sock.connect_ex((host, port))
        sock.close()
        
        if result == 0:
            return {'status': 'open'}
        else:
            return {'status': 'closed'}
    except Exception as e:
        return {'status': 'error', 'error': str(e)}

def check_http_response(url, timeout=5):
    """检查HTTP响应"""
    try:
        import urllib.request
        import urllib.error
        
        start_time = time.time()
        req = urllib.request.Request(url)
        req.add_header('User-Agent', 'NetworkDiagnostics/1.0')
        
        with urllib.request.urlopen(req, timeout=timeout) as response:
            response_time = (time.time() - start_time) * 1000
            return {
                'status': 'success',
                'code': response.getcode(),
                'response_time': f"{response_time:.2f}ms"
            }
    except urllib.error.HTTPError as e:
        return {'status': 'http_error', 'code': e.code}
    except Exception as e:
        return {'status': 'error', 'error': str(e)}

def get_network_interfaces():
    """获取网络接口信息"""
    try:
        result = subprocess.run(['ip', 'addr', 'show'], capture_output=True, text=True)
        if result.returncode == 0:
            return result.stdout
        return "无法获取网络接口信息"
    except:
        return "ip命令不可用"

def get_routing_table():
    """获取路由表"""
    try:
        result = subprocess.run(['ip', 'route', 'show'], capture_output=True, text=True)
        if result.returncode == 0:
            return result.stdout
        return "无法获取路由表"
    except:
        return "ip命令不可用"

def main():
    print_header()
    
    # 配置检查项目
    ping_hosts = ['8.8.8.8', 'baidu.com', 'google.com']
    dns_domains = ['baidu.com', 'google.com', 'github.com']
    port_checks = [
        ('baidu.com', 80),
        ('baidu.com', 443),
        ('google.com', 80),
        ('google.com', 443)
    ]
    http_urls = ['http://baidu.com', 'https://baidu.com']
    
    print("🌐 开始网络诊断...")
    print()
    
    # 1. 网络接口检查
    print("🔌 网络接口信息:")
    interfaces = get_network_interfaces()
    print(interfaces)
    print()
    
    # 2. 路由表检查
    print("🛣️  路由表信息:")
    routes = get_routing_table()
    print(routes)
    print()
    
    # 3. Ping连通性检查
    print("📡 Ping连通性检查:")
    for host in ping_hosts:
        print(f"  检查主机: {host}")
        result = check_ping(host)
        if result['status'] == 'success':
            print(f"    ✅ 连通正常，丢包率: {result['loss']}")
        else:
            print(f"    ❌ 连通失败: {result.get('error', '未知错误')}")
    print()
    
    # 4. DNS解析检查
    print("🔍 DNS解析检查:")
    for domain in dns_domains:
        print(f"  解析域名: {domain}")
        result = check_dns(domain)
        if result['status'] == 'success':
            print(f"    ✅ 解析成功: {', '.join(result['addresses'])}")
        else:
            print(f"    ❌ 解析失败: {result.get('error', '未知错误')}")
    print()
    
    # 5. 端口连通性检查
    print("🔌 端口连通性检查:")
    with ThreadPoolExecutor(max_workers=5) as executor:
        future_to_port = {
            executor.submit(check_port, host, port): (host, port)
            for host, port in port_checks
        }
        
        for future in as_completed(future_to_port):
            host, port = future_to_port[future]
            try:
                result = future.result()
                print(f"  {host}:{port}")
                if result['status'] == 'open':
                    print(f"    ✅ 端口开放")
                elif result['status'] == 'closed':
                    print(f"    ❌ 端口关闭")
                else:
                    print(f"    ❌ 检查失败: {result.get('error', '未知错误')}")
            except Exception as e:
                print(f"  {host}:{port}")
                print(f"    ❌ 检查异常: {str(e)}")
    print()
    
    # 6. HTTP响应检查
    print("🌍 HTTP响应检查:")
    for url in http_urls:
        print(f"  请求URL: {url}")
        result = check_http_response(url)
        if result['status'] == 'success':
            print(f"    ✅ 响应正常: HTTP {result['code']}, 响应时间: {result['response_time']}")
        elif result['status'] == 'http_error':
            print(f"    ⚠️  HTTP错误: {result['code']}")
        else:
            print(f"    ❌ 请求失败: {result.get('error', '未知错误')}")
    print()
    
    # 7. 网络统计信息
    print("📊 网络统计信息:")
    try:
        # 显示网络连接统计
        netstat_result = subprocess.run(['netstat', '-s'], capture_output=True, text=True)
        if netstat_result.returncode == 0:
            lines = netstat_result.stdout.split('\n')[:20]  # 只显示前20行
            for line in lines:
                if line.strip():
                    print(f"  {line}")
        else:
            print("  无法获取网络统计信息")
    except:
        print("  netstat命令不可用")
    
    print()
    print("=" * 50)
    print(f"网络诊断完成 - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print("=" * 50)

if __name__ == "__main__":
    main()
