import { Navigate, Outlet, useLocation } from "react-router-dom"

import { loadSession } from "@/lib/auth"

export function RequireAuth() {
  const location = useLocation()
  const session = loadSession()
  if (!session) {
    return <Navigate to="/auth/login" state={{ from: location }} replace />
  }
  return <Outlet />
}
