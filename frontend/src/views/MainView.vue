<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  NLayout,
  NLayoutSider,
  NLayoutContent,
  useMessage,
  NFlex
} from 'naive-ui'
import { ArrowBack, SettingsOutline } from '@vicons/ionicons5'
import type { Workspace, Command, SplitPaneData, SplitDirection, TerminalTab } from '../types'
import { useWorkspaceStore } from '../stores/workspace'
import { useRecentPathStore } from '../stores/recentPath'
import { useTerminalStore } from '../stores/terminal'
import SplitPane from '../components/SplitPane.vue'
import CommandLibrary from '../components/command/CommandLibrary.vue'
import RecentPaths from '../components/RecentPaths.vue'
import WorkspaceSettings from '../components/WorkspaceSettings.vue'
import { onFileDrop, setupFileDropListener } from '../utils/fileDrop'

const router = useRouter()
const message = useMessage()
const workspaceStore = useWorkspaceStore()
const recentPathStore = useRecentPathStore()
const terminalStore = useTerminalStore()

const workspace = computed(() => workspaceStore.currentWorkspace as Workspace)

const showSettingsModal = ref(false)

const createSplitPaneData = (): SplitPaneData => {
  return {
    id: `pane-${Math.random().toString(36).substring(2, 15)}`,
    direction: null,
    tabs: [
      { id: `terminal-${Math.random().toString(36).substring(2, 15)}`, name: '终端 1' }
    ],
    activeTab: 0
  }
}

const splitRoot = ref<SplitPaneData>(createSplitPaneData())

const findPaneById = (root: SplitPaneData, id: string): SplitPaneData | null => {
  if (root.id === id) {
    return root
  }
  if (root.first) {
    const found = findPaneById(root.first, id)
    if (found) return found
  }
  if (root.second) {
    const found = findPaneById(root.second, id)
    if (found) return found
  }
  return null
}

const findParentPane = (root: SplitPaneData, id: string): SplitPaneData | null => {
  if (root.first && root.first.id === id) return root
  if (root.second && root.second.id === id) return root
  if (root.first) {
    const found = findParentPane(root.first, id)
    if (found) return found
  }
  if (root.second) {
    const found = findParentPane(root.second, id)
    if (found) return found
  }
  return null
}

const deepClonePane = (pane: SplitPaneData): SplitPaneData => {
  const cloned: SplitPaneData = {
    id: pane.id,
    direction: pane.direction,
    tabs: [...pane.tabs],
    activeTab: pane.activeTab
  }
  if (pane.first) {
    cloned.first = deepClonePane(pane.first)
  }
  if (pane.second) {
    cloned.second = deepClonePane(pane.second)
  }
  return cloned
}

const handleSplit = (id: string, direction: SplitDirection) => {
  const pane = findPaneById(splitRoot.value, id)
  if (!pane || pane.direction) return

  pane.direction = direction
  pane.first = {
    id: `pane-${Math.random().toString(36).substring(2, 15)}`,
    direction: null,
    tabs: [...pane.tabs],
    activeTab: pane.activeTab
  }
  pane.second = {
    id: `pane-${Math.random().toString(36).substring(2, 15)}`,
    direction: null,
    tabs: [
      { id: `terminal-${Math.random().toString(36).substring(2, 15)}`, name: '终端 1' }
    ],
    activeTab: 0
  }
}

const handleAddTab = (id: string) => {
  const pane = findPaneById(splitRoot.value, id)
  if (!pane) return

  const newId = `terminal-${Math.random().toString(36).substring(2, 15)}`
  pane.tabs.push({ id: newId, name: `终端 ${pane.tabs.length + 1}` })
  pane.activeTab = pane.tabs.length - 1
}

