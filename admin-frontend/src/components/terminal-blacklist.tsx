import { useState } from 'react'
import useSWR from 'swr'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
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
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogClose,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { IconButton } from '@/components/xui/icon-button'
import {
  getTerminalBlacklist,
  createTerminalBlacklist,
  updateTerminalBlacklist,
  deleteTerminalBlacklist,
} from '@/api/terminal-audit'
import type { TerminalBlacklist } from '@/types/terminal-audit'

const blacklistFormSchema = z.object({
  pattern: z.string().min(1, '匹配模式不能为空'),
  description: z.string().min(1, '描述不能为空'),
  action: z.enum(['block', 'warn', 'log']),
  enabled: z.boolean(),
})

type BlacklistFormData = z.infer<typeof blacklistFormSchema>

interface BlacklistCardProps {
  data?: TerminalBlacklist
  mutate: () => void
}

const BlacklistCard: React.FC<BlacklistCardProps> = ({ data, mutate }) => {
  const [open, setOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)

  const form = useForm<BlacklistFormData>({
    resolver: zodResolver(blacklistFormSchema),
    defaultValues: data
      ? {
          pattern: data.pattern,
          description: data.description,
          action: data.action,
          enabled: data.enabled,
        }
      : {
          pattern: '',
          description: '',
          action: 'block',
          enabled: true,
        },
  })

  const onSubmit = async (values: BlacklistFormData) => {
    try {
      if (data?.id) {
        await updateTerminalBlacklist(data.id, values)
        toast.success('成功', {
          description: '黑名单规则已更新',
        })
      } else {
        await createTerminalBlacklist(values)
        toast.success('成功', {
          description: '黑名单规则已创建',
        })
      }
      setOpen(false)
      mutate()
      form.reset()
    } catch (e) {
      console.error(e)
      toast.error('错误', {
        description: '保存黑名单规则失败',
      })
    }
  }

  const handleDelete = async () => {
    if (!data?.id) return
    try {
      await deleteTerminalBlacklist(data.id)
      toast.success('成功', {
        description: '黑名单规则已删除',
      })
      setDeleteDialogOpen(false)
      mutate()
    } catch (e) {
      console.error(e)
      toast.error('错误', {
        description: '删除黑名单规则失败',
      })
    }
  }

  return (
    <>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogTrigger asChild>
          {data ? (
            <IconButton variant="outline" icon="edit" />
          ) : (
            <IconButton icon="plus" />
          )}
        </DialogTrigger>
        <DialogContent className="sm:max-w-xl">
          <ScrollArea className="max-h-[calc(100dvh-5rem)] p-3">
            <div className="items-center mx-1">
              <DialogHeader>
                <DialogTitle>
                  {data ? '编辑黑名单规则' : '创建黑名单规则'}
                </DialogTitle>
                <DialogDescription>
                  配置命令匹配模式以拦截、警告或记录
                </DialogDescription>
              </DialogHeader>
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4 my-4">
                  <FormField
                    control={form.control}
                    name="pattern"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>匹配模式</FormLabel>
                        <FormControl>
                          <Input placeholder="^rm\s+-rf.*" {...field} />
                        </FormControl>
                        <FormDescription>
                          用于匹配命令的正则表达式
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="description"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>描述</FormLabel>
                        <FormControl>
                          <Textarea
                            placeholder="描述此规则的作用"
                            {...field}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="action"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>动作</FormLabel>
                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="选择动作" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            <SelectItem value="block">拦截</SelectItem>
                            <SelectItem value="warn">警告</SelectItem>
                            <SelectItem value="log">仅记录</SelectItem>
                          </SelectContent>
                        </Select>
                        <FormDescription>
                          拦截：阻止执行，警告：显示警告，仅记录：只记录不阻止
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="enabled"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3">
                        <div className="space-y-0.5">
                          <FormLabel>启用</FormLabel>
                          <FormDescription>
                            启用或禁用此规则
                          </FormDescription>
                        </div>
                        <FormControl>
                          <Switch checked={field.value} onCheckedChange={field.onChange} />
                        </FormControl>
                      </FormItem>
                    )}
                  />

                  <DialogFooter>
                    <DialogClose asChild>
                      <Button variant="outline">取消</Button>
                    </DialogClose>
                    <Button type="submit">保存</Button>
                  </DialogFooter>
                </form>
              </Form>
            </div>
          </ScrollArea>
        </DialogContent>
      </Dialog>

      {data && (
        <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
          <DialogTrigger asChild>
            <IconButton variant="outline" icon="trash" />
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>删除黑名单规则</DialogTitle>
              <DialogDescription>
                确定要删除此规则吗？此操作无法撤销。
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <DialogClose asChild>
                <Button variant="outline">取消</Button>
              </DialogClose>
              <Button variant="destructive" onClick={handleDelete}>
                删除
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </>
  )
}

export const TerminalBlacklistManager: React.FC = () => {
  const { data, error, isLoading, mutate } = useSWR(
    'terminal-blacklist',
    getTerminalBlacklist,
  )

  const getActionBadge = (action: string) => {
    switch (action) {
      case 'block':
        return <Badge variant="destructive">拦截</Badge>
      case 'warn':
        return <Badge variant="default">警告</Badge>
      case 'log':
        return <Badge variant="secondary">记录</Badge>
      default:
        return <Badge variant="outline">{action}</Badge>
    }
  }

  if (isLoading) {
    return <div className="text-center py-8">加载中...</div>
  }

  if (error) {
    return (
      <div className="text-center py-8 text-red-500">
        <div>加载数据失败</div>
        <div className="text-sm mt-2">{error.message || '未知错误'}</div>
      </div>
    )
  }

  const rules = data || []

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <div>
          <h3 className="text-lg font-medium">命令黑名单规则</h3>
          <p className="text-sm text-muted-foreground">
            管理命令匹配模式以拦截、警告或记录
          </p>
        </div>
        <BlacklistCard mutate={mutate} />
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>匹配模式</TableHead>
              <TableHead>描述</TableHead>
              <TableHead>动作</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rules.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                  暂无黑名单规则
                </TableCell>
              </TableRow>
            ) : (
              rules.map((rule) => (
                <TableRow key={rule.id}>
                  <TableCell>
                    <code className="text-sm bg-muted px-2 py-1 rounded">{rule.pattern}</code>
                  </TableCell>
                  <TableCell className="max-w-md">
                    <div className="truncate" title={rule.description}>
                      {rule.description}
                    </div>
                  </TableCell>
                  <TableCell>{getActionBadge(rule.action)}</TableCell>
                  <TableCell>
                    {rule.enabled ? (
                      <Badge variant="default">已启用</Badge>
                    ) : (
                      <Badge variant="outline">已禁用</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2">
                      <BlacklistCard data={rule} mutate={mutate} />
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
