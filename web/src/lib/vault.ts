import { api } from "@/lib/api"

export type Account = {
  id: string
  name: string
  description: string
  url: string
  access_group_names: string[]
  created_by_entity_id: string
  updated_by_entity_id: string
  deleted_at: string | null
  created_at: string
  updated_at: string
}

export type Secret = {
  id: string
  account_id: string
  key: string
  label: string
  type: string
  sensitive: boolean
  plain_value: string
  key_id: string
  algorithm: string
  created_by_entity_id: string
  updated_by_entity_id: string
  deleted_at: string | null
  created_at: string
  updated_at: string
}

export type AccountWithSecrets = Account & {
  secrets: Secret[]
}

export type AccountListItem = Account & {
  secret_count: number
}

export type AccountInput = {
  name: string
  description: string
  url: string
  access_group_names: string[]
}

export type SentinelGroup = {
  id: string
  name: string
  description: string
  allowed_sources: string[]
  created_by: string
  member_count: number
  owner_count: number
  pending_count: number
  updated_at: string
  created_at: string
}

export type SecretInput = {
  key: string
  label: string
  type: string
  sensitive: boolean
  plain_value: string
}

export const commonSecretTypes = [
  "username",
  "email",
  "password",
  "secret_key",
  "api_key",
  "client_id",
  "client_secret",
  "totp_seed",
  "url",
  "note",
]

export async function listAccounts() {
  const response = await api.get<AccountListItem[]>("/accounts")
  return response.data
}

export async function getAccount(id: string) {
  const response = await api.get<AccountWithSecrets>(`/accounts/${id}`)
  return response.data
}

export async function createAccount(input: AccountInput) {
  const response = await api.post<Account>("/accounts", input)
  return response.data
}

export async function listSentinelGroups() {
  const response = await api.get<SentinelGroup[]>("/groups")
  return response.data
}

export async function updateAccount(id: string, input: AccountInput) {
  const response = await api.put<Account>(`/accounts/${id}`, input)
  return response.data
}

export async function deleteAccount(id: string) {
  await api.delete(`/accounts/${id}`)
}

export async function createSecret(accountID: string, input: SecretInput) {
  const response = await api.post<Secret>(`/accounts/${accountID}/secrets`, input)
  return response.data
}

export async function deleteSecret(accountID: string, secretID: string) {
  await api.delete(`/accounts/${accountID}/secrets/${secretID}`)
}

export async function revealSecret(accountID: string, secretID: string) {
  const response = await api.post<{ value: string }>(
    `/accounts/${accountID}/secrets/${secretID}/reveal`,
  )
  return response.data.value
}
