<script setup lang="ts">
import { ref } from 'vue'
import {
  NModal,
  NButton,
  NSpace,
  NIcon,
  NAlert,
  NDivider,
  NText,
  NFlex,
  NStatistic,
  useMessage,
  useDialog,
  NInputGroup,
  NInput
} from 'naive-ui'
import { DownloadOutline, CloudUploadOutline } from '@vicons/ionicons5'
import { models } from '../../wailsjs/go/models'
import * as App from '../../wailsjs/go/main/App'

interface Props {
  show: boolean
}

interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'dataChanged'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
const message = useMessage()
const dialog = useDialog()

const localShow = ref(false)
const isExporting = ref(false)
const isImporting = ref(false)
const importFilePath = ref('')
const importPreview = ref<DatabaseBackupPreview | null>(null)

interface DatabaseBackupPreview {
  workspaceCount: number
  groupCount: number
  commandCount: number
  recentPathCount: number
}

const handleShowChange = (value: boolean) => {
  localShow.value = value
  emit('update:show', value)
  if (!value) {
    resetImportState()
  }
}

const resetImportState = () => {
  importFilePath.value = ''
  importPreview.value = null
}

const handleExport = async () => {
  try {
    isExporting.value = true
    const backupData = await App.ExportDatabase()
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19)
    const fileName = `QuickTerm_backup_${timestamp}.json`
    const savePath = await App.SaveDatabaseBackupDialog(fileName)
    if (savePath) {
      await App.WriteFile(savePath, JSON.stringify(backupData, null, 2))
      message.success('数据库备份成功')
    }
  } catch (error) {
    message.error('数据库备份失败')
    console.error(error)
  } finally {
    isExporting.value = false
  }
}

const handleSelectBackupFile = async () => {
  try {
    const path = await App.OpenDatabaseBackupDialog()
    if (path) {
      importFilePath.value = path
      const content = await App.ReadFile(path)
      const backupData = models.DatabaseBackup.createFrom(content)
      importPreview.value = {
        workspaceCount: backupData.workspaces.length,
        groupCount: backupData.groups.length,
        commandCount: backupData.commands.length,
        recentPathCount: backupData.recentPaths.length
      }
    }
  } catch (error) {
    message.error('读取备份文件失败，请检查文件格式')
    console.error(error)
    importFilePath.value = ''
    importPreview.value = null
  }
}

const handleImport = async () => {
  if (!importFilePath.value) {
    message.warning('请先选择备份文件')
    return
  }

  try {
    const hasData = await App.DatabaseHasData()
    if (hasData) {
      dialog.warning({
        title: '数据覆盖确认',
        content: '当前数据库中已存在数据，恢复备份将覆盖所有现有数据（包括工作空间、命令库、分组和最近路径）。此操作不可撤销，是否继续？',
        positiveText: '确认覆盖',
        negativeText: '取消',
        onPositiveClick: async () => {
          await executeImport()
        }
      })
    } else {
      await executeImport()
    }
  } catch (error) {
    message.error('检查数据库状态失败')
    console.error(error)
  }
}

const executeImport = async () => {
  try {
    isImporting.value = true
    const content = await App.ReadFile(importFilePath.value)
    const backupData = models.DatabaseBackup.createFrom(content)
    await App.ImportDatabase(backupData)
    message.success('数据库恢复成功')
    localShow.value = false
    emit('dataChanged')
    resetImportState()
  } catch (error) {
    message.error('数据库恢复失败')
    console.error(error)
  } finally {
    isImporting.value = false
  }
}
</script>

<template>
  <NModal :show="show" preset="card" title="数据备份与恢复" style="width: 550px;" @update:show="handleShowChange">
    <NFlex vertical :size="20">
      <NFlex vertical :size="12">
        <NText strong>备份数据库</NText>
        <NText depth="3">将所有工作空间、命令库、分组和最近路径导出为JSON文件</NText>
        <NButton type="info" :loading="isExporting" @click="handleExport">
          <template #icon>
            <NIcon><DownloadOutline /></NIcon>
          </template>
          导出备份
        </NButton>
      </NFlex>

      <NDivider />

      <NFlex vertical :size="12">
        <NText strong>恢复数据库</NText>
        <NText depth="3">从备份文件恢复所有数据，已存在数据将被覆盖</NText>
        <NInputGroup>
          <NInput v-model:value="importFilePath" readonly placeholder="请选择备份文件" />
          <NButton @click="handleSelectBackupFile">选择文件</NButton>
        </NInputGroup>

        <template v-if="importPreview">
          <NAlert type="info" :bordered="false">
            <NFlex justify="space-around">
              <NStatistic label="工作空间" :value="importPreview.workspaceCount" />
              <NStatistic label="命令分组" :value="importPreview.groupCount" />
              <NStatistic label="命令" :value="importPreview.commandCount" />
              <NStatistic label="最近路径" :value="importPreview.recentPathCount" />
            </NFlex>
          </NAlert>
        </template>

        <NButton type="warning" :loading="isImporting" :disabled="!importFilePath" @click="handleImport">
          <template #icon>
            <NIcon><CloudUploadOutline /></NIcon>
          </template>
          恢复备份
        </NButton>
      </NFlex>
    </NFlex>

    <template #footer>
      <NSpace justify="end">
        <NButton @click="handleShowChange(false)">关闭</NButton>
      </NSpace>
    </template>
  </NModal>
</template>
