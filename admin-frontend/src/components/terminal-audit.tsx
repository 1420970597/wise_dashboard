import { useState } from 'react'
import useSWR from 'swr'
import { format } from 'date-fns'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { getTerminalSessions, getTerminalCommands } from '@/api/terminal-audit'
import type { TerminalSession } from '@/types/terminal-audit'

export const TerminalAuditSessions: React.FC = () => {
  const [page, setPage] = useState(1)
  const [selectedSession, setSelectedSession] = useState<TerminalSession | null>(null)
  const [commandsDialogOpen, setCommandsDialogOpen] = useState(false)

  const { data, error, isLoading } = useSWR(
    ['terminal-sessions', page],
    () => getTerminalSessions(page, 20),
    { refreshInterval: 10000 },
  )

  const formatDuration = (seconds: number) => {
    if (seconds < 60) return `${seconds}秒`
    if (seconds < 3600) return `${Math.floor(seconds / 60)}分 ${seconds % 60}秒`
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    return `${hours}小时 ${minutes}分`
  }

  const handleViewCommands = (session: TerminalSession) => {
    setSelectedSession(session)
    setCommandsDialogOpen(true)
  }

  if (isLoading) {
    return <div className="text-center py-8">加载中...</div>
  }

  if (error) {
    return <div className="text-center py-8 text-red-500">加载数据失败</div>
  }

  const sessions = data?.sessions || []
  const total = data?.total || 0
  const totalPages = Math.ceil(total / 20)

  return (
    <div className="space-y-4">
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>用户</TableHead>
              <TableHead>服务器</TableHead>
              <TableHead>开始时间</TableHead>
              <TableHead>持续时长</TableHead>
              <TableHead>命令数</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {sessions.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                  暂无会话记录
                </TableCell>
              </TableRow>
            ) : (
              sessions.map((session) => (
                <TableRow key={session.id}>
                  <TableCell className="font-medium">{session.username}</TableCell>
                  <TableCell>{session.server_name}</TableCell>
                  <TableCell>
                    {format(new Date(session.started_at), 'yyyy-MM-dd HH:mm:ss')}
                  </TableCell>
                  <TableCell>{formatDuration(session.duration)}</TableCell>
                  <TableCell>
                    <Badge variant="secondary">{session.command_count}</Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-2">
                      {session.ended_at ? (
                        <Badge variant="outline">已结束</Badge>
                      ) : (
                        <Badge variant="default">活跃中</Badge>
                      )}
                      {session.recording_enabled && (
                        <Badge variant="secondary">已录像</Badge>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleViewCommands(session)}
                      >
                        查看命令
                      </Button>
                      {session.recording_enabled && session.recording_path && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            window.open(`/api/v1/terminal/recording/${session.id}`, '_blank')
                          }}
                        >
                          下载录像
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-muted-foreground">
            第 {page} 页，共 {totalPages} 页（总计 {total} 条）
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
            >
              上一页
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
            >
              下一页
            </Button>
          </div>
        </div>
      )}

      {/* Commands Dialog */}
      <CommandsDialog
        session={selectedSession}
        open={commandsDialogOpen}
        onOpenChange={setCommandsDialogOpen}
      />
    </div>
  )
}

interface CommandsDialogProps {
  session: TerminalSession | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

const CommandsDialog: React.FC<CommandsDialogProps> = ({ session, open, onOpenChange }) => {
  const [page, setPage] = useState(1)

  const { data, error, isLoading } = useSWR(
    session && open ? ['terminal-commands', session.id, page] : null,
    () => getTerminalCommands(page, 50, session!.id),
  )

  if (!session) return null

  const commands = data?.commands || []
  const total = data?.total || 0
  const totalPages = Math.ceil(total / 50)

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl">
        <DialogHeader>
          <DialogTitle>
            终端命令 - {session.username}@{session.server_name}
          </DialogTitle>
          <DialogDescription>
            会话开始时间: {format(new Date(session.started_at), 'yyyy-MM-dd HH:mm:ss')}
          </DialogDescription>
        </DialogHeader>

        <ScrollArea className="max-h-[60vh]">
          {isLoading ? (
            <div className="text-center py-8">加载中...</div>
          ) : error ? (
            <div className="text-center py-8 text-red-500">加载命令失败</div>
          ) : commands.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              暂无命令记录
            </div>
          ) : (
            <div className="space-y-2 p-2">
              {commands.map((cmd) => (
                <div
                  key={cmd.id}
                  className={`p-3 rounded-lg border ${
                    cmd.blocked ? 'bg-red-50 border-red-200 dark:bg-red-950/20' : 'bg-muted/50'
                  }`}
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex-1 space-y-1">
                      <div className="flex items-center gap-2">
                        <code className="text-sm font-mono bg-background px-2 py-1 rounded">
                          {cmd.command}
                        </code>
                        {cmd.blocked && (
                          <Badge variant="destructive" className="text-xs">
                            已拦截
                          </Badge>
                        )}
                      </div>
                      <div className="text-xs text-muted-foreground space-y-0.5">
                        <div>
                          工作目录: <code>{cmd.working_dir}</code>
                        </div>
                        <div>
                          执行时间: {format(new Date(cmd.executed_at), 'yyyy-MM-dd HH:mm:ss')}
                        </div>
                        {!cmd.blocked && (
                          <div>
                            退出码: <code>{cmd.exit_code}</code>
                          </div>
                        )}
                        {cmd.blocked && cmd.block_reason && (
                          <div className="text-red-600 dark:text-red-400">
                            拦截原因: {cmd.block_reason}
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </ScrollArea>

        {totalPages > 1 && (
          <div className="flex items-center justify-between pt-4 border-t">
            <div className="text-sm text-muted-foreground">
              第 {page} 页，共 {totalPages} 页
            </div>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
              >
                上一页
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
              >
                下一页
              </Button>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
