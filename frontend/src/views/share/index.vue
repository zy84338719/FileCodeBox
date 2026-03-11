<template>
  <div class="share-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <h3>文件分享</h3>
          <el-radio-group v-model="activeTab">
            <el-radio-button label="upload">上传文件</el-radio-button>
            <el-radio-button label="text">分享文本</el-radio-button>
            <el-radio-button label="get">获取分享</el-radio-button>
          </el-radio-group>
        </div>
      </template>
      
      <div class="tab-content">
        <!-- 文件上传 -->
        <FileUpload v-if="activeTab === 'upload'" @success="handleSuccess" />
        
        <!-- 文本分享 -->
        <TextShare v-if="activeTab === 'text'" @success="handleSuccess" />
        
        <!-- 获取分享 -->
        <GetShare v-if="activeTab === 'get'" :initial-code="initialCode" />
      </div>
    </el-card>
    
    <!-- 分享成功对话框 -->
    <el-dialog v-model="showSuccessDialog" title="分享成功" width="500px">
      <div class="success-content">
        <el-result icon="success" title="分享成功">
          <template #sub-title>
            您的分享链接已生成
          </template>
        </el-result>
        
        <el-input
          v-model="shareUrl"
          readonly
          :suffix-icon="CopyDocument"
          @click="copyUrl"
        />
      </div>
      
      <template #footer>
        <el-button @click="showSuccessDialog = false">关闭</el-button>
        <el-button type="primary" @click="copyUrl">复制链接</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { CopyDocument } from '@element-plus/icons-vue'
import FileUpload from '@/components/upload/FileUpload.vue'
import TextShare from '@/components/upload/TextShare.vue'
import GetShare from '@/components/upload/GetShare.vue'

const route = useRoute()
const activeTab = ref('upload')
const showSuccessDialog = ref(false)
const shareUrl = ref('')
const initialCode = ref('')

interface ShareResult {
  code: string
  share_url: string
  full_share_url: string
  qr_code_data: string
}

const handleSuccess = (result: ShareResult) => {
  shareUrl.value = result.full_share_url || result.share_url
  showSuccessDialog.value = true
}

const copyUrl = async () => {
  try {
    await navigator.clipboard.writeText(shareUrl.value)
    ElMessage.success('链接已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败，请手动复制')
  }
}

// 检查 URL 中是否有分享码
onMounted(() => {
  const code = route.params.code as string
  if (code) {
    initialCode.value = code
    activeTab.value = 'get'
  }
})
</script>

<style scoped lang="scss">
.share-container {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
  
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    
    h3 {
      margin: 0;
    }
  }
  
  .tab-content {
    margin-top: 20px;
  }
  
  .success-content {
    text-align: center;
    
    .el-input {
      margin-top: 20px;
    }
  }
}
</style>
