import { Navigate, Outlet, useLocation } from "react-router-dom"

import { needsEmailVerification, saveLoginReturnFrom, useAuth } from "@/lib/auth"

export function RequireVerifiedEmail() {
  const location = useLocation()
  const { user, isLoading } = useAuth()
  if (isLoading) return null
  if (needsEmailVerification(user)) {
    saveLoginReturnFrom(location)
    return <Navigate to="/verify-email" replace />
  }
  return <Outlet />
}
