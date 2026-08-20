import { Navigate, Outlet, useLocation } from "react-router-dom"

import { loadSession, saveLoginReturnFrom } from "@/lib/auth"

export function RequireAuth() {
  const location = useLocation()
  const session = loadSession()
  if (!session) {
    saveLoginReturnFrom(location)
    return <Navigate to="/auth/login" state={{ from: location }} replace />
  }
  return <Outlet />
}
