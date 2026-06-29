import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Edit2, Monitor, Moon, Plus, Server, Sun, Trash2, type LucideIcon } from "lucide-react"
import { useState, type ReactNode } from "react"
import { toast } from "sonner"

import { ConfirmDialog } from "@/components/ConfirmDialog"
import { GithubIcon } from "@/components/icons/socials"
import { PageContainer, PageHeader } from "@/components/PageContainer"
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
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Skeleton } from "@/components/ui/skeleton"
import { Textarea } from "@/components/ui/textarea"
import { useTheme, type Theme } from "@/lib/theme"
import {
  createGitHubActionsRule,
  createKubernetesSecretRule,
  deleteGitHubActionsRule,
  deleteKubernetesSecretRule,
  listGitHubActionsRules,
  listKubernetesSecretRules,
  updateGitHubActionsRule,
  updateKubernetesSecretRule,
  type GitHubActionsRule,
  type GitHubActionsRuleInput,
  type KubernetesSecretRule,
  type KubernetesSecretRuleInput,
} from "@/lib/vault"

const themeOptions: Array<{ value: Theme; label: string; Icon: LucideIcon }> = [
  { value: "system", label: "System", Icon: Monitor },
  { value: "light", label: "Light", Icon: Sun },
  { value: "dark", label: "Dark", Icon: Moon },
]

function errorMessage(error: unknown, fallback: string) {
  return (error as { response?: { data?: { error?: string } } })?.response?.data?.error ?? fallback
}

function normalizeIdentifier(value: string) {
  return value.trim().toLowerCase()
}

function normalizeList(value: string, lower = false) {
  return Array.from(
    new Set(
      value
        .split(/[\n,]+/)
        .map((item) => item.trim())
        .filter(Boolean)
        .map((item) => (lower ? item.toLowerCase() : item)),
    ),
  )
}

function listText(values?: string[]) {
  return (values ?? []).join("\n")
}

function patternSummary(patterns: string[], singular: string) {
  return `${patterns.length} ${singular}${patterns.length === 1 ? "" : "s"}`
}

