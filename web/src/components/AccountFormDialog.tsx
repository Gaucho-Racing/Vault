import { useState, type ReactNode } from "react"

import { AccountForm } from "@/components/AccountForm"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import type { Account, AccountInput } from "@/lib/vault"

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

  async function handleSubmit(input: AccountInput) {
    await onSubmit(input)
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{account ? "Edit account" : "New account"}</DialogTitle>
        </DialogHeader>
        <AccountForm
          account={account}
          isPending={isPending}
          submitLabel="Save"
          onSubmit={handleSubmit}
        />
      </DialogContent>
    </Dialog>
  )
}
