import { ChevronRight, Unlink } from "lucide-react"
import { Link } from "react-router-dom"

import { ApplicationIcon } from "@/components/ApplicationIcon"
import { GroupCard } from "@/components/GroupCard"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import type { Application, ApplicationGroupLink } from "@/lib/applications"
import type { Group } from "@/lib/groups"

type Section = {
  application: Application | null
  groups: { group: Group; gating: boolean }[]
}

// A group linked to three applications shows up under all three — the sections
// are a lens on the links, not a partition of the groups.
function buildSections(
  groups: Group[],
  applications: Application[],
  links: ApplicationGroupLink[],
): Section[] {
  const visible = new Map(groups.map((g) => [g.id, g]))
  const linkedGroupIDs = new Set<string>()

  const sections: Section[] = []
  for (const application of applications) {
    const members: Section["groups"] = []
    for (const link of links) {
      if (link.application_id !== application.id) continue
      linkedGroupIDs.add(link.group_id)
      const group = visible.get(link.group_id)
      if (group) members.push({ group, gating: link.required })
    }
    if (members.length > 0) sections.push({ application, groups: members })
  }

  sections.sort(
    (a, b) =>
      b.groups.length - a.groups.length ||
      (a.application?.name ?? "").localeCompare(b.application?.name ?? ""),
  )

  const unlinked = groups
    .filter((g) => !linkedGroupIDs.has(g.id))
    .map((group) => ({ group, gating: false }))
  if (unlinked.length > 0) sections.push({ application: null, groups: unlinked })

  return sections
}

function GatingBadge() {
  return (
    <Badge
      variant="outline"
      className="h-5 border-gr-pink/40 bg-gr-pink/10 text-[10px] text-gr-pink"
    >
      gates access
    </Badge>
  )
}

export function GroupsByApplication({
  groups,
  applications,
  links,
  isLoading,
  emptyLabel,
}: {
  groups: Group[]
  applications: Application[]
  links: ApplicationGroupLink[]
  isLoading: boolean
  emptyLabel: string
}) {
  if (isLoading) {
    return (
      <div className="space-y-8">
        {Array.from({ length: 2 }).map((_, i) => (
          <div key={i} className="space-y-3">
            <Skeleton className="h-16 rounded-lg" />
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {Array.from({ length: 3 }).map((_, j) => (
                <Skeleton key={j} className="h-32 rounded-lg" />
              ))}
            </div>
          </div>
        ))}
      </div>
    )
  }

  const sections = buildSections(groups, applications, links)
  if (sections.length === 0) {
    return <p className="py-12 text-center text-sm text-muted-foreground">{emptyLabel}</p>
  }

  return (
    <div className="space-y-8">
      {sections.map((section) => {
        const gatingCount = section.groups.filter((g) => g.gating).length
        return (
          <section key={section.application?.id ?? "unlinked"}>
            {section.application ? (
              <Link
                to={`/applications/${section.application.id}`}
                className="mb-3 flex items-center gap-3 rounded-lg border border-border/60 bg-card px-4 py-3 transition-colors hover:border-gr-purple/50 hover:bg-muted/40"
              >
                <ApplicationIcon
                  name={section.application.name}
                  iconUrl={section.application.icon_url}
                  className="size-9 rounded-md text-sm"
                />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium leading-none">
                    {section.application.name}
                  </p>
                  <p className="mt-1 truncate text-xs text-muted-foreground">
                    {section.application.description || "No description."}
                  </p>
                </div>
                <div className="flex shrink-0 items-center gap-2 text-xs text-muted-foreground">
                  {gatingCount > 0 && <GatingBadge />}
                  <span>
                    {section.groups.length} group{section.groups.length === 1 ? "" : "s"}
                  </span>
                  <ChevronRight className="size-4" />
                </div>
              </Link>
            ) : (
              <div className="mb-3 flex items-center gap-2 px-1 py-3 text-muted-foreground">
                <Unlink className="size-4" />
                <p className="text-sm font-medium">Not linked to an application</p>
                <span className="text-xs">
                  · {section.groups.length} group{section.groups.length === 1 ? "" : "s"}
                </span>
              </div>
            )}

            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {section.groups.map(({ group, gating }) => (
                <GroupCard
                  key={group.id}
                  group={group}
                  accessory={gating ? <GatingBadge /> : undefined}
                />
              ))}
            </div>
          </section>
        )
      })}
    </div>
  )
}
