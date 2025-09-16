// 系统配置模块

/**
 * 安全的显示警告函数
 */
function safeShowAlert(message, type = 'info') {
    if (typeof window.showAlert === 'function') {
        window.showAlert(message, type);
    } else {
        console.log(`[${type.toUpperCase()}] ${message}`);
    }
}

/**
 * 加载系统配置
 */
async function loadConfig() {
    // 检查是否已认证
    if (!authToken && !window.authToken) {
        console.log('未认证，跳过配置加载');
        return;
    }
    
    try {
        const result = await apiRequest('/admin/config');
        
        if (result.code === 200) {
            const config = result.data;
            fillConfigForm(config);
            safeShowAlert('配置加载成功', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('加载配置失败:', error);
        // 只在已认证的情况下显示错误提示
        if (authToken || window.authToken) {
            safeShowAlert('加载配置失败: ' + error.message, 'error');
        }
    }
}

/**
 * 填充配置表单
 */
function fillConfigForm(config) {
    try {
        // 基础设置
        setFieldValue('base_name', config.base?.name);
        setFieldValue('base_description', config.base?.description);
        setFieldValue('base_keywords', config.base?.keywords);
        setFieldValue('admin_token', ''); // 不回显密码
        setFieldValue('notify_title', config.notify_title);
        setFieldValue('notify_content', config.notify_content);
        setFieldValue('page_explain', config.page_explain);
        
        // 上传限制设置
        setFieldValue('upload_size_mb', bytesToMB(config.transfer?.upload?.upload_size || 0));
        setFieldValue('chunk_size_mb', bytesToMB(config.transfer?.upload?.chunk_size || 0));
        setFieldValue('max_save_seconds', config.transfer?.upload?.max_save_seconds);
        setCheckboxValue('open_upload', config.transfer?.upload?.open_upload);
        setCheckboxValue('enable_chunk', config.transfer?.upload?.enable_chunk);
        
        // 性能设置
        setCheckboxValue('enable_concurrent_download', config.transfer?.download?.enable_concurrent_download);
        setFieldValue('max_concurrent_downloads', config.transfer?.download?.max_concurrent_downloads);
        setFieldValue('download_timeout', config.transfer?.download?.download_timeout);
        setFieldValue('opacity', config.opacity);
        setFieldValue('themes_select', config.themes_select);
        
        // 用户系统设置 (始终启用)
    // config.user.allow_user_registration 可能为 0/1，setCheckboxValue 接受布尔化
    setCheckboxValue('allow_user_registration', config.user?.allow_user_registration);
        setCheckboxValue('require_email_verify', config.user?.require_email_verify);
        setFieldValue('user_storage_quota_mb', bytesToMB(config.user?.user_storage_quota || 0));
        setFieldValue('user_upload_size_mb', bytesToMB(config.user?.user_upload_size || 0));
        setFieldValue('session_expiry_hours', config.user?.session_expiry_hours);
        setFieldValue('max_sessions_per_user', config.user?.max_sessions_per_user);
        
        // 用户系统始终启用，无需切换显示
        
        console.log('配置表单填充完成');
    } catch (error) {
        console.error('填充配置表单失败:', error);
        safeShowAlert('填充配置表单失败: ' + error.message, 'error');
    }
}

/**
 * 设置表单字段值
 */
function setFieldValue(fieldId, value) {
    const field = document.getElementById(fieldId);
    if (field && value !== undefined && value !== null) {
        field.value = value;
    }
}

/**
 * 设置复选框值
 */
function setCheckboxValue(fieldId, value) {
    const field = document.getElementById(fieldId);
    if (field) {
        field.checked = Boolean(value);
    }
}

/**
 * 字节转MB
 */
function bytesToMB(bytes) {
    if (!bytes) return 0;
    return Math.round(bytes / (1024 * 1024));
}

/**
 * MB转字节
 */
function mbToBytes(mb) {
    if (!mb) return 0;
    return mb * 1024 * 1024;
}

/**
 * 处理配置表单提交
 */
async function handleConfigSubmit(e) {
    e.preventDefault();
    
    try {
        safeShowAlert('正在保存配置...', 'info');
        
        // 构建配置对象
        const config = {
            base: {
                name: getFieldValue('base_name'),
                description: getFieldValue('base_description'),
                keywords: getFieldValue('base_keywords')
            },
            transfer: {
                upload: {
                    open_upload: getCheckboxValue('open_upload') ? 1 : 0,
                    upload_size: mbToBytes(getFieldValue('upload_size_mb', 'number')),
                    enable_chunk: getCheckboxValue('enable_chunk') ? 1 : 0,
                    chunk_size: mbToBytes(getFieldValue('chunk_size_mb', 'number')),
                    max_save_seconds: getFieldValue('max_save_seconds', 'number')
                },
                download: {
                    enable_concurrent_download: getCheckboxValue('enable_concurrent_download') ? 1 : 0,
                    max_concurrent_downloads: getFieldValue('max_concurrent_downloads', 'number'),
                    download_timeout: getFieldValue('download_timeout', 'number')
                }
            },
            user: {
                allow_user_registration: getCheckboxValue('allow_user_registration') ? 1 : 0,
                require_email_verify: getCheckboxValue('require_email_verify') ? 1 : 0,
                user_storage_quota: mbToBytes(getFieldValue('user_storage_quota_mb', 'number')),
                user_upload_size: mbToBytes(getFieldValue('user_upload_size_mb', 'number')),
                session_expiry_hours: getFieldValue('session_expiry_hours', 'number'),
                max_sessions_per_user: getFieldValue('max_sessions_per_user', 'number')
            },
            notify_title: getFieldValue('notify_title'),
            notify_content: getFieldValue('notify_content'),
            page_explain: getFieldValue('page_explain'),
            opacity: getFieldValue('opacity', 'number'),
            themes_select: getFieldValue('themes_select')
        };
        
        // 如果密码字段有值，添加到配置中
        const adminToken = getFieldValue('admin_token');
        if (adminToken && adminToken.trim()) {
            config.admin_token = adminToken.trim();
        }
        
        console.log('准备提交的配置:', config);
        
        const result = await apiRequest('/admin/config', {
            method: 'PUT',
            body: JSON.stringify(config)
        });
        
        if (result.code === 200) {
            safeShowAlert('配置保存成功！', 'success');
            // 清空密码字段
            setFieldValue('admin_token', '');
        } else {
            throw new Error(result.message || '保存失败');
        }
    } catch (error) {
        console.error('保存配置失败:', error);
        safeShowAlert('保存配置失败: ' + error.message, 'error');
    }
}

/**
 * 获取表单字段值
 */
function getFieldValue(fieldId, type = 'string') {
    const field = document.getElementById(fieldId);
    if (!field) return type === 'number' ? 0 : '';
    
    const value = field.value.trim();
    if (type === 'number') {
        const num = parseInt(value) || 0;
        return num;
    }
    return value;
}

/**
 * 获取复选框值
 */
function getCheckboxValue(fieldId) {
    const field = document.getElementById(fieldId);
    return field ? field.checked : false;
}

/**
 * 重置配置
 */
async function resetConfig() {
    if (!confirm('确定要重置为默认配置吗？此操作不可恢复。')) {
        return;
    }
    
    try {
        safeShowAlert('正在重置配置...', 'info');
        
        const result = await apiRequest('/admin/config/reset', {
            method: 'POST'
        });
        
        if (result.code === 200) {
            await loadConfig();
            safeShowAlert('配置已重置为默认值', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('重置配置失败:', error);
        safeShowAlert('重置配置失败，将尝试重新加载当前配置', 'warning');
        await loadConfig();
    }
}

/**
 * 初始化配置表单
 */
function initConfigForm() {
    const form = document.getElementById('config-form');
    if (form) {
        form.addEventListener('submit', handleConfigSubmit);
    }
    
    // 用户系统始终启用，无需切换事件
    
    // 注意：不在这里立即加载配置
    // 配置加载将在用户登录并切换到配置标签页时进行
    console.log('配置表单已初始化，等待认证后加载数据');
}

// 将函数暴露到全局作用域
window.loadConfig = loadConfig;
window.handleConfigSubmit = handleConfigSubmit;
window.resetConfig = resetConfig;
window.initConfigForm = initConfigForm;
