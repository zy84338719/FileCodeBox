// NOTE: @/api/* 和 @/types/user 的真实模块已存在且有完整类型，
// 不再在此用 any 声明覆盖（避免降级类型检查）。仅保留无真实文件的类型声明。

declare module '@/types/share' {
  export interface ShareInfo {
    code: string
    filename: string
    file_size: number
    content_type: 'text' | 'file'
    content?: string
    has_password: boolean
    created_at: string
    expire_time?: string
    download_count: number
    max_downloads?: number
    username?: string
  }
}

declare module 'swagger-ui-dist/swagger-ui-es-bundle' {
  export const SwaggerUIBundle: any
  export const SwaggerUIStandalonePreset: any
}

declare module 'swagger-ui-dist/swagger-ui.css' {
  const css: string
  export default css
}