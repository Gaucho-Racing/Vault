import { QrCode } from "lucide-react"
import { useState, type ReactNode } from "react"
import { toast } from "sonner"

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
import { commonSecretTypes, decodeTOTPRegistrationQRCode, type SecretInput } from "@/lib/vault"

const CUSTOM_SECRET_TYPE = "__custom__"
const TOTP_SECRET_TYPE = "totp_seed"

function getClipboardImageFile(event: React.ClipboardEvent) {
  for (const item of Array.from(event.clipboardData.items)) {
    if (!item.type.startsWith("image/")) continue
    const file = item.getAsFile()
    if (file) return file
  }
  return null
}

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
  const [isDecodingQRCode, setIsDecodingQRCode] = useState(false)
  const selectedType = commonSecretTypes.includes(type) ? type : isCustomType ? CUSTOM_SECRET_TYPE : undefined
  const hasTOTPRegistrationURL = plainValue.trim().toLowerCase().startsWith("otpauth://totp/")

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
    setIsDecodingQRCode(false)
  }

  async function handlePaste(event: React.ClipboardEvent) {
    const imageFile = getClipboardImageFile(event)
    if (!imageFile) return

    event.preventDefault()
    setIsDecodingQRCode(true)
    try {
      const totpRegistration = await decodeTOTPRegistrationQRCode(imageFile)
      setType(TOTP_SECRET_TYPE)
      setIsCustomType(false)
      setSensitive(true)
      setPlainValue(totpRegistration.value)
      setKey((current) => current || totpRegistration.suggested_key)
      setLabel((current) => current || totpRegistration.suggested_label)
      toast.success("TOTP QR loaded")
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Failed to read TOTP QR")
    } finally {
      setIsDecodingQRCode(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <form onSubmit={handleSubmit} onPaste={(event) => void handlePaste(event)} className="space-y-4">
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

          {type === TOTP_SECRET_TYPE && (
            <div
              tabIndex={0}
              className="flex items-center gap-3 rounded-lg border border-dashed border-gr-pink/35 bg-gr-pink/8 px-3 py-2.5 outline-none transition-colors focus-visible:border-gr-purple focus-visible:ring-3 focus-visible:ring-gr-purple/20 dark:bg-gr-pink/12"
            >
              <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-background text-gr-pink shadow-sm shadow-black/[0.03] dark:shadow-black/20">
                <QrCode className="size-4" />
              </div>
              <div className="min-w-0">
                <div className="text-sm font-medium">
                  {isDecodingQRCode ? "Scanning QR" : "Paste QR screenshot"}
                </div>
                <div className="truncate text-xs text-muted-foreground">
                  {hasTOTPRegistrationURL ? "Registration QR loaded" : "TOTP registration"}
                </div>
              </div>
            </div>
          )}

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
