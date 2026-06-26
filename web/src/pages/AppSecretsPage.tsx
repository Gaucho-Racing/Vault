import { useQuery } from "@tanstack/react-query"
import { ArrowUpRight, Clock3, CodeXml, Plus, Search } from "lucide-react"
import { useMemo, useState } from "react"
import { Link } from "react-router-dom"

import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { listAppSecretApplications, type AppSecretApplicationListItem } from "@/lib/vault"

function formatDate(value: string) {
  return new Date(value).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  })
}

function applicationMatches(application: AppSecretApplicationListItem, query: string) {
  const normalizedQuery = query.trim().toLowerCase()
  if (!normalizedQuery) return true
  return (
    application.name.toLowerCase().includes(normalizedQuery) ||
    application.access_group_names.some((group) => group.toLowerCase().includes(normalizedQuery))
  )
}

function secretCountLabel(count: number) {
  return `${count} secret${count === 1 ? "" : "s"}`
}

export default function AppSecretsPage() {
  const [search, setSearch] = useState("")

  const applicationsQuery = useQuery({
    queryKey: ["appSecretApplications"],
    queryFn: listAppSecretApplications,
  })

  const applications = useMemo(() => applicationsQuery.data ?? [], [applicationsQuery.data])
  const filteredApplications = useMemo(
    () => applications.filter((application) => applicationMatches(application, search)),
    [applications, search],
  )

  return (
    <PageContainer>
      <PageHeader
        title="App Secrets"
        description="Manage encrypted application keys and environment values."
        action={
          <Button asChild>
            <Link to="/app-secrets/new">
              <Plus className="size-4" />
              Application
            </Link>
          </Button>
        }
      />

      <div className="mb-6 flex h-11 max-w-xl min-w-0 items-center gap-2 rounded-lg bg-card px-3 shadow-sm shadow-black/[0.03] dark:shadow-black/20">
        <Search className="size-4 shrink-0 text-muted-foreground" />
        <Input
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          placeholder="Search applications"
          className="h-9 border-0 bg-transparent px-0 shadow-none focus-visible:ring-0"
        />
      </div>

      {applicationsQuery.isLoading ? (
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          {Array.from({ length: 6 }).map((_, index) => (
            <Skeleton key={index} className="h-44 rounded-lg" />
          ))}
        </div>
      ) : filteredApplications.length === 0 ? (
        <Card>
          <CardContent className="flex min-h-56 flex-col items-center justify-center py-10 text-center">
            <div className="flex size-10 items-center justify-center rounded-lg bg-muted">
              <Search className="size-5 text-muted-foreground" />
            </div>
            <div className="mt-4 text-sm font-medium">No applications found</div>
            <div className="mt-1 text-sm text-muted-foreground">
              Try a different search or create a new application.
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          {filteredApplications.map((application) => (
            <Link key={application.id} to={`/app-secrets/${application.id}`} className="group">
              <Card className="h-full gap-5 transition-all hover:-translate-y-0.5 hover:border-primary/25 hover:shadow-lg hover:shadow-black/5 dark:hover:border-primary/30 dark:hover:shadow-black/25">
                <CardHeader className="gap-3">
                  <CardTitle className="flex items-start justify-between gap-3">
                    <span className="min-w-0 truncate font-mono text-base font-semibold">
                      {application.name}
                    </span>
                    <span className="flex size-7 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground transition-colors group-hover:bg-primary group-hover:text-primary-foreground">
                      <ArrowUpRight className="size-4" />
                    </span>
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="flex items-center gap-2 rounded-lg bg-muted/60 px-2.5 py-2 text-xs">
                    <CodeXml className="size-3.5 shrink-0 text-primary" />
                    <span className="min-w-0 truncate">
                      {secretCountLabel(application.secret_count)}
                    </span>
                  </div>
                  <div className="flex min-h-6 flex-wrap gap-1.5">
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
                  <div className="flex items-center gap-1.5 border-t border-border/50 pt-3 text-xs text-muted-foreground">
                    <Clock3 className="size-3.5" />
                    Updated {formatDate(application.updated_at)}
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </PageContainer>
  )
}
