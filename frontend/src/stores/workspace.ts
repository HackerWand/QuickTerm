import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Workspace } from '../types'
import * as App from '../../wailsjs/go/main/App'
import { models } from '../../wailsjs/go/models'

export const useWorkspaceStore = defineStore('workspace', () => {
  const currentWorkspace = ref<Workspace | null>(null)

  const setWorkspace = (workspace: Workspace) => {
    currentWorkspace.value = workspace
  }

  const clearWorkspace = () => {
    currentWorkspace.value = null
  }

  const updateWorkspace = async (updates: Partial<Workspace>) => {
    if (currentWorkspace.value) {
      const workspaceToSave = new models.Workspace({
        id: currentWorkspace.value.id,
        name: updates.name ?? currentWorkspace.value.name,
        path: updates.path ?? currentWorkspace.value.path,
        ignoredCommands: updates.ignoredCommands ?? currentWorkspace.value.ignoredCommands
      })
      await App.UpdateWorkspace(workspaceToSave)
      currentWorkspace.value = { ...currentWorkspace.value, ...updates }
    }
  }

  return {
    currentWorkspace,
    setWorkspace,
    clearWorkspace,
    updateWorkspace
  }
})
