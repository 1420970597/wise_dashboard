import { fetcher, FetcherMethod } from "./api"

export interface AutoSSH {
    id: number
    user_id: number
    name: string
    server_id: number
    mapping_type: string
    source_port: number
    target_host: string
    target_port: number
    enabled: boolean
    status: string
    last_error?: string
    last_start_at?: string
    created_at: string
    updated_at: string
}

export interface AutoSSHForm {
    name: string
    server_id: number
    mapping_type: string
    source_port: number
    target_host: string
    target_port: number
    enabled: boolean
}

export async function createAutoSSH(data: AutoSSHForm) {
    return fetcher<{ id: number }>(FetcherMethod.POST, "/api/v1/autossh", data)
}

export async function updateAutoSSH(id: number, data: AutoSSHForm) {
    return fetcher(FetcherMethod.PATCH, `/api/v1/autossh/${id}`, data)
}

export async function startAutoSSH(id: number) {
    return fetcher(FetcherMethod.POST, `/api/v1/autossh/${id}/start`)
}

export async function stopAutoSSH(id: number) {
    return fetcher(FetcherMethod.POST, `/api/v1/autossh/${id}/stop`)
}

export async function deleteAutoSSH(ids: number[]) {
    return fetcher(FetcherMethod.POST, "/api/v1/batch-delete/autossh", ids)
}
