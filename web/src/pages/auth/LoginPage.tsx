import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export default function LoginPage() {
  return (
    <main className="mx-auto flex min-h-svh max-w-md flex-col justify-center gap-6 p-8">
      <Card>
        <CardHeader>
          <CardTitle>Sign in to Sentinel</CardTitle>
          <CardDescription>Login methods will go here.</CardDescription>
        </CardHeader>
        <CardContent>
          <Button className="w-full" disabled>
            Sign in with Discord
          </Button>
        </CardContent>
      </Card>
    </main>
  )
}
