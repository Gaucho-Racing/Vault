import { useEffect, useState } from "react"

export type Theme = "light" | "dark"

const STORAGE_KEY = "vault_theme"
const EVENT_NAME = "vault-theme-change"

function systemTheme(): Theme {
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light"
}

export function currentTheme(): Theme {
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored === "light" || stored === "dark") return stored
  return systemTheme()
}

function applyTheme(theme: Theme) {
  document.documentElement.classList.toggle("dark", theme === "dark")
  document.documentElement.style.colorScheme = theme
}

export function setTheme(theme: Theme) {
  localStorage.setItem(STORAGE_KEY, theme)
  applyTheme(theme)
  window.dispatchEvent(new CustomEvent(EVENT_NAME, { detail: theme }))
}

export function useTheme() {
  const [theme, setThemeState] = useState<Theme>(() => currentTheme())

  useEffect(() => {
    applyTheme(theme)

    function handleThemeChange() {
      setThemeState(currentTheme())
    }

    window.addEventListener(EVENT_NAME, handleThemeChange)
    window.addEventListener("storage", handleThemeChange)
    return () => {
      window.removeEventListener(EVENT_NAME, handleThemeChange)
      window.removeEventListener("storage", handleThemeChange)
    }
  }, [theme])

  function updateTheme(nextTheme: Theme) {
    setThemeState(nextTheme)
    setTheme(nextTheme)
  }

  function toggleTheme() {
    updateTheme(theme === "dark" ? "light" : "dark")
  }

  return { theme, setTheme: updateTheme, toggleTheme }
}
