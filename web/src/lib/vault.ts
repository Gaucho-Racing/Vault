import { api } from "@/lib/api"

export type Account = {
  id: string
  name: string
  description: string
  url: string
  access_group_names: string[]
  created_by_entity_id: string
  updated_by_entity_id: string
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
  created_at: string
  updated_at: string
}

export type TOTPCode = {
  code: string
  period: number
  digits: number
  algorithm: string
  seconds_remaining: number
  expires_at: string
}

export type TOTPRegistration = {
  value: string
  issuer: string
  account_name: string
  suggested_key: string
  suggested_label: string
}

export type AuditActor = {
  user_id: string
  entity_id: string
  username: string
  first_name: string
  last_name: string
  email: string
  avatar_url: string
}

export type AuditLog = {
  id: string
  action: string
  actor_entity_id: string
  actor_user_id: string
  actor_group_names: string[]
  account_id: string
  account_name: string
  secret_id: string
  secret_key: string
  secret_label: string
  request_method: string
  request_path: string
  ip_address: string
  user_agent: string
  created_at: string
  actor?: AuditActor
}

export type AppSecretApplication = {
  id: string
  name: string
  access_group_names: string[]
  created_by_entity_id: string
  updated_by_entity_id: string
  created_at: string
  updated_at: string
}

export type AccountWithSecrets = Account & {
  secrets: Secret[]
}

export type AppSecret = {
  id: string
  application_id: string
  key: string
  key_id: string
  algorithm: string
  created_by_entity_id: string
  updated_by_entity_id: string
  created_at: string
  updated_at: string
}

export type AppSecretApplicationWithSecrets = AppSecretApplication & {
  secrets: AppSecret[]
}

export type AccountListItem = Account & {
  secret_count: number
  can_access: boolean
}

export type AppSecretApplicationListItem = AppSecretApplication & {
  secret_count: number
  can_access: boolean
}

export type AccountInput = {
  name: string
  description: string
  url: string
  access_group_names: string[]
}

export type AppSecretApplicationInput = {
  name: string
  access_group_names: string[]
}

export type GitHubActionsRule = {
  id: string
  name: string
  repository_patterns: string[]
  ref_patterns: string[]
  secret_selectors: string[]
  enabled: boolean
  created_by_entity_id: string
  updated_by_entity_id: string
  created_at: string
  updated_at: string
}

export type GitHubActionsRuleInput = {
  name: string
  repository_patterns: string[]
  ref_patterns: string[]
  secret_selectors: string[]
  enabled: boolean
}

export type KubernetesCluster = {
  id: string
  name: string
  issuer: string
  audience: string
  enabled: boolean
  created_by_entity_id: string
  updated_by_entity_id: string
  created_at: string
  updated_at: string
}

export type KubernetesClusterInput = {
  name: string
  issuer: string
  audience: string
  enabled: boolean
}

export type KubernetesSecretRule = {
  id: string
  name: string
  cluster_ids: string[]
  namespace_patterns: string[]
  service_account_patterns: string[]
  secret_selectors: string[]
  enabled: boolean
  created_by_entity_id: string
  updated_by_entity_id: string
  created_at: string
  updated_at: string
}

