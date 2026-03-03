<template>
  <div class="h-screen w-screen flex flex-col bg-bgApp text-textMain overflow-hidden font-sans transition-colors duration-300">
    <!-- Header Tool Bar -->
    <header class="h-14 glass flex items-center px-6 justify-between shrink-0 z-20 border-b border-black/5 dark:border-white/5 gap-4">
      <div class="flex items-center gap-4 shrink-0">
        <div class="w-9 h-9 rounded-lg bg-primary flex items-center justify-center font-bold text-white shadow-lg shadow-primary/20 transform transition-transform hover:scale-105">
          OBS
        </div>
        <span class="font-bold text-lg tracking-tight whitespace-nowrap">Huawei OBS Client</span>
      </div>
      
      <div class="flex-1 max-w-xl relative group min-w-[200px] h-9 bg-black/5 dark:bg-white/5 border border-black/5 dark:border-white/10 rounded-full flex items-center px-3 focus-within:bg-white dark:focus-within:bg-bgCard focus-within:border-primary/50 transition-all">
        <el-icon class="text-textLight group-focus-within:text-primary transition-colors shrink-0 mr-2"><Search /></el-icon>
        <input 
          v-model="connStore.searchKeyword"
          type="text" 
          placeholder="搜索您的桶或文件..." 
          class="flex-1 bg-transparent border-none outline-none text-xs text-textMain placeholder:text-textLight/50"
        />
      </div>

      <div class="flex items-center gap-2 shrink-0">
        <el-tooltip :content="isDark ? '切换亮色模式' : '切换暗色模式'" placement="bottom">
          <button @click="connStore.toggleTheme" class="w-8 h-8 rounded-full hover:bg-black/5 dark:hover:bg-white/10 flex items-center justify-center transition-colors">
            <el-icon><component :is="isDark ? Sunny : Moon" /></el-icon>
          </button>
        </el-tooltip>
        <el-tooltip content="设置" placement="bottom">
          <button class="w-8 h-8 rounded-full hover:bg-black/5 dark:hover:bg-white/10 flex items-center justify-center transition-colors">
            <el-icon><Setting /></el-icon>
          </button>
        </el-tooltip>
        <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-primary to-blue-400 p-[1px] ml-1">
          <div class="w-full h-full rounded-full bg-bgCard flex items-center justify-center overflow-hidden">
             <el-icon size="16"><User /></el-icon>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Layout Content -->
    <div class="flex-1 flex overflow-hidden">
      <!-- Left Sidebar: Connections -->
      <aside class="w-64 glass flex flex-col shrink-0 m-2 mr-0 rounded-xl overflow-hidden border border-black/5 dark:border-white/5">
        <div class="p-4 flex justify-between items-center">
          <span class="font-bold text-[10px] uppercase tracking-widest text-textLight">Connections</span>
          <button 
            @click="openAddDialog"
            class="w-6 h-6 rounded-full bg-primary/10 dark:bg-primary/20 text-primary flex items-center justify-center hover:bg-primary hover:text-white transition-all transform hover:rotate-90"
          >
            <el-icon><Plus /></el-icon>
          </button>
        </div>
        
        <div class="flex-1 overflow-y-auto px-2 pb-4 space-y-1">
          <div v-if="connections.length === 0" class="text-xs text-textLight p-4 text-center italic">
            点击上方 + 号添加首个连接
          </div>
          <div 
            v-for="conn in connections" :key="conn.id"
            class="group relative flex items-center px-4 py-3 rounded-lg cursor-pointer transition-all duration-300 overflow-hidden"
            :class="activeConnId === conn.id ? 'bg-primary/5 dark:bg-primary/10 text-primary shadow-sm' : 'hover:bg-black/5 dark:hover:bg-white/5 text-textMain/70 hover:text-primary'"
            @click="selectConnection(conn.id)"
          >
            <!-- Active indicator dot -->
            <div 
              class="absolute left-0 w-1 h-6 bg-primary rounded-r-full transition-transform duration-300"
              :class="activeConnId === conn.id ? 'scale-y-100' : 'scale-y-0 opacity-0'"
            ></div>

            <el-icon class="mr-3 text-lg transition-colors" :class="activeConnId === conn.id ? 'text-primary' : 'text-textLight group-hover:text-primary'"><Link /></el-icon>
            <span class="truncate font-medium text-sm transition-transform group-hover:translate-x-1">{{ conn.name }}</span>
            
            <div class="ml-auto flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-all">
              <button 
                @click.stop="duplicateConnection(conn.id)"
                class="p-1 hover:bg-black/5 dark:hover:bg-white/10 rounded group/btn"
                title="复制连接"
              >
                <el-icon size="14" class="group-hover/btn:text-primary"><CopyDocument /></el-icon>
              </button>
              <button 
                @click.stop="editConnection(conn)"
                class="p-1 hover:bg-black/5 dark:hover:bg-white/10 rounded group/btn"
                title="编辑连接"
              >
                <el-icon size="14" class="group-hover/btn:text-primary"><Edit /></el-icon>
              </button>
              <button 
                @click.stop="deleteConnection(conn.id)"
                class="p-1 hover:bg-black/5 dark:hover:bg-white/10 rounded group/btn"
                title="删除连接"
              >
                <el-icon size="14" class="group-hover/btn:text-danger"><Delete /></el-icon>
              </button>
            </div>
          </div>
        </div>

        <!-- Transfer Status Mini Card -->
        <div class="p-4 mt-auto">
          <div class="bg-black/5 dark:bg-white/5 rounded-xl p-3 border border-black/5 dark:border-white/5">
            <div class="flex justify-between items-center mb-2">
              <span class="text-[10px] text-textLight uppercase font-bold">传输任务</span>
              <span class="text-[10px] text-primary cursor-pointer hover:underline" @click="openTransferPanel">查看全部</span>
            </div>
            <div class="space-y-2">
              <div v-if="transferStore.tasks.length > 0" class="space-y-2">
                <div v-for="task in transferStore.tasks.slice(0, 2)" :key="task.id" class="flex flex-col gap-1">
                  <div class="flex justify-between text-[9px] text-textLight truncate">
                    <span>{{ task.name }}</span>
                    <span>{{ task.progress }}%</span>
                  </div>
                  <div class="h-1 bg-black/10 dark:bg-white/10 rounded-full overflow-hidden">
                    <div class="bg-primary h-full transition-all duration-300" :style="{ width: task.progress + '%' }"></div>
                  </div>
                </div>
              </div>
              <div v-else class="text-[9px] text-textLight italic text-center py-1">暂无任务</div>
            </div>
          </div>
        </div>
      </aside>

      <!-- Center Area: File List -->
      <main class="flex-1 flex flex-col min-w-0 m-2 rounded-xl bg-bgCard border border-black/5 dark:border-white/5 relative overflow-hidden shadow-premium">
        <!-- Navigation/Address Bar -->
        <div class="h-12 flex items-center px-4 gap-4 bg-black/2 dark:bg-white/2 border-b border-black/5 dark:border-white/5 shrink-0">
          <div class="flex items-center gap-1">
            <button class="nav-btn"><el-icon><ArrowLeft /></el-icon></button>
            <button class="nav-btn"><el-icon><ArrowRight /></el-icon></button>
            <button class="nav-btn ml-1"><el-icon><RefreshRight /></el-icon></button>
          </div>
          
          <div class="flex-1 h-8 bg-black/5 dark:bg-white/5 border border-black/5 dark:border-white/5 rounded-lg flex items-center px-3 text-[11px] group hover:border-black/10 dark:hover:border-white/10 transition-colors focus-within:bg-white dark:focus-within:bg-bgCard focus-within:border-primary/50">
            <el-icon class="mr-2 opacity-50 text-textLight"><FolderOpened /></el-icon>
            <span class="text-textLight opacity-50 select-none mr-1 font-mono">obs://</span>
            <input 
              v-model="editablePath"
              type="text"
              class="flex-1 bg-transparent border-none outline-none text-textMain font-mono h-full"
              @keyup.enter="handlePathNavigation"
              @blur="syncPath"
            />
          </div>
        </div>
        
        <!-- Files View Router -->
        <div class="flex-1 overflow-hidden relative">
          <router-view />
        </div>
      </main>
    </div>

    <!-- Footer Status Bar -->
    <footer class="h-7 px-4 flex items-center justify-between text-[10px] text-textLight select-none border-t border-black/5 dark:border-white/5 shrink-0 bg-transparent">
      <div class="flex items-center gap-4">
        <span class="flex items-center gap-1.5">
          <span class="w-1.5 h-1.5 rounded-full transition-all" :class="activeConnId ? 'bg-emerald-500 shadow-sm shadow-emerald-500/50' : 'bg-gray-400 dark:bg-gray-600'"></span> 
          {{ activeConnId ? 'CONNECTED' : 'DISCONNECTED' }}
        </span>
        <div class="w-px h-3 bg-black/5 dark:bg-white/10"></div>
        <span class="uppercase tracking-tighter">Ready</span>
      </div>
      <div class="flex items-center gap-3">
        <span class="font-mono">v1.0.0-beta</span>
        <el-icon class="hover:text-primary cursor-pointer transition-colors" size="14"><InfoFilled /></el-icon>
      </div>
    </footer>
    
    <ConnectionDialog v-model:visible="showAddConnection" :edit-data="editingConn" @saved="loadConnections" />
    <TransferPanel ref="transferPanelRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import ConnectionDialog from './components/ConnectionDialog.vue'
