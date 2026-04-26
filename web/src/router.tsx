import { createBrowserRouter } from "react-router-dom"

import { AppShell } from "@/components/AppShell"
import AnalyticsPage from "@/pages/analytics/AnalyticsPage"
import ApplicationsPage from "@/pages/applications/ApplicationsPage"
import LoginPage from "@/pages/auth/LoginPage"
import GroupsPage from "@/pages/groups/GroupsPage"
import HomePage from "@/pages/HomePage"
import NotFoundPage from "@/pages/NotFoundPage"
import AuthorizePage from "@/pages/oauth/AuthorizePage"
import SettingsPage from "@/pages/settings/SettingsPage"

export const router = createBrowserRouter([
  {
    element: <AppShell />,
    children: [
      { path: "/", element: <HomePage /> },
      { path: "/applications", element: <ApplicationsPage /> },
      { path: "/groups", element: <GroupsPage /> },
      { path: "/analytics", element: <AnalyticsPage /> },
      { path: "/settings", element: <SettingsPage /> },
    ],
  },
  { path: "/auth/login", element: <LoginPage /> },
  { path: "/oauth/authorize", element: <AuthorizePage /> },
  { path: "*", element: <NotFoundPage /> },
])
