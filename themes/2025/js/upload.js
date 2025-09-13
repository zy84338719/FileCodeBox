// 文件上传模块 - 处理文件选择、拖拽、上传进度等

/**
 * 文件上传管理器
 */
const FileUpload = {
    // 上传配置
    config: {
        maxSize: 10485760, // 默认10MB
        allowedTypes: [], // 允许的文件类型
        chunkSize: 1024 * 1024, // 分片大小 1MB
        enableChunk: false // 是否启用分片上传
    },
    
    // 当前上传状态
    currentUpload: null,
    
    /**
     * 初始化文件上传
     */
    init() {
        this.setupFileInput();
        this.setupDragAndDrop();
        this.setupFormSubmit();
        this.loadConfig();
    },
    
    /**
     * 加载服务器配置
     */
    async loadConfig() {
        try {
            const response = await fetch('/', {
                method: 'POST'
            });
            const result = await response.json();
            
            if (result.code === 200) {
                this.config.maxSize = result.data.uploadSize;
                this.config.enableChunk = result.data.enableChunk;
                console.log('上传配置已加载:', this.config);
            }
        } catch (error) {
            console.error('获取配置失败:', error);
        }
    },
    
    /**
     * 设置文件选择器
     */
    setupFileInput() {
        const fileInput = document.getElementById('file-input');
        if (!fileInput) return;
        
        fileInput.addEventListener('change', (e) => {
            const file = e.target.files[0];
            if (file) {
                this.updateFileDisplay(file);
            }
        });
    },
    
    /**
     * 设置拖拽上传
     */
    setupDragAndDrop() {
        const uploadArea = document.querySelector('.upload-area');
        if (!uploadArea) return;
        
        // 点击选择文件
        uploadArea.addEventListener('click', () => {
            document.getElementById('file-input')?.click();
        });
        
        // 拖拽事件
        uploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            uploadArea.classList.add('dragover');
        });
        
        uploadArea.addEventListener('dragleave', (e) => {
            e.preventDefault();
            uploadArea.classList.remove('dragover');
        });
        
        uploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            uploadArea.classList.remove('dragover');
            
            const files = e.dataTransfer.files;
            if (files.length > 0) {
                const fileInput = document.getElementById('file-input');
                if (fileInput) {
                    fileInput.files = files;
                    this.updateFileDisplay(files[0]);
                }
            }
        });
    },
    
    /**
     * 设置表单提交
     */
    setupFormSubmit() {
        const form = document.getElementById('file-form');
        if (!form) return;
        
        form.addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleFileUpload(e);
        });
    },
    
    /**
     * 更新文件显示
     */
    updateFileDisplay(file) {
        const uploadText = document.querySelector('.upload-text');
        if (uploadText && file) {
            const fileSizeMB = (file.size / 1024 / 1024).toFixed(2);
            uploadText.textContent = `已选择: ${file.name} (${fileSizeMB}MB)`;
        }
    },
    
    /**
     * 验证文件
     */
    validateFile(file) {
        // 检查文件大小
        if (file.size > this.config.maxSize) {
            const maxSizeMB = (this.config.maxSize / 1024 / 1024).toFixed(2);
            const fileSizeMB = (file.size / 1024 / 1024).toFixed(2);
            throw new Error(`文件大小超过限制！\n文件大小: ${fileSizeMB}MB\n最大允许: ${maxSizeMB}MB\n\n请选择更小的文件或使用管理后台调整上传大小限制。`);
        }
        
        // 检查文件类型（如果配置了允许的类型）
        if (this.config.allowedTypes.length > 0 && !validateFileType(file, this.config.allowedTypes)) {
            throw new Error(`不支持的文件类型: ${getFileExtension(file.name)}`);
        }
        
        return true;
    },
    
    /**
     * 处理文件上传
     */
    async handleFileUpload(event) {
        const fileInput = document.getElementById('file-input');
        const file = fileInput?.files[0];
        
        if (!file) {
            showNotification('请选择文件', 'error');
            return;
        }
        
        try {
            // 验证文件
            this.validateFile(file);
            
            // 获取表单数据
            const formData = new FormData();
            formData.append('file', file);
            formData.append('expire_style', event.target.expire_style.value);
            formData.append('expire_value', event.target.expire_value.value);
            
            // 开始上传
            await this.uploadFile(formData, file);
            
        } catch (error) {
            showNotification(error.message, 'error');
        }
    },
    
    /**
     * 上传文件
     */
    async uploadFile(formData, file) {
        const progressContainer = document.getElementById('upload-progress');
        const progressFill = document.getElementById('progress-fill');
        const progressText = document.getElementById('progress-text');
        const uploadStatus = document.getElementById('upload-status');
        const uploadBtn = document.getElementById('upload-btn');
        
        // 显示进度条和禁用按钮
        if (progressContainer) progressContainer.classList.add('show');
        if (uploadBtn) uploadBtn.disabled = true;
        if (uploadStatus) {
            uploadStatus.textContent = '正在上传...';
            uploadStatus.className = 'upload-status status-uploading';
        }
        
        try {
            const xhr = new XMLHttpRequest();
            
            // 上传进度监听
            const uploadStartTime = Date.now();
            let lastUpdateTime = uploadStartTime;
            let smoothProgress = 0;
            
            xhr.upload.addEventListener('progress', (e) => {
                if (e.lengthComputable) {
                    const currentTime = Date.now();
                    const timeDiff = currentTime - lastUpdateTime;
                    
                    // 计算真实进度
                    const realProgress = (e.loaded / e.total) * 100;
                    
                    // 平滑进度更新，避免进度条跳跃
                    if (timeDiff > 50) { // 每50ms最多更新一次
                        // 如果真实进度比当前显示的进度快很多，加快追赶速度
                        const progressDiff = realProgress - smoothProgress;
                        if (progressDiff > 10) {
                            smoothProgress += progressDiff * 0.5; // 快速追赶
                        } else {
                            smoothProgress += progressDiff * 0.3; // 平滑更新
                        }
                        
                        // 确保进度不超过真实进度
                        smoothProgress = Math.min(smoothProgress, realProgress);
                        
                        // 更新UI
                        const displayProgress = Math.floor(smoothProgress);
                        if (progressFill) progressFill.style.width = smoothProgress + '%';
                        if (progressText) progressText.textContent = displayProgress + '%';
                        
                        // 计算上传速度
                        const speed = e.loaded / ((currentTime - uploadStartTime) / 1000);
                        const speedText = formatSpeed(speed);
                        
                        // 估算剩余时间
                        const remainingBytes = e.total - e.loaded;
                        const estimatedTime = remainingBytes / speed;
                        const timeText = formatTime(estimatedTime);
                        
                        if (uploadStatus) {
                            if (displayProgress < 100) {
                                uploadStatus.textContent = `正在上传... ${displayProgress}% (${speedText}, 剩余${timeText})`;
                            } else {
                                uploadStatus.textContent = '处理中...';
                            }
                        }
                        
                        lastUpdateTime = currentTime;
                    }
                }
            });
            
            // 响应处理
            xhr.addEventListener('load', () => {
                try {
                    const result = JSON.parse(xhr.responseText);
                    
                    if (result.code === 200) {
                        if (uploadStatus) {
                            uploadStatus.textContent = '上传成功！';
                            uploadStatus.className = 'upload-status status-success';
                        }
                        
                        // 自动复制提取码到剪贴板
                        const shareCode = result.data.code;
                        copyToClipboardAuto(shareCode);
                        
                        setTimeout(() => {
                            showResult(`
                                <h3>文件上传成功！</h3>
                                <div class="result-code">${result.data.code}</div>
                                <p>文件名: ${result.data.file_name}</p>
                                <p>文件大小: ${formatFileSize(file.size)}</p>
                                <p>✅ 提取码已自动复制到剪贴板</p>
                            `);
                            
                            // 重置表单
                            this.resetUpload();
                        }, 1000);
                    } else {
                        throw new Error(result.message || '上传失败');
                    }
                } catch (error) {
                    this.handleUploadError(error.message);
                }
            });
            
            // 错误处理
            xhr.addEventListener('error', () => {
                this.handleUploadError('网络错误，请重试');
            });
            
            // 超时处理
            xhr.addEventListener('timeout', () => {
                this.handleUploadError('上传超时，请重试');
            });
            
            // 发送请求
            xhr.timeout = 300000; // 5分钟超时
            xhr.open('POST', '/share/file/');
            
            // 添加认证头
            const token = UserAuth.getToken();
            if (token) {
                xhr.setRequestHeader('Authorization', 'Bearer ' + token);
            }
            
            xhr.send(formData);
            
        } catch (error) {
            this.handleUploadError(error.message);
        }
    },
    
    /**
     * 处理上传错误
     */
    handleUploadError(message) {
        const uploadStatus = document.getElementById('upload-status');
        const uploadBtn = document.getElementById('upload-btn');
        
        if (uploadStatus) {
            uploadStatus.textContent = '上传失败: ' + message;
            uploadStatus.className = 'upload-status status-error';
        }
        
        if (uploadBtn) {
            uploadBtn.disabled = false;
        }
        
        setTimeout(() => {
            showNotification('上传失败: ' + message, 'error');
        }, 500);
    },
    
    /**
     * 重置上传状态
     */
    resetUpload() {
        const progressContainer = document.getElementById('upload-progress');
        const progressFill = document.getElementById('progress-fill');
        const progressText = document.getElementById('progress-text');
        const uploadBtn = document.getElementById('upload-btn');
        const fileInput = document.getElementById('file-input');
        const uploadText = document.querySelector('.upload-text');
        
        if (progressContainer) progressContainer.classList.remove('show');
        if (uploadBtn) uploadBtn.disabled = false;
        if (progressFill) progressFill.style.width = '0%';
        if (progressText) progressText.textContent = '0%';
        if (fileInput) fileInput.value = '';
        if (uploadText) uploadText.textContent = '点击选择文件或拖拽到此处';
    }
};