// API 请求封装

// 注意：authToken 全局变量已在 main.js 中声明

/**
 * API请求封装
 * @param {string} url - 请求URL
 * @param {Object} options - 请求选项
 * @returns {Promise} 请求结果
 */
async function apiRequest(url, options = {}) {
    // 动态获取当前的authToken（优先检查内存变量，再检查 localStorage 的新键 auth_token）
    const currentAuthToken = window.authToken || localStorage.getItem('auth_token');
    
    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json'
        }
    };
    
    // 如果有token，添加Authorization头
    if (currentAuthToken) {
        defaultOptions.headers['Authorization'] = 'Bearer ' + currentAuthToken;
    }
    
    const finalOptions = {
        ...defaultOptions,
        ...options,
        headers: {
            ...defaultOptions.headers,
            ...options.headers
        }
    };
    
    try {
        const response = await fetch(url, finalOptions);
        
        // 处理认证失败
        if (response.status === 401) {
            handleAuthError();
            throw new Error('认证失败，请重新登录');
        }
        
        // 处理其他HTTP错误
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const result = await response.json();
        return result;
    } catch (error) {
        console.error('API Request Error:', error);
        throw error;
    }
}

/**
 * GET 请求
 * @param {string} url - 请求URL
 * @param {Object} params - 查询参数
 * @returns {Promise} 请求结果
 */
async function apiGet(url, params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const fullUrl = queryString ? `${url}?${queryString}` : url;
    
    return apiRequest(fullUrl, {
        method: 'GET'
    });
}

/**
 * POST 请求
 * @param {string} url - 请求URL
 * @param {Object} data - 请求数据
 * @returns {Promise} 请求结果
 */
async function apiPost(url, data = {}) {
    return apiRequest(url, {
        method: 'POST',
        body: JSON.stringify(data)
    });
}

/**
 * PUT 请求
 * @param {string} url - 请求URL
 * @param {Object} data - 请求数据
 * @returns {Promise} 请求结果
 */
async function apiPut(url, data = {}) {
    return apiRequest(url, {
        method: 'PUT',
        body: JSON.stringify(data)
    });
}

/**
 * DELETE 请求
 * @param {string} url - 请求URL
 * @returns {Promise} 请求结果
 */
async function apiDelete(url) {
    return apiRequest(url, {
        method: 'DELETE'
    });
}

/**
 * 文件上传请求
 * @param {string} url - 请求URL
 * @param {FormData} formData - 表单数据
 * @param {Function} onProgress - 进度回调
 * @returns {Promise} 请求结果
 */
async function apiUpload(url, formData, onProgress = null) {
    return new Promise((resolve, reject) => {
        const currentAuthToken = window.authToken || localStorage.getItem('auth_token');
        const xhr = new XMLHttpRequest();
        
        // 设置上传进度监听
        if (onProgress) {
            xhr.upload.addEventListener('progress', (e) => {
                if (e.lengthComputable) {
                    const percentComplete = (e.loaded / e.total) * 100;
                    onProgress(percentComplete);
                }
            });
        }
        
        // 设置完成监听
        xhr.addEventListener('load', () => {
            if (xhr.status >= 200 && xhr.status < 300) {
                try {
                    const result = JSON.parse(xhr.responseText);
                    resolve(result);
                } catch (error) {
                    reject(new Error('响应解析失败'));
                }
            } else if (xhr.status === 401) {
                handleAuthError();
                reject(new Error('认证失败，请重新登录'));
            } else {
                reject(new Error(`上传失败: ${xhr.status} ${xhr.statusText}`));
            }
        });
        
        // 设置错误监听
        xhr.addEventListener('error', () => {
            reject(new Error('网络错误'));
        });
        
        // 设置超时监听
        xhr.addEventListener('timeout', () => {
            reject(new Error('请求超时'));
        });
        
        // 配置请求
        xhr.open('POST', url);
        if (currentAuthToken) {
            xhr.setRequestHeader('Authorization', 'Bearer ' + currentAuthToken);
        }
        xhr.timeout = 300000; // 5分钟超时
        
        // 发送请求
        xhr.send(formData);
    });
}

/**
 * 文件下载请求
 * @param {string} url - 请求URL
 * @param {string} filename - 文件名
 * @returns {Promise} 下载结果
 */
async function apiDownload(url, filename = 'download') {
    try {
    const currentAuthToken = window.authToken || localStorage.getItem('auth_token');
        const headers = {};
        
        if (currentAuthToken) {
            headers['Authorization'] = 'Bearer ' + currentAuthToken;
        }
        
        const response = await fetch(url, { headers });
        
        if (response.status === 401) {
            handleAuthError();
            throw new Error('认证失败，请重新登录');
        }
        
        if (!response.ok) {
            throw new Error(`下载失败: ${response.status} ${response.statusText}`);
        }
        
        const blob = await response.blob();
        
        // 尝试从响应头获取文件名
        const contentDisposition = response.headers.get('Content-Disposition');
        if (contentDisposition) {
            const matches = contentDisposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/);
            if (matches != null && matches[1]) {
                filename = matches[1].replace(/['"]/g, '');
            }
        }
        
        // 创建下载链接
        const url_obj = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url_obj;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url_obj);
        
        return { success: true, filename };
    } catch (error) {
        console.error('Download Error:', error);
        throw error;
    }
}

/**
 * 批量请求
 * @param {Array} requests - 请求数组
 * @param {number} concurrency - 并发数
 * @returns {Promise} 请求结果数组
 */
