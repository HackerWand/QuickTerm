import { defineStore } from 'pinia'
import { nextTick, ref } from 'vue'
import { Terminal as XTerminal, IDisposable } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebglAddon } from '@xterm/addon-webgl'
import * as runtime from '../../wailsjs/runtime/runtime'
import * as App from '../../wailsjs/go/main/App'
import type { TerminalInstance, TerminalConfig } from '../types'
import { ThrottleMap } from '../utils/throttle'

const TERMINAL_CONFIG: TerminalConfig = {
  cursorBlink: true,
  cursorStyle: 'bar',
  theme: {
    background: '#0d1117',
    foreground: '#c9d1d9',
    cursor: '#c9d1d9',
    black: '#0d1117',
    red: '#f85149',
    green: '#7ee787',
    yellow: '#f2cc60',
    blue: '#79c0ff',
    magenta: '#d2a8ff',
    cyan: '#7ee787',
    white: '#f6f8fa',
    brightBlack: '#484f58',
    brightRed: '#f85149',
    brightGreen: '#7ee787',
    brightYellow: '#f2cc60',
    brightBlue: '#79c0ff',
    brightMagenta: '#d2a8ff',
    brightCyan: '#7ee787',
    brightWhite: '#f6f8fa'
  },
  fontSize: 14,
  fontFamily: 'Menlo, Monaco, "Courier New", monospace',
  allowTransparency: true,
  convertEol: true,
  macOptionIsMeta: false,
  macOptionClickForcesSelection: false,
  scrollback: 1000
}

export interface TerminalInstanceWithAddon extends TerminalInstance {
  fitAddon: FitAddon
}

export const useTerminalStore = defineStore('terminal', () => {
  const terminalMap = ref<Map<string, TerminalInstanceWithAddon>>(new Map())
  const activeTerminalId = ref<string | null>(null)
  const resizeThrottler = new ThrottleMap(100)

  const getTerminalInstance = (id: string): TerminalInstanceWithAddon | undefined => {
    return terminalMap.value.get(id)
  }

  const createTerminalInstance = async (id: string, workspacePath: string): Promise<TerminalInstanceWithAddon> => {
    const existing = terminalMap.value.get(id)
    if (existing) {
      return existing
    }

    const container = document.createElement('div')
    container.className = 'terminal-container'
    container.style.width = '100%'
    container.style.height = '100%'
    container.style.overflow = 'hidden'

    const term = new XTerminal(TERMINAL_CONFIG)
    const fitAddon = new FitAddon()
    term.loadAddon(fitAddon)
    term.loadAddon(new WebglAddon())
    term.open(container)
    fitAddon.fit()

    term.textarea?.addEventListener('focus', () => {
      activeTerminalId.value = id
    })

    const instance: TerminalInstanceWithAddon = {
      id,
      terminal: term,
      container,
      workspacePath,
      fitAddon
    }

    terminalMap.value.set(id, instance)

    try {
      await App.CreateTerminal(id, '', workspacePath, [])
    } catch (error) {
      console.error('Failed to create terminal:', error)
      term.writeln('Error: Failed to create terminal')
      return instance
    }

    const outputEventName = `terminal-output-${id}`
    instance.outputListener = runtime.EventsOn(outputEventName, (data: string) => {
      const terminalInst = terminalMap.value.get(id)
      if (terminalInst) {
        terminalInst.terminal.write(data)
      }
    })

    const closeEventName = `terminal-closed-${id}`
    instance.closeListener = runtime.EventsOn(closeEventName, () => {
      const terminalInst = terminalMap.value.get(id)
      if (terminalInst) {
        terminalInst.terminal.writeln('\nTerminal closed')
      }
    })

    return instance
  }

  const destroyTerminalInstance = async (id: string): Promise<void> => {
    const instance = terminalMap.value.get(id)
    if (!instance) {
      return
    }

    if (activeTerminalId.value === id) {
      activeTerminalId.value = null
    }

    resizeThrottler.cancel(id)

    if (instance.outputListener) {
      instance.outputListener()
    }
    if (instance.closeListener) {
      instance.closeListener()
    }

    try {
      await App.CloseTerminal(id)
    } catch (error) {
      console.error('Failed to close terminal:', error)
    }

    instance.terminal.dispose()
    terminalMap.value.delete(id)
  }

  const destroyAllTerminalInstances = async (): Promise<void> => {
    resizeThrottler.cancelAll()

    const terminalIds = Array.from(terminalMap.value.keys())
    for (const id of terminalIds) {
      destroyTerminalInstance(id)
    }
    
    // 清理所有后端终端
    try {
      await App.ClearAllTerminals()
    } catch (error) {
      console.error('Failed to clear all terminals:', error)
    }
  }

  const attachTerminalToContainer = (terminalId: string, targetContainer: HTMLElement): void => {
    const instance = terminalMap.value.get(terminalId)
    if (!instance) {
      return
    }

    if (!targetContainer.contains(instance.container)) {
      targetContainer.innerHTML = ''
      targetContainer.appendChild(instance.container)
    }
    nextTick(() => {
      resizeTerminal(terminalId)
    })
  }

  const focusTerminal = (id: string): void => {
    const instance = terminalMap.value.get(id)
    if (instance) {
      activeTerminalId.value = id
      instance.terminal.focus()
    }
  }

  const resizeTerminal = (id: string): void => {
    const instance = terminalMap.value.get(id)
    if (!instance) {
      return
    }

    instance.fitAddon.fit()

    resizeThrottler.call(id, () => {
      const currentInstance = terminalMap.value.get(id)
      if (!currentInstance) {
        return
      }
      try {
        App.ResizeTerminal(id, currentInstance.terminal.rows, currentInstance.terminal.cols)
      } catch (error) {
        console.error('Failed to resize terminal:', error)
      }
    })
  }

  const writeToTerminal = (id: string, data: string): void => {
    const instance = terminalMap.value.get(id)
    if (instance) {
      try {
        App.WriteTerminal(id, data)
      } catch (error) {
        console.error('Failed to write to terminal:', error)
        instance.terminal.writeln('\nError: Failed to send input')
      }
    }
  }

  const onTerminalData = (id: string, callback: (data: string) => void): IDisposable | undefined => {
    const instance = terminalMap.value.get(id)
    if (!instance) {
      return
    }

    return instance.terminal.onData(callback)
  }

  return {
    terminalMap,
    activeTerminalId,
    getTerminalInstance,
    createTerminalInstance,
    destroyTerminalInstance,
    destroyAllTerminalInstances,
    attachTerminalToContainer,
    focusTerminal,
    resizeTerminal,
    writeToTerminal,
    onTerminalData
  }
})
