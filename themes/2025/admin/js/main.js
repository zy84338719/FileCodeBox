// 主入口文件 - 应用程序初始化和全局控制

// ========== 立即可用的全局函数 ==========

/**
 * 切换标签页 - 立即可用版本
 * @param {string} tabName - 标签页名称
 */
function switchTab(tabName) {
    try {
        // 如果未认证，显示登录提示
        const authToken = localStorage.getItem('user_token');
        if (!authToken) {
            showLoginPrompt();
            return;
        }
        
        // 更新按钮状态
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        
        // 找到被点击的按钮并激活
        const clickedBtn = event ? event.target : document.querySelector(`.tab-btn[onclick*="${tabName}"]`);
        if (clickedBtn) {
            clickedBtn.classList.add('active');
        }
        
        // 隐藏登录提示
        const loginPrompt = document.getElementById('login-prompt');
        if (loginPrompt) {
            loginPrompt.classList.remove('active');
        }
        
        // 更新内容显示
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.remove('active');
        });
        
        const targetTab = document.getElementById(tabName + '-tab');
        if (targetTab) {
            targetTab.classList.add('active');
        }
        
        // 根据标签页加载相应数据
        if (typeof loadTabData === 'function') {
            loadTabData(tabName);
        }
        
        console.log(`Switched to tab: ${tabName}`);
    } catch (error) {
        console.error('Failed to switch tab:', error);
        if (typeof showAlert === 'function') {
            showAlert('切换标签页失败', 'error');
        }
    }
}

/**
 * 显示登录提示 - 立即可用版本
 */
function showLoginPrompt() {
    try {
        const loginPrompt = document.getElementById('login-prompt');
        if (loginPrompt) {
            loginPrompt.classList.add('active');
        }
        
        // 显示登录模态框或重定向到登录页面
        if (typeof showLoginModal === 'function') {
            showLoginModal();
        } else {
            alert('请先登录！');
        }
    } catch (error) {
        console.error('Failed to show login prompt:', error);
    }
}

// ========== 应用状态管理 ==========

// 全局状态管理
const AppState = {
    currentTab: 'dashboard',
    isLoading: false,
    modals: new Set(),
    intervals: new Map(),
    timeouts: new Map()
};

// 全局变量
let currentPage = 1;
let currentSearch = '';
let authToken = localStorage.getItem('user_token'); // 使用统一的user_token
let currentStorageType = 'local';
let storageData = {};

/**
 * 应用程序初始化
 */
function initApp() {
    console.log('Initializing FileCodeBox Admin Panel...');
    
    try {
        // 初始化事件监听器
        initEventListeners();
        
        // 检查认证状态
        if (authToken) {
            // 验证token有效性
            verifyToken().then(async valid => {
                if (valid) {
                    await showAdminPage();
                } else {
                    // token无效，清除token但不立即跳转
                    authToken = null;
                    localStorage.removeItem('user_token');
                    window.authToken = null;
                    showLoginPrompt();
                }
            }).catch((error) => {
                // 验证失败，清除token但不立即跳转
                authToken = null;
                localStorage.removeItem('user_token');
                window.authToken = null;
                showLoginPrompt();
            });
        } else {
            // 没有token，显示登录提示
            showLoginPrompt();
        }
        
        console.log('FileCodeBox Admin Panel initialized successfully');
    } catch (error) {
        console.error('Failed to initialize app:', error);
        showAlert('应用程序初始化失败: ' + error.message, 'error');
    }
}

/**
 * 处理管理员登录
 */
async function handleAdminLogin(event) {
    event.preventDefault();
    
    const username = document.getElementById('admin-username').value;
    const password = document.getElementById('admin-password').value;
    const errorDiv = document.getElementById('login-error');
    
    if (!username || !password) {
        errorDiv.textContent = '请输入用户名和密码';
        errorDiv.style.display = 'block';
        return;
    }
    
    try {
        showLoading('正在登录...');
        
        const response = await fetch('/admin/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: username,
                password: password
            })
        });
        
        const result = await response.json();
        
        if (result.code === 200 && result.data && result.data.token) {
            // 保存token
            authToken = result.data.token;
            window.authToken = authToken;
            localStorage.setItem('user_token', authToken);
            
            // 隐藏错误信息
            errorDiv.style.display = 'none';
            
            // 显示管理页面
            await showAdminPage();
            
            showAlert('登录成功！', 'success');
        } else {
            errorDiv.textContent = result.message || '登录失败';
            errorDiv.style.display = 'block';
        }
    } catch (error) {
        console.error('Login error:', error);
        errorDiv.textContent = '登录请求失败: ' + error.message;
        errorDiv.style.display = 'block';
    } finally {
        hideLoading();
    }
}

