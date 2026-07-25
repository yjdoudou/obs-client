<template>
  <el-dialog 
    :model-value="visible"
    @update:model-value="emit('update:visible', $event)"
    :title="fileName" 
    width="90vw" 
    height="90vh"
    :close-on-click-modal="false"
    class="file-preview-dialog"
  >
    <div class="preview-container">
      <div v-if="loading" class="loading-overlay">
        <div class="flex flex-col items-center gap-3">
          <div class="w-10 h-10 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
          <span class="text-[10px] text-primary font-bold tracking-widest uppercase">Loading</span>
        </div>
      </div>
      
      <div v-else-if="error" class="error-content">
        <div class="text-center py-20">
          <el-icon size="64" class="text-red-400 mb-4"><Warning /></el-icon>
          <p class="text-textLight text-sm">{{ error }}</p>
        </div>
      </div>
      
      <div v-else class="content-wrapper">
        <!-- 图片预览 -->
        <div v-if="isImage" class="image-preview">
          <img 
            :src="blobUrl" 
            :alt="fileName"
            class="max-w-full max-h-full object-contain"
            @error="handleImageError"
          />
        </div>
        
        <!-- PDF预览 -->
        <div v-else-if="isPdf" class="pdf-preview">
          <iframe 
            :src="blobUrl" 
            class="w-full h-full"
            title="PDF Preview"
            @load="onPdfLoaded"
          ></iframe>
        </div>
        
        <!-- 文本预览 -->
        <div v-else-if="isText" class="text-preview">
          <pre class="text-sm font-mono whitespace-pre-wrap break-all text-textMain">{{ textContent }}</pre>
        </div>
        
        <!-- Excel预览 -->
        <div v-else-if="isExcel" class="excel-preview">
          <div ref="excelContainer" class="w-full h-full overflow-auto"></div>
        </div>
        
        <!-- Word预览 -->
        <div v-else-if="isWord" class="word-preview">
          <div ref="wordContainer" class="w-full h-full overflow-auto p-4"></div>
        </div>
        
        <!-- 未知文件类型 -->
        <div v-else class="unknown-preview">
          <div class="text-center py-20">
            <el-icon size="64" class="text-gray-400 mb-4"><Document /></el-icon>
            <p class="text-textLight text-base mb-2">暂不支持此文件格式预览</p>
            <p class="text-textLight/50 text-xs">{{ fileName }}</p>
            <p class="text-textLight/40 text-xs mt-1">{{ contentType }}</p>
          </div>
        </div>
      </div>
    </div>
    
    <template #footer>
      <div class="flex items-center justify-between w-full">
        <div class="text-xs text-textLight">
          {{ formatSize(size) }} | {{ contentType }}
        </div>
        <div class="flex gap-2">
          <el-button @click="handleDownload">
            <el-icon class="mr-1"><Download /></el-icon>
            下载
          </el-button>
          <el-button type="primary" @click="emit('update:visible', false)">
            关闭
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted, nextTick } from 'vue'
import { PreviewFile } from '../../wailsjs/go/app/App'
import { ElMessage } from 'element-plus'
import { Warning, Document, Download } from '@element-plus/icons-vue'
import * as XLSX from 'xlsx'
import { renderAsync } from 'docx-preview'

interface Props {
  visible: boolean
  connId: string
  location: string
  bucketName: string
  objectKey: string
  fileName: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'download', data: { objectKey: string; fileName: string }): void
}>()

const loading = ref(false)
const error = ref('')
const content = ref('')
const contentType = ref('')
const size = ref(0)
const currentBlobUrl = ref('')
const excelContainer = ref<HTMLElement | null>(null)
const wordContainer = ref<HTMLElement | null>(null)

