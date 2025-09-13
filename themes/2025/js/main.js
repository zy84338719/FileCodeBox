// 主入口文件 - 应用程序初始化和全局控制

/**
 * 应用程序状态管理
 */
const AppState = {
    initialized: false,
    currentTab: 'file',
    isLoading: false,
    config: null
};

/**
 * 应用程序主类
 */
class FileCodeBoxApp {
    constructor() {
        this.modules = [];
        this.eventListeners = [];
    }
    
    /**
     * 注册模块
     */
    registerModule(module) {
        this.modules.push(module);
    }
    
    /**
     * 初始化应用程序
     */
    async init() {
        if (AppState.initialized) {
            console.warn('应用程序已初始化');
            return;
        }
        
        console.log('初始化 FileCodeBox 应用程序...');
        
        try {
            // 设置全局错误处理
            this.setupGlobalErrorHandling();
            
            // 加载配置
            await this.loadConfig();
            
            // 初始化各个模块
            await this.initModules();
            
            // 设置全局事件监听器
            this.setupGlobalEvents();
            
            // 应用动态配置
            this.applyDynamicConfig();
            
            AppState.initialized = true;
            console.log('FileCodeBox 应用程序初始化完成');
            
        } catch (error) {
            console.error('应用程序初始化失败:', error);
            showNotification('应用程序初始化失败', 'error');
        }
    }
    
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
                AppState.config = result.data;
                console.log('应用配置已加载:', AppState.config);
            } else {
                console.warn('配置加载失败:', result.message);
            }
        } catch (error) {
            console.error('加载配置时出错:', error);
        }
    }
    
    /**
     * 初始化所有模块
     */
    async initModules() {
        // 初始化用户系统
        await UserSystem.init();
        
        // 初始化文件上传
        FileUpload.init();
        
        // 初始化分享功能
        ShareManager.init();
        
        // 初始化标签页管理
        TabManager.init();
        
        console.log('所有模块初始化完成');
    }
    
    /**
     * 设置全局事件监听器
     */
    setupGlobalEvents() {
        // 页面可见性变化
        document.addEventListener('visibilitychange', () => {
            if (document.visibilityState === 'visible') {
                // 页面变为可见时，检查用户状态
                if (UserAuth.isLoggedIn()) {
                    UserAuth.updateUI();
                }
            }
        });
        
        // 窗口大小变化
        window.addEventListener('resize', debounce(() => {
            this.handleResize();
        }, 250));
        
        // 在线/离线状态
        window.addEventListener('online', () => {
            showNotification('网络连接已恢复', 'success');
        });
        
        window.addEventListener('offline', () => {
            showNotification('网络连接已断开', 'warning');
        });
        
        // 键盘快捷键
        document.addEventListener('keydown', (e) => {
            this.handleKeyboard(e);
        });
        
        console.log('全局事件监听器已设置');
    }
    
    /**
     * 处理窗口大小变化
     */
    handleResize() {
        // 如果是移动端，调整布局
        if (isMobile()) {
            document.body.classList.add('mobile');
        } else {
            document.body.classList.remove('mobile');
        }
    }
    
    /**
     * 处理键盘事件
     */
    handleKeyboard(e) {
        // Escape键隐藏结果
        if (e.key === 'Escape') {
            hideResult();
        }
        
        // Ctrl/Cmd + V 粘贴文件（如果支持）
        if ((e.ctrlKey || e.metaKey) && e.key === 'v') {
            this.handlePaste(e);
        }
    }
    
    /**
     * 处理粘贴事件
     */
    async handlePaste(e) {
        // 检查当前是否在文本输入框中
        const activeElement = document.activeElement;
        if (activeElement && (activeElement.tagName === 'INPUT' || activeElement.tagName === 'TEXTAREA')) {
            return; // 让默认的粘贴行为处理
        }
        
        // 检查剪贴板是否有文件
        if (navigator.clipboard && navigator.clipboard.read) {
            try {
                const clipboardItems = await navigator.clipboard.read();
                for (const clipboardItem of clipboardItems) {
                    for (const type of clipboardItem.types) {
                        if (type.startsWith('image/')) {
                            e.preventDefault();
                            const blob = await clipboardItem.getType(type);
                            // 创建File对象并处理上传
                            const file = new File([blob], `pasted-image.${type.split('/')[1]}`, { type });
                            this.handlePastedFile(file);
                            return;
                        }
                    }
                }
            } catch (error) {
                console.log('无法读取剪贴板:', error);
            }
        }
    }
    
    /**
     * 处理粘贴的文件
     */
    handlePastedFile(file) {
        // 切换到文件上传标签页
        TabManager.switchTab('file');
        
        // 设置文件到文件输入框
        const fileInput = document.getElementById('file-input');
        if (fileInput) {
            const dataTransfer = new DataTransfer();
            dataTransfer.items.add(file);
            fileInput.files = dataTransfer.files;
            FileUpload.updateFileDisplay(file);
            showNotification('已粘贴文件: ' + file.name, 'success');
        }
    }
    
    /**
     * 应用动态配置
     */
    applyDynamicConfig() {
        // 应用从后端模板传递的样式配置
        this.applyTemplateConfig();
        
        if (!AppState.config) return;
        
        // 设置页面标题
        if (AppState.config.name) {
            document.title = AppState.config.name;
        }
        
        // 应用自定义背景（如果有）
        this.applyCustomBackground();
        
        // 设置通知内容
        if (AppState.config.notify_title && AppState.config.notify_content) {
            this.showWelcomeNotification();
        }
    }
    
    /**
     * 应用模板配置（从后端传递的动态样式）
     */
    applyTemplateConfig() {
        if (window.AppConfig) {
            // 应用不透明度
            if (window.AppConfig.opacity && window.AppConfig.opacity !== '{{opacity}}') {
                document.body.style.opacity = window.AppConfig.opacity;
            }
            
            // 应用背景图片
            if (window.AppConfig.background && window.AppConfig.background !== '{{background}}') {
                document.body.style.backgroundImage = window.AppConfig.background;
            }
        }
    }
    
    /**
     * 应用自定义背景
     */
    applyCustomBackground() {
        // 这里可以根据配置设置自定义背景
        // 例如从配置中读取背景图片URL和透明度
        const backgroundImage = AppState.config.background_image;
        const backgroundOpacity = AppState.config.background_opacity || 1;
        
        if (backgroundImage) {
            document.body.style.setProperty('--background-image', `url(${backgroundImage})`);
            document.body.style.setProperty('--background-opacity', backgroundOpacity);
            document.body.classList.add('custom-background');
        }
    }
    
    /**
     * 显示欢迎通知
     */
    showWelcomeNotification() {
        // 延迟显示欢迎消息，避免与其他通知冲突
        setTimeout(() => {
            if (AppState.config.notify_title) {
                showNotification(AppState.config.notify_title, 'info');
            }
        }, 2000);
    }
    
    /**
     * 设置全局错误处理
     */
    setupGlobalErrorHandling() {
        // 捕获未处理的Promise错误
        window.addEventListener('unhandledrejection', (event) => {
            console.error('未处理的Promise错误:', event.reason);
            showNotification('发生了未知错误', 'error');
            event.preventDefault();
        });
        
        // 捕获JavaScript错误
        window.addEventListener('error', (event) => {
            console.error('JavaScript错误:', event.error);
            // 只在开发模式下显示详细错误
            if (window.location.hostname === 'localhost') {
                showNotification('JavaScript错误: ' + event.error?.message, 'error');
            }
        });
    }
    
    /**
     * 销毁应用程序
     */
    destroy() {
        // 清理事件监听器
        this.eventListeners.forEach(({ element, event, handler }) => {
            element.removeEventListener(event, handler);
        });
        
        // 重置状态
        AppState.initialized = false;
        AppState.currentTab = 'file';
        AppState.isLoading = false;
        AppState.config = null;
        
        console.log('应用程序已销毁');
    }
}

/**
 * 创建应用程序实例
 */
const app = new FileCodeBoxApp();

/**
 * DOM内容加载完成后初始化应用程序
 */
document.addEventListener('DOMContentLoaded', () => {
    app.init();
});

/**
 * 导出全局变量和函数（保持向后兼容性）
 */
window.app = app;
window.AppState = AppState;

// 确保这些函数在全局作用域中可用
window.showResult = showResult;
window.hideResult = hideResult;
window.showNotification = showNotification;
window.copyToClipboard = copyToClipboard;
window.copyToClipboardAuto = copyToClipboardAuto;