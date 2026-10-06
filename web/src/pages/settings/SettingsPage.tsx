import { ShieldCheck } from "lucide-react"
import { Link } from "react-router-dom"

import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { useAdmins } from "@/lib/admin"

export default function SettingsPage() {
  const { isAdmin } = useAdmins()

  return (
    <PageContainer>
      <PageHeader
        title="Settings"
        description="Manage your account and profile."
      />
      <div className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>Your profile</CardTitle>
            <CardDescription>Update your name, team details, and occupation.</CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild variant="outline">
              <Link to="/profile">Edit profile</Link>
            </Button>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Connected accounts</CardTitle>
            <CardDescription>View your Discord connection and manage your GitHub account.</CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild variant="outline">
              <Link to="/settings/connected-accounts">Manage connected accounts</Link>
            </Button>
          </CardContent>
        </Card>
      </div>

      {isAdmin && (
        <section className="mt-10 border-t border-border/60 pt-8">
          <div className="mb-4 flex items-center gap-2">
            <ShieldCheck className="size-4 text-muted-foreground" />
            <h2 className="text-sm font-medium">Admin</h2>
          </div>
          <div className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Audit log</CardTitle>
                <CardDescription>
                  Administrative actions across the team, and who performed them.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Button asChild variant="outline">
                  <Link to="/settings/audit">View audit log</Link>
                </Button>
              </CardContent>
            </Card>
          </div>
        </section>
      )}
    </PageContainer>
  )
}
