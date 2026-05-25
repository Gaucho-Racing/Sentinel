import { createBrowserRouter } from "react-router-dom"

import { AppShell } from "@/components/AppShell"
import { RequireAuth } from "@/components/RequireAuth"
import AnalyticsPage from "@/pages/analytics/AnalyticsPage"
import ApplicationDetailsPage from "@/pages/applications/ApplicationDetailsPage"
import ApplicationEditPage from "@/pages/applications/ApplicationEditPage"
import ApplicationNewPage from "@/pages/applications/ApplicationNewPage"
import ApplicationsPage from "@/pages/applications/ApplicationsPage"
import LoginPage from "@/pages/auth/LoginPage"
import DebugPage from "@/pages/debug/DebugPage"
import GroupDetailsPage from "@/pages/groups/GroupDetailsPage"
import GroupsPage from "@/pages/groups/GroupsPage"
import HomePage from "@/pages/HomePage"
import NotFoundPage from "@/pages/NotFoundPage"
import AuthorizePage from "@/pages/oauth/AuthorizePage"
import OnboardingPage from "@/pages/onboarding/OnboardingPage"
import SettingsPage from "@/pages/settings/SettingsPage"

export const router = createBrowserRouter([
  {
    element: <RequireAuth />,
    children: [
      {
        element: <AppShell />,
        children: [
          { path: "/", element: <HomePage /> },
          { path: "/applications", element: <ApplicationsPage /> },
          { path: "/applications/new", element: <ApplicationNewPage /> },
          { path: "/applications/:id", element: <ApplicationDetailsPage /> },
          { path: "/applications/:id/edit", element: <ApplicationEditPage /> },
          { path: "/groups", element: <GroupsPage /> },
          { path: "/groups/:id", element: <GroupDetailsPage /> },
          { path: "/analytics", element: <AnalyticsPage /> },
          { path: "/settings", element: <SettingsPage /> },
          { path: "/debug", element: <DebugPage /> },
        ],
      },
    ],
  },
  { path: "/auth/login", element: <LoginPage /> },
  { path: "/oauth/authorize", element: <AuthorizePage /> },
  { path: "/onboard", element: <OnboardingPage /> },
  { path: "*", element: <NotFoundPage /> },
])
