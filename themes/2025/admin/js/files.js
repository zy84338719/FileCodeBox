// 文件管理模块
const FilesManager = {
    currentPage: 1,
    currentSearch: '',
    currentFilter: 'all',
    currentLimit: 10,  // 添加动态limit支持
    currentView: 'list',
    selectedFiles: new Set(),
    filesStats: {},
    
    // 安全调用showAlert函数
    showAlert: function(message, type = 'info', duration = 3000) {
        if (typeof window.showAlert === 'function') {
            window.showAlert(message, type, duration);
        } else {
            console.log(`[${type.toUpperCase()}] ${message}`);
        }
    }
};

// 辅助函数：更新元素内容
function updateElement(id, value) {
    const elements = document.querySelectorAll('#' + id);
    elements.forEach(element => {
        if (element) element.textContent = value;
    });
}

function initFileInterface() {
    // 移除loadFileStats调用，因为后端没有这个API
    // loadFileStats();
    loadFiles(1);
    
    const searchInput = document.getElementById('files-search-input');
    if (searchInput) {
        searchInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                searchFiles();
            }
        });
    }
    
    const selectAllCheckbox = document.getElementById('select-all-files');
    if (selectAllCheckbox) {
        selectAllCheckbox.addEventListener('change', toggleSelectAll);
    }
    
    const filterTags = document.querySelectorAll('.filter-tag');
    filterTags.forEach(tag => {
        tag.addEventListener('click', function() {
            setFileFilter(this.dataset.type);
        });
    });
    
    console.log('文件管理界面初始化完成');
}

// 注释掉loadFileStats函数，因为后端没有/admin/files/stats端点
/*
async function loadFileStats() {
    try {
        const result = await apiRequest('/admin/files/stats');
        if (result.code === 200) {
            const stats = result.data;
            FilesManager.filesStats = stats;
            
            updateElement('total-files', stats.total_files || 0);
            updateElement('total-size', formatFileSize(stats.total_size || 0));
            updateElement('today-uploads', stats.today_uploads || 0);
            updateElement('public-files', stats.public_files || 0);
        }
    } catch (error) {
        console.error('加载文件统计失败:', error);
    }
}
*/

async function loadFiles(page = 1, search = '', filter = 'all') {
    try {
        showLoading('正在加载文件列表...');
        
        const params = new URLSearchParams({
            page: page,
            limit: FilesManager.currentLimit,  // 使用动态limit
            search: search
            // 暂时移除type参数，后端不支持
        });
        
        const result = await apiRequest('/admin/files?' + params);
        
        if (result.code === 200) {
            const data = result.data;
            FilesManager.currentPage = page;
            FilesManager.currentSearch = search;
            FilesManager.currentFilter = filter;
            
            FilesManager.selectedFiles.clear();
            updateBulkActions();
            
            // 修复：使用正确的数据结构
            let filesList = data.list || [];
            const pagination = data.pagination || {};
            
            // 前端文件类型过滤（临时解决方案，因为后端不支持type参数）
            if (filter && filter !== 'all') {
                filesList = filesList.filter(file => {
                    const fileName = file.uuid_file_name || '';
                    return matchesFileType(fileName, filter);
                });
            }
            
            if (FilesManager.currentView === 'list') {
                displayFilesList(filesList);
            } else {
                displayFilesGrid(filesList);
            }
            
            updateFilesPagination(pagination.total || 0, page, FilesManager.currentLimit);
            
            // 更新分页信息显示
            updateFilesInfo(filesList, pagination, page);
            
            // 根据文件列表数据计算并更新统计信息
            updateFileStatsFromData(filesList, pagination);
            
        } else {
            throw new Error(result.message || '加载失败');
        }
    } catch (error) {
        console.error('加载文件列表失败:', error);
        FilesManager.showAlert('加载文件列表失败: ' + error.message, 'error');
    } finally {
        hideLoading();
    }
}

