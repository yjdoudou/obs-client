<template>
  <div class="theme-panel">
    <div class="flex items-center justify-between mb-4">
      <span class="font-bold text-sm text-textMain">选择皮肤</span>
      <button 
        @click="$emit('close')" 
        class="w-6 h-6 rounded-full hover:bg-black/5 dark:hover:bg-white/10 flex items-center justify-center transition-colors"
      >
        <el-icon size="14"><Close /></el-icon>
      </button>
    </div>
    
    <div class="grid grid-cols-2 gap-3 mb-4">
      <div 
        v-for="theme in themeStore.themes" 
        :key="theme.id"
        @click="themeStore.setTheme(theme.id)"
        class="theme-card cursor-pointer rounded-xl p-3 border-2 transition-all duration-200 hover:scale-[1.02]"
        :class="themeStore.currentThemeId === theme.id ? 'border-primary shadow-md shadow-primary/20' : 'border-transparent hover:border-black/5 dark:hover:border-white/10'"
      >
        <div 
          class="w-full h-12 rounded-lg mb-2 shadow-inner"
          :style="{ backgroundColor: theme.preview }"
        ></div>
        <div class="flex items-center justify-between">
          <div>
            <div class="font-medium text-xs text-textMain">{{ theme.name }}</div>
            <div class="text-[10px] text-textLight mt-0.5">{{ theme.description }}</div>
          </div>
          <div 
            v-if="themeStore.currentThemeId === theme.id"
            class="w-5 h-5 rounded-full bg-primary flex items-center justify-center"
          >
            <el-icon size="10" class="text-white"><Check /></el-icon>
          </div>
        </div>
        <div class="flex items-center gap-1 mt-2">
          <span 
            v-if="theme.glassEnabled"
            class="px-1.5 py-0.5 rounded text-[9px] bg-primary/10 text-primary"
          >
            玻璃
          </span>
          <span 
            v-if="theme.id.includes('dark')"
            class="px-1.5 py-0.5 rounded text-[9px] bg-black/10 dark:bg-white/10 text-textLight"
          >
            暗色
          </span>
          <span 
            v-if="theme.id.includes('light')"
            class="px-1.5 py-0.5 rounded text-[9px] bg-black/5 dark:bg-white/5 text-textLight"
          >
            亮色
          </span>
        </div>
      </div>
    </div>
    
    <div class="border-t border-black/5 dark:border-white/5 pt-4">
      <div class="flex items-center justify-between mb-3">
        <span class="font-medium text-xs text-textMain">自定义背景</span>
        <button 
          v-if="themeStore.hasBackground"
          @click="removeBackground"
          class="text-[10px] text-red-500 hover:text-red-600 transition-colors flex items-center gap-1"
        >
          <el-icon size="12"><Delete /></el-icon>
          移除
        </button>
      </div>
      
      <div 
        v-if="themeStore.hasBackground && themeStore.backgroundImage"
        class="w-full h-20 rounded-lg bg-cover bg-center mb-3 border border-black/5 dark:border-white/5 relative overflow-hidden"
        :style="{ backgroundImage: `url(${themeStore.backgroundImage})` }"
      >
        <div class="absolute inset-0 bg-black/30 flex items-center justify-center">
          <span class="text-white text-xs font-medium">当前背景预览</span>
        </div>
      </div>
      
      <button 
        @click="triggerUpload"
        class="w-full h-12 rounded-lg border-2 border-dashed border-black/10 dark:border-white/10 hover:border-primary/50 transition-colors flex flex-col items-center justify-center gap-1"
      >
        <el-icon size="18" class="text-textLight"><UploadFilled /></el-icon>
        <span class="text-[11px] text-textLight">上传背景图片</span>
      </button>
      <input 
        ref="fileInput"
        type="file" 
        accept="image/*" 
        class="hidden"
        @change="handleFileSelect"
      />
      
      <div v-if="themeStore.hasBackground" class="mt-4">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[11px] text-textLight">遮罩透明度</span>
          <span class="text-[10px] text-textMain">{{ Math.round(themeStore.overlayOpacity * 100) }}%</span>
        </div>
        <input 
          type="range" 
          min="0" 
          max="1" 
          step="0.05" 
          :value="themeStore.overlayOpacity"
          @input="handleOpacityChange"
          class="w-full h-2 rounded-full appearance-none bg-black/10 dark:bg-white/10 cursor-pointer"
          style="accent-color: #0066ff"
        />
        <div class="flex justify-between text-[9px] text-textLight mt-1">
          <span>背景清晰</span>
          <span>背景模糊</span>
        </div>
      </div>
    </div>
    
    <div class="mt-4 pt-3 border-t border-black/5 dark:border-white/5">
      <div class="text-[11px] text-textLight text-center">
        皮肤设置将自动保存到本地
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useThemeStore } from '../store/theme'
import { Close, Check, UploadFilled, Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

defineEmits(['close'])

const themeStore = useThemeStore()
const fileInput = ref<HTMLInputElement | null>(null)

const triggerUpload = () => {
  fileInput.value?.click()
}

const handleFileSelect = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  
  if (file) {
    try {
      await themeStore.uploadBackground(file)
      ElMessage.success('背景图片上传成功')
    } catch (error) {
      ElMessage.error('背景图片上传失败')
    }
  }
  
  target.value = ''
}

const handleOpacityChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  const value = parseFloat(target.value)
  themeStore.setOverlayOpacity(value)
}

const removeBackground = async () => {
  try {
    await themeStore.removeBackground()
    ElMessage.success('背景图片已移除')
  } catch (error) {
    ElMessage.error('移除背景失败')
  }
}
</script>

<style scoped>
.theme-panel {
  padding: 16px;
  background: var(--bg-card);
  border-radius: 12px;
}

.theme-card:hover {
  transform: translateY(-2px);
}

input[type="range"]::-webkit-slider-thumb {
  cursor: pointer;
}
</style>
