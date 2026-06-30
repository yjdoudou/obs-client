import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { GetAllThemes, SaveTheme, SaveBackgroundImage, GetBackgroundInfo, SaveOverlayOpacity, ClearBackgroundImage } from '../../wailsjs/go/app/App'

export interface ThemeVariable {
  'primary-color': string
  'bg-app': string
  'bg-card': string
  'text-main': string
  'text-light': string
  'border-color': string
  'glass-bg': string
}

export interface ThemeConfig {
  id: string
  name: string
  description: string
  preview: string
  glassEnabled: boolean
  variables: ThemeVariable
}

export const useThemeStore = defineStore('theme', () => {
  const themes = ref<ThemeConfig[]>([])
  const currentThemeId = ref('glass-light')
  const isDark = ref(false)
  
  const backgroundImage = ref('')
  const overlayOpacity = ref(0.5)
  const hasBackground = ref(false)

  const currentTheme = computed(() => {
    return themes.value.find(t => t.id === currentThemeId.value) || themes.value[0]
  })

  const loadThemes = async () => {
    try {
      const data = await GetAllThemes()
      themes.value = data || []
      
      const savedThemeId = localStorage.getItem('theme-id')
      const savedDark = localStorage.getItem('theme-dark') === 'true'
      
      if (savedThemeId) {
        const exists = themes.value.some(t => t.id === savedThemeId)
        if (exists) {
          currentThemeId.value = savedThemeId
        }
      }
      isDark.value = savedDark
      
      applyTheme(currentTheme.value)
      loadBackgroundInfo()
    } catch (error) {
      console.error('加载主题失败:', error)
      themes.value = getDefaultThemes()
      applyTheme(themes.value[0])
    }
  }

  const loadBackgroundInfo = async () => {
    try {
      const info = await GetBackgroundInfo('default-user')
      backgroundImage.value = info.imageBase64 || ''
      overlayOpacity.value = info.opacity
      hasBackground.value = info.hasImage
      applyBackground()
    } catch (error) {
      console.error('加载背景信息失败:', error)
    }
  }

  const setTheme = async (themeId: string) => {
    const theme = themes.value.find(t => t.id === themeId)
    if (!theme) return

    currentThemeId.value = themeId
    isDark.value = themeId.includes('dark')
    
    localStorage.setItem('theme-id', themeId)
    localStorage.setItem('theme-dark', String(isDark.value))
    
    applyTheme(theme)

    try {
      await SaveTheme('default-user', themeId, isDark.value)
    } catch (error) {
      console.error('保存主题失败:', error)
    }
  }

  const applyTheme = (theme: ThemeConfig) => {
    const root = document.documentElement
    
    if (theme.glassEnabled) {
      root.classList.add('glass-mode')
    } else {
      root.classList.remove('glass-mode')
    }
    
    if (theme.id.includes('dark')) {
      root.classList.add('dark')
    } else {
      root.classList.remove('dark')
    }

    Object.entries(theme.variables).forEach(([key, value]) => {
      root.style.setProperty(`--${key}`, value as string)
    })

    root.style.setProperty('--bg-overlay-opacity', String(overlayOpacity.value))
    
    const overlayLayer = document.getElementById('app-overlay')
    if (overlayLayer && hasBackground.value) {
      overlayLayer.style.backgroundColor = theme.variables['bg-app'] || '#ffffff'
    }
    
    applyElementPlusTheme(theme)
    applyBackground()
  }

  const applyElementPlusTheme = (theme: ThemeConfig) => {
    const root = document.documentElement
    
    root.style.setProperty('--el-color-primary', theme.variables['primary-color'])
    root.style.setProperty('--el-text-color-primary', theme.variables['text-main'])
    root.style.setProperty('--el-text-color-regular', theme.variables['text-light'])
    root.style.setProperty('--el-bg-color-page', theme.variables['bg-app'])
    root.style.setProperty('--el-bg-color', theme.variables['bg-card'])
    root.style.setProperty('--el-border-color', theme.variables['border-color'])
    root.style.setProperty('--el-border-color-light', theme.variables['border-color'])
  }

  const applyBackground = () => {
    const root = document.documentElement
    const bgLayer = document.getElementById('app-background')
    const overlayLayer = document.getElementById('app-overlay')
    
    if (!bgLayer || !overlayLayer) return
    
    if (hasBackground.value && backgroundImage.value) {
      bgLayer.style.backgroundImage = `url(${backgroundImage.value})`
      bgLayer.style.display = 'block'
      overlayLayer.style.display = 'block'
      overlayLayer.style.backgroundColor = getComputedStyle(root).getPropertyValue('--bg-app') || '#ffffff'
      overlayLayer.style.opacity = String(overlayOpacity.value)
      root.classList.add('has-background')
    } else {
      bgLayer.style.backgroundImage = ''
      bgLayer.style.display = 'none'
      overlayLayer.style.display = 'none'
      root.classList.remove('has-background')
    }
  }

  const uploadBackground = async (file: File) => {
    try {
      const reader = new FileReader()
      return new Promise<void>((resolve, reject) => {
        reader.onload = async (e) => {
          const result = e.target?.result as string
          const base64Data = result.split(',')[1]
          
          if (base64Data) {
            const success = await SaveBackgroundImage('default-user', base64Data)
            if (success) {
              backgroundImage.value = result
              hasBackground.value = true
              applyBackground()
              resolve()
            } else {
              reject(new Error('保存背景失败'))
            }
          } else {
            reject(new Error('图片解析失败'))
          }
        }
        reader.onerror = reject
        reader.readAsDataURL(file)
      })
    } catch (error) {
      console.error('上传背景失败:', error)
      throw error
    }
  }

  const setOverlayOpacity = async (opacity: number) => {
    overlayOpacity.value = opacity
    const root = document.documentElement
    root.style.setProperty('--bg-overlay-opacity', String(opacity))
    
    const overlayLayer = document.getElementById('app-overlay')
    if (overlayLayer) {
      overlayLayer.style.opacity = String(opacity)
    }
    
    try {
      await SaveOverlayOpacity('default-user', opacity)
    } catch (error) {
      console.error('保存透明度失败:', error)
    }
  }

  const removeBackground = async () => {
    try {
      await ClearBackgroundImage('default-user')
      backgroundImage.value = ''
      hasBackground.value = false
      applyBackground()
    } catch (error) {
      console.error('清除背景失败:', error)
    }
  }

  const getDefaultThemes = (): ThemeConfig[] => [
    {
      id: 'glass-light',
      name: '玻璃亮色',
      description: '透明玻璃效果，现代感十足',
      preview: '#f5f7fa',
      glassEnabled: true,
      variables: {
        'primary-color': '#0066ff',
        'bg-app': 'rgba(245, 247, 250, 0.6)',
        'bg-card': 'rgba(255, 255, 255, 0.8)',
        'text-main': '#333333',
        'text-light': '#666666',
        'border-color': 'rgba(0, 0, 0, 0.08)',
        'glass-bg': 'rgba(255, 255, 255, 0.6)',
      },
    },
    {
      id: 'glass-dark',
      name: '玻璃暗色',
      description: '深色透明玻璃效果，护眼模式',
      preview: '#121212',
      glassEnabled: true,
      variables: {
        'primary-color': '#0066ff',
        'bg-app': 'rgba(18, 18, 18, 0.6)',
        'bg-card': 'rgba(30, 30, 30, 0.8)',
        'text-main': '#e0e0e0',
        'text-light': '#888888',
        'border-color': 'rgba(255, 255, 255, 0.08)',
        'glass-bg': 'rgba(30, 30, 30, 0.6)',
      },
    },
    {
      id: 'solid-light',
      name: '经典亮色',
      description: '传统不透明风格，简洁清爽',
      preview: '#ffffff',
      glassEnabled: false,
      variables: {
        'primary-color': '#0066ff',
        'bg-app': '#ffffff',
        'bg-card': '#ffffff',
        'text-main': '#333333',
        'text-light': '#666666',
        'border-color': '#e8e8e8',
        'glass-bg': '#ffffff',
      },
    },
    {
      id: 'solid-dark',
      name: '经典暗色',
      description: '深色不透明风格，沉稳专业',
      preview: '#1a1a1a',
      glassEnabled: false,
      variables: {
        'primary-color': '#0066ff',
        'bg-app': '#1a1a1a',
        'bg-card': '#242424',
        'text-main': '#e0e0e0',
        'text-light': '#888888',
        'border-color': '#333333',
        'glass-bg': '#242424',
      },
    },
    {
      id: 'minimal-light',
      name: '极简亮色',
      description: '极简风格，专注内容',
      preview: '#fafafa',
      glassEnabled: false,
      variables: {
        'primary-color': '#0070f3',
        'bg-app': '#fafafa',
        'bg-card': '#ffffff',
        'text-main': '#213547',
        'text-light': '#6b7280',
        'border-color': '#e5e7eb',
        'glass-bg': '#ffffff',
      },
    },
    {
      id: 'ocean-blue',
      name: '海洋蓝',
      description: '清新蓝色主题，舒适视觉',
      preview: '#f0f9ff',
      glassEnabled: true,
      variables: {
        'primary-color': '#0ea5e9',
        'bg-app': 'rgba(240, 249, 255, 0.7)',
        'bg-card': 'rgba(255, 255, 255, 0.9)',
        'text-main': '#0c4a6e',
        'text-light': '#075985',
        'border-color': 'rgba(14, 165, 233, 0.1)',
        'glass-bg': 'rgba(255, 255, 255, 0.7)',
      },
    },
  ]

  return {
    themes,
    currentThemeId,
    isDark,
    currentTheme,
    backgroundImage,
    overlayOpacity,
    hasBackground,
    loadThemes,
    setTheme,
    uploadBackground,
    setOverlayOpacity,
    removeBackground,
  }
})
