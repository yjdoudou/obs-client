import { defineStore } from 'pinia'

export const useConnectionStore = defineStore('connection', {
    state: () => ({
        activeConnectionId: '',
        currentBucket: '',
        currentLocation: '',
        currentPrefix: '',
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

        getFullPath() {
            if (!this.currentBucket) return 'root'
            return `${this.currentBucket}/${this.currentPrefix}`
        }
    }
})
