/**
 * 管理员配置API客户端
 * 支持新的结构化DTO和平面化DTO格式
 */

class AdminConfigClient {
    constructor(baseUrl = '', adminToken = '') {
        this.baseUrl = baseUrl;
        this.adminToken = adminToken;
        this.headers = {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${adminToken}`
        };
    }

    /**
     * 获取配置
     */
    async getConfig() {
        try {
            const response = await fetch(`${this.baseUrl}/admin/config`, {
                method: 'GET',
                headers: this.headers
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const result = await response.json();
            return result.data;
        } catch (error) {
            console.error('获取配置失败:', error);
            throw error;
        }
    }

    /**
     * 使用结构化DTO更新配置
     * @param {Object} configUpdate - 结构化配置更新对象
     * @example
     * updateConfigStructured({
     *   mcp: {
     *     enable_mcp_server: 1,
     *     mcp_port: "8081",
     *     mcp_host: "0.0.0.0"
     *   },
     *   user: {
     *     allow_user_registration: 1,
     *     jwt_secret: "newSecret123456"
     *   },
     *   base: {
     *     name: "新的文件快递柜",
     *     description: "更新的描述"
     *   }
     * })
     */
    async updateConfigStructured(configUpdate) {
        try {
            const response = await fetch(`${this.baseUrl}/admin/config`, {
                method: 'PUT',
                headers: this.headers,
                body: JSON.stringify(configUpdate)
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const result = await response.json();
            return result;
        } catch (error) {
            console.error('更新配置失败:', error);
            throw error;
        }
    }

    /**
     * 使用平面化格式更新配置
     * @param {Object} flatConfig - 平面化配置对象
     * @example
     * updateConfigFlat({
     *   enable_mcp_server: 1,
     *   mcp_port: "8081",
     *   allow_user_registration: 1,
     *   jwt_secret: "newSecret123456",
     *   name: "新的文件快递柜"
     * })
     */
    async updateConfigFlat(flatConfig) {
        try {
            const response = await fetch(`${this.baseUrl}/admin/config`, {
                method: 'PUT',
                headers: this.headers,
                body: JSON.stringify(flatConfig)
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const result = await response.json();
            return result;
        } catch (error) {
            console.error('更新配置失败:', error);
            throw error;
        }
    }

    /**
     * 更新MCP配置
     */
    async updateMCPConfig(enable, port = "8081", host = "0.0.0.0") {
        return this.updateConfigStructured({
            mcp: {
                enable_mcp_server: enable ? 1 : 0,
                mcp_port: port,
                mcp_host: host
            }
        });
    }

    /**
     * 更新用户系统配置
     */
    async updateUserConfig(allowRegistration, jwtSecret = null, uploadSize = null, storageQuota = null) {
        const userConfig = {
            allow_user_registration: allowRegistration ? 1 : 0
        };

        if (jwtSecret) userConfig.jwt_secret = jwtSecret;
        if (uploadSize !== null) userConfig.user_upload_size = uploadSize;
        if (storageQuota !== null) userConfig.user_storage_quota = storageQuota;

        return this.updateConfigStructured({
            user: userConfig
        });
    }

    /**
     * 更新基础配置
     */
    async updateBaseConfig(name = null, description = null, keywords = null) {
        const baseConfig = {};

        if (name) baseConfig.name = name;
        if (description) baseConfig.description = description;
        if (keywords) baseConfig.keywords = keywords;

        if (Object.keys(baseConfig).length === 0) {
            throw new Error('至少需要提供一个基础配置字段');
        }

        return this.updateConfigStructured({
            base: baseConfig
        });
    }

    /**
     * 更新传输配置
     */
    async updateTransferConfig({
        openUpload = null,
        uploadSize = null,
        enableChunk = null,
        chunkSize = null,
        enableConcurrentDownload = null,
        maxConcurrentDownloads = null,
        downloadTimeout = null
    } = {}) {
        const transferConfig = {};

        // 上传配置
        const uploadConfig = {};
        if (openUpload !== null) uploadConfig.open_upload = openUpload ? 1 : 0;
        if (uploadSize !== null) uploadConfig.upload_size = uploadSize;
        if (enableChunk !== null) uploadConfig.enable_chunk = enableChunk ? 1 : 0;
        if (chunkSize !== null) uploadConfig.chunk_size = chunkSize;

        if (Object.keys(uploadConfig).length > 0) {
            transferConfig.upload = uploadConfig;
        }

        // 下载配置
        const downloadConfig = {};
        if (enableConcurrentDownload !== null) downloadConfig.enable_concurrent_download = enableConcurrentDownload ? 1 : 0;
        if (maxConcurrentDownloads !== null) downloadConfig.max_concurrent_downloads = maxConcurrentDownloads;
        if (downloadTimeout !== null) downloadConfig.download_timeout = downloadTimeout;

        if (Object.keys(downloadConfig).length > 0) {
            transferConfig.download = downloadConfig;
        }

        if (Object.keys(transferConfig).length === 0) {
            throw new Error('至少需要提供一个传输配置字段');
        }

        return this.updateConfigStructured({
            transfer: transferConfig
        });
    }

    /**
     * 批量更新多个配置部分
     */
    async updateMultipleConfigs({
        base = null,
        transfer = null,
        user = null,
        mcp = null,
        notifyTitle = null,
        notifyContent = null,
        pageExplain = null,
        opacity = null,
        themesSelect = null
    } = {}) {
        const configUpdate = {};

        if (base) configUpdate.base = base;
        if (transfer) configUpdate.transfer = transfer;
        if (user) configUpdate.user = user;
        if (mcp) configUpdate.mcp = mcp;
        if (notifyTitle) configUpdate.notify_title = notifyTitle;
        if (notifyContent) configUpdate.notify_content = notifyContent;
        if (pageExplain) configUpdate.page_explain = pageExplain;
        if (opacity !== null) configUpdate.opacity = opacity;
        if (themesSelect) configUpdate.themes_select = themesSelect;

        if (Object.keys(configUpdate).length === 0) {
            throw new Error('至少需要提供一个配置更新');
        }

        return this.updateConfigStructured(configUpdate);
    }
}

// 使用示例
const configClient = new AdminConfigClient('http://localhost:12345', 'FileCodeBox2025');

// 示例1: 更新MCP配置
async function enableMCP() {
    try {
        const result = await configClient.updateMCPConfig(true, "8082", "127.0.0.1");
        console.log('MCP配置更新成功:', result);
    } catch (error) {
        console.error('MCP配置更新失败:', error);
    }
}

// 示例2: 更新用户配置
async function updateUserSettings() {
    try {
        const result = await configClient.updateUserConfig(
            true, // 允许用户注册
            "newJWTSecret123456789", // 新的JWT密钥
            100 * 1024 * 1024, // 100MB上传限制
            2 * 1024 * 1024 * 1024 // 2GB存储配额
        );
        console.log('用户配置更新成功:', result);
    } catch (error) {
        console.error('用户配置更新失败:', error);
    }
}

// 示例3: 批量更新配置
async function batchUpdateConfig() {
    try {
        const result = await configClient.updateMultipleConfigs({
            base: {
                name: "我的文件快递柜",
                description: "高效的文件分享平台"
            },
            user: {
                allow_user_registration: 1,
                jwt_secret: "batchUpdateSecret123456"
            },
            mcp: {
                enable_mcp_server: 1,
                mcp_port: "8083"
            },
            opacity: 80
        });
        console.log('批量配置更新成功:', result);
    } catch (error) {
        console.error('批量配置更新失败:', error);
    }
}

// 示例4: 使用平面化格式更新（向后兼容）
async function flatUpdateExample() {
    try {
        const result = await configClient.updateConfigFlat({
            enable_mcp_server: 0,
            allow_user_registration: 1,
            jwt_secret: "flatUpdateSecret123456",
            name: "平面更新测试",
            open_upload: 1,
            upload_size: 50 * 1024 * 1024 // 50MB
        });
        console.log('平面化配置更新成功:', result);
    } catch (error) {
        console.error('平面化配置更新失败:', error);
    }
}

// 导出供其他模块使用
if (typeof module !== 'undefined' && module.exports) {
    module.exports = AdminConfigClient;
}

// 浏览器环境下全局可用
if (typeof window !== 'undefined') {
    window.AdminConfigClient = AdminConfigClient;
}