/**
 * 跳转到用户登录页面
 */
function redirectToUserLogin() {
    // 保存当前页面路径，登录后可以返回
    sessionStorage.setItem('redirect_after_login', '/admin/');
    // 跳转到用户登录页面
    window.location.href = '/user/login';
}

/**
 * 显示登录提示页面
 */
function showLoginPrompt() {
    // 隐藏所有标签页内容
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
    });
    
    // 显示或创建登录提示页面
    let loginPrompt = document.getElementById('login-prompt');
    if (!loginPrompt) {
        loginPrompt = document.createElement('div');
        loginPrompt.id = 'login-prompt';
        loginPrompt.className = 'tab-content active';
        loginPrompt.innerHTML = `
            <div style="text-align: center; padding: 60px 20px;">
                <div style="max-width: 400px; margin: 0 auto; background: white; padding: 40px; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.1);">
                    <i class="fas fa-user-shield" style="font-size: 48px; color: #007bff; margin-bottom: 20px;"></i>
                    <h2 style="color: #333; margin-bottom: 16px;">管理员登录</h2>
                    <form id="admin-login-form" style="text-align: left;">
                        <div style="margin-bottom: 15px;">
                            <label style="display: block; margin-bottom: 5px; color: #555;">用户名</label>
                            <input type="text" id="admin-username" placeholder="请输入管理员用户名" style="width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box;">
                        </div>
                        <div style="margin-bottom: 20px;">
                            <label style="display: block; margin-bottom: 5px; color: #555;">密码</label>
                            <input type="password" id="admin-password" placeholder="请输入密码" style="width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box;">
                        </div>
                        <button type="submit" style="width: 100%; padding: 12px; background: #007bff; color: white; border: none; border-radius: 6px; cursor: pointer; font-size: 16px;">
                            登录
                        </button>
                        <div id="login-error" style="margin-top: 10px; color: #dc3545; display: none;"></div>
                    </form>
                </div>
            </div>
        `;
        
        // 添加到标签页容器中
        const tabsContainer = document.querySelector('#tab-content-container');
        if (tabsContainer) {
            tabsContainer.appendChild(loginPrompt);
        } else {
            document.body.appendChild(loginPrompt);
        }
        
        // 绑定登录表单事件
        const form = document.getElementById('admin-login-form');
        if (form) {
            form.addEventListener('submit', handleAdminLogin);
        }
    } else {
        loginPrompt.classList.add('active');
    }
    
    // 隐藏所有标签按钮的active状态
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
}

/**
 * 初始化事件监听器
 */
function initEventListeners() {
    // 移除了管理员登录表单的事件监听器，因为现在使用统一登录

    // 配置表单 - 由 config-simple.js 处理
    // const configForm = document.getElementById('config-form');
    // if (configForm) {
    //     configForm.addEventListener('submit', handleConfigSubmit);
    // }

    // 编辑文件表单 - 由 files.js 处理  
    // const editForm = document.getElementById('edit-form');
    // if (editForm) {
    //     editForm.addEventListener('submit', handleEditSubmit);
    // }

    // 搜索输入框 - 由 files.js 处理
    // const searchInput = document.getElementById('search-input');
    // if (searchInput) {
    //     searchInput.addEventListener('keypress', function(e) {
    //         if (e.key === 'Enter') {
    //             searchFiles();
    //         }
    //     });
    // }

    // 用户系统开关 - 由 config-simple.js 处理
    // const enableUserSystem = document.getElementById('enable_user_system');
    // if (enableUserSystem) {
    //     enableUserSystem.addEventListener('change', toggleUserSystemOptions);
    // }

    // 模态框关闭 - 由各自模块处理
    // const closeBtn = document.querySelector('.close');
    // if (closeBtn) {
    //     closeBtn.onclick = closeModal;
    // }

    // 点击模态框外部关闭 - 由各自模块处理
    window.onclick = function(event) {
        const modal = document.getElementById('edit-modal');
        if (event.target == modal) {
            closeModal();
        }
    }

    // 存储卡片点击事件 - 由 storage-simple.js 处理
    // ['local', 'webdav', 'nfs', 's3'].forEach(type => {
    //     const card = document.getElementById(`${type}-card`);
    //     if (card) {
    //         card.addEventListener('click', () => selectStorageCard(type));
    //     }
    // });
}

