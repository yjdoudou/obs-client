<template>
  <div class="h-full w-full flex flex-col bg-transparent text-textMain transition-colors duration-300">
    <!-- Unselected State -->
    <div v-if="!connId" class="flex-1 flex flex-col items-center justify-center p-8 text-center">
      <div class="relative mb-8">
        <div class="absolute inset-0 bg-primary/20 blur-3xl rounded-full"></div>
        <div class="relative w-32 h-32 rounded-3xl bg-gradient-to-br from-primary to-blue-400 p-[1px]">
          <div class="w-full h-full rounded-[23px] bg-bgCard flex items-center justify-center shadow-premium">
            <el-icon size="64" class="text-primary"><Cloudy /></el-icon>
          </div>
        </div>
      </div>
      <h1 class="text-4xl font-black mb-4 bg-gradient-to-r from-textMain to-textLight bg-clip-text text-transparent italic">多云存储客户端</h1>
      <p class="text-textLight max-w-sm mb-10 leading-relaxed text-sm">连接您的云对象存储，支持华为云、腾讯云、阿里云、百度云及 MinIO，体验极速、稳定、现代化的文件管理服务。</p>
      
      <el-button type="primary" size="large" round class="!px-10 !h-12 shadow-lg shadow-primary/20" @click="emit('open-add')">
        添加云端存储
      </el-button>
    </div>

    <!-- Active Connection State -->
    <div v-else class="flex-1 flex flex-col overflow-hidden">
      <!-- Secondary Toolbar -->
      <div class="h-14 flex items-center justify-between px-6 bg-black/2 dark:bg-white/2 border-b border-black/5 dark:border-white/5 shrink-0">
        <div class="flex items-center gap-2 overflow-hidden flex-1 mr-4">
          <div 
            @click="selectBucket('', '')" 
            class="flex items-center gap-1 px-3 py-1.5 rounded-lg hover:bg-black/5 dark:hover:bg-white/10 cursor-pointer transition-all text-[11px] font-bold shrink-0 uppercase tracking-tight"
            :class="!currentBucket ? 'text-primary bg-primary/5' : 'text-textLight'"
          >
            <el-icon><Menu /></el-icon>
            <span>所有存储桶</span>
          </div>
          
          <template v-if="currentBucket">
            <el-icon class="text-textLight/30 shrink-0" size="12"><ArrowRight /></el-icon>
            <div 
              @click="goBackToPrefix('')"
              class="flex items-center gap-1 px-2 py-1.5 rounded-lg hover:bg-black/5 dark:hover:bg-white/10 cursor-pointer transition-all text-xs font-bold text-primary shrink-0"
            >
              <el-icon><Coin /></el-icon>
              <span>{{ currentBucket }}</span>
            </div>
            
            <!-- Breadcrumbs for currentPrefix -->
            <template v-for="(path, index) in breadcrumbs" :key="index">
              <el-icon class="text-textLight/30 shrink-0" size="12"><ArrowRight /></el-icon>
              <div 
                @click="goBackToPrefix(path.fullPath)"
                class="px-2 py-1 rounded-lg hover:bg-black/5 dark:hover:bg-white/10 cursor-pointer transition-all text-xs font-medium text-textLight truncate max-w-[120px]"
              >
                {{ path.name }}
              </div>
            </template>
          </template>
        </div>
        
        <div class="flex items-center gap-3 shrink-0">
          <template v-if="currentBucket">
            <el-button type="primary" size="default" class="!rounded-lg !px-4 shadow-md shadow-primary/20" @click="handleUpload" :loading="isUploading">
              <el-icon class="mr-1.5"><Upload /></el-icon>
              上传文件
            </el-button>
            <div class="w-px h-4 bg-black/10 dark:bg-white/10 mx-1"></div>
          </template>
          
          <el-button-group>
            <el-tooltip content="刷新" placement="bottom">
              <el-button circle size="small" @click="currentBucket ? fetchObjects() : fetchBuckets()" class="!bg-black/5 dark:!bg-white/5 !border-none">
                <el-icon><Refresh /></el-icon>
              </el-button>
            </el-tooltip>
          </el-button-group>
        </div>
      </div>

      <!-- Main Content Area -->
      <div class="flex-1 overflow-hidden relative">
        <!-- Loading Overlay -->
        <div v-if="loading" class="absolute inset-0 z-20 flex justify-center items-center bg-bgApp/60 backdrop-blur-sm">
          <div class="flex flex-col items-center gap-3">
             <div class="w-10 h-10 border-4 border-primary/20 border-t-primary rounded-full animate-spin"></div>
             <span class="text-[10px] text-primary font-bold tracking-widest animate-pulse uppercase">Syncing</span>
          </div>
        </div>

        <!-- Buckets List Grid -->
        <div v-if="!currentBucket" class="h-full overflow-y-auto p-6 scroll-smooth bg-bgApp/30">
          <div v-if="filteredBuckets.length === 0" class="h-full flex items-center justify-center">
             <el-empty :description="searchKeyword ? '没有找到匹配的桶' : '该账号下没有桶'" :image-size="120" />
          </div>
          <div v-else class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-6">
            <div 
              v-for="bucket in filteredBuckets" :key="bucket.Name"
              @click="selectBucket(bucket.Name, bucket.Location)"
              class="group relative bg-bgCard border border-black/5 dark:border-white/5 rounded-2xl p-5 cursor-pointer hover:border-primary/30 shadow-card hover:shadow-premium transition-all duration-300 transform hover:-translate-y-1"
            >
              <div class="flex items-start justify-between mb-4">
                <div class="w-12 h-12 rounded-xl bg-primary/5 dark:bg-primary/10 flex items-center justify-center text-primary dark:text-blue-400 group-hover:bg-primary group-hover:text-white transition-all duration-300">
                  <el-icon size="24"><Coin /></el-icon>
                </div>
                <span class="text-[9px] font-black px-2 py-0.5 rounded-md bg-black/5 dark:bg-white/5 border border-black/5 dark:border-white/5 text-textLight uppercase tracking-tighter">{{ bucket.Location }}</span>
              </div>
              <h3 class="font-bold text-sm truncate mb-1 group-hover:text-primary transition-colors text-textMain" :title="bucket.Name">{{ bucket.Name }}</h3>
              <p class="text-[10px] text-textLight/70">创建于 {{ new Date(bucket.CreationDate).toLocaleDateString() }}</p>
              
              <!-- Hover arrow icon -->
              <el-icon class="absolute bottom-4 right-4 text-primary opacity-0 group-hover:opacity-100 transition-all translate-x-2 group-hover:translate-x-0 duration-300"><Right /></el-icon>
            </div>
          </div>
        </div>

        <!-- Objects List View -->
        <div v-else class="h-full flex flex-col bg-bgCard overflow-hidden">
          <el-table 
            ref="tableRef"
            :data="filteredItems" 
            height="100%"
            class="custom-table"
            :row-class-name="() => 'group'"
            :empty-text="searchKeyword ? '没有找到匹配的文件' : '该目录目前为空'"
            @row-click="handleRowClick"
            @selection-change="handleSelectionChange"
          >
            <el-table-column type="selection" width="55" />
            <el-table-column label="名称" min-width="300">
              <template #default="{ row }">
                <div class="flex items-center gap-3 py-1 transition-transform group-hover:translate-x-1 duration-300">
                  <div 
                    class="w-8 h-8 rounded flex items-center justify-center transition-all"
                    :class="getIconColorClass(row.name, row.type)"
                  >
                    <el-icon size="16"><component :is="getFileIcon(row.name, row.type)" /></el-icon>
                  </div>
                  <span class="truncate font-medium text-sm text-textMain group-hover:text-primary transition-colors">{{ row.name }}</span>
                </div>
              </template>
            </el-table-column>
            
            <el-table-column label="大小" width="120">
              <template #default="{ row }">
                <span v-if="row.type === 'file'" class="text-xs text-gray-400 dark:text-gray-500 font-mono">{{ formatBytes(row.Size) }}</span>
                <span v-else class="text-xs text-gray-300 dark:text-gray-600">-</span>
              </template>
            </el-table-column>
            
            <el-table-column label="修改时间" width="200">
              <template #default="{ row }">
                <span v-if="row.type === 'file'" class="text-xs text-gray-400 dark:text-gray-500">{{ new Date(row.LastModified).toLocaleString() }}</span>
                <span v-else class="text-xs text-gray-300 dark:text-gray-600">-</span>
              </template>
            </el-table-column>
            
            <el-table-column label="操作" width="160" align="right">
              <template #default="{ row }">
                <div class="flex justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <template v-if="row.type === 'file'">
                    <el-tooltip content="编辑" placement="top">
                      <button class="action-btn text-blue-500 hover:bg-blue-500/10 dark:hover:bg-blue-500/20" @click.stop="handleEdit(row)">
                        <el-icon><Edit /></el-icon>
                      </button>
                    </el-tooltip>
                    <el-tooltip content="下载" placement="top">
                      <button class="action-btn text-primary hover:bg-primary/10 dark:hover:bg-primary/20" @click.stop="handleDownload(row)">
                        <el-icon><Download /></el-icon>
                      </button>
                    </el-tooltip>
                    <el-tooltip content="删除" placement="top">
                      <button class="action-btn text-red-400 hover:bg-red-400/10 dark:hover:bg-red-400/20" @click.stop="handleDelete(row)">
                        <el-icon><Delete /></el-icon>
                      </button>
                    </el-tooltip>
                  </template>
                </div>
              </template>
            </el-table-column>
          </el-table>

          <!-- Batch Action Toolbar -->
          <transition name="el-zoom-in-bottom">
            <div v-if="selectedItems.length > 0" class="absolute bottom-6 left-1/2 -translate-x-1/2 z-30 flex items-center gap-4 bg-bgCard/90 backdrop-blur-md border border-primary/30 rounded-2xl px-6 py-3 shadow-premium animate-glow">
              <div class="flex items-center gap-3 pr-4 border-r border-black/5 dark:border-white/10">
                <div class="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center text-primary font-bold text-xs">
                  {{ selectedItems.length }}
                </div>
                <span class="text-xs font-medium text-textMain">已选择项目</span>
              </div>
              
              <div class="flex items-center gap-2">
                <el-button type="primary" size="default" class="!rounded-lg" @click="handleBatchDownload">
                  <el-icon class="mr-1"><Download /></el-icon>
                  批量下载
                </el-button>
                <el-button type="danger" plain size="default" class="!rounded-lg" @click="handleBatchDelete">
                  <el-icon class="mr-1"><Delete /></el-icon>
                  批量删除
                </el-button>
                <div class="w-px h-4 bg-black/10 dark:bg-white/10 mx-2"></div>
                <el-button size="default" class="!rounded-lg !border-none !bg-black/5 dark:!bg-white/5" @click="clearSelection">
                  取消
                </el-button>
              </div>
            </div>
          </transition>
        </div>
      </div>
    </div>
    
    <FilePreview 
      v-if="previewFile"
      v-model:visible="showPreview"
      :conn-id="previewFile.connId"
      :location="previewFile.location"
      :bucket-name="previewFile.bucketName"
      :object-key="previewFile.objectKey"
      :file-name="previewFile.fileName"
      @download="handlePreviewDownload"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useConnectionStore } from '../store/connection'
