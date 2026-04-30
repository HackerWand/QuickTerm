import { ref, computed, watch, type Ref } from 'vue'
import type { TemplateParam, TemplateOption } from '../types'

const PLACEHOLDER_PATTERN = /\{\{([a-zA-Z0-9_\-]*)\}\}/g

export type ParamValueType = 'number' | 'select' | 'input' | 'file' | 'directory'

function isInsideQuotes(text: string, offset: number): boolean {
  let quoteCount = 0
  for (let i = 0; i < offset; i++) {
    if (text[i] === '"' && (i === 0 || text[i - 1] !== '\\')) {
      quoteCount++
    }
  }
  return quoteCount % 2 === 1
}

export function useTemplateParams(content: Ref<string>, savedParams: Ref<TemplateParam[]>) {
  const placeholderCount = computed(() => {
    const matches = content.value.match(PLACEHOLDER_PATTERN)
    return matches ? matches.length : 0
  })

  const hasPlaceholders = computed(() => placeholderCount.value > 0)

  const params = ref<TemplateParam[]>([])

  const paramValues = ref<Record<string, string | number | null>>({})

  const syncParamsFromContent = () => {
    const allMatches = [...content.value.matchAll(PLACEHOLDER_PATTERN)]
    const seenNames = new Set<string>()
    const uniqueNames: string[] = []

    for (const match of allMatches) {
      const name = match[1]
      if (!seenNames.has(name)) {
        seenNames.add(name)
        uniqueNames.push(name)
      }
    }

    const existingParamsByName = new Map<string, TemplateParam>()
    for (const p of params.value) {
      existingParamsByName.set(p.name, { ...p } as unknown as TemplateParam)
    }

    const savedParamsByName = new Map<string, TemplateParam>()
    for (const p of savedParams.value) {
      savedParamsByName.set(p.name, { ...p } as unknown as TemplateParam)
    }

    const newParams: TemplateParam[] = []
    for (const name of uniqueNames) {
      if (existingParamsByName.has(name)) {
        newParams.push(existingParamsByName.get(name)!)
      } else if (savedParamsByName.has(name)) {
        newParams.push(savedParamsByName.get(name)!)
      } else {
        newParams.push({
          name,
          type: 'input' as ParamValueType,
          description: '',
          options: []
        } as unknown as TemplateParam)
      }
    }

    params.value = newParams

    const newValues: Record<string, string | number | null> = {}
    for (const param of newParams) {
      if (param.name in paramValues.value && paramValues.value[param.name] !== undefined) {
        newValues[param.name] = paramValues.value[param.name]
      } else {
        if (param.type === 'number') {
          newValues[param.name] = 0
        } else if (param.type === 'select' && param.options.length > 0) {
          newValues[param.name] = param.options[0].value
        } else {
          newValues[param.name] = ''
        }
      }
    }
    paramValues.value = newValues
  }

  const updateParamType = (index: number, type: ParamValueType) => {
    if (index >= 0 && index < params.value.length) {
      const paramName = params.value[index].name
      params.value[index] = { ...params.value[index], type } as unknown as TemplateParam
      if (type === 'select') {
        if (!params.value[index].options || params.value[index].options.length === 0) {
          params.value[index].options = []
        }
        paramValues.value[paramName] = null
      } else if (type === 'number') {
        paramValues.value[paramName] = 0
      } else {
        paramValues.value[paramName] = ''
      }
    }
  }

  const updateParamName = (index: number, name: string) => {
    if (index >= 0 && index < params.value.length) {
      const oldName = params.value[index].name
      params.value[index] = { ...params.value[index], name } as unknown as TemplateParam
      if (oldName !== name && oldName in paramValues.value) {
        paramValues.value[name] = paramValues.value[oldName]
        delete paramValues.value[oldName]
      }
    }
  }

  const updateParamDescription = (index: number, description: string) => {
    if (index >= 0 && index < params.value.length) {
      params.value[index] = { ...params.value[index], description } as unknown as TemplateParam
    }
  }

  const updateParamOptions = (index: number, options: TemplateOption[]) => {
    if (index >= 0 && index < params.value.length) {
      const paramName = params.value[index].name
      params.value[index] = { ...params.value[index], options } as unknown as TemplateParam
      if (params.value[index].type === 'select') {
        const currentVal = paramValues.value[paramName]
        if (!options.some(opt => opt.value === String(currentVal))) {
          paramValues.value[paramName] = options.length > 0 ? options[0].value : ''
        }
      }
    }
  }

  const resolveContent = (autoQuote: boolean = false): string => {
    let result = content.value
    result = result.replace(PLACEHOLDER_PATTERN, (match: string, name: string, offset: number) => {
      const value = paramValues.value[name]
      const resolvedValue = value !== undefined && value !== null ? String(value) : ''

      if (autoQuote && !isInsideQuotes(result, offset)) {
        return `"${resolvedValue}"`
      }

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
