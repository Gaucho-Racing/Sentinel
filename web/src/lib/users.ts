import { useQuery } from "@tanstack/react-query"

import { api } from "@/lib/api"

// The subset the people-pickers and chips need. Kept narrow on purpose so the
// helpers below stay usable with partial records; `Member` satisfies it.
export type UserOption = {
  id: string
  entity_id: string
  username: string
  first_name: string
  last_name: string
  email: string
  avatar_url: string
}

export type LinkedAccount = {
  provider: "GITHUB" | "DISCORD"
  external_id: string
  username: string
}

// Full mirror of core/model/user.go. `GET /users` has always returned every
// field — including email and phone, which PopulateUser fills from the entity's
// auth rows — the directory is just the first surface to show them.
export type Member = UserOption & {
  phone_number: string
  gender: string
  birthday: string
  graduate_level: string
  graduation_year: number
  major: string
  shirt_size: string
  jacket_size: string
  sae_registration_number: string
  occupation_title: string
  occupation_company: string
  // Group *names*, not IDs. Filter by group through the membership table
  // instead (useAllGroupMembers) so a rename can't break the filter.
  groups: string[] | null
  linked_accounts?: LinkedAccount[]
  initial_role: string
  updated_at: string
  created_at: string
}

export function useUsers({ enabled = true }: { enabled?: boolean } = {}) {
  return useQuery({
    queryKey: ["users"],
    queryFn: async () => {
      const res = await api.get<Member[]>("/users")
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

export function memberDisplayRole(member: Member): string {
  return member.initial_role || "member"
}

// "Undergraduate · Mechanical Engineering · '27" with the empty parts dropped,
// so a sparsely-filled profile doesn't render a row of stray separators.
export function academicSummary(member: Member): string {
  return [
    member.graduate_level,
    member.major,
    member.graduation_year > 0 ? `'${String(member.graduation_year).slice(2)}` : "",
  ]
    .filter(Boolean)
    .join(" · ")
}
