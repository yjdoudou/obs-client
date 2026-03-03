<template>
  <div class="h-8 flex items-center justify-between bg-transparent select-none drag z-50">
    <div class="flex items-center gap-2 pl-3 pointer-events-none">
      <img src="../assets/logo.png" class="w-4 h-4 object-contain shadow-sm" alt="logo" />
      <span class="text-[10px] font-bold tracking-widest text-textLight opacity-80">OBS CLIENT</span>
    </div>
    
    <div class="flex h-full no-drag">
      <div 
        @click="minimize" 
        class="w-12 h-full flex items-center justify-center hover:bg-black/10 dark:hover:bg-white/10 transition-colors cursor-default"
      >
        <el-icon size="14"><SemiSelect /></el-icon>
      </div>
      <div 
        @click="toggleMaximize" 
        class="w-12 h-full flex items-center justify-center hover:bg-black/10 dark:hover:bg-white/10 transition-colors cursor-default"
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
.drag {
  --wails-draggable: drag;
}
.no-drag {
  --wails-draggable: no-drag;
}
</style>
