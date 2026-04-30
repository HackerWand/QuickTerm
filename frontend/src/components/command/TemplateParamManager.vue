<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  NModal,
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NButton,
  NSpace,
  NList,
  NListItem,
  useMessage
} from 'naive-ui'
import type { ParsedTemplateParam, ParamValueType, TemplateOptionData } from '../../types'
import SelectOptionManager from './SelectOptionManager.vue'

interface Props {
  show: boolean
  params: ParsedTemplateParam[]
}

interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'update:params', params: ParsedTemplateParam[]): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
const message = useMessage()

const localShow = ref(false)
const localParams = ref<ParsedTemplateParam[]>([])
const showOptionManager = ref(false)
const managingOptionIndex = ref(-1)

const paramTypeOptions = [
  { label: '输入框', value: 'input' },
  { label: '数字', value: 'number' },
  { label: '选项', value: 'select' },
  { label: '文件', value: 'file' },
  { label: '目录', value: 'directory' }
]

watch(() => props.show, (newVal) => {
  localShow.value = newVal
  if (newVal) {
    localParams.value = props.params.map(p => ({
      name: p.name,
      type: p.type,
      description: p.description ?? '',
      options: p.options.map(o => ({ label: o.label, value: o.value }))
    }))
  }
})

watch(localShow, (newVal) => {
  emit('update:show', newVal)
})

const updateLocalParamName = (index: number, name: string) => {
  if (index >= 0 && index < localParams.value.length) {
    localParams.value[index] = { ...localParams.value[index], name }
  }
}

const updateLocalParamDescription = (index: number, description: string) => {
  if (index >= 0 && index < localParams.value.length) {
    localParams.value[index] = { ...localParams.value[index], description }
  }
}

const updateLocalParamType = (index: number, type: ParamValueType) => {
  if (index >= 0 && index < localParams.value.length) {
    const updated: ParsedTemplateParam = {
      ...localParams.value[index],
      type,
      options: type === 'select' ? localParams.value[index].options : []
    }
    localParams.value[index] = updated
  }
}

const updateLocalParamOptions = (index: number, options: TemplateOptionData[]) => {
  if (index >= 0 && index < localParams.value.length) {
    localParams.value[index] = { ...localParams.value[index], options }
  }
}

const handleManageOptions = (index: number) => {
  managingOptionIndex.value = index
  showOptionManager.value = true
}

const handleOptionsUpdate = (newOptions: TemplateOptionData[]) => {
  if (managingOptionIndex.value >= 0) {
    updateLocalParamOptions(managingOptionIndex.value, newOptions)
  }
}

const handleSave = () => {
  for (let i = 0; i < localParams.value.length; i++) {
    const param = localParams.value[i]
    if (!param.name.trim()) {
      message.error(`参数 ${i + 1} 的名称不能为空`)
      return
    }
    if (param.type === 'select' && param.options.length === 0) {
      message.error(`参数 ${i + 1} 为选项类型，至少需要一个选项`)
      return
    }
  }
  emit('update:params', localParams.value.map(p => ({
    name: p.name,
    type: p.type,
    description: p.description,
    options: p.options.map(o => ({ label: o.label, value: o.value }))
  })))
  localShow.value = false
}
</script>

<template>
  <NModal
    v-model:show="localShow"
    preset="card"
    title="管理模版参数"
    style="width: 550px"
    :bordered="false"
  >
    <div class="param-manager-content">
      <NList v-if="localParams.length > 0" bordered size="small">
        <NListItem v-for="(param, index) in localParams" :key="index">
          <div class="param-config">
            <NForm :model="{}" :show-feedback="false">
              <NFormItem :label="`参数 ${index + 1}`">
                <div class="param-config-inner">
                  <div class="param-config-row">
                    <NInput
                      :value="param.name"
                      @update:value="(val: string) => updateLocalParamName(index, val)"
                      placeholder="参数名称"
                      style="flex: 1;"
                    />
                    <NSelect
                      :value="param.type"
                      @update:value="(val: ParamValueType) => updateLocalParamType(index, val)"
                      :options="paramTypeOptions"
                      style="width: 100px;"
                    />
                  </div>
                  <NInput
                    :value="param.description"
                    @update:value="(val: string) => updateLocalParamDescription(index, val)"
                    placeholder="参数描述（可选）"
                    size="small"
                  />
                  <div v-if="param.type === 'select'" class="param-options-row">
                    <span class="options-label">选项列表 ({{ param.options.length }}个)</span>
                    <NButton size="small" @click="handleManageOptions(index)">
                      管理选项
                    </NButton>
                  </div>
                </div>
              </NFormItem>
            </NForm>
          </div>
        </NListItem>
      </NList>
      <div v-else class="empty-params">暂无模版参数</div>
    </div>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="localShow = false">取消</NButton>
        <NButton type="primary" @click="handleSave">确定</NButton>
      </NSpace>
    </template>
  </NModal>

  <SelectOptionManager
    v-model:show="showOptionManager"
    :options="managingOptionIndex >= 0 ? localParams[managingOptionIndex]?.options || [] : []"
    @update:options="handleOptionsUpdate"
  />
</template>

<style scoped>
.param-manager-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.param-config {
  width: 100%;
}

.param-config-inner {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.param-config-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.param-options-row {
  display: flex;
  gap: 8px;
  align-items: center;
  justify-content: space-between;
}

.options-label {
  font-size: 12px;
  color: var(--n-color-text-3);
}

.empty-params {
  font-size: 12px;
  color: var(--n-color-text-3);
  text-align: center;
  padding: 12px 0;
}
</style>