function displayFilesList(files) {
    const tbody = document.getElementById('files-tbody');
    if (!tbody) return;
    
    if (files.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="text-center" style="padding: 40px;">暂无文件</td></tr>';
        return;
    }
    
    tbody.innerHTML = files.map(file => {
        // 安全地从uuid_file_name中提取原始文件名
        let fileName = '未知文件';
        if (file.uuid_file_name && typeof file.uuid_file_name === 'string') {
            const match = file.uuid_file_name.match(/^[A-Za-z0-9]+-(.+)$/);
            fileName = match ? match[1] : file.uuid_file_name;
        }
        
        return '<tr>' +
            '<td><input type="checkbox" class="file-checkbox" value="' + file.ID + '" onchange="toggleFileSelection(\'' + file.ID + '\')"></td>' +
            '<td><div class="file-info">' +
                '<div class="file-icon ' + getFileTypeClass(fileName) + '"><i class="' + getFileIcon(fileName) + '"></i></div>' +
                '<div class="file-details">' +
                    '<div class="file-name" title="' + fileName + '">' + fileName + '</div>' +
                    '<div class="file-meta">' +
                        '<span class="file-code">' + file.code + '</span>' +
                        '<span class="file-date">' + formatDateTime(file.CreatedAt) + '</span>' +
                    '</div>' +
                '</div>' +
            '</div></td>' +
            '<td>' + formatFileSize(file.size) + '</td>' +
            '<td><span class="file-status ' + (file.require_auth ? 'private' : 'public') + '">' + (file.require_auth ? '私有' : '公开') + '</span></td>' +
            '<td>' + formatDateTime(file.CreatedAt) + '</td>' +
            '<td><div class="file-actions">' +
                '<button class="file-action-btn btn-view" onclick="viewFile(\'' + file.code + '\')" title="预览"><i class="fas fa-eye"></i></button>' +
                '<button class="file-action-btn btn-download" onclick="downloadFile(\'' + file.code + '\')" title="下载"><i class="fas fa-download"></i></button>' +
                '<button class="file-action-btn btn-copy" onclick="copyFileLink(\'' + file.code + '\')" title="复制链接"><i class="fas fa-copy"></i></button>' +
                '<button class="file-action-btn btn-edit" onclick="editFile(\'' + file.code + '\')" title="编辑"><i class="fas fa-edit"></i></button>' +
                '<button class="file-action-btn btn-delete" onclick="deleteFile(\'' + file.code + '\')" title="删除"><i class="fas fa-trash"></i></button>' +
            '</div></td>' +
        '</tr>';
    }).join('');
}

