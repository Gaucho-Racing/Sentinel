import { Plus, Search } from "lucide-react"
import { useState } from "react"

import { GroupCard, type GroupSummary } from "@/components/GroupCard"
import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

const MOCK_GROUPS: GroupSummary[] = [
  {
    id: "grp_officers",
    name: "Officers",
    description: "Team leadership — President, VPs, Captains, and Safety.",
    member_count: 8,
    owner_count: 3,
    allowed_sources: ["DIRECT"],
    pending_requests: 0,
  },
  {
    id: "grp_suspension",
    name: "Suspension Team",
    description: "Geometry, dampers, kinematics. Synced with the Suspension Discord role.",
    member_count: 12,
    owner_count: 2,
    allowed_sources: ["DIRECT", "DISCORD"],
    pending_requests: 1,
  },
  {
    id: "grp_aero",
    name: "Aero Team",
    description: "CFD, bodywork, undertray. Synced with the Aero Discord role.",
    member_count: 15,
    owner_count: 2,
    allowed_sources: ["DIRECT", "DISCORD"],
    pending_requests: 3,
  },
  {
    id: "grp_powertrain",
    name: "Powertrain Team",
    description: "Engine, drivetrain, fuel system. Synced with the Powertrain Discord role.",
    member_count: 10,
    owner_count: 2,
    allowed_sources: ["DIRECT", "DISCORD"],
    pending_requests: 0,
  },
  {
    id: "grp_electrical",
    name: "Electrical Team",
    description: "Wiring harness, sensors, data acquisition.",
    member_count: 9,
    owner_count: 2,
    allowed_sources: ["DIRECT", "DISCORD"],
    pending_requests: 2,
  },
  {
    id: "grp_all_members",
    name: "All Members",
    description: "Auto-populated from active membership — every onboarded member.",
    member_count: 89,
    owner_count: 1,
    allowed_sources: ["CONDITIONAL"],
    pending_requests: 0,
  },
  {
    id: "grp_alumni",
    name: "Alumni",
    description: "Past members. Auto-populated based on graduation year.",
    member_count: 47,
    owner_count: 1,
    allowed_sources: ["CONDITIONAL"],
    pending_requests: 0,
  },
  {
    id: "grp_mentors",
    name: "Mentors",
    description: "Industry mentors and alumni advisors with project oversight.",
    member_count: 6,
    owner_count: 1,
    allowed_sources: ["DIRECT"],
    pending_requests: 2,
  },
  {
    id: "grp_faculty",
    name: "Faculty Advisors",
    description: "UCSB faculty with formal advisory roles.",
    member_count: 3,
    owner_count: 1,
    allowed_sources: ["DIRECT"],
    pending_requests: 0,
  },
]

export default function GroupsPage() {
  const [query, setQuery] = useState("")

  const needle = query.trim().toLowerCase()
  const filtered = needle
    ? MOCK_GROUPS.filter(
        (g) =>
          g.name.toLowerCase().includes(needle) ||
          g.description.toLowerCase().includes(needle),
      )
    : MOCK_GROUPS
  const sorted = [...filtered].sort((a, b) => a.name.localeCompare(b.name))

  return (
    <PageContainer>
      <div className="mb-6 flex items-start justify-between gap-4">
        <PageHeader
          title="Groups"
          description="Group memberships drive who can access which apps."
        />
        <Button>
          <Plus className="mr-1 size-3.5" />
          New group
        </Button>
      </div>

      <div className="relative mb-6">
        <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="search"
          placeholder="Search groups…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="pl-9"
        />
      </div>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {sorted.length === 0 ? (
          <p className="col-span-full py-12 text-center text-sm text-muted-foreground">
            No groups match "{query}".
          </p>
        ) : (
          sorted.map((group) => <GroupCard key={group.id} group={group} />)
        )}
      </div>
    </PageContainer>
  )
}
