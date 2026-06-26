import { Navigate, createBrowserRouter } from "react-router-dom"

import { AppShell } from "@/components/AppShell"
import { RequireAuth } from "@/components/RequireAuth"
import AccountCreatePage from "@/pages/AccountCreatePage"
import AccountDetailsPage from "@/pages/AccountDetailsPage"
import AccountsPage from "@/pages/AccountsPage"
import AppSecretApplicationCreatePage from "@/pages/AppSecretApplicationCreatePage"
import AppSecretApplicationDetailsPage from "@/pages/AppSecretApplicationDetailsPage"
import AppSecretsPage from "@/pages/AppSecretsPage"
import LoginPage from "@/pages/LoginPage"
import NotFoundPage from "@/pages/NotFoundPage"
import SettingsPage from "@/pages/SettingsPage"

export const router = createBrowserRouter([
  {
    element: <RequireAuth />,
    children: [
      {
        element: <AppShell />,
        children: [
          { path: "/", element: <Navigate to="/accounts" replace /> },
          { path: "/accounts", element: <AccountsPage /> },
          { path: "/accounts/new", element: <AccountCreatePage /> },
          { path: "/accounts/:id", element: <AccountDetailsPage /> },
          { path: "/app-secrets", element: <AppSecretsPage /> },
          { path: "/app-secrets/new", element: <AppSecretApplicationCreatePage /> },
          { path: "/app-secrets/:id", element: <AppSecretApplicationDetailsPage /> },
          { path: "/settings", element: <SettingsPage /> },
        ],
      },
    ],
  },
  { path: "/auth/login", element: <LoginPage /> },
  { path: "*", element: <NotFoundPage /> },
])
