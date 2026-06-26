import { CodeXml, Globe2, ShieldCheck } from "lucide-react"
import { useState } from "react"

import { GroupPicker } from "@/components/AccountForm"
import { GithubIcon } from "@/components/icons/socials"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { cn } from "@/lib/utils"
import type { AppSecretApplication, AppSecretApplicationInput } from "@/lib/vault"

function normalizeIdentifier(value: string) {
  return value.trim().toLowerCase()
}

function normalizeGroups(groups: string[]) {
  return Array.from(new Set(groups.map((group) => group.trim()).filter(Boolean)))
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

function AccessSummary({ restricted }: { restricted: boolean }) {
  const Icon = restricted ? ShieldCheck : Globe2

  return (
    <Card className={cn("border-0 bg-muted/45 py-0 shadow-none", restricted && "bg-primary/10")}>
      <CardContent className="flex items-start gap-3 p-3">
        <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-background text-primary shadow-sm shadow-black/[0.03] dark:shadow-black/20">
          <Icon className="size-4" />
        </div>
        <div className="min-w-0">
          <div className="text-sm font-medium">{restricted ? "Restricted" : "Public"}</div>
          <div className="mt-1 text-sm leading-5 text-muted-foreground">
            {restricted
              ? "Only selected Sentinel groups can access this application. Admins always have access."
              : "All Sentinel users can access this application. Add groups to limit access."}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function GitHubActionsSummary({
  repositoryCount,
  refCount,
}: {
  repositoryCount: number
  refCount: number
}) {
  const enabled = repositoryCount > 0 && refCount > 0

  return (
    <Card className={cn("border-0 bg-muted/45 py-0 shadow-none", enabled && "bg-primary/10")}>
      <CardContent className="flex items-start gap-3 p-3">
        <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-background text-primary shadow-sm shadow-black/[0.03] dark:shadow-black/20">
          <GithubIcon className="size-4" />
        </div>
        <div className="min-w-0">
          <div className="text-sm font-medium">{enabled ? "Enabled" : "Disabled"}</div>
          <div className="mt-1 text-sm leading-5 text-muted-foreground">
            {enabled
              ? `${repositoryCount} repo${repositoryCount === 1 ? "" : "s"} and ${refCount} ref pattern${
                  refCount === 1 ? "" : "s"
                } can export these secrets.`
              : "Add repositories and ref patterns to allow GitHub Actions exports."}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export function AppSecretApplicationForm({
  application,
  isPending,
  submitLabel = "Save application",
  onSubmit,
  onCancel,
}: {
  application?: AppSecretApplication
  isPending: boolean
  submitLabel?: string
  onSubmit: (input: AppSecretApplicationInput) => Promise<void>
  onCancel?: () => void
}) {
  const [name, setName] = useState(application?.name ?? "")
  const [groups, setGroups] = useState(() => normalizeGroups(application?.access_group_names ?? []))
  const [githubRepositories, setGithubRepositories] = useState(() =>
    listText(application?.github_actions_repositories),
  )
  const [githubRefs, setGithubRefs] = useState(() => listText(application?.github_actions_refs))
  const normalizedGithubRepositories = normalizeList(githubRepositories, true)
  const normalizedGithubRefs = normalizeList(githubRefs)

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    await onSubmit({
      name: normalizeIdentifier(name),
      access_group_names: groups,
      github_actions_repositories: normalizedGithubRepositories,
      github_actions_refs: normalizedGithubRefs,
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="application-name">Name</Label>
        <div className="relative">
          <CodeXml className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            id="application-name"
            value={name}
            onChange={(event) => setName(event.target.value.toLowerCase())}
            placeholder="mapache-prod"
            className="pl-9 font-mono"
            required
          />
        </div>
      </div>

      <section className="space-y-3">
        <div>
          <h2 className="text-sm font-medium">Access</h2>
          <p className="mt-1 text-sm leading-5 text-muted-foreground">
            Applications are public by default. Select Sentinel groups to limit access.
          </p>
        </div>
        <AccessSummary restricted={groups.length > 0} />
        <GroupPicker selectedGroups={groups} onChange={setGroups} />
      </section>

      <section className="space-y-3">
        <h2 className="text-sm font-medium">GitHub Actions</h2>
        <GitHubActionsSummary
          repositoryCount={normalizedGithubRepositories.length}
          refCount={normalizedGithubRefs.length}
        />
        <div className="grid gap-3 md:grid-cols-2">
          <div className="space-y-2">
            <Label htmlFor="github-actions-repositories">Repositories</Label>
            <Textarea
              id="github-actions-repositories"
              value={githubRepositories}
              onChange={(event) => setGithubRepositories(event.target.value)}
              placeholder="gaucho-racing/mapache"
              className="min-h-24 font-mono"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="github-actions-refs">Refs</Label>
            <Textarea
              id="github-actions-refs"
              value={githubRefs}
              onChange={(event) => setGithubRefs(event.target.value)}
              placeholder={"refs/heads/main\nrefs/tags/v*"}
              className="min-h-24 font-mono"
            />
          </div>
        </div>
      </section>

      <div className="flex flex-wrap justify-end gap-2">
        {onCancel && (
          <Button type="button" variant="secondary" onClick={onCancel}>
            Cancel
          </Button>
        )}
        <Button type="submit" disabled={isPending}>
          {isPending ? "Saving" : submitLabel}
        </Button>
      </div>
    </form>
  )
}
