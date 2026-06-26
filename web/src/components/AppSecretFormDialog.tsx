import { useState, type ReactNode } from "react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import type { AppSecret, AppSecretInput } from "@/lib/vault"

function normalizeIdentifier(value: string) {
  return value.trim().toLowerCase()
}

export function AppSecretFormDialog({
  appSecret,
  trigger,
  isPending,
  onSubmit,
}: {
  appSecret?: AppSecret
  trigger: ReactNode
  isPending: boolean
  onSubmit: (input: AppSecretInput) => Promise<unknown>
}) {
  const [open, setOpen] = useState(false)
  const [key, setKey] = useState(appSecret?.key ?? "")
  const [value, setValue] = useState("")
  const isEditing = !!appSecret

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen)
    if (nextOpen) return
    setValue("")
    if (!isEditing) {
      setKey("")
    }
  }

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    await onSubmit({
      key: normalizeIdentifier(key),
      value,
    })
    handleOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <form onSubmit={handleSubmit} className="space-y-4">
          <DialogHeader>
            <DialogTitle>{isEditing ? "Edit app secret" : "New app secret"}</DialogTitle>
          </DialogHeader>

          <div className="space-y-2">
            <Label htmlFor={isEditing ? `app-secret-key-${appSecret.id}` : "app-secret-key"}>
              Key
            </Label>
            <Input
              id={isEditing ? `app-secret-key-${appSecret.id}` : "app-secret-key"}
              value={key}
              onChange={(event) => setKey(event.target.value.toLowerCase())}
              placeholder="vehicle_id"
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor={isEditing ? `app-secret-value-${appSecret.id}` : "app-secret-value"}>
              Value
            </Label>
            <Textarea
              id={isEditing ? `app-secret-value-${appSecret.id}` : "app-secret-value"}
              value={value}
              onChange={(event) => setValue(event.target.value)}
              rows={4}
              required={!isEditing}
            />
          </div>

          <DialogFooter>
            <Button type="submit" disabled={isPending}>
              {isPending ? "Saving" : "Save"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
