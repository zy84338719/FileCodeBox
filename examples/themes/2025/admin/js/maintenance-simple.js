// 简化的维护工具模块

/**
 * 系统重启
 */
async function restartSystem() {
    if (!confirm('确定要重启系统吗？这将中断所有当前连接。')) {
        return;
    }
    
    try {
        const result = await apiRequest('/admin/system/restart', {
            method: 'POST'
        });
        
        if (result.code === 200) {
            showAlert('系统重启命令已发送', 'success');
            setTimeout(() => {
                window.location.reload();
            }, 3000);
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('系统重启失败:', error);
        showAlert('系统重启失败: ' + error.message, 'error');
    }
}

/**
 * 系统关闭
 */
async function shutdownSystem() {
    if (!confirm('确定要关闭程序吗？这将停止所有服务。')) {
        return;
    }

    try {
        const result = await apiRequest('/admin/maintenance/shutdown', {
            method: 'POST'
        });

        if (result.code === 200) {
            showAlert('系统关闭指令已发送', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('系统关闭失败:', error);
        showAlert('系统关闭失败: ' + error.message, 'error');
    }
}

/**
 * 清理临时文件
 */
async function cleanTempFiles() {
    if (!confirm('确定要清理临时文件吗？')) {
        return;
    }
    
    try {
        const result = await apiRequest('/admin/maintenance/clean-temp', {
            method: 'POST'
        });
        
        if (result.code === 200) {
            // 修复：使用正确的字段名，并提供默认值
            const cleanedCount = result.data.count || 0;
            const freedSpace = result.data.freed_space || 0;
            
            if (freedSpace > 0) {
                showAlert(`清理完成，删除 ${cleanedCount} 个临时文件，释放空间: ${formatFileSize(freedSpace)}`, 'success');
            } else {
                showAlert(`清理完成，删除 ${cleanedCount} 个临时文件`, 'success');
            }
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('清理临时文件失败:', error);
        showAlert('清理临时文件失败: ' + error.message, 'error');
    }
}

/**
 * 数据库优化
 */
async function optimizeDatabase() {
    if (!confirm('确定要优化数据库吗？此操作可能需要一些时间。')) {
        return;
    }
    
    try {
        const result = await apiRequest('/admin/maintenance/db/optimize', {
            method: 'POST'
        });
        
        if (result.code === 200) {
            showAlert('数据库优化完成', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('数据库优化失败:', error);
        showAlert('数据库优化失败: ' + error.message, 'error');
    }
}

/**
 * 清理过期文件
 */
async function cleanExpiredFiles() {
    if (!confirm('确定要清理过期文件吗？此操作不可逆。')) {
        return;
    }
    
    try {
        const result = await apiRequest('/admin/maintenance/clean-expired', {
            method: 'POST'
        });
        
        if (result.code === 200) {
            const data = result.data;
            // 修复：使用正确的字段名，并提供默认值
            const deletedCount = data.cleaned_count || data.deleted_count || 0;
            const freedSpace = data.freed_space || 0;
            
            if (freedSpace > 0) {
                showAlert(`清理完成，删除 ${deletedCount} 个文件，释放空间: ${formatFileSize(freedSpace)}`, 'success');
            } else {
                showAlert(`清理完成，删除 ${deletedCount} 个文件`, 'success');
            }
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('清理过期文件失败:', error);
        showAlert('清理过期文件失败: ' + error.message, 'error');
    }
}

/**
 * 备份数据库
 */
async function backupDatabase() {
    try {
        const result = await apiRequest('/admin/maintenance/db/backup', {
            method: 'POST'
        });
        
        if (result.code === 200) {
            showAlert(`数据库备份完成: ${result.data.backup_file}`, 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('数据库备份失败:', error);
        showAlert('数据库备份失败: ' + error.message, 'error');
    }
}

/**
 * 导出系统配置
 */
async function exportConfig() {
    try {
        const result = await apiRequest('/admin/config');
        
        if (result.code === 200) {
            // 创建下载链接
            const blob = new Blob([JSON.stringify(result.data, null, 2)], {
                type: 'application/json'
            });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `filecodebox-config-${new Date().toISOString().split('T')[0]}.json`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
            
            showAlert('配置导出成功', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('导出配置失败:', error);
        showAlert('导出配置失败: ' + error.message, 'error');
    }
}

/**
 * 显示导入配置模态框
 */
function showImportConfigModal() {
    const modal = document.getElementById('import-config-modal');
    if (modal) {
        modal.style.display = 'block';
    }
}

/**
 * 导入系统配置
 */
async function importConfig() {
    const fileInput = document.getElementById('config-file');
    const file = fileInput.files[0];
    
    if (!file) {
        showAlert('请选择配置文件', 'warning');
        return;
    }
    
    try {
        const text = await file.text();
        const config = JSON.parse(text);
        
        const result = await apiRequest('/admin/config', {
            method: 'PUT',
            body: JSON.stringify(config)
        });
        
        if (result.code === 200) {
            closeImportConfigModal();
            showAlert('配置导入成功，请重启系统以生效', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('导入配置失败:', error);
        showAlert('导入配置失败: ' + error.message, 'error');
    }
}

/**
 * 关闭导入配置模态框
 */
function closeImportConfigModal() {
    const modal = document.getElementById('import-config-modal');
    if (modal) {
        modal.style.display = 'none';
    }
}

/**
 * 系统诊断
 */
async function systemDiagnosis() {
    try {
        const result = await apiRequest('/admin/maintenance/system-info');
        
        if (result.code === 200) {
            displayDiagnosisResult(result.data);
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('获取系统信息失败:', error);
        showAlert('获取系统信息失败: ' + error.message, 'error');
    }
}

/**
 * 显示诊断结果
 */
function displayDiagnosisResult(diagnosis) {
    const modal = document.getElementById('diagnosis-modal');
    const content = document.getElementById('diagnosis-content');
    
    if (!modal || !content) return;
    
    let html = '<div class="diagnosis-result">';
    
    // 系统信息
    html += '<h4>系统信息</h4>';
    html += '<ul>';
    html += `<li>版本: ${diagnosis.version || 'Unknown'}</li>`;
    html += `<li>运行时间: ${diagnosis.uptime || 'Unknown'}</li>`;
    html += `<li>内存使用: ${formatFileSize(diagnosis.memory_usage || 0)}</li>`;
    html += `<li>磁盘使用: ${formatFileSize(diagnosis.disk_usage || 0)}</li>`;
    html += '</ul>';
    
    // 检查项目
    html += '<h4>检查结果</h4>';
    if (diagnosis.checks && diagnosis.checks.length > 0) {
        html += '<ul>';
        diagnosis.checks.forEach(check => {
            const status = check.passed ? '✅' : '❌';
            html += `<li>${status} ${check.name}: ${check.message}</li>`;
        });
        html += '</ul>';
    }
    
    // 建议
    if (diagnosis.recommendations && diagnosis.recommendations.length > 0) {
        html += '<h4>建议</h4>';
        html += '<ul>';
        diagnosis.recommendations.forEach(rec => {
            html += `<li>💡 ${rec}</li>`;
        });
        html += '</ul>';
    }
    
    html += '</div>';
    
    content.innerHTML = html;
    modal.style.display = 'block';
}

/**
 * 关闭诊断模态框
 */
function closeDiagnosisModal() {
    const modal = document.getElementById('diagnosis-modal');
    if (modal) {
        modal.style.display = 'none';
    }
}

/**
 * 查看系统日志
 */
async function viewSystemLogs() {
    try {
        const result = await apiRequest('/admin/maintenance/logs');
        
        if (result.code === 200) {
            displaySystemLogs(result.data);
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('获取系统日志失败:', error);
        showAlert('获取系统日志失败: ' + error.message, 'error');
    }
}

/**
 * 显示系统日志
 */
function displaySystemLogs(logs) {
    const modal = document.getElementById('logs-modal');
    const content = document.getElementById('logs-content');
    
    if (!modal || !content) return;
    
    const html = logs.map(log => `
        <div class="log-entry ${log.level}">
            <span class="log-time">${formatDateTime(log.timestamp)}</span>
            <span class="log-level">[${log.level.toUpperCase()}]</span>
            <span class="log-message">${log.message}</span>
        </div>
    `).join('');
    
    content.innerHTML = html;
    modal.style.display = 'block';
}

/**
 * 关闭日志模态框
 */
function closeLogsModal() {
    const modal = document.getElementById('logs-modal');
    if (modal) {
        modal.style.display = 'none';
    }
}

/**
 * 清理系统日志
 */
async function clearSystemLogs() {
    if (!confirm('确定要清理系统日志吗？此操作不可逆。')) {
        return;
    }
    
    try {
        const result = await apiRequest('/admin/maintenance/logs/clear-system', {
            method: 'POST'
        });
        
        if (result.code === 200) {
            showAlert('系统日志清理完成', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('清理系统日志失败:', error);
        showAlert('清理系统日志失败: ' + error.message, 'error');
    }
}

/**
 * 清理系统日志
 */
async function clearSystemLogs() {
    if (!confirm('确定要清理系统日志吗？此操作不可逆。')) {
        return;
    }
    
    try {
        const result = await apiRequest('/admin/maintenance/logs/clear-system', {
            method: 'POST'
        });
        
        if (result.code === 200) {
            showAlert('系统日志清理完成', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('清理系统日志失败:', error);
        showAlert('清理系统日志失败: ' + error.message, 'error');
    }
}

/**
 * 数据库分析
 */
async function analyzeDatabase() {
    try {
        const result = await apiRequest('/admin/maintenance/db/analyze');
        
        if (result.code === 200) {
            showAlert('数据库分析完成', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('数据库分析失败:', error);
        showAlert('数据库分析失败: ' + error.message, 'error');
    }
}

/**
 * 清空缓存
 */
async function clearCache() {
    try {
        const result = await apiRequest('/admin/maintenance/cache/clear-system', {
            method: 'POST'
        });
        
        if (result.code === 200) {
            showAlert('缓存清理完成', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('清理缓存失败:', error);
        showAlert('清理缓存失败: ' + error.message, 'error');
    }
}

/**
 * 重建缓存
 */
async function refreshCache() {
    showAlert('缓存重建功能暂未实现', 'info');
}

/**
 * 显示缓存统计
 */
function showCacheStats() {
    showAlert('缓存统计功能暂未实现', 'info');
}

/**
 * 显示系统状态
 */
async function showSystemStatus() {
    try {
        const result = await apiRequest('/admin/maintenance/system-info');
        
        if (result.code === 200) {
            displayDiagnosisResult(result.data);
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('获取系统状态失败:', error);
        showAlert('获取系统状态失败: ' + error.message, 'error');
    }
}

/**
 * 显示性能指标
 */
async function showPerformanceMetrics() {
    try {
        const result = await apiRequest('/admin/maintenance/monitor/performance');
        
        if (result.code === 200) {
            showAlert('性能指标获取成功', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('获取性能指标失败:', error);
        showAlert('获取性能指标失败: ' + error.message, 'error');
    }
}

/**
 * 生成系统报告
 */
function generateSystemReport() {
    showAlert('系统报告生成功能暂未实现', 'info');
}

/**
 * 安全扫描
 */
async function securityScan() {
    try {
        const result = await apiRequest('/admin/maintenance/security/scan', {
            method: 'POST'
        });
        
        if (result.code === 200) {
            showAlert('安全扫描完成', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('安全扫描失败:', error);
        showAlert('安全扫描失败: ' + error.message, 'error');
    }
}

/**
 * 显示访问日志
 */
function showAccessLogs() {
    viewSystemLogs(); // 使用现有的日志查看功能
}

/**
 * 清理会话
 */
function clearSessions() {
    showAlert('会话清理功能暂未实现', 'info');
}

/**
 * 下载日志
 */
async function downloadLogs() {
    try {
        const result = await apiRequest('/admin/maintenance/logs/export');
        
        if (result.code === 200) {
            showAlert('日志导出成功', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('下载日志失败:', error);
        showAlert('下载日志失败: ' + error.message, 'error');
    }
}

/**
 * 清理旧日志
 */
function clearOldLogs() {
    clearSystemLogs(); // 使用现有的清理功能
}

/**
 * 显示日志分析
 */
function showLogAnalysis() {
    showAlert('日志分析功能暂未实现', 'info');
}

/**
 * 一键全面清理
 */
async function quickCleanAll() {
    if (!confirm('确定要执行一键全面清理吗？这将清理过期文件、临时文件和缓存。')) {
        return;
    }
    
    try {
        // 依次执行清理操作
        await cleanExpiredFiles();
        await cleanTempFiles();
        await clearCache();
        showAlert('一键全面清理完成', 'success');
    } catch (error) {
        console.error('一键清理失败:', error);
        showAlert('一键清理失败: ' + error.message, 'error');
    }
}

/**
 * 一键优化系统
 */
async function quickOptimize() {
    if (!confirm('确定要执行一键系统优化吗？这将优化数据库。')) {
        return;
    }
    
    try {
        await optimizeDatabase();
        showAlert('一键系统优化完成', 'success');
    } catch (error) {
        console.error('一键优化失败:', error);
        showAlert('一键优化失败: ' + error.message, 'error');
    }
}

/**
 * 系统健康检查
 */
function systemHealthCheck() {
    systemDiagnosis(); // 使用现有的系统诊断功能
}

/**
 * 清理无效记录
 */
async function cleanInvalidRecords() {
    if (!confirm('确定要清理无效记录吗？此操作不可逆。')) {
        return;
    }
    
    try {
        const result = await apiRequest('/admin/maintenance/clean-invalid', {
            method: 'POST'
        });
        
        if (result.code === 200) {
            showAlert('无效记录清理完成', 'success');
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('清理无效记录失败:', error);
        showAlert('清理无效记录失败: ' + error.message, 'error');
    }
}

// 将函数暴露到全局作用域
window.restartSystem = restartSystem;
window.cleanTempFiles = cleanTempFiles;
window.optimizeDatabase = optimizeDatabase;
window.cleanExpiredFiles = cleanExpiredFiles;
window.backupDatabase = backupDatabase;
window.exportConfig = exportConfig;
window.showImportConfigModal = showImportConfigModal;
window.importConfig = importConfig;
window.closeImportConfigModal = closeImportConfigModal;
window.systemDiagnosis = systemDiagnosis;
window.closeDiagnosisModal = closeDiagnosisModal;
window.viewSystemLogs = viewSystemLogs;
window.closeLogsModal = closeLogsModal;
window.clearSystemLogs = clearSystemLogs;
window.analyzeDatabase = analyzeDatabase;
window.clearCache = clearCache;
window.refreshCache = refreshCache;
window.showCacheStats = showCacheStats;
window.showSystemStatus = showSystemStatus;
window.showPerformanceMetrics = showPerformanceMetrics;
window.generateSystemReport = generateSystemReport;
window.securityScan = securityScan;
window.showAccessLogs = showAccessLogs;
window.clearSessions = clearSessions;
window.downloadLogs = downloadLogs;
window.clearOldLogs = clearOldLogs;
window.showLogAnalysis = showLogAnalysis;
window.quickCleanAll = quickCleanAll;
window.quickOptimize = quickOptimize;
window.systemHealthCheck = systemHealthCheck;
window.cleanInvalidRecords = cleanInvalidRecords;
