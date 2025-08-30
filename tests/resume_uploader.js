/**
 * FileCodeBox 断点续传上传器
 * 支持大文件分片上传、断点续传、进度跟踪等功能
 */
class ResumeUploader {
    constructor(options = {}) {
        this.baseUrl = options.baseUrl || 'http://localhost:12345';
        this.chunkSize = options.chunkSize || 1024 * 1024; // 1MB
        this.maxRetries = options.maxRetries || 3;
        this.onProgress = options.onProgress || (() => {});
        this.onComplete = options.onComplete || (() => {});
        this.onError = options.onError || (() => {});
        
        this.uploadId = null;
        this.file = null;
        this.fileHash = null;
        this.totalChunks = 0;
        this.uploadedChunks = new Set();
        this.isUploading = false;
        this.isPaused = false;
    }

    /**
     * 计算文件哈希值
     */
    async calculateFileHash(file) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = async (event) => {
                try {
                    const arrayBuffer = event.target.result;
                    const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer);
                    const hashArray = Array.from(new Uint8Array(hashBuffer));
                    const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
                    resolve(hashHex);
                } catch (error) {
                    reject(error);
                }
            };
            reader.onerror = reject;
            reader.readAsArrayBuffer(file);
        });
    }

    /**
     * 初始化上传
     */
    async initUpload(file, expireValue = 1, expireStyle = 'day') {
        this.file = file;
        this.expireValue = expireValue;
        this.expireStyle = expireStyle;
        
        try {
            // 计算文件哈希
            this.onProgress({ stage: 'hashing', progress: 0 });
            this.fileHash = await this.calculateFileHash(file);
            
            // 初始化上传
            this.onProgress({ stage: 'initializing', progress: 0 });
            const response = await fetch(`${this.baseUrl}/chunk/upload/init/`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    file_name: file.name,
                    file_size: file.size,
                    chunk_size: this.chunkSize,
                    file_hash: this.fileHash
                })
            });

            const result = await response.json();
            
            if (result.code !== 200) {
                throw new Error(result.message);
            }

            // 文件已存在（秒传）
            if (result.detail.existed) {
                this.onComplete({
                    code: result.detail.file_code,
                    mode: 'instant'
                });
                return;
            }

            // 设置上传信息
            this.uploadId = result.detail.upload_id;
            this.totalChunks = result.detail.total_chunks;
            this.uploadedChunks = new Set(result.detail.uploaded_chunks || []);

            console.log(`断点续传初始化成功:`, {
                uploadId: this.uploadId,
                totalChunks: this.totalChunks,
                uploadedChunks: this.uploadedChunks.size,
                progress: result.detail.progress
            });

            // 开始上传
            await this.startUpload();

        } catch (error) {
            this.onError(error);
        }
    }

    /**
     * 开始上传
     */
    async startUpload() {
        this.isUploading = true;
        this.isPaused = false;

        try {
            // 上传未完成的分片
            for (let chunkIndex = 0; chunkIndex < this.totalChunks; chunkIndex++) {
                if (this.isPaused || !this.isUploading) {
                    break;
                }

                if (!this.uploadedChunks.has(chunkIndex)) {
                    await this.uploadChunk(chunkIndex);
                }

                // 更新进度
                const progress = (this.uploadedChunks.size / this.totalChunks) * 100;
                this.onProgress({
                    stage: 'uploading',
                    progress: progress,
                    chunkIndex: chunkIndex,
                    totalChunks: this.totalChunks
                });
            }

            // 如果没有暂停，完成上传
            if (!this.isPaused && this.uploadedChunks.size === this.totalChunks) {
                await this.completeUpload();
            }

        } catch (error) {
            this.isUploading = false;
            this.onError(error);
        }
    }

    /**
     * 上传单个分片
     */
    async uploadChunk(chunkIndex, retryCount = 0) {
        try {
            const start = chunkIndex * this.chunkSize;
            const end = Math.min(start + this.chunkSize, this.file.size);
            const chunkBlob = this.file.slice(start, end);

            const formData = new FormData();
            formData.append('chunk', chunkBlob);

            const response = await fetch(`${this.baseUrl}/chunk/upload/chunk/${this.uploadId}/${chunkIndex}`, {
                method: 'POST',
                body: formData
            });

            const result = await response.json();

            if (result.code !== 200) {
                throw new Error(result.message);
            }

            this.uploadedChunks.add(chunkIndex);
            console.log(`分片 ${chunkIndex} 上传成功`);

        } catch (error) {
            console.error(`分片 ${chunkIndex} 上传失败:`, error);

            // 重试机制
            if (retryCount < this.maxRetries) {
                console.log(`重试分片 ${chunkIndex}, 第 ${retryCount + 1} 次重试`);
                await new Promise(resolve => setTimeout(resolve, 1000 * (retryCount + 1)));
                return this.uploadChunk(chunkIndex, retryCount + 1);
            } else {
                throw error;
            }
        }
    }

    /**
     * 完成上传
     */
    async completeUpload() {
        try {
            this.onProgress({ stage: 'completing', progress: 100 });

            const response = await fetch(`${this.baseUrl}/chunk/upload/complete/${this.uploadId}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    expire_value: this.expireValue,
                    expire_style: this.expireStyle
                })
            });

            const result = await response.json();

            if (result.code !== 200) {
                throw new Error(result.message);
            }

            this.isUploading = false;
            this.onComplete({
                code: result.detail.code,
                name: result.detail.name,
                mode: 'upload'
            });

        } catch (error) {
            this.isUploading = false;
            throw error;
        }
    }

    /**
     * 暂停上传
     */
    pause() {
        this.isPaused = true;
        console.log('上传已暂停');
    }

    /**
     * 恢复上传
     */
    async resume() {
        if (!this.uploadId) {
            throw new Error('没有可恢复的上传会话');
        }

        this.isPaused = false;
        console.log('恢复上传');

        // 获取最新状态
        await this.getUploadStatus();
        
        // 继续上传
        await this.startUpload();
    }

    /**
     * 取消上传
     */
    async cancel() {
        this.isUploading = false;
        this.isPaused = false;

        if (this.uploadId) {
            try {
                await fetch(`${this.baseUrl}/chunk/upload/cancel/${this.uploadId}`, {
                    method: 'DELETE'
                });
                console.log('上传已取消');
            } catch (error) {
                console.error('取消上传失败:', error);
            }

            this.uploadId = null;
            this.uploadedChunks.clear();
        }
    }

    /**
     * 获取上传状态
     */
    async getUploadStatus() {
        if (!this.uploadId) {
            return null;
        }

        try {
            const response = await fetch(`${this.baseUrl}/chunk/upload/status/${this.uploadId}`);
            const result = await response.json();

            if (result.code === 200) {
                this.uploadedChunks = new Set(result.detail.uploaded_chunks || []);
                return result.detail;
            }

            return null;
        } catch (error) {
            console.error('获取上传状态失败:', error);
            return null;
        }
    }

    /**
     * 验证分片
     */
    async verifyChunk(chunkIndex, chunkHash) {
        try {
            const response = await fetch(`${this.baseUrl}/chunk/upload/verify/${this.uploadId}/${chunkIndex}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    chunk_hash: chunkHash
                })
            });

            const result = await response.json();
            return result.code === 200 && result.detail.valid;
        } catch (error) {
            console.error('验证分片失败:', error);
            return false;
        }
    }
}

// 使用示例
/*
const uploader = new ResumeUploader({
    baseUrl: 'http://localhost:12345',
    chunkSize: 1024 * 1024, // 1MB
    onProgress: (info) => {
        console.log('上传进度:', info);
        // 更新UI进度条
    },
    onComplete: (result) => {
        console.log('上传完成:', result);
        // 显示分享码
    },
    onError: (error) => {
        console.error('上传失败:', error);
        // 显示错误信息
    }
});

// 开始上传
const fileInput = document.getElementById('file-input');
fileInput.addEventListener('change', (event) => {
    const file = event.target.files[0];
    if (file) {
        uploader.initUpload(file, 7, 'day'); // 7天过期
    }
});

// 暂停上传
document.getElementById('pause-btn').addEventListener('click', () => {
    uploader.pause();
});

// 恢复上传
document.getElementById('resume-btn').addEventListener('click', () => {
    uploader.resume();
});

// 取消上传
document.getElementById('cancel-btn').addEventListener('click', () => {
    uploader.cancel();
});
*/

export default ResumeUploader;