import { useTransferStore } from '../store/transfer'
import { storeToRefs } from 'pinia'
import { 
  ListBuckets, ListObjects, UploadFile, DownloadFile, 
  DeleteObject, SelectFile, SelectSaveFile,
  SelectDirectory, DownloadDirectory, EditFile
} from '../../wailsjs/go/app/App'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Cloudy, Menu, ArrowRight, Coin, Upload, 
  Refresh, Right, Document, Download, Delete, Folder,
  Picture, VideoCamera, Headset, Box, Tickets, View
} from '@element-plus/icons-vue'
import FilePreview from '../components/FilePreview.vue'

interface Bucket { Name: string; CreationDate: string; Location: string }
interface Content { Key: string; Size: number; LastModified: string; StorageClass: string }

const emit = defineEmits(['open-add'])
const connStore = useConnectionStore()
const transferStore = useTransferStore()
const { 
  activeConnectionId: connId, 
  currentBucket, 
  currentLocation, 
  currentPrefix, 
  searchKeyword 
} = storeToRefs(connStore)

const loading = ref(false)
const isUploading = ref(false)
const buckets = ref<Bucket[]>([])
const objects = ref<Content[]>([])
const folders = ref<string[]>([])
const selectedItems = ref<any[]>([])
const tableRef = ref<any>(null)

