<template>
  <div class="fixed bottom-6 right-6 z-50">
    <!-- Float Button -->
    <el-button 
      type="primary" 
      class="!w-14 !h-14 !rounded-full shadow-premium !border-none !bg-primary animate-glow flex items-center justify-center p-0"
      @click="visible = !visible"
    >
      <el-badge :value="activeCount" :hidden="activeCount === 0" class="flex items-center justify-center">
        <el-icon :size="24" class="text-white"><Switch /></el-icon>
      </el-badge>
    </el-button>

    <!-- Transfer Panel -->
    <transition name="el-zoom-in-bottom">
      <div v-if="visible" class="absolute bottom-20 right-0 w-[400px] max-h-[500px] bg-bgCard/95 backdrop-blur-xl border border-white/10 rounded-3xl shadow-2xl flex flex-col overflow-hidden">
        <div class="px-6 py-4 border-b border-black/5 dark:border-white/5 flex items-center justify-between bg-white/5">
          <h3 class="text-base font-bold text-textMain flex items-center gap-2">
            传输列表
            <span v-if="tasks.length" class="text-xs font-normal text-textLight">({{ tasks.length }})</span>
          </h3>
          <div class="flex items-center gap-2">
            <el-button link size="small" @click="clearCompleted" v-if="hasCompleted">清除完成</el-button>
            <el-button link size="small" @click="visible = false">
              <el-icon><Close /></el-icon>
            </el-button>
          </div>
        </div>

        <div class="flex-1 overflow-y-auto p-4 custom-scrollbar">
          <div v-if="tasks.length === 0" class="h-40 flex flex-col items-center justify-center text-textLight opacity-50">
            <el-icon :size="48"><Files /></el-icon>
            <p class="mt-2 text-sm">暂无传输任务</p>
          </div>
          
          <div v-else class="space-y-4">
            <div v-for="task in tasks" :key="task.id" class="p-3 rounded-2xl bg-black/5 dark:bg-white/5 border border-transparent hover:border-primary/20 transition-all group">
              <div class="flex items-start gap-3">
                <div :class="[
                  'w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0',
                  task.type === 'upload' ? 'bg-blue-500/10 text-blue-500' : 'bg-green-500/10 text-green-500'
                ]">
                  <el-icon :size="20">
                    <component :is="task.type === 'upload' ? 'Upload' : 'Download'" />
                  </el-icon>
                </div>
                
                <div class="flex-1 min-w-0">
                  <div class="flex items-center justify-between mb-1">
                    <span class="text-sm font-semibold text-textMain truncate pr-2" :title="task.name">
                      {{ task.name }}
                    </span>
                    <el-button link size="small" class="opacity-0 group-hover:opacity-100 transition-opacity" @click="removeTask(task.id)">
                      <el-icon><Close /></el-icon>
                    </el-button>
                  </div>
                  
                  <div class="flex items-center justify-between text-[11px] text-textLight mb-2">
                    <span>{{ formatSize(task.transferred) }} / {{ formatSize(task.size) }}</span>
                    <span :class="getStatusClass(task.status)">{{ getStatusText(task.status) }}</span>
                  </div>

                  <el-progress 
                    :percentage="task.progress" 
                    :status="task.status === 'completed' ? 'success' : (task.status === 'error' ? 'exception' : '')"
                    :stroke-width="6"
                    :show-text="false"
                    class="custom-progress"
                  />
                  
                  <div v-if="task.error" class="mt-2 text-[10px] text-red-500 truncate">
                    {{ task.error }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useTransferStore } from '../store/transfer'
import { storeToRefs } from 'pinia'
import { 
  Switch, Close, Files, Upload, 
  Download, Check, Warning 
} from '@element-plus/icons-vue'

const store = useTransferStore()
const { tasks } = storeToRefs(store)
const { clearCompleted, removeTask } = store
const visible = ref(false)

const activeCount = computed(() => 
  tasks.value.filter(t => t.status === 'transferring' || t.status === 'pending').length
)

const hasCompleted = computed(() => 
  tasks.value.some(t => t.status === 'completed' || t.status === 'error')
)

const formatSize = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'pending': return '等待中'
    case 'transferring': return '传输中'
    case 'completed': return '已完成'
    case 'error': return '出错'
    default: return status
  }
}

const getStatusClass = (status: string) => {
  switch (status) {
    case 'completed': return 'text-green-500'
    case 'error': return 'text-red-500 font-bold'
    case 'transferring': return 'text-primary'
    default: return ''
  }
}

defineExpose({
  visible
})
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.05);
  border-radius: 10px;
}
.dark .custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.05);
}

.custom-progress :deep(.el-progress-bar__outer) {
  background-color: rgba(0, 0, 0, 0.05) !important;
}
.dark .custom-progress :deep(.el-progress-bar__outer) {
  background-color: rgba(255, 255, 255, 0.05) !important;
}

.animate-glow {
  animation: bg-glow 4s infinite alternate;
}

@keyframes bg-glow {
  from { box-shadow: 0 0 10px rgba(var(--primary-rgb), 0.2); }
  to { box-shadow: 0 0 20px rgba(var(--primary-rgb), 0.5); }
}

.shadow-premium {
  box-shadow: 0 10px 30px -10px rgba(var(--primary-rgb), 0.5);
}
</style>
