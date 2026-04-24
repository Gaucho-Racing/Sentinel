import { Link } from "react-router-dom"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export default function HomePage() {
  return (
    <main className="mx-auto flex min-h-svh max-w-2xl flex-col justify-center gap-6 p-8">
      <Card>
        <CardHeader>
          <CardTitle>Sentinel</CardTitle>
          <CardDescription>Gaucho Racing's authentication service.</CardDescription>
        </CardHeader>
        <CardContent className="flex gap-2">
          <Button asChild>
            <Link to="/auth/login">Sign in</Link>
          </Button>
          <Button asChild variant="outline">
            <Link to="/applications">Applications</Link>
          </Button>
        </CardContent>
      </Card>
    </main>
  )
}
