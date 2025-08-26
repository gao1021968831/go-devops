#!/bin/bash

# Docker清理脚本
# 清理无用的Docker镜像、容器、网络和卷

echo "========================================="
echo "Docker清理脚本 - $(date)"
echo "========================================="

# 检查Docker是否安装和运行
if ! command -v docker >/dev/null 2>&1; then
    echo "❌ Docker未安装"
    exit 1
fi

if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker服务未运行"
    exit 1
fi

echo "🐳 Docker清理开始..."
echo ""

# 显示清理前的状态
echo "🔍 清理前Docker资源使用情况:"
docker system df
echo ""

# 停止所有已退出的容器
echo "🛑 清理已停止的容器..."
stopped_containers=$(docker ps -aq --filter "status=exited")
if [ -n "$stopped_containers" ]; then
    docker rm $stopped_containers
    echo "  ✅ 已清理 $(echo $stopped_containers | wc -w) 个已停止的容器"
else
    echo "  ℹ️  没有需要清理的已停止容器"
fi
echo ""

# 清理悬空镜像
echo "🖼️  清理悬空镜像..."
dangling_images=$(docker images -f "dangling=true" -q)
if [ -n "$dangling_images" ]; then
    docker rmi $dangling_images
    echo "  ✅ 已清理 $(echo $dangling_images | wc -w) 个悬空镜像"
else
    echo "  ℹ️  没有悬空镜像需要清理"
fi
echo ""

# 清理无用的网络
echo "🌐 清理无用的网络..."
unused_networks=$(docker network ls --filter "dangling=true" -q)
if [ -n "$unused_networks" ]; then
    docker network rm $unused_networks 2>/dev/null
    echo "  ✅ 已清理 $(echo $unused_networks | wc -w) 个无用网络"
else
    echo "  ℹ️  没有无用网络需要清理"
fi
echo ""

# 清理无用的卷
echo "💾 清理无用的卷..."
unused_volumes=$(docker volume ls --filter "dangling=true" -q)
if [ -n "$unused_volumes" ]; then
    docker volume rm $unused_volumes
    echo "  ✅ 已清理 $(echo $unused_volumes | wc -w) 个无用卷"
else
    echo "  ℹ️  没有无用卷需要清理"
fi
echo ""

# 清理构建缓存
echo "🔧 清理构建缓存..."
docker builder prune -f >/dev/null 2>&1
echo "  ✅ 构建缓存已清理"
echo ""

# 清理超过30天的镜像
echo "🗓️  清理30天前的未使用镜像..."
old_images=$(docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.ID}}\t{{.CreatedAt}}" | awk 'NR>1' | while read repo tag id created; do
    # 检查镜像是否超过30天且未被使用
    if [ "$(docker ps -a --filter ancestor=$id -q)" = "" ]; then
        created_timestamp=$(date -d "$created" +%s 2>/dev/null || echo 0)
        current_timestamp=$(date +%s)
        age_days=$(( (current_timestamp - created_timestamp) / 86400 ))
        if [ $age_days -gt 30 ]; then
            echo $id
        fi
    fi
done)

if [ -n "$old_images" ]; then
    echo "$old_images" | xargs docker rmi -f 2>/dev/null
    echo "  ✅ 已清理旧镜像"
else
    echo "  ℹ️  没有需要清理的旧镜像"
fi
echo ""

# 系统级清理
echo "🧹 执行系统级清理..."
docker system prune -f >/dev/null 2>&1
echo "  ✅ 系统级清理完成"
echo ""

# 显示清理后的状态
echo "🔍 清理后Docker资源使用情况:"
docker system df
echo ""

# 显示当前运行的容器
echo "📊 当前运行的容器:"
running_containers=$(docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}")
if [ "$(docker ps -q)" ]; then
    echo "$running_containers"
else
    echo "  ℹ️  没有运行中的容器"
fi
echo ""

echo "✅ Docker清理完成 - $(date)"
echo "========================================="
