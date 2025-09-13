// 用户管理模块

/**
 * 安全调用showAlert函数
 */
function safeShowAlert(message, type = 'info', duration = 3000) {
    if (typeof window.showAlert === 'function') {
        window.showAlert(message, type, duration);
    } else {
        console.log(`[${type.toUpperCase()}] ${message}`);
    }
}

// 全局变量
let userFilters = {
    status: '',
    dateRange: '',
    sortBy: 'created_at_desc'
};

let currentUserPage = 1;
let currentUserSearch = '';
let currentUserLimit = 10;  // 添加动态limit支持

/**
 * 初始化用户管理界面
 */
function initUserInterface() {
    // 移除loadUserStats调用，因为后端没有这个API端点
    // loadUserStats();
    loadUsers();
}

// 注释掉loadUserStats函数，因为后端没有/admin/users/stats端点
/*
async function loadUserStats() {
    try {
        const result = await apiRequest('/admin/users/stats');
        
        if (result.code === 200) {
            const stats = result.data;
            updateUserStats(stats);
        }
    } catch (error) {
        console.error('加载用户统计失败:', error);
    }
}
*/

/**
 * 更新用户统计数据
 */
function updateUserStats(stats) {
    const elements = {
        'total-users': stats.total_users || 0,
        'active-users': stats.active_users || 0,
        'new-users-today': stats.today_registrations || 0,
        'online-users': stats.today_uploads || 0  // 暂时用今日上传数替代在线用户数
    };
    
    Object.entries(elements).forEach(([id, value]) => {
        const element = document.getElementById(id);
        if (element) {
            element.textContent = value;
        }
    });
}

/**
 * 加载用户列表
 */