// ========== 认证相关功能 ==========

/**
 * 显示管理页面
 */
async function showAdminPage() {
    console.log('Showing admin page...');
    
    // 默认显示dashboard标签
    switchTab('dashboard');
    
    // 异步加载仪表板数据（不阻塞页面显示）
    try {
        await loadStats();
    } catch (error) {
        console.error('加载统计数据失败:', error);
        // 即使统计数据加载失败，也不影响页面显示
    }
}

/**
 * 验证token有效性
 */
async function verifyToken() {
    try {
        // 使用用户API验证token并检查管理员权限
        const result = await apiRequest('/user/profile');
        if (result.code === 200 && result.data && result.data.role === 'admin') {
            return true;
        }
        return false;
    } catch (error) {
        console.warn('Token验证失败:', error);
        return false;
    }
}

/**
 * 退出登录
 */
function logout() {
    authToken = null;
    window.authToken = null; // 清除全局变量
    localStorage.removeItem('user_token');
    redirectToUserLogin();
    showAlert('已退出登录', 'info');
}

/**
 * 跳转到用户页面
 */
function goToUser() {
    window.location.href = '/user/dashboard';
}

// ========== API请求封装 ==========

/**
 * API请求封装
 */
async function apiRequest(url, options = {}) {
    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json'
        }
    };
    
    const finalOptions = {
        ...defaultOptions,
        ...options,
        headers: {
            ...defaultOptions.headers,
            ...options.headers
        }
    };
    
    if (authToken) {
        finalOptions.headers['Authorization'] = `Bearer ${authToken}`;
        console.log('🔑 使用Bearer token进行API请求:', url);
    } else {
        console.log('🔓 无token，发送匿名API请求:', url);
    }
    
    const response = await fetch(url, finalOptions);
    console.log('📡 API响应状态:', response.status, response.statusText);
    
    if (response.status === 401) {
        console.log('🚫 收到401未授权响应，执行自动登出');
        logout();
        throw new Error('认证失败');
    }
    
    return response.json();
}

// ========== 统计数据 ==========

/**
 * 加载统计数据
 */
