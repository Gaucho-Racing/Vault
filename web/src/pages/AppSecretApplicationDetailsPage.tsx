import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  ArrowLeft,
  Clock3,
  Copy,
  Download,
  Eye,
  EyeOff,
  Plus,
  CodeXml,
  Trash2,
  UsersRound,
} from "lucide-react"
import { useState, type ReactNode } from "react"
import { Link, useNavigate, useParams } from "react-router-dom"
import { toast } from "sonner"

import { AppSecretApplicationForm } from "@/components/AppSecretApplicationForm"
import { AppSecretFormDialog } from "@/components/AppSecretFormDialog"
import { ConfirmDialog } from "@/components/ConfirmDialog"
import { PageContainer } from "@/components/PageContainer"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Skeleton } from "@/components/ui/skeleton"
import {
  createAppSecret,
  deleteAppSecret,
  deleteAppSecretApplication,
  downloadAppSecretEnvFile,
  getAppSecretApplication,
  revealAppSecret,
  updateAppSecret,
  updateAppSecretApplication,
  type AppSecret,
  type AppSecretApplication,
  type AppSecretApplicationInput,
  type AppSecretInput,
} from "@/lib/vault"

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

async function copyValue(value: string) {
  try {
    await navigator.clipboard.writeText(value)
    toast.success("Copied")
  } catch {
    toast.error("Failed to copy")
  }
}

function saveBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement("a")
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  window.setTimeout(() => URL.revokeObjectURL(url), 0)
}

function ApplicationFormDialog({
  application,
  isPending,
  trigger,
  onSubmit,
}: {
  application: AppSecretApplication
  isPending: boolean
  trigger: ReactNode
  onSubmit: (input: AppSecretApplicationInput) => Promise<void>
}) {
  const [open, setOpen] = useState(false)

  async function handleSubmit(input: AppSecretApplicationInput) {
    await onSubmit(input)
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Edit application</DialogTitle>
        </DialogHeader>
        <AppSecretApplicationForm
          application={application}
          isPending={isPending}
          onSubmit={handleSubmit}
          onCancel={() => setOpen(false)}
        />
      </DialogContent>
    </Dialog>
  )
}

function SecretValue({
  applicationID,
  secret,
  revealed,
  revealing,
  copying,
  onReveal,
  onCopy,
  onHide,
}: {
  applicationID: string
  secret: AppSecret
  revealed?: string
  revealing: boolean
  copying: boolean
  onReveal: (applicationID: string, secretID: string) => void
  onCopy: (applicationID: string, secretID: string) => void
  onHide: (secretID: string) => void
}) {
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
        onClick={() => onCopy(applicationID, secret.id)}
      >
        <Copy className="size-3.5" />
        {copying ? "Copying" : "Copy"}
      </Button>
      <Button
        variant="secondary"
        size="sm"
        disabled={revealing}
        onClick={() => onReveal(applicationID, secret.id)}
      >
        <Eye className="size-3.5" />
        {revealing ? "Revealing" : "Reveal"}
      </Button>
    </div>
  )
}

export default function AppSecretApplicationDetailsPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [revealed, setRevealed] = useState<Record<string, string>>({})

  const applicationQuery = useQuery({
    queryKey: ["appSecretApplication", id],
    queryFn: () => getAppSecretApplication(id ?? ""),
    enabled: !!id,
  })

  const createSecretMutation = useMutation({
    mutationFn: (input: AppSecretInput) => createAppSecret(id ?? "", input),
    onSuccess: () => {
      toast.success("App secret created")
      void queryClient.invalidateQueries({ queryKey: ["appSecretApplication", id] })
      void queryClient.invalidateQueries({ queryKey: ["appSecretApplications"] })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to create app secret")),
  })

  const updateApplicationMutation = useMutation({
    mutationFn: (input: AppSecretApplicationInput) => updateAppSecretApplication(id ?? "", input),
    onSuccess: () => {
      toast.success("Application updated")
      void queryClient.invalidateQueries({ queryKey: ["appSecretApplication", id] })
      void queryClient.invalidateQueries({ queryKey: ["appSecretApplications"] })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to update application")),
  })

  const deleteApplicationMutation = useMutation({
    mutationFn: () => deleteAppSecretApplication(id ?? ""),
    onSuccess: () => {
      toast.success("Application deleted")
      void queryClient.invalidateQueries({ queryKey: ["appSecretApplications"] })
      navigate("/app-secrets", { replace: true })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to delete application")),
  })

  const updateSecretMutation = useMutation({
    mutationFn: ({ secretID, input }: { secretID: string; input: AppSecretInput }) =>
      updateAppSecret(id ?? "", secretID, input),
    onSuccess: (_, variables) => {
      toast.success("App secret updated")
      setRevealed((current) => {
        const next = { ...current }
        delete next[variables.secretID]
        return next
      })
      void queryClient.invalidateQueries({ queryKey: ["appSecretApplication", id] })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to update app secret")),
  })

  const deleteSecretMutation = useMutation({
    mutationFn: (secretID: string) => deleteAppSecret(id ?? "", secretID),
    onSuccess: (_, secretID) => {
      toast.success("App secret deleted")
      setRevealed((current) => {
        const next = { ...current }
        delete next[secretID]
        return next
      })
      void queryClient.invalidateQueries({ queryKey: ["appSecretApplication", id] })
      void queryClient.invalidateQueries({ queryKey: ["appSecretApplications"] })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to delete app secret")),
  })

  const revealMutation = useMutation({
    mutationFn: ({ applicationID, secretID }: { applicationID: string; secretID: string }) =>
      revealAppSecret(applicationID, secretID),
    onSuccess: (value, variables) => {
      setRevealed((current) => ({ ...current, [variables.secretID]: value }))
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to reveal app secret")),
  })

  const revealAllMutation = useMutation({
    mutationFn: async (secrets: AppSecret[]) => {
      const entries = await Promise.all(
        secrets.map(async (secret) => [secret.id, await revealAppSecret(id ?? "", secret.id)] as const),
      )
      return Object.fromEntries(entries)
    },
    onSuccess: (values) => setRevealed(values),
    onError: (error) => toast.error(errorMessage(error, "Failed to reveal app secrets")),
  })

  const copySecretMutation = useMutation({
    mutationFn: async ({ applicationID, secretID }: { applicationID: string; secretID: string }) => {
      const value = await revealAppSecret(applicationID, secretID)
      await navigator.clipboard.writeText(value)
    },
    onSuccess: () => toast.success("Copied"),
    onError: (error) => toast.error(errorMessage(error, "Failed to copy app secret")),
  })

  const downloadEnvMutation = useMutation({
    mutationFn: () => downloadAppSecretEnvFile(id ?? ""),
    onSuccess: (blob) => saveBlob(blob, `${applicationQuery.data?.name ?? "application"}.env`),
    onError: (error) => toast.error(errorMessage(error, "Failed to download env file")),
  })

  const application = applicationQuery.data
  const allSecretsRevealed =
    !!application &&
    application.secrets.length > 0 &&
    application.secrets.every((secret) => revealed[secret.id] !== undefined)

  if (applicationQuery.isLoading) {
    return (
      <PageContainer>
        <Skeleton className="mb-4 h-8 w-48" />
        <Skeleton className="h-32 rounded-lg" />
        <Skeleton className="mt-4 h-80 rounded-lg" />
      </PageContainer>
    )
  }

  if (!application) {
    return (
      <PageContainer>
        <Card>
          <CardContent className="py-10 text-center text-sm text-muted-foreground">
            Application not found.
          </CardContent>
        </Card>
      </PageContainer>
    )
  }

  async function handleUpdateApplication(input: AppSecretApplicationInput) {
    await updateApplicationMutation.mutateAsync(input)
  }

  async function handleCreateSecret(input: AppSecretInput) {
    await createSecretMutation.mutateAsync(input)
  }

  return (
    <PageContainer>
      <div className="mb-6">
        <div className="mb-4 flex items-center justify-between gap-3">
          <Button asChild variant="ghost" size="sm" className="-ml-2">
            <Link to="/app-secrets">
              <ArrowLeft className="size-4" />
              App Secrets
            </Link>
          </Button>
          <div className="flex flex-wrap justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              disabled={downloadEnvMutation.isPending}
              onClick={() => downloadEnvMutation.mutate()}
            >
              <Download className="size-4" />
              {downloadEnvMutation.isPending ? "Downloading" : ".env"}
            </Button>
            <ApplicationFormDialog
              application={application}
              isPending={updateApplicationMutation.isPending}
              onSubmit={handleUpdateApplication}
              trigger={<Button variant="outline">Edit</Button>}
            />
            <ConfirmDialog
              title="Delete application"
              description="This will delete the application and its app secrets from Vault."
              confirmLabel="Delete application"
              isPending={deleteApplicationMutation.isPending}
              onConfirm={() => deleteApplicationMutation.mutateAsync()}
              trigger={
                <Button variant="destructive" disabled={deleteApplicationMutation.isPending}>
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
                  <CodeXml className="size-5" />
                </div>
                <h1 className="truncate font-mono text-2xl font-semibold">{application.name}</h1>
              </div>
            </div>
            <div className="grid gap-2 sm:grid-cols-2 lg:w-[420px]">
              <div className="rounded-lg bg-muted/45 p-3 dark:bg-muted/35">
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <UsersRound className="size-3.5" />
                  Access
                </div>
                <div className="mt-1 truncate text-sm font-medium">
                  {application.access_group_names.length === 0
                    ? "Public"
                    : `${application.access_group_names.length} group${
                        application.access_group_names.length === 1 ? "" : "s"
                      }`}
                </div>
              </div>
              <div className="rounded-lg bg-muted/45 p-3 dark:bg-muted/35">
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <Clock3 className="size-3.5" />
                  Updated
                </div>
                <div className="mt-1 truncate text-sm font-medium">
                  {formatDate(application.updated_at)}
                </div>
              </div>
            </div>
          </div>
          <div className="mt-4 flex flex-wrap gap-1.5 border-t border-border/50 pt-4">
            {application.access_group_names.length === 0 ? (
              <Badge variant="secondary">Public</Badge>
            ) : (
              application.access_group_names.map((group) => (
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
              {application.secrets.length} item{application.secrets.length === 1 ? "" : "s"}
            </div>
          </div>
          <div className="flex flex-wrap justify-end gap-2">
            {allSecretsRevealed ? (
              <Button type="button" variant="secondary" size="sm" onClick={() => setRevealed({})}>
                <EyeOff className="size-4" />
                Hide all
              </Button>
            ) : (
              <Button
                type="button"
                variant="secondary"
                size="sm"
                disabled={application.secrets.length === 0 || revealAllMutation.isPending}
                onClick={() => revealAllMutation.mutate(application.secrets)}
              >
                <Eye className="size-4" />
                {revealAllMutation.isPending ? "Showing" : "Show all"}
              </Button>
            )}
            <AppSecretFormDialog
              isPending={createSecretMutation.isPending}
              onSubmit={handleCreateSecret}
              trigger={
                <Button size="sm">
                  <Plus className="size-4" />
                  Secret
                </Button>
              }
            />
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {application.secrets.length === 0 ? (
            <div className="flex min-h-56 flex-col items-center justify-center px-4 py-10 text-center">
              <div className="flex size-10 items-center justify-center rounded-lg bg-muted">
                <CodeXml className="size-5 text-muted-foreground" />
              </div>
              <div className="mt-4 text-sm font-medium">No app secrets yet</div>
              <div className="mt-1 text-sm text-muted-foreground">
                Add the first environment value for this application.
              </div>
            </div>
          ) : (
            <div>
              <div className="hidden grid-cols-[minmax(180px,1fr)_minmax(220px,1.4fr)_132px] border-b border-border/50 px-4 py-3 text-xs font-medium text-muted-foreground lg:grid">
                <div>Secret</div>
                <div>Value</div>
                <div />
              </div>
              {application.secrets.map((secret) => (
                <div
                  key={secret.id}
                  className="grid gap-3 border-b border-border/45 px-4 py-4 last:border-b-0 lg:grid-cols-[minmax(180px,1fr)_minmax(220px,1.4fr)_132px] lg:items-center"
                >
                  <div className="min-w-0">
                    <div className="truncate font-mono text-sm font-medium">{secret.key}</div>
                    <div className="mt-1 truncate font-mono text-xs text-muted-foreground">
                      {application.name}.{secret.key}
                    </div>
                  </div>
                  <SecretValue
                    applicationID={application.id}
                    secret={secret}
                    revealed={revealed[secret.id]}
                    revealing={
                      revealMutation.isPending && revealMutation.variables?.secretID === secret.id
                    }
                    copying={
                      copySecretMutation.isPending &&
                      copySecretMutation.variables?.secretID === secret.id
                    }
                    onReveal={(applicationID, secretID) =>
                      revealMutation.mutate({ applicationID, secretID })
                    }
                    onCopy={(applicationID, secretID) =>
                      copySecretMutation.mutate({ applicationID, secretID })
                    }
                    onHide={(secretID) =>
                      setRevealed((current) => {
                        const next = { ...current }
                        delete next[secretID]
                        return next
                      })
                    }
                  />
                  <div className="flex justify-start gap-1 lg:justify-end">
                    <AppSecretFormDialog
                      appSecret={secret}
                      isPending={updateSecretMutation.isPending}
                      onSubmit={(input) =>
                        updateSecretMutation.mutateAsync({ secretID: secret.id, input })
                      }
                      trigger={
                        <Button variant="ghost" size="sm">
                          Edit
                        </Button>
                      }
                    />
                    <ConfirmDialog
                      title="Delete app secret"
                      description="This will delete this app secret from the application."
                      confirmLabel="Delete app secret"
                      isPending={deleteSecretMutation.isPending}
                      onConfirm={() => deleteSecretMutation.mutateAsync(secret.id)}
                      trigger={
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          disabled={deleteSecretMutation.isPending}
                        >
                          <Trash2 className="size-3.5" />
                          <span className="sr-only">Delete app secret</span>
                        </Button>
                      }
                    />
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </PageContainer>
  )
}
