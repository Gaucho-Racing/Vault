import { useMutation, useQueryClient } from "@tanstack/react-query"
import { ArrowLeft } from "lucide-react"
import { Link, useNavigate } from "react-router-dom"
import { toast } from "sonner"

import { AppSecretApplicationForm } from "@/components/AppSecretApplicationForm"
import { PageContainer } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { createAppSecretApplication, type AppSecretApplicationInput } from "@/lib/vault"

function errorMessage(error: unknown, fallback: string) {
  return (error as { response?: { data?: { error?: string } } })?.response?.data?.error ?? fallback
}

export default function AppSecretApplicationCreatePage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: createAppSecretApplication,
    onSuccess: (application) => {
      toast.success("Application created")
      void queryClient.invalidateQueries({ queryKey: ["appSecretApplications"] })
      navigate(`/app-secrets/${application.id}`, { replace: true })
    },
    onError: (error) => toast.error(errorMessage(error, "Failed to create application")),
  })

  async function handleCreate(input: AppSecretApplicationInput) {
    await createMutation.mutateAsync(input)
  }

  return (
    <PageContainer className="max-w-4xl">
      <div className="mb-6">
        <Button asChild variant="ghost" size="sm" className="-ml-2">
          <Link to="/app-secrets">
            <ArrowLeft className="size-4" />
            App Secrets
          </Link>
        </Button>
        <div className="mt-4">
          <h1 className="text-2xl font-semibold">New application</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Create an application namespace for encrypted development secrets.
          </p>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Application details</CardTitle>
        </CardHeader>
        <CardContent>
          <AppSecretApplicationForm
            isPending={createMutation.isPending}
            submitLabel="Create application"
            onSubmit={handleCreate}
            onCancel={() => navigate("/app-secrets")}
          />
        </CardContent>
      </Card>
    </PageContainer>
  )
}
