import type {
  TerminalBlacklist,
  TerminalBlacklistForm,
  TerminalSessionsResponse,
  TerminalCommandsResponse,
} from '@/types/terminal-audit'
import { fetcher, FetcherMethod } from './api'

// Get terminal sessions list
export const getTerminalSessions = async (
  page: number = 1,
  pageSize: number = 20,
  userId?: number,
  serverId?: number,
): Promise<TerminalSessionsResponse> => {
  const params = new URLSearchParams({
    page: page.toString(),
    page_size: pageSize.toString(),
  })
  if (userId) params.append('user_id', userId.toString())
  if (serverId) params.append('server_id', serverId.toString())

  const response = await fetcher<TerminalSessionsResponse>(
    FetcherMethod.GET,
    `/api/v1/terminal/sessions?${params.toString()}`,
  )
  return response
}

// Get terminal commands list
export const getTerminalCommands = async (
  page: number = 1,
  pageSize: number = 50,
  sessionId?: number,
): Promise<TerminalCommandsResponse> => {
  const params = new URLSearchParams({
    page: page.toString(),
    page_size: pageSize.toString(),
  })
  if (sessionId) params.append('session_id', sessionId.toString())

  const response = await fetcher<TerminalCommandsResponse>(
    FetcherMethod.GET,
    `/api/v1/terminal/commands?${params.toString()}`,
  )
  return response
}

// Get blacklist rules
export const getTerminalBlacklist = async (): Promise<TerminalBlacklist[]> => {
  const response = await fetcher<TerminalBlacklist[]>(
    FetcherMethod.GET,
    '/api/v1/terminal/blacklist',
  )
  return response || []
}

// Create blacklist rule
export const createTerminalBlacklist = async (
  data: TerminalBlacklistForm,
): Promise<number> => {
  const response = await fetcher<number>(
    FetcherMethod.POST,
    '/api/v1/terminal/blacklist',
    data,
  )
  return response
}

// Update blacklist rule
export const updateTerminalBlacklist = async (
  id: number,
  data: TerminalBlacklistForm,
): Promise<void> => {
  await fetcher<void>(
    FetcherMethod.PATCH,
    `/api/v1/terminal/blacklist/${id}`,
    data,
  )
}

// Delete blacklist rule
export const deleteTerminalBlacklist = async (id: number): Promise<void> => {
  await fetcher<void>(
    FetcherMethod.DELETE,
    `/api/v1/terminal/blacklist/${id}`,
  )
}
