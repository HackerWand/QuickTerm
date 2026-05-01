<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { NButton, NIcon, NPopconfirm, useMessage, useDialog, useLoadingBar } from 'naive-ui'
import { Trash, Add, OpenOutline } from '@vicons/ionicons5'
import { useRecentPathStore } from '../stores/recentPath'
import { useWorkspaceStore } from '../stores/workspace'
import { useTerminalStore } from '../stores/terminal'
import * as App from '../../wailsjs/go/main/App'
import type { RecentPath } from '../types'

interface Emits {
  (e: 'copy-path', path: string): void
}

const emit = defineEmits<Emits>()
const message = useMessage()
const dialog = useDialog()
const loadingBar = useLoadingBar()
const recentPathStore = useRecentPathStore()
const workspaceStore = useWorkspaceStore()
const terminalStore = useTerminalStore()

const workspaceId = computed(() => {
  return workspaceStore.currentWorkspace?.id || 0
})

const safePaths = computed(() => {
  return recentPathStore.safePaths()
})

const loadData = async () => {
  if (!workspaceId.value) return
  
  loadingBar.start()
  try {
    await recentPathStore.loadPaths(workspaceId.value)
    loadingBar.finish()
  } catch (error) {
    loadingBar.error()
    message.error('加载最近路径失败')
    console.error(error)
  }
}

const handleCopyPath = async (path: string) => {
  try {
    await navigator.clipboard.writeText(path)
    message.success('路径已复制到剪贴板')
  } catch (error) {
    message.error('复制失败')
  }
  emit('copy-path', path)
}

const handleNavigateToPath = async (path: string) => {
  const activeId = terminalStore.activeTerminalId
  if (!activeId) {
    message.warning('没有激活的终端，请先点击一个终端')
    return
  }
  try {
    const dirPath = await App.GetDirectoryPath(path)
    terminalStore.focusTerminal(activeId)
    await new Promise(resolve => setTimeout(resolve, 100))
    terminalStore.writeToTerminal(activeId, `cd "${dirPath}"\n`)
    message.success(`已导航到: ${dirPath}`)
  } catch (error) {
    dialog.error({
      title: '导航到路径失败',
      content: '请检查路径是否存在或是否可访问:' + (error as Error).message
    })
    console.error(error)
  }
}

const handleDeletePath = async (path: RecentPath) => {
  try {
    await recentPathStore.deletePath(path.id, workspaceId.value)
    message.success('路径已删除')
  } catch (error) {
    message.error('删除失败')
    console.error(error)
  }
}

const handleClearPaths = async () => {
  try {
    await recentPathStore.clearPaths(workspaceId.value)
    message.success('已清空所有最近路径')
  } catch (error) {
    message.error('清空失败')
    console.error(error)
  }
}

const handleAddPath = async () => {
  try {
    const selectedPath = await App.SelectPathDialog()
    if (selectedPath && selectedPath.trim() !== '') {
      await recentPathStore.addPath(workspaceId.value, selectedPath)
      message.success('路径已添加')
      await loadData()
    }
  } catch (error) {
    message.error('添加路径失败')
    console.error(error)
  }
}

const truncatePath = (path: string, maxLength: number = 30) => {
  if (path.length <= maxLength) {
    return path
  }
  const start = Math.floor((maxLength - 3) / 2)
  const end = Math.ceil((maxLength - 3) / 2)
  return path.slice(0, start) + '...' + path.slice(-end)
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="recent-paths">
    <div class="paths-header">
      <h3 class="section-title">最近路径</h3>
      <div class="header-actions">
        <NButton
          quaternary
          size="tiny"
          @click="handleAddPath"
        >
          <template #icon>
            <NIcon><Add /></NIcon>
          </template>
        </NButton>
        <NPopconfirm
          v-if="safePaths.length > 0"
          positive-text="确定"
          negative-text="取消"
          @positive-click="handleClearPaths"
        >
          确定要清空所有最近路径吗？
          <template #trigger>
            <NButton quaternary size="tiny">清空</NButton>
          </template>
        </NPopconfirm>
      </div>
    </div>
    
    <div class="paths-content">
      <div v-if="safePaths.length === 0" class="empty-state">
        暂无最近路径
      </div>
      <div
        v-else
        class="paths-list"
      >
        <div
          v-for="path in safePaths"
          :key="path.id"
          class="path-item"
          :title="path.path"
          @click="handleCopyPath(path.path)"
        >
          <span class="path-text">{{ truncatePath(path.path) }}</span>
          <NButton
            quaternary
            size="tiny"
            circle
            class="navigate-button"
            @click.stop="handleNavigateToPath(path.path)"
          >
            <template #icon>
              <NIcon><OpenOutline /></NIcon>
            </template>
          </NButton>
          <NPopconfirm
            positive-text="确定"
            negative-text="取消"
            @positive-click.stop="handleDeletePath(path)"
          >
            确定要删除该路径吗？
            <template #trigger>
              <NButton
                quaternary
                size="tiny"
                circle
                class="delete-button"
              >
                <template #icon>
                  <NIcon><Trash /></NIcon>
                </template>
              </NButton>
            </template>
          </NPopconfirm>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.recent-paths {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.paths-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  flex-shrink: 0;
}

.header-actions {
  display: flex;
  gap: 4px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--n-color-text-3);
  margin: 0;
}

.paths-content {
  flex: 1;
  overflow-y: auto;
  padding-right: 4px;
}

.paths-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 150px;
  overflow-y: auto;
}

.path-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background-color: var(--n-color-modal);
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.path-item:hover {
  background-color: var(--n-color-popover);
}

.path-text {
  font-size: 12px;
  color: var(--n-color-text-2);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
  min-width: 0;
  margin-right: 8px;
}

.delete-button,
.navigate-button {
  opacity: 0;
  transition: opacity 0.2s;
}

.path-item:hover .delete-button,
.path-item:hover .navigate-button {
  opacity: 1;
}

.empty-state {
  font-size: 12px;
  color: var(--n-color-text-3);
  text-align: center;
  padding: 16px;
}

.paths-list::-webkit-scrollbar {
  width: 6px;
}

.paths-list::-webkit-scrollbar-track {
  background: transparent;
}

.paths-list::-webkit-scrollbar-thumb {
  background: var(--n-color-border);
  border-radius: 3px;
}
</style>
