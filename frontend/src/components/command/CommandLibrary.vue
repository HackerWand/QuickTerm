<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { NButton, NIcon, NTooltip, NInput, useMessage, useLoadingBar } from 'naive-ui'
import { Add, LayersOutline, Search } from '@vicons/ionicons5'
import { useCommandStore } from '../../stores/command'
import { hasTemplateParams } from '../../composables/useTemplateParams'
import { useWorkspaceStore } from '../../stores/workspace'
import * as App from '../../../wailsjs/go/main/App'
import { models } from '../../../wailsjs/go/models'
import CommandGroup from './CommandGroup.vue'
import CommandDialog from './CommandDialog.vue'
import GroupDialog from './GroupDialog.vue'
import QuickEditDialog from './QuickEditDialog.vue'
import type { Command, CommandGroup as CommandGroupType, Workspace, CommandData } from '../../types'

interface Emits {
  (e: 'execute-command', command: Command): void
}

const emit = defineEmits<Emits>()
const message = useMessage()
const loadingBar = useLoadingBar()
const commandStore = useCommandStore()
const workspaceStore = useWorkspaceStore()

const showCommandDialog = ref(false)
const showGroupDialog = ref(false)
const showQuickEditDialog = ref(false)
const editingCommand = ref<Command | null>(null)
const editingGroup = ref<CommandGroupType | null>(null)
const quickEditingCommand = ref<Command | null>(null)
const currentAddingGroupId = ref<number | undefined>(undefined)
const searchKeyword = ref('')

const matchesSearch = (command: Command, keyword: string): boolean => {
  if (!keyword.trim()) return true
  const lowerKeyword = keyword.toLowerCase()
  if (command.name.toLowerCase().includes(lowerKeyword)) return true
  if (command.content.toLowerCase().includes(lowerKeyword)) return true
  if (command.description.toLowerCase().includes(lowerKeyword)) return true
  return command.templateParams.some((param) => {
    if (param.name.toLowerCase().includes(lowerKeyword)) return true
    if (param.description.toLowerCase().includes(lowerKeyword)) return true
    return param.options.some(
      (option) =>
        option.label.toLowerCase().includes(lowerKeyword) ||
        option.value.toLowerCase().includes(lowerKeyword)
    )
  })
}

const filteredCommands = computed(() => {
  const keyword = searchKeyword.value
  if (!keyword.trim()) return commandStore.commands
  return commandStore.commands.filter((cmd: Command) => matchesSearch(cmd, keyword))
})

const filteredGroups = computed(() => {
  const keyword = searchKeyword.value
  const commands = filteredCommands.value
  if (!keyword.trim()) return commandStore.groups
  const groupIds = commands.map(v => v.groupId)
  return commandStore.groups.filter((group: CommandGroupType) => groupIds.includes(group.id))
})

const ungroupedCommands = computed(() => {
  return filteredCommands.value.filter((cmd: Command) => !cmd.groupId)
})

const isSearching = computed(() => {
  return searchKeyword.value.trim().length > 0
})

const workspaceId = computed(() => {
  return workspaceStore.currentWorkspace?.id || 0
})

const loadData = async () => {
  if (!workspaceId.value) return
  
  loadingBar.start()
  try {
    await Promise.all([
      commandStore.loadCommands(workspaceId.value),
      commandStore.loadGroups(workspaceId.value)
    ])
    loadingBar.finish()
  } catch (error) {
    loadingBar.error()
    message.error('加载数据失败')
    console.error(error)
  }
}

const handleAddCommand = (groupId: number | undefined = undefined) => {
  currentAddingGroupId.value = groupId
  editingCommand.value = null
  showCommandDialog.value = true
}

const handleEditCommand = (command: Command) => {
  editingCommand.value = command
  currentAddingGroupId.value = undefined
  showCommandDialog.value = true
}

const handleSaveCommand = async (command: Omit<CommandData, 'id'>) => {
  try {
    const cmd = {
      ...command,
      groupId: currentAddingGroupId.value !== undefined ? currentAddingGroupId.value : command.groupId
    } as Command
    await commandStore.createCommand(cmd)
    message.success('命令创建成功')
  } catch (error) {
    message.error('创建命令失败')
    console.error(error)
  }
}

const handleUpdateCommand = async (command: CommandData) => {
  try {
    await commandStore.updateCommand(command as unknown as Command)
    message.success('命令更新成功')
  } catch (error) {
    message.error('更新命令失败')
    console.error(error)
  }
}

const handleDeleteCommand = async (command: Command) => {
  try {
    await commandStore.deleteCommand(command.id, command.workspaceId)
    message.success('命令删除成功')
  } catch (error) {
    message.error('删除命令失败')
    console.error(error)
  }
}

const handleAddGroup = () => {
  editingGroup.value = null
  showGroupDialog.value = true
}

const handleEditGroup = (group: CommandGroupType) => {
  editingGroup.value = group
  showGroupDialog.value = true
}

const handleSaveGroup = async (group: Omit<CommandGroupType, 'id'>) => {
  try {
    await commandStore.createGroup(group)
    message.success('分组创建成功')
  } catch (error) {
    message.error('创建分组失败')
    console.error(error)
  }
}

const handleUpdateGroup = async (group: CommandGroupType) => {
  try {
    await commandStore.updateGroup(group)
    message.success('分组更新成功')
  } catch (error) {
    message.error('更新分组失败')
    console.error(error)
  }
}

const handleDeleteGroup = async (group: CommandGroupType) => {
  try {
    await commandStore.deleteGroup(group.id, group.workspaceId)
    message.success('分组删除成功')
  } catch (error) {
    message.error('删除分组失败')
    console.error(error)
  }
}

