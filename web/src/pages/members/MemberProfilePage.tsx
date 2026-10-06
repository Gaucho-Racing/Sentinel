import { useQuery } from "@tanstack/react-query"
import { ArrowLeft, ArrowUpRight } from "lucide-react"
import type { ReactNode } from "react"
import { Link, useParams } from "react-router-dom"

import { ApplicationIcon } from "@/components/ApplicationIcon"
import { DiscordIcon, GithubIcon } from "@/components/icons/socials"
import { MemberAvatar } from "@/components/MemberCard"
import { PageContainer } from "@/components/PageContainer"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import type { Application } from "@/lib/applications"
import type { Entity } from "@/lib/auth"
import { SOURCE_LABEL, useAllGroupMembers, useGroups, type GroupSource } from "@/lib/groups"
import { userName, type Member } from "@/lib/users"

type EntityLogin = {
  id: string
  entity_id: string
  client_id: string
  ip_address: string
  created_at: string
}

type AccessedApplication = Application & { last_accessed_at: string }

const RECENT_LIMIT = 5

function formatDate(iso: string) {
  if (!iso || iso.startsWith("0001-01-01")) return null
  return new Date(iso).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  })
}

function formatDateTime(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  })
}

type Field = { label: string; value: ReactNode }

// Fields are passed as data rather than as children so the emptiness check is
// on the values. Filtering rendered <Field> elements would never drop anything
// — an element that renders null is still a truthy object.
function DetailCard({ title, fields }: { title: string; fields: Field[] }) {
  const present = fields.filter((field) => !!field.value)
  if (present.length === 0) return null
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-2 gap-4 sm:grid-cols-3">
        {present.map((field) => (
          <div key={field.label}>
            <p className="text-[11px] uppercase tracking-wide text-muted-foreground">
              {field.label}
            </p>
            <p className="mt-0.5 break-words text-sm">{field.value}</p>
          </div>
        ))}
      </CardContent>
    </Card>
  )
}