async function loadStats() {
    // 检查认证状态
    if (!authToken && !window.authToken) {
        console.log('未认证，跳过统计数据加载');
        return;
    }
    
    try {
        const result = await apiRequest('/admin/dashboard');
        
        if (result.code === 200) {
            const stats = result.data;
            
            // 更新文件标签页的统计数据（保持兼容性）
            const totalFilesEl = document.getElementById('total-files');
            const todayUploadsEl = document.getElementById('today-uploads');
            const activeFilesEl = document.getElementById('active-files');
            const totalStorageEl = document.getElementById('total-storage');
            
            if (totalFilesEl) totalFilesEl.textContent = stats.total_files || 0;
            if (todayUploadsEl) todayUploadsEl.textContent = stats.today_uploads || 0;
            if (activeFilesEl) activeFilesEl.textContent = stats.active_files || 0;
            if (totalStorageEl) totalStorageEl.textContent = formatFileSize(stats.total_size || 0);
            
            // 更新仪表板页面的统计数据
            const dashboardTotalFilesEl = document.getElementById('dashboard-total-files');
            const dashboardTodayUploadsEl = document.getElementById('dashboard-today-uploads');
            const dashboardActiveUsersEl = document.getElementById('dashboard-active-users');
            const dashboardTotalStorageEl = document.getElementById('dashboard-total-storage');
            
            if (dashboardTotalFilesEl) dashboardTotalFilesEl.textContent = stats.total_files || 0;
            if (dashboardTodayUploadsEl) dashboardTodayUploadsEl.textContent = stats.today_uploads || 0;
            if (dashboardActiveUsersEl) dashboardActiveUsersEl.textContent = stats.active_files || 0; // 临时使用active_files作为活跃用户数
            if (dashboardTotalStorageEl) dashboardTotalStorageEl.textContent = formatFileSize(stats.total_size || 0);
            
            // 更新趋势百分比（如果后端提供）
            const filesTrendEl = document.getElementById('files-trend');
            const uploadsTrendEl = document.getElementById('uploads-trend');
            const usersTrendEl = document.getElementById('users-trend');
            const storageTrendEl = document.getElementById('storage-trend');

            if (filesTrendEl) {
                if (stats.files_change_percent !== undefined && stats.files_change_percent !== null) {
                    filesTrendEl.textContent = (stats.files_change_percent > 0 ? '+' : '') + stats.files_change_percent + '%';
                } else {
                    filesTrendEl.textContent = '—';
                }
            }

            if (uploadsTrendEl) {
                if (stats.uploads_change_percent !== undefined && stats.uploads_change_percent !== null) {
                    uploadsTrendEl.textContent = (stats.uploads_change_percent > 0 ? '+' : '') + stats.uploads_change_percent + '%';
                } else {
                    uploadsTrendEl.textContent = '—';
                }
            }

            if (usersTrendEl) {
                if (stats.users_change_percent !== undefined && stats.users_change_percent !== null) {
                    usersTrendEl.textContent = (stats.users_change_percent > 0 ? '+' : '') + stats.users_change_percent + '%';
                } else {
                    usersTrendEl.textContent = '—';
                }
            }

            if (storageTrendEl) {
                if (stats.storage_change_percent !== undefined && stats.storage_change_percent !== null) {
                    storageTrendEl.textContent = (stats.storage_change_percent > 0 ? '+' : '') + stats.storage_change_percent + '%';
                } else {
                    storageTrendEl.textContent = '—';
                }
            }
            // 更新存储使用率（如果API提供了相关数据）
            const storageUsageEl = document.getElementById('storage-usage');
            if (storageUsageEl && stats.storage_usage_percent) {
                storageUsageEl.textContent = `${stats.storage_usage_percent}% 已使用`;
            }
        }
    } catch (error) {
        console.error('加载统计数据失败:', error);
        // 即使统计数据加载失败，也不要阻止页面显示
    }
}

// ========== 标签页数据加载 ==========

/**
 * 加载标签页数据
 * @param {string} tabName - 标签页名称
 */
function loadTabData(tabName) {
    // 检查认证状态，未认证时不加载数据
    if (!authToken && !window.authToken) {
        console.log(`未认证，跳过标签页 ${tabName} 的数据加载`);
        return;
    }
    
    switch (tabName) {
        case 'dashboard':
            // 加载仪表板统计数据
            loadStats();
            break;
        case 'files':
            // 由 files.js 处理
            if (typeof initFileInterface === 'function') {
                initFileInterface();
            }
            break;
        case 'users':
            // 由 users.js 处理
            if (typeof initUserInterface === 'function') {
                initUserInterface();
            } else if (typeof loadUsers === 'function') {
                loadUsers();
            }
            break;
        case 'storage':
            // 由 storage-simple.js 处理
            if (typeof loadStorageInfo === 'function') {
                loadStorageInfo();
            }
            break;
        case 'mcp':
            // 由 mcp-simple.js 处理
            if (typeof loadMCPConfig === 'function') {
                loadMCPConfig();
            }
            if (typeof loadMCPStatus === 'function') {
                loadMCPStatus();
            }
            break;
        case 'config':
            // 由 config-simple.js 处理
            if (typeof loadConfig === 'function') {
                loadConfig();
            }
            break;
        case 'maintenance':
            // 维护页面不需要预加载数据
            break;
        default:
            console.warn(`Unknown tab: ${tabName}`);
    }
}

// ========== 工具函数 ==========

/**
 * 显示加载提示
 */
function showLoading(message = '加载中...') {
    // 创建或更新加载提示
    let loadingDiv = document.getElementById('global-loading');
    if (!loadingDiv) {
        loadingDiv = document.createElement('div');
        loadingDiv.id = 'global-loading';
        loadingDiv.style.cssText = `
            position: fixed;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            background: rgba(0, 0, 0, 0.8);
            color: white;
            padding: 20px;
            border-radius: 8px;
            z-index: 10000;
            text-align: center;
        `;
        document.body.appendChild(loadingDiv);
    }
    loadingDiv.innerHTML = `
        <div style="margin-bottom: 10px;">
            <i class="fas fa-spinner fa-spin" style="font-size: 24px;"></i>
        </div>
        <div>${message}</div>
    `;
    loadingDiv.style.display = 'block';
}

