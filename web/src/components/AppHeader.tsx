import { KeyRound, LogOut, Menu, ShieldCheck } from "lucide-react"
import { Link, useLocation } from "react-router-dom"

import { ThemeToggle } from "@/components/ThemeToggle"
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
  const section = pathname.startsWith("/accounts") ? "Accounts" : "Vault"

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center gap-3 border-b border-border/60 bg-background/90 px-4 backdrop-blur supports-[backdrop-filter]:bg-background/75 lg:px-6">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="icon" className="lg:hidden">
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
        <div className="flex size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-sm shadow-primary/20">
          <KeyRound className="size-4" />
        </div>
        <span className="text-sm font-semibold">Vault</span>
      </Link>

      <div className="hidden min-w-0 lg:block">
        <div className="flex items-center gap-2 text-xs font-medium uppercase text-muted-foreground">
          <ShieldCheck className="size-3.5 text-primary" />
          Gaucho Racing Vault
        </div>
        <div className="mt-1 text-lg font-semibold leading-none">{section}</div>
      </div>

      <div className="flex-1" />

      <ThemeToggle />

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <button className="rounded-full outline-none ring-offset-background focus-visible:ring-2 focus-visible:ring-ring/35 focus-visible:ring-offset-2">
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
