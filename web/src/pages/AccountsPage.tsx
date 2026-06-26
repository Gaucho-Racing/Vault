import { useQuery } from "@tanstack/react-query"
import { ArrowUpRight, Clock3, KeyRound, Lock, Plus, Search } from "lucide-react"
import { useMemo, useState } from "react"
import { Link } from "react-router-dom"

import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { listAccounts } from "@/lib/vault"
import { cn } from "@/lib/utils"

function formatDate(value: string) {
  return new Date(value).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  })
}

function accountMatches(accountName: string, query: string) {
  return accountName.toLowerCase().includes(query.trim().toLowerCase())
}

function secretCountLabel(count: number) {
  return `${count} secret${count === 1 ? "" : "s"}`
}

export default function AccountsPage() {
  const [search, setSearch] = useState("")

  const accountsQuery = useQuery({
    queryKey: ["accounts"],
    queryFn: listAccounts,
  })

  const accounts = useMemo(() => accountsQuery.data ?? [], [accountsQuery.data])
  const filteredAccounts = useMemo(
    () => accounts.filter((account) => accountMatches(account.name, search)),
    [accounts, search],
  )

  return (
    <PageContainer>
      <PageHeader
        title="Accounts"
        description="Manage shared credentials, API keys, and service logins."
        action={
          <Button asChild>
            <Link to="/accounts/new">
              <Plus className="size-4" />
              Account
            </Link>
          </Button>
        }
      />

      <div className="mb-6 flex h-11 max-w-xl min-w-0 items-center gap-2 rounded-lg bg-card px-3 shadow-sm shadow-black/[0.03] dark:shadow-black/20">
        <Search className="size-4 shrink-0 text-muted-foreground" />
        <Input
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          placeholder="Search accounts"
          className="h-9 border-0 bg-transparent px-0 shadow-none focus-visible:ring-0"
        />
      </div>

      {accountsQuery.isLoading ? (
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          {Array.from({ length: 6 }).map((_, index) => (
            <Skeleton key={index} className="h-44 rounded-lg" />
          ))}
        </div>
      ) : filteredAccounts.length === 0 ? (
        <Card>
          <CardContent className="flex min-h-56 flex-col items-center justify-center py-10 text-center">
            <div className="flex size-10 items-center justify-center rounded-lg bg-muted">
              <Search className="size-5 text-muted-foreground" />
            </div>
            <div className="mt-4 text-sm font-medium">No accounts found</div>
            <div className="mt-1 text-sm text-muted-foreground">
              Try a different search or create a new account.
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          {filteredAccounts.map((account) => {
            const card = (
              <Card
                className={cn(
                  "h-full gap-5 transition-all",
                  account.can_access
                    ? "hover:-translate-y-0.5 hover:border-primary/25 hover:shadow-lg hover:shadow-black/5 dark:hover:border-primary/30 dark:hover:shadow-black/25"
                    : "opacity-60",
                )}
              >
                <CardHeader className="gap-3">
                  <CardTitle className="flex items-start justify-between gap-3">
                    <span className="min-w-0 truncate text-base font-semibold">{account.name}</span>
                    <span
                      className={cn(
                        "flex size-7 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground",
                        account.can_access && "transition-colors group-hover:bg-primary group-hover:text-primary-foreground",
                      )}
                    >
                      {account.can_access ? <ArrowUpRight className="size-4" /> : <Lock className="size-3.5" />}
                    </span>
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  <p className="line-clamp-2 text-sm leading-5 text-muted-foreground">
                    {account.description || "No description"}
                  </p>
                  <div className="flex items-center gap-2 rounded-lg bg-muted/60 px-2.5 py-2 text-xs">
                    <KeyRound className="size-3.5 shrink-0 text-primary" />
                    <span className="min-w-0 truncate">{secretCountLabel(account.secret_count)}</span>
                  </div>
                  <div className="flex min-h-6 flex-wrap gap-1.5">
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
                  <div className="flex items-center gap-1.5 border-t border-border/50 pt-3 text-xs text-muted-foreground">
                    <Clock3 className="size-3.5" />
                    Updated {formatDate(account.updated_at)}
                  </div>
                </CardContent>
              </Card>
            )

            return account.can_access ? (
              <Link key={account.id} to={`/accounts/${account.id}`} className="group">
                {card}
              </Link>
            ) : (
              <div key={account.id} className="cursor-not-allowed" title="You don't have access to this account">
                {card}
              </div>
            )
          })}
        </div>
      )}
    </PageContainer>
  )
}
