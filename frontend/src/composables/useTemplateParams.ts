import { ref, computed, watch, type Ref } from 'vue'
import { models } from '../../wailsjs/go/models'
import type { TemplateParam, ParsedTemplateParam, ParamValueType, TemplateOptionData } from '../types'

const PLACEHOLDER_SOURCE = /\{\{([a-zA-Z0-9_\-]*)\}\}/g

function createPlaceholderRegex(): RegExp {
  return new RegExp(PLACEHOLDER_SOURCE.source, PLACEHOLDER_SOURCE.flags)
}

export function hasTemplateParams(content: string): boolean {
  return createPlaceholderRegex().test(content)
}

function isInsideQuotes(text: string, offset: number): boolean {
  let quoteCount = 0
  for (let i = 0; i < offset; i++) {
    if (text[i] === '"' && (i === 0 || text[i - 1] !== '\\')) {
      quoteCount++
    }
  }
  return quoteCount % 2 === 1
}

function templateParamToParsed(tp: TemplateParam): ParsedTemplateParam {
  return {
    name: tp.name,
    type: tp.type as ParamValueType,
    description: tp.description,
    options: tp.options.map(o => ({ label: o.label, value: o.value }))
  }
}

export function parsedToTemplateParam(parsed: ParsedTemplateParam): TemplateParam {
  return models.TemplateParam.createFrom({
    name: parsed.name,
    type: parsed.type,
    description: parsed.description,
    options: parsed.options.map(o => ({ label: o.label, value: o.value }))
  })
}

export function useTemplateParams(content: Ref<string>, savedParams: Ref<TemplateParam[]>) {
  const placeholderCount = computed(() => {
    const matches = content.value.match(createPlaceholderRegex())
    return matches ? matches.length : 0
  })

  const hasPlaceholders = computed(() => placeholderCount.value > 0)

  const params = ref<ParsedTemplateParam[]>([])

  const paramValues = ref<Record<string, string | number | null>>({})

  const syncParamsFromContent = () => {
    const regex = createPlaceholderRegex()
    const allMatches = [...content.value.matchAll(regex)]
    const seenNames = new Set<string>()
    const uniqueNames: string[] = []

    for (const match of allMatches) {
      const name = match[1]
      if (!seenNames.has(name)) {
        seenNames.add(name)
        uniqueNames.push(name)
      }
    }

    const existingParamsByName = new Map<string, ParsedTemplateParam>()
    for (const p of params.value) {
      existingParamsByName.set(p.name, { ...p, options: [...p.options] })
    }

    const savedParamsByName = new Map<string, ParsedTemplateParam>()
    for (const p of savedParams.value) {
      savedParamsByName.set(p.name, templateParamToParsed(p))
    }

    const newParams: ParsedTemplateParam[] = []
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
        })
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
      const updated: ParsedTemplateParam = {
        ...params.value[index],
        type,
        options: type === 'select' ? params.value[index].options : []
      }
      params.value[index] = updated
      if (type === 'select') {
        if (updated.options.length === 0) {
          paramValues.value[paramName] = null
        }
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
      params.value[index] = { ...params.value[index], name }
      if (oldName !== name && oldName in paramValues.value) {
        paramValues.value[name] = paramValues.value[oldName]
        delete paramValues.value[oldName]
      }
    }
  }

  const updateParamDescription = (index: number, description: string) => {
    if (index >= 0 && index < params.value.length) {
      params.value[index] = { ...params.value[index], description }
    }
  }

  const updateParamOptions = (index: number, options: TemplateOptionData[]) => {
    if (index >= 0 && index < params.value.length) {
      const paramName = params.value[index].name
      params.value[index] = { ...params.value[index], options }
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
    result = result.replace(createPlaceholderRegex(), (_match: string, name: string, offset: number) => {
      const value = paramValues.value[name]
      const resolvedValue = value !== undefined && value !== null ? String(value) : ''

      if (autoQuote && !isInsideQuotes(result, offset)) {
        return `"${resolvedValue}"`
      }

      return resolvedValue
    })
    return result
  }

  watch([content, savedParams], ([, newSavedParams], [, oldSavedParams]) => {
    if (newSavedParams !== oldSavedParams) {
      params.value = []
      paramValues.value = {}
    }
    syncParamsFromContent()
  })

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