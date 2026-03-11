// 工具函数模块 - 通用的辅助函数

/**
 * 格式化文件大小
 * @param {number} bytes 字节数
 * @returns {string} 格式化后的文件大小
 */
function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

/**
 * 格式化上传速度
 * @param {number} bytesPerSecond 每秒字节数
 * @returns {string} 格式化后的速度
 */
function formatSpeed(bytesPerSecond) {
    if (bytesPerSecond === 0) return '0 B/s';
    
    const units = ['B/s', 'KB/s', 'MB/s', 'GB/s'];
    let size = bytesPerSecond;
    let unitIndex = 0;
    
    while (size >= 1024 && unitIndex < units.length - 1) {
        size /= 1024;
        unitIndex++;
    }
    
    return size.toFixed(1) + ' ' + units[unitIndex];
}

/**
 * 格式化剩余时间
 * @param {number} seconds 剩余秒数
 * @returns {string} 格式化后的时间
 */
function formatTime(seconds) {
    if (!isFinite(seconds) || seconds <= 0) return '计算中...';
    
    if (seconds < 60) {
        return Math.round(seconds) + '秒';
    } else if (seconds < 3600) {
        const minutes = Math.round(seconds / 60);
        return minutes + '分钟';
    } else {
        const hours = Math.floor(seconds / 3600);
        const minutes = Math.round((seconds % 3600) / 60);
        return hours + '小时' + (minutes > 0 ? minutes + '分钟' : '');
    }
}

/**
 * 复制文本到剪贴板
 * @param {string} text 要复制的文本
 * @param {HTMLElement} button 按钮元素（可选）
 */
function copyToClipboard(text, button = null) {
    const originalText = button ? button.textContent : null;
    
    if (navigator.clipboard && window.isSecureContext) {
        // 现代浏览器的 Clipboard API
        navigator.clipboard.writeText(text).then(() => {
            if (button) {
                button.textContent = '✅ 已复制';
                button.style.background = '#28a745';
                setTimeout(() => {
                    button.textContent = originalText;
                    button.style.background = '#17a2b8';
                }, 2000);
            }
        }).catch(err => {
            console.error('复制失败:', err);
            fallbackCopyTextToClipboard(text, button, originalText);
        });
    } else {
        // 备用方案
        fallbackCopyTextToClipboard(text, button, originalText);
    }
}

/**
 * 备用复制方案
 * @param {string} text 要复制的文本
 * @param {HTMLElement} button 按钮元素
 * @param {string} originalText 原始按钮文本
 */
function fallbackCopyTextToClipboard(text, button, originalText) {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    
    try {
        const successful = document.execCommand('copy');
        if (successful && button) {
            button.textContent = '✅ 已复制';
            button.style.background = '#28a745';
            setTimeout(() => {
                button.textContent = originalText;
                button.style.background = '#17a2b8';
            }, 2000);
        } else if (button) {
            button.textContent = '❌ 复制失败';
            button.style.background = '#dc3545';
            setTimeout(() => {
                button.textContent = originalText;
                button.style.background = '#17a2b8';
            }, 2000);
        }
    } catch (err) {
        console.error('备用复制方案失败:', err);
        if (button) {
            button.textContent = '❌ 复制失败';
            button.style.background = '#dc3545';
            setTimeout(() => {
                button.textContent = originalText;
                button.style.background = '#17a2b8';
            }, 2000);
        }
    }
    
    document.body.removeChild(textArea);
}

/**
 * 自动复制到剪贴板（无UI反馈）
 * @param {string} text 要复制的文本
 */
function copyToClipboardAuto(text) {
    if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(text).then(() => {
            console.log('提取码已自动复制到剪贴板:', text);
            showNotification('提取码已复制: ' + text, 'success');
        }).catch(err => {
            console.error('自动复制失败:', err);
            fallbackCopyTextToClipboardAuto(text);
        });
    } else {
        fallbackCopyTextToClipboardAuto(text);
    }
}

/**
 * 备用自动复制方案
 * @param {string} text 要复制的文本
 */
function fallbackCopyTextToClipboardAuto(text) {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    
    try {
        const successful = document.execCommand('copy');
        if (successful) {
            console.log('提取码已自动复制到剪贴板:', text);
            showNotification('提取码已复制: ' + text, 'success');
        } else {
            console.error('自动复制失败');
        }
    } catch (err) {
        console.error('备用自动复制方案失败:', err);
    }
    
    document.body.removeChild(textArea);
}

/**
 * 显示通知消息
 * @param {string} message 消息内容
 * @param {string} type 消息类型 (success, error, warning, info)
 */
function showNotification(message, type = 'success') {
    // 创建通知元素
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    
    document.body.appendChild(notification);
    
    // 显示动画
    setTimeout(() => {
        notification.classList.add('show');
    }, 100);
    
    // 3秒后自动隐藏
    setTimeout(() => {
        notification.classList.remove('show');
        setTimeout(() => {
            if (notification.parentNode) {
                document.body.removeChild(notification);
            }
        }, 300);
    }, 3000);
}

/**
 * 显示结果
 * @param {string} content HTML内容
 */
function showResult(content) {
    const resultElement = document.getElementById('result-content');
    const resultContainer = document.getElementById('result');
    
    if (resultElement && resultContainer) {
        resultElement.innerHTML = content;
        resultContainer.classList.add('show');
    }
}

/**
 * 隐藏结果
 */
function hideResult() {
    const resultContainer = document.getElementById('result');
    if (resultContainer) {
        resultContainer.classList.remove('show');
    }
}

/**
 * 转义HTML字符
 * @param {string} text 原始文本
 * @returns {string} 转义后的文本
 */
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

/**
 * 防抖函数
 * @param {Function} func 要防抖的函数
 * @param {number} wait 等待时间
 * @returns {Function} 防抖后的函数
 */
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

/**
 * 节流函数
 * @param {Function} func 要节流的函数
 * @param {number} limit 时间限制
 * @returns {Function} 节流后的函数
 */
function throttle(func, limit) {
    let inThrottle;
    return function executedFunction(...args) {
        if (!inThrottle) {
            func.apply(this, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}

/**
 * 检查是否为移动端
 * @returns {boolean} 是否为移动端
 */
function isMobile() {
    return window.innerWidth <= 768 || /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
}

/**
 * 获取文件扩展名
 * @param {string} filename 文件名
 * @returns {string} 文件扩展名
 */
function getFileExtension(filename) {
    return filename.slice((filename.lastIndexOf('.') - 1 >>> 0) + 2);
}

/**
 * 验证文件类型
 * @param {File} file 文件对象
 * @param {Array} allowedTypes 允许的类型数组
 * @returns {boolean} 是否为允许的类型
 */
function validateFileType(file, allowedTypes) {
    return allowedTypes.some(type => {
        if (type.startsWith('.')) {
            return file.name.toLowerCase().endsWith(type.toLowerCase());
        }
        return file.type.startsWith(type);
    });
}

/**
 * 格式化日期时间
 * @param {Date|string} date 日期对象或字符串
 * @returns {string} 格式化后的日期时间
 */
function formatDateTime(date) {
    const d = new Date(date);
    return d.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}