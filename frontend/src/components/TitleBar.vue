<template>
  <div class="h-8 flex items-center justify-between bg-transparent select-none" style="--wails-drop-target: drag">
    <div class="flex items-center gap-2 pl-4 pointer-events-none">
      <div class="w-4 h-4 rounded-sm bg-primary/20 flex items-center justify-center">
        <div class="w-2 h-2 rounded-[1px] bg-primary"></div>
      </div>
      <span class="text-[10px] font-bold tracking-widest text-textLight">OBS CLIENT</span>
    </div>
    
    <div class="flex h-full no-drag" style="--wails-drop-target: none">
      <div 
        @click="minimize" 
        class="w-12 h-full flex items-center justify-center hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-default"
      >
        <el-icon size="14"><SemiSelect /></el-icon>
      </div>
      <div 
        @click="toggleMaximize" 
        class="w-12 h-full flex items-center justify-center hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-default"
      >
        <el-icon size="12"><CopyDocument v-if="isMaximized" /><FullScreen v-else /></el-icon>
      </div>
      <div 
        @click="close" 
        class="w-12 h-full flex items-center justify-center hover:bg-red-500 hover:text-white transition-colors cursor-default"
      >
        <el-icon size="14"><Close /></el-icon>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { WindowMinimise, WindowToggleMaximise, Quit, WindowIsMaximised } from '../../wailsjs/runtime/runtime'
import { SemiSelect, FullScreen, CopyDocument, Close } from '@element-plus/icons-vue'

const isMaximized = ref(false)

const minimize = () => WindowMinimise()
const toggleMaximize = async () => {
  await WindowToggleMaximise()
  isMaximized.value = await WindowIsMaximised()
}
const close = () => Quit()

onMounted(async () => {
  // Check initial state
  isMaximized.value = await WindowIsMaximised()
  
  // You might want to listen for window changes if wails emits them
})
</script>

<style scoped>
.no-drag {
  -webkit-app-region: no-drag;
}
</style>
