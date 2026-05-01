<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  NCard,
  NButton,
  NList,
  NListItem,
  NModal,
  NForm,
  NFormItem,
  NInput,
  useMessage,
  NSpace,
  NText,
  NAlert,
  NEllipsis,
  NFlex,
  NPopconfirm,
  NIcon,
  useDialog,
  NInputGroup,
} from 'naive-ui'
import type { Workspace } from '../types'
import { models } from '../../wailsjs/go/models'
import { useWorkspaceStore } from '../stores/workspace'
import * as App from '../../wailsjs/go/main/App'
import { setupFileDropListener, onFileDrop } from '../utils/fileDrop'
import { TrashOutline, DownloadOutline, CloudUploadOutline, ServerOutline } from '@vicons/ionicons5'
import DataBackupDialog from '../components/DataBackupDialog.vue'

const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const workspaceStore = useWorkspaceStore()

const workspaces = ref<Workspace[]>([])
const showCreateModal = ref(false)
const showImportModal = ref(false)
const showBackupDialog = ref(false)
const newWorkspaceName = ref('')
const newWorkspacePath = ref('')
const isLoading = ref(false)
const importPath = ref('')
const importWorkspaceName = ref('')
const importWorkspacePath = ref('')
let importData: models.WorkspaceExport | null = null

const loadWorkspaces = async () => {
  try {
    isLoading.value = true
    const result = await App.GetWorkspaces()
    workspaces.value = result
  } catch (error) {
    dialog.error({
      title: '加载工作空间失败',
      content: (error as Error).message
    })
    console.error(error)
  } finally {
    isLoading.value = false
  }
}

const handleCreateWorkspace = async () => {
  if (!newWorkspaceName.value.trim()) {
    message.warning('请输入工作空间名称')
    return
  }
  try {
    isLoading.value = true
    await App.CreateWorkspace(newWorkspaceName.value, newWorkspacePath.value)
    showCreateModal.value = false
    newWorkspaceName.value = ''
    newWorkspacePath.value = ''
    message.success('工作空间创建成功')
    await loadWorkspaces()
  } catch (error) {
    message.error('创建工作空间失败')
    console.error(error)
  } finally {
    isLoading.value = false
  }
}

const handleOpenWorkspace = async (workspace: Workspace) => {
  try {
    await App.OpenWorkspaceWindow(workspace.id)
    workspaceStore.setWorkspace(workspace)
    await router.push({
      name: 'MainView',
      params: { id: workspace.id }
    })
  } catch (error) {
    message.error('打开工作空间失败')
    console.error(error)
  }
}

const handleDeleteWorkspace = async (workspace: Workspace) => {
  try {
    isLoading.value = true
    await App.DeleteWorkspace(workspace.id)
    message.success('工作空间删除成功')
    await loadWorkspaces()
  } catch (error) {
    if (error instanceof Error && error.message === 'cannot delete the last workspace') {
      message.error('至少需要保留一个工作空间')
    } else {
      message.error('删除工作空间失败')
    }
    console.error(error)
  } finally {
    isLoading.value = false
  }
}

const handleExportWorkspace = async (workspace: Workspace) => {
  try {
    isLoading.value = true
    const exportData = await App.ExportWorkspace(workspace.id)
    const fileName = `${workspace.name}_QuickTerm.json`
    const savePath = await App.SaveFileDialog(fileName)
    if (savePath) {
      await App.WriteFile(savePath, JSON.stringify(exportData, null, 2))
      message.success('工作空间导出成功')
    }
  } catch (error) {
    message.error('导出工作空间失败')
    console.error(error)
  } finally {
    isLoading.value = false
  }
}

const handleSelectImportFile = async () => {
  try {
    const path = await App.OpenFileDialog()
    if (path) {
      importPath.value = path
      const content = await App.ReadFile(path)
      importData = models.WorkspaceExport.createFrom(content)
      importWorkspaceName.value = importData.name
    }
  } catch (error) {
    message.error('选择文件失败')
    console.error(error)
  }
}

const handleImportWorkspace = async () => {
  if (!importData) {
    message.warning('请先选择要导入的文件')
    return
  }
  if (!importWorkspaceName.value.trim()) {
    message.warning('请输入工作空间名称')
    return
  }
  try {
    isLoading.value = true
    importData.name = importWorkspaceName.value
    await App.ImportWorkspace(importData, importWorkspacePath.value)
    showImportModal.value = false
    importData = null
    importPath.value = ''
    importWorkspaceName.value = ''
    importWorkspacePath.value = ''
    message.success('工作空间导入成功')
    await loadWorkspaces()
  } catch (error) {
    message.error('导入工作空间失败')
    console.error(error)
  } finally {
    isLoading.value = false
  }
}

const handleFileDrop = async (event: { x: number; y: number; paths: string[] }) => {
  try {
    if (event.paths.length > 0) {
      const path = event.paths[0]
      const dirPath = await App.GetDirectoryPath(path)
      newWorkspacePath.value = dirPath
      message.success(`工作空间路径已设置为: ${dirPath}`)
    }
  } catch (error) {
    message.error('处理路径失败')
    console.error(error)
  }
}