function displayFilesGrid(files) {
    const gridContainer = document.getElementById('files-grid-view');
    if (!gridContainer) return;
    
    if (files.length === 0) {
        gridContainer.innerHTML = '<div style="grid-column: 1 / -1; text-align: center; padding: 40px;">暂无文件</div>';
        return;
    }
    
    gridContainer.innerHTML = files.map(file => {
        // 安全地从uuid_file_name中提取原始文件名
        let fileName = '未知文件';
        if (file.uuid_file_name && typeof file.uuid_file_name === 'string') {
            const match = file.uuid_file_name.match(/^[A-Za-z0-9]+-(.+)$/);
            fileName = match ? match[1] : file.uuid_file_name;
        }
        
        const previewContent = isImageFile(fileName) ? 
            '<img src="/file/' + file.code + '" alt="' + fileName + '" onerror="this.style.display=\'none\'; this.nextElementSibling.style.display=\'flex\';">' +
            '<div class="file-card-icon ' + getFileTypeClass(fileName) + '" style="display: none;"><i class="' + getFileIcon(fileName) + '"></i></div>' :
            '<div class="file-card-icon ' + getFileTypeClass(fileName) + '"><i class="' + getFileIcon(fileName) + '"></i></div>';
            
        return '<div class="file-card" onclick="selectFileCard(\'' + file.ID + '\')">' +
            '<div class="file-card-preview">' + previewContent + '</div>' +
            '<div class="file-card-body">' +
                '<div class="file-card-name" title="' + fileName + '">' + fileName + '</div>' +
                '<div class="file-card-meta">' +
                    '<span>' + formatFileSize(file.size) + '</span>' +
                    '<span class="file-status ' + (file.require_auth ? 'private' : 'public') + '">' + (file.require_auth ? '私有' : '公开') + '</span>' +
                '</div>' +
                '<div class="file-card-actions">' +
                    '<button class="file-action-btn btn-view" onclick="event.stopPropagation(); viewFile(\'' + file.code + '\')" title="预览"><i class="fas fa-eye"></i></button>' +
                    '<button class="file-action-btn btn-download" onclick="event.stopPropagation(); downloadFile(\'' + file.code + '\')" title="下载"><i class="fas fa-download"></i></button>' +
                    '<button class="file-action-btn btn-copy" onclick="event.stopPropagation(); copyFileLink(\'' + file.code + '\')" title="复制链接"><i class="fas fa-copy"></i></button>' +
                    '<button class="file-action-btn btn-edit" onclick="event.stopPropagation(); editFile(\'' + file.code + '\')" title="编辑"><i class="fas fa-edit"></i></button>' +
                    '<button class="file-action-btn btn-delete" onclick="event.stopPropagation(); deleteFile(\'' + file.code + '\')" title="删除"><i class="fas fa-trash"></i></button>' +
                '</div>' +
            '</div>' +
        '</div>';
    }).join('');
}

