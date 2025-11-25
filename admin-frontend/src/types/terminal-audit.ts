// Terminal Audit Types

export interface TerminalSession {
  id: number
  created_at: string
  updated_at: string
  user_id: number
  username: string
  server_id: number
  server_name: string
  stream_id: string
  started_at: string
  ended_at?: string
  duration: number
  command_count: number
  recording_path?: string
  recording_enabled: boolean
}

export interface TerminalCommand {
  id: number
  created_at: string
  updated_at: string
  session_id: number
  user_id: number
  server_id: number
  command: string
  working_dir: string
  executed_at: string
  exit_code: number
  blocked: boolean
  block_reason?: string
}

export interface TerminalBlacklist {
  id: number
  created_at: string
  updated_at: string
  pattern: string
  description: string
  action: 'block' | 'warn' | 'log'
  enabled: boolean
  created_by: number
}

export interface TerminalBlacklistForm {
  pattern: string
  description: string
  action: 'block' | 'warn' | 'log'
  enabled: boolean
}

export interface TerminalSessionsResponse {
  sessions: TerminalSession[]
  total: number
  page: number
  pageSize: number
}

export interface TerminalCommandsResponse {
  commands: TerminalCommand[]
  total: number
  page: number
  pageSize: number
}

export interface CommonResponse<T> {
  success: boolean
  data?: T
  error?: string
}