const handleRemoveTab = async (paneId: string, tabIndex: number) => {
  const pane = findPaneById(splitRoot.value, paneId)
  if (!pane) return

  const tabId = pane.tabs[tabIndex].id

  terminalStore.destroyTerminalInstance(tabId)
  pane.tabs.splice(tabIndex, 1)

  if (pane.tabs.length !== 0) {
    if (pane.activeTab >= pane.tabs.length) {
      pane.activeTab = pane.tabs.length - 1
    } else if (pane.activeTab > tabIndex) {
      pane.activeTab = pane.activeTab - 1
    }
  } else {
    pane.activeTab = 0
    const parent = findParentPane(splitRoot.value, paneId)
    if (parent) {
      const sibling = parent.first?.id === paneId ? parent.second : parent.first
      if (sibling) {
        parent.id = sibling.id
        parent.direction = sibling.direction
        parent.tabs = [...sibling.tabs]
        parent.activeTab = sibling.activeTab
        parent.first = sibling.first
        parent.second = sibling.second
      }
    }
  }
}

const handleBackToSelector = async () => {
  workspaceStore.clearWorkspace()
  await router.push({ name: 'WorkspaceSelector' })
}

const handleExecuteCommand = async (command: Command) => {
  try {
    const activeId = terminalStore.activeTerminalId
    if (activeId) {
      terminalStore.focusTerminal(activeId)
      await new Promise(resolve => setTimeout(resolve, 100))
      terminalStore.writeToTerminal(activeId, command.content + '\n')
    } else {
      message.warning('No active terminal, please click a terminal first')
    }
  } catch (error) {
    message.error('执行命令失败')
    console.error(error)
  }
}

const handleDragDrop = async (event: { x: number; y: number; paths: string[] }) => {
  try {
    const currentWorkspaceId = workspace.value.id
    for (const path of event.paths) {
      if (recentPathStore.paths.some(p => p.path === path)) continue
      await recentPathStore.addPath(currentWorkspaceId, path)
      await navigator.clipboard.writeText(path)
      message.success(`已添加路径: ${path}`)
    }
  } catch (error) {
    message.error('添加路径失败')
    console.error(error)
  }
}

let unsubscribe: (() => void) | null = null

onMounted(() => {
  setupFileDropListener()
  unsubscribe = onFileDrop(handleDragDrop)
})

onUnmounted(() => {
  if (unsubscribe) {
    unsubscribe()
  }
})
</script>

<template>
  <NLayout has-sider style="height: 100%;">
    <NLayoutSider show-trigger="arrow-circle" bordered collapse-mode="transform" :width="280" :collapsed-width="64">
      <div class="sider-content">
        <div class="back-button-container">
          <NFlex justify="space-between" align="center">
            <NButton quaternary @click="handleBackToSelector">
              <NIcon><ArrowBack /></NIcon>
              返回工作空间
            </NButton>
            <NButton 
              quaternary 
              @click="showSettingsModal = true"
              class="settings-button"
            >
              <NIcon><SettingsOutline /></NIcon>
            </NButton>
          </NFlex>
        </div>
        <div class="sidebar-section command-section">
          <CommandLibrary @execute-command="handleExecuteCommand" />
        </div>
        <div class="sidebar-section">
          <RecentPaths />
        </div>
      </div>
    </NLayoutSider>
    
    <NLayoutContent v-if="workspace" style="height: 100%; overflow: hidden; padding: 5px;">
      <SplitPane
        :data="splitRoot"
        @split="handleSplit"
        @addTab="handleAddTab"
        @removeTab="handleRemoveTab"
      />
    </NLayoutContent>
  </NLayout>
  <WorkspaceSettings v-model:show="showSettingsModal"/>
</template>

<style scoped>
.sider-content {
  padding: 16px;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.back-button-container {
  flex-shrink: 0;
}

.settings-button {
  flex-shrink: 0;
}

.sidebar-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sidebar-section.command-section {
  flex: 1;
  overflow: hidden;
  min-height: 0;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--n-color-text-3);
  margin: 0;
}

.section-content {
  flex: 1;
  min-height: 100px;
  color: var(--n-color-text-3);
}
</style>