const handleExecuteCommand = (command: Command) => {
  if (hasTemplateParams(command.content)) {
    message.info('命令包含模版参数，已自动打开快捷编辑')
    handleQuickEditCommand(command)
    return
  }
  emit('execute-command', command)
}

const handleQuickEditCommand = (command: Command) => {
  quickEditingCommand.value = command
  showQuickEditDialog.value = true
}

const handleQuickEditSave = async (command: CommandData) => {
  try {
    await commandStore.updateCommand(command as unknown as Command)
    message.success('命令更新成功')
  } catch (error) {
    message.error('更新命令失败')
    console.error(error)
  }
}

const handleQuickEditExecute = async (commandContent: string) => {
  const tempCommand: CommandData = {
    id: 0,
    name: '临时命令',
    content: commandContent,
    description: '',
    workspaceId: workspaceId.value,
    templateParams: []
  }
  emit('execute-command', tempCommand as unknown as Command)
}

const handleAddToIgnore = async (command: Command) => {
  try {
    const workspace = workspaceStore.currentWorkspace
    if (!workspace) {
      message.error('工作空间未加载')
      return
    }

    const newIgnoreRule = {
      pattern: command.content,
      isRegex: false
    }

    const updatedWorkspace = {
      id: workspace.id,
      name: workspace.name,
      path: workspace.path,
      ignoredCommands: [...workspace.ignoredCommands, newIgnoreRule]
    } as models.Workspace

    await App.UpdateWorkspace(updatedWorkspace)
    await commandStore.deleteCommand(command.id, command.workspaceId)
    workspaceStore.setWorkspace(updatedWorkspace)
    message.success('已添加到忽略列表')
  } catch (error) {
    message.error('添加到忽略列表失败')
    console.error(error)
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="command-library">
    <div class="library-header">
      <h3 class="section-title">命令库</h3>
      <div class="header-actions">
        <NTooltip>
          <template #trigger>
            <NButton
              quaternary
              size="tiny"
              @click="handleAddGroup"
            >
              <template #icon>
                <NIcon><LayersOutline /></NIcon>
              </template>
            </NButton>
          </template>
          添加分组
        </NTooltip>
        <NTooltip>
          <template #trigger>
            <NButton
              quaternary
              size="tiny"
              @click="handleAddCommand()"
            >
              <template #icon>
                <NIcon><Add /></NIcon>
              </template>
            </NButton>
          </template>
          添加命令
        </NTooltip>
      </div>
    </div>

    <!-- 在这里新增一个搜索框，用于实现命令的搜索功能 -->
    <NInput
      v-model:value="searchKeyword"
      placeholder="搜索命令..."
      clearable
      size="small"
      style="margin-bottom: 12px; flex-shrink: 0;"
    >
      <template #prefix>
        <NIcon :size="14"><Search /></NIcon>
      </template>
    </NInput>
    
    <div class="library-content">
      <CommandGroup
        v-for="group in filteredGroups"
        :key="group.id"
        :group="group"
        :commands="filteredCommands.filter((cmd: Command) => cmd.groupId === group.id)"
        @add-command="handleAddCommand(group.id)"
        @edit-command="handleEditCommand"
        @delete-command="handleDeleteCommand"
        @edit-group="handleEditGroup(group)"
        @delete-group="handleDeleteGroup(group)"
        @execute-command="handleExecuteCommand"
        @copy-command="() => {}"
        @quick-edit-command="handleQuickEditCommand"
        @add-to-ignore="handleAddToIgnore"
      />
      
      <CommandGroup
        v-if="ungroupedCommands.length > 0 || (!isSearching && commandStore.groups.length === 0)"
        :group="undefined"
        :commands="ungroupedCommands"
        @add-command="handleAddCommand(undefined)"
        @edit-command="handleEditCommand"
        @delete-command="handleDeleteCommand"
        @execute-command="handleExecuteCommand"
        @copy-command="() => {}"
        @quick-edit-command="handleQuickEditCommand"
        @add-to-ignore="handleAddToIgnore"
      />

      <div v-if="isSearching && filteredCommands.length === 0" class="no-results">
        未找到匹配的命令
      </div>
    </div>
    
    <CommandDialog
      v-model:show="showCommandDialog"
      :command="editingCommand"
      :groups="commandStore.groups"
      :workspace-id="workspaceId"
      @save="handleSaveCommand"
      @update="handleUpdateCommand"
    />
    
    <GroupDialog
      v-model:show="showGroupDialog"
      :group="editingGroup"
      :workspace-id="workspaceId"
      @save="handleSaveGroup"
      @update="handleUpdateGroup"
    />
    
    <QuickEditDialog
      v-model:show="showQuickEditDialog"
      :command="quickEditingCommand"
      @save="handleQuickEditSave"
      @execute="handleQuickEditExecute"
    />
  </div>
</template>

<style scoped>
.command-library {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.library-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-shrink: 0;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--n-color-text-3);
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 4px;
}

.library-content {
  flex: 1;
  overflow-y: auto;
  padding-right: 4px;
}

.library-content::-webkit-scrollbar {
  width: 6px;
}

.library-content::-webkit-scrollbar-track {
  background: transparent;
}

.library-content::-webkit-scrollbar-thumb {
  background: var(--n-color-border);
  border-radius: 3px;
}

.no-results {
  text-align: center;
  color: var(--n-color-text-3);
  font-size: 13px;
  padding: 24px 0;
}
</style>