export default function MemberProfilePage() {
  const { entityID = "" } = useParams()

  const entityQuery = useQuery({
    queryKey: ["entity", entityID],
    queryFn: async () => (await api.get<Entity>(`/entities/${entityID}`)).data,
    enabled: !!entityID,
  })

  const profile = entityQuery.data?.user
  const userID = profile?.id

  const groupsQuery = useGroups()
  const membershipsQuery = useAllGroupMembers()

  const loginsQuery = useQuery({
    queryKey: ["logins", userID],
    queryFn: async () =>
      (
        await api.get<EntityLogin[]>(`/users/${userID}/logins`, {
          params: { limit: RECENT_LIMIT },
        })
      ).data,
    enabled: !!userID,
  })

  const appsQuery = useQuery({
    queryKey: ["recentApplications", userID],
    queryFn: async () =>
      (
        await api.get<AccessedApplication[]>(`/users/${userID}/recent-applications`, {
          params: { limit: RECENT_LIMIT },
        })
      ).data,
    enabled: !!userID,
  })

  const memberships = (membershipsQuery.data ?? []).filter((m) => m.entity_id === entityID)
  const groupByID = new Map((groupsQuery.data ?? []).map((g) => [g.id, g]))
  const held = memberships
    .map((m) => ({ group: groupByID.get(m.group_id), source: m.source }))
    .filter((entry) => entry.group)
    .sort((a, b) => a.group!.name.localeCompare(b.group!.name))

  const discord = entityQuery.data?.external_auths?.find((a) => a.provider === "DISCORD")
  const github = entityQuery.data?.external_auths?.find((a) => a.provider === "GITHUB")

  if (entityQuery.isLoading) {
    return (
      <PageContainer>
        <Skeleton className="mb-6 h-24 rounded-lg" />
        <div className="space-y-4">
          <Skeleton className="h-40 rounded-lg" />
          <Skeleton className="h-40 rounded-lg" />
        </div>
      </PageContainer>
    )
  }

  if (entityQuery.isError || !profile) {
    return (
      <PageContainer>
        <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
          <Link to="/members">
            <ArrowLeft className="size-4" />
            Members
          </Link>
        </Button>
        <Card>
          <CardHeader>
            <CardTitle>Member not found</CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            No member profile exists for <code className="font-mono">{entityID}</code>. The
            entity may be a service account, or it may have been deleted.
          </CardContent>
        </Card>
      </PageContainer>
    )
  }

  const member: Member = profile
  const birthday = formatDate(member.birthday)

  return (
    <PageContainer>
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to="/members">
          <ArrowLeft className="size-4" />
          Members
        </Link>
      </Button>

      <header className="mb-8 flex flex-wrap items-start gap-4">
        <MemberAvatar member={member} className="size-16" fallbackClassName="text-xl" />
        <div className="min-w-0 flex-1">
          <h1 className="text-2xl font-semibold tracking-tight">{userName(member)}</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {member.username ? `@${member.username}` : member.entity_id}
          </p>
          <div className="mt-2 flex flex-wrap items-center gap-1.5">
            {member.initial_role && (
              <Badge variant="secondary" className="h-5 text-[10px]">
                {member.initial_role}
              </Badge>
            )}
            {discord && (
              <Badge variant="outline" className="h-5 gap-1 text-[10px]">
                <DiscordIcon className="size-3" />
                {discord.metadata?.username ?? "linked"}
              </Badge>
            )}
            {github && (
              <Badge variant="outline" className="h-5 gap-1 text-[10px]">
                <GithubIcon className="size-3" />
                {github.metadata?.username ?? "linked"}
              </Badge>
            )}
          </div>
        </div>
      </header>

      <div className="space-y-4">
        <DetailCard
          title="Contact"
          fields={[
            { label: "Email", value: member.email },
            { label: "Phone", value: member.phone_number },
            {
              label: "Entity ID",
              value: <code className="font-mono text-xs">{entityID}</code>,
            },
          ]}
        />

        <DetailCard
          title="Academic"
          fields={[
            { label: "Level", value: member.graduate_level },
            { label: "Major", value: member.major },
            {
              label: "Graduation",
              value: member.graduation_year > 0 ? String(member.graduation_year) : "",
            },
          ]}
        />

        <DetailCard
          title="Personal"
          fields={[
            { label: "Gender", value: member.gender },
            { label: "Birthday", value: birthday },
            { label: "Joined", value: formatDate(member.created_at) },
          ]}
        />

        <DetailCard
          title="Team"
          fields={[
            { label: "Shirt size", value: member.shirt_size },
            { label: "Jacket size", value: member.jacket_size },
            { label: "SAE number", value: member.sae_registration_number },
          ]}
        />

        <DetailCard
          title="Occupation"
          fields={[
            { label: "Title", value: member.occupation_title },
            { label: "Company", value: member.occupation_company },
          ]}
        />

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Groups</CardTitle>
          </CardHeader>
          <CardContent>
            {membershipsQuery.isLoading ? (
              <Skeleton className="h-6 w-48" />
            ) : held.length === 0 ? (
              <p className="text-sm text-muted-foreground">Not a member of any group.</p>
            ) : (
              <div className="flex flex-wrap gap-1.5">
                {held.map(({ group, source }) => (
                  <Link
                    key={group!.id}
                    to={`/groups/${group!.id}`}
                    className="inline-flex items-center gap-1.5 rounded-md border border-border/60 bg-muted/40 px-2 py-1 text-xs transition-colors hover:border-gr-purple/60 hover:bg-muted"
                  >
                    {group!.name}
                    {source && (
                      <span className="font-mono text-[10px] text-muted-foreground">
                        {SOURCE_LABEL[source as GroupSource] ?? source.toLowerCase()}
                      </span>
                    )}
                  </Link>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Applications used</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {appsQuery.isLoading ? (
              <div className="space-y-2 p-6">
                <Skeleton className="h-8 w-full" />
                <Skeleton className="h-8 w-2/3" />
              </div>
            ) : (appsQuery.data ?? []).length === 0 ? (
              <p className="px-6 pb-6 text-sm text-muted-foreground">
                Hasn't signed into any application yet.
              </p>
            ) : (
              <ul className="divide-y divide-border">
                {(appsQuery.data ?? []).map((app) => (
                  <li key={app.id}>
                    <Link
                      to={`/applications/${app.id}`}
                      className="flex items-center gap-3 px-6 py-3 transition-colors hover:bg-muted/40"
                    >
                      <ApplicationIcon
                        name={app.name}
                        iconUrl={app.icon_url}
                        className="size-7 rounded-md text-xs"
                      />
                      <span className="min-w-0 flex-1 truncate text-sm">{app.name}</span>
                      <span className="shrink-0 text-xs text-muted-foreground">
                        {formatDateTime(app.last_accessed_at)}
                      </span>
                      <ArrowUpRight className="size-3.5 shrink-0 text-muted-foreground" />
                    </Link>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Recent sign-ins</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {loginsQuery.isLoading ? (
              <div className="space-y-2 p-6">
                <Skeleton className="h-8 w-full" />
                <Skeleton className="h-8 w-2/3" />
              </div>
            ) : (loginsQuery.data ?? []).length === 0 ? (
              <p className="px-6 pb-6 text-sm text-muted-foreground">No sign-in activity yet.</p>
            ) : (
              <ul className="divide-y divide-border">
                {(loginsQuery.data ?? []).map((login) => (
                  <li
                    key={login.id}
                    className="flex flex-wrap items-center gap-x-3 gap-y-1 px-6 py-3 text-sm"
                  >
                    <span className="font-mono text-xs">{login.client_id}</span>
                    <span className="ml-auto whitespace-nowrap text-xs text-muted-foreground">
                      {formatDateTime(login.created_at)}
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      </div>
    </PageContainer>
  )
}
