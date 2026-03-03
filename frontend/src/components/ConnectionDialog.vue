<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? '编辑连接' : '新建连接'"
    width="500px"
    class="connection-dialog"
    :close-on-click-modal="false"
    destroy-on-close
  >
    <div class="px-2">
      <el-form :model="form" :rules="rules" ref="formRef" label-position="top">
        <el-form-item label="连接名称" prop="name">
          <el-input v-model="form.name" placeholder="例如: 我的华为云" class="modern-input" />
        </el-form-item>
        
        <div class="grid grid-cols-2 gap-4">
          <el-form-item label="Access Key (AK)" prop="accessKeyId">
            <el-input v-model="form.accessKeyId" placeholder="输入 AK" class="modern-input" />
          </el-form-item>
          <el-form-item label="Secret Key (SK)" prop="secretAccessKey">
            <el-input v-model="form.secretAccessKey" type="password" show-password placeholder="输入 SK" class="modern-input" />
          </el-form-item>
        </div>

        <el-form-item label="Region (可选)" prop="region">
          <el-input v-model="form.region" placeholder="例如: cn-north-4" class="modern-input" />
        </el-form-item>

        <el-form-item label="Endpoint (可选)" prop="endpoint">
          <el-input v-model="form.endpoint" placeholder="例如: https://obs.cn-north-4.myhuaweicloud.com" class="modern-input" />
        </el-form-item>
      </el-form>
    </div>

    <template #footer>
      <div class="flex items-center justify-between px-2 pb-2">
        <el-button 
          :loading="testing" 
          plain 
          class="!border-primary/30 !text-primary hover:!bg-primary/10"
          @click="handleTest"
        >
          <el-icon class="mr-1.5"><MagicStick /></el-icon>
          测试连接
        </el-button>
        <div class="flex gap-2">
          <el-button @click="visible = false" class="!bg-black/5 dark:!bg-white/5 !border-none text-gray-400">取消</el-button>
          <el-button type="primary" :loading="saving" @click="handleSave" class="shadow-lg shadow-primary/20">
            {{ isEdit ? '更新连接' : '保存连接' }}
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { TestConnection, CreateConnection, UpdateConnection } from '../../wailsjs/go/app/App'
import { connection } from '../../wailsjs/go/models'
import { MagicStick } from '@element-plus/icons-vue'

const visible = defineModel<boolean>('visible', { default: false })
const emit = defineEmits(['saved'])

const props = defineProps<{
  editData: connection.Connection | null
}>()

const formRef = ref<FormInstance>()
const testing = ref(false)
const saving = ref(false)
const isEdit = ref(false)

const form = reactive({
  id: '',
  name: '',
  accessKeyId: '',
  secretAccessKey: '',
  region: '',
  endpoint: '',
})

const rules = reactive<FormRules>({
  name: [{ required: true, message: '请输入连接名称', trigger: 'blur' }],
  accessKeyId: [{ required: true, message: '请输入 Access Key', trigger: 'blur' }],
  secretAccessKey: [{ required: true, message: '请输入 Secret Key', trigger: 'blur' }],
})

watch(() => props.editData, (val) => {
  if (val) {
    isEdit.value = true
    form.id = val.id
    form.name = val.name
    form.accessKeyId = val.accessKeyId
    form.secretAccessKey = val.secretAccessKey
    form.region = val.region
    form.endpoint = val.endpoint
  } else {
    isEdit.value = false
    form.id = ''
    form.name = ''
    form.accessKeyId = ''
    form.secretAccessKey = ''
    form.region = ''
    form.endpoint = ''
  }
}, { immediate: true })

const handleTest = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      testing.value = true
      try {
        const conn = new connection.Connection({
          id: form.id,
          name: form.name,
          accessKeyId: form.accessKeyId,
          secretAccessKey: form.secretAccessKey,
          region: form.region,
          endpoint: form.endpoint
        })
        const ok = await TestConnection(conn)
        if (ok) ElMessage.success('测试成功')
        else ElMessage.error('连接失败，请检查配置')
      } catch (err: any) {
        ElMessage.error('测试出错: ' + (err.message || err))
      } finally {
        testing.value = false
      }
    }
  })
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      saving.value = true
      try {
        const conn = new connection.Connection({
          id: form.id,
          name: form.name,
          accessKeyId: form.accessKeyId,
          secretAccessKey: form.secretAccessKey,
          region: form.region,
          endpoint: form.endpoint,
          delimiter: '/'
        })
        
        let success = false
        if (isEdit.value) {
          success = await UpdateConnection(conn)
        } else {
          const res = await CreateConnection(conn)
          success = !!res
        }

        if (success) {
          ElMessage.success(isEdit.value ? '更新成功' : '保存成功')
          visible.value = false
          emit('saved')
        }
      } catch (err: any) {
        ElMessage.error('保存失败: ' + (err.message || err))
      } finally {
        saving.value = false
      }
    }
  })
}
</script>

<style scoped>
:deep(.modern-input .el-input__wrapper) {
  @apply !bg-black/5 dark:!bg-white/5 !border-black/5 dark:!border-white/10 !rounded-lg !h-10 !transition-all;
}

:deep(.modern-input .el-input__wrapper.is-focus) {
  @apply !border-primary/50 !shadow-lg shadow-primary/5;
}

:deep(.el-form-item__label) {
  @apply !text-[11px] !font-bold !uppercase !tracking-wider !mb-1 text-textLight;
}

:deep(.connection-dialog.el-dialog) {
  @apply !bg-bgCard !backdrop-blur-xl border border-black/5 dark:border-white/5 shadow-premium !rounded-2xl overflow-hidden;
}

:deep(.el-dialog__header) {
  @apply !pt-6 !px-6 !mb-0;
}

:deep(.el-dialog__title) {
  @apply !text-base !font-bold !text-textMain;
}

:deep(.el-dialog__body) {
  @apply !px-6 !py-4;
}

:deep(.el-dialog__footer) {
  @apply !px-6 !pb-6 !pt-0;
}
</style>
