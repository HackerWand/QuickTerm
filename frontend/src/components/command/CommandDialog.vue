<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import {
  NModal,
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NButton,
  NSpace
} from 'naive-ui'
import type { Command, CommandGroup, CommandData } from '../../types'

interface Props {
  show: boolean
  command: Command | null
  groups: CommandGroup[]
  workspaceId: number
}

interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'save', command: Omit<CommandData, 'id'>): void
  (e: 'update', command: CommandData): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const localShow = ref(false)
const name = ref('')
const content = ref('')
const description = ref('')
const groupId = ref<number | undefined>(undefined)
const isLoading = ref(false)

const groupOptions = computed(() => {
  return [
    { label: '无分组', value: undefined as number | undefined },
    ...props.groups.map(group => ({
      label: group.name,
      value: group.id
    }))
  ]
})

watch(() => props.show, (newVal) => {
  localShow.value = newVal
  if (newVal && props.command) {
    name.value = props.command.name
    content.value = props.command.content
    description.value = props.command.description
    groupId.value = props.command.groupId
  } else if (newVal) {
    name.value = ''
    content.value = ''
    description.value = ''
    groupId.value = undefined
  }
})

watch(localShow, (newVal) => {
  emit('update:show', newVal)
})

const handleSave = async () => {
  if (!name.value.trim() || !content.value.trim()) {
    return
  }
  
  isLoading.value = true
  
  if (props.command) {
    emit('update', {
      id: props.command.id,
      name: name.value,
      content: content.value,
      description: description.value,
      groupId: groupId.value,
      workspaceId: props.command.workspaceId,
      templateParams: props.command.templateParams || []
    })
  } else {
    emit('save', {
      name: name.value,
      content: content.value,
      description: description.value,
      groupId: groupId.value,
      workspaceId: props.workspaceId,
      templateParams: []
    })
  }
  
  isLoading.value = false
  emit('update:show', false)
}
</script>

<template>
  <NModal
    v-model:show="localShow"
    preset="card"
    :title="command ? '编辑命令' : '新建命令'"
    style="width: 500px"
    :bordered="false"
    size="huge"
    :segmented="{ content: 'soft', footer: 'soft' }"
  >
    <NForm :model="{ name, content, description, groupId }">
      <NFormItem label="名称" required>
        <NInput v-model:value="name" placeholder="请输入命令名称" maxlength="100" />
      </NFormItem>
      <NFormItem label="命令内容" required>
        <NInput
          v-model:value="content"
          type="textarea"
          placeholder="请输入命令内容"
          :rows="3"
        />
      </NFormItem>
      <NFormItem label="描述">
        <NInput
          v-model:value="description"
          type="textarea"
          placeholder="请输入命令描述（可选）"
          :rows="2"
        />
      </NFormItem>
      <NFormItem label="分组">
        <NSelect
          v-model:value="groupId"
          :options="groupOptions"
          placeholder="选择分组"
        />
      </NFormItem>
    </NForm>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="localShow = false" :disabled="isLoading">取消</NButton>
        <NButton type="primary" @click="handleSave" :loading="isLoading">保存</NButton>
      </NSpace>
    </template>
  </NModal>
</template>
