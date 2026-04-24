import { useSearchParams } from "react-router-dom"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export default function AuthorizePage() {
  const [params] = useSearchParams()
  const clientId = params.get("client_id") ?? ""
  const redirectUri = params.get("redirect_uri") ?? ""
  const scope = params.get("scope") ?? ""

  return (
    <main className="mx-auto flex min-h-svh max-w-lg flex-col justify-center gap-6 p-8">
      <Card>
        <CardHeader>
          <CardTitle>Authorize application</CardTitle>
          <CardDescription>
            The consent screen will render the app name, icon, and requested scopes.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
            <dt className="text-muted-foreground">client_id</dt>
            <dd className="font-mono">{clientId || "-"}</dd>
            <dt className="text-muted-foreground">redirect_uri</dt>
            <dd className="truncate font-mono">{redirectUri || "-"}</dd>
            <dt className="text-muted-foreground">scope</dt>
            <dd className="font-mono">{scope || "-"}</dd>
          </dl>
          <div className="flex gap-2">
            <Button className="flex-1" disabled>
              Approve
            </Button>
            <Button className="flex-1" variant="outline" disabled>
              Deny
            </Button>
          </div>
        </CardContent>
      </Card>
    </main>
  )
}
