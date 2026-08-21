import { Activity, BarChart3, Boxes, LogIn, ScrollText, Users, UsersRound } from "lucide-react"
import { useState } from "react"

import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useAdmins } from "@/lib/admin"
import { cn } from "@/lib/utils"

import { ApplicationsTab } from "./tabs/ApplicationsTab"
import { AuditTab } from "./tabs/AuditTab"
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
  { key: "audit", label: "Audit", icon: ScrollText, Component: AuditTab },
] as const

type TabKey = (typeof TABS)[number]["key"]

export default function AnalyticsPage() {
  const [tab, setTab] = useState<TabKey>("overview")
  const { isAdmin, isLoading } = useAdmins()

  const Active = TABS.find((t) => t.key === tab)?.Component ?? OverviewTab

  return (
    <PageContainer>
      <PageHeader
        title="Analytics"
        description="Sign-in trends, application usage, membership, and audit activity across the team."
      />

      {isLoading ? (
        <div className="space-y-6">
          <Skeleton className="h-10 w-full max-w-lg" />
          <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <Skeleton key={i} className="h-24 rounded-lg" />
            ))}
          </div>
        </div>
      ) : !isAdmin ? (
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <BarChart3 className="size-4 text-muted-foreground" />
              <CardTitle>Restricted</CardTitle>
            </div>
            <CardDescription>Team analytics are available to admins.</CardDescription>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            Ask an existing admin to add you to the Admins group if you need access to sign-in,
            membership, and audit insights.
          </CardContent>
        </Card>
      ) : (
        <>
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
        </>
      )}
    </PageContainer>
  )
}
