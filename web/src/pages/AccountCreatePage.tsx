import { useMutation, useQueryClient } from "@tanstack/react-query"
import { ArrowLeft } from "lucide-react"
import { Link, useNavigate } from "react-router-dom"
import { toast } from "sonner"

import { AccountForm } from "@/components/AccountForm"
import { PageContainer } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { createAccount, type AccountInput } from "@/lib/vault"

export default function AccountCreatePage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: createAccount,
    onSuccess: (account) => {
      toast.success("Account created")
      void queryClient.invalidateQueries({ queryKey: ["accounts"] })
      navigate(`/accounts/${account.id}`, { replace: true })
    },
    onError: (error) => {
      const message =
        (error as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Failed to create account"
      toast.error(message)
    },
  })

  async function handleCreate(input: AccountInput) {
    await createMutation.mutateAsync(input)
  }

  return (
    <PageContainer className="max-w-4xl">
      <div className="mb-6">
        <Button asChild variant="ghost" size="sm" className="-ml-2">
          <Link to="/accounts">
            <ArrowLeft className="size-4" />
            Accounts
          </Link>
        </Button>
        <div className="mt-4">
          <h1 className="text-2xl font-semibold">New account</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Add a shared login, API key set, or service account.
          </p>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Account details</CardTitle>
        </CardHeader>
        <CardContent>
          <AccountForm
            isPending={createMutation.isPending}
            submitLabel="Create account"
            onSubmit={handleCreate}
            onCancel={() => navigate("/accounts")}
          />
        </CardContent>
      </Card>
    </PageContainer>
  )
}
