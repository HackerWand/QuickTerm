<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import {
  NModal,
  NForm,
  NFormItem,
  NInput,
  NList,
  NListItem,
  NSwitch,
  NSpace,
  NFlex,
  NButton,
  NIcon,
  NThing,
  useMessage,
  NPopconfirm,
  NTag,
  NSelect
} from 'naive-ui'
import { Add, TrashOutline, DownloadOutline } from '@vicons/ionicons5'
import type { Workspace } from '../types'
import { useWorkspaceStore } from '../stores/workspace'
import { onFileDrop, setupFileDropListener } from '../utils/fileDrop'
import { GetDirectoryPath, ExportWorkspace, SaveFileDialog, WriteFile } from '../../wailsjs/go/main/App'


interface Props {
  show: boolean
}

interface Emits {
  (e: 'update:show', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
const message = useMessage()
const workspaceStore = useWorkspaceStore()

const localShow = ref(false)
const editingWorkspace = ref<Partial<Workspace>>({})
const newIgnorePattern = ref('')
const newIgnoreIsRegex = ref(0)
const isLoading = ref(false)

watch(
  () => props.show,
  (newVal) => {
    localShow.value = newVal
    if (newVal && workspaceStore.currentWorkspace) {
      editingWorkspace.value = {
        name: workspaceStore.currentWorkspace.name,
        path: workspaceStore.currentWorkspace.path,
        ignoredCommands: [...workspaceStore.currentWorkspace.ignoredCommands]
      }
    }
  }
)

watch(
  localShow,
  (newVal) => {
    emit('update:show', newVal)
  }
)

const saveSettings = async () => {
  if (!editingWorkspace.value.name?.trim()) {
    message.error('工作空间名称不能为空')
    return
  }
  try {
    await workspaceStore.updateWorkspace(editingWorkspace.value)
    emit('update:show', false)
    message.success('设置已保存')
  } catch (error) {
    message.error('保存设置失败')
    console.error(error)
  }
}

const handleExport = async () => {
  if (!workspaceStore.currentWorkspace) return
  try {
    isLoading.value = true
    const exportData = await ExportWorkspace(workspaceStore.currentWorkspace.id)
    const fileName = `${workspaceStore.currentWorkspace.name.replace(/[^\w\s]/g, '_')}_workspace.json`
    const savePath = await SaveFileDialog(fileName)
    if (savePath) {
      await WriteFile(savePath, JSON.stringify(exportData, null, 2))
      message.success('工作空间导出成功')
    }
  } catch (error) {
    message.error('导出工作空间失败')
    console.error(error)
  } finally {
    isLoading.value = false
  }
}

const addIgnoreRule = () => {
  if (!newIgnorePattern.value.trim()) {
    message.error('忽略模式不能为空')
    return
  }
  if (!editingWorkspace.value.ignoredCommands) {
    editingWorkspace.value.ignoredCommands = []
  }
  if (newIgnoreIsRegex.value) {
    try {
      new RegExp(newIgnorePattern.value)
    } catch (e) {
      message.error('正则表达式格式错误')
      return
    }
  }
  if (editingWorkspace.value.ignoredCommands.some(rule => {
    return rule.pattern === newIgnorePattern.value.trim() && rule.isRegex === (newIgnoreIsRegex.value === 1)
  })) {
    message.error('忽略模式已存在')
    return
  }
  editingWorkspace.value.ignoredCommands.push({
    pattern: newIgnorePattern.value.trim(),
    isRegex: newIgnoreIsRegex.value === 1
  })
  newIgnorePattern.value = ''
  newIgnoreIsRegex.value = 0
}

const removeIgnoreRule = (index: number) => {
  if (editingWorkspace.value.ignoredCommands) {
    editingWorkspace.value.ignoredCommands.splice(index, 1)
  }
}

const handleFileDrop = async (event: { x: number; y: number; paths: string[] }) => {
  try {
    if (event.paths.length > 0 && localShow.value) {
      const path = event.paths[0]
      const dirPath = await GetDirectoryPath(path)
      editingWorkspace.value.path = dirPath
      message.success(`工作空间路径已设置为: ${dirPath}`)
    }
  } catch (error) {
    message.error('处理路径失败')
    console.error(error)
  }
}

let unsubscribe: (() => void) | null = null

onMounted(() => {
  setupFileDropListener()
  unsubscribe = onFileDrop(handleFileDrop)
})

onUnmounted(() => {
  if (unsubscribe) {
    unsubscribe()
  }
})
</script>

<template>
  <NModal v-model:show="localShow" preset="card" title="工作空间设置" style="width: 600px;">
    <NForm :model="editingWorkspace" label-placement="left" label-width="100">
      <NFormItem label="工作空间名称">
        <NInput v-model:value="editingWorkspace.name" placeholder="请输入工作空间名称" />
      </NFormItem>
      <NFormItem label="工作空间路径">
        <NInput v-model:value="editingWorkspace.path" placeholder="请输入工作空间路径" />
      </NFormItem>
      <NFormItem label="忽略命令">
        <div class="ignore-commands-section">
          <NList bordered v-if="editingWorkspace.ignoredCommands?.length">
            <NListItem v-for="(rule, index) in editingWorkspace.ignoredCommands" :key="index">
              <NThing>
                <template #header>
                  {{ rule.pattern }}
                </template>
                  <NTag size="small" :bordered="false" :type="rule.isRegex ? 'info' : 'primary'">{{ rule.isRegex ? '正则' : '字符串' }}</NTag>
                <template #header-extra>
                  <NPopconfirm @positive-click="removeIgnoreRule(index)">
                    确认删除吗？
                    <template #trigger>
                      <NButton text size="tiny">
                        <template #icon>
                          <NIcon><TrashOutline /></NIcon>
                        </template>
                      </NButton>
                    </template>
                  </NPopconfirm>
                </template>
              </NThing>
            </NListItem>
          </NList>
          <NFlex justify="space-between" align="center">
            <NInputGroup>
              <NSelect style="width: 90px;" v-model:value="newIgnoreIsRegex" placeholder="识别模式" :options="[{ label: '正则', value: 1 }, { label: '字符串', value: 0 }]"></NSelect>
              <NInput 
                v-model:value="newIgnorePattern" 
                placeholder="输入忽略模式"
                style="flex: 1;"
                @keyup.enter="addIgnoreRule"
              />
              <NButton @click="addIgnoreRule">
                <template #icon>
                  <NIcon><Add /></NIcon>
                </template>
                添加
              </NButton>
            </NInputGroup>
          </NFlex>
        </div>
      </NFormItem>
    </NForm>
    <NAlert type="info" :bordered="false">可拖拽文件或目录到此处将自动设置为工作空间路径</NAlert>
    <template #footer>
      <NSpace justify="space-between">
        <NButton type="info" @click="handleExport" :loading="isLoading">
          <template #icon>
            <NIcon><DownloadOutline /></NIcon>
          </template>
          导出工作空间
        </NButton>
        <NSpace>
          <NButton @click="localShow = false">取消</NButton>
          <NButton type="primary" @click="saveSettings">保存</NButton>
        </NSpace>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped>
.ignore-commands-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.rule-type-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  background-color: var(--n-color-info);
  color: var(--n-color-info-text);
}

.rule-pattern {
  flex: 1;
  margin-left: 12px;
  font-family: monospace;
  font-size: 13px;
}
</style>
