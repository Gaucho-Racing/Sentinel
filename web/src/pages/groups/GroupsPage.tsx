import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export default function GroupsPage() {
  return (
    <PageContainer>
      <PageHeader
        title="Groups"
        description="Group memberships drive who can access which apps. Manage your memberships and join requests here."
      />
      <Card>
        <CardHeader>
          <CardTitle>Coming soon</CardTitle>
          <CardDescription>Group browser, members, and join-request inbox will live here.</CardDescription>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          Placeholder page during design phase.
        </CardContent>
      </Card>
    </PageContainer>
  )
}
