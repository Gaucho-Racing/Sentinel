import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export default function AnalyticsPage() {
  return (
    <PageContainer>
      <PageHeader
        title="Analytics"
        description="Sign-in trends, top applications, and group activity across the team."
      />
      <Card>
        <CardHeader>
          <CardTitle>Coming soon</CardTitle>
          <CardDescription>Charts and aggregate metrics will live here.</CardDescription>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          Placeholder page during design phase.
        </CardContent>
      </Card>
    </PageContainer>
  )
}
