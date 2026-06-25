import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { ArrowUpRight, Plus, Search } from "lucide-react"
import { useMemo, useState } from "react"
import { Link } from "react-router-dom"
import { toast } from "sonner"

import { AccountFormDialog } from "@/components/AccountFormDialog"
import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { createAccount, listAccounts, type AccountInput } from "@/lib/vault"

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

export default function AccountsPage() {
  const queryClient = useQueryClient()
  const [search, setSearch] = useState("")

  const accountsQuery = useQuery({
    queryKey: ["accounts"],
    queryFn: listAccounts,
  })

  const createMutation = useMutation({
    mutationFn: createAccount,
    onSuccess: () => {
      toast.success("Account created")
      void queryClient.invalidateQueries({ queryKey: ["accounts"] })
    },
    onError: (error) => {
      const message =
        (error as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Failed to create account"
      toast.error(message)
    },
  })

  const accounts = useMemo(() => accountsQuery.data ?? [], [accountsQuery.data])
  const filteredAccounts = useMemo(
    () => accounts.filter((account) => accountMatches(account.name, search)),
    [accounts, search],
  )

  async function handleCreate(input: AccountInput) {
    await createMutation.mutateAsync(input)
  }

  return (
    <PageContainer>
      <PageHeader
        title="Accounts"
        description={`${accounts.length} account${accounts.length === 1 ? "" : "s"}`}
        action={
          <AccountFormDialog
            isPending={createMutation.isPending}
            onSubmit={handleCreate}
            trigger={
              <Button>
                <Plus className="size-4" />
                Account
              </Button>
            }
          />
        }
      />

      <div className="mb-5 flex max-w-sm items-center gap-2 rounded-lg border bg-background px-3">
        <Search className="size-4 text-muted-foreground" />
        <Input
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          placeholder="Search accounts"
          className="border-0 px-0 shadow-none focus-visible:ring-0"
        />
      </div>

      {accountsQuery.isLoading ? (
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          {Array.from({ length: 6 }).map((_, index) => (
            <Skeleton key={index} className="h-40 rounded-xl" />
          ))}
        </div>
      ) : filteredAccounts.length === 0 ? (
        <Card>
          <CardContent className="py-10 text-center text-sm text-muted-foreground">
            No accounts found.
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          {filteredAccounts.map((account) => (
            <Link key={account.id} to={`/accounts/${account.id}`}>
              <Card className="h-full transition-colors hover:bg-muted/40">
                <CardHeader>
                  <CardTitle className="flex items-start justify-between gap-3">
                    <span className="min-w-0 truncate">{account.name}</span>
                    <ArrowUpRight className="size-4 shrink-0 text-muted-foreground" />
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <p className="line-clamp-2 min-h-10 text-sm text-muted-foreground">
                    {account.description || "No description"}
                  </p>
                  <div className="flex flex-wrap gap-1.5">
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
                  <div className="text-xs text-muted-foreground">
                    Updated {formatDate(account.updated_at)}
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