async function loadUsers(page = 1, search = '') {
    try {
        const params = new URLSearchParams({
            page: page,
            limit: currentUserLimit,  // 使用动态limit
            search: search,
            status: userFilters.status,
            date_range: userFilters.dateRange,
            sort_by: userFilters.sortBy
        });
        
        // 移除空值参数
        for (let [key, value] of [...params]) {
            if (!value) params.delete(key);
        }
        
        const url = `/admin/users?${params.toString()}`;
        const result = await apiRequest(url);
        
        if (result.code === 200) {
            displayUsers(result.data?.users || []);
            updateUserPagination(result.data?.pagination || {});
            // 更新用户统计数据
            if (result.data?.stats) {
                updateUserStats(result.data.stats);
            }
            currentUserPage = page;
            currentUserSearch = search;
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('加载用户列表失败:', error);
        safeShowAlert('加载用户列表失败: ' + error.message, 'error');
        displayUsersError(error.message);
    }
}

/**
 * 显示用户列表
 */
function displayUsers(users) {
    const tbody = document.getElementById('users-tbody');
    if (!tbody) return;
    
    if (users.length === 0) {
        tbody.innerHTML = `
            <tr>
                <td colspan="9" class="loading-cell">
                    <div class="empty-state">
                        <i class="fas fa-users" style="font-size: 48px; color: #dee2e6; margin-bottom: 16px;"></i>
                        <h4>暂无用户数据</h4>
                        <p>还没有用户注册，或者当前搜索条件下没有匹配的用户。</p>
                    </div>
                </td>
            </tr>
        `;
        return;
    }
    
    tbody.innerHTML = users.map(user => `
        <tr>
            <td>
                <input type="checkbox" class="user-checkbox" value="${user.id}">
            </td>
            <td>
                <div class="user-info">
                    <div class="user-avatar">
                        ${user.username.charAt(0).toUpperCase()}
                    </div>
                    <div class="user-details">
                        <span class="user-name">${escapeHtml(user.username)}</span>
                        <small class="user-id">ID: ${user.id}</small>
                    </div>
                </div>
            </td>
            <td>${escapeHtml(user.email) || '<span style="color: #9ca3af; font-style: italic;">未设置邮箱</span>'}</td>
            <td>
                <div style="display: flex; flex-direction: column; gap: 4px;">
                    <span style="font-weight: 500; color: #374151;">${formatDateTime(user.created_at)}</span>
                    <small style="color: #6b7280; font-size: 11px;">${formatRelativeTime(user.created_at)}</small>
                </div>
            </td>
            <td>
                <div style="display: flex; flex-direction: column; gap: 4px;">
                    <span style="font-weight: 500; color: #374151;">${user.last_login ? formatDateTime(user.last_login) : '从未登录'}</span>
                    ${user.last_login ? `<small style="color: #6b7280; font-size: 11px;">${formatRelativeTime(user.last_login)}</small>` : ''}
                </div>
            </td>
            <td>
                <span class="status-badge ${user.status === 'active' ? 'status-active' : 'status-inactive'}">
                    ${user.status === 'active' ? '活跃' : '禁用'}
                </span>
            </td>
            <td>
                <span style="font-weight: 600; color: #1e293b; font-size: 15px;">${user.file_count || 0}</span>
                ${user.file_limit > 0 ? `<small style="color: #6c757d;">/${user.file_limit}</small>` : ''}
            </td>
            <td>
                <div style="display: flex; flex-direction: column; gap: 2px;">
                    <span style="font-weight: 600;">${formatFileSize(user.storage_used || 0)}</span>
                    ${user.storage_limit > 0 ? `<small style="color: #6c757d;">/${formatFileSize(user.storage_limit * 1024 * 1024)}</small>` : ''}
                    ${user.storage_limit > 0 ? `
                        <div style="width: 60px; height: 4px; background: #e9ecef; border-radius: 2px; margin-top: 2px;">
                            <div style="width: ${Math.min(100, (user.storage_used / (user.storage_limit * 1024 * 1024)) * 100)}%; height: 100%; background: ${(user.storage_used / (user.storage_limit * 1024 * 1024)) > 0.8 ? '#dc3545' : '#007bff'}; border-radius: 2px;"></div>
                        </div>
                    ` : ''}
                </div>
            </td>
            <td class="actions">
                <button onclick="editUser('${user.id}')" class="btn-small" title="编辑用户">
                    <i class="fas fa-edit"></i>
                </button>
                <button onclick="deleteUser('${user.id}')" class="btn-small" title="删除用户">
                    <i class="fas fa-trash-alt"></i>
                </button>
                <button onclick="${user.status === 'active' ? 'disableUser' : 'enableUser'}('${user.id}')" 
                        class="btn-small ${user.status === 'active' ? '' : 'inactive'}" 
                        title="${user.status === 'active' ? '禁用用户' : '启用用户'}">
                    <i class="fas ${user.status === 'active' ? 'fa-user-times' : 'fa-user-check'}"></i>
                </button>
            </td>
        </tr>
    `).join('');
    
    // 更新全选状态
    updateSelectAllState();
}

/**
 * 显示用户列表错误
 */
function displayUsersError(error) {
    const tbody = document.getElementById('users-tbody');
    if (!tbody) return;
    
    tbody.innerHTML = `
        <tr>
            <td colspan="9" class="loading-cell">
                <div class="error-state">
                    <i class="fas fa-exclamation-triangle" style="font-size: 48px; color: #dc3545; margin-bottom: 16px;"></i>
                    <h4>加载失败</h4>
                    <p>${error}</p>
                    <button onclick="loadUsers()" class="btn btn-primary">
                        <i class="fas fa-redo"></i> 重新加载
                    </button>
                </div>
            </td>
        </tr>
    `;
}

/**
 * 更新分页信息
 */
function updateUserPagination(pagination) {
    const paginationContainer = document.getElementById('user-pagination-container');
    const paginationElement = document.getElementById('user-pagination');
    const pageStartEl = document.getElementById('page-start');
    const pageEndEl = document.getElementById('page-end');
    const totalCountEl = document.getElementById('total-count');
    
    if (!pagination || !paginationContainer || !paginationElement) return;
    
    const current_page = pagination.page || 1;
    const per_page = pagination.page_size || 10;
    const total = pagination.total || 0;
    const last_page = pagination.pages || 1;
    
    // 如果总数为0，隐藏分页容器
    if (total === 0) {
        paginationContainer.style.display = 'none';
        return;
    }
    
    // 显示分页容器
    paginationContainer.style.display = 'block';
    
    // 更新统计信息
    const start = (current_page - 1) * per_page + 1;
    const end = Math.min(current_page * per_page, total);
    
    if (pageStartEl) pageStartEl.textContent = total > 0 ? start : 0;
    if (pageEndEl) pageEndEl.textContent = end;
    if (totalCountEl) totalCountEl.textContent = total;
    
    // 生成分页按钮
    if (last_page <= 1) {
        paginationElement.innerHTML = '';
        return;
    }
    
    let paginationHTML = '<div class="pagination-wrapper">';
    
    // 首页按钮
    if (current_page > 1) {
        paginationHTML += `<button onclick="loadUsers(1, currentUserSearch)" class="btn-page btn-page-first" title="首页">
            <i class="fas fa-angle-double-left"></i>
        </button>`;
    }
    
    // 上一页按钮
    if (current_page > 1) {
        paginationHTML += `<button onclick="loadUsers(${current_page - 1}, currentUserSearch)" class="btn-page btn-page-prev" title="上一页">
            <i class="fas fa-angle-left"></i> 上一页
        </button>`;
    }
    
    // 页码按钮组
    paginationHTML += '<div class="pagination-numbers">';
    
    const startPage = Math.max(1, current_page - 2);
    const endPage = Math.min(last_page, current_page + 2);
    
    // 如果开始页码不是1，显示省略号
    if (startPage > 1) {
        paginationHTML += `<button onclick="loadUsers(1, currentUserSearch)" class="btn-page">1</button>`;
        if (startPage > 2) {
            paginationHTML += '<span class="pagination-ellipsis">...</span>';
        }
    }
    
    // 页码按钮
    for (let i = startPage; i <= endPage; i++) {
        const activeClass = i === current_page ? 'active' : '';
        paginationHTML += `<button onclick="loadUsers(${i}, currentUserSearch)" class="btn-page ${activeClass}" title="第${i}页">${i}</button>`;
    }
    
    // 如果结束页码不是最后一页，显示省略号
    if (endPage < last_page) {
        if (endPage < last_page - 1) {
            paginationHTML += '<span class="pagination-ellipsis">...</span>';
        }
        paginationHTML += `<button onclick="loadUsers(${last_page}, currentUserSearch)" class="btn-page">${last_page}</button>`;
    }
    
    paginationHTML += '</div>';
    
    // 下一页按钮
    if (current_page < last_page) {
        paginationHTML += `<button onclick="loadUsers(${current_page + 1}, currentUserSearch)" class="btn-page btn-page-next" title="下一页">
            下一页 <i class="fas fa-angle-right"></i>
        </button>`;
    }
    
    // 末页按钮
    if (current_page < last_page) {
        paginationHTML += `<button onclick="loadUsers(${last_page}, currentUserSearch)" class="btn-page btn-page-last" title="末页">
            <i class="fas fa-angle-double-right"></i>
        </button>`;
    }
    
    // 页面跳转控件
    paginationHTML += `<div class="pagination-jump">
        <span>跳转到</span>
        <input type="number" id="users-page-jump" min="1" max="${last_page}" value="${current_page}" style="width: 60px; text-align: center;">
        <button onclick="jumpToUserPage()" class="btn btn-sm">跳转</button>
    </div>`;
    
    // 页面大小选择器
    paginationHTML += `<div class="pagination-size">
        <span>每页显示</span>
        <select id="users-page-size" onchange="changeUserPageSize()">
            <option value="5" ${per_page === 5 ? 'selected' : ''}>5条</option>
            <option value="10" ${per_page === 10 ? 'selected' : ''}>10条</option>
            <option value="20" ${per_page === 20 ? 'selected' : ''}>20条</option>
            <option value="50" ${per_page === 50 ? 'selected' : ''}>50条</option>
        </select>
    </div>`;
    
    paginationHTML += '</div>';
    
    paginationElement.innerHTML = paginationHTML;
}

/**
 * 跳转到指定用户页面
 */
function jumpToUserPage() {
    const input = document.getElementById('users-page-jump');
    if (!input) return;
    
    const page = parseInt(input.value);
    if (isNaN(page) || page < 1) {
        safeShowAlert('请输入有效的页码', 'warning');
        return;
    }
    
    loadUsers(page, currentUserSearch);
}

/**
 * 更改用户页面大小
 */
function changeUserPageSize() {
    const select = document.getElementById('users-page-size');
    if (!select) return;
    
    const newLimit = parseInt(select.value);
    if (isNaN(newLimit) || newLimit < 1) return;
    
    // 更新limit参数并重新加载第一页
    currentUserLimit = newLimit;
    loadUsers(1, currentUserSearch);
}

/**
 * 搜索用户
 */
function searchUsers() {
    const searchInput = document.getElementById('user-search-input');
    if (!searchInput) return;
    
    const searchValue = searchInput.value.trim();
    loadUsers(1, searchValue);
}

/**
 * 清空搜索
 */
function clearUserSearch() {
    const searchInput = document.getElementById('user-search-input');
    if (searchInput) {
        searchInput.value = '';
        loadUsers(1, '');
    }
}

/**
 * 过滤用户
 */
function filterUsers() {
    const statusFilter = document.getElementById('status-filter');
    const dateFilter = document.getElementById('date-filter');
    const sortFilter = document.getElementById('sort-filter');
    
    userFilters.status = statusFilter ? statusFilter.value : '';
    userFilters.dateRange = dateFilter ? dateFilter.value : '';
    userFilters.sortBy = sortFilter ? sortFilter.value : 'created_at_desc';
    
    loadUsers(1, currentUserSearch);
}

/**
 * 刷新用户列表
 */
function refreshUsers() {
    // 移除loadUserStats调用，因为后端没有这个API端点
    // loadUserStats();
    loadUsers(currentUserPage, currentUserSearch);
    safeShowAlert('用户列表已刷新', 'success');
}

/**
 * 切换全选状态
 */
function toggleSelectAllUsers() {
    const selectAll = document.getElementById('select-all-users');
    const checkboxes = document.querySelectorAll('.user-checkbox');
    
    if (selectAll && checkboxes.length > 0) {
        checkboxes.forEach(checkbox => {
            checkbox.checked = selectAll.checked;
        });
    }
}

/**
 * 更新全选状态
 */
function updateSelectAllState() {
    const selectAll = document.getElementById('select-all-users');
    const checkboxes = document.querySelectorAll('.user-checkbox');
    
    if (selectAll && checkboxes.length > 0) {
        const checkedCount = document.querySelectorAll('.user-checkbox:checked').length;
        selectAll.checked = checkedCount === checkboxes.length;
        selectAll.indeterminate = checkedCount > 0 && checkedCount < checkboxes.length;
    }
}

/**
 * 获取选中的用户ID
 */
function getSelectedUserIds() {
    const checkboxes = document.querySelectorAll('.user-checkbox:checked');
    return Array.from(checkboxes).map(cb => cb.value);
}

/**
 * 批量启用用户
 */
async function batchEnableUsers() {
    const userIds = getSelectedUserIds();
    if (userIds.length === 0) {
        safeShowAlert('请先选择要启用的用户', 'warning');
        return;
    }
    
    if (!confirm(`确定要启用选中的 ${userIds.length} 个用户吗？`)) {
        return;
    }
    
    try {
        const result = await apiRequest('/admin/users/batch-enable', {
            method: 'POST',
            body: JSON.stringify({ user_ids: userIds })
        });
        
        if (result.code === 200) {
            refreshUsers();
            safeShowAlert(`成功启用 ${userIds.length} 个用户`, 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('批量启用用户失败:', error);
        safeShowAlert('批量启用用户失败: ' + error.message, 'error');
    }
}

/**
 * 批量禁用用户
 */
async function batchDisableUsers() {
    const userIds = getSelectedUserIds();
    if (userIds.length === 0) {
        safeShowAlert('请先选择要禁用的用户', 'warning');
        return;
    }
    
    if (!confirm(`确定要禁用选中的 ${userIds.length} 个用户吗？`)) {
        return;
    }
    
    try {
        const result = await apiRequest('/admin/users/batch-disable', {
            method: 'POST',
            body: JSON.stringify({ user_ids: userIds })
        });
        
        if (result.code === 200) {
            refreshUsers();
            safeShowAlert(`成功禁用 ${userIds.length} 个用户`, 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('批量禁用用户失败:', error);
        safeShowAlert('批量禁用用户失败: ' + error.message, 'error');
    }
}

/**
 * 批量删除用户
 */
async function batchDeleteUsers() {
    const userIds = getSelectedUserIds();
    if (userIds.length === 0) {
        safeShowAlert('请先选择要删除的用户', 'warning');
        return;
    }
    
    if (!confirm(`确定要删除选中的 ${userIds.length} 个用户吗？此操作不可恢复。`)) {
        return;
    }
    
    try {
        const result = await apiRequest('/admin/users/batch-delete', {
            method: 'POST',
            body: JSON.stringify({ user_ids: userIds })
        });
        
        if (result.code === 200) {
            refreshUsers();
            safeShowAlert(`成功删除 ${userIds.length} 个用户`, 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('批量删除用户失败:', error);
        safeShowAlert('批量删除用户失败: ' + error.message, 'error');
    }
}

/**
 * 切换用户操作菜单
 */
function toggleUserActions() {
    const menu = document.getElementById('user-actions-menu');
    if (menu) {
        menu.classList.toggle('show');
    }
}

/**
 * 显示添加用户模态框
 */
function showAddUserModal() {
    const modal = document.getElementById('user-modal');
    if (!modal) return;
    
    // 重置表单
    const form = document.getElementById('user-form');
    if (form) form.reset();
    
    // 设置默认值
    document.getElementById('user-modal-title').textContent = '✨ 添加用户';
    document.getElementById('user-id').value = '';
    document.getElementById('user-active').checked = true;
    document.getElementById('user-storage-limit').value = '100';
    document.getElementById('user-file-limit').value = '1000';
    document.getElementById('password-hint').textContent = '密码长度至少6位';
    document.getElementById('user-password').required = true;
    
    modal.style.display = 'block';
}

/**
 * 查看用户文件
 */
async function viewUserFiles(userId) {
    try {
        // 切换到文件管理标签并过滤该用户的文件
        switchTab('files');
        
        // 等待一下确保标签切换完成
        setTimeout(() => {
            const searchInput = document.getElementById('search-input');
            if (searchInput) {
                searchInput.value = `user:${userId}`;
                if (typeof searchFiles === 'function') {
                    searchFiles();
                }
            }
        }, 100);
        
        safeShowAlert('已切换到文件管理页面，正在加载用户文件', 'info');
    } catch (error) {
        console.error('查看用户文件失败:', error);
        safeShowAlert('查看用户文件失败: ' + error.message, 'error');
    }
}

/**
 * 编辑用户
 */
async function editUser(userId) {
    try {
        const result = await apiRequest(`/admin/users/${userId}`);
        
        if (result.code === 200) {
            const user = result.data;
            showUserModal(user);
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('获取用户信息失败:', error);
        safeShowAlert('获取用户信息失败: ' + error.message, 'error');
    }
}

/**
 * 显示用户编辑模态框
 */
function showUserModal(user) {
    const modal = document.getElementById('user-modal');
    if (!modal) return;
    
    document.getElementById('user-modal-title').textContent = user ? '✏️ 编辑用户' : '✨ 添加用户';
    document.getElementById('user-id').value = user ? user.id : '';
    document.getElementById('user-username').value = user ? user.username : '';
    document.getElementById('user-email').value = user ? user.email : '';
    document.getElementById('user-role').value = user ? user.role : 'user';
    document.getElementById('user-storage-limit').value = user ? user.storage_limit : 100;
    document.getElementById('user-file-limit').value = user ? user.file_limit : 1000;
    document.getElementById('user-active').checked = user ? (user.status === 'active') : true;
    
    if (user) {
        // 编辑模式，密码字段不是必需的
        document.getElementById('password-hint').textContent = '留空保持原密码不变';
        document.getElementById('user-password').required = false;
        document.getElementById('user-password').value = '';
    } else {
        // 添加模式，密码是必需的
        document.getElementById('password-hint').textContent = '密码长度至少6位';
        document.getElementById('user-password').required = true;
    }
    
    modal.style.display = 'block';
}

/**
 * 关闭用户模态框
 */
function closeUserModal() {
    const modal = document.getElementById('user-modal');
    if (modal) {
        modal.style.display = 'none';
    }
}

/**
 * 提交用户表单
 */
async function submitUserForm(event) {
    event.preventDefault();
    
    const userId = document.getElementById('user-id').value;
    const username = document.getElementById('user-username').value.trim();
    const email = document.getElementById('user-email').value.trim();
    const password = document.getElementById('user-password').value;
    const role = document.getElementById('user-role').value;
    const storageLimit = parseInt(document.getElementById('user-storage-limit').value) || 0;
    const fileLimit = parseInt(document.getElementById('user-file-limit').value) || 0;
    const status = document.getElementById('user-active').checked ? 'active' : 'inactive';
    
    // 清除之前的错误提示
    clearFormErrors();
    
    // 验证输入
    const errors = [];
    
    if (!username) {
        errors.push({ field: 'user-username', message: '用户名不能为空' });
    } else if (username.length < 3) {
        errors.push({ field: 'user-username', message: '用户名长度至少3位' });
    } else if (!/^[a-zA-Z0-9_-]+$/.test(username)) {
        errors.push({ field: 'user-username', message: '用户名只能包含字母、数字、下划线和连字符' });
    }
    
    if (!userId && !password) {
        errors.push({ field: 'user-password', message: '添加用户时密码不能为空' });
    } else if (password && password.length < 6) {
        errors.push({ field: 'user-password', message: '密码长度至少6位' });
    }
    
    if (email && !isValidEmail(email)) {
        errors.push({ field: 'user-email', message: '邮箱格式不正确' });
    }
    
    if (storageLimit < 0) {
        errors.push({ field: 'user-storage-limit', message: '存储限制不能为负数' });
    }
    
    if (fileLimit < 0) {
        errors.push({ field: 'user-file-limit', message: '文件数量限制不能为负数' });
    }
    
    // 显示验证错误
    if (errors.length > 0) {
        displayFormErrors(errors);
        return;
    }
    
    // 显示提交状态
    const submitBtn = event.target.querySelector('button[type="submit"]');
    const originalText = submitBtn.innerHTML;
    submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 保存中...';
    submitBtn.disabled = true;
    
    try {
        const data = {
            username: username,
            nickname: username, // 如果没有昵称字段，使用用户名作为昵称
            is_admin: role === 'admin',
            is_active: status === 'active'
        };
        
        // 只在有邮箱时才添加邮箱字段
        if (email) {
            data.email = email;
        }
        
        if (password) {
            data.password = password;
        }
        
        let result;
        if (userId) {
            // 编辑用户
            result = await apiRequest(`/admin/users/${userId}`, {
                method: 'PUT',
                body: JSON.stringify(data)
            });
        } else {
            // 添加用户
            result = await apiRequest('/admin/users', {
                method: 'POST',
                body: JSON.stringify(data)
            });
        }
        
        if (result.code === 200) {
            closeUserModal();
            refreshUsers();
            safeShowAlert(userId ? '用户信息更新成功' : '用户添加成功', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('保存用户失败:', error);
        
        // 根据错误类型显示不同的提示
        let errorMessage = error.message;
        if (errorMessage.includes('用户名已存在')) {
            displayFormErrors([{ field: 'user-username', message: '该用户名已被使用' }]);
        } else if (errorMessage.includes('邮箱已') || errorMessage.includes('email')) {
            displayFormErrors([{ field: 'user-email', message: '该邮箱已被使用' }]);
        } else {
            safeShowAlert('保存用户失败: ' + errorMessage, 'error');
        }
    } finally {
        // 恢复提交按钮状态
        submitBtn.innerHTML = originalText;
        submitBtn.disabled = false;
    }
}

/**
 * 清除表单错误提示
 */
function clearFormErrors() {
    const errorElements = document.querySelectorAll('.form-error');
    errorElements.forEach(el => el.remove());
    
    const fieldElements = document.querySelectorAll('.form-control.error');
    fieldElements.forEach(el => el.classList.remove('error'));
}

/**
 * 显示表单错误提示
 */
function displayFormErrors(errors) {
    errors.forEach(error => {
        const field = document.getElementById(error.field);
        if (field) {
            field.classList.add('error');
            
            // 添加错误消息
            const errorEl = document.createElement('div');
            errorEl.className = 'form-error';
            errorEl.textContent = error.message;
            errorEl.style.color = '#dc3545';
            errorEl.style.fontSize = '12px';
            errorEl.style.marginTop = '4px';
            
            field.parentNode.appendChild(errorEl);
        }
    });
    
    // 显示总体错误提示
    if (errors.length > 0) {
        safeShowAlert(`表单验证失败，请检查${errors.length}个错误`, 'warning');
    }
}

/**
 * 重置用户密码
 */
async function resetUserPassword(userId) {
    const newPassword = prompt('请输入新密码（至少6位）:');
    if (!newPassword) {
        return;
    }
    
    if (newPassword.length < 6) {
        safeShowAlert('密码长度至少6位', 'warning');
        return;
    }
    
    try {
        const result = await apiRequest(`/admin/users/${userId}`, {
            method: 'PUT',
            body: JSON.stringify({ password: newPassword })
        });
        
        if (result.code === 200) {
            safeShowAlert('密码重置成功', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('重置密码失败:', error);
        safeShowAlert('重置密码失败: ' + error.message, 'error');
    }
}

/**
 * 禁用用户
 */
async function disableUser(userId) {
    if (!confirm('确定要禁用这个用户吗？用户将无法登录。')) {
        return;
    }
    
    try {
        const result = await apiRequest(`/admin/users/${userId}/status`, {
            method: 'PUT',
            body: JSON.stringify({ is_active: false })
        });
        
        if (result.code === 200) {
            refreshUsers();
            safeShowAlert('用户已禁用', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('禁用用户失败:', error);
        safeShowAlert('禁用用户失败: ' + error.message, 'error');
    }
}

/**
 * 启用用户
 */
async function enableUser(userId) {
    try {
        const result = await apiRequest(`/admin/users/${userId}/status`, {
            method: 'PUT',
            body: JSON.stringify({ is_active: true })
        });
        
        if (result.code === 200) {
            refreshUsers();
            safeShowAlert('用户已启用', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('启用用户失败:', error);
        safeShowAlert('启用用户失败: ' + error.message, 'error');
    }
}

/**
 * 删除用户
 */
async function deleteUser(userId) {
    if (!confirm('确定要删除这个用户吗？\n\n此操作将删除用户的所有数据，包括上传的文件，且不可恢复。')) {
        return;
    }
    
    // 二次确认
    const confirmText = prompt('请输入 "DELETE" 来确认删除操作:');
    if (confirmText !== 'DELETE') {
        safeShowAlert('删除操作已取消', 'info');
        return;
    }
    
    try {
        const result = await apiRequest(`/admin/users/${userId}`, {
            method: 'DELETE'
        });
        
        if (result.code === 200) {
            refreshUsers();
            safeShowAlert('用户删除成功', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('删除用户失败:', error);
        safeShowAlert('删除用户失败: ' + error.message, 'error');
    }
}

/**
 * 导出用户列表
 */
async function exportUsers() {
    try {
        safeShowAlert('正在导出用户列表...', 'info');
        
        const currentAuthToken = window.authToken || localStorage.getItem('admin_token');
        const response = await fetch('/admin/users/export', {
            headers: {
                'Authorization': `Bearer ${currentAuthToken}`
            }
        });
        
        if (!response.ok) {
            throw new Error(`导出失败: ${response.status} ${response.statusText}`);
        }
        
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = url;
        a.download = `users_export_${new Date().toISOString().split('T')[0]}.csv`;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
        
        safeShowAlert('用户列表导出成功', 'success');
    } catch (error) {
        console.error('导出用户列表失败:', error);
        safeShowAlert('导出用户列表失败: ' + error.message, 'error');
    }
}

// 工具函数

/**
 * HTML转义
 */
function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

/**
 * 验证邮箱格式
 */
function isValidEmail(email) {
    const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return re.test(email);
}

/**
 * 格式化相对时间
 */
function formatRelativeTime(dateString) {
    if (!dateString) return '';
    
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now - date;
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
    
    if (diffDays === 0) return '今天';
    if (diffDays === 1) return '昨天';
    if (diffDays < 7) return `${diffDays}天前`;
    if (diffDays < 30) return `${Math.floor(diffDays / 7)}周前`;
    if (diffDays < 365) return `${Math.floor(diffDays / 30)}个月前`;
    return `${Math.floor(diffDays / 365)}年前`;
}

// 事件监听器
document.addEventListener('click', function(e) {
    // 点击模态框外部关闭
    if (e.target.classList.contains('modal')) {
        closeUserModal();
    }
    
    // 点击下拉菜单外部关闭
    if (!e.target.closest('.dropdown')) {
        const menus = document.querySelectorAll('.dropdown-menu');
        menus.forEach(menu => menu.classList.remove('show'));
    }
    
    // 更新复选框状态
    if (e.target.classList.contains('user-checkbox')) {
        updateSelectAllState();
    }
});

// 键盘事件监听
document.addEventListener('keydown', function(e) {
    // ESC键关闭模态框
    if (e.key === 'Escape') {
        closeUserModal();
    }
    
    // 回车键搜索
    if (e.key === 'Enter' && e.target.id === 'user-search-input') {
        searchUsers();
    }
});

// 将函数暴露到全局作用域
window.initUserInterface = initUserInterface;
window.loadUsers = loadUsers;
window.searchUsers = searchUsers;
window.clearUserSearch = clearUserSearch;
window.filterUsers = filterUsers;
window.refreshUsers = refreshUsers;
window.toggleSelectAllUsers = toggleSelectAllUsers;
window.batchEnableUsers = batchEnableUsers;
window.batchDisableUsers = batchDisableUsers;
window.batchDeleteUsers = batchDeleteUsers;
window.toggleUserActions = toggleUserActions;
window.showAddUserModal = showAddUserModal;
window.viewUserFiles = viewUserFiles;
window.editUser = editUser;
window.closeUserModal = closeUserModal;
window.submitUserForm = submitUserForm;
window.resetUserPassword = resetUserPassword;
window.disableUser = disableUser;
window.enableUser = enableUser;
window.deleteUser = deleteUser;
window.exportUsers = exportUsers;
