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
import { commonSecretTypes, type SecretInput } from "@/lib/vault"

export function SecretFormDialog({
  trigger,
  isPending,
  onSubmit,
}: {
  trigger: ReactNode
  isPending: boolean
  onSubmit: (input: SecretInput) => Promise<void>
}) {
  const [open, setOpen] = useState(false)
  const [key, setKey] = useState("")
  const [label, setLabel] = useState("")
  const [type, setType] = useState("")
  const [plainValue, setPlainValue] = useState("")
  const [sensitive, setSensitive] = useState(true)

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    await onSubmit({
      key,
      label,
      type,
      sensitive,
      plain_value: plainValue,
    })
    setOpen(false)
    setKey("")
    setLabel("")
    setType("")
    setPlainValue("")
    setSensitive(true)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <form onSubmit={handleSubmit} className="space-y-4">
          <DialogHeader>
            <DialogTitle>New secret</DialogTitle>
          </DialogHeader>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="secret-key">Key</Label>
              <Input
                id="secret-key"
                value={key}
                onChange={(event) => setKey(event.target.value)}
                placeholder="primary_password"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="secret-type">Type</Label>
              <Input
                id="secret-type"
                list="secret-types"
                value={type}
                onChange={(event) => setType(event.target.value)}
                placeholder="password"
              />
              <datalist id="secret-types">
                {commonSecretTypes.map((secretType) => (
                  <option key={secretType} value={secretType} />
                ))}
              </datalist>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="secret-label">Label</Label>
            <Input
              id="secret-label"
              value={label}
              onChange={(event) => setLabel(event.target.value)}
              placeholder="Primary password"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="secret-value">Value</Label>
            <Textarea
              id="secret-value"
              value={plainValue}
              onChange={(event) => setPlainValue(event.target.value)}
              rows={4}
              required={sensitive}
            />
          </div>

          <label className="flex items-center justify-between gap-3 rounded-lg bg-muted/40 px-3 py-2 text-sm dark:bg-muted/30">
            <span className="font-medium">Sensitive</span>
            <input
              type="checkbox"
              checked={sensitive}
              onChange={(event) => setSensitive(event.target.checked)}
              className="size-4 accent-gr-pink"
            />
          </label>

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
