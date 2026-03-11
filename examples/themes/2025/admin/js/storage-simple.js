// 存储管理功能

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

// 本地变量（避免与全局变量冲突）
let storageInfo = null;
let currentSelectedStorageType = null;

/**
 * 初始化存储管理界面
 */
function initStorageInterface() {
    // 注意：不在这里立即加载存储信息
    // 数据加载将在用户登录并切换到存储标签页时进行
    console.log('存储管理界面已初始化，等待认证后加载数据');
}

/**
 * 加载存储信息
 */
async function loadStorageInfo() {
    // 检查是否已认证
    if (!authToken && !window.authToken) {
        console.log('未认证，跳过存储信息加载');
        return;
    }
    
    try {
        const result = await apiRequest('/admin/storage');
        
        if (result.code === 200) {
            storageInfo = result.data;
            displayStorageInfo(result.data);
            safeShowAlert('存储信息加载成功', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('加载存储信息失败:', error);
        // 只在已认证的情况下显示错误提示
        if (authToken || window.authToken) {
            safeShowAlert('加载存储信息失败: ' + error.message, 'error');
            displayStorageError(error.message);
        }
    }
}

/**
 * 显示存储信息
 */
function displayStorageInfo(data) {
    // 更新当前存储状态
    updateCurrentStorageDisplay(data);
    
    // 更新存储卡片
    updateStorageCards(data);
    
    // 填充配置表单
    fillStorageConfigForms(data.storage_config);
}

/**
 * 更新当前存储显示
 */
function updateCurrentStorageDisplay(data) {
    const currentStorageContainer = document.getElementById('current-storage-display');
    if (!currentStorageContainer) return;
    const currentType = data && data.current ? data.current : null;
    const storageDetails = data && data.storage_details ? data.storage_details : {};
    const detail = currentType ? (storageDetails[currentType] || {}) : {};

    const typeNames = {
        'local': '本地存储',
        'webdav': 'WebDAV存储',
        'nfs': 'NFS网络存储',
        's3': 'S3对象存储'
    };

    const available = Boolean(detail.available);
    const usage = detail.usage_percent !== undefined ? detail.usage_percent : null;
    const storagePath = detail.storage_path || detail.path || '';

    const html = `
        <div class="current-storage-card">
            <div class="current-storage-overview">
                <div class="current-storage-label">
                    <div class="current-storage-icon"><i class="fas fa-hdd"></i></div>
                    <div class="current-storage-title">
                        <h4>当前存储</h4>
                        <p>${typeNames[currentType] || (currentType || '未配置')}</p>
                    </div>
                </div>
                <span class="current-storage-chip ${available ? 'chip-success' : 'chip-error'}">
                    <i class="fas fa-${available ? 'check-circle' : 'exclamation-circle'}"></i>
                    ${available ? '正常运行' : '当前异常'}
                </span>
            </div>
            <div class="current-storage-grid">
                <div class="current-storage-item">
                    <span class="item-label">存储类型</span>
                    <span class="item-value">${typeNames[currentType] || (currentType || '未配置')}</span>
                </div>
                <div class="current-storage-item">
                    <span class="item-label">运行状态</span>
                    <span class="item-value ${available ? 'status-ok' : 'status-error'}">${available ? '正常' : '异常'}</span>
                </div>
                <div class="current-storage-item">
                    <span class="item-label">存储路径</span>
                    <span class="item-value item-path" title="${storagePath || '未配置'}">${storagePath || '未配置'}</span>
                </div>
                ${usage !== null ? `
                <div class="current-storage-item">
                    <span class="item-label">使用率</span>
                    <span class="item-value status-info">${usage}%</span>
                </div>` : ''}
                ${!available && (detail.error || '') ? `
                <div class="current-storage-item current-storage-alert">
                    <span class="item-label"><i class="fas fa-exclamation-triangle"></i> 异常信息</span>
                    <span class="item-value">${detail.error || '存储连接异常'}</span>
                </div>` : ''}
            </div>
        </div>
    `;

    currentStorageContainer.innerHTML = html;
}

/**
 * 更新存储卡片
 */
function updateStorageCards(data) {
    const currentType = data && data.current ? data.current : null;
    const storageDetails = data && data.storage_details ? data.storage_details : {};

    // 更新每个存储卡片的状态
    Object.keys(storageDetails).forEach(type => {
        const card = document.getElementById(`${type}-storage-card`);
        if (!card) return;

        const detail = storageDetails[type] || {};

        // 移除所有状态类
        card.classList.remove('current-storage', 'storage-available', 'storage-unavailable');

        // 添加当前状态类
        const isCurrent = type === currentType;
        if (isCurrent) {
            card.classList.add('current-storage');
        }

        if (detail.available) {
            card.classList.add('storage-available');
        } else {
            card.classList.add('storage-unavailable');
        }

        // 更新状态徽章
        const statusBadge = card.querySelector('.storage-status-badge');
        if (statusBadge) {
            const available = Boolean(detail.available);
            if (isCurrent) {
                if (available) {
                    statusBadge.className = 'storage-status-badge status-current';
                    statusBadge.innerHTML = `
                        <i class="fas fa-star"></i>
                        当前使用
                    `;
                } else {
                    statusBadge.className = 'storage-status-badge status-error';
                    statusBadge.innerHTML = `
                        <i class="fas fa-exclamation-circle"></i>
                        当前异常
                    `;
                }
            } else {
                statusBadge.className = `storage-status-badge ${available ? 'status-success' : 'status-error'}`;
                statusBadge.innerHTML = `
                    <i class="fas fa-${available ? 'check-circle' : 'exclamation-circle'}"></i>
                    ${available ? '可用' : '不可用'}
                `;
            }
        }

        // 更新错误信息与显示路径/usage
        const errorDisplay = card.querySelector('.storage-error-display');
        if (errorDisplay) {
            if (!detail.available && detail.error) {
                errorDisplay.style.display = 'block';
                errorDisplay.innerHTML = `<i class="fas fa-exclamation-triangle"></i> ${detail.error}`;
            } else {
                errorDisplay.style.display = 'none';
            }
        }

        // 在卡片底部显示路径和使用率（防御性）
        let metaEl = card.querySelector('.storage-meta');
        if (!metaEl) {
            metaEl = document.createElement('div');
            metaEl.className = 'storage-meta';
            card.appendChild(metaEl);
        }

        const pathText = detail.storage_path || detail.path || '未配置';
        const usageText = detail.usage_percent !== undefined ? `${detail.usage_percent}% 已用` : '';
        metaEl.innerHTML = `<div class="meta-path">路径: <strong>${pathText}</strong></div>${usageText ? `<div class="meta-usage">${usageText}</div>` : ''}`;
    });
}

/**
 * 填充存储配置表单
 */
function fillStorageConfigForms(storageConfig) {
    if (!storageConfig) return;
    
    // 本地存储配置
    const localPath = document.getElementById('local-storage-path');
    if (localPath) {
        localPath.value = storageConfig.storage_path || '';
    }
    
    // WebDAV存储配置
    if (storageConfig.webdav) {
        const webdavHostname = document.getElementById('webdav-hostname');
        const webdavUsername = document.getElementById('webdav-username');
        const webdavPassword = document.getElementById('webdav-password');
        const webdavRootPath = document.getElementById('webdav-root-path');
        
        if (webdavHostname) webdavHostname.value = storageConfig.webdav.webdav_hostname || '';
        if (webdavUsername) webdavUsername.value = storageConfig.webdav.webdav_username || '';
        if (webdavPassword) webdavPassword.value = ''; // 不回显密码
        if (webdavRootPath) webdavRootPath.value = storageConfig.webdav.webdav_root_path || '';
    }
    
    // NFS存储配置
    if (storageConfig.nfs) {
        const nfsServer = document.getElementById('nfs-server');
        const nfsPath = document.getElementById('nfs-path');
        const nfsMountPoint = document.getElementById('nfs-mount-point');
        const nfsVersion = document.getElementById('nfs-version');
        const nfsOptions = document.getElementById('nfs-options');
        
        if (nfsServer) nfsServer.value = storageConfig.nfs.nfs_server || '';
        if (nfsPath) nfsPath.value = storageConfig.nfs.nfs_path || '';
        if (nfsMountPoint) nfsMountPoint.value = storageConfig.nfs.nfs_mount_point || '';
        if (nfsVersion) nfsVersion.value = storageConfig.nfs.nfs_version || '4';
        if (nfsOptions) nfsOptions.value = storageConfig.nfs.nfs_options || '';
    }
    
    // S3存储配置
    if (storageConfig.s3) {
        const s3AccessKeyID = document.getElementById('s3-access-key-id');
        const s3SecretAccessKey = document.getElementById('s3-secret-access-key');
        const s3BucketName = document.getElementById('s3-bucket-name');
        const s3EndpointURL = document.getElementById('s3-endpoint-url');
        const s3RegionName = document.getElementById('s3-region-name');
        
        if (s3AccessKeyID) s3AccessKeyID.value = storageConfig.s3.s3_access_key_id || '';
        if (s3SecretAccessKey) s3SecretAccessKey.value = ''; // 不回显密钥
        if (s3BucketName) s3BucketName.value = storageConfig.s3.s3_bucket_name || '';
        if (s3EndpointURL) s3EndpointURL.value = storageConfig.s3.s3_endpoint_url || '';
        if (s3RegionName) s3RegionName.value = storageConfig.s3.s3_region_name || '';
    }
}

/**
 * 选择存储卡片
 */
function selectStorageCard(type) {
    // 移除其他卡片的选中状态
    document.querySelectorAll('.storage-card').forEach(card => {
        card.classList.remove('selected');
    });
    
    // 添加选中状态
    const selectedCard = document.getElementById(`${type}-storage-card`);
    if (selectedCard) {
        selectedCard.classList.add('selected');
    }
    
    currentSelectedStorageType = type;
    
    // 显示配置区域和操作按钮
    showStorageActions(type);
}

/**
 * 显示存储操作按钮
 */
function showStorageActions(type) {
    const actionsContainer = document.getElementById('storage-actions');
    if (!actionsContainer) return;
    
    const isCurrentType = storageInfo && storageInfo.current === type;
    
    actionsContainer.innerHTML = `
        <button onclick="toggleStorageConfig('${type}')" class="btn btn-info">
            <i class="fas fa-cog"></i> 配置 ${getStorageTypeName(type)}
        </button>
        <button onclick="testStorageConnection('${type}')" class="btn btn-success">
            <i class="fas fa-link"></i> 测试连接
        </button>
        ${!isCurrentType ? `
            <button onclick="confirmStorageSwitch('${type}')" class="btn btn-warning">
                <i class="fas fa-exchange-alt"></i> 切换到此存储
            </button>
        ` : `
            <span class="current-storage-indicator">
                <i class="fas fa-check-circle"></i> 当前使用的存储
            </span>
        `}
    `;
}

/**
 * 获取存储类型名称
 */
function getStorageTypeName(type) {
    const names = {
        'local': '本地存储',
        'webdav': 'WebDAV存储',
        'nfs': 'NFS网络存储',
        's3': 'S3对象存储'
    };
    return names[type] || type;
}

/**
 * 切换存储配置显示
 */
function toggleStorageConfig(type) {
    const configPanel = document.getElementById(`${type}-config-panel`);
    if (!configPanel) return;
    
    // 隐藏其他配置面板
    document.querySelectorAll('.storage-config-panel').forEach(panel => {
        if (panel.id !== `${type}-config-panel`) {
            panel.style.display = 'none';
        }
    });
    
    // 切换当前面板
    if (configPanel.style.display === 'none' || !configPanel.style.display) {
        configPanel.style.display = 'block';
    } else {
        configPanel.style.display = 'none';
    }
}

/**
 * 测试存储连接
 */
async function testStorageConnection(type) {
    try {
        safeShowAlert('正在测试存储连接...', 'info');
        
        const result = await apiRequest(`/admin/storage/test/${type}`, {
            method: 'GET'
        });
        
        if (result.code === 200) {
            safeShowAlert(`${getStorageTypeName(type)}连接测试成功`, 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('测试存储连接失败:', error);
        safeShowAlert(`${getStorageTypeName(type)}连接测试失败: ` + error.message, 'error');
    }
}

/**
 * 保存存储配置
 */
async function saveStorageConfig(type) {
    try {
        safeShowAlert('正在保存配置...', 'info');
        
        const config = getStorageConfigByType(type);
        
        const result = await apiRequest('/admin/storage/config', {
            method: 'PUT',
            body: JSON.stringify({
                storage_type: type,
                config: config
            })
        });
        
        if (result.code === 200) {
            safeShowAlert(`${getStorageTypeName(type)}配置保存成功`, 'success');
            // 重新加载存储信息
            await loadStorageInfo();
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('保存存储配置失败:', error);
        safeShowAlert('保存存储配置失败: ' + error.message, 'error');
    }
}

/**
 * 根据类型获取存储配置
 */
function getStorageConfigByType(type) {
    const config = {};
    
    switch (type) {
        case 'local':
            const localPath = document.getElementById('local-storage-path');
            if (localPath) {
                config.storage_path = localPath.value;
            }
            break;
            
        case 'webdav':
            const webdavHostname = document.getElementById('webdav-hostname');
            const webdavUsername = document.getElementById('webdav-username');
            const webdavPassword = document.getElementById('webdav-password');
            const webdavRootPath = document.getElementById('webdav-root-path');
            
            if (webdavHostname) config.hostname = webdavHostname.value;
            if (webdavUsername) config.username = webdavUsername.value;
            if (webdavPassword && webdavPassword.value) config.password = webdavPassword.value;
            if (webdavRootPath) config.root_path = webdavRootPath.value;
            break;
            
        case 'nfs':
            const nfsServer = document.getElementById('nfs-server');
            const nfsPath = document.getElementById('nfs-path');
            const nfsMountPoint = document.getElementById('nfs-mount-point');
            const nfsVersion = document.getElementById('nfs-version');
            const nfsOptions = document.getElementById('nfs-options');
            
            if (nfsServer) config.server = nfsServer.value;
            if (nfsPath) config.nfs_path = nfsPath.value;
            if (nfsMountPoint) config.mount_point = nfsMountPoint.value;
            if (nfsVersion) config.version = nfsVersion.value;
            if (nfsOptions) config.options = nfsOptions.value;
            break;
            
        case 's3':
            const s3AccessKeyID = document.getElementById('s3-access-key-id');
            const s3SecretAccessKey = document.getElementById('s3-secret-access-key');
            const s3BucketName = document.getElementById('s3-bucket-name');
            const s3EndpointURL = document.getElementById('s3-endpoint-url');
            const s3RegionName = document.getElementById('s3-region-name');
            
            if (s3AccessKeyID) config.access_key_id = s3AccessKeyID.value;
            if (s3SecretAccessKey && s3SecretAccessKey.value) config.secret_access_key = s3SecretAccessKey.value;
            if (s3BucketName) config.bucket_name = s3BucketName.value;
            if (s3EndpointURL) config.endpoint_url = s3EndpointURL.value;
            if (s3RegionName) config.region_name = s3RegionName.value;
            break;
    }
    
    return config;
}

/**
 * 确认切换存储
 */
function confirmStorageSwitch(type) {
    if (!confirm(`确定要切换到${getStorageTypeName(type)}吗？这将影响新上传文件的存储位置。`)) {
        return;
    }
    
    switchStorage(type);
}

/**
 * 切换存储
 */
async function switchStorage(type) {
    try {
        safeShowAlert('正在切换存储...', 'info');
        
        const result = await apiRequest('/admin/storage/switch', {
            method: 'POST',
            body: JSON.stringify({
                storage_type: type
            })
        });
        
        if (result.code === 200) {
            safeShowAlert(`成功切换到${getStorageTypeName(type)}`, 'success');
            // 重新加载存储信息
            await loadStorageInfo();
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('切换存储失败:', error);
        safeShowAlert('切换存储失败: ' + error.message, 'error');
    }
}

/**
 * 显示存储错误
 */
function displayStorageError(error) {
    const currentStorageContainer = document.getElementById('current-storage-display');
    if (!currentStorageContainer) return;
    
    currentStorageContainer.innerHTML = `
        <div class="storage-error-card">
            <i class="fas fa-exclamation-triangle"></i>
            <h4>加载存储信息失败</h4>
            <p>${error}</p>
            <button onclick="loadStorageInfo()" class="btn btn-primary">
                <i class="fas fa-redo"></i> 重新加载
            </button>
        </div>
    `;
}

// 将函数暴露到全局作用域
window.initStorageInterface = initStorageInterface;
window.loadStorageInfo = loadStorageInfo;
window.selectStorageCard = selectStorageCard;
window.toggleStorageConfig = toggleStorageConfig;
window.testStorageConnection = testStorageConnection;
window.saveStorageConfig = saveStorageConfig;
window.confirmStorageSwitch = confirmStorageSwitch;
