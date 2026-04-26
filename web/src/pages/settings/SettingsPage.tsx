import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export default function SettingsPage() {
  return (
    <PageContainer>
      <PageHeader
        title="Settings"
        description="Profile, authentication methods, linked accounts, and active sessions."
      />
      <Card>
        <CardHeader>
          <CardTitle>Coming soon</CardTitle>
          <CardDescription>Profile editor and security settings will live here.</CardDescription>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          Placeholder page during design phase.
        </CardContent>
      </Card>
    </PageContainer>
  )
}
