<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  NModal,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace
} from 'naive-ui'
import type { CommandGroup } from '../../types'

interface Props {
  show: boolean
  group: CommandGroup | null
  workspaceId: number
}

interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'save', group: Omit<CommandGroup, 'id'>): void
  (e: 'update', group: CommandGroup): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const localShow = ref(false)
const name = ref('')
const isLoading = ref(false)

watch(() => props.show, (newVal) => {
  localShow.value = newVal
  if (newVal && props.group) {
    name.value = props.group.name
  } else if (newVal) {
    name.value = ''
  }
})

watch(localShow, (newVal) => {
  emit('update:show', newVal)
})

const handleSave = async () => {
  if (!name.value.trim()) {
    return
  }
  
  isLoading.value = true
  
  if (props.group) {
    emit('update', {
      ...props.group,
      name: name.value
    })
  } else {
    emit('save', {
      name: name.value,
      workspaceId: props.workspaceId
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
    :title="group ? '编辑分组' : '新建分组'"
    style="width: 400px"
    :bordered="false"
    size="huge"
    :segmented="{ content: 'soft', footer: 'soft' }"
  >
    <NForm :model="{ name }">
      <NFormItem label="名称" required>
        <NInput v-model:value="name" placeholder="请输入分组名称" maxlength="50" />
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
