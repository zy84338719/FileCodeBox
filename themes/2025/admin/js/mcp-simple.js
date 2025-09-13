// MCP服务器管理模块

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

/**
 * 加载MCP配置
 */
async function loadMCPConfig() {
    // 检查是否已认证
    if (!authToken && !window.authToken) {
        console.log('未认证，跳过MCP配置加载');
        return;
    }
    
    try {
        const result = await apiRequest('/admin/mcp/config');
        
        if (result.code === 200) {
            const config = result.data;
            fillMCPConfigForm(config);
            safeShowAlert('MCP配置加载成功', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('加载MCP配置失败:', error);
        // 只在已认证的情况下显示错误提示
        if (authToken || window.authToken) {
            safeShowAlert('加载MCP配置失败: ' + error.message, 'error');
        }
    }
}

/**
 * 填充MCP配置表单
 */
function fillMCPConfigForm(config) {
    try {
        // 启用状态
        const enableMCPCheckbox = document.getElementById('enable_mcp_server');
        if (enableMCPCheckbox) {
            enableMCPCheckbox.checked = config.enable_mcp_server === 1;
        }
        
        // 端口配置
        const portField = document.getElementById('mcp_port');
        if (portField) {
            portField.value = config.mcp_port || '8081';
        }
        
        // 主机配置
        const hostField = document.getElementById('mcp_host');
        if (hostField) {
            hostField.value = config.mcp_host || '0.0.0.0';
        }
        
        // 更新端口显示
        updatePortDisplay(config.mcp_port || '8081');
        
        // 切换配置选项显示
        toggleMCPConfigOptions();
        
        console.log('MCP配置表单填充完成');
    } catch (error) {
        console.error('填充MCP配置表单失败:', error);
        safeShowAlert('填充MCP配置表单失败: ' + error.message, 'error');
    }
}

/**
 * 更新端口显示
 */
function updatePortDisplay(port) {
    const portDisplay = document.getElementById('display-port');
    if (portDisplay) {
        portDisplay.textContent = port || '8081';
    }
}

/**
 * 切换MCP配置选项显示
 */
function toggleMCPConfigOptions() {
    const enableMCP = document.getElementById('enable_mcp_server');
    const mcpOptions = document.getElementById('mcp-config-options');
    
    if (enableMCP && mcpOptions) {
        mcpOptions.style.display = enableMCP.checked ? 'block' : 'none';
    }
}

/**
 * 保存MCP配置
 */
async function saveMCPConfig(e) {
    if (e) e.preventDefault();
    
    try {
        safeShowAlert('正在保存MCP配置...', 'info');
        
        // 获取表单数据
        const mcpConfig = {
            enable_mcp_server: getMCPCheckboxValue('enable_mcp_server') ? 1 : 0,
            mcp_port: getMCPFieldValue('mcp_port') || '8081',
            mcp_host: getMCPFieldValue('mcp_host') || '0.0.0.0'
        };
        
        console.log('准备提交的MCP配置:', mcpConfig);
        
        const result = await apiRequest('/admin/mcp/config', {
            method: 'PUT',
            body: JSON.stringify(mcpConfig)
        });
        
        if (result.code === 200) {
            safeShowAlert('MCP配置保存成功！', 'success');
            // 重新加载状态
            await loadMCPStatus();
        } else {
            throw new Error(result.message || '保存失败');
        }
    } catch (error) {
        console.error('保存MCP配置失败:', error);
        safeShowAlert('保存MCP配置失败: ' + error.message, 'error');
    }
}

/**
 * 获取MCP表单字段值
 */
function getMCPFieldValue(fieldId) {
    const field = document.getElementById(fieldId);
    return field ? field.value.trim() : '';
}

/**
 * 获取MCP复选框值
 */
function getMCPCheckboxValue(fieldId) {
    const field = document.getElementById(fieldId);
    return field ? field.checked : false;
}

/**
 * 加载MCP状态
 */
async function loadMCPStatus() {
    // 检查是否已认证
    if (!authToken && !window.authToken) {
        console.log('未认证，跳过MCP状态加载');
        return;
    }
    
    try {
        const result = await apiRequest('/admin/mcp/status');
        
        if (result.code === 200) {
            const status = result.data;
            displayMCPStatus(status);
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('加载MCP状态失败:', error);
        // 只在已认证的情况下显示错误提示
        if (authToken || window.authToken) {
            safeShowAlert('加载MCP状态失败: ' + error.message, 'error');
        }
    }
}

/**
 * 显示MCP状态
 */
function displayMCPStatus(status) {
    const statusContainer = document.getElementById('mcp-status-display');
    if (!statusContainer) return;
    
    const isRunning = status.running;
    const config = status.config || {};
    
    const statusHtml = `
        <div class="status-card">
            <div class="status-row">
                <span class="status-label">服务器状态</span>
                <span class="status-badge ${isRunning ? 'status-running' : 'status-stopped'}">
                    <i class="fas ${isRunning ? 'fa-play-circle' : 'fa-stop-circle'}"></i>
                    ${isRunning ? '运行中' : '已停止'}
                </span>
            </div>
            <div class="status-row">
                <span class="status-label">启用状态</span>
                <span class="status-badge ${config.enabled ? 'status-enabled' : 'status-disabled'}">
                    ${config.enabled ? '已启用' : '已禁用'}
                </span>
            </div>
            <div class="status-row">
                <span class="status-label">监听端口</span>
                <span class="status-value">${config.port || '8081'}</span>
            </div>
            <div class="status-row">
                <span class="status-label">绑定地址</span>
                <span class="status-value">${config.host || '0.0.0.0'}</span>
            </div>
            ${status.timestamp ? `
            <div class="status-row">
                <span class="status-label">状态更新时间</span>
                <span class="status-value">${formatDateTime(status.timestamp)}</span>
            </div>
            ` : ''}
        </div>
        
        <div class="control-buttons">
            ${config.enabled ? `
                <button onclick="controlMCPServer('${isRunning ? 'stop' : 'start'}')" 
                        class="btn ${isRunning ? 'btn-danger' : 'btn-success'}" 
                        id="mcp-control-btn">
                    <i class="fas ${isRunning ? 'fa-stop' : 'fa-play'}"></i>
                    ${isRunning ? '停止服务器' : '启动服务器'}
                </button>
                <button onclick="restartMCPServer()" class="btn btn-warning">
                    <i class="fas fa-redo"></i> 重启服务器
                </button>
            ` : `
                <p class="text-muted">
                    <i class="fas fa-info-circle"></i>
                    MCP服务器未启用，请先在配置中启用后保存配置
                </p>
            `}
        </div>
    `;
    
    statusContainer.innerHTML = statusHtml;
}

/**
 * 控制MCP服务器（启动/停止）
 */
async function controlMCPServer(action) {
    try {
        const button = document.getElementById('mcp-control-btn');
        if (button) {
            button.disabled = true;
            button.innerHTML = `<i class="fas fa-spinner fa-spin"></i> ${action === 'start' ? '启动中...' : '停止中...'}`;
        }
        
        const result = await apiRequest('/admin/mcp/control', {
            method: 'POST',
            body: JSON.stringify({ action: action })
        });
        
        if (result.code === 200) {
            safeShowAlert(`MCP服务器${action === 'start' ? '启动' : '停止'}成功`, 'success');
            // 等待一会再刷新状态
            setTimeout(() => {
                loadMCPStatus();
            }, 1000);
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error(`${action}MCP服务器失败:`, error);
        safeShowAlert(`${action === 'start' ? '启动' : '停止'}MCP服务器失败: ` + error.message, 'error');
        
        // 重新加载状态以恢复按钮
        setTimeout(() => {
            loadMCPStatus();
        }, 1000);
    }
}

/**
 * 重启MCP服务器
 */
async function restartMCPServer() {
    if (!confirm('确定要重启MCP服务器吗？这会暂时中断正在进行的连接。')) {
        return;
    }
    
    try {
        const result = await apiRequest('/admin/mcp/restart', {
            method: 'POST'
        });
        
        if (result.code === 200) {
            safeShowAlert('MCP服务器重启成功', 'success');
            // 等待一会再刷新状态
            setTimeout(() => {
                loadMCPStatus();
            }, 2000);
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('重启MCP服务器失败:', error);
        safeShowAlert('重启MCP服务器失败: ' + error.message, 'error');
    }
}

/**
 * 测试MCP连接
 */
async function testMCPConnection() {
    try {
        safeShowAlert('正在测试MCP连接...', 'info');
        
        // 获取当前配置
        const port = getMCPFieldValue('mcp_port') || '8081';
        const host = getMCPFieldValue('mcp_host') || '0.0.0.0';
        
        // 这里可以添加简单的端口连通性测试
        // 由于是前端，我们只能通过后端API来测试
        const result = await apiRequest('/admin/mcp/test', {
            method: 'POST',
            body: JSON.stringify({
                port: port,
                host: host
            })
        });
        
        if (result.code === 200) {
            safeShowAlert('MCP连接测试成功', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('测试MCP连接失败:', error);
        safeShowAlert('测试MCP连接失败: ' + error.message, 'warning');
    }
}

/**
 * 初始化MCP管理界面
 */
function initMCPInterface() {
    // 绑定启用状态切换事件
    const enableMCP = document.getElementById('enable_mcp_server');
    if (enableMCP) {
        enableMCP.addEventListener('change', toggleMCPConfigOptions);
    }
    
    // 绑定端口字段变化事件
    const portField = document.getElementById('mcp_port');
    if (portField) {
        portField.addEventListener('input', function() {
            updatePortDisplay(this.value);
        });
    }
    
    // 绑定表单提交事件
    const form = document.getElementById('mcp-config-form');
    if (form) {
        form.addEventListener('submit', saveMCPConfig);
    }
    
    // 注意：不在这里立即加载配置和状态
    // 数据加载将在用户登录并切换到MCP标签页时进行
}

/**
 * 格式化日期时间
 */
function formatDateTime(dateString) {
    if (!dateString) return '-';
    try {
        const date = new Date(dateString);
        return date.toLocaleString('zh-CN');
    } catch (error) {
        return dateString;
    }
}

// 将函数暴露到全局作用域
window.loadMCPConfig = loadMCPConfig;
window.saveMCPConfig = saveMCPConfig;
window.toggleMCPConfigOptions = toggleMCPConfigOptions;
window.loadMCPStatus = loadMCPStatus;
window.controlMCPServer = controlMCPServer;
window.restartMCPServer = restartMCPServer;
window.testMCPConnection = testMCPConnection;
window.initMCPInterface = initMCPInterface;
window.updatePortDisplay = updatePortDisplay;
