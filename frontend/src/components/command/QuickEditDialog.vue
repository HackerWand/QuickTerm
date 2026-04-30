<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import {
  NModal,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSelect,
  NButton,
  NSpace,
  NInputGroup,
  NIcon,
  NFlex,
  NText,
  useMessage,
  NAlert
} from 'naive-ui'
import { ClipboardOutline, FolderOpenOutline, DocumentTextOutline } from '@vicons/ionicons5'
import { OpenFileSelectorDialog, OpenDirectorySelectorDialog } from '../../../wailsjs/go/main/App'
import { useRecentPathStore } from '../../stores/recentPath'
import { useWorkspaceStore } from '../../stores/workspace'
import { useTemplateParams, parsedToTemplateParam } from '../../composables/useTemplateParams'
import type { Command, TemplateParam, ParsedTemplateParam, CommandData } from '../../types'
import TemplateParamManager from './TemplateParamManager.vue'

interface Props {
  show: boolean
  command: Command | null
}

interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'save', command: CommandData): void
  (e: 'execute', commandContent: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
const message = useMessage()
const recentPathStore = useRecentPathStore()
const workspaceStore = useWorkspaceStore()

const localShow = ref(false)
const content = ref('')
const isLoading = ref(false)
const savedParams = ref<TemplateParam[]>([])
const showParamManager = ref(false)

const workspaceId = computed(() => {
  return workspaceStore.currentWorkspace?.id || 0
})

const safePaths = computed(() => {
  return recentPathStore.safePaths()
})

const {
  hasPlaceholders,
  params,
  paramValues,
  resolveContent
} = useTemplateParams(content, savedParams)

const getSelectOptions = (param: ParsedTemplateParam) => {
  return param.options.map(opt => ({ label: opt.label, value: opt.value }))
}

const handleParamsUpdate = (updatedParams: ParsedTemplateParam[]) => {
  params.value = updatedParams.map(p => ({
    name: p.name,
    type: p.type,
    description: p.description,
    options: p.options.map(o => ({ label: o.label, value: o.value }))
  }))
  const newValues: Record<string, string | number | null> = {}
  for (const param of updatedParams) {
    const currentVal = paramValues.value[param.name]
    if (param.type === 'number') {
      newValues[param.name] = typeof currentVal === 'number' ? currentVal : 0
    } else if (param.type === 'select') {
      const optionValues = param.options.map(opt => opt.value)
      if (typeof currentVal === 'string' && optionValues.includes(currentVal)) {
        newValues[param.name] = currentVal
      } else {
        newValues[param.name] = optionValues.length > 0 ? optionValues[0] : ''
      }
    } else {
      newValues[param.name] = typeof currentVal === 'string' ? currentVal : ''
    }
  }
  paramValues.value = newValues
}

const loadPaths = async () => {
  if (!workspaceId.value) return
  try {
    await recentPathStore.loadPaths(workspaceId.value)
  } catch (error) {
    console.error('Failed to load recent paths:', error)
  }
}

watch(() => props.show, (newVal) => {
  localShow.value = newVal
  if (newVal) {
    if (props.command) {
      content.value = props.command.content
      savedParams.value = props.command.templateParams
        ? [...props.command.templateParams]
        : []
    } else {
      content.value = ''
      savedParams.value = []
    }
    loadPaths()
  }
})

watch(localShow, (newVal) => {
  emit('update:show', newVal)
})

const handleCopyPath = async (path: string) => {
  try {
    await navigator.clipboard.writeText(path)
    message.success('路径已复制到剪贴板')
  } catch (error) {
    message.error('复制失败')
  }
}

const handlePasteToParam = async (index: number) => {
  try {
    const text = await navigator.clipboard.readText()
    const paramName = params.value[index].name
    paramValues.value[paramName] = text
    message.success('已从剪贴板粘贴')
  } catch (error) {
    console.error('Failed to paste from clipboard:', error)
    message.error('粘贴失败，请检查剪贴板权限')
  }
}

const handleSelectFile = async (index: number) => {
  try {
    const path = await OpenFileSelectorDialog()
    if (path) {
      const paramName = params.value[index].name
      paramValues.value[paramName] = path
    }
  } catch (error) {
    console.error('Failed to select file:', error)
    message.error('选择文件失败')
  }
}

const handleSelectDirectory = async (index: number) => {
  try {
    const path = await OpenDirectorySelectorDialog()
    if (path) {
      const paramName = params.value[index].name
      paramValues.value[paramName] = path
    }
  } catch (error) {
    console.error('Failed to select directory:', error)
    message.error('选择目录失败')
  }
}

const buildUpdatedCommand = (): CommandData => {
  if (!props.command) {
    return {} as CommandData
  }
  return {
    id: props.command.id,
    name: props.command.name,
    content: content.value,
    description: props.command.description,
    groupId: props.command.groupId,
    workspaceId: props.command.workspaceId,
    templateParams: params.value.map(parsedToTemplateParam)
  }
}

const handleSave = async () => {
  if (!content.value.trim() || !props.command) {
    return
  }

  isLoading.value = true

  const updatedCommand = buildUpdatedCommand()
  emit('save', updatedCommand)

  isLoading.value = false
}

const handleExecute = async () => {
  if (!content.value.trim()) {
    return
  }

  const resolvedContent = resolveContent(true)
  emit('execute', resolvedContent)
}

const truncatePath = (path: string, maxLength: number = 35) => {
  if (path.length <= maxLength) {
    return path
  }
  const start = Math.floor((maxLength - 3) / 2)
  const end = Math.ceil((maxLength - 3) / 2)
  return path.slice(0, start) + '...' + path.slice(-end)
}
</script>

<template>
  <NModal
    v-model:show="localShow"
    preset="card"
    :title="command?.name + ' - 快捷编辑命令'"
    style="width: 550px"
    :bordered="false"
  >
    <div class="quick-edit-content">
      <div class="recent-paths-section" v-if="safePaths.length">
        <h4 class="section-subtitle">最近路径</h4>
        <div class="paths-list">
          <div v-if="safePaths.length === 0" class="empty-paths">
            暂无最近路径
          </div>
          <div v-else class="paths-container">
            <div
              v-for="path in safePaths"
              :key="path.id"
              class="path-item"
              :title="path.path"
              @click="handleCopyPath(path.path)"
            >
              <span class="path-text">{{ truncatePath(path.path, 100) }}</span>
            </div>
          </div>
        </div>
      </div>

      <NForm :model="{ content }">
        <NFormItem label="命令内容" required>
          <NInput
            v-model:value="content"
            type="textarea"
            placeholder="使用 ___ 作为占位符，例如：echo ___ && ___"
            :rows="4"
          />
        </NFormItem>

        <template v-if="hasPlaceholders">
          <div class="template-params-section">
            <div class="section-header">
              <h4 class="section-subtitle">模版参数</h4>
              <NButton text size="small" type="primary" @click="showParamManager = true">
                管理
              </NButton>
            </div>
            <NForm label-placement="left">
              <NFormItem v-for="(param, index) in params" :key="param.name" :label="param.name">
                <NFlex style="flex: 1;" vertical>
                  <NInputNumber
                    v-if="param.type === 'number'"
                    v-model:value="paramValues[param.name] as number"
                    placeholder="请输入数值"
                    clearable
                    :update-value-on-input="true"
                  />
                  <NSelect
                    clearable
                    v-else-if="param.type === 'select'"
                    v-model:value="paramValues[param.name] as string"
                    :options="getSelectOptions(param)"
                    placeholder="请选择"
                  />
                  <NInputGroup v-else-if="param.type === 'file'">
                    <NInput
                      clearable
                      v-model:value="paramValues[param.name] as string"
                      placeholder="请选择文件路径"
                      style="flex: 1;"
                    />
                    <NButton @click="handleSelectFile(index)">
                      <template #icon>
                        <NIcon><DocumentTextOutline /></NIcon>
                      </template>
                      选择文件
                    </NButton>
                  </NInputGroup>
                  <NInputGroup v-else-if="param.type === 'directory'">
                    <NInput
                      clearable
                      v-model:value="paramValues[param.name] as string"
                      placeholder="请选择目录路径"
                      style="flex: 1;"
                    />
                    <NButton @click="handleSelectDirectory(index)">
                      <template #icon>
                        <NIcon><FolderOpenOutline /></NIcon>
                      </template>
                      选择目录
                    </NButton>
                  </NInputGroup>
                  <NInputGroup v-else>
                    <NInput
                      clearable
                      v-model:value="paramValues[param.name] as string"
                      placeholder="请输入值"
                      style="flex: 1;"
                    />
                    <NButton @click="handlePasteToParam(index)">
                      <template #icon>
                        <NIcon><ClipboardOutline /></NIcon>
                      </template>
                      粘贴
                    </NButton>
                  </NInputGroup>
                  <NText v-if="param.description" :depth="3">{{ param.description }}</NText>
                </NFlex>
              </NFormItem>
            </NForm>
          </div>
        </template>
        <NAlert title="模版参数" type="info" :bordered="false">
          <span v-pre>使用{{name}}作为占位符，例如：echo {{param1}} && {{param2}}，可生成模版参数，用于动态替换命令中的参数。</span>
        </NAlert>
      </NForm>
    </div>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="localShow = false" :disabled="isLoading">取消</NButton>
        <NButton type="info" @click="handleExecute" :disabled="isLoading">直接执行</NButton>
        <NButton type="primary" @click="handleSave" :loading="isLoading" :disabled="!command">保存</NButton>
      </NSpace>
    </template>
  </NModal>

  <TemplateParamManager
    v-model:show="showParamManager"
    :params="params"
    @update:params="handleParamsUpdate"
  />
</template>

<style scoped>
.quick-edit-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.recent-paths-section {
  margin-bottom: 8px;
}

.section-subtitle {
  font-size: 13px;
  font-weight: 600;
  color: var(--n-color-text-2);
  margin: 0;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.paths-list {
  display: flex;
  flex-direction: column;
}

.paths-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  max-height: 100px;
  overflow-y: auto;
}

.path-item {
  display: inline-flex;
  align-items: center;
  padding: 6px 12px;
  background-color: var(--n-color-modal);
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
  border: 1px solid var(--n-color-border);
}

.path-item:hover {
  background-color: var(--n-color-popover);
}

.path-text {
  font-size: 12px;
  color: var(--n-color-text-2);
  white-space: nowrap;
}

.empty-paths {
  font-size: 12px;
  color: var(--n-color-text-3);
}

.paths-container::-webkit-scrollbar {
  width: 6px;
}

.paths-container::-webkit-scrollbar-track {
  background: transparent;
}

.paths-container::-webkit-scrollbar-thumb {
  background: var(--n-color-border);
  border-radius: 3px;
}

.template-params-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
</style>
