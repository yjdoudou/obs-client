import { defineStore } from 'pinia'

export interface TransferTask {
    id: string
    name: string
    type: 'upload' | 'download'
    size: number
    transferred: number
    status: 'pending' | 'transferring' | 'completed' | 'error'
    progress: number
    error?: string
    timestamp: number
}

export const useTransferStore = defineStore('transfer', {
    state: () => ({
        tasks: [] as TransferTask[],
    }),
    actions: {
        addTask(task: Omit<TransferTask, 'progress' | 'transferred' | 'status' | 'timestamp'>) {
            this.tasks.unshift({
                ...task,
                transferred: 0,
                progress: 0,
                status: 'pending',
                timestamp: Date.now()
            })
        },
        updateProgress(id: string, transferred: number, total: number) {
            const task = this.tasks.find(t => t.id === id)
            if (task) {
                task.transferred = transferred
                task.progress = Math.floor((transferred / total) * 100)
                task.status = 'transferring'
            }
        },
        completeTask(id: string) {
            const task = this.tasks.find(t => t.id === id)
            if (task) {
                task.progress = 100
                task.status = 'completed'
            }
        },
        failTask(id: string, error: string) {
            const task = this.tasks.find(t => t.id === id)
            if (task) {
                task.status = 'error'
                task.error = error
            }
        },
        clearCompleted() {
            this.tasks = this.tasks.filter(t => t.status !== 'completed' && t.status !== 'error')
        },
        removeTask(id: string) {
            this.tasks = this.tasks.filter(t => t.id !== id)
        }
    }
})
