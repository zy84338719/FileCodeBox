// 前端 API 请求封装

/**
 * 通用请求方法
 * @param {string} url 请求URL
 * @param {Object} options fetch选项
 * @returns {Promise<Object>} 解析后的JSON结果
 */
async function apiRequest(url, options = {}) {
    const currentAuthToken =
        (typeof window !== 'undefined' ? window.authToken : null) ||
        localStorage.getItem('user_token');

    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json'
        }
    };

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

        if (response.status === 401) {
            handleAuthError();
            throw new Error('认证失败，请重新登录');
        }

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
 */
async function apiGet(url, params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const fullUrl = queryString ? `${url}?${queryString}` : url;
    return apiRequest(fullUrl, { method: 'GET' });
}

/**
 * POST 请求
 */
async function apiPost(url, data = {}) {
    return apiRequest(url, {
        method: 'POST',
        body: JSON.stringify(data)
    });
}

/**
 * PUT 请求
 */
async function apiPut(url, data = {}) {
    return apiRequest(url, {
        method: 'PUT',
        body: JSON.stringify(data)
    });
}

/**
 * DELETE 请求
 */
async function apiDelete(url) {
    return apiRequest(url, { method: 'DELETE' });
}

/**
 * 文件上传请求
 */
async function apiUpload(url, formData, onProgress = null) {
    return new Promise((resolve, reject) => {
        const currentAuthToken =
            (typeof window !== 'undefined' ? window.authToken : null) ||
            localStorage.getItem('user_token');
        const xhr = new XMLHttpRequest();

        if (onProgress) {
            xhr.upload.addEventListener('progress', (e) => {
                if (e.lengthComputable) {
                    const percentComplete = (e.loaded / e.total) * 100;
                    onProgress(percentComplete);
                }
            });
        }

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

        xhr.addEventListener('error', () => {
            reject(new Error('网络错误'));
        });

        xhr.addEventListener('timeout', () => {
            reject(new Error('请求超时'));
        });

        xhr.open('POST', url);
        if (currentAuthToken) {
            xhr.setRequestHeader('Authorization', 'Bearer ' + currentAuthToken);
        }
        xhr.timeout = 300000;
        xhr.send(formData);
    });
}

/**
 * 文件下载请求
 */
async function apiDownload(url, filename = 'download') {
    try {
        const currentAuthToken =
            (typeof window !== 'undefined' ? window.authToken : null) ||
            localStorage.getItem('user_token');
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
        const contentDisposition = response.headers.get('Content-Disposition');
        if (contentDisposition) {
            const matches = contentDisposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/);
            if (matches != null && matches[1]) {
                filename = matches[1].replace(/['"]/g, '');
            }
        }

        const urlObj = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = urlObj;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(urlObj);

        return { success: true, filename };
    } catch (error) {
        console.error('Download Error:', error);
        throw error;
    }
}

/**
 * 批量请求
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
 */
async function apiRetry(requestFn, maxRetries = 3, delay = 1000) {
    let lastError;
    for (let i = 0; i <= maxRetries; i++) {
        try {
            return await requestFn();
        } catch (error) {
            lastError = error;
            if (error.message.includes('401') || error.message.includes('认证失败')) {
                throw error;
            }
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
    if (typeof window !== 'undefined') {
        window.authToken = null;
    }
    localStorage.removeItem('user_token');
    if (typeof UserAuth !== 'undefined') {
        try {
            UserAuth.removeToken();
            UserAuth.removeUserInfo();
            UserAuth.updateUI();
        } catch (e) {
            console.error('处理认证错误失败:', e);
        }
    } else {
        location.reload();
    }
}

/**
 * 设置认证令牌
 */
function setAuthToken(token) {
    if (typeof window !== 'undefined') {
        window.authToken = token;
    }
    if (token) {
        localStorage.setItem('user_token', token);
    } else {
        localStorage.removeItem('user_token');
    }
}

/**
 * 获取认证令牌
 */
function getAuthToken() {
    return (typeof window !== 'undefined' ? window.authToken : null) || localStorage.getItem('user_token');
}

/**
 * 检查是否已认证
 */
function isAuthenticated() {
    const currentAuthToken =
        (typeof window !== 'undefined' ? window.authToken : null) ||
        localStorage.getItem('user_token');
    return !!currentAuthToken;
}

/**
 * 请求拦截器（占位）
 */
function addRequestInterceptor(interceptor) {
    // 可在此实现请求拦截逻辑
}

/**
 * 响应拦截器（占位）
 */
function addResponseInterceptor(interceptor) {
    // 可在此实现响应拦截逻辑
}

// 常用API端点
const API_ENDPOINTS = {
    CONFIG: '/',
    SHARE_TEXT: '/share/text/',
    SHARE_FILE: '/share/file/',
    SHARE_SELECT: '/share/select/',
    USER_PROFILE: '/user/profile',
    USER_STATS: '/user/stats',
    USER_FILES: '/user/files',
    USER_FILE_DETAIL: (id) => `/user/files/${id}`,
    USER_CHANGE_PASSWORD: '/user/change-password',
    USER_LOGOUT: '/user/logout',
    USER_SYSTEM_INFO: '/user/system-info'
};

// 导出（用于模块系统）
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

