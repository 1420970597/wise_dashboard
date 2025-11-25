import { createUser, updateUser } from "@/api/user"
import { Button } from "@/components/ui/button"
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog"
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"
import { IconButton } from "@/components/xui/icon-button"
import { MultiSelect } from "@/components/xui/multi-select"
import { useServer } from "@/hooks/useServer"
import { ModelUser } from "@/types"
import { zodResolver } from "@hookform/resolvers/zod"
import { useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { KeyedMutator } from "swr"
import { z } from "zod"

interface UserCardProps {
    data?: ModelUser
    mutate: KeyedMutator<ModelUser[]>
}

const userFormSchema = z.object({
    username: z.string().min(1),
    role: z.number().int().min(0).max(1),
    password: z.string().min(8).max(72).optional().or(z.literal("")),
    server_ids: z.array(z.number()).optional(),
})

export const UserCard: React.FC<UserCardProps> = ({ data, mutate }) => {
    const { t } = useTranslation()
    const form = useForm<z.infer<typeof userFormSchema>>({
        resolver: zodResolver(userFormSchema),
        defaultValues: data
            ? {
                  username: data.username,
                  role: data.role,
                  password: "",
                  server_ids: data.server_ids || [],
              }
            : {
                  username: "",
                  role: 1,
                  password: "",
                  server_ids: [],
              },
        resetOptions: {
            keepDefaultValues: false,
        },
    })

    const [open, setOpen] = useState(false)

    const onSubmit = async (values: z.infer<typeof userFormSchema>) => {
        try {
            if (data) {
                // Edit mode - only send fields that are set
                const updateData: any = {
                    username: values.username,
                    role: values.role,
                }
                if (values.password && values.password.length > 0) {
                    updateData.password = values.password
                }
                if (values.server_ids) {
                    updateData.server_ids = values.server_ids
                }
                await updateUser(data.id, updateData)
            } else {
                // Create mode - password is required
                if (!values.password || values.password.length === 0) {
                    toast(t("Error"), {
                        description: t("PasswordRequired"),
                    })
                    return
                }
                await createUser(values)
            }
        } catch (e) {
            console.error(e)
            toast(t("Error"), {
                description: t("Results.UnExpectedError"),
            })
            return
        }
        setOpen(false)
        await mutate()
        form.reset()
    }

    const { servers } = useServer()
    const serverList = servers?.map((s) => ({
        value: `${s.id}`,
        label: s.name,
    })) || []

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                {data ? <IconButton variant="outline" icon="edit" /> : <IconButton icon="plus" />}
            </DialogTrigger>
            <DialogContent className="sm:max-w-xl">
                <ScrollArea className="max-h-[calc(100dvh-5rem)] p-3">
                    <div className="items-center mx-1">
                        <DialogHeader>
                            <DialogTitle>{data ? t("EditUser") : t("NewUser")}</DialogTitle>
                            <DialogDescription />
                        </DialogHeader>
                        <Form {...form}>
                            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-2 my-2">
                                <FormField
                                    control={form.control}
                                    name="username"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>{t("Username")}</FormLabel>
                                            <FormControl>
                                                <Input {...field} />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                <FormField
                                    control={form.control}
                                    name="password"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>
                                                {t("Password")}
                                                {data && ` (${t("LeaveBlankToKeepCurrent")})`}
                                            </FormLabel>
                                            <FormControl>
                                                <Input type="password" {...field} />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                <FormField
                                    control={form.control}
                                    name="role"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>{t("Role")}</FormLabel>
                                            <Select
                                                onValueChange={(value) =>
                                                    field.onChange(parseInt(value))
                                                }
                                                defaultValue={field.value.toString()}
                                            >
                                                <FormControl>
                                                    <SelectTrigger>
                                                        <SelectValue
                                                            placeholder={t("SelectRole")}
                                                        />
                                                    </SelectTrigger>
                                                </FormControl>
                                                <SelectContent>
                                                    <SelectItem value="0">{t("Admin")}</SelectItem>
                                                    <SelectItem value="1">{t("User")}</SelectItem>
                                                </SelectContent>
                                            </Select>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                <FormField
                                    control={form.control}
                                    name="server_ids"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>{t("ServerPermissions")}</FormLabel>
                                            <FormControl>
                                                <MultiSelect
                                                    options={serverList}
                                                    onValueChange={(e) => {
                                                        const arr = e.map(Number)
                                                        field.onChange(arr)
                                                    }}
                                                    defaultValue={field.value?.map(String)}
                                                    placeholder={t("SelectServers")}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                <DialogFooter className="justify-end">
                                    <DialogClose asChild>
                                        <Button type="button" className="my-2" variant="secondary">
                                            {t("Close")}
                                        </Button>
                                    </DialogClose>
                                    <Button type="submit" className="my-2">
                                        {t("Confirm")}
                                    </Button>
                                </DialogFooter>
                            </form>
                        </Form>
                    </div>
                </ScrollArea>
            </DialogContent>
        </Dialog>
    )
}
