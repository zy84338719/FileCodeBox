declare module '@/api/user' {
  export const userApi: any
}

declare module '@/api/share' {
  export const shareApi: any
}

declare module '@/api/admin' {
  export const adminApi: any
}

declare module '@/api' {
  export * from '@/api/user'
  export * from '@/api/share'
  export * from '@/api/admin'
}

declare module '@/types/user' {
  export interface UserInfo {
    id: number
    username: string
    email: string
    nickname: string
    avatar?: string
    status: number
    role?: string
    created_at: string
  }

  export interface UserStats {
    total_uploads: number
    total_downloads: number
    total_storage: number
    max_storage_quota: number
    current_files: number
    file_count: number
  }
}

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