/**
 * 隐藏加载提示
 */
function hideLoading() {
    const loadingDiv = document.getElementById('global-loading');
    if (loadingDiv) {
        loadingDiv.style.display = 'none';
    }
}

/**
 * 显示提示信息
 */
function showAlert(message, type = 'info') {
    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${type}`;
    alertDiv.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 15px;
        border-radius: 5px;
        color: white;
        font-weight: bold;
        z-index: 9999;
        max-width: 400px;
        word-wrap: break-word;
    `;
    
    // 根据类型设置背景色
    switch(type) {
        case 'success':
            alertDiv.style.background = '#28a745';
            break;
        case 'error':
            alertDiv.style.background = '#dc3545';
            break;
        case 'warning':
            alertDiv.style.background = '#ffc107';
            alertDiv.style.color = '#212529';
            break;
        default:
            alertDiv.style.background = '#17a2b8';
    }
    
    alertDiv.textContent = message;
    document.body.appendChild(alertDiv);
    
    // 3秒后自动移除
    setTimeout(() => {
        if (alertDiv.parentNode) {
            alertDiv.parentNode.removeChild(alertDiv);
        }
    }, 3000);
}

/**
 * 格式化文件大小
 */
function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

/**
 * 格式化时间
 */
function formatDateTime(dateString) {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN');
}

/**
 * 安全的时间格式化函数
 */
function formatDateTimeLocal(dateString) {
    try {
        if (!dateString) return '';
        
        let date;
        if (dateString.includes('+') || dateString.includes('Z')) {
            date = new Date(dateString);
        } else {
            date = new Date(dateString + '+08:00');
        }
        
        if (isNaN(date.getTime())) {
            console.warn('Invalid date string:', dateString);
            return '';
        }
        
        // 转换为本地时间的datetime-local格式
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');
        
        return `${year}-${month}-${day}T${hours}:${minutes}`;
    } catch (error) {
        console.warn('Error formatting date:', dateString, error);
        return '';
    }
}

// 移动端菜单切换
function toggleMobileMenu() {
    const tabHeader = document.querySelector('.tab-header');
    const overlay = document.querySelector('.mobile-menu-overlay');
    
    if (tabHeader) {
        tabHeader.classList.toggle('mobile-active');
        
        // 如果没有遮罩层，创建一个
        if (!overlay && tabHeader.classList.contains('mobile-active')) {
            const newOverlay = document.createElement('div');
            newOverlay.className = 'mobile-menu-overlay';
            newOverlay.onclick = closeMobileMenu;
            document.body.appendChild(newOverlay);
        } else if (overlay && !tabHeader.classList.contains('mobile-active')) {
            overlay.remove();
        }
    }
}

// 关闭移动端菜单
function closeMobileMenu() {
    const tabHeader = document.querySelector('.tab-header');
    const overlay = document.querySelector('.mobile-menu-overlay');
    
    if (tabHeader) {
        tabHeader.classList.remove('mobile-active');
    }
    
    if (overlay) {
        overlay.remove();
    }
}

// DOM 加载完成后初始化应用程序
document.addEventListener('DOMContentLoaded', () => {
    console.log('DOM loaded, initializing app...');
    initApp();
    
    // 点击标签页项目时自动关闭移动端菜单
    document.querySelectorAll('.tab-item').forEach(item => {
        item.addEventListener('click', () => {
            if (window.innerWidth <= 768) {
                closeMobileMenu();
            }
        });
    });
});

// 将关键函数和变量暴露到全局作用域
window.switchTab = switchTab;
window.logout = logout;
window.goToUser = goToUser;
window.showAlert = showAlert;
window.showLoading = showLoading;
window.hideLoading = hideLoading;
window.toggleMobileMenu = toggleMobileMenu;
window.closeMobileMenu = closeMobileMenu;
window.redirectToUserLogin = redirectToUserLogin;
window.showLoginPrompt = showLoginPrompt;
window.authToken = authToken;