const showPreview = ref(false)
const previewFile = ref<{
  connId: string
  location: string
  bucketName: string
  objectKey: string
  fileName: string
} | null>(null)

const handleSelectionChange = (val: any[]) => {
  selectedItems.value = val
}

const clearSelection = () => {
  if (tableRef.value) {
    tableRef.value.clearSelection()
  }
}

const getFileIcon = (name: string, type: 'file' | 'folder') => {
  if (type === 'folder') return Folder
  const ext = name.split('.').pop()?.toLowerCase() || ''
  
  const iconMap: Record<string, any> = {
    // Images
    jpg: Picture, jpeg: Picture, png: Picture, gif: Picture, svg: Picture, webp: Picture, bmp: Picture,
    // Videos
    mp4: VideoCamera, mkv: VideoCamera, avi: VideoCamera, mov: VideoCamera, wmv: VideoCamera,
    // Audio
    mp3: Headset, wav: Headset, ogg: Headset, flac: Headset,
    // Archives
    zip: Box, rar: Box, '7z': Box, tar: Box, gz: Box,
    // Documents
    pdf: Tickets, doc: Tickets, docx: Tickets, xls: Tickets, xlsx: Tickets, ppt: Tickets, pptx: Tickets, txt: Tickets
  }
  
  return iconMap[ext] || Document
}

