import { Activity, Boxes, LogIn, Users, UsersRound } from "lucide-react"
import { useState } from "react"

import { PageContainer, PageHeader } from "@/components/PageContainer"
import { cn } from "@/lib/utils"

import { ApplicationsTab } from "./tabs/ApplicationsTab"
import { GroupsTab } from "./tabs/GroupsTab"
import { MembersTab } from "./tabs/MembersTab"
import { OverviewTab } from "./tabs/OverviewTab"
import { SigninsTab } from "./tabs/SigninsTab"

const TABS = [
  { key: "overview", label: "Overview", icon: Activity, Component: OverviewTab },
  { key: "signins", label: "Sign-ins", icon: LogIn, Component: SigninsTab },
  { key: "applications", label: "Applications", icon: Boxes, Component: ApplicationsTab },
  { key: "members", label: "Members", icon: Users, Component: MembersTab },
  { key: "groups", label: "Groups", icon: UsersRound, Component: GroupsTab },
] as const

type TabKey = (typeof TABS)[number]["key"]

export default function AnalyticsPage() {
  const [tab, setTab] = useState<TabKey>("overview")

  const Active = TABS.find((t) => t.key === tab)?.Component ?? OverviewTab

  return (
    <PageContainer>
      <PageHeader
        title="Analytics"
        description="Sign-in trends, application usage, and membership across the team."
      />

      <div className="mb-6 flex gap-1 overflow-x-auto border-b border-border">
        {TABS.map((t) => (
          <button
            key={t.key}
            type="button"
            onClick={() => setTab(t.key)}
            className={cn(
              "flex items-center gap-1.5 whitespace-nowrap border-b-2 px-3 py-2 text-sm font-medium transition-colors",
              tab === t.key
                ? "border-gr-pink text-foreground"
                : "border-transparent text-muted-foreground hover:text-foreground",
            )}
          >
            <t.icon className="size-4" />
            {t.label}
          </button>
        ))}
      </div>
      <Active />
    </PageContainer>
  )
}