async function apiBatch(requests, concurrency = 3) {
    const results = [];
    
    for (let i = 0; i < requests.length; i += concurrency) {
        const batch = requests.slice(i, i + concurrency);
        const batchResults = await Promise.allSettled(batch);
        results.push(...batchResults);
    }
    
    return results;
}

/**
 * 重试请求
 * @param {Function} requestFn - 请求函数
 * @param {number} maxRetries - 最大重试次数
 * @param {number} delay - 重试延迟（毫秒）
 * @returns {Promise} 请求结果
 */
async function apiRetry(requestFn, maxRetries = 3, delay = 1000) {
    let lastError;
    
    for (let i = 0; i <= maxRetries; i++) {
        try {
            return await requestFn();
        } catch (error) {
            lastError = error;
            
            // 如果是认证错误，不重试
            if (error.message.includes('401') || error.message.includes('认证失败')) {
                throw error;
            }
            
            // 如果不是最后一次尝试，等待后重试
            if (i < maxRetries) {
                await new Promise(resolve => setTimeout(resolve, delay * (i + 1)));
            }
        }
    }
    
    throw lastError;
}

/**
 * 处理认证错误
 */
function handleAuthError() {
    // 清除全局authToken变量
    if (typeof window !== 'undefined') {
        window.authToken = null;
    }
    localStorage.removeItem('auth_token');
    
    // 如果当前不在登录页面，跳转到登录页面
    if (typeof showLoginPage === 'function') {
        showLoginPage();
    } else {
        // 刷新页面回到登录状态
        location.reload();
    }
}

/**
 * 设置认证令牌
 * @param {string} token - 认证令牌
 */
function setAuthToken(token) {
    // 同时设置全局变量和localStorage
    if (typeof window !== 'undefined') {
        window.authToken = token;
    }
    if (token) {
        localStorage.setItem('auth_token', token);
    } else {
        localStorage.removeItem('auth_token');
    }
}

/**
 * 获取认证令牌
 * @returns {string|null} 认证令牌
 */
function getAuthToken() {
    return (typeof window !== 'undefined' ? window.authToken : null) || localStorage.getItem('auth_token');
}

/**
 * 检查是否已认证
 * @returns {boolean} 是否已认证
 */
function isAuthenticated() {
    const currentAuthToken = (typeof window !== 'undefined' ? window.authToken : null) || localStorage.getItem('auth_token');
    return !!currentAuthToken;
}

/**
 * 请求拦截器（在请求发送前执行）
 * @param {Function} interceptor - 拦截器函数
 */
function addRequestInterceptor(interceptor) {
    // 这里可以实现请求拦截器逻辑
    // 例如添加通用的请求头、参数等
}

/**
 * 响应拦截器（在响应处理后执行）
 * @param {Function} interceptor - 拦截器函数
 */
function addResponseInterceptor(interceptor) {
    // 这里可以实现响应拦截器逻辑
    // 例如统一的错误处理、数据转换等
}

// 常用的API端点
const API_ENDPOINTS = {
    // 认证相关
    LOGIN: '/admin/login',
    LOGOUT: '/admin/logout',
    
    // 仪表板
    DASHBOARD: '/admin/dashboard',
    
    // 文件管理
    FILES: '/admin/files',
    FILE_DETAIL: (id) => `/admin/files/${id}`,
    FILE_DOWNLOAD: '/admin/files/download',
    
    // 用户管理
    USERS: '/admin/users',
    USER_DETAIL: (id) => `/admin/users/${id}`,
    USER_STATUS: (id) => `/admin/users/${id}/status`,
    USER_FILES: (id) => `/admin/users/${id}/files`,
    
    // 存储管理
    STORAGE: '/admin/storage',
    STORAGE_CONFIG: '/admin/storage/config',
    STORAGE_TEST: (type) => `/admin/storage/test/${type}`,
    STORAGE_SWITCH: '/admin/storage/switch',
    
    // MCP 服务器
    MCP_CONFIG: '/admin/mcp/config',
    MCP_STATUS: '/admin/mcp/status',
    MCP_CONTROL: '/admin/mcp/control',
    MCP_RESTART: '/admin/mcp/restart',
    
    // 系统配置
    CONFIG: '/admin/config',
    
    // 系统维护
    MAINTENANCE_CLEAN_EXPIRED: '/admin/maintenance/clean-expired',
    MAINTENANCE_CLEAN_TEMP: '/admin/maintenance/clean-temp',
    MAINTENANCE_CLEAN_INVALID: '/admin/maintenance/clean-invalid',
    MAINTENANCE_DB_OPTIMIZE: '/admin/maintenance/db/optimize',
    MAINTENANCE_DB_ANALYZE: '/admin/maintenance/db/analyze',
    MAINTENANCE_DB_BACKUP: '/admin/maintenance/db/backup',
    MAINTENANCE_CACHE_CLEAR: '/admin/maintenance/cache/clear-system',
    MAINTENANCE_MONITOR_SYSTEM: '/admin/maintenance/monitor/system'
};

// 导出API函数（如果使用模块系统）
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        apiRequest,
        apiGet,
        apiPost,
        apiPut,
        apiDelete,
        apiUpload,
        apiDownload,
        apiBatch,
        apiRetry,
        setAuthToken,
        getAuthToken,
        isAuthenticated,
        handleAuthError,
        addRequestInterceptor,
        addResponseInterceptor,
        API_ENDPOINTS
    };
}