function GitHubActionsRuleDialog({
  rule,
  trigger,
  isPending,
  onSubmit,
}: {
  rule?: GitHubActionsRule
  trigger: ReactNode
  isPending: boolean
  onSubmit: (input: GitHubActionsRuleInput) => Promise<unknown>
}) {
  const [open, setOpen] = useState(false)
  const [name, setName] = useState(rule?.name ?? "")
  const [repositoryPatterns, setRepositoryPatterns] = useState(() => listText(rule?.repository_patterns))
  const [refPatterns, setRefPatterns] = useState(() => listText(rule?.ref_patterns))
  const [secretSelectors, setSecretSelectors] = useState(() => listText(rule?.secret_selectors))
  const [enabled, setEnabled] = useState(rule?.enabled ?? true)
  const isEditing = !!rule

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen)
    if (nextOpen) return
    setName(rule?.name ?? "")
    setRepositoryPatterns(listText(rule?.repository_patterns))
    setRefPatterns(listText(rule?.ref_patterns))
    setSecretSelectors(listText(rule?.secret_selectors))
    setEnabled(rule?.enabled ?? true)
  }

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    await onSubmit({
      name: normalizeIdentifier(name),
      repository_patterns: normalizeList(repositoryPatterns, true),
      ref_patterns: normalizeList(refPatterns),
      secret_selectors: normalizeList(secretSelectors, true),
      enabled,
    })
    handleOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{isEditing ? "Edit GitHub Actions rule" : "New GitHub Actions rule"}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_160px]">
            <div className="space-y-2">
              <Label htmlFor={isEditing ? `github-rule-name-${rule.id}` : "github-rule-name"}>
                Name
              </Label>
              <Input
                id={isEditing ? `github-rule-name-${rule.id}` : "github-rule-name"}
                value={name}
                onChange={(event) => setName(event.target.value.toLowerCase())}
                placeholder="publish-packages"
                required
              />
            </div>
            <label className="flex items-center justify-between gap-3 self-end rounded-lg bg-muted/40 px-3 py-2 text-sm dark:bg-muted/30">
              <span className="font-medium">Enabled</span>
              <input
                type="checkbox"
                checked={enabled}
                onChange={(event) => setEnabled(event.target.checked)}
                className="size-4 accent-gr-pink"
              />
            </label>
          </div>

          <div className="grid gap-3 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor={isEditing ? `github-rule-repos-${rule.id}` : "github-rule-repos"}>
                Repositories
              </Label>
              <Textarea
                id={isEditing ? `github-rule-repos-${rule.id}` : "github-rule-repos"}
                value={repositoryPatterns}
                onChange={(event) => setRepositoryPatterns(event.target.value)}
                placeholder={"gaucho-racing/mapache\ngaucho-racing/*"}
                className="min-h-28 font-mono"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor={isEditing ? `github-rule-refs-${rule.id}` : "github-rule-refs"}>
                Refs
              </Label>
              <Textarea
                id={isEditing ? `github-rule-refs-${rule.id}` : "github-rule-refs"}
                value={refPatterns}
                onChange={(event) => setRefPatterns(event.target.value)}
                placeholder={"refs/heads/main\nrefs/tags/v*"}
                className="min-h-28 font-mono"
                required
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor={isEditing ? `github-rule-selectors-${rule.id}` : "github-rule-selectors"}>
              Secret selectors
            </Label>
            <Textarea
              id={isEditing ? `github-rule-selectors-${rule.id}` : "github-rule-selectors"}
              value={secretSelectors}
              onChange={(event) => setSecretSelectors(event.target.value)}
              placeholder={"pypi.publish_token\nmapache-prod.*"}
              className="min-h-28 font-mono"
              required
            />
          </div>

          <div className="flex flex-wrap justify-end gap-2">
            <Button type="button" variant="secondary" onClick={() => handleOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? "Saving" : "Save"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function KubernetesSecretRuleDialog({
  rule,
  trigger,
  isPending,
  onSubmit,
}: {
  rule?: KubernetesSecretRule
  trigger: ReactNode
  isPending: boolean
  onSubmit: (input: KubernetesSecretRuleInput) => Promise<unknown>
}) {
  const [open, setOpen] = useState(false)
  const [name, setName] = useState(rule?.name ?? "")
  const [clusterPatterns, setClusterPatterns] = useState(() => listText(rule?.cluster_patterns))
  const [namespacePatterns, setNamespacePatterns] = useState(() => listText(rule?.namespace_patterns))
  const [serviceAccountPatterns, setServiceAccountPatterns] = useState(() =>
    listText(rule?.service_account_patterns),
  )
  const [secretSelectors, setSecretSelectors] = useState(() => listText(rule?.secret_selectors))
  const [enabled, setEnabled] = useState(rule?.enabled ?? true)
  const isEditing = !!rule

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen)
    if (nextOpen) return
    setName(rule?.name ?? "")
    setClusterPatterns(listText(rule?.cluster_patterns))
    setNamespacePatterns(listText(rule?.namespace_patterns))
    setServiceAccountPatterns(listText(rule?.service_account_patterns))
    setSecretSelectors(listText(rule?.secret_selectors))
    setEnabled(rule?.enabled ?? true)
  }

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    await onSubmit({
      name: normalizeIdentifier(name),
      cluster_patterns: normalizeList(clusterPatterns, true),
      namespace_patterns: normalizeList(namespacePatterns, true),
      service_account_patterns: normalizeList(serviceAccountPatterns, true),
      secret_selectors: normalizeList(secretSelectors, true),
      enabled,
    })
    handleOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{isEditing ? "Edit Kubernetes rule" : "New Kubernetes rule"}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_160px]">
            <div className="space-y-2">
              <Label htmlFor={isEditing ? `kubernetes-rule-name-${rule.id}` : "kubernetes-rule-name"}>
                Name
              </Label>
              <Input
                id={isEditing ? `kubernetes-rule-name-${rule.id}` : "kubernetes-rule-name"}
                value={name}
                onChange={(event) => setName(event.target.value.toLowerCase())}
                placeholder="production-sync"
                required
              />
            </div>
            <label className="flex items-center justify-between gap-3 self-end rounded-lg bg-muted/40 px-3 py-2 text-sm dark:bg-muted/30">
              <span className="font-medium">Enabled</span>
              <input
                type="checkbox"
                checked={enabled}
                onChange={(event) => setEnabled(event.target.checked)}
                className="size-4 accent-gr-pink"
              />
            </label>
          </div>

          <div className="grid gap-3 md:grid-cols-3">
            <div className="space-y-2">
              <Label htmlFor={isEditing ? `kubernetes-rule-clusters-${rule.id}` : "kubernetes-rule-clusters"}>
                Clusters
              </Label>
              <Textarea
                id={isEditing ? `kubernetes-rule-clusters-${rule.id}` : "kubernetes-rule-clusters"}
                value={clusterPatterns}
                onChange={(event) => setClusterPatterns(event.target.value)}
                placeholder={"default\nprod-*"}
                className="min-h-28 font-mono"
                required
              />
            </div>
            <div className="space-y-2">
              <Label
                htmlFor={isEditing ? `kubernetes-rule-namespaces-${rule.id}` : "kubernetes-rule-namespaces"}
              >
                Namespaces
              </Label>
              <Textarea
                id={isEditing ? `kubernetes-rule-namespaces-${rule.id}` : "kubernetes-rule-namespaces"}
                value={namespacePatterns}
                onChange={(event) => setNamespacePatterns(event.target.value)}
                placeholder={"default\nmapache-*"}
                className="min-h-28 font-mono"
                required
              />
            </div>
            <div className="space-y-2">
              <Label
                htmlFor={
                  isEditing
                    ? `kubernetes-rule-service-accounts-${rule.id}`
                    : "kubernetes-rule-service-accounts"
                }
              >
                Service accounts
              </Label>
              <Textarea
                id={
                  isEditing
                    ? `kubernetes-rule-service-accounts-${rule.id}`
                    : "kubernetes-rule-service-accounts"
                }
                value={serviceAccountPatterns}
                onChange={(event) => setServiceAccountPatterns(event.target.value)}
                placeholder={"vault-sync\n*-deployer"}
                className="min-h-28 font-mono"
                required
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor={isEditing ? `kubernetes-rule-selectors-${rule.id}` : "kubernetes-rule-selectors"}>
              Secret selectors
            </Label>
            <Textarea
              id={isEditing ? `kubernetes-rule-selectors-${rule.id}` : "kubernetes-rule-selectors"}
              value={secretSelectors}
              onChange={(event) => setSecretSelectors(event.target.value)}
              placeholder={"pypi.publish_token\nmapache-prod.*"}
              className="min-h-28 font-mono"
              required
            />
          </div>

          <div className="flex flex-wrap justify-end gap-2">
            <Button type="button" variant="secondary" onClick={() => handleOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? "Saving" : "Save"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function PatternBadges({ values, variant = "outline" }: { values: string[]; variant?: "outline" | "secondary" }) {
  return (
    <div className="flex min-h-6 flex-wrap gap-1.5">
      {values.map((value) => (
        <Badge key={value} variant={variant} className="max-w-full truncate" title={value}>
          {value}
        </Badge>
      ))}
    </div>
  )
}

export default function SettingsPage() {
  const { theme, setTheme } = useTheme()
  const queryClient = useQueryClient()

  const rulesQuery = useQuery({
    queryKey: ["githubActionsRules"],
    queryFn: listGitHubActionsRules,
  })

  const kubernetesRulesQuery = useQuery({
    queryKey: ["kubernetesSecretRules"],
    queryFn: listKubernetesSecretRules,
  })

  const createRuleMutation = useMutation({
    mutationFn: createGitHubActionsRule,
    onSuccess: () => {
      toast.success("GitHub Actions rule created")
      void queryClient.invalidateQueries({ queryKey: ["githubActionsRules"] })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to create GitHub Actions rule")),
  })

  const updateRuleMutation = useMutation({
    mutationFn: ({ id, input }: { id: string; input: GitHubActionsRuleInput }) =>
      updateGitHubActionsRule(id, input),
    onSuccess: () => {
      toast.success("GitHub Actions rule updated")
      void queryClient.invalidateQueries({ queryKey: ["githubActionsRules"] })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to update GitHub Actions rule")),
  })

  const deleteRuleMutation = useMutation({
    mutationFn: deleteGitHubActionsRule,
    onSuccess: () => {
      toast.success("GitHub Actions rule deleted")
      void queryClient.invalidateQueries({ queryKey: ["githubActionsRules"] })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to delete GitHub Actions rule")),
  })

  const createKubernetesRuleMutation = useMutation({
    mutationFn: createKubernetesSecretRule,
    onSuccess: () => {
      toast.success("Kubernetes rule created")
      void queryClient.invalidateQueries({ queryKey: ["kubernetesSecretRules"] })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to create Kubernetes rule")),
  })

  const updateKubernetesRuleMutation = useMutation({
    mutationFn: ({ id, input }: { id: string; input: KubernetesSecretRuleInput }) =>
      updateKubernetesSecretRule(id, input),
    onSuccess: () => {
      toast.success("Kubernetes rule updated")
      void queryClient.invalidateQueries({ queryKey: ["kubernetesSecretRules"] })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to update Kubernetes rule")),
  })

  const deleteKubernetesRuleMutation = useMutation({
    mutationFn: deleteKubernetesSecretRule,
    onSuccess: () => {
      toast.success("Kubernetes rule deleted")
      void queryClient.invalidateQueries({ queryKey: ["kubernetesSecretRules"] })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to delete Kubernetes rule")),
  })

  const rules = rulesQuery.data ?? []
  const kubernetesRules = kubernetesRulesQuery.data ?? []

  return (
    <PageContainer>
      <PageHeader title="Settings" />

      <div className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>Appearance</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-2">
              {themeOptions.map(({ value, label, Icon }) => (
                <Button
                  key={value}
                  variant={theme === value ? "default" : "secondary"}
                  onClick={() => setTheme(value)}
                  aria-pressed={theme === value}
                >
                  <Icon className="size-4" />
                  {label}
                </Button>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between gap-3 border-b border-border/50 pb-4">
            <div>
              <CardTitle>GitHub Actions</CardTitle>
              <div className="mt-1 text-sm text-muted-foreground">
                {rules.length} rule{rules.length === 1 ? "" : "s"}
              </div>
            </div>
            <GitHubActionsRuleDialog
              isPending={createRuleMutation.isPending}
              onSubmit={(input) => createRuleMutation.mutateAsync(input)}
              trigger={
                <Button size="sm">
                  <Plus className="size-4" />
                  Rule
                </Button>
              }
            />
          </CardHeader>
          <CardContent className="p-0">
            {rulesQuery.isLoading ? (
              <div className="space-y-3 p-4">
                <Skeleton className="h-24 rounded-lg" />
                <Skeleton className="h-24 rounded-lg" />
              </div>
            ) : rules.length === 0 ? (
              <div className="flex min-h-44 flex-col items-center justify-center px-4 py-10 text-center">
                <div className="flex size-10 items-center justify-center rounded-lg bg-muted">
                  <GithubIcon className="size-5 text-muted-foreground" />
                </div>
                <div className="mt-4 text-sm font-medium">No GitHub Actions rules</div>
              </div>
            ) : (
              <div>
                {rules.map((rule) => (
                  <div
                    key={rule.id}
                    className="grid gap-4 border-b border-border/45 px-4 py-4 last:border-b-0 xl:grid-cols-[minmax(180px,0.8fr)_minmax(220px,1fr)_minmax(220px,1fr)_minmax(220px,1fr)_92px] xl:items-start"
                  >
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground">
                          <GithubIcon className="size-4" />
                        </div>
                        <div className="min-w-0">
                          <div className="truncate font-mono text-sm font-medium">{rule.name}</div>
                          <div className="mt-1 text-xs text-muted-foreground">
                            {rule.enabled ? "Enabled" : "Disabled"}
                          </div>
                        </div>
                      </div>
                    </div>
                    <div className="min-w-0 space-y-2">
                      <div className="text-xs font-medium text-muted-foreground">
                        {patternSummary(rule.repository_patterns, "repo pattern")}
                      </div>
                      <PatternBadges values={rule.repository_patterns} />
                    </div>
                    <div className="min-w-0 space-y-2">
                      <div className="text-xs font-medium text-muted-foreground">
                        {patternSummary(rule.ref_patterns, "ref pattern")}
                      </div>
                      <PatternBadges values={rule.ref_patterns} />
                    </div>
                    <div className="min-w-0 space-y-2">
                      <div className="text-xs font-medium text-muted-foreground">
                        {patternSummary(rule.secret_selectors, "selector")}
                      </div>
                      <PatternBadges values={rule.secret_selectors} variant="secondary" />
                    </div>
                    <div className="flex gap-1 xl:justify-end">
                      <GitHubActionsRuleDialog
                        rule={rule}
                        isPending={updateRuleMutation.isPending}
                        onSubmit={(input) => updateRuleMutation.mutateAsync({ id: rule.id, input })}
                        trigger={
                          <Button variant="ghost" size="icon-sm">
                            <Edit2 className="size-3.5" />
                            <span className="sr-only">Edit GitHub Actions rule</span>
                          </Button>
                        }
                      />
                      <ConfirmDialog
                        title="Delete GitHub Actions rule"
                        description="This will remove this GitHub Actions access rule."
                        confirmLabel="Delete rule"
                        isPending={deleteRuleMutation.isPending}
                        onConfirm={() => deleteRuleMutation.mutateAsync(rule.id)}
                        trigger={
                          <Button variant="ghost" size="icon-sm" disabled={deleteRuleMutation.isPending}>
                            <Trash2 className="size-3.5" />
                            <span className="sr-only">Delete GitHub Actions rule</span>
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

        <Card>
          <CardHeader className="flex flex-row items-center justify-between gap-3 border-b border-border/50 pb-4">
            <div>
              <CardTitle>Kubernetes</CardTitle>
              <div className="mt-1 text-sm text-muted-foreground">
                {kubernetesRules.length} rule{kubernetesRules.length === 1 ? "" : "s"}
              </div>
            </div>
            <KubernetesSecretRuleDialog
              isPending={createKubernetesRuleMutation.isPending}
              onSubmit={(input) => createKubernetesRuleMutation.mutateAsync(input)}
              trigger={
                <Button size="sm">
                  <Plus className="size-4" />
                  Rule
                </Button>
              }
            />
          </CardHeader>
          <CardContent className="p-0">
            {kubernetesRulesQuery.isLoading ? (
              <div className="space-y-3 p-4">
                <Skeleton className="h-24 rounded-lg" />
                <Skeleton className="h-24 rounded-lg" />
              </div>
            ) : kubernetesRules.length === 0 ? (
              <div className="flex min-h-44 flex-col items-center justify-center px-4 py-10 text-center">
                <div className="flex size-10 items-center justify-center rounded-lg bg-muted">
                  <Server className="size-5 text-muted-foreground" />
                </div>
                <div className="mt-4 text-sm font-medium">No Kubernetes rules</div>
              </div>
            ) : (
              <div>
                {kubernetesRules.map((rule) => (
                  <div
                    key={rule.id}
                    className="grid gap-4 border-b border-border/45 px-4 py-4 last:border-b-0 xl:grid-cols-[minmax(180px,0.75fr)_minmax(160px,0.9fr)_minmax(160px,0.9fr)_minmax(190px,1fr)_minmax(220px,1.1fr)_92px] xl:items-start"
                  >
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground">
                          <Server className="size-4" />
                        </div>
                        <div className="min-w-0">
                          <div className="truncate font-mono text-sm font-medium">{rule.name}</div>
                          <div className="mt-1 text-xs text-muted-foreground">
                            {rule.enabled ? "Enabled" : "Disabled"}
                          </div>
                        </div>
                      </div>
                    </div>
                    <div className="min-w-0 space-y-2">
                      <div className="text-xs font-medium text-muted-foreground">
                        {patternSummary(rule.cluster_patterns, "cluster")}
                      </div>
                      <PatternBadges values={rule.cluster_patterns} />
                    </div>
                    <div className="min-w-0 space-y-2">
                      <div className="text-xs font-medium text-muted-foreground">
                        {patternSummary(rule.namespace_patterns, "namespace")}
                      </div>
                      <PatternBadges values={rule.namespace_patterns} />
                    </div>
                    <div className="min-w-0 space-y-2">
                      <div className="text-xs font-medium text-muted-foreground">
                        {patternSummary(rule.service_account_patterns, "service account")}
                      </div>
                      <PatternBadges values={rule.service_account_patterns} />
                    </div>
                    <div className="min-w-0 space-y-2">
                      <div className="text-xs font-medium text-muted-foreground">
                        {patternSummary(rule.secret_selectors, "selector")}
                      </div>
                      <PatternBadges values={rule.secret_selectors} variant="secondary" />
                    </div>
                    <div className="flex gap-1 xl:justify-end">
                      <KubernetesSecretRuleDialog
                        rule={rule}
                        isPending={updateKubernetesRuleMutation.isPending}
                        onSubmit={(input) =>
                          updateKubernetesRuleMutation.mutateAsync({ id: rule.id, input })
                        }
                        trigger={
                          <Button variant="ghost" size="icon-sm">
                            <Edit2 className="size-3.5" />
                            <span className="sr-only">Edit Kubernetes rule</span>
                          </Button>
                        }
                      />
                      <ConfirmDialog
                        title="Delete Kubernetes rule"
                        description="This will remove this Kubernetes access rule."
                        confirmLabel="Delete rule"
                        isPending={deleteKubernetesRuleMutation.isPending}
                        onConfirm={() => deleteKubernetesRuleMutation.mutateAsync(rule.id)}
                        trigger={
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            disabled={deleteKubernetesRuleMutation.isPending}
                          >
                            <Trash2 className="size-3.5" />
                            <span className="sr-only">Delete Kubernetes rule</span>
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
      </div>
    </PageContainer>
  )
}
