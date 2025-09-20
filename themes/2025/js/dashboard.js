// 用户仪表板模块 - 处理用户仪表板相关功能

/**
 * 仪表板管理器
 */
const Dashboard = {
    // 分页配置
    currentPage: 1,
    pageSize: 20,

    // Helper: 安全解析 JSON
    async parseJsonSafe(response) {
        try {
            return await response.json();
        } catch (err) {
            console.error('[dashboard] 解析 JSON 失败:', err);
            return null;
        }
    },

    // Helper: 处理认证相关返回（401/403）
    handleAuthError(result) {
        if (!result) return false;
        if (result.code === 401 || result.code === 403) {
            // 清理本地登录信息并提示重新登录
            UserAuth.removeToken();
            UserAuth.removeUserInfo();
            UserAuth.updateUI();
            this.showLoginPrompt();
            return true;
        }
        return false;
    },
    
    /**
     * 初始化仪表板
     */
    async init() {
        // 如果有 token 但缺少 user_info，先尝试在初始化阶段拉取用户信息（自愈），最多重试3次
        const token = UserAuth.getToken();
        if (token && !UserAuth.getUserInfo()) {
            console.log('[dashboard] 检测到 token 存在但 user_info 缺失，开始最多 3 次尝试拉取用户信息');
            let success = false;
            for (let attempt = 1; attempt <= 3; attempt++) {
                try {
                    console.log(`[dashboard] 拉取 user_info 尝试 #${attempt}`);
                    const userInfo = await this.loadUserInfo();
                    if (userInfo) {
                        success = true;
                        break;
                    }
                } catch (err) {
                    console.error('[dashboard] 尝试拉取 user_info 时出错:', err);
                }
                // 指数退避等待
                await new Promise(res => setTimeout(res, 300 * attempt));
            }
            if (!success) {
                console.warn('[dashboard] 多次尝试后仍无法获取 user_info');
                this.showProfileRetryPrompt();
            }
        }

        // 认证检查（如果没有 token，会在页面内显示登录提示）
        if (!this.checkAuth()) return;
        
        const userInfo = UserAuth.getUserInfo();
        if (userInfo) {
            this.updateUserDisplay(userInfo);
            
            // 如果是管理员，显示管理后台按钮
            if (userInfo.role === 'admin') {
                const adminBtn = document.getElementById('admin-btn');
                if (adminBtn) {
                    adminBtn.style.display = 'inline-block';
                }
            }
        }
        
        // 加载仪表板数据
        this.loadDashboard();
        
        // 设置功能模块
        this.setupFileUpload();
        this.setupForms();
    },
    
    /**
     * 检查认证状态
     */
    checkAuth() {
        const token = UserAuth.getToken();
        if (!token) {
            // 不再直接重定向到登录页，避免在某些环境下导致页面闪现为空白。
            // 改为在页面内显示友好的登录提示，用户可以点击跳转登录。
            this.showLoginPrompt();
            return false;
        }
        return true;
    },

    /**
     * 在页面中间显示登录提示（当用户未登录或 token 缺失时）
     */
    showLoginPrompt() {
        try {
            const container = document.querySelector('.container') || document.body;
            // 避免重复创建
            if (document.getElementById('dashboard-login-prompt')) return;

            const prompt = document.createElement('div');
            prompt.id = 'dashboard-login-prompt';
            prompt.style.position = 'fixed';
            prompt.style.left = '50%';
            prompt.style.top = '50%';
            prompt.style.transform = 'translate(-50%, -50%)';
            prompt.style.zIndex = '9999';
            prompt.style.background = 'rgba(255,255,255,0.96)';
            prompt.style.padding = '24px 32px';
            prompt.style.borderRadius = '8px';
            prompt.style.boxShadow = '0 6px 20px rgba(0,0,0,0.12)';
            prompt.style.textAlign = 'center';
            prompt.innerHTML = `
                <h3 style="margin:0 0 8px 0;">您尚未登录</h3>
                <p style="margin:0 0 12px 0;color:#666;">要访问用户中心，请先登录账户。</p>
                <div>
                    <button id="dashboard-login-btn" class="btn" style="margin-right:8px;">去登录</button>
                    <button id="dashboard-refresh-btn" class="btn btn-secondary">刷新页面</button>
                </div>
            `;

            container.appendChild(prompt);

            document.getElementById('dashboard-login-btn').addEventListener('click', () => {
                window.location.href = '/user/login';
            });
            document.getElementById('dashboard-refresh-btn').addEventListener('click', () => {
                window.location.reload();
            });
        } catch (err) {
            console.error('显示登录提示失败:', err);
        }
    },
    
    /**
     * 更新用户显示信息
     */
    updateUserDisplay(userInfo) {
        const userNameElement = document.getElementById('user-name');
        const userAvatarElement = document.getElementById('user-avatar');
        
        if (userNameElement) {
            userNameElement.textContent = userInfo.nickname || userInfo.username;
        }
        if (userAvatarElement) {
            userAvatarElement.textContent = (userInfo.nickname || userInfo.username).charAt(0).toUpperCase();
        }
    },
    
    /**
     * 从API加载用户信息并保存到localStorage
     */
    async loadUserInfo() {
        try {
            const response = await fetch('/user/profile', {
                headers: UserAuth.getAuthHeaders()
            });
            const result = await this.parseJsonSafe(response);
            if (this.handleAuthError(result)) return null;
            if (result && result.code === 200 && result.data) {
                const userInfo = result.data;
                UserAuth.setUserInfo(userInfo);
                // 更新 UI 状态以反映登录状态
                UserAuth.updateUI();
                console.log('[dashboard] 已获取并保存 user_info');
                return userInfo;
            } else {
                console.warn('[dashboard] /user/profile 返回结构非预期:', result);
                return null;
            }
        } catch (error) {
            console.error('获取用户信息失败:', error);
        }
        return null;
    },

    /**
     * 当拉取 user_info 多次失败时，提供一个可操作提示（重试或重新登录）
     */
    showProfileRetryPrompt() {
        try {
            const container = document.querySelector('.container') || document.body;
            // 避免重复创建
            if (document.getElementById('dashboard-profile-retry')) return;

            const prompt = document.createElement('div');
            prompt.id = 'dashboard-profile-retry';
            prompt.style.position = 'fixed';
            prompt.style.left = '50%';
            prompt.style.top = '60%';
            prompt.style.transform = 'translate(-50%, -50%)';
            prompt.style.zIndex = '9999';
            prompt.style.background = 'rgba(255,255,255,0.96)';
            prompt.style.padding = '16px 20px';
            prompt.style.borderRadius = '6px';
            prompt.style.boxShadow = '0 6px 20px rgba(0,0,0,0.12)';
            prompt.style.textAlign = 'center';
            prompt.innerHTML = `
                <div style="margin-bottom:8px;color:#333;">获取用户信息失败</div>
                <div style="margin-bottom:12px;color:#666;font-size:13px;">系统检测到你已登录（token 存在），但无法获取到账户信息，可能是网络或会话问题。</div>
                <div>
                    <button id="dashboard-retry-profile" class="btn" style="margin-right:8px;">重试获取用户信息</button>
                    <button id="dashboard-rel-login" class="btn btn-secondary">重新登录</button>
                </div>
            `;

            container.appendChild(prompt);

            document.getElementById('dashboard-retry-profile').addEventListener('click', async () => {
                document.getElementById('dashboard-profile-retry').remove();
                console.log('[dashboard] 用户触发重试获取 user_info');
                await this.loadUserInfo();
                const ui = UserAuth.getUserInfo();
                if (ui) {
                    this.updateUserDisplay(ui);
                    this.loadDashboard();
                } else {
                    // 如果仍失败，重新展示提示
                    this.showProfileRetryPrompt();
                }
            });

            document.getElementById('dashboard-rel-login').addEventListener('click', () => {
                // 清理本地登录信息并跳转登录页
                UserAuth.removeToken();
                UserAuth.removeUserInfo();
                window.location.href = '/user/login';
            });
        } catch (err) {
            console.error('显示 profile 重试提示失败:', err);
        }
    },

    /**
     * 切换标签页
     */
    switchTab(tabName, event) {
        // 移除所有active类
        document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
        
        // 添加active类到当前标签
        if (event && event.target) {
            event.target.classList.add('active');
        }
        const tabContent = document.getElementById(tabName + '-content');
        if (tabContent) {
            tabContent.classList.add('active');
        }
        
        // 根据标签页加载相应内容
        switch(tabName) {
            case 'dashboard':
                this.loadDashboard();
                break;
            case 'files':
                this.loadMyFiles();
                break;
            case 'profile':
                this.loadProfile();
                break;
        }
    },
    
    /**
     * 跳转到管理后台
     */
    goToAdmin() {
        const token = UserAuth.getToken();
        if (token) {
            window.location.href = '/admin/';
        } else {
            alert('请先登录');
            window.location.href = '/user/login';
        }
    },
    
    /**
     * 加载仪表板数据
     */
    async loadDashboard() {
        try {
            const response = await fetch('/user/stats', {
                headers: UserAuth.getAuthHeaders()
            });
            const result = await this.parseJsonSafe(response);
            if (this.handleAuthError(result)) return;
            if (result && result.code === 200 && result.data) {
                const stats = result.data;
                this.updateStatsCards(stats);
                this.updateStorageProgress(stats);
            } else {
                console.warn('[dashboard] /user/stats 返回非预期结果:', result);
            }
        } catch (error) {
            console.error('加载仪表板数据失败:', error);
        }
    },
    
    /**
     * 更新统计卡片
     */
    updateStatsCards(stats) {
        const statsGrid = document.getElementById('stats-grid');
        if (!statsGrid) return;
        
        statsGrid.innerHTML = `
            <div class="stat-card">
                <div class="stat-icon">📄</div>
                <div class="stat-value">${stats.current_files}</div>
                <div class="stat-label">总文件数</div>
            </div>
            <div class="stat-card">
                <div class="stat-icon">📤</div>
                <div class="stat-value">${stats.total_uploads}</div>
                <div class="stat-label">总上传数</div>
            </div>
            <div class="stat-card">
                <div class="stat-icon">📥</div>
                <div class="stat-value">${stats.total_downloads}</div>
                <div class="stat-label">总下载次数</div>
            </div>
            <div class="stat-card">
                <div class="stat-icon">💾</div>
                <div class="stat-value">${formatFileSize(stats.total_storage)}</div>
                <div class="stat-label">已用存储</div>
            </div>
        `;
    },
    
    /**
     * 更新存储进度条
     */
    updateStorageProgress(stats) {
        const storageProgress = document.getElementById('storage-progress');
        const storageText = document.getElementById('storage-text');
        
        if (storageProgress && storageText) {
            const storagePercent = (stats.total_storage / stats.max_storage_quota) * 100 || 0;
            storageProgress.style.width = storagePercent + '%';
            storageText.textContent = 
                `${formatFileSize(stats.total_storage)} / ${formatFileSize(stats.max_storage_quota)} (${storagePercent.toFixed(1)}%)`;
        }
    },
    
    /**
     * 加载我的文件
     */
    async loadMyFiles(page = 1) {
        try {
            const response = await fetch(`/user/files?page=${page}&page_size=${this.pageSize}`, {
                headers: UserAuth.getAuthHeaders()
            });
            const result = await this.parseJsonSafe(response);
            if (this.handleAuthError(result)) return;
            if (result && result.code === 200 && result.data) {
                const files = result.data.files || [];
                const pagination = result.data.pagination || { page: 1, total_pages: 1, total: 0 };
                this.renderFilesList(files, pagination);
            } else {
                console.warn('[dashboard] /user/files 返回非预期结果:', result);
            }
        } catch (error) {
            console.error('加载文件列表失败:', error);
        }
    },
    
    /**
     * 渲染文件列表
     */
    renderFilesList(files, pagination) {
        const filesList = document.getElementById('files-list');
        if (!filesList) return;
        
        if (files.length === 0) {
            filesList.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">📁</div>
                    <p>还没有上传任何文件</p>
                    <p style="color: #9ca3af; font-size: 14px;">点击下方按钮开始上传您的第一个文件</p>
                    <a href="#" class="btn" onclick="Dashboard.switchTab('upload', event); return false;">📤 立即上传</a>
                </div>
            `;
            return;
        }
        
        let tableHTML = this.generateFilesTable(files);
        
        // 添加分页
        if (pagination.total_pages > 1) {
            tableHTML += this.generatePagination(pagination);
        }
        
        filesList.innerHTML = tableHTML;
    },
    
    /**
     * 生成文件表格
     */
    generateFilesTable(files) {
        let tableHTML = `
            <table class="file-table">
                <thead>
                    <tr>
                        <th>文件信息</th>
                        <th>提取码</th>
                        <th>大小</th>
                        <th>类型</th>
                        <th>过期时间</th>
                        <th>下载次数</th>
                        <th>操作</th>
                    </tr>
                </thead>
                <tbody>
        `;
        
        files.forEach(file => {
            const fileName = file.file_name || (file.prefix + file.suffix);
            const uploadType = file.upload_type === 'authenticated' ? '认证上传' : '匿名上传';
            const authRequired = file.require_auth ? '🔒' : '🔓';
            const fileExtension = fileName.split('.').pop().toUpperCase();
            
            // 根据文件扩展名选择图标
            const fileIcon = this.getFileIcon(fileExtension);
            
            tableHTML += `
                <tr>
                    <td>
                        <div class="file-name">
                            <span class="file-icon">${fileIcon}</span>
                            <div>
                                <div>${authRequired} ${escapeHtml(fileName)}</div>
                                <span class="file-upload-type">${uploadType}</span>
                            </div>
                        </div>
                    </td>
                    <td><span class="file-code">${file.code}</span></td>
                    <td><span class="file-size">${formatFileSize(file.size)}</span></td>
                    <td><span class="file-type">${fileExtension}</span></td>
                    <td><span class="file-expire">${formatDateTime(file.expired_at)}</span></td>
                    <td><span class="file-downloads">${file.used_count}</span></td>
                    <td>
                        <div class="file-actions">
                            <button class="btn-sm btn-info" onclick="Dashboard.copyCode('${file.code}')" title="复制提取码">
                                📋 复制
                            </button>
                            <a href="/share/download?code=${file.code}" class="btn-sm btn-success" title="下载文件">
                                📥 下载
                            </a>
                            <button class="btn-sm btn-danger" onclick="Dashboard.deleteFile('${file.id}')" title="删除文件">
                                🗑️ 删除
                            </button>
                        </div>
                    </td>
                </tr>
            `;
        });
        
        tableHTML += `
                </tbody>
            </table>
        `;
        
        return tableHTML;
    },
    
    /**
     * 根据文件扩展名获取图标
     */
    getFileIcon(extension) {
        const iconMap = {
            // 图片文件
            'JPG': '🖼️', 'JPEG': '🖼️', 'PNG': '🖼️', 'GIF': '🖼️', 'BMP': '🖼️', 'SVG': '🖼️', 'WEBP': '🖼️',
            // 文档文件
            'PDF': '📄', 'DOC': '📝', 'DOCX': '📝', 'XLS': '📊', 'XLSX': '📊', 'PPT': '📑', 'PPTX': '📑',
            'TXT': '📃', 'RTF': '📃', 'MD': '📃',
            // 代码文件
            'JS': '💻', 'HTML': '💻', 'CSS': '💻', 'PHP': '💻', 'JAVA': '💻', 'PY': '💻', 'GO': '💻',
            'CPP': '💻', 'C': '💻', 'H': '💻', 'JSON': '💻', 'XML': '💻', 'SQL': '💻',
            // 音频文件
            'MP3': '🎵', 'WAV': '🎵', 'FLAC': '🎵', 'AAC': '🎵', 'OGG': '🎵', 'M4A': '🎵',
            // 视频文件
            'MP4': '🎬', 'AVI': '🎬', 'MKV': '🎬', 'MOV': '🎬', 'WMV': '🎬', 'FLV': '🎬', 'WEBM': '🎬',
            // 压缩文件
            'ZIP': '📦', 'RAR': '📦', '7Z': '📦', 'TAR': '📦', 'GZ': '📦', 'BZ2': '📦',
            // 可执行文件
            'EXE': '⚙️', 'MSI': '⚙️', 'APP': '⚙️', 'DEB': '⚙️', 'RPM': '⚙️', 'DMG': '⚙️'
        };
        
        return iconMap[extension.toUpperCase()] || '📄';
    },
    
    /**
     * 生成分页组件
     */
    generatePagination(pagination) {
        let paginationHTML = '<div class="pagination">';
        
        // 上一页
        if (pagination.page > 1) {
            paginationHTML += `<button class="page-btn" onclick="Dashboard.loadMyFiles(${pagination.page - 1})">‹ 上一页</button>`;
        } else {
            paginationHTML += `<button class="page-btn" disabled>‹ 上一页</button>`;
        }
        
        // 页码按钮 - 智能显示
        const maxVisiblePages = 5;
        let startPage = Math.max(1, pagination.page - Math.floor(maxVisiblePages / 2));
        let endPage = Math.min(pagination.total_pages, startPage + maxVisiblePages - 1);
        
        // 调整起始页面以确保显示正确数量的页码
        if (endPage - startPage + 1 < maxVisiblePages) {
            startPage = Math.max(1, endPage - maxVisiblePages + 1);
        }
        
        // 如果不是从第1页开始，显示第1页和省略号
        if (startPage > 1) {
            paginationHTML += `<button class="page-btn" onclick="Dashboard.loadMyFiles(1)">1</button>`;
            if (startPage > 2) {
                paginationHTML += `<span class="page-ellipsis">...</span>`;
            }
        }
        
        // 显示页码
        for (let i = startPage; i <= endPage; i++) {
            const active = i === pagination.page ? 'active' : '';
            paginationHTML += `<button class="page-btn ${active}" onclick="Dashboard.loadMyFiles(${i})">${i}</button>`;
        }
        
        // 如果不是到最后一页，显示省略号和最后一页
        if (endPage < pagination.total_pages) {
            if (endPage < pagination.total_pages - 1) {
                paginationHTML += `<span class="page-ellipsis">...</span>`;
            }
            paginationHTML += `<button class="page-btn" onclick="Dashboard.loadMyFiles(${pagination.total_pages})">${pagination.total_pages}</button>`;
        }
        
        // 下一页
        if (pagination.page < pagination.total_pages) {
            paginationHTML += `<button class="page-btn" onclick="Dashboard.loadMyFiles(${pagination.page + 1})">下一页 ›</button>`;
        } else {
            paginationHTML += `<button class="page-btn" disabled>下一页 ›</button>`;
        }
        
        // 显示分页信息
        paginationHTML += `<span class="page-info">第 ${pagination.page} 页，共 ${pagination.total_pages} 页 (${pagination.total} 个文件)</span>`;
        
        paginationHTML += '</div>';
        return paginationHTML;
    },
    
    /**
     * 加载个人资料
     */
    async loadProfile() {
        try {
            const response = await fetch('/user/profile', {
                headers: UserAuth.getAuthHeaders()
            });
            const result = await this.parseJsonSafe(response);
            if (this.handleAuthError(result)) return;
            if (result && result.code === 200 && result.data) {
                const profile = result.data;
                const form = document.getElementById('profile-form');
                if (form) {
                    form.username.value = profile.username || '';
                    form.email.value = profile.email || '';
                    form.nickname.value = profile.nickname || '';
                    form.created_at.value = profile.created_at ? formatDateTime(profile.created_at) : '暂无数据';
                    form.last_login_at.value = profile.last_login_at ? formatDateTime(profile.last_login_at) : '暂无数据';
                }
            } else {
                console.warn('[dashboard] /user/profile 返回非预期结果:', result);
            }
        } catch (error) {
            console.error('加载个人资料失败:', error);
        }
    },
    
    /**
     * 复制提取码（使用utils.js中的copyToClipboard函数）
     */
    copyCode(code) {
        // 使用utils.js中的copyToClipboard函数
        if (typeof copyToClipboard === 'function') {
            const tempButton = document.createElement('button');
            tempButton.textContent = '复制';
            copyToClipboard(code, tempButton);
        } else {
            // 降级方案
            this.fallbackCopyCode(code);
        }
    },
    
    /**
     * 降级复制方案
     */
    fallbackCopyCode(code) {
        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(code).then(() => {
                showNotification('提取码已复制到剪贴板', 'success');
            }).catch(err => {
                console.error('复制失败:', err);
                alert('提取码: ' + code);
            });
        } else {
            alert('提取码: ' + code);
        }
    },
    
    /**
     * 删除文件
     */
    async deleteFile(fileId) {
        if (!confirm('确定要删除这个文件吗？')) {
            return;
        }
        
        try {
            const response = await fetch(`/user/files/${fileId}`, {
                method: 'DELETE',
                headers: UserAuth.getAuthHeaders()
            });
            const result = await this.parseJsonSafe(response);
            if (this.handleAuthError(result)) return;
            if (result && result.code === 200) {
                showNotification('文件删除成功', 'success');
                this.loadMyFiles(this.currentPage);
            } else {
                showNotification('删除失败: ' + (result && result.message ? result.message : '未知错误'), 'error');
            }
        } catch (error) {
            console.error('删除文件失败:', error);
            showNotification('删除失败: ' + error.message, 'error');
        }
    },
    
    /**
     * 设置文件上传
     */
    setupFileUpload() {
        const uploadArea = document.querySelector('.upload-area');
        const fileInput = document.getElementById('file-input');
        const uploadText = document.getElementById('upload-text');
        
        if (!uploadArea || !fileInput || !uploadText) return;
        
        // 点击选择文件
        uploadArea.addEventListener('click', () => fileInput.click());
        
        // 拖拽上传
        uploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            uploadArea.classList.add('dragover');
        });
        
        uploadArea.addEventListener('dragleave', () => {
            uploadArea.classList.remove('dragover');
        });
        
        uploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            uploadArea.classList.remove('dragover');
            
            const files = e.dataTransfer.files;
            if (files.length > 0) {
                fileInput.files = files;
                const fileSizeMB = (files[0].size / 1024 / 1024).toFixed(2);
                uploadText.textContent = `已选择: ${files[0].name} (${fileSizeMB}MB)`;
            }
        });
        
        // 文件选择
        fileInput.addEventListener('change', (e) => {
            const file = e.target.files[0];
            if (file) {
                const fileSizeMB = (file.size / 1024 / 1024).toFixed(2);
                uploadText.textContent = `已选择: ${file.name} (${fileSizeMB}MB)`;
            }
        });
    },
    
    /**
     * 设置表单提交
     */
    setupForms() {
        this.setupUploadForm();
        this.setupProfileForm();
        this.setupPasswordForm();
    },
    
    /**
     * 设置文件上传表单
     */
    setupUploadForm() {
        const uploadForm = document.getElementById('upload-form');
        if (!uploadForm) return;
        
        uploadForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const fileInput = document.getElementById('file-input');
            const file = fileInput.files[0];
            
            if (!file) {
                showNotification('请选择文件', 'error');
                return;
            }
            
            await this.handleFileUpload(e.target, file);
        });
    },
    
    /**
     * 处理文件上传
     */
    async handleFileUpload(form, file) {
        const uploadBtn = document.getElementById('upload-btn');
        const uploadProgress = document.getElementById('upload-progress');
        const uploadProgressFill = document.getElementById('upload-progress-fill');
        const uploadResult = document.getElementById('upload-result');
        
        if (!uploadBtn || !uploadProgress || !uploadProgressFill || !uploadResult) return;
        
        uploadBtn.disabled = true;
        uploadBtn.textContent = '上传中...';
        uploadProgress.style.display = 'block';
        
        const formData = new FormData();
        formData.append('file', file);
        formData.append('expire_style', form.expire_style.value);
        formData.append('expire_value', form.expire_value.value);
        formData.append('require_auth', form.require_auth.checked ? 'true' : 'false');
        
        try {
            const xhr = new XMLHttpRequest();
            
            // 上传进度
            xhr.upload.addEventListener('progress', (e) => {
                if (e.lengthComputable) {
                    const percentComplete = (e.loaded / e.total) * 100;
                    uploadProgressFill.style.width = percentComplete + '%';
                }
            });
            
            xhr.onload = () => {
                if (xhr.status === 200) {
                    const result = JSON.parse(xhr.responseText);
                    if (result.code === 200) {
                        this.showUploadSuccess(result.data, uploadResult, form);
                    } else {
                        throw new Error(result.message);
                    }
                } else {
                    throw new Error('上传失败');
                }
            };
            
            xhr.onerror = () => {
                throw new Error('网络错误');
            };
            
            xhr.open('POST', '/share/file/');
            xhr.setRequestHeader('Authorization', 'Bearer ' + UserAuth.getToken());
            xhr.send(formData);
            
        } catch (error) {
            this.showUploadError(error.message, uploadResult);
        } finally {
            uploadBtn.disabled = false;
            uploadBtn.textContent = '上传文件';
            setTimeout(() => {
                uploadProgress.style.display = 'none';
                uploadProgressFill.style.width = '0%';
            }, 1000);
        }
    },
    
    /**
     * 显示上传成功结果
     */
    showUploadSuccess(data, uploadResult, form) {
        uploadResult.innerHTML = `
            <div style="background: #d4edda; color: #155724; padding: 15px; border-radius: 5px;">
                <h4>上传成功！</h4>
                <p>提取码: <strong>${data.code}</strong></p>
                <button class="btn-sm btn-info" onclick="Dashboard.copyCode('${data.code}')">复制提取码</button>
            </div>
        `;
        
        // 重置表单
        form.reset();
        const uploadText = document.getElementById('upload-text');
        if (uploadText) {
            uploadText.textContent = '点击选择文件或拖拽到此处';
        }
        
        // 刷新统计
        this.loadDashboard();
        
        // 显示成功通知
        showNotification('文件上传成功', 'success');
    },
    
    /**
     * 显示上传错误结果
     */
    showUploadError(message, uploadResult) {
        uploadResult.innerHTML = `
            <div style="background: #f8d7da; color: #721c24; padding: 15px; border-radius: 5px;">
                上传失败: ${escapeHtml(message)}
            </div>
        `;
        showNotification('上传失败: ' + message, 'error');
    },
    
    /**
     * 设置个人资料表单
     */
    setupProfileForm() {
        const profileForm = document.getElementById('profile-form');
        if (!profileForm) return;
        
        profileForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const formData = new FormData(e.target);
            const data = {
                nickname: formData.get('nickname')
            };
            
            try {
                const response = await fetch('/user/profile', {
                    method: 'PUT',
                    headers: UserAuth.getAuthHeaders(),
                    body: JSON.stringify(data)
                });
                const result = await this.parseJsonSafe(response);
                if (this.handleAuthError(result)) return;
                if (result && result.code === 200) {
                    showNotification('资料更新成功', 'success');
                    // 更新本地存储的用户信息
                    const userInfo = UserAuth.getUserInfo();
                    if (userInfo) {
                        userInfo.nickname = data.nickname;
                        UserAuth.setUserInfo(userInfo);
                        this.updateUserDisplay(userInfo);
                    }
                } else {
                    showNotification('更新失败: ' + (result && result.message ? result.message : '未知错误'), 'error');
                }
            } catch (error) {
                showNotification('更新失败: ' + error.message, 'error');
            }
        });
    },
    
    /**
     * 设置修改密码表单
     */
    setupPasswordForm() {
        const passwordForm = document.getElementById('password-form');
        if (!passwordForm) return;
        
        passwordForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const formData = new FormData(e.target);
            const newPassword = formData.get('new_password');
            const confirmPassword = formData.get('confirm_password');
            
            if (newPassword !== confirmPassword) {
                showNotification('两次输入的新密码不一致', 'error');
                return;
            }
            
            const data = {
                old_password: formData.get('old_password'),
                new_password: newPassword
            };
            
            try {
                const response = await fetch('/user/change-password', {
                    method: 'POST',
                    headers: UserAuth.getAuthHeaders(),
                    body: JSON.stringify(data)
                });
                
                if (response.ok) {
                    showNotification('密码修改成功，请重新登录', 'success');
                    setTimeout(() => {
                        UserAuth.logout();
                    }, 2000);
                } else {
                    const result = await response.json();
                    showNotification('修改失败: ' + result.message, 'error');
                }
            } catch (error) {
                showNotification('修改失败: ' + error.message, 'error');
            }
        });
    }
};

// 全局函数，供HTML调用
window.Dashboard = Dashboard;

// 页面加载完成后初始化
window.addEventListener('load', () => {
    Dashboard.init();
});