export type KubernetesSecretRuleInput = {
  name: string
  cluster_ids: string[]
  namespace_patterns: string[]
  service_account_patterns: string[]
  secret_selectors: string[]
  enabled: boolean
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

export type AppSecretInput = {
  value: string
  key: string
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

type AppSecretApplicationResponseFields = {
  access_group_names?: string[] | null
}

type GitHubActionsRuleResponseFields = {
  repository_patterns?: string[] | null
  ref_patterns?: string[] | null
  secret_selectors?: string[] | null
}

type KubernetesSecretRuleResponseFields = {
  cluster_ids?: string[] | null
  namespace_patterns?: string[] | null
  service_account_patterns?: string[] | null
  secret_selectors?: string[] | null
}

function normalizeStringArray(value: string[] | null | undefined) {
  return Array.isArray(value) ? value : []
}

function normalizeAppSecretApplication<T extends AppSecretApplicationResponseFields>(application: T) {
  return {
    ...application,
    access_group_names: normalizeStringArray(application.access_group_names),
  }
}

function normalizeGitHubActionsRule<T extends GitHubActionsRuleResponseFields>(rule: T) {
  return {
    ...rule,
    repository_patterns: normalizeStringArray(rule.repository_patterns),
    ref_patterns: normalizeStringArray(rule.ref_patterns),
    secret_selectors: normalizeStringArray(rule.secret_selectors),
  }
}

function normalizeKubernetesSecretRule<T extends KubernetesSecretRuleResponseFields>(rule: T) {
  return {
    ...rule,
    cluster_ids: normalizeStringArray(rule.cluster_ids),
    namespace_patterns: normalizeStringArray(rule.namespace_patterns),
    service_account_patterns: normalizeStringArray(rule.service_account_patterns),
    secret_selectors: normalizeStringArray(rule.secret_selectors),
  }
}

export async function listAccounts() {
  const response = await api.get<AccountListItem[]>("/accounts")
  return response.data
}

export async function listAppSecretApplications() {
  const response = await api.get<AppSecretApplicationListItem[]>("/app-secrets")
  return response.data.map(normalizeAppSecretApplication)
}

export async function createAppSecretApplication(input: AppSecretApplicationInput) {
  const response = await api.post<AppSecretApplication>("/app-secrets", input)
  return normalizeAppSecretApplication(response.data)
}

export async function getAppSecretApplication(id: string) {
  const response = await api.get<AppSecretApplicationWithSecrets>(`/app-secrets/${id}`)
  return normalizeAppSecretApplication(response.data)
}

export async function updateAppSecretApplication(id: string, input: AppSecretApplicationInput) {
  const response = await api.put<AppSecretApplication>(`/app-secrets/${id}`, input)
  return normalizeAppSecretApplication(response.data)
}

export async function deleteAppSecretApplication(id: string) {
  await api.delete(`/app-secrets/${id}`)
}

export async function listGitHubActionsRules() {
  const response = await api.get<GitHubActionsRule[]>("/integrations/github/actions/rules")
  return response.data.map(normalizeGitHubActionsRule)
}

export async function createGitHubActionsRule(input: GitHubActionsRuleInput) {
  const response = await api.post<GitHubActionsRule>("/integrations/github/actions/rules", input)
  return normalizeGitHubActionsRule(response.data)
}

export async function updateGitHubActionsRule(id: string, input: GitHubActionsRuleInput) {
  const response = await api.put<GitHubActionsRule>(`/integrations/github/actions/rules/${id}`, input)
  return normalizeGitHubActionsRule(response.data)
}

export async function deleteGitHubActionsRule(id: string) {
  await api.delete(`/integrations/github/actions/rules/${id}`)
}

export async function listKubernetesClusters() {
  const response = await api.get<KubernetesCluster[]>("/integrations/kubernetes/clusters")
  return response.data
}

export async function createKubernetesCluster(input: KubernetesClusterInput) {
  const response = await api.post<KubernetesCluster>("/integrations/kubernetes/clusters", input)
  return response.data
}

export async function updateKubernetesCluster(id: string, input: KubernetesClusterInput) {
  const response = await api.put<KubernetesCluster>(`/integrations/kubernetes/clusters/${id}`, input)
  return response.data
}

export async function verifyKubernetesCluster(input: KubernetesClusterInput) {
  const response = await api.post<{ valid: boolean }>("/integrations/kubernetes/clusters/verify", input)
  return response.data
}

export async function deleteKubernetesCluster(id: string) {
  await api.delete(`/integrations/kubernetes/clusters/${id}`)
}

export async function listKubernetesSecretRules() {
  const response = await api.get<KubernetesSecretRule[]>("/integrations/kubernetes/rules")
  return response.data.map(normalizeKubernetesSecretRule)
}

export async function createKubernetesSecretRule(input: KubernetesSecretRuleInput) {
  const response = await api.post<KubernetesSecretRule>("/integrations/kubernetes/rules", input)
  return normalizeKubernetesSecretRule(response.data)
}

export async function updateKubernetesSecretRule(id: string, input: KubernetesSecretRuleInput) {
  const response = await api.put<KubernetesSecretRule>(`/integrations/kubernetes/rules/${id}`, input)
  return normalizeKubernetesSecretRule(response.data)
}

export async function deleteKubernetesSecretRule(id: string) {
  await api.delete(`/integrations/kubernetes/rules/${id}`)
}

export async function createAppSecret(applicationID: string, input: AppSecretInput) {
  const response = await api.post<AppSecret>(`/app-secrets/${applicationID}/secrets`, input)
  return response.data
}

export async function updateAppSecret(applicationID: string, secretID: string, input: AppSecretInput) {
  const response = await api.put<AppSecret>(`/app-secrets/${applicationID}/secrets/${secretID}`, input)
  return response.data
}

export async function deleteAppSecret(applicationID: string, secretID: string) {
  await api.delete(`/app-secrets/${applicationID}/secrets/${secretID}`)
}

export async function revealAppSecret(applicationID: string, secretID: string) {
  const response = await api.post<{ value: string }>(
    `/app-secrets/${applicationID}/secrets/${secretID}/reveal`,
  )
  return response.data.value
}

export async function downloadAppSecretEnvFile(applicationID: string) {
  const response = await api.get<Blob>(`/app-secrets/${applicationID}/env`, {
    responseType: "blob",
  })
  return response.data
}

export async function getAccount(id: string) {
  const response = await api.get<AccountWithSecrets>(`/accounts/${id}`)
  return response.data
}

export async function listAccountAuditLogs(id: string, limit = 10) {
  const response = await api.get<AuditLog[]>(`/accounts/${id}/audit-logs`, {
    params: { limit },
  })
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

export async function generateTOTPCode(accountID: string, secretID: string) {
  const response = await api.post<TOTPCode>(`/accounts/${accountID}/secrets/${secretID}/totp`)
  return response.data
}

export async function decodeTOTPRegistrationQRCode(file: File) {
  const formData = new FormData()
  formData.append("file", file, file.name || "totp-qr.png")
  const response = await api.post<TOTPRegistration>("/secrets/totp/qr", formData)
  return response.data
}
