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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { commonSecretTypes, type SecretInput } from "@/lib/vault"

const CUSTOM_SECRET_TYPE = "__custom__"

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
  const [isCustomType, setIsCustomType] = useState(false)
  const [plainValue, setPlainValue] = useState("")
  const [sensitive, setSensitive] = useState(true)
  const selectedType = commonSecretTypes.includes(type) ? type : isCustomType ? CUSTOM_SECRET_TYPE : undefined

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
    setIsCustomType(false)
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
              <Select
                value={selectedType}
                onValueChange={(value) => {
                  if (value === CUSTOM_SECRET_TYPE) {
                    setType("")
                    setIsCustomType(true)
                    return
                  }
                  setType(value)
                  setIsCustomType(false)
                }}
              >
                <SelectTrigger id="secret-type" className="h-9 w-full bg-background">
                  <SelectValue placeholder="Select type" />
                </SelectTrigger>
                <SelectContent position="popper" className="w-(--radix-select-trigger-width)">
                  {commonSecretTypes.map((secretType) => (
                    <SelectItem key={secretType} value={secretType}>
                      {secretType}
                    </SelectItem>
                  ))}
                  <SelectItem value={CUSTOM_SECRET_TYPE}>Custom</SelectItem>
                </SelectContent>
              </Select>
              {isCustomType && (
                <Input
                  value={type}
                  onChange={(event) => setType(event.target.value)}
                  placeholder="custom_type"
                  autoFocus
                />
              )}
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
