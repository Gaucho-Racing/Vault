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
import type { Account, AccountInput } from "@/lib/vault"

function parseGroups(value: string) {
  return value
    .split(/[,\n]/)
    .map((group) => group.trim())
    .filter(Boolean)
}

function groupsToText(groups: string[]) {
  return groups.join(", ")
}

export function AccountFormDialog({
  account,
  trigger,
  isPending,
  onSubmit,
}: {
  account?: Account
  trigger: ReactNode
  isPending: boolean
  onSubmit: (input: AccountInput) => Promise<void>
}) {
  const [open, setOpen] = useState(false)
  const [name, setName] = useState(account?.name ?? "")
  const [description, setDescription] = useState(account?.description ?? "")
  const [url, setURL] = useState(account?.url ?? "")
  const [groups, setGroups] = useState(groupsToText(account?.access_group_names ?? []))

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    await onSubmit({
      name,
      description,
      url,
      access_group_names: parseGroups(groups),
    })
    setOpen(false)
    if (!account) {
      setName("")
      setDescription("")
      setURL("")
      setGroups("")
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <form onSubmit={handleSubmit} className="space-y-4">
          <DialogHeader>
            <DialogTitle>{account ? "Edit account" : "New account"}</DialogTitle>
          </DialogHeader>

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

          <div className="space-y-2">
            <Label htmlFor="account-description">Description</Label>
            <Textarea
              id="account-description"
              value={description}
              onChange={(event) => setDescription(event.target.value)}
              rows={3}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="account-groups">Sentinel groups</Label>
            <Input
              id="account-groups"
              value={groups}
              onChange={(event) => setGroups(event.target.value)}
              placeholder="Admins, SocialManagers"
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