function switchView(view) {
    FilesManager.currentView = view;
    
    document.querySelectorAll('.view-toggle-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    const activeBtn = document.querySelector('.view-toggle-btn[onclick*="' + view + '"]');
    if (activeBtn) activeBtn.classList.add('active');
    
    const listView = document.getElementById('files-list-view');
    const gridView = document.getElementById('files-grid-view');
    
    if (view === 'list') {
        if (listView) listView.style.display = 'block';
        if (gridView) gridView.style.display = 'none';
    } else {
        if (listView) listView.style.display = 'none';
        if (gridView) gridView.style.display = 'grid';
    }
    
    loadFiles(FilesManager.currentPage, FilesManager.currentSearch, FilesManager.currentFilter);
}

function searchFiles() {
    const searchInput = document.getElementById('files-search-input');
    const searchTerm = searchInput ? searchInput.value.trim() : '';
    loadFiles(1, searchTerm, FilesManager.currentFilter);
}

function setFileFilter(type) {
    FilesManager.currentFilter = type;
    
    document.querySelectorAll('.filter-tag').forEach(tag => {
        tag.classList.remove('active');
    });
    const activeFilter = document.querySelector('[data-type="' + type + '"]');
    if (activeFilter) activeFilter.classList.add('active');
    
    loadFiles(1, FilesManager.currentSearch, type);
}

function toggleFileSelection(fileId) {
    if (FilesManager.selectedFiles.has(fileId)) {
        FilesManager.selectedFiles.delete(fileId);
    } else {
        FilesManager.selectedFiles.add(fileId);
    }
    updateBulkActions();
}

function selectFileCard(fileId) {
    toggleFileSelection(fileId);
    
    const card = event.currentTarget;
    if (FilesManager.selectedFiles.has(fileId)) {
        card.classList.add('selected');
    } else {
        card.classList.remove('selected');
    }
}

function toggleSelectAll() {
    const checkboxes = document.querySelectorAll('.file-checkbox');
    const selectAllCheckbox = document.getElementById('select-all-files');
    
    if (selectAllCheckbox && selectAllCheckbox.checked) {
        checkboxes.forEach(checkbox => {
            checkbox.checked = true;
            FilesManager.selectedFiles.add(checkbox.value);
        });
    } else {
        checkboxes.forEach(checkbox => {
            checkbox.checked = false;
            FilesManager.selectedFiles.delete(checkbox.value);
        });
    }
    
    updateBulkActions();
}

function updateBulkActions() {
    const bulkActions = document.querySelector('.bulk-actions');
    const selectedCount = document.getElementById('selected-count');
    
    if (FilesManager.selectedFiles.size > 0) {
        if (bulkActions) bulkActions.classList.add('show');
        if (selectedCount) selectedCount.textContent = FilesManager.selectedFiles.size;
    } else {
        if (bulkActions) bulkActions.classList.remove('show');
    }
}

function updateFilesPagination(total, page, limit) {
    const paginationContainer = document.getElementById('files-pagination-container');
    const pagination = document.getElementById('files-pagination');
    
    if (!paginationContainer || !pagination) return;
    
    // 如果总数为0，隐藏分页容器
    if (total === 0) {
        paginationContainer.style.display = 'none';
        return;
    }
    
    // 显示分页容器
    paginationContainer.style.display = 'block';
    
    // 更新统计信息
    const startItem = Math.min((page - 1) * limit + 1, total);
    const endItem = Math.min(page * limit, total);
    document.getElementById('files-current-start').textContent = startItem;
    document.getElementById('files-current-end').textContent = endItem;
    document.getElementById('files-total').textContent = total;
    
    const totalPages = Math.ceil(total / limit);
    
    if (totalPages <= 1) {
        pagination.innerHTML = '';
        return;
    }
    
    let paginationHTML = '<div class="pagination-wrapper">';
    
    // 首页按钮
    if (page > 1) {
        paginationHTML += `<button onclick="loadFiles(1, '${FilesManager.currentSearch}', '${FilesManager.currentFilter}')" class="btn-page btn-page-first" title="首页">
            <i class="fas fa-angle-double-left"></i>
        </button>`;
    }
    
    // 上一页按钮
    if (page > 1) {
        paginationHTML += `<button onclick="loadFiles(${page - 1}, '${FilesManager.currentSearch}', '${FilesManager.currentFilter}')" class="btn-page btn-page-prev" title="上一页">
            <i class="fas fa-angle-left"></i> 上一页
        </button>`;
    }
    
    // 页码按钮组
    paginationHTML += '<div class="pagination-numbers">';
    
    const startPage = Math.max(1, page - 2);
    const endPage = Math.min(totalPages, page + 2);
    
    // 如果开始页码不是1，显示省略号
    if (startPage > 1) {
        paginationHTML += `<button onclick="loadFiles(1, '${FilesManager.currentSearch}', '${FilesManager.currentFilter}')" class="btn-page">1</button>`;
        if (startPage > 2) {
            paginationHTML += '<span class="pagination-ellipsis">...</span>';
        }
    }
    
    // 页码按钮
    for (let i = startPage; i <= endPage; i++) {
        const activeClass = i === page ? 'active' : '';
        paginationHTML += `<button onclick="loadFiles(${i}, '${FilesManager.currentSearch}', '${FilesManager.currentFilter}')" class="btn-page ${activeClass}" title="第${i}页">${i}</button>`;
    }
    
    // 如果结束页码不是最后一页，显示省略号
    if (endPage < totalPages) {
        if (endPage < totalPages - 1) {
            paginationHTML += '<span class="pagination-ellipsis">...</span>';
        }
        paginationHTML += `<button onclick="loadFiles(${totalPages}, '${FilesManager.currentSearch}', '${FilesManager.currentFilter}')" class="btn-page">${totalPages}</button>`;
    }
    
    paginationHTML += '</div>';
    
    // 下一页按钮
    if (page < totalPages) {
        paginationHTML += `<button onclick="loadFiles(${page + 1}, '${FilesManager.currentSearch}', '${FilesManager.currentFilter}')" class="btn-page btn-page-next" title="下一页">
            下一页 <i class="fas fa-angle-right"></i>
        </button>`;
    }
    
    // 末页按钮
    if (page < totalPages) {
        paginationHTML += `<button onclick="loadFiles(${totalPages}, '${FilesManager.currentSearch}', '${FilesManager.currentFilter}')" class="btn-page btn-page-last" title="末页">
            <i class="fas fa-angle-double-right"></i>
        </button>`;
    }
    
    // 页面跳转控件
    paginationHTML += `<div class="pagination-jump">
        <span>跳转到</span>
        <input type="number" id="files-page-jump" min="1" max="${totalPages}" value="${page}" style="width: 60px; text-align: center;">
        <button onclick="jumpToPage('files')" class="btn btn-sm">跳转</button>
    </div>`;
    
    // 页面大小选择器
    paginationHTML += `<div class="pagination-size">
        <span>每页显示</span>
        <select id="files-page-size" onchange="changePageSize('files')">
            <option value="5" ${limit === 5 ? 'selected' : ''}>5条</option>
            <option value="10" ${limit === 10 ? 'selected' : ''}>10条</option>
            <option value="20" ${limit === 20 ? 'selected' : ''}>20条</option>
            <option value="50" ${limit === 50 ? 'selected' : ''}>50条</option>
        </select>
    </div>`;
    
    paginationHTML += '</div>';
    
    pagination.innerHTML = paginationHTML;
}

/**
 * 跳转到指定页面
 */
function jumpToPage(module) {
    const input = document.getElementById(module + '-page-jump');
    if (!input) return;
    
    const page = parseInt(input.value);
    if (isNaN(page) || page < 1) {
        FilesManager.showAlert('请输入有效的页码', 'warning');
        return;
    }
    
    if (module === 'files') {
        loadFiles(page, FilesManager.currentSearch, FilesManager.currentFilter);
    }
}

/**
 * 更改页面大小
 */
function changePageSize(module) {
    const select = document.getElementById(module + '-page-size');
    if (!select) return;
    
    const newLimit = parseInt(select.value);
    if (isNaN(newLimit) || newLimit < 1) return;
    
    if (module === 'files') {
        // 更新limit参数并重新加载第一页
        FilesManager.currentLimit = newLimit;
        loadFiles(1, FilesManager.currentSearch, FilesManager.currentFilter);
    }
}

async function downloadFile(fileCode) {
    try {
        const response = await fetch('/admin/files/download?code=' + encodeURIComponent(fileCode), {
            headers: {
                'Authorization': 'Bearer ' + (authToken || '')
            }
        });
        
        if (!response.ok) {
            throw new Error('下载失败');
        }
        
        const disposition = response.headers.get('Content-Disposition');
        let filename = 'download';
        if (disposition) {
            const filenameMatch = disposition.match(/filename="(.+)"/);
            if (filenameMatch) {
                filename = filenameMatch[1];
            }
        }
        
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        a.click();
        
        window.URL.revokeObjectURL(url);
        
    } catch (error) {
        console.error('下载文件失败:', error);
        FilesManager.showAlert('下载文件失败: ' + error.message, 'error');
    }
}

async function copyFileLink(fileCode) {
    try {
        const link = window.location.origin + '/share/download?code=' + fileCode;
        await navigator.clipboard.writeText(link);
        FilesManager.showAlert('链接已复制到剪贴板', 'success');
    } catch (error) {
        console.error('复制链接失败:', error);
        FilesManager.showAlert('复制链接失败', 'error');
    }
}

async function deleteFile(fileCode) {
    if (!confirm('确定要删除这个文件吗？此操作不可恢复！')) {
        return;
    }
    
    try {
        const result = await apiRequest('/admin/files/' + encodeURIComponent(fileCode), {
            method: 'DELETE'
        });
        
        if (result.code === 200) {
            FilesManager.showAlert('文件已删除', 'success');
            loadFiles(FilesManager.currentPage, FilesManager.currentSearch, FilesManager.currentFilter);
            // loadFileStats(); // 后端未提供统计接口，已移除
        } else {
            throw new Error(result.message || '删除失败');
        }
    } catch (error) {
        console.error('删除文件失败:', error);
        FilesManager.showAlert('删除文件失败: ' + error.message, 'error');
    }
}

function getFileIcon(filename) {
    const ext = filename.split('.').pop().toLowerCase();
    const iconMap = {
        'jpg': 'fas fa-image', 'jpeg': 'fas fa-image', 'png': 'fas fa-image', 'gif': 'fas fa-image',
        'mp4': 'fas fa-video', 'avi': 'fas fa-video', 'mov': 'fas fa-video',
        'mp3': 'fas fa-music', 'wav': 'fas fa-music',
        'pdf': 'fas fa-file-pdf', 'doc': 'fas fa-file-word', 'docx': 'fas fa-file-word',
        'zip': 'fas fa-file-archive', 'rar': 'fas fa-file-archive',
        'html': 'fas fa-file-code', 'css': 'fas fa-file-code', 'js': 'fas fa-file-code'
    };
    return iconMap[ext] || 'fas fa-file';
}

function getFileTypeClass(filename) {
    const ext = filename.split('.').pop().toLowerCase();
    if (['jpg', 'jpeg', 'png', 'gif'].includes(ext)) return 'file-type-image';
    if (['mp4', 'avi', 'mov'].includes(ext)) return 'file-type-video';
    if (['mp3', 'wav'].includes(ext)) return 'file-type-audio';
    if (['pdf'].includes(ext)) return 'file-type-pdf';
    return 'file-type-other';
}

function isImageFile(filename) {
    const ext = filename.split('.').pop().toLowerCase();
    return ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg'].includes(ext);
}

function viewFile(fileCode) {
    // 直接在新标签打开下载/预览链接，由后端按类型决定返回
    window.open('/share/download?code=' + encodeURIComponent(fileCode), '_blank');
}

function editFile(fileCode) {
    FilesManager.showAlert('文件编辑功能', 'info');
}

/**
 * 判断文件是否匹配指定类型
 */
function matchesFileType(fileName, fileType) {
    if (!fileName) return false;
    
    const ext = fileName.toLowerCase().split('.').pop();
    
    switch (fileType) {
        case 'image':
            return ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'svg', 'webp', 'ico'].includes(ext);
        case 'video':
            return ['mp4', 'avi', 'mkv', 'mov', 'wmv', 'flv', 'webm', '3gp'].includes(ext);
        case 'audio':
            return ['mp3', 'wav', 'flac', 'aac', 'ogg', 'm4a', 'wma'].includes(ext);
        case 'document':
            return ['pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx', 'txt', 'rtf', 'odt'].includes(ext);
        case 'archive':
            return ['zip', 'rar', '7z', 'tar', 'gz', 'bz2', 'xz', 'dmg'].includes(ext);
        case 'code':
            return ['js', 'html', 'css', 'php', 'py', 'java', 'cpp', 'c', 'h', 'go', 'rs', 'ts', 'json', 'xml', 'sql'].includes(ext);
        case 'other':
            return !matchesFileType(fileName, 'image') && 
                   !matchesFileType(fileName, 'video') && 
                   !matchesFileType(fileName, 'audio') && 
                   !matchesFileType(fileName, 'document') && 
                   !matchesFileType(fileName, 'archive') && 
                   !matchesFileType(fileName, 'code');
        default:
            return true;
    }
}

