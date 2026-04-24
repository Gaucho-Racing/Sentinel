import { createBrowserRouter } from "react-router-dom"

import ApplicationsPage from "@/pages/applications/ApplicationsPage"
import LoginPage from "@/pages/auth/LoginPage"
import HomePage from "@/pages/HomePage"
import NotFoundPage from "@/pages/NotFoundPage"
import AuthorizePage from "@/pages/oauth/AuthorizePage"

export const router = createBrowserRouter([
  { path: "/", element: <HomePage /> },
  { path: "/auth/login", element: <LoginPage /> },
  { path: "/oauth/authorize", element: <AuthorizePage /> },
  { path: "/applications", element: <ApplicationsPage /> },
  { path: "*", element: <NotFoundPage /> },
])
