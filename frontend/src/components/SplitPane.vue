<script setup lang="ts">
import { ref, computed } from 'vue'
import { NSplit, NTabs, NTabPane, NButton, NIcon, NFlex, NPopconfirm } from 'naive-ui'
import { Add } from '@vicons/ionicons5'
import { ArrowAutofitHeight20Filled, ArrowAutofitWidth20Filled } from '@vicons/fluent'
import type { SplitPaneData, SplitDirection, Workspace } from '../types'
import Terminal from './Terminal.vue'
import { useWorkspaceStore } from '../stores/workspace'
import { useTerminalStore } from '../stores/terminal'

const props = defineProps<{
  data: SplitPaneData
}>()

const emit = defineEmits<{
  split: [id: string, direction: SplitDirection]
  addTab: [id: string]
  removeTab: [paneId: string, tabIndex: number]
}>()

const workspaceStore = useWorkspaceStore()
const terminalStore = useTerminalStore()

const terminalRefs = ref<Map<string, InstanceType<typeof Terminal>>>(new Map())
const firstPaneSize = ref(0.5)

const workspace = computed(() => workspaceStore.currentWorkspace as Workspace)

const handleAddTab = () => {
  emit('addTab', props.data.id)
}

const handleRemoveTab = (index: number) => {
  emit('removeTab', props.data.id, index)
}

const handleSplit = (direction: SplitDirection) => {
  emit('split', props.data.id, direction)
}

const registerTerminal = (id: string, el: any) => {
  if (el) {
    terminalRefs.value.set(id, el as InstanceType<typeof Terminal>)
  }
}
</script>

<template>
  <NSplit 
    v-if="data.direction" 
    :direction="data.direction === 'horizontal' ? 'horizontal' : 'vertical'"
    :default-size="firstPaneSize"
    :resize-trigger-size="1"
  >
    <template #1>
      <SplitPane
        v-if="data.first"
        :data="data.first"
        @split="(id, dir) => emit('split', id, dir)"
        @addTab="(id) => emit('addTab', id)"
        @removeTab="(paneId, tabIndex) => emit('removeTab', paneId, tabIndex)"
      />
    </template>
    <template #2>
      <SplitPane
        v-if="data.second"
        :data="data.second"
        @split="(id, dir) => emit('split', id, dir)"
        @addTab="(id) => emit('addTab', id)"
        @removeTab="(paneId, tabIndex) => emit('removeTab', paneId, tabIndex)"
      />
    </template>
  </NSplit>
  <div v-else class="pane-container" :class="{
    active: data.tabs.some(v => v.id === terminalStore.activeTerminalId)
  }">
    <NTabs v-model:value="data.activeTab" type="line" :style="{
      flex: 1,
      display: 'flex',
      'flex-direction': 'column',
      overflow: 'hidden',
      height: '100%'
    }" :pane-style="{
      flex: 1,
      height: 0
    }">
      <template #suffix v-if="data.tabs.length > 0">
        <NFlex :size="2">
          <NButton size="small" @click="handleSplit('horizontal')" text>
            <template #icon>
              <NIcon><ArrowAutofitWidth20Filled /></NIcon>
            </template>
          </NButton>
          <NButton size="small" @click="handleSplit('vertical')" text>
            <template #icon>
              <NIcon><ArrowAutofitHeight20Filled /></NIcon>
            </template>
          </NButton>
          <NButton size="small" @click="handleAddTab" text>
            <template #icon>
              <NIcon><Add /></NIcon>
            </template>
          </NButton>
        </NFlex>
      </template>
      <NTabPane v-if="data.tabs.length === 0" :name="0">
        <template #tab>
          <NFlex justify="space-between" @click="handleAddTab">
            <span>新增终端</span>
            <NButton size="small" text>
              <template #icon>
                <NIcon><Add /></NIcon>
              </template>
            </NButton>
          </NFlex>
        </template>
      </NTabPane>
      <NTabPane
        v-for="(tab, index) in data.tabs"
        :key="tab.id"
        :name="index"
        display-directive="show:lazy"
      >
        <template #tab>
          <NFlex justify="space-between">
            <span>{{ tab.name }}</span>
            <NPopconfirm @positive-click="handleRemoveTab(index)">
              确定关闭该终端吗？
              <template #trigger>
                <NButton text>x</NButton>
              </template>
            </NPopconfirm>
          </NFlex>
        </template>
        <Terminal
          :key="tab.id"
          :ref="(el) => el && registerTerminal(tab.id, el)"
          :workspace-path="workspace.path"
          :terminal-id="tab.id"
        />
      </NTabPane>
    </NTabs>
  </div>
</template>

<style scoped>
.pane-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  border: 1px solid transparent;
  transition: all .3s;
}
.pane-container.active {
  border-color: rgba(255, 255, 255, 0.5);
}
</style>
