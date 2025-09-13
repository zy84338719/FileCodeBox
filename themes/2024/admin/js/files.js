// 文件管理模块
const FilesManager = {
    currentPage: 1,
    currentSearch: '',
    currentFilter: 'all',
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
            limit: 20,
            search: search,
            type: filter
        });
        
        const result = await apiRequest('/admin/files?' + params);
        
        if (result.code === 200) {
            const data = result.data;
            FilesManager.currentPage = page;
            FilesManager.currentSearch = search;
            FilesManager.currentFilter = filter;
            
            FilesManager.selectedFiles.clear();
            updateBulkActions();
            
            if (FilesManager.currentView === 'list') {
                displayFilesList(data.files || []);
            } else {
                displayFilesGrid(data.files || []);
            }
            
            updateFilesPagination(data.total || 0, page, data.limit || 20);
            
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
        return '<tr>' +
            '<td><input type="checkbox" class="file-checkbox" value="' + file.id + '" onchange="toggleFileSelection(\'' + file.id + '\')"></td>' +
            '<td><div class="file-info">' +
                '<div class="file-icon ' + getFileTypeClass(file.name) + '"><i class="' + getFileIcon(file.name) + '"></i></div>' +
                '<div class="file-details">' +
                    '<div class="file-name" title="' + file.name + '">' + file.name + '</div>' +
                    '<div class="file-meta">' +
                        '<span class="file-code">' + file.code + '</span>' +
                        '<span class="file-date">' + formatDateTime(file.created_at) + '</span>' +
                    '</div>' +
                '</div>' +
            '</div></td>' +
            '<td>' + formatFileSize(file.size) + '</td>' +
            '<td><span class="file-status ' + (file.is_public ? 'public' : 'private') + '">' + (file.is_public ? '公开' : '私有') + '</span></td>' +
            '<td>' + formatDateTime(file.created_at) + '</td>' +
            '<td><div class="file-actions">' +
                '<button class="file-action-btn btn-view" onclick="viewFile(\'' + file.id + '\')" title="预览"><i class="fas fa-eye"></i></button>' +
                '<button class="file-action-btn btn-download" onclick="downloadFile(\'' + file.id + '\')" title="下载"><i class="fas fa-download"></i></button>' +
                '<button class="file-action-btn btn-copy" onclick="copyFileLink(\'' + file.code + '\')" title="复制链接"><i class="fas fa-copy"></i></button>' +
                '<button class="file-action-btn btn-edit" onclick="editFile(\'' + file.id + '\')" title="编辑"><i class="fas fa-edit"></i></button>' +
                '<button class="file-action-btn btn-delete" onclick="deleteFile(\'' + file.id + '\')" title="删除"><i class="fas fa-trash"></i></button>' +
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
        const previewContent = isImageFile(file.name) ? 
            '<img src="/file/' + file.code + '" alt="' + file.name + '" onerror="this.style.display=\'none\'; this.nextElementSibling.style.display=\'flex\';">' +
            '<div class="file-card-icon ' + getFileTypeClass(file.name) + '" style="display: none;"><i class="' + getFileIcon(file.name) + '"></i></div>' :
            '<div class="file-card-icon ' + getFileTypeClass(file.name) + '"><i class="' + getFileIcon(file.name) + '"></i></div>';
            
        return '<div class="file-card" onclick="selectFileCard(\'' + file.id + '\')">' +
            '<div class="file-card-preview">' + previewContent + '</div>' +
            '<div class="file-card-body">' +
                '<div class="file-card-name" title="' + file.name + '">' + file.name + '</div>' +
                '<div class="file-card-meta">' +
                    '<span>' + formatFileSize(file.size) + '</span>' +
                    '<span class="file-status ' + (file.is_public ? 'public' : 'private') + '">' + (file.is_public ? '公开' : '私有') + '</span>' +
                '</div>' +
                '<div class="file-card-actions">' +
                    '<button class="file-action-btn btn-view" onclick="event.stopPropagation(); viewFile(\'' + file.id + '\')" title="预览"><i class="fas fa-eye"></i></button>' +
                    '<button class="file-action-btn btn-download" onclick="event.stopPropagation(); downloadFile(\'' + file.id + '\')" title="下载"><i class="fas fa-download"></i></button>' +
                    '<button class="file-action-btn btn-copy" onclick="event.stopPropagation(); copyFileLink(\'' + file.code + '\')" title="复制链接"><i class="fas fa-copy"></i></button>' +
                    '<button class="file-action-btn btn-edit" onclick="event.stopPropagation(); editFile(\'' + file.id + '\')" title="编辑"><i class="fas fa-edit"></i></button>' +
                    '<button class="file-action-btn btn-delete" onclick="event.stopPropagation(); deleteFile(\'' + file.id + '\')" title="删除"><i class="fas fa-trash"></i></button>' +
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
    const totalPages = Math.ceil(total / limit);
    const pagination = document.getElementById('files-pagination');
    
    if (!pagination || totalPages <= 1) {
        if (pagination) pagination.innerHTML = '';
        return;
    }
    
    let paginationHTML = '';
    
    if (page > 1) {
        paginationHTML += '<button onclick="loadFiles(' + (page - 1) + ', \'' + FilesManager.currentSearch + '\', \'' + FilesManager.currentFilter + '\')" class="btn-page">上一页</button>';
    }
    
    const startPage = Math.max(1, page - 2);
    const endPage = Math.min(totalPages, page + 2);
    
    for (let i = startPage; i <= endPage; i++) {
        const activeClass = i === page ? 'active' : '';
        paginationHTML += '<button onclick="loadFiles(' + i + ', \'' + FilesManager.currentSearch + '\', \'' + FilesManager.currentFilter + '\')" class="btn-page ' + activeClass + '">' + i + '</button>';
    }
    
    if (page < totalPages) {
        paginationHTML += '<button onclick="loadFiles(' + (page + 1) + ', \'' + FilesManager.currentSearch + '\', \'' + FilesManager.currentFilter + '\')" class="btn-page">下一页</button>';
    }
    
    pagination.innerHTML = paginationHTML;
}

async function downloadFile(fileId) {
    try {
        const response = await fetch('/admin/files/' + fileId + '/download', {
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
        const link = window.location.origin + '/file/' + fileCode;
        await navigator.clipboard.writeText(link);
        FilesManager.showAlert('链接已复制到剪贴板', 'success');
    } catch (error) {
        console.error('复制链接失败:', error);
        FilesManager.showAlert('复制链接失败', 'error');
    }
}

async function deleteFile(fileId) {
    if (!confirm('确定要删除这个文件吗？此操作不可恢复！')) {
        return;
    }
    
    try {
        const result = await apiRequest('/admin/files/' + fileId, {
            method: 'DELETE'
        });
        
        if (result.code === 200) {
            FilesManager.showAlert('文件已删除', 'success');
            loadFiles(FilesManager.currentPage, FilesManager.currentSearch, FilesManager.currentFilter);
            loadFileStats();
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

function viewFile(fileId) {
    FilesManager.showAlert('文件预览功能', 'info');
}

function editFile(fileId) {
    FilesManager.showAlert('文件编辑功能', 'info');
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
