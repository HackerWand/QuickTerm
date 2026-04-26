import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Command, CommandGroup } from '../types'
import * as App from '../../wailsjs/go/main/App'

export const useCommandStore = defineStore('command', () => {
  const commands = ref<Command[]>([])
  const groups = ref<CommandGroup[]>([])
  const isLoading = ref(false)

  const loadCommands = async (workspaceId: number) => {
    try {
      isLoading.value = true
      const result = await App.GetCommands(workspaceId)
      commands.value = result || []
    } catch (error) {
      console.error('Failed to load commands:', error)
      commands.value = []
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const loadGroups = async (workspaceId: number) => {
    try {
      isLoading.value = true
      const result = await App.GetCommandGroups(workspaceId)
      groups.value = result || []
    } catch (error) {
      console.error('Failed to load command groups:', error)
      groups.value = []
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const createCommand = async (command: Omit<Command, 'id'>) => {
    try {
      await App.CreateCommand(command as Command)
      await loadCommands(command.workspaceId)
    } catch (error) {
      console.error('Failed to create command:', error)
      throw error
    }
  }

  const updateCommand = async (command: Command) => {
    try {
      await App.UpdateCommand(command)
      await loadCommands(command.workspaceId)
    } catch (error) {
      console.error('Failed to update command:', error)
      throw error
    }
  }

  const deleteCommand = async (id: number, workspaceId: number) => {
    try {
      await App.DeleteCommand(id)
      await loadCommands(workspaceId)
    } catch (error) {
      console.error('Failed to delete command:', error)
      throw error
    }
  }

  const createGroup = async (group: Omit<CommandGroup, 'id'>) => {
    try {
      await App.CreateCommandGroup(group as CommandGroup)
      await loadGroups(group.workspaceId)
    } catch (error) {
      console.error('Failed to create command group:', error)
      throw error
    }
  }

  const updateGroup = async (group: CommandGroup) => {
    try {
      await App.UpdateCommandGroup(group)
      await loadGroups(group.workspaceId)
    } catch (error) {
      console.error('Failed to update command group:', error)
      throw error
    }
  }

  const deleteGroup = async (id: number, workspaceId: number) => {
    try {
      await App.DeleteCommandGroup(id)
      await loadGroups(workspaceId)
      await loadCommands(workspaceId)
    } catch (error) {
      console.error('Failed to delete command group:', error)
      throw error
    }
  }

  const getCommandsByGroup = (groupId: number | undefined) => {
    return commands.value.filter((cmd: Command) => cmd.groupId === groupId)
  }

  const getGroupById = (groupId: number | undefined) => {
    if (!groupId) return null
    return groups.value.find((group: CommandGroup) => group.id === groupId) || null
  }

  const autoSaveCommand = async (workspaceId: number, command: string) => {
    try {
      await App.AutoSaveCommand(workspaceId, command)
      // 自动保存后刷新命令列表，确保新保存的命令显示在命令库中
      await loadCommands(workspaceId)
    } catch (error) {
      console.error('Failed to auto save command:', error)
      throw error
    }
  }

  return {
    commands,
    groups,
    isLoading,
    loadCommands,
    loadGroups,
    createCommand,
    updateCommand,
    deleteCommand,
    createGroup,
    updateGroup,
    deleteGroup,
    getCommandsByGroup,
    getGroupById,
    autoSaveCommand
  }
})
