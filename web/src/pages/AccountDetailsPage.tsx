import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { ArrowLeft, Copy, Eye, EyeOff, ExternalLink, Plus, Trash2 } from "lucide-react"
import { useState } from "react"
import { Link, useNavigate, useParams } from "react-router-dom"
import { toast } from "sonner"

import { AccountFormDialog } from "@/components/AccountFormDialog"
import { PageContainer } from "@/components/PageContainer"
import { SecretFormDialog } from "@/components/SecretFormDialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import {
  archiveAccount,
  archiveSecret,
  createSecret,
  getAccount,
  revealSecret,
  updateAccount,
  type AccountInput,
  type Secret,
  type SecretInput,
} from "@/lib/vault"

function errorMessage(error: unknown, fallback: string) {
  return (error as { response?: { data?: { error?: string } } })?.response?.data?.error ?? fallback
}

async function copyValue(value: string) {
  await navigator.clipboard.writeText(value)
  toast.success("Copied")
}

function SecretValue({
  accountID,
  secret,
  revealed,
  revealing,
  onReveal,
  onHide,
}: {
  accountID: string
  secret: Secret
  revealed?: string
  revealing: boolean
  onReveal: (accountID: string, secretID: string) => void
  onHide: (secretID: string) => void
}) {
  if (!secret.sensitive) {
    return (
      <div className="flex min-w-0 items-center gap-2">
        <code className="min-w-0 truncate rounded bg-muted px-2 py-1 text-xs">
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
        <code className="min-w-0 truncate rounded bg-muted px-2 py-1 text-xs">{revealed}</code>
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
    <Button
      variant="outline"
      size="sm"
      disabled={revealing}
      onClick={() => onReveal(accountID, secret.id)}
    >
      <Eye className="size-3.5" />
      {revealing ? "Revealing" : "Reveal"}
    </Button>
  )
}

export default function AccountDetailsPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [revealed, setRevealed] = useState<Record<string, string>>({})

  const accountQuery = useQuery({
    queryKey: ["account", id],
    queryFn: () => getAccount(id ?? ""),
    enabled: !!id,
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
      void queryClient.invalidateQueries({ queryKey: ["accounts"] })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to update account")),
  })

  const archiveAccountMutation = useMutation({
    mutationFn: () => archiveAccount(id ?? ""),
    onSuccess: () => {
      toast.success("Account archived")
      void queryClient.invalidateQueries({ queryKey: ["accounts"] })
      navigate("/accounts", { replace: true })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to archive account")),
  })

  const archiveSecretMutation = useMutation({
    mutationFn: (secretID: string) => archiveSecret(id ?? "", secretID),
    onSuccess: () => {
      toast.success("Secret archived")
      void queryClient.invalidateQueries({ queryKey: ["account", id] })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to archive secret")),
  })

  const revealMutation = useMutation({
    mutationFn: ({ accountID, secretID }: { accountID: string; secretID: string }) =>
      revealSecret(accountID, secretID),
    onSuccess: (value, variables) => {
      setRevealed((current) => ({ ...current, [variables.secretID]: value }))
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to reveal secret")),
  })

  const account = accountQuery.data

  if (accountQuery.isLoading) {
    return (
      <PageContainer>
        <Skeleton className="mb-6 h-8 w-48" />
        <Skeleton className="h-96 rounded-xl" />
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
      <div className="mb-8 flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div className="min-w-0">
          <Button asChild variant="ghost" size="sm" className="-ml-2 mb-3">
            <Link to="/accounts">
              <ArrowLeft className="size-4" />
              Accounts
            </Link>
          </Button>
          <h1 className="truncate text-2xl font-semibold tracking-tight">{account.name}</h1>
          <p className="mt-1 max-w-2xl text-sm text-muted-foreground">
            {account.description || "No description"}
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
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
          <Button
            variant="destructive"
            disabled={archiveAccountMutation.isPending}
            onClick={() => archiveAccountMutation.mutate()}
          >
            <Trash2 className="size-4" />
            Archive
          </Button>
        </div>
      </div>

      <div className="mb-8 flex flex-wrap gap-1.5">
        {account.access_group_names.length === 0 ? (
          <Badge variant="secondary">All Sentinel users</Badge>
        ) : (
          account.access_group_names.map((group) => (
            <Badge key={group} variant="outline">
              {group}
            </Badge>
          ))
        )}
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0">
          <CardTitle>Secrets</CardTitle>
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
            <div className="px-4 py-10 text-center text-sm text-muted-foreground">
              No secrets yet.
            </div>
          ) : (
            <div className="divide-y">
              {account.secrets.map((secret) => (
                <div
                  key={secret.id}
                  className="grid gap-3 px-4 py-4 lg:grid-cols-[minmax(180px,1fr)_160px_minmax(220px,1.4fr)_auto] lg:items-center"
                >
                  <div className="min-w-0">
                    <div className="truncate text-sm font-medium">{secret.label || secret.key}</div>
                    <div className="mt-1 truncate font-mono text-xs text-muted-foreground">
                      {secret.key}
                    </div>
                  </div>
                  <div className="flex flex-wrap gap-1.5">
                    {secret.type && <Badge variant="outline">{secret.type}</Badge>}
                    <Badge variant={secret.sensitive ? "secondary" : "outline"}>
                      {secret.sensitive ? "Sensitive" : "Plain"}
                    </Badge>
                  </div>
                  <SecretValue
                    accountID={account.id}
                    secret={secret}
                    revealed={revealed[secret.id]}
                    revealing={revealMutation.isPending}
                    onReveal={(accountID, secretID) => revealMutation.mutate({ accountID, secretID })}
                    onHide={(secretID) =>
                      setRevealed((current) => {
                        const next = { ...current }
                        delete next[secretID]
                        return next
                      })
                    }
                  />
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    disabled={archiveSecretMutation.isPending}
                    onClick={() => archiveSecretMutation.mutate(secret.id)}
                  >
                    <Trash2 className="size-3.5" />
                    <span className="sr-only">Archive secret</span>
                  </Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </PageContainer>
  )
}
