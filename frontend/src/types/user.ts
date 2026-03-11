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

export interface LoginForm {
  username: string
  password: string
}

export interface RegisterForm {
  username: string
  email: string
  password: string
  confirmPassword: string
  nickname?: string
}