import TransferPanel from './components/TransferPanel.vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { GetConnections, DuplicateConnection, DeleteConnection } from '../wailsjs/go/app/App'
import { useConnectionStore } from './store/connection'
import { useTransferStore } from './store/transfer'
import type { connection } from '../wailsjs/go/models'
import { 
  Search, Setting, User, Plus, Link, Edit, 
  Download, ArrowLeft, ArrowRight, RefreshRight, 
  FolderOpened, InfoFilled, Moon, Sunny, CopyDocument, Delete
} from '@element-plus/icons-vue'
import { EventsOn } from '../wailsjs/runtime/runtime'

const showAddConnection = ref(false)
const editingConn = ref<connection.Connection | null>(null)
const connections = ref<connection.Connection[]>([])
const connStore = useConnectionStore()
const transferStore = useTransferStore()
const transferPanelRef = ref<any>(null)
const activeConnId = computed(() => connStore.activeConnectionId)
const isDark = computed(() => connStore.isDark)

const editablePath = ref('')

// Watch for store path changes and sync local editablePath
watch(() => connStore.getFullPath(), (newPath) => {
  editablePath.value = newPath
}, { immediate: true })

const syncPath = () => {
  editablePath.value = connStore.getFullPath()
}

const handlePathNavigation = () => {
  const path = editablePath.value.trim()
  if (!path || path === 'root') {
    connStore.setBucket('')
    return
  }

  // Parse path: bucket/prefix
  const parts = path.split('/')
  const bucket = parts[0]
  const prefix = parts.slice(1).join('/')
  
  connStore.setBucket(bucket)
  if (prefix) {
    connStore.setPrefix(prefix.endsWith('/') ? prefix : prefix + '/')
  }
}

