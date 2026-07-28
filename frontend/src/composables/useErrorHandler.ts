// Stub - full implementation in Task 7
import { ElMessage } from 'element-plus'

interface ErrInput {
  code?: number
  message?: string
  response?: {
    status?: number
    data?: { code?: number; message?: string; trace_id?: string }
  }
}

export const useErrorHandler = () => {
  const handleError = (e: ErrInput | unknown) => {
    const err = e as ErrInput
    if (err && typeof err === 'object' && 'code' in err && err.code !== undefined) {
      ElMessage.error(err.message || 'Error')
    } else {
      ElMessage.error('Unknown error')
    }
  }
  return { handleError }
}