const handleDataChanged = async () => {
  await loadWorkspaces()
}

let unsubscribe: (() => void) | null = null

onUnmounted(() => {
  if (unsubscribe) {
    unsubscribe()
  }
})

onMounted(() => {
  loadWorkspaces()
  setupFileDropListener()
  unsubscribe = onFileDrop(handleFileDrop)
})
</script>

<template>
  <NCard content-scrollable title="QuickTerm" hoverable style="width: 80vw; max-width: 800px; max-height: 80vh; position: absolute; top: 10%; left: 50%; transform: translate(-50%, 0);">
    <NList v-if="workspaces.length > 0" :bordered="false">
      <NListItem v-for="workspace in workspaces" :key="workspace.id">
        <NFlex justify="space-between" align="center">
          <NFlex vertical style="flex: 1; overflow: hidden;">
            <NText strong>{{ workspace.name }}</NText>
            <NEllipsis style="font-size: 13px;">
              <NText depth="3">{{ workspace.path }}</NText>
            </NEllipsis>
          </NFlex>
          <NFlex align="center">
            <NButton size="small" @click="handleExportWorkspace(workspace)">
              <template #icon>
                <NIcon><DownloadOutline /></NIcon>
              </template>
              导出
            </NButton>
            <NButton size="small" type="primary" @click="handleOpenWorkspace(workspace)">
              打开
            </NButton>
            <NPopconfirm
              v-if="workspaces.length > 1"
              positive-text="删除"
              negative-text="取消"
              @positive-click="handleDeleteWorkspace(workspace)"
            >
              确定要删除此工作空间吗？
              <template #trigger>
                <NButton type="warning" size="small" :disabled="workspaces.length <= 1">
                  <template #icon>
                    <NIcon><TrashOutline /></NIcon>
                  </template>
                </NButton>
              </template>
            </NPopconfirm>
          </NFlex>
        </NFlex>
      </NListItem>
    </NList>
    <template #action>
      <NFlex justify="end" :size="10">
        <NButton type="default" @click="showBackupDialog = true">
          <template #icon>
            <NIcon><ServerOutline /></NIcon>
          </template>
          备份与恢复
        </NButton>
        <NButton type="default" @click="showImportModal = true">
          <template #icon>
            <NIcon><CloudUploadOutline /></NIcon>
          </template>
          导入工作空间
        </NButton>
        <NButton type="info" @click="dialog.warning({
          title: '暂不支持',
          content: '受限当前wailsv2的限制，暂不支持多窗口模式'
        })">
          打开新窗口
        </NButton>
        <NButton type="primary" @click="showCreateModal = true" :loading="isLoading">
          新建工作空间
        </NButton>
      </NFlex>
    </template>
  </NCard>
  <NModal
    v-model:show="showCreateModal"
    preset="card"
    title="新建工作空间"
    style="width: 500px"
    size="huge"
  >
    <NForm :model="{ name: newWorkspaceName, path: newWorkspacePath }">
      <NFormItem label="名称" required>
        <NInput v-model:value="newWorkspaceName" placeholder="请输入工作空间名称" maxlength="50" />
      </NFormItem>
      <NFormItem label="路径">
        <NInput v-model:value="newWorkspacePath" placeholder="留空则使用用户主目录" />
      </NFormItem>
    </NForm>
    <NAlert type="info" :bordered="false">可拖拽文件或目录到此处将自动设置为工作空间路径</NAlert>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="showCreateModal = false" :disabled="isLoading">取消</NButton>
        <NButton type="primary" @click="handleCreateWorkspace" :loading="isLoading">创建</NButton>
      </NSpace>
    </template>
  </NModal>
  <NModal
    v-model:show="showImportModal"
    preset="card"
    title="导入工作空间"
    style="width: 500px"
    size="huge"
  >
    <NForm>
      <NFormItem label="选择文件">
        <NInputGroup>
          <NInput v-model:value="importPath" readonly placeholder="请选择要导入的JSON文件" />
          <NButton @click="handleSelectImportFile">选择文件</NButton>
        </NInputGroup>
      </NFormItem>
      <NFormItem label="工作空间名称" required>
        <NInput v-model:value="importWorkspaceName" placeholder="请输入工作空间名称" maxlength="50" />
      </NFormItem>
      <NFormItem label="工作空间路径">
        <NInput v-model:value="importWorkspacePath" placeholder="留空则使用用户主目录" />
      </NFormItem>
    </NForm>
    <NAlert type="info" :bordered="false">导入将创建新的工作空间，包含命令库、分组和忽略规则</NAlert>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="showImportModal = false" :disabled="isLoading">取消</NButton>
        <NButton type="primary" @click="handleImportWorkspace" :loading="isLoading">导入</NButton>
      </NSpace>
    </template>
  </NModal>
  <DataBackupDialog v-model:show="showBackupDialog" @data-changed="handleDataChanged" />
</template>

<style scoped>
</style>