/**
 * 根据文件列表数据更新统计信息
 */
function updateFileStatsFromData(files, pagination) {
    // 总文件数使用分页信息中的total
    const totalFiles = pagination.total || 0;
    updateElement('total-files', totalFiles);
    
    // 计算总大小（只能计算当前页面的文件大小）
    let totalSize = 0;
    let publicFiles = 0;
    let todayUploads = 0;
    
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    
    files.forEach(file => {
        if (file.size) {
            totalSize += file.size;
        }
        
        if (!file.require_auth) {
            publicFiles++;
        }
        
        if (file.CreatedAt) {
            const fileDate = new Date(file.CreatedAt);
            fileDate.setHours(0, 0, 0, 0);
            if (fileDate.getTime() === today.getTime()) {
                todayUploads++;
            }
        }
    });
    
    // 注意：总大小只是当前页面的估算值
    updateElement('total-size', formatFileSize(totalSize));
    updateElement('public-files', publicFiles);
    updateElement('today-uploads', todayUploads);
}

/**
 * 更新文件列表分页信息显示
 */
function updateFilesInfo(files, pagination, currentPage) {
    const total = pagination.total || 0;
    const pageSize = FilesManager.currentLimit;  // 使用动态limit
    
    if (total === 0) {
        updateElement('files-current-start', 0);
        updateElement('files-current-end', 0);
        updateElement('files-total', 0);
        return;
    }
    
    const start = (currentPage - 1) * pageSize + 1;
    const end = Math.min(currentPage * pageSize, total);
    
    updateElement('files-current-start', start);
    updateElement('files-current-end', end);
    updateElement('files-total', total);
}

