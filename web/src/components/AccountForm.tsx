import { useQuery } from "@tanstack/react-query"
import { Globe2, Plus, Search, ShieldCheck, X } from "lucide-react"
import { useMemo, useState } from "react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Skeleton } from "@/components/ui/skeleton"
import { Textarea } from "@/components/ui/textarea"
import { cn } from "@/lib/utils"
import { listSentinelGroups, type Account, type AccountInput, type SentinelGroup } from "@/lib/vault"

function sortGroups(groups: SentinelGroup[]) {
  return [...groups].sort((a, b) => a.name.localeCompare(b.name))
}

function normalizeGroups(groups: string[]) {
  return Array.from(new Set(groups.map((group) => group.trim()).filter(Boolean)))
}

function groupMatches(group: SentinelGroup, query: string) {
  const normalizedQuery = query.trim().toLowerCase()
  if (!normalizedQuery) return true
  return (
    group.name.toLowerCase().includes(normalizedQuery) ||
    (group.description ?? "").toLowerCase().includes(normalizedQuery)
  )
}

function AccessSummary({ restricted }: { restricted: boolean }) {
  const Icon = restricted ? ShieldCheck : Globe2

  return (
    <Card className={cn("border-0 bg-muted/45 shadow-none", restricted && "bg-primary/10")}>
      <CardContent className="flex items-start gap-3 p-3">
        <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-background text-primary shadow-sm shadow-black/[0.03] dark:shadow-black/20">
          <Icon className="size-4" />
        </div>
        <div className="min-w-0">
          <div className="text-sm font-medium">{restricted ? "Restricted" : "Public"}</div>
          <div className="mt-1 text-sm leading-5 text-muted-foreground">
            {restricted
              ? "Only selected Sentinel groups can access this account. Admins always have access."
              : "All Sentinel users can access this account. Add groups to limit access."}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function GroupPicker({
  selectedGroups,
  onChange,
}: {
  selectedGroups: string[]
  onChange: (groups: string[]) => void
}) {
  const [search, setSearch] = useState("")
  const groupsQuery = useQuery({
    queryKey: ["sentinelGroups"],
    queryFn: listSentinelGroups,
    staleTime: 5 * 60 * 1000,
  })

  const selectedSet = useMemo(() => new Set(selectedGroups), [selectedGroups])
  const filteredGroups = useMemo(() => {
    const groups = sortGroups(groupsQuery.data ?? [])
    return groups
      .filter((group) => !selectedSet.has(group.name))
      .filter((group) => groupMatches(group, search))
      .slice(0, 8)
  }, [groupsQuery.data, search, selectedSet])

  function addGroup(groupName: string) {
    onChange(normalizeGroups([...selectedGroups, groupName]))
    setSearch("")
  }

  function removeGroup(groupName: string) {
    onChange(selectedGroups.filter((group) => group !== groupName))
  }

  return (
    <div className="space-y-3">
      <div className="space-y-2">
        <Label htmlFor="account-groups">Sentinel groups</Label>
        <div className="relative">
          <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            id="account-groups"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search groups"
            className="pl-9"
          />
        </div>
      </div>

      <div className="rounded-lg bg-muted/35 p-2">
        {groupsQuery.isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 3 }).map((_, index) => (
              <Skeleton key={index} className="h-9 rounded-md" />
            ))}
          </div>
        ) : groupsQuery.isError ? (
          <div className="px-2 py-3 text-sm text-muted-foreground">
            Could not load Sentinel groups.
          </div>
        ) : filteredGroups.length === 0 ? (
          <div className="px-2 py-3 text-sm text-muted-foreground">
            {search.trim() ? "No matching groups." : "No more groups available."}
          </div>
        ) : (
          <div className="max-h-72 space-y-1 overflow-y-auto">
            {filteredGroups.map((group) => (
              <button
                key={group.id}
                type="button"
                onClick={() => addGroup(group.name)}
                className="flex w-full items-center gap-3 rounded-md px-2 py-2 text-left text-sm transition-colors hover:bg-background focus-visible:ring-2 focus-visible:ring-ring/25 focus-visible:outline-none"
              >
                <div className="min-w-0 flex-1">
                  <div className="truncate font-medium">{group.name}</div>
                  {group.description && (
                    <div className="mt-0.5 line-clamp-1 text-xs text-muted-foreground">
                      {group.description}
                    </div>
                  )}
                </div>
                <Plus className="size-4 shrink-0 text-primary" />
              </button>
            ))}
          </div>
        )}
      </div>

      <div className="flex min-h-8 flex-wrap gap-1.5">
        {selectedGroups.length === 0 ? (
          <Badge variant="secondary">Public by default</Badge>
        ) : (
          selectedGroups.map((group) => (
            <Badge key={group} variant="secondary" className="gap-1.5 pr-1">
              {group}
              <button
                type="button"
                onClick={() => removeGroup(group)}
                className="rounded-sm p-0.5 hover:bg-background/80 focus-visible:ring-2 focus-visible:ring-ring/25 focus-visible:outline-none"
              >
                <X className="size-3" />
                <span className="sr-only">Remove {group}</span>
              </button>
            </Badge>
          ))
        )}
      </div>
    </div>
  )
}

export function AccountForm({
  account,
  isPending,
  submitLabel = "Save account",
  onSubmit,
  onCancel,
}: {
  account?: Account
  isPending: boolean
  submitLabel?: string
  onSubmit: (input: AccountInput) => Promise<void>
  onCancel?: () => void
}) {
  const [name, setName] = useState(account?.name ?? "")
  const [description, setDescription] = useState(account?.description ?? "")
  const [url, setURL] = useState(account?.url ?? "")
  const [groups, setGroups] = useState(() => normalizeGroups(account?.access_group_names ?? []))

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    await onSubmit({
      name,
      description,
      url,
      access_group_names: groups,
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="account-name">Name</Label>
          <Input
            id="account-name"
            value={name}
            onChange={(event) => setName(event.target.value)}
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="account-url">URL</Label>
          <Input
            id="account-url"
            type="url"
            value={url}
            onChange={(event) => setURL(event.target.value)}
            placeholder="https://"
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="account-description">Description</Label>
        <Textarea
          id="account-description"
          value={description}
          onChange={(event) => setDescription(event.target.value)}
          rows={3}
        />
      </div>

      <section className="space-y-3">
        <div>
          <h2 className="text-sm font-medium">Access</h2>
          <p className="mt-1 text-sm leading-5 text-muted-foreground">
            Accounts are public by default. Select Sentinel groups to limit access.
          </p>
        </div>
        <AccessSummary restricted={groups.length > 0} />
        <GroupPicker selectedGroups={groups} onChange={setGroups} />
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
