import { Monitor, Moon, Sun, type LucideIcon } from "lucide-react"

import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { useTheme, type Theme } from "@/lib/theme"

const themeOptions: Array<{ value: Theme; label: string; Icon: LucideIcon }> = [
  { value: "system", label: "System", Icon: Monitor },
  { value: "light", label: "Light", Icon: Sun },
  { value: "dark", label: "Dark", Icon: Moon },
]

export default function SettingsPage() {
  const { theme, setTheme } = useTheme()

  return (
    <PageContainer>
      <PageHeader title="Settings" />

      <Card>
        <CardHeader>
          <CardTitle>Appearance</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-2">
            {themeOptions.map(({ value, label, Icon }) => (
              <Button
                key={value}
                variant={theme === value ? "default" : "secondary"}
                onClick={() => setTheme(value)}
                aria-pressed={theme === value}
              >
                <Icon className="size-4" />
                {label}
              </Button>
            ))}
          </div>
        </CardContent>
      </Card>
    </PageContainer>
  )
}