/**
 * 安全更新DOM元素内容
 */
function updateElement(id, content) {
    const element = document.getElementById(id);
    if (element) {
        element.textContent = content;
    }
}

/**
 * 批量删除文件
 */
function bulkDelete() {
    const selectedFiles = Array.from(FilesManager.selectedFiles);
    
    if (selectedFiles.length === 0) {
        FilesManager.showAlert('请先选择要删除的文件', 'warning');
        return;
    }
    
    if (!confirm(`确定要删除选中的 ${selectedFiles.length} 个文件吗？此操作不可恢复！`)) {
        return;
    }
    
    FilesManager.showAlert('批量删除功能开发中', 'info');
    // TODO: 实现批量删除功能
    console.log('批量删除文件:', selectedFiles);
}

/**
 * 批量下载文件
 */
function bulkDownload() {
    const selectedFiles = Array.from(FilesManager.selectedFiles);
    
    if (selectedFiles.length === 0) {
        FilesManager.showAlert('请先选择要下载的文件', 'warning');
        return;
    }
    
    FilesManager.showAlert('批量下载功能开发中', 'info');
    // TODO: 实现批量下载功能
    console.log('批量下载文件:', selectedFiles);
}

/**
 * 批量切换可见性
 */
function bulkToggleVisibility() {
    const selectedFiles = Array.from(FilesManager.selectedFiles);
    
    if (selectedFiles.length === 0) {
        FilesManager.showAlert('请先选择要操作的文件', 'warning');
        return;
    }
    
    FilesManager.showAlert('批量切换可见性功能开发中', 'info');
    // TODO: 实现批量切换可见性功能
    console.log('批量切换可见性:', selectedFiles);
}

