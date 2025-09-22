// 传输日志独立管理页面逻辑

(function () {
    const state = {
        page: 1,
        pageSize: 20,
        totalPages: 1,
        operation: '',
        search: '',
        initialized: false
    };

    function formatDuration(ms) {
        if (ms === undefined || ms === null || ms < 0) {
            return '—';
        }
        if (ms < 1000) {
            return `${ms} ms`;
        }
        return `${(ms / 1000).toFixed(2)} s`;
    }

    function formatDateTime(value) {
        if (!value) {
            return '—';
        }
        const date = new Date(value);
        if (Number.isNaN(date.getTime())) {
            return value;
        }
        return date.toLocaleString('zh-CN');
    }

    function ensureBindings() {
        if (state.initialized) {
            return;
        }
        const searchInput = document.getElementById('transfer-log-search');
        if (searchInput) {
            searchInput.addEventListener('keypress', event => {
                if (event.key === 'Enter') {
                    applyFilters();
                }
            });
        }
        const operationSelect = document.getElementById('transfer-log-operation');
        if (operationSelect) {
            operationSelect.addEventListener('change', () => applyFilters());
        }
        state.initialized = true;
    }

    async function load(page = 1) {
        ensureBindings();

        state.page = page;
        const params = new URLSearchParams();
        params.set('page', page);
        params.set('page_size', state.pageSize);

        if (state.operation) {
            params.set('operation', state.operation);
        }
        if (state.search) {
            params.set('search', state.search);
        }

        try {
            const result = await apiRequest(`/admin/logs/transfer?${params.toString()}`);
            if (result.code === 200 && result.data) {
                render(result.data);
            } else {
                throw new Error(result.message || '获取传输日志失败');
            }
        } catch (error) {
            console.error('加载传输日志失败:', error);
            const message = (error && error.message) ? error.message.substring(0, 200) : '未知错误';
            showAlert('加载传输日志失败: ' + message, 'error');
        }
    }

    function render(data) {
        const tbody = document.getElementById('transfer-log-table-body');
        const pagination = document.getElementById('transfer-log-pagination');

        if (!tbody) {
            return;
        }

        const logs = Array.isArray(data.logs) ? data.logs : [];
        tbody.innerHTML = '';

        if (!logs.length) {
            tbody.innerHTML = '<tr><td colspan="8" class="empty-cell">暂无记录</td></tr>';
        } else {
            logs.forEach(log => {
                const tr = document.createElement('tr');
                const operationLabel = log.operation === 'upload' ? '上传' : (log.operation === 'download' ? '下载' : (log.operation || '—'));
                const userLabel = log.username ? log.username : (log.user_id ? `用户 #${log.user_id}` : '匿名');
                const sizeLabel = log.file_size ? formatFileSize(log.file_size) : '—';

                tr.innerHTML = `
                    <td>${operationLabel}</td>
                    <td>${log.file_code || '—'}</td>
                    <td>${log.file_name || '—'}</td>
                    <td>${sizeLabel}</td>
                    <td>${userLabel}</td>
                    <td>${log.ip || '—'}</td>
                    <td>${formatDuration(log.duration_ms)}</td>
                    <td>${formatDateTime(log.created_at)}</td>
                `;
                tbody.appendChild(tr);
            });
        }

        if (!pagination || !data.pagination) {
            return;
        }

        const { page, pages, total, page_size: pageSize } = data.pagination;
        state.page = page || 1;
        state.totalPages = pages || 1;
        if (pageSize) {
            state.pageSize = pageSize;
        }

        if (!total) {
            pagination.style.display = 'none';
            pagination.innerHTML = '';
            return;
        }

        pagination.innerHTML = '';
        pagination.style.display = 'flex';

        const info = document.createElement('div');
        info.className = 'pagination-info';
        info.textContent = `第 ${state.page} / ${Math.max(state.totalPages, 1)} 页 · 共 ${total} 条记录`;

        const actions = document.createElement('div');
        actions.className = 'transfer-log-pagination-actions';

        const prevBtn = document.createElement('button');
        prevBtn.className = 'btn btn-secondary';
        prevBtn.innerHTML = '<i class="fas fa-arrow-left"></i> 上一页';
        prevBtn.disabled = state.page <= 1;
        prevBtn.onclick = () => changePage(-1);

        const nextBtn = document.createElement('button');
        nextBtn.className = 'btn btn-secondary';
        nextBtn.innerHTML = '下一页 <i class="fas fa-arrow-right"></i>';
        nextBtn.disabled = state.page >= state.totalPages;
        nextBtn.onclick = () => changePage(1);

        actions.appendChild(prevBtn);
        actions.appendChild(nextBtn);

        pagination.appendChild(info);
        pagination.appendChild(actions);
    }

    function applyFilters() {
        const operationSelect = document.getElementById('transfer-log-operation');
        const searchInput = document.getElementById('transfer-log-search');
        state.operation = operationSelect ? operationSelect.value : '';
        state.search = searchInput ? searchInput.value.trim() : '';
        load(1);
    }

    function refresh() {
        load(state.page || 1);
    }

    function changePage(delta) {
        const next = (state.page || 1) + delta;
        if (next < 1 || next > (state.totalPages || 1)) {
            return;
        }
        load(next);
    }

    window.initTransferLogsTab = function () {
        ensureBindings();
        load(1);
    };

    window.applyTransferLogFilters = applyFilters;
    window.refreshTransferLogs = refresh;
    window.changeTransferLogPage = changePage;
})();
