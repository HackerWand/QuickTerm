<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  NModal,
  NInput,
  NButton,
  NSpace,
  NList,
  NListItem,
  NIcon,
  NPopconfirm,
  useMessage,
  NInputGroup
} from 'naive-ui'
import { Add, TrashOutline } from '@vicons/ionicons5'
import type { TemplateOption } from '../../types'

interface Props {
  show: boolean
  options: TemplateOption[]
}

interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'update:options', options: TemplateOption[]): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
const message = useMessage()

const localShow = ref(false)
const localOptions = ref<TemplateOption[]>([])
const newLabel = ref('')
const newValue = ref('')

watch(() => props.show, (newVal) => {
  localShow.value = newVal
  if (newVal) {
    localOptions.value = [...props.options]
    newLabel.value = ''
    newValue.value = ''
  }
})

watch(localShow, (newVal) => {
  emit('update:show', newVal)
})

const addOption = () => {
  const trimmedLabel = newLabel.value.trim()
  const trimmedValue = newValue.value.trim()
  if (!trimmedLabel || !trimmedValue) {
    message.error('选项标签和值不能为空')
    return
  }
  if (localOptions.value.some(opt => opt.value === trimmedValue)) {
    message.error('选项值已存在')
    return
  }
  localOptions.value.push({ label: trimmedLabel, value: trimmedValue })
  newLabel.value = ''
  newValue.value = ''
}

const removeOption = (index: number) => {
  localOptions.value.splice(index, 1)
}

const handleSave = () => {
  if (localOptions.value.length === 0) {
    message.error('至少需要一个选项')
    return
  }
  emit('update:options', [...localOptions.value])
  localShow.value = false
}
</script>

<template>
  <NModal
    v-model:show="localShow"
    preset="card"
    title="管理选项"
    style="width: 500px"
    :bordered="false"
  >
    <div class="option-manager-content">
      <NList bordered v-if="localOptions.length > 0" size="small">
        <NListItem v-for="(option, index) in localOptions" :key="index">
          <span class="option-text">{{ option.label }} ({{ option.value }})</span>
          <template #suffix>
            <NPopconfirm @positive-click="removeOption(index)">
              确认删除吗？
              <template #trigger>
                <NButton text size="tiny" type="error">
                  <template #icon>
                    <NIcon><TrashOutline /></NIcon>
                  </template>
                </NButton>
              </template>
            </NPopconfirm>
          </template>
        </NListItem>
      </NList>
      <div v-else class="empty-options">暂无选项</div>
      <div class="add-option-section" style="margin-top: 12px;">
        <NInputGroup>
          <NInput
            v-model:value="newLabel"
            placeholder="选项标签"
            style="flex: 1;"
            @keyup.enter="addOption"
          />
          <NInput
            v-model:value="newValue"
            placeholder="选项值"
            style="flex: 1;"
            @keyup.enter="addOption"
          />
          <NButton @click="addOption">
            <template #icon>
              <NIcon><Add /></NIcon>
            </template>
            添加
          </NButton>
        </NInputGroup>
      </div>
    </div>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="localShow = false">取消</NButton>
        <NButton type="primary" @click="handleSave">确定</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped>
.option-manager-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.option-text {
  font-size: 13px;
}

.empty-options {
  font-size: 12px;
  color: var(--n-color-text-3);
  text-align: center;
  padding: 12px 0;
}

.add-option-section {
  padding-top: 4px;
}
</style>
