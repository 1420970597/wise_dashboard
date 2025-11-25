import { swrFetcher } from "@/api/api"
import { AutoSSH, AutoSSHForm, createAutoSSH, deleteAutoSSH, startAutoSSH, stopAutoSSH, updateAutoSSH } from "@/api/autossh"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table"
import { useServer } from "@/hooks/useServer"
import { ColumnDef, flexRender, getCoreRowModel, useReactTable } from "@tanstack/react-table"
import { Play, Square, Trash2 } from "lucide-react"
import { useCallback, useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import useSWR from "swr"

export default function AutoSSHPage() {
    const { t } = useTranslation()
    const { data, mutate, error, isLoading } = useSWR<AutoSSH[]>("/api/v1/autossh", swrFetcher)
    const { servers } = useServer()

    const [dialogOpen, setDialogOpen] = useState(false)
    const [editingMapping, setEditingMapping] = useState<AutoSSH | null>(null)
    const [formData, setFormData] = useState<AutoSSHForm>({
        name: "",
        server_id: 0,
        mapping_type: "remote",
        source_port: 0,
        target_host: "localhost",
        target_port: 0,
        enabled: true,
    })

    useEffect(() => {
        if (error)
            toast(t("Error"), {
                description: t("Results.ErrorFetchingResource", {
                    error: error.message,
                }),
            })
    }, [error, t])

    const handleCreate = useCallback(() => {
        setEditingMapping(null)
        setFormData({
            name: "",
            server_id: 0,
            mapping_type: "remote",
            source_port: 0,
            target_host: "localhost",
            target_port: 0,
            enabled: true,
        })
        setDialogOpen(true)
    }, [])

    const handleEdit = useCallback((mapping: AutoSSH) => {
        setEditingMapping(mapping)
        setFormData({
            name: mapping.name,
            server_id: mapping.server_id,
            mapping_type: mapping.mapping_type,
            source_port: mapping.source_port,
            target_host: mapping.target_host,
            target_port: mapping.target_port,
            enabled: mapping.enabled,
        })
        setDialogOpen(true)
    }, [])

    const handleSubmit = useCallback(async () => {
        try {
            if (editingMapping) {
                await updateAutoSSH(editingMapping.id, formData)
                toast(t("Success"), { description: t("AutoSSH mapping updated") })
            } else {
                await createAutoSSH(formData)
                toast(t("Success"), { description: t("AutoSSH mapping created") })
            }
            setDialogOpen(false)
            mutate()
        } catch (error: any) {
            toast(t("Error"), { description: error.message })
        }
    }, [editingMapping, formData, mutate, t])

    const handleStart = useCallback(async (id: number) => {
        try {
            await startAutoSSH(id)
            toast(t("Success"), { description: t("AutoSSH mapping started") })
            mutate()
        } catch (error: any) {
            toast(t("Error"), { description: error.message })
        }
    }, [mutate, t])

    const handleStop = useCallback(async (id: number) => {
        try {
            await stopAutoSSH(id)
            toast(t("Success"), { description: t("AutoSSH mapping stopped") })
            mutate()
        } catch (error: any) {
            toast(t("Error"), { description: error.message })
        }
    }, [mutate, t])

    const columns: ColumnDef<AutoSSH>[] = useMemo(() => [
        {
            id: "select",
            header: ({ table }) => (
                <Checkbox
                    checked={
                        table.getIsAllPageRowsSelected() ||
                        (table.getIsSomePageRowsSelected() && "indeterminate")
                    }
                    onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
                    aria-label="Select all"
                />
            ),
            cell: ({ row }) => (
                <Checkbox
                    checked={row.getIsSelected()}
                    onCheckedChange={(value) => row.toggleSelected(!!value)}
                    aria-label="Select row"
                />
            ),
            enableSorting: false,
            enableHiding: false,
        },
        {
            header: "ID",
            accessorKey: "id",
            accessorFn: (row) => row.id,
        },
        {
            header: t("Name"),
            accessorKey: "name",
            accessorFn: (row) => row.name,
            cell: ({ row }) => {
                const mapping = row.original
                return <div className="max-w-32 whitespace-normal break-words">{mapping.name}</div>
            },
        },
        {
            header: t("Server"),
            accessorKey: "server_id",
            accessorFn: (row) => row.server_id,
            cell: ({ row }) => {
                const server = servers?.find((s: any) => s.id === row.original.server_id)
                return <div className="max-w-32 whitespace-normal break-words">{server?.name || `ID: ${row.original.server_id}`}</div>
            },
        },
        {
            header: t("MappingType"),
            accessorKey: "mapping_type",
            accessorFn: (row) => row.mapping_type,
            cell: ({ row }) => {
                const type = row.original.mapping_type
                return (
                    <div>
                        {type === "remote" ? t("RemoteForward") : t("LocalForward")}
                    </div>
                )
            },
        },
        {
            header: t("SourcePort"),
            accessorKey: "source_port",
            accessorFn: (row) => row.source_port,
        },
        {
            header: t("Target"),
            accessorKey: "target",
            cell: ({ row }) => {
                const mapping = row.original
                return (
                    <div className="max-w-32 whitespace-normal break-words">
                        {mapping.target_host}:{mapping.target_port}
                    </div>
                )
            },
        },
        {
            header: t("Status"),
            accessorKey: "status",
            accessorFn: (row) => row.status,
            cell: ({ row }) => {
                const status = row.original.status
                const statusColor =
                    status === "running"
                        ? "text-green-500"
                        : status === "error"
                          ? "text-red-500"
                          : "text-gray-500"
                return <div className={statusColor}>{t(status)}</div>
            },
        },
        {
            header: t("Enabled"),
            accessorKey: "enabled",
            accessorFn: (row) => row.enabled,
            cell: ({ row }) => {
                return <div>{row.original.enabled ? t("Yes") : t("No")}</div>
            },
        },
        {
            header: t("Actions"),
            id: "actions",
            cell: ({ row }) => {
                const mapping = row.original
                return (
                    <div className="flex gap-2">
                        {mapping.status === "running" ? (
                            <Button
                                size="sm"
                                variant="outline"
                                onClick={() => handleStop(mapping.id)}
                            >
                                <Square className="h-4 w-4" />
                            </Button>
                        ) : (
                            <Button
                                size="sm"
                                variant="outline"
                                onClick={() => handleStart(mapping.id)}
                            >
                                <Play className="h-4 w-4" />
                            </Button>
                        )}
                        <Button size="sm" variant="outline" onClick={() => handleEdit(mapping)}>
                            {t("Edit")}
                        </Button>
                    </div>
                )
            },
        },
    ], [t, servers, handleStart, handleStop, handleEdit])

    const tableData = useMemo(() => data || [], [data])

    const table = useReactTable({
        data: tableData,
        columns,
        getCoreRowModel: getCoreRowModel(),
    })

    return (
        <div className="mx-auto w-full max-w-5xl px-4 py-4">
            <div className="mb-4 flex items-center justify-between">
                <h1 className="text-2xl font-bold">{t("AutoSSH")}</h1>
                <div>
                    <Button onClick={handleCreate}>{t("Add")}</Button>
                </div>
            </div>

            <div className="rounded-lg border">
                <Table>
                    <TableHeader>
                        {table.getHeaderGroups().map((headerGroup) => (
                            <TableRow key={headerGroup.id}>
                                {headerGroup.headers.map((header) => (
                                    <TableHead key={header.id}>
                                        {header.isPlaceholder
                                            ? null
                                            : flexRender(
                                                  header.column.columnDef.header,
                                                  header.getContext(),
                                              )}
                                    </TableHead>
                                ))}
                            </TableRow>
                        ))}
                    </TableHeader>
                    <TableBody>
                        {isLoading ? (
                            <TableRow>
                                <TableCell colSpan={columns.length} className="text-center">
                                    {t("Loading")}...
                                </TableCell>
                            </TableRow>
                        ) : table.getRowModel().rows?.length ? (
                            table.getRowModel().rows.map((row) => (
                                <TableRow key={row.id}>
                                    {row.getVisibleCells().map((cell) => (
                                        <TableCell key={cell.id}>
                                            {flexRender(
                                                cell.column.columnDef.cell,
                                                cell.getContext(),
                                            )}
                                        </TableCell>
                                    ))}
                                </TableRow>
                            ))
                        ) : (
                            <TableRow>
                                <TableCell colSpan={columns.length} className="text-center">
                                    {t("NoData")}
                                </TableCell>
                            </TableRow>
                        )}
                    </TableBody>
                </Table>
            </div>

            {table.getSelectedRowModel().rows.length > 0 && (
                <div className="mt-4 flex gap-2">
                    <Button
                        variant="destructive"
                        onClick={async () => {
                            const ids = table.getSelectedRowModel().rows.map((row) => row.original.id)
                            try {
                                await deleteAutoSSH(ids)
                                toast(t("Success"), { description: t("Deleted successfully") })
                                mutate()
                                table.resetRowSelection()
                            } catch (error: any) {
                                toast(t("Error"), { description: error.message })
                            }
                        }}
                    >
                        <Trash2 className="h-4 w-4 mr-2" />
                        {t("Delete")} ({table.getSelectedRowModel().rows.length})
                    </Button>
                </div>
            )}

            <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
                <DialogContent className="max-w-2xl">
                    <DialogHeader>
                        <DialogTitle>
                            {editingMapping ? t("EditAutoSSH") : t("AddAutoSSH")}
                        </DialogTitle>
                        <DialogDescription>
                            {t("ConfigureAutoSSHPortMapping")}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid gap-2">
                            <Label htmlFor="name">{t("Name")}</Label>
                            <Input
                                id="name"
                                value={formData.name}
                                onChange={(e) =>
                                    setFormData({ ...formData, name: e.target.value })
                                }
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="server">{t("Server")}</Label>
                            <Select
                                value={formData.server_id.toString()}
                                onValueChange={(value) =>
                                    setFormData({ ...formData, server_id: parseInt(value) })
                                }
                            >
                                <SelectTrigger>
                                    <SelectValue placeholder={t("SelectServer")} />
                                </SelectTrigger>
                                <SelectContent>
                                    {servers?.map((server: any) => (
                                        <SelectItem key={server.id} value={server.id.toString()}>
                                            {server.name}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="mapping_type">{t("MappingType")}</Label>
                            <Select
                                value={formData.mapping_type}
                                onValueChange={(value) =>
                                    setFormData({ ...formData, mapping_type: value })
                                }
                            >
                                <SelectTrigger>
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="remote">{t("RemoteForward")}</SelectItem>
                                    <SelectItem value="local">{t("LocalForward")}</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="source_port">{t("SourcePort")}</Label>
                            <Input
                                id="source_port"
                                type="number"
                                value={formData.source_port}
                                onChange={(e) =>
                                    setFormData({
                                        ...formData,
                                        source_port: parseInt(e.target.value),
                                    })
                                }
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="target_host">{t("TargetHost")}</Label>
                            <Input
                                id="target_host"
                                value={formData.target_host}
                                onChange={(e) =>
                                    setFormData({ ...formData, target_host: e.target.value })
                                }
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="target_port">{t("TargetPort")}</Label>
                            <Input
                                id="target_port"
                                type="number"
                                value={formData.target_port}
                                onChange={(e) =>
                                    setFormData({
                                        ...formData,
                                        target_port: parseInt(e.target.value),
                                    })
                                }
                            />
                        </div>
                        <div className="flex items-center gap-2">
                            <Switch
                                id="enabled"
                                checked={formData.enabled}
                                onCheckedChange={(checked) =>
                                    setFormData({ ...formData, enabled: checked })
                                }
                            />
                            <Label htmlFor="enabled">{t("Enabled")}</Label>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setDialogOpen(false)}>
                            {t("Cancel")}
                        </Button>
                        <Button onClick={handleSubmit}>{t("Save")}</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    )
}
