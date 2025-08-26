/**
 * 格式化文件大小
 * @param {number} bytes 字节数
 * @returns {string} 格式化后的文件大小
 */
export function formatFileSize(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

/**
 * 格式化日期时间
 * @param {string|Date} dateTime 日期时间
 * @returns {string} 格式化后的日期时间
 */
export function formatDateTime(dateTime) {
  if (!dateTime) return '-'
  
  const date = new Date(dateTime)
  if (isNaN(date.getTime())) return '-'
  
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const diffDays = Math.floor(diff / (1000 * 60 * 60 * 24))
  
  // 如果是今天
  if (diffDays === 0) {
    return date.toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })
  }
  
  // 如果是昨天
  if (diffDays === 1) {
    return '昨天 ' + date.toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit'
    })
  }
  
  // 如果是今年
  if (date.getFullYear() === now.getFullYear()) {
    return date.toLocaleDateString('zh-CN', {
      month: '2-digit',
      day: '2-digit'
    }) + ' ' + date.toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit'
    })
  }
  
  // 其他情况显示完整日期
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  }) + ' ' + date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit'
  })
}

/**
 * 格式化相对时间
 * @param {string|Date} dateTime 日期时间
 * @returns {string} 相对时间描述
 */
export function formatRelativeTime(dateTime) {
  if (!dateTime) return '-'
  
  const date = new Date(dateTime)
  if (isNaN(date.getTime())) return '-'
  
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const diffSeconds = Math.floor(diff / 1000)
  const diffMinutes = Math.floor(diffSeconds / 60)
  const diffHours = Math.floor(diffMinutes / 60)
  const diffDays = Math.floor(diffHours / 24)
  
  if (diffSeconds < 60) {
    return '刚刚'
  } else if (diffMinutes < 60) {
    return `${diffMinutes}分钟前`
  } else if (diffHours < 24) {
    return `${diffHours}小时前`
  } else if (diffDays < 7) {
    return `${diffDays}天前`
  } else {
    return formatDateTime(dateTime)
  }
}

/**
 * 格式化持续时间
 * @param {number} seconds 秒数
 * @returns {string} 格式化后的持续时间
 */
export function formatDuration(seconds) {
  if (!seconds || seconds < 0) return '0秒'
  
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const remainingSeconds = seconds % 60
  
  if (hours > 0) {
    return `${hours}时${minutes}分${remainingSeconds}秒`
  } else if (minutes > 0) {
    return `${minutes}分${remainingSeconds}秒`
  } else {
    return `${remainingSeconds}秒`
  }
}

/**
 * 格式化百分比
 * @param {number} value 数值
 * @param {number} total 总数
 * @param {number} decimals 小数位数
 * @returns {string} 百分比字符串
 */
export function formatPercentage(value, total, decimals = 1) {
  if (!total || total === 0) return '0%'
  
  const percentage = (value / total) * 100
  return percentage.toFixed(decimals) + '%'
}

/**
 * 格式化数字，添加千分位分隔符
 * @param {number} num 数字
 * @returns {string} 格式化后的数字
 */
export function formatNumber(num) {
  if (num === null || num === undefined) return '0'
  
  return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

/**
 * 截断文本
 * @param {string} text 文本
 * @param {number} maxLength 最大长度
 * @param {string} suffix 后缀
 * @returns {string} 截断后的文本
 */
export function truncateText(text, maxLength = 50, suffix = '...') {
  if (!text || text.length <= maxLength) return text || ''
  
  return text.substring(0, maxLength) + suffix
}

/**
 * 格式化状态文本
 * @param {string} status 状态
 * @returns {string} 格式化后的状态文本
 */
export function formatStatus(status) {
  const statusMap = {
    'pending': '等待中',
    'running': '运行中',
    'completed': '已完成',
    'failed': '失败',
    'cancelled': '已取消',
    'online': '在线',
    'offline': '离线',
    'active': '活跃',
    'inactive': '非活跃'
  }
  
  return statusMap[status] || status
}
