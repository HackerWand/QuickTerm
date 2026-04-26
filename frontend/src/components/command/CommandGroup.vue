<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  NButton,
  NIcon,
  NPopconfirm,
  NTooltip,
  useMessage,
  NFlex,
  NThing,
  NEllipsis,
  NDivider
} from 'naive-ui'
import { Create, Trash, Play, Copy, EyeOff, CodeWorkingOutline, ChevronForward } from '@vicons/ionicons5'
import type { Command, CommandGroup } from '../../types'

interface Props {
  group: CommandGroup | undefined
  commands: Command[]
}

interface Emits {
  (e: 'add-command'): void
  (e: 'edit-command', command: Command): void
  (e: 'delete-command', command: Command): void
  (e: 'edit-group'): void
  (e: 'delete-group'): void
  (e: 'execute-command', command: Command): void
  (e: 'copy-command', command: Command): void
  (e: 'quick-edit-command', command: Command): void
  (e: 'add-to-ignore', command: Command): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
const message = useMessage()
const collapsed = ref(props.group ? true : false)

const groupName = computed(() => {
  return props.group ? props.group.name : '未分组'
})

const toggleCollapse = () => {
  collapsed.value = !collapsed.value
}

const handleExecute = (command: Command) => {
  emit('execute-command', command)
}

const handleCopy = async (command: Command) => {
  try {
    await navigator.clipboard.writeText(command.content)
    message.success('命令已复制到剪贴板')
  } catch (error) {
    message.error('复制失败')
  }
  emit('copy-command', command)
}
</script>

<template>
  <div class="command-group">
    <div class="group-header">
      <div class="group-title" @click="toggleCollapse">
        <NIcon size="16" class="collapse-icon" :class="{ 'collapse-icon-rotated': !collapsed }">
          <ChevronForward />
        </NIcon>
        {{ groupName }}
        ({{ commands.length }})
      </div>
      <div class="group-actions">
        <NTooltip v-if="group">
          <template #trigger>
            <NButton
              quaternary
              size="tiny"
              @click="emit('edit-group')"
            >
              <template #icon>
                <NIcon><Create /></NIcon>
              </template>
            </NButton>
          </template>
          编辑分组
        </NTooltip>
        <NTooltip v-if="group">
          <template #trigger>
            <NPopconfirm
              positive-text="确定"
              negative-text="取消"
              @positive-click="emit('delete-group')"
            >
              确定要删除该分组吗？
              <template #trigger>
                <NButton quaternary size="tiny">
                  <template #icon>
                    <NIcon><Trash /></NIcon>
                  </template>
                </NButton>
              </template>
            </NPopconfirm>
          </template>
          删除分组
        </NTooltip>
      </div>
    </div>
    <NFlex v-show="!collapsed" direction="column" :gap="10">
      <template v-for="(command, index) in commands" :key="command.id">
        <NThing style="cursor: pointer;" @click="emit('quick-edit-command', command)">
          <template #header>
            <span style="font-size: 14px; font-weight: normal;">{{ command.name }}</span>
          </template>
          <template v-if="command.content && command.content !== command.name">
            <NEllipsis :line-clamp="2">
              <NText :depth="3">{{ command.content }}</NText>
            </NEllipsis>
          </template>
          <template #description v-if="command.description">
            <NEllipsis :line-clamp="2">
              <NText :depth="3">{{ command.description }}</NText>
            </NEllipsis>
          </template>
          <template #action>
            <NFlex :size="10" justify="space-between">
              <NTooltip>
                <template #trigger>
                  <NButton quaternary size="tiny" @click.stop="handleExecute(command)">
                    <template #icon>
                      <NIcon><Play /></NIcon>
                    </template>
                  </NButton>
                </template>
                执行命令
              </NTooltip>
              <NTooltip>
                <template #trigger>
                  <NButton
                    quaternary
                    size="tiny"
                    @click.stop="handleCopy(command)"
                  >
                    <template #icon>
                      <NIcon><Copy /></NIcon>
                    </template>
                  </NButton>
                </template>
                复制命令
              </NTooltip>
              <NTooltip>
                <template #trigger>
                  <NButton
                    quaternary
                    size="tiny"
                    @click.stop="emit('quick-edit-command', command)"
                  >
                    <template #icon>
                      <NIcon><CodeWorkingOutline /></NIcon>
                    </template>
                  </NButton>
                </template>
                快速编辑
              </NTooltip>
              <NTooltip>
                <template #trigger>
                  <NButton
                    quaternary
                    size="tiny"
                    @click.stop="emit('edit-command', command)"
                  >
                    <template #icon>
                      <NIcon><Create /></NIcon>
                    </template>
                  </NButton>
                </template>
                编辑命令
              </NTooltip>
              <NTooltip>
                <template #trigger>
                  <NPopconfirm
                    positive-text="确定"
                    negative-text="取消"
                    @positive-click="emit('add-to-ignore', command)"
                  >
                  确定要将该命令添加到忽略列表吗？该命令也将从命令库中删除。
                  <template #trigger>
                    <NButton @click.stop quaternary size="tiny">
                      <template #icon>
                        <NIcon><EyeOff /></NIcon>
                      </template>
                    </NButton>
                  </template>
                  </NPopconfirm>
                </template>
                添加到忽略列表
              </NTooltip>
              <NTooltip>
                <template #trigger>
                  <NPopconfirm
                    positive-text="确定"
                    negative-text="取消"
                    @positive-click="emit('delete-command', command)"
                  >
                  确定要删除该命令吗？
                  <template #trigger>
                    <NButton @click.stop quaternary size="tiny">
                      <template #icon>
                        <NIcon><Trash /></NIcon>
                      </template>
                    </NButton>
                  </template>
                  </NPopconfirm>
                </template>
                删除命令
              </NTooltip>
            </NFlex>
          </template>
        </NThing>
        <NDivider v-if="index < commands.length - 1" style="margin: 5px 0;"/>
      </template>
      <NThing v-if="commands.length === 0">
        暂无命令
      </NThing>
    </NFlex>
  </div>
</template>

<style scoped>
.command-group {
  margin-bottom: 16px;
}

.group-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.group-title {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  user-select: none;
}

.collapse-icon {
  transition: transform 0.2s ease;
}

.collapse-icon-rotated {
  transform: rotate(90deg);
}

.group-actions {
  display: flex;
  gap: 4px;
}
</style>
