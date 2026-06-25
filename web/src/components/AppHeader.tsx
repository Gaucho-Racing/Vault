import { KeyRound, LogOut, Menu, Settings } from "lucide-react"
import { Link, useLocation, useNavigate } from "react-router-dom"

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
import { Skeleton } from "@/components/ui/skeleton"
import { useAuth } from "@/lib/auth"
import { cn } from "@/lib/utils"

const mobileItems = [
  { to: "/accounts", label: "Accounts", icon: KeyRound },
  { to: "/settings", label: "Settings", icon: Settings },
]

function sectionTitle(pathname: string) {
  if (pathname.startsWith("/settings")) return "Settings"
  if (pathname.startsWith("/accounts")) return "Accounts"
  return "Vault"
}

function avatarLabel(value?: string) {
  const cleaned = value?.replace(/[^a-z0-9]/gi, "") ?? ""
  return (cleaned.slice(0, 2) || "VA").toUpperCase()
}

function shortID(value?: string) {
  if (!value) return "Unknown"
  if (value.length <= 18) return value
  return `${value.slice(0, 10)}...${value.slice(-6)}`
}

function HeaderUserMenu() {
  const navigate = useNavigate()
  const { session, isLoading, logout } = useAuth()

  if (isLoading || !session) {
    return <Skeleton className="size-8 rounded-full" />
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button className="rounded-full outline-none ring-offset-background focus-visible:ring-2 focus-visible:ring-ring/35 focus-visible:ring-offset-2">
          <Avatar className="size-8 cursor-pointer">
            <AvatarFallback>{avatarLabel(session.entity_id)}</AvatarFallback>
          </Avatar>
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" sideOffset={10} className="w-64">
        <DropdownMenuLabel className="flex flex-col">
          <span className="text-sm font-medium">Sentinel entity</span>
          <span className="font-mono text-xs font-normal text-muted-foreground">
            {shortID(session.entity_id)}
          </span>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem onSelect={() => navigate("/settings")}>
          <Settings className="size-4" />
          Settings
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onSelect={logout} className="text-destructive focus:text-destructive">
          <LogOut className="size-4" />
          Sign out
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

export function AppHeader() {
  const { pathname } = useLocation()
  const section = sectionTitle(pathname)

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
        <div className="text-lg font-semibold leading-none">{section}</div>
      </div>

      <div className="flex-1" />

      <HeaderUserMenu />
    </header>
  )
}
