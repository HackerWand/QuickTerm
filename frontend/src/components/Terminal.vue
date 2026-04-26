<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useTerminalStore } from '../stores/terminal'
import { useWorkspaceStore } from '../stores/workspace'
import { useCommandStore } from '../stores/command'
import * as App from '../../wailsjs/go/main/App'
import type { IDisposable } from '@xterm/xterm'

const props = defineProps<{
  workspacePath: string
  terminalId: string
}>()

const terminalStore = useTerminalStore()
const workspaceStore = useWorkspaceStore()
const commandStore = useCommandStore()
const currentInput = ref('')
const terminalRef = ref<HTMLElement>()
const dataDisposable = ref<IDisposable | undefined>()
const resizeObserver = ref<ResizeObserver | null>(null)

const processInputData = (data: string): void => {
  let i = 0
  while (i < data.length) {
    const char = data[i]
    if (char === '\x1b') {
      i++
      if (i < data.length && data[i] === '[') {
        i++
        while (i < data.length && data[i] >= '\x30' && data[i] <= '\x3f') {
          i++
        }
        while (i < data.length && data[i] >= '\x20' && data[i] <= '\x2f') {
          i++
        }
        if (i < data.length && ((data[i] >= '\x40' && data[i] <= '\x7e'))) {
          i++
        }
      }
      continue
    }

    if (char === '\n' || char === '\r') {
      const trimmedInput = currentInput.value.trimEnd()
      if (trimmedInput.endsWith('\\')) {
        currentInput.value = currentInput.value.substring(0, currentInput.value.length - 1) + ' '
      } else {
        if (currentInput.value.trim() !== '') {
          handleCommandAutoSave(currentInput.value)
        }
        currentInput.value = ''
      }
      i++
    } else if (char === '\x7f' || char === '\b') {
      if (currentInput.value.length > 0) {
        currentInput.value = currentInput.value.slice(0, -1)
      }
      i++
    } else {
      currentInput.value += char
      i++
    }
  }
}

const handleCommandAutoSave = async (command: string): Promise<void> => {
  try {
    const workspace = workspaceStore.currentWorkspace
    if (!workspace) return

    let filteredCommand = command
      .replace(/\x1b\[[0-9:;<=>?]*[ -/]*[@A-Z[\]^_`a-z{|}~]/g, '')
      .replace(/\x1b\[200~([\s\S]*?)\x1b\[201~/g, '$1')
      .replace(/\[200~([^\[]*)\[201~/g, '$1')
      .replace(/[\x00-\x09\x0b\x0c\x0e-\x1f\x7f]/g, '')
      .trim()

    if (filteredCommand === '') return

    await commandStore.autoSaveCommand(workspace.id, filteredCommand)
  } catch (error) {
    console.error('Failed to auto save command:', error)
  }
}

const handleTerminalData = (data: string): void => {
  processInputData(data)
  try {
    App.WriteTerminal(props.terminalId, data)
  } catch (error) {
    console.error('Failed to write to terminal:', error)
  }
}

const handleResize = (): void => {
  terminalStore.resizeTerminal(props.terminalId)
}

const focus = (): void => {
  terminalStore.focusTerminal(props.terminalId)
}

defineExpose({
  focus,
  terminalId: props.terminalId
})

onMounted(() => {
  nextTick(async () => {
    await terminalStore.createTerminalInstance(props.terminalId, props.workspacePath)

    if (terminalRef.value) {
      terminalStore.attachTerminalToContainer(props.terminalId, terminalRef.value)

      resizeObserver.value = new ResizeObserver(() => {
        handleResize()
      })
      resizeObserver.value.observe(terminalRef.value)
    }

    dataDisposable.value = terminalStore.onTerminalData(props.terminalId, handleTerminalData)

    handleResize()
  })
})

onUnmounted(async () => {
  if (resizeObserver.value) {
    resizeObserver.value.disconnect()
    resizeObserver.value = null
  }

  if (dataDisposable.value) {
    dataDisposable.value.dispose()
  }
})
</script>

<template>
  <div ref="terminalRef" class="terminal-wrapper"></div>
</template>

<style scoped>
.terminal-wrapper {
  width: 100%;
  height: 100%;
  overflow: hidden;
}
</style>
