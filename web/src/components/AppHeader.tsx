import { KeyRound, LogOut, Menu } from "lucide-react"
import { Link, useLocation } from "react-router-dom"

import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useAuth } from "@/lib/auth"
import { cn } from "@/lib/utils"

const mobileItems = [{ to: "/accounts", label: "Accounts", icon: KeyRound }]

export function AppHeader() {
  const { pathname } = useLocation()
  const { logout } = useAuth()

  return (
    <header className="sticky top-0 z-30 flex h-14 items-center gap-3 border-b bg-background/85 px-4 backdrop-blur supports-[backdrop-filter]:bg-background/65 lg:px-6">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon" className="lg:hidden">
            <Menu className="size-4" />
            <span className="sr-only">Open navigation</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="w-52">
          <DropdownMenuLabel>Vault</DropdownMenuLabel>
          <DropdownMenuSeparator />
          {mobileItems.map((item) => (
            <DropdownMenuItem key={item.to} asChild>
              <Link
                to={item.to}
                className={cn(pathname.startsWith(item.to) && "text-gr-pink")}
              >
                <item.icon className="size-4" />
                {item.label}
              </Link>
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>

      <Link to="/accounts" className="flex items-center gap-2 lg:hidden">
        <img src="/favicon.svg" alt="Vault" className="size-8 rounded-lg" />
        <span className="text-sm font-semibold">Vault</span>
      </Link>

      <div className="flex-1" />

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <button className="rounded-full outline-none ring-offset-background focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2">
            <Avatar className="size-8 cursor-pointer">
              <AvatarFallback>V</AvatarFallback>
            </Avatar>
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" sideOffset={10} className="w-56">
          <DropdownMenuLabel className="flex flex-col">
            <span className="text-sm font-medium">Sentinel session</span>
            <span className="text-xs font-normal text-muted-foreground">Vault API access</span>
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem onSelect={logout} className="text-destructive focus:text-destructive">
            <LogOut className="size-4" />
            Sign out
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </header>
  )
}
