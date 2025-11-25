import { useState } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { TerminalAuditSessions } from '@/components/terminal-audit'
import { TerminalBlacklistManager } from '@/components/terminal-blacklist'

export default function TerminalAuditPage() {
  const [activeTab, setActiveTab] = useState('sessions')

  return (
    <div className="px-3 max-w-7xl mx-auto">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between w-full gap-3 mt-6 mb-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">终端审计</h1>
          <p className="text-sm text-muted-foreground mt-1">
            监控终端会话并管理命令黑名单
          </p>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full max-w-md grid-cols-2">
          <TabsTrigger value="sessions">终端会话</TabsTrigger>
          <TabsTrigger value="blacklist">命令黑名单</TabsTrigger>
        </TabsList>

        <TabsContent value="sessions" className="mt-6">
          <TerminalAuditSessions />
        </TabsContent>

        <TabsContent value="blacklist" className="mt-6">
          <TerminalBlacklistManager />
        </TabsContent>
      </Tabs>
    </div>
  )
}