/**
 * 清除选择
 */
function clearSelection() {
    FilesManager.selectedFiles.clear();
    updateBulkActions();
    
    // 取消所有复选框的选中状态
    const checkboxes = document.querySelectorAll('.file-checkbox');
    checkboxes.forEach(checkbox => {
        checkbox.checked = false;
    });
    
    const selectAllCheckbox = document.getElementById('select-all-files');
    if (selectAllCheckbox) {
        selectAllCheckbox.checked = false;
    }
}

/**
 * 显示上传模态框
 */
function showUploadModal() {
    FilesManager.showAlert('文件上传功能开发中', 'info');
    // TODO: 实现文件上传模态框
    console.log('显示文件上传模态框');
}

// 将文件管理函数暴露到全局作用域
window.switchView = switchView;
window.searchFiles = searchFiles;
window.setFileFilter = setFileFilter;
window.toggleFileSelection = toggleFileSelection;
window.selectFileCard = selectFileCard;
window.downloadFile = downloadFile;
window.copyFileLink = copyFileLink;
window.deleteFile = deleteFile;
window.viewFile = viewFile;
window.editFile = editFile;
window.showUploadModal = showUploadModal;
window.bulkDelete = bulkDelete;
window.bulkDownload = bulkDownload;
window.bulkToggleVisibility = bulkToggleVisibility;
window.clearSelection = clearSelection;
window.jumpToPage = jumpToPage;
window.changePageSize = changePageSize;
