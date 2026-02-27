import { defineStore } from 'pinia'

export const useConnectionStore = defineStore('connection', {
    state: () => ({
        activeConnectionId: '',
        currentBucket: '',
        currentLocation: '',
        currentPrefix: '',
        isDark: localStorage.getItem('theme') === 'dark',
        searchKeyword: ''
    }),
    actions: {
        setCurrentConnection(id: string) {
            this.activeConnectionId = id
            this.currentBucket = ''
            this.currentLocation = ''
            this.currentPrefix = ''
            this.searchKeyword = ''
        },

        setBucket(bucket: string, location: string = '') {
            this.currentBucket = bucket
            this.currentLocation = location
            this.currentPrefix = ''
        },

        setLocation(location: string) {
            this.currentLocation = location
        },

        setPrefix(prefix: string) {
            this.currentPrefix = prefix
        },

        // Get formatted full path
        getFullPath() {
            if (!this.currentBucket) return 'root'
            return `${this.currentBucket}/${this.currentPrefix}`
        },

        toggleTheme() {
            this.isDark = !this.isDark
            localStorage.setItem('theme', this.isDark ? 'dark' : 'light')
            if (this.isDark) {
                document.documentElement.classList.add('dark')
            } else {
                document.documentElement.classList.remove('dark')
            }
        }
    }
})
