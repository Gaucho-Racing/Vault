import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  ArrowLeft,
  Clock3,
  Copy,
  Eye,
  EyeOff,
  ExternalLink,
  History,
  Plus,
  ShieldCheck,
  Trash2,
  UsersRound,
} from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { Link, useNavigate, useParams } from "react-router-dom"
import { toast } from "sonner"

import { AccountFormDialog } from "@/components/AccountFormDialog"
import { ConfirmDialog } from "@/components/ConfirmDialog"
import { PageContainer } from "@/components/PageContainer"
import { SecretFormDialog } from "@/components/SecretFormDialog"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useAuth } from "@/lib/auth"
import {
  createSecret,
  deleteAccount,
  deleteSecret,
  generateTOTPCode,
  getAccount,
  listAccountAuditLogs,
  revealSecret,
  updateAccount,
  type AuditLog,
  type AccountInput,
  type Secret,
  type SecretInput,
  type TOTPCode,
} from "@/lib/vault"

const TOTP_SECRET_TYPE = "totp_seed"
const AUDIT_LOG_PAGE_SIZE = 10

function errorMessage(error: unknown, fallback: string) {
  return (error as { response?: { data?: { error?: string } } })?.response?.data?.error ?? fallback
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  })
}

function formatDateTime(value: string) {
  return new Date(value).toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
  })
}

async function copyValue(value: string) {
  try {
    await navigator.clipboard.writeText(value)
    toast.success("Copied")
  } catch {
    toast.error("Failed to copy")
  }
}

function isTOTPSecret(secret: Secret) {
  return secret.type.trim().toLowerCase() === TOTP_SECRET_TYPE
}

function formatTOTPCode(code: string) {
  if (code.length === 6) return `${code.slice(0, 3)} ${code.slice(3)}`
  if (code.length === 8) return `${code.slice(0, 4)} ${code.slice(4)}`
  return code
}

function auditActionLabel(action: string) {
  switch (action) {
    case "account.created":
      return "Created account"
    case "account.updated":
      return "Updated account"
    case "account.deleted":
      return "Deleted account"
    case "account.viewed":
      return "Viewed account"
    case "secret.viewed":
      return "Viewed secret"
    default:
      return action
  }
}

function auditActorName(auditLog: AuditLog) {
  const actor = auditLog.actor
  if (!actor) return auditLog.actor_user_id || auditLog.actor_entity_id || "Unknown actor"
  return (
    [actor.first_name, actor.last_name].filter(Boolean).join(" ") ||
    actor.username ||
    actor.email ||
    actor.user_id ||
    auditLog.actor_entity_id ||
    "Unknown actor"
  )
}

function auditActorInitials(auditLog: AuditLog) {
  const actor = auditLog.actor
  const source =
    [actor?.first_name, actor?.last_name].filter(Boolean).join(" ") ||
    actor?.username ||
    actor?.email ||
    auditLog.actor_user_id ||
    auditLog.actor_entity_id
  const initials = source
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join("")
  return initials || "?"
}

function auditTargetLabel(auditLog: AuditLog) {
  if (auditLog.secret_id) {
    return auditLog.secret_label || auditLog.secret_key || auditLog.secret_id
  }
  return auditLog.account_name || auditLog.account_id
}

function auditTargetSecondary(auditLog: AuditLog) {
  if (auditLog.secret_id) {
    return auditLog.secret_key || auditLog.secret_id
  }
  return ""
}

