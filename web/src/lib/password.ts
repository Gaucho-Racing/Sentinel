// Mirrors core/service/email.go:ValidatePassword. Keep these rules in sync.
export function validatePassword(password: string): string | null {
  if (password.length < 8) return "Password must be at least 8 characters"
  if (password.length > 64) return "Password must be at most 64 characters"
  if (!/\d/.test(password)) return "Password must contain at least one number"
  if (!/[A-Z]/.test(password)) return "Password must contain at least one capital letter"
  return null
}
