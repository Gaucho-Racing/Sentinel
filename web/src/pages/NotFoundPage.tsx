import { Link } from "react-router-dom"

import { Button } from "@/components/ui/button"

export default function NotFoundPage() {
  return (
    <main className="mx-auto flex min-h-svh max-w-md flex-col items-center justify-center gap-4 p-8 text-center">
      <h1 className="text-2xl font-semibold">404</h1>
      <p className="text-muted-foreground">This page does not exist.</p>
      <Button asChild variant="outline">
        <Link to="/">Go home</Link>
      </Button>
    </main>
  )
}