function TOTPSecretValue({ accountID, secret }: { accountID: string; secret: Secret }) {
  const [code, setCode] = useState<TOTPCode | null>(null)
  const [nowMs, setNowMs] = useState(() => Date.now())
  const refreshAttemptedFor = useRef<string | null>(null)
  const totpMutation = useMutation({
    mutationFn: () => generateTOTPCode(accountID, secret.id),
    onSuccess: (nextCode) => {
      setCode(nextCode)
      setNowMs(Date.now())
      refreshAttemptedFor.current = null
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to generate TOTP code")),
  })
  const secondsRemaining = code
    ? Math.max(0, Math.ceil((new Date(code.expires_at).getTime() - nowMs) / 1000))
    : 0

  useEffect(() => {
    if (!code) return
    const intervalID = window.setInterval(() => {
      setNowMs(Date.now())
    }, 1000)
    return () => window.clearInterval(intervalID)
  }, [code])

  useEffect(() => {
    if (
      !code ||
      totpMutation.isPending ||
      secondsRemaining > 1 ||
      refreshAttemptedFor.current === code.expires_at
    ) {
      return
    }
    refreshAttemptedFor.current = code.expires_at
    totpMutation.mutate()
  }, [code, secondsRemaining, totpMutation])

  if (!code) {
    return (
      <Button
        variant="secondary"
        size="sm"
        disabled={totpMutation.isPending}
        onClick={() => totpMutation.mutate()}
      >
        <Clock3 className="size-3.5" />
        {totpMutation.isPending ? "Generating" : "Generate code"}
      </Button>
    )
  }

  return (
    <div className="flex min-w-0 flex-wrap items-center gap-2">
      <code className="min-w-24 rounded-md bg-gr-pink/10 px-2.5 py-1.5 text-center font-mono text-sm font-semibold text-gr-pink dark:bg-gr-pink/20">
        {formatTOTPCode(code.code)}
      </code>
      <span className="rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground">
        {secondsRemaining}s
      </span>
      <Button variant="ghost" size="icon-sm" onClick={() => void copyValue(code.code)}>
        <Copy className="size-3.5" />
        <span className="sr-only">Copy TOTP code</span>
      </Button>
    </div>
  )
}

function SecretValue({
  accountID,
  secret,
  revealed,
  revealing,
  copying,
  onReveal,
  onCopy,
  onHide,
}: {
  accountID: string
  secret: Secret
  revealed?: string
  revealing: boolean
  copying: boolean
  onReveal: (accountID: string, secretID: string) => void
  onCopy: (accountID: string, secretID: string) => void
  onHide: (secretID: string) => void
}) {
  if (isTOTPSecret(secret)) {
    return <TOTPSecretValue accountID={accountID} secret={secret} />
  }

  if (!secret.sensitive) {
    return (
      <div className="flex min-w-0 items-center gap-2">
        <code className="min-w-0 truncate rounded-md bg-muted px-2.5 py-1.5 text-xs">
          {secret.plain_value || "empty"}
        </code>
        {secret.plain_value && (
          <Button variant="ghost" size="icon-sm" onClick={() => void copyValue(secret.plain_value)}>
            <Copy className="size-3.5" />
            <span className="sr-only">Copy</span>
          </Button>
        )}
      </div>
    )
  }

  if (revealed !== undefined) {
    return (
      <div className="flex min-w-0 items-center gap-2">
        <code className="min-w-0 truncate rounded-md bg-muted px-2.5 py-1.5 text-xs">
          {revealed}
        </code>
        <Button variant="ghost" size="icon-sm" onClick={() => void copyValue(revealed)}>
          <Copy className="size-3.5" />
          <span className="sr-only">Copy</span>
        </Button>
        <Button variant="ghost" size="icon-sm" onClick={() => onHide(secret.id)}>
          <EyeOff className="size-3.5" />
          <span className="sr-only">Hide</span>
        </Button>
      </div>
    )
  }

  return (
    <div className="flex flex-wrap gap-2">
      <Button
        variant="secondary"
        size="sm"
        disabled={copying}
        onClick={() => onCopy(accountID, secret.id)}
      >
        <Copy className="size-3.5" />
        {copying ? "Copying" : "Copy"}
      </Button>
      <Button
        variant="secondary"
        size="sm"
        disabled={revealing}
        onClick={() => onReveal(accountID, secret.id)}
      >
        <Eye className="size-3.5" />
        {revealing ? "Revealing" : "Reveal"}
      </Button>
    </div>
  )
}

export default function AccountDetailsPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { user } = useAuth()
  const [revealed, setRevealed] = useState<Record<string, string>>({})
  const [auditLogLimit, setAuditLogLimit] = useState(AUDIT_LOG_PAGE_SIZE)
  const canViewAuditLog = user?.groups.includes("Admins") ?? false

  const accountQuery = useQuery({
    queryKey: ["account", id],
    queryFn: () => getAccount(id ?? ""),
    enabled: !!id,
  })

  const auditLogsQuery = useQuery({
    queryKey: ["accountAuditLogs", id, auditLogLimit],
    queryFn: () => listAccountAuditLogs(id ?? "", auditLogLimit),
    enabled: !!id && canViewAuditLog,
  })

  const createSecretMutation = useMutation({
    mutationFn: (input: SecretInput) => createSecret(id ?? "", input),
    onSuccess: () => {
      toast.success("Secret created")
      void queryClient.invalidateQueries({ queryKey: ["account", id] })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to create secret")),
  })

  const updateAccountMutation = useMutation({
    mutationFn: (input: AccountInput) => updateAccount(id ?? "", input),
    onSuccess: () => {
      toast.success("Account updated")
      void queryClient.invalidateQueries({ queryKey: ["account", id] })
      void queryClient.invalidateQueries({ queryKey: ["accountAuditLogs", id] })
      void queryClient.invalidateQueries({ queryKey: ["accounts"] })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to update account")),
  })

  const deleteAccountMutation = useMutation({
    mutationFn: () => deleteAccount(id ?? ""),
    onSuccess: () => {
      toast.success("Account deleted")
      void queryClient.invalidateQueries({ queryKey: ["accountAuditLogs", id] })
      void queryClient.invalidateQueries({ queryKey: ["accounts"] })
      navigate("/accounts", { replace: true })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to delete account")),
  })

  const deleteSecretMutation = useMutation({
    mutationFn: (secretID: string) => deleteSecret(id ?? "", secretID),
    onSuccess: () => {
      toast.success("Secret deleted")
      void queryClient.invalidateQueries({ queryKey: ["account", id] })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to delete secret")),
  })

  const revealMutation = useMutation({
    mutationFn: ({ accountID, secretID }: { accountID: string; secretID: string }) =>
      revealSecret(accountID, secretID),
    onSuccess: (value, variables) => {
      setRevealed((current) => ({ ...current, [variables.secretID]: value }))
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to reveal secret")),
  })

  const copySensitiveSecretMutation = useMutation({
    mutationFn: async ({ accountID, secretID }: { accountID: string; secretID: string }) => {
      const value = await revealSecret(accountID, secretID)
      await navigator.clipboard.writeText(value)
    },
    onSuccess: () => toast.success("Copied"),
    onError: (error) => toast.error(errorMessage(error, "Failed to copy secret")),
  })

  const account = accountQuery.data

  if (accountQuery.isLoading) {
    return (
      <PageContainer>
        <Skeleton className="mb-4 h-8 w-48" />
        <Skeleton className="h-32 rounded-lg" />
        <Skeleton className="mt-4 h-96 rounded-lg" />
      </PageContainer>
    )
  }

  if (!account) {
    return (
      <PageContainer>
        <Card>
          <CardContent className="py-10 text-center text-sm text-muted-foreground">
            Account not found.
          </CardContent>
        </Card>
      </PageContainer>
    )
  }

  async function handleUpdate(input: AccountInput) {
    await updateAccountMutation.mutateAsync(input)
  }

  async function handleCreateSecret(input: SecretInput) {
    await createSecretMutation.mutateAsync(input)
  }

  return (
    <PageContainer>
      <div className="mb-6">
        <div className="mb-4 flex items-center justify-between gap-3">
          <Button asChild variant="ghost" size="sm" className="-ml-2">
            <Link to="/accounts">
              <ArrowLeft className="size-4" />
              Accounts
            </Link>
          </Button>
          <div className="flex flex-wrap justify-end gap-2">
            {account.url && (
              <Button asChild variant="outline">
                <a href={account.url} target="_blank" rel="noreferrer">
                  <ExternalLink className="size-4" />
                  Open
                </a>
              </Button>
            )}
            <AccountFormDialog
              account={account}
              isPending={updateAccountMutation.isPending}
              onSubmit={handleUpdate}
              trigger={<Button variant="outline">Edit</Button>}
            />
            <ConfirmDialog
              title="Delete account"
              description="This will delete the account and its secrets from Vault."
              confirmLabel="Delete account"
              isPending={deleteAccountMutation.isPending}
              onConfirm={() => deleteAccountMutation.mutateAsync()}
              trigger={
                <Button variant="destructive" disabled={deleteAccountMutation.isPending}>
                  <Trash2 className="size-4" />
                  Delete
                </Button>
              }
            />
          </div>
        </div>

        <div className="rounded-lg bg-card p-4 shadow-sm shadow-black/[0.03] dark:shadow-black/20">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                  <ShieldCheck className="size-5" />
                </div>
                <h1 className="truncate text-2xl font-semibold">{account.name}</h1>
              </div>
              <p className="mt-3 max-w-3xl text-sm leading-6 text-muted-foreground">
                {account.description || "No description"}
              </p>
            </div>
            <div className="grid gap-2 sm:grid-cols-2 lg:w-[420px]">
              <div className="rounded-lg bg-muted/45 p-3 dark:bg-muted/35">
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <UsersRound className="size-3.5" />
                  Access
                </div>
                <div className="mt-1 truncate text-sm font-medium">
                  {account.access_group_names.length === 0
                    ? "Public"
                    : `${account.access_group_names.length} group${account.access_group_names.length === 1 ? "" : "s"}`}
                </div>
              </div>
              <div className="rounded-lg bg-muted/45 p-3 dark:bg-muted/35">
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <Clock3 className="size-3.5" />
                  Updated
                </div>
                <div className="mt-1 truncate text-sm font-medium">{formatDate(account.updated_at)}</div>
              </div>
            </div>
          </div>
          <div className="mt-4 flex flex-wrap gap-1.5 border-t border-border/50 pt-4">
            {account.access_group_names.length === 0 ? (
              <Badge variant="secondary">Public</Badge>
            ) : (
              account.access_group_names.map((group) => (
                <Badge key={group} variant="outline">
                  {group}
                </Badge>
              ))
            )}
          </div>
        </div>
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between gap-3 border-b border-border/50 pb-4">
          <div>
            <CardTitle>Secrets</CardTitle>
            <div className="mt-1 text-sm text-muted-foreground">
              {account.secrets.length} item{account.secrets.length === 1 ? "" : "s"}
            </div>
          </div>
          <SecretFormDialog
            isPending={createSecretMutation.isPending}
            onSubmit={handleCreateSecret}
            trigger={
              <Button size="sm">
                <Plus className="size-4" />
                Secret
              </Button>
            }
          />
        </CardHeader>
        <CardContent className="p-0">
          {account.secrets.length === 0 ? (
            <div className="flex min-h-56 flex-col items-center justify-center px-4 py-10 text-center">
              <div className="flex size-10 items-center justify-center rounded-lg bg-muted">
                <ShieldCheck className="size-5 text-muted-foreground" />
              </div>
              <div className="mt-4 text-sm font-medium">No secrets yet</div>
              <div className="mt-1 text-sm text-muted-foreground">
                Add the first username, password, token, or note for this account.
              </div>
            </div>
          ) : (
            <div>
              <div className="hidden grid-cols-[minmax(180px,1fr)_160px_minmax(220px,1.4fr)_44px] border-b border-border/50 px-4 py-3 text-xs font-medium text-muted-foreground lg:grid">
                <div>Secret</div>
                <div>Type</div>
                <div>Value</div>
                <div />
              </div>
              {account.secrets.map((secret) => (
                <div
                  key={secret.id}
                  className="grid gap-3 border-b border-border/45 px-4 py-4 last:border-b-0 lg:grid-cols-[minmax(180px,1fr)_160px_minmax(220px,1.4fr)_44px] lg:items-center"
                >
                  <div className="min-w-0">
                    <div className="truncate text-sm font-medium">{secret.label || secret.key}</div>
                    <div className="mt-1 truncate font-mono text-xs text-muted-foreground">
                      {secret.key}
                    </div>
                  </div>
                  <div className="flex flex-wrap gap-1.5">
                    {secret.type && <Badge variant="outline">{secret.type}</Badge>}
                    <Badge
                      variant="outline"
                      className={
                        secret.sensitive
                          ? "border-transparent bg-gr-pink/10 text-gr-pink dark:bg-gr-pink/20"
                          : undefined
                      }
                    >
                      {secret.sensitive ? "Sensitive" : "Plain"}
                    </Badge>
                  </div>
                  <SecretValue
                    accountID={account.id}
                    secret={secret}
                    revealed={revealed[secret.id]}
                    revealing={
                      revealMutation.isPending && revealMutation.variables?.secretID === secret.id
                    }
                    copying={
                      copySensitiveSecretMutation.isPending &&
                      copySensitiveSecretMutation.variables?.secretID === secret.id
                    }
                    onReveal={(accountID, secretID) => revealMutation.mutate({ accountID, secretID })}
                    onCopy={(accountID, secretID) =>
                      copySensitiveSecretMutation.mutate({ accountID, secretID })
                    }
                    onHide={(secretID) =>
                      setRevealed((current) => {
                        const next = { ...current }
                        delete next[secretID]
                        return next
                      })
                    }
                  />
                  <ConfirmDialog
                    title="Delete secret"
                    description="This will delete this secret from the account."
                    confirmLabel="Delete secret"
                    isPending={deleteSecretMutation.isPending}
                    onConfirm={() => deleteSecretMutation.mutateAsync(secret.id)}
                    trigger={
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        disabled={deleteSecretMutation.isPending}
                      >
                        <Trash2 className="size-3.5" />
                        <span className="sr-only">Delete secret</span>
                      </Button>
                    }
                  />
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {canViewAuditLog && (
        <Card className="mt-4">
          <CardHeader className="flex flex-row items-center justify-between gap-3 border-b border-border/50 pb-4">
            <div>
              <CardTitle>Audit log</CardTitle>
              <div className="mt-1 text-sm text-muted-foreground">
                Recent account activity
              </div>
            </div>
            <div className="flex size-8 items-center justify-center rounded-lg bg-muted text-muted-foreground">
              <History className="size-4" />
            </div>
          </CardHeader>
          <CardContent className="p-0">
            {auditLogsQuery.isLoading ? (
              <div className="space-y-3 p-4">
                <Skeleton className="h-12 rounded-lg" />
                <Skeleton className="h-12 rounded-lg" />
                <Skeleton className="h-12 rounded-lg" />
              </div>
            ) : auditLogsQuery.isError ? (
              <div className="px-4 py-8 text-center text-sm text-muted-foreground">
                Could not load audit log.
              </div>
            ) : (auditLogsQuery.data ?? []).length === 0 ? (
              <div className="px-4 py-8 text-center text-sm text-muted-foreground">
                No audit events yet.
              </div>
            ) : (
              <div>
                {(auditLogsQuery.data ?? []).map((auditLog) => (
                  <div
                    key={auditLog.id}
                    className="grid gap-3 border-b border-border/45 px-4 py-4 last:border-b-0 md:grid-cols-[minmax(180px,1fr)_minmax(240px,1fr)_180px] md:items-center"
                  >
                    <div className="min-w-0">
                      <div className="truncate text-sm font-medium">
                        {auditActionLabel(auditLog.action)}
                      </div>
                      <div className="mt-1 truncate font-mono text-xs text-muted-foreground">
                        {auditTargetLabel(auditLog)}
                      </div>
                      {auditTargetSecondary(auditLog) !== auditTargetLabel(auditLog) && (
                        <div className="mt-1 truncate font-mono text-xs text-muted-foreground/80">
                          {auditTargetSecondary(auditLog)}
                        </div>
                      )}
                    </div>
                    <div className="flex min-w-0 items-center gap-3">
                      <Avatar>
                        {auditLog.actor?.avatar_url && (
                          <AvatarImage src={auditLog.actor.avatar_url} alt={auditActorName(auditLog)} />
                        )}
                        <AvatarFallback>{auditActorInitials(auditLog)}</AvatarFallback>
                      </Avatar>
                      <div className="min-w-0">
                        <div className="truncate text-sm">{auditActorName(auditLog)}</div>
                        {auditLog.ip_address && (
                          <div className="mt-1 truncate text-xs text-muted-foreground">
                            {auditLog.ip_address}
                          </div>
                        )}
                      </div>
                    </div>
                    <div className="text-sm text-muted-foreground md:text-right">
                      {formatDateTime(auditLog.created_at)}
                    </div>
                  </div>
                ))}
                {(auditLogsQuery.data ?? []).length >= auditLogLimit && auditLogLimit < 100 && (
                  <div className="border-t border-border/45 px-4 py-3">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      className="w-full"
                      disabled={auditLogsQuery.isFetching}
                      onClick={() =>
                        setAuditLogLimit((current) => Math.min(current + AUDIT_LOG_PAGE_SIZE, 100))
                      }
                    >
                      {auditLogsQuery.isFetching ? "Loading" : "Load more"}
                    </Button>
                  </div>
                )}
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </PageContainer>
  )
}