const getIconColorClass = (name: string, type: 'file' | 'folder') => {
  if (type === 'folder') return 'bg-orange-500/10 text-orange-500 group-hover:bg-orange-500 group-hover:text-white'
  const ext = name.split('.').pop()?.toLowerCase() || ''
  
  if (['jpg', 'jpeg', 'png', 'gif', 'svg', 'webp', 'bmp'].includes(ext)) {
    return 'bg-emerald-500/10 text-emerald-500 group-hover:bg-emerald-500 group-hover:text-white'
  }
  if (['mp4', 'mkv', 'avi', 'mov', 'wmv'].includes(ext)) {
    return 'bg-purple-500/10 text-purple-500 group-hover:bg-purple-500 group-hover:text-white'
  }
  if (['mp3', 'wav', 'ogg', 'flac'].includes(ext)) {
    return 'bg-pink-500/10 text-pink-500 group-hover:bg-pink-500 group-hover:text-white'
  }
  if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) {
    return 'bg-amber-500/10 text-amber-500 group-hover:bg-amber-500 group-hover:text-white'
  }
  if (['pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx', 'txt'].includes(ext)) {
    return 'bg-blue-500/10 text-blue-500 group-hover:bg-blue-500 group-hover:text-white'
  }
  
  return 'bg-gray-500/10 text-gray-500 group-hover:bg-gray-500 group-hover:text-white'
}

const breadcrumbs = computed(() => {
  if (!currentPrefix.value) return []
  const parts = currentPrefix.value.split('/').filter(p => p)
  let fullPath = ''
  return parts.map(name => {
    fullPath += name + '/'
    return { name, fullPath }
  })
})

const filteredBuckets = computed(() => {
  if (!searchKeyword.value) return buckets.value
  return buckets.value.filter(b => b.Name.toLowerCase().includes(searchKeyword.value.toLowerCase()))
})

const filteredItems = computed(() => {
  const keyword = searchKeyword.value.toLowerCase()
  const items = [
    ...folders.value.map(f => ({
      name: f.split('/').filter(p => p).pop() + '/',
      fullPath: f,
      type: 'folder'
    })),
    ...objects.value
      .filter(o => o.Key !== currentPrefix.value && !o.Key.endsWith('/')) // Filter out directory marker objects
      .map(o => ({
        ...o,
        name: o.Key.split('/').pop() || o.Key,
        type: 'file'
      }))
  ]
  if (!keyword) return items
  return items.filter(i => i.name.toLowerCase().includes(keyword))
})

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const fetchBuckets = async () => {
  if (!connId.value) return
  loading.value = true
  try {
    const data = await ListBuckets(connId.value)
    buckets.value = (data || []) as unknown as Bucket[]
  } catch (err: any) {
    ElMessage.error('获取桶列表失败: ' + (err.message || err))
    buckets.value = []
  } finally {
    loading.value = false
  }
}

