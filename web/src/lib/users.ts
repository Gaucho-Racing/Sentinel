import { useQuery } from "@tanstack/react-query"

import { api } from "@/lib/api"

export type UserOption = {
  id: string
  entity_id: string
  username: string
  first_name: string
  last_name: string
  email: string
  avatar_url: string
}

export function useUsers({ enabled = true }: { enabled?: boolean } = {}) {
  return useQuery({
    queryKey: ["users"],
    queryFn: async () => {
      const res = await api.get<UserOption[]>("/users")
      return res.data
    },
    staleTime: 5 * 60 * 1000,
    enabled,
  })
}

export function userName(user: UserOption) {
  const fullName = `${user.first_name} ${user.last_name}`.trim()
  return fullName || user.username || user.entity_id
}

export function userInitials(user: UserOption) {
  return userName(user)
    .split(/\s+/)
    .map((part) => part[0])
    .filter(Boolean)
    .slice(0, 2)
    .join("")
    .toUpperCase()
}

export function userSearchKeys(user: UserOption) {
  return [
    user.first_name,
    user.last_name,
    user.username,
    user.email,
    user.entity_id,
  ].filter(Boolean)
}