const base64ToUint8Array = (base64: string): Uint8Array => {
  const binary = atob(base64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes
}

const blobUrl = computed(() => {
  if (!content.value) return ''
  
  if (currentBlobUrl.value) {
    URL.revokeObjectURL(currentBlobUrl.value)
  }
  
  try {
    const bytes = base64ToUint8Array(content.value)
    const blob = new Blob([bytes], { type: contentType.value })
    currentBlobUrl.value = URL.createObjectURL(blob)
    return currentBlobUrl.value
  } catch {
    return ''
  }
})

const textContent = computed(() => {
  if (!content.value) return ''
  try {
    const bytes = base64ToUint8Array(content.value)
    return new TextDecoder('utf-8').decode(bytes)
  } catch {
    return '无法解析文本内容'
  }
})

const isImage = computed(() => {
  return contentType.value.startsWith('image/')
})

const isPdf = computed(() => {
  return contentType.value === 'application/pdf' || props.fileName.endsWith('.pdf')
})

const isText = computed(() => {
  const textTypes = [
    'text/plain', 'text/html', 'text/css', 'text/javascript',
    'application/json', 'application/xml', 'text/markdown',
    'text/x-python', 'text/x-go', 'text/x-c', 'text/x-cpp',
    'text/x-java', 'text/x-php', 'text/x-shellscript',
    'application/javascript', 'application/typescript',
  ]
  const textExtensions = ['txt', 'md', 'json', 'xml', 'html', 'css', 'js', 'ts', 'py', 'go', 'c', 'cpp', 'java', 'php', 'sh', 'yaml', 'yml', 'ini', 'conf', 'log', 'csv']
  const lowerFileName = props.fileName.toLowerCase()
  
  return textTypes.some(t => contentType.value.includes(t)) || 
         textExtensions.some(ext => lowerFileName.endsWith('.' + ext))
})

const isExcel = computed(() => {
  const excelExtensions = ['.xls', '.xlsx', '.xlsm', '.xlsb']
  const lowerFileName = props.fileName.toLowerCase()
  
  return excelExtensions.some(ext => lowerFileName.endsWith(ext))
})

const isWord = computed(() => {
  const wordExtensions = ['.docx']
  const lowerFileName = props.fileName.toLowerCase()
  
  return wordExtensions.some(ext => lowerFileName.endsWith(ext))
})

const formatSize = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const loadContent = async () => {
  if (!props.visible || !props.connId || !props.bucketName || !props.objectKey) return
  
  loading.value = true
  error.value = ''
  content.value = ''
  
  try {
    const result = await PreviewFile(props.connId, props.location, props.bucketName, props.objectKey)
    if (result) {
      content.value = result.content
      contentType.value = result.contentType
      size.value = result.size
    }
  } catch (err: any) {
    error.value = err.message || '预览失败'
    ElMessage.error(error.value)
  } finally {
    loading.value = false
  }
}

const renderExcel = async () => {
  if (!excelContainer.value || !content.value) return
  
  try {
    const bytes = base64ToUint8Array(content.value)
    const workbook = XLSX.read(bytes, { type: 'array', cellDates: true })
    
    excelContainer.value.innerHTML = ''
    
    for (const sheetName of workbook.SheetNames) {
      const worksheet = workbook.Sheets[sheetName]
      const table = XLSX.utils.sheet_to_html(worksheet)
      
      const sheetDiv = document.createElement('div')
      sheetDiv.className = 'mb-6'
      sheetDiv.innerHTML = `
        <div class="text-sm font-bold text-textMain mb-2 px-2 py-1 bg-gray-100 dark:bg-gray-800 rounded">
          ${sheetName}
        </div>
        <div class="overflow-x-auto border border-gray-200 dark:border-gray-700 rounded">
          ${table}
        </div>
      `
      excelContainer.value.appendChild(sheetDiv)
    }
  } catch (err: any) {
    error.value = 'Excel 解析失败: ' + (err.message || '未知错误')
    ElMessage.error(error.value)
  }
}

const renderWord = async () => {
  if (!wordContainer.value || !content.value) return
  
  try {
    const bytes = base64ToUint8Array(content.value)
    const blob = new Blob([bytes], { type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' })
    
    wordContainer.value.innerHTML = ''
    
    await renderAsync(blob, wordContainer.value, undefined, {
      className: 'docx-wrapper',
      inWrapper: true,
      ignoreWidth: false,
      ignoreHeight: false,
      ignoreFonts: false,
      breakPages: true,
      ignoreLastRenderedPageBreak: true,
      experimental: false,
      trimXmlDeclaration: true,
      useBase64URL: true,
      useMathMLPolyfill: false,
      renderChanges: true,
      renderHeaders: true,
      renderFooters: true,
      renderFootnotes: true,
      renderEndnotes: true,
    })
    
    const wrapper = wordContainer.value.querySelector('.docx-wrapper')
    if (!wrapper || wrapper.innerHTML.trim() === '') {
      throw new Error('docx-preview 渲染结果为空')
    }
  } catch (err: any) {
    console.error('Word 预览错误:', err)
    wordContainer.value.innerHTML = `
      <div class="text-center py-10">
        <el-icon size="48" class="text-red-400 mb-3"><Warning /></el-icon>
        <p class="text-textLight text-sm mb-2">Word 预览失败</p>
        <p class="text-textLight/50 text-xs">${err.message || '未知错误'}</p>
        <p class="text-textLight/40 text-xs mt-2">建议下载文件后使用 Word 打开</p>
      </div>
    `
    ElMessage.error('Word 预览失败: ' + (err.message || '未知错误'))
  }
}

watch([() => props.visible, content], async ([visible, contentVal]) => {
  if (visible && contentVal) {
    await nextTick()
    if (isExcel.value && excelContainer.value) {
      await renderExcel()
    }
    if (isWord.value && wordContainer.value) {
      await renderWord()
    }
  }
})

watch([() => props.visible, () => props.connId, () => props.bucketName, () => props.objectKey], ([visible, connId, bucketName, objectKey]) => {
  if (visible && connId && bucketName && objectKey) {
    loadContent()
  } else if (!visible) {
    revokeBlobUrl()
  }
})

const handleImageError = () => {
  error.value = '图片加载失败'
}

const onPdfLoaded = () => {
  loading.value = false
}

const handleDownload = () => {
  emit('download', {
    objectKey: props.objectKey,
    fileName: props.fileName
  })
}

const revokeBlobUrl = () => {
  if (currentBlobUrl.value) {
    URL.revokeObjectURL(currentBlobUrl.value)
    currentBlobUrl.value = ''
  }
}

onUnmounted(() => {
  revokeBlobUrl()
})
</script>

<style scoped>
.preview-container {
  height: 60vh;
  position: relative;
  background: var(--bg-app);
  border-radius: 8px;
  overflow: hidden;
}

.loading-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-app);
  z-index: 10;
}

.error-content,
.unknown-preview {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.content-wrapper {
  height: 100%;
  overflow: auto;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.image-preview {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pdf-preview {
  width: 100%;
  height: 100%;
  border: none;
}

.text-preview {
  width: 100%;
  height: 100%;
  overflow: auto;
  background: var(--bg-card);
  padding: 16px;
  border-radius: 8px;
}

.excel-preview {
  width: 100%;
  height: 100%;
  overflow: auto;
  padding: 16px;
}

.word-preview {
  width: 100%;
  height: 100%;
  overflow: auto;
  background: white;
}

:deep(.excel-preview table) {
  border-collapse: collapse;
  font-size: 12px;
}

:deep(.excel-preview th),
:deep(.excel-preview td) {
  border: 1px solid #e5e7eb;
  padding: 4px 8px;
  text-align: left;
  white-space: nowrap;
}

:deep(.excel-preview th) {
  background: #f3f4f6;
  font-weight: 600;
}

:deep(.excel-preview tr:nth-child(even)) {
  background: #f9fafb;
}

:deep(.word-preview .docx-wrapper) {
  background: white;
  padding: 20px;
}

:deep(.word-preview section) {
  margin-bottom: 20px;
}
</style>