const fetchObjects = async () => {
  if (!connId.value || !currentBucket.value) return
  // Important: Clear current lists immediately to free up memory/reactivity processing
  objects.value = []
  folders.value = []
  loading.value = true
  
  // If location is missing but we have bucket, try to find it in the list
  if (!currentLocation.value && buckets.value.length > 0) {
    const bucket = buckets.value.find(b => b.Name === currentBucket.value)
    if (bucket) {
      connStore.setLocation(bucket.Location)
    }
  }

  try {
    const data = await ListObjects(connId.value, currentLocation.value, currentBucket.value, currentPrefix.value, '/') as any
    objects.value = (data.objects || []) as unknown as Content[]
    folders.value = (data.folders || []) as string[]
  } catch (err: any) {
    ElMessage.error('获取对象失败: ' + (err.message || err))
    objects.value = []
    folders.value = []
  } finally {
    loading.value = false
  }
}

const selectBucket = (bucketName: string, location: string = '') => {
  if (currentBucket.value === bucketName && currentLocation.value === location) return
  loading.value = true
  connStore.setBucket(bucketName, location)
}

const goBackToPrefix = (prefix: string) => {
  if (currentPrefix.value === prefix) return
  loading.value = true
  connStore.setPrefix(prefix)
}

const handleRowClick = (row: any) => {
  if (row.type === 'folder') {
    if (currentPrefix.value === row.fullPath) return
    loading.value = true
    connStore.setPrefix(row.fullPath)
  } else {
    const now = Date.now()
    const lastClick = (row as any)._lastClickTime || 0
    if (now - lastClick < 300) {
      handlePreview(row)
    }
    ;(row as any)._lastClickTime = now
  }
}

const handlePreview = (row: Content) => {
  if (!connId.value || !currentBucket.value) return
  
  const fileName = row.Key.split('/').pop() || row.Key
  previewFile.value = {
    connId: connId.value,
    location: currentLocation.value,
    bucketName: currentBucket.value,
    objectKey: row.Key,
    fileName
  }
  showPreview.value = true
}

const handleEdit = async (row: Content) => {
  if (!connId.value || !currentBucket.value) return
  
  try {
    const result = await EditFile(connId.value, currentLocation.value, currentBucket.value, row.Key)
    ElMessage.success(result)
  } catch (error: any) {
    ElMessage.error('编辑文件失败: ' + (error.message || '未知错误'))
  }
}

const handlePreviewDownload = async (data: { objectKey: string; fileName: string }) => {
  if (!connId.value || !currentBucket.value) return
  try {
    const savePath = await SelectSaveFile(data.fileName)
    if (!savePath) return 
    
    const taskID = `dl-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    transferStore.addTask({
      id: taskID,
      name: data.fileName,
      type: 'download',
      size: 0
    })

    await DownloadFile(connId.value, currentLocation.value, currentBucket.value, data.objectKey, savePath, taskID)
    ElMessage.success("下载任务已提交")
  } catch (e: any) {
    ElMessage.error("下载出错: " + (e.message || e))
  }
}

const handleUpload = async () => {
  if (!connId.value || !currentBucket.value) return
  try {
    const filePaths = await SelectFile()
    if (!filePaths || filePaths.length === 0) return
    
    isUploading.value = true
    const uploadPromises = []
    
    for (const localPath of filePaths) {
      // 使用更可靠的文件名提取方法
      const fileName = localPath.replace(/\\/g, '/').split('/').pop() || 'upload.file'
      const objectKey = currentPrefix.value + fileName
      
      const taskID = `up-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
      transferStore.addTask({
        id: taskID,
        name: fileName,
        type: 'upload',
        size: 0 // Backend will report total size in first progress event
      })

      uploadPromises.push(UploadFile(connId.value, currentLocation.value, currentBucket.value, objectKey, localPath, taskID))
    }
    
    await Promise.all(uploadPromises)
    ElMessage.success("上传任务已提交")
    fetchObjects()
  } catch (e: any) {
    ElMessage.error("上传错误: " + (e.message || e))
  } finally {
    isUploading.value = false
  }
}

