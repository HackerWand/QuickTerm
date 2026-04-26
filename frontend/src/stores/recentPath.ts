import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { RecentPath } from '../types'
import * as App from '../../wailsjs/go/main/App'

export const useRecentPathStore = defineStore('recentPath', () => {
  const paths = ref<RecentPath[]>([])
  const isLoading = ref(false)

  const safePaths = () => paths.value || []

  const loadPaths = async (workspaceId: number) => {
    try {
      isLoading.value = true
      const result = await App.GetRecentPaths(workspaceId)
      paths.value = result || []
    } catch (error) {
      console.error('Failed to load recent paths:', error)
      paths.value = []
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const addPath = async (workspaceId: number, path: string) => {
    try {
      await App.AddRecentPath(workspaceId, path)
      await loadPaths(workspaceId)
    } catch (error) {
      console.error('Failed to add recent path:', error)
      throw error
    }
  }

  const deletePath = async (id: number, workspaceId: number) => {
    try {
      await App.DeleteRecentPath(id)
      await loadPaths(workspaceId)
    } catch (error) {
      console.error('Failed to delete recent path:', error)
      throw error
    }
  }

  const clearPaths = async (workspaceId: number) => {
    try {
      await App.ClearRecentPaths(workspaceId)
      await loadPaths(workspaceId)
    } catch (error) {
      console.error('Failed to clear recent paths:', error)
      throw error
    }
  }

  return {
    paths,
    isLoading,
    safePaths,
    loadPaths,
    addPath,
    deletePath,
    clearPaths
  }
})
