import { ArrowRight } from "lucide-react"
import { Link } from "react-router-dom"

import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

type LinkGroup = {
  title: string
  description: string
  links: Array<{ to: string; label: string; note?: string }>
}

const SAMPLE_AUTHORIZE = new URLSearchParams({
  client_id: "sentinel",
  redirect_uri: "http://localhost:3000/auth/callback",
  scope: "user:read groups:read",
}).toString()

const GROUPS: LinkGroup[] = [
  {
    title: "Dashboard",
    description: "Pages reachable from the sidebar.",
    links: [
      { to: "/", label: "Home" },
      { to: "/applications", label: "Applications" },
      { to: "/groups", label: "Groups" },
      { to: "/analytics", label: "Analytics" },
      { to: "/settings", label: "Settings" },
    ],
  },
  {
    title: "Auth flow",
    description: "Full-bleed pages outside the dashboard shell.",
    links: [
      { to: "/auth/login", label: "Login" },
      {
        to: `/oauth/authorize?${SAMPLE_AUTHORIZE}`,
        label: "OAuth authorize",
        note: "with sample client_id, redirect_uri, scope",
      },
    ],
  },
  {
    title: "Edge cases",
    description: "States that aren't part of the normal nav.",
    links: [{ to: "/this-route-does-not-exist", label: "404" }],
  },
]

function LinkRow({ to, label, note }: { to: string; label: string; note?: string }) {
  return (
    <Link
      to={to}
      className="flex items-center justify-between gap-4 px-6 py-3 transition-colors hover:bg-muted/40"
    >
      <div>
        <p className="text-sm font-medium leading-none">{label}</p>
        {note && <p className="mt-1 text-xs text-muted-foreground">{note}</p>}
      </div>
      <div className="flex items-center gap-3 text-xs text-muted-foreground">
        <span className="font-mono">{to}</span>
        <ArrowRight className="size-3.5" />
      </div>
    </Link>
  )
}

export default function DebugPage() {
  return (
    <PageContainer>
      <PageHeader
        title="Debug"
        description="Index of every page in the app, for design and QA."
      />
      <div className="space-y-6">
        {GROUPS.map((group) => (
          <Card key={group.title}>
            <CardHeader>
              <CardTitle>{group.title}</CardTitle>
              <CardDescription>{group.description}</CardDescription>
            </CardHeader>
            <CardContent className="p-0">
              <ul className="divide-y divide-border">
                {group.links.map((link) => (
                  <li key={link.to}>
                    <LinkRow {...link} />
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>
        ))}
      </div>
    </PageContainer>
  )
}