const handleDownload = async (row: Content) => {
  if (!connId.value || !currentBucket.value) return
  try {
    const fileName = row.Key.split('/').pop() || row.Key
    const savePath = await SelectSaveFile(fileName)
    if (!savePath) return 
    
    const taskID = `dl-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    transferStore.addTask({
      id: taskID,
      name: fileName,
      type: 'download',
      size: row.Size || 0
    })

    await DownloadFile(connId.value, currentLocation.value, currentBucket.value, row.Key, savePath, taskID)
    ElMessage.success("下载任务已提交")
  } catch (e: any) {
    ElMessage.error("下载出错: " + (e.message || e))
  }
}

const handleDelete = async (row: Content) => {
  if (!connId.value || !currentBucket.value) return
  try {
    const name = row.Key.split('/').pop() || row.Key
    await ElMessageBox.confirm(`确定要删除对象 ${name} 吗?`, '警告', {
      type: 'warning',
    })
    await DeleteObject(connId.value, currentLocation.value, currentBucket.value, row.Key)
    ElMessage.success("删除成功")
    fetchObjects()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error("删除失败: " + (e.message || e))
  }
}

const handleBatchDownload = async () => {
  if (!connId.value || !currentBucket.value || selectedItems.value.length === 0) return
  
  try {
    const localDir = await SelectDirectory()
    if (!localDir) return

    ElMessage.info(`开始下载 ${selectedItems.value.length} 个项目...`)
    const downloadPromises = []
    
    for (const item of selectedItems.value) {
      const taskID = `dl-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
      if (item.type === 'folder') {
        const folderName = item.name
        transferStore.addTask({
          id: taskID,
          name: folderName,
          type: 'download',
          size: 0
        })
        downloadPromises.push(DownloadDirectory(connId.value, currentLocation.value, currentBucket.value, item.fullPath, localDir, taskID))
      } else {
        const fileName = item.Key.split('/').pop() || item.Key
        const localPath = `${localDir}${localDir.includes('\\') ? '\\' : '/'}${fileName}`
        
        transferStore.addTask({
          id: taskID,
          name: fileName,
          type: 'download',
          size: item.Size || 0
        })
        
        downloadPromises.push(DownloadFile(connId.value, currentLocation.value, currentBucket.value, item.Key, localPath, taskID))
      }
    }
    
    await Promise.all(downloadPromises)
    ElMessage.success("批量下载任务已提交")
    clearSelection()
  } catch (e: any) {
    ElMessage.error("批量下载出错: " + (e.message || e))
  }
}

const handleBatchDelete = async () => {
  if (!connId.value || !currentBucket.value || selectedItems.value.length === 0) return
  
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedItems.value.length} 个项目吗? 此操作不可撤销！`, '严重警告', {
      type: 'error',
    })
    
    ElMessage.info("正在批量删除...")
    for (const item of selectedItems.value) {
      if (item.type === 'file') {
        await DeleteObject(connId.value, currentLocation.value, currentBucket.value, item.Key)
      }
      // Note: Folder deletion would require recursive logic on backend too, 
      // but for now we focus on files and simple folders as markers.
    }
    
    ElMessage.success("批量删除操作已提交")
    fetchObjects()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error("批量删除出错: " + (e.message || e))
  }
}

// Watch store state instead of local
watch(currentBucket, (newBucket) => {
  if (newBucket) {
    fetchObjects()
  } else {
    objects.value = []
    folders.value = []
    fetchBuckets()
  }
})

watch(currentPrefix, () => {
  if (currentBucket.value) {
    fetchObjects()
  }
})

// Watch connection changes
watch(connId, (newId) => {
  if (newId) {
    connStore.setBucket('', '') // reset to bucket list
    fetchBuckets()
  }
})

onMounted(() => {
  if (connId.value) {
    fetchBuckets()
    if (currentBucket.value) fetchObjects()
  }
})

</script>

<style scoped>
.action-btn {
  @apply w-7 h-7 flex items-center justify-center rounded-md transition-all active:scale-90;
}

:deep(.custom-table) {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-header-bg-color: transparent;
  background-color: transparent !important;
  --el-table-border-color: var(--border-color);
  --el-table-text-color: var(--text-main);
}

:deep(.el-table__row) {
  cursor: pointer;
  @apply transition-colors;
}

:deep(.el-table__row:hover > td) {
  background-color: rgba(0, 0, 0, 0.02) !important;
}

.dark :deep(.el-table__row:hover > td) {
  background-color: rgba(255, 255, 255, 0.05) !important;
}

:deep(.el-table th.el-table__cell) {
  @apply bg-black/[0.02] dark:bg-white/[0.04] text-[11px] font-bold uppercase tracking-wider py-3 text-textLight;
}

.nav-btn {
  @apply w-7 h-7 flex items-center justify-center rounded-md text-textLight hover:bg-black/5 dark:hover:bg-white/10 hover:text-primary transition-all active:scale-95;
}
</style>
