import mitt from 'mitt'
import { EventsOn, EventsOff } from '../../wailsjs/runtime'

export interface FileDropEvent {
  x: number
  y: number
  paths: string[]
}

type FileDropEvents = {
  'file-drop': FileDropEvent
}

const emitter = mitt<FileDropEvents>()
let isWailsListenerSet = false

const handleWailsFileDrop = (x: number, y: number, paths: string[]) => {
  console.log('handleWailsFileDrop', x, y, paths)
  emitter.emit('file-drop', { x, y, paths })
}

export const setupFileDropListener = () => {
  if (!isWailsListenerSet) {
    EventsOn('wails:file-drop', handleWailsFileDrop)
    isWailsListenerSet = true
  }
}

export const removeFileDropListener = () => {
  if (isWailsListenerSet) {
    EventsOff('wails:file-drop')
    isWailsListenerSet = false
  }
}

export const onFileDrop = (handler: (event: FileDropEvent) => void) => {
  emitter.on('file-drop', handler)
  return () => emitter.off('file-drop', handler)
}

export const offFileDrop = (handler: (event: FileDropEvent) => void) => {
  emitter.off('file-drop', handler)
}
