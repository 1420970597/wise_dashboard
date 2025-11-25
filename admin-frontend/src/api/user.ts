import { ModelProfile, ModelProfileForm, ModelUser, ModelUserForm } from "@/types"

import { FetcherMethod, fetcher } from "./api"

export const getProfile = async (): Promise<ModelProfile> => {
    return fetcher<ModelProfile>(FetcherMethod.GET, "/api/v1/profile", null)
}

export const login = async (username: string, password: string): Promise<void> => {
    return fetcher<void>(FetcherMethod.POST, "/api/v1/login", { username, password })
}

export const createUser = async (data: ModelUserForm): Promise<number> => {
    return fetcher<number>(FetcherMethod.POST, "/api/v1/user", data)
}

export const deleteUser = async (id: number[]): Promise<void> => {
    return fetcher<void>(FetcherMethod.POST, "/api/v1/batch-delete/user", id)
}

export const updateProfile = async (data: ModelProfileForm): Promise<void> => {
    return fetcher<void>(FetcherMethod.POST, "/api/v1/profile", data)
}

export const getUser = async (id: number): Promise<ModelUser> => {
    return fetcher<ModelUser>(FetcherMethod.GET, `/api/v1/user/${id}`, null)
}

export const updateUser = async (id: number, data: ModelUserForm): Promise<void> => {
    return fetcher<void>(FetcherMethod.PATCH, `/api/v1/user/${id}`, data)
}

export const updateUserServers = async (id: number, serverIds: number[]): Promise<void> => {
    return fetcher<void>(FetcherMethod.PUT, `/api/v1/user/${id}/servers`, serverIds)
}
