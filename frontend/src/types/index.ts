import type { Terminal as XTerminal } from '@xterm/xterm'
import type { models } from '../../wailsjs/go/models'

export type Workspace = models.Workspace
export type IgnoreRule = models.IgnoreRule
export type CommandGroup = models.CommandGroup
export type Command = models.Command
export type TemplateParam = models.TemplateParam
export type TemplateOption = models.TemplateOption
export type RecentPath = models.RecentPath

export interface TemplateOptionData {
  label: string
  value: string
}

export interface TemplateParamData {
  name: string
  type: string
  description: string
  options: TemplateOptionData[]
}

export type ParamValueType = 'number' | 'select' | 'input' | 'file' | 'directory'

export interface ParsedTemplateParam {
  name: string
  type: ParamValueType
  description: string
  options: TemplateOptionData[]
}
export type WorkspaceExport = models.WorkspaceExport
export type DatabaseBackup = models.DatabaseBackup

export interface CommandData {
  id: number
  name: string
  content: string
  description: string
  groupId?: number
  workspaceId: number
  templateParams: TemplateParam[]
}

export interface TerminalInstance {
  id: string
  terminal: XTerminal
  container: HTMLDivElement
  workspacePath: string
  outputListener?: () => void
  closeListener?: () => void
}

export type CursorStyle = 'block' | 'bar' | 'underline'

export interface TerminalTab {
  id: string
  name: string
}

export type SplitDirection = 'horizontal' | 'vertical' | null

export interface SplitPaneData {
  id: string
  direction: SplitDirection
  first?: SplitPaneData
  second?: SplitPaneData
  tabs: TerminalTab[]
  activeTab: number
}

export interface TerminalConfig {
  cursorBlink: boolean
  cursorStyle: CursorStyle
  theme: {
    background: string
    foreground: string
    cursor: string
    black: string
    red: string
    green: string
    yellow: string
    blue: string
    magenta: string
    cyan: string
    white: string
    brightBlack: string
    brightRed: string
    brightGreen: string
    brightYellow: string
    brightBlue: string
    brightMagenta: string
    brightCyan: string
    brightWhite: string
  }
  fontSize: number
  fontFamily: string
  allowTransparency: boolean
  convertEol: boolean
  macOptionIsMeta: boolean
  macOptionClickForcesSelection: boolean
  scrollback: number
}
