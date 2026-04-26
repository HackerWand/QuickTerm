import { ref, computed, watch, type Ref } from 'vue'
import type { TemplateParam, TemplateOption } from '../types'

const PLACEHOLDER_PATTERN = /\{\{[a-zA-Z0-9_\-]*\}\}/g

export type ParamValueType = 'number' | 'select' | 'input' | 'file' | 'directory'

export function useTemplateParams(content: Ref<string>, savedParams: Ref<TemplateParam[]>) {
  const placeholderCount = computed(() => {
    const matches = content.value.match(PLACEHOLDER_PATTERN)
    return matches ? matches.length : 0
  })

  const hasPlaceholders = computed(() => placeholderCount.value > 0)

  const params = ref<TemplateParam[]>([])

  const paramValues = ref<Record<number, string | number | null>>({})

  const syncParamsFromContent = () => {
    const matches = content.value.match(PLACEHOLDER_PATTERN) ?? []
    const count = matches.length
    const existingParams = [...params.value]
    const newParams: TemplateParam[] = []

    for (let i = 0; i < count; i++) {
      if (i < existingParams.length) {
        newParams.push({ ...existingParams[i] } as unknown as TemplateParam)
      } else if (i < savedParams.value.length) {
        newParams.push({ ...savedParams.value[i] } as unknown as TemplateParam)
      } else {
        const extractedName = matches[i].replace(/^\{\{|\}\}$/g, '')
        newParams.push({
          name: extractedName,
          type: 'input' as ParamValueType,
          description: '',
          options: []
        } as unknown as TemplateParam)
      }
    }

    params.value = newParams

    const newValues: Record<number, string | number | null> = {}
    for (let i = 0; i < count; i++) {
      if (i < Object.keys(paramValues.value).length && paramValues.value[i] !== undefined) {
        newValues[i] = paramValues.value[i]
      } else {
        const param = newParams[i]
        if (param.type === 'number') {
          newValues[i] = 0
        } else if (param.type === 'select' && param.options.length > 0) {
          newValues[i] = param.options[0].value
        } else {
          newValues[i] = ''
        }
      }
    }
    paramValues.value = newValues
  }

  const updateParamType = (index: number, type: ParamValueType) => {
    if (index >= 0 && index < params.value.length) {
      params.value[index] = { ...params.value[index], type } as unknown as TemplateParam
      if (type === 'select') {
        if (!params.value[index].options || params.value[index].options.length === 0) {
          params.value[index].options = []
        }
        paramValues.value[index] = null
      } else if (type === 'number') {
        paramValues.value[index] = 0
      } else {
        paramValues.value[index] = ''
      }
    }
  }

  const updateParamName = (index: number, name: string) => {
    if (index >= 0 && index < params.value.length) {
      params.value[index] = { ...params.value[index], name } as unknown as TemplateParam
    }
  }

  const updateParamDescription = (index: number, description: string) => {
    if (index >= 0 && index < params.value.length) {
      params.value[index] = { ...params.value[index], description } as unknown as TemplateParam
    }
  }

  const updateParamOptions = (index: number, options: TemplateOption[]) => {
    if (index >= 0 && index < params.value.length) {
      params.value[index] = { ...params.value[index], options } as unknown as TemplateParam
      if (params.value[index].type === 'select') {
        const currentVal = paramValues.value[index]
        if (!options.some(opt => opt.value === String(currentVal))) {
          paramValues.value[index] = options.length > 0 ? options[0].value : ''
        }
      }
    }
  }

  const resolveContent = (): string => {
    let result = content.value
    let paramIndex = 0
    result = result.replace(PLACEHOLDER_PATTERN, () => {
      const value = paramValues.value[paramIndex]
      const resolvedValue = value !== undefined ? String(value) : ''
      paramIndex++
      return resolvedValue
    })
    return result
  }

  watch(content, () => {
    syncParamsFromContent()
  })

  watch(savedParams, () => {
    syncParamsFromContent()
  }, { deep: true })

  return {
    placeholderCount,
    hasPlaceholders,
    params,
    paramValues,
    syncParamsFromContent,
    updateParamType,
    updateParamName,
    updateParamDescription,
    updateParamOptions,
    resolveContent
  }
}