// Watch for dark mode changes and apply to root element
watch(isDark, (val) => {
  if (val) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}, { immediate: true })

const openAddDialog = () => {
  editingConn.value = null
  showAddConnection.value = true
}

const editConnection = (conn: connection.Connection) => {
  editingConn.value = conn
  showAddConnection.value = true
}

const loadConnections = async () => {
  try {
    const data = await GetConnections()
    connections.value = data || []
  } catch (err) {
    console.error("加载连接失败:", err)
  }
}

const selectConnection = (id: string) => {
  connStore.setCurrentConnection(id)
}

const duplicateConnection = async (id: string) => {
  try {
    const success = await DuplicateConnection(id)
    if (success) {
      ElMessage.success('复制成功')
      loadConnections()
    } else {
      ElMessage.error('复制失败')
    }
  } catch (err) {
    ElMessage.error('复制出错: ' + err)
  }
}

const deleteConnection = async (id: string) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除该连接吗？此操作不可撤销。',
      '删除确认',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    const success = await DeleteConnection(id)
    if (success) {
      ElMessage.success('删除成功')
      if (connStore.activeConnectionId === id) {
        connStore.setCurrentConnection('')
      }
      loadConnections()
    } else {
      ElMessage.error('删除失败')
    }
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('删除出错: ' + err)
    }
  }
}

const openTransferPanel = () => {
  if (transferPanelRef.value) {
    transferPanelRef.value.visible = true
  }
}

onMounted(() => {
  loadConnections()
  EventsOn('transfer-progress', (data: any) => {
    if (data && data.id) {
      transferStore.updateProgress(data.id, data.transferred, data.total)
    }
  })
  EventsOn('transfer-complete', (data: any) => {
    if (data && data.id) {
      transferStore.completeTask(data.id)
    }
  })
  EventsOn('transfer-error', (data: any) => {
    if (data && data.id) {
      transferStore.failTask(data.id, data.error || '传输失败')
    }
  })
})
</script>


<style scoped>
.nav-btn {
  @apply w-7 h-7 flex items-center justify-center rounded-md text-gray-500 hover:bg-white/10 hover:text-textLight transition-all active:scale-95;
}
</style>
