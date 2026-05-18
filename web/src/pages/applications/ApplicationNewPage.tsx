import { ArrowLeft } from "lucide-react"
import { useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { toast } from "sonner"

import { OutlineButton } from "@/components/OutlineButton"
import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { api } from "@/lib/api"
import type { Application } from "@/lib/applications"

export default function ApplicationNewPage() {
  const navigate = useNavigate()
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [iconURL, setIconURL] = useState("")
  const [launchURL, setLaunchURL] = useState("")
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (submitting) return
    setSubmitting(true)
    try {
      const res = await api.post<Application>("/applications", {
        name,
        description,
        icon_url: iconURL,
        launch_url: launchURL,
      })
      toast.success("Application created")
      navigate(`/applications/${res.data.id}`)
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't create the application."
      toast.error(message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <PageContainer>
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to="/applications">
          <ArrowLeft className="mr-1 size-3.5" />
          All applications
        </Link>
      </Button>

      <PageHeader
        title="New application"
        description="Register an OAuth client so the app can sign users in through Sentinel."
      />

      <form onSubmit={handleSubmit} className="max-w-xl space-y-5">
        <div className="space-y-2">
          <Label htmlFor="name">Name</Label>
          <Input
            id="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Blix"
            required
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="description">Description</Label>
          <Textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Telemetry visualization and dashboarding"
            rows={2}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="launch_url">Launch URL</Label>
          <Input
            id="launch_url"
            type="url"
            value={launchURL}
            onChange={(e) => setLaunchURL(e.target.value)}
            placeholder="https://blix.gauchoracing.com"
          />
          <p className="text-xs text-muted-foreground">
            Where the dashboard's launch button sends users.
          </p>
        </div>
        <div className="space-y-2">
          <Label htmlFor="icon_url">Icon URL</Label>
          <Input
            id="icon_url"
            type="url"
            value={iconURL}
            onChange={(e) => setIconURL(e.target.value)}
            placeholder="https://blix.gauchoracing.com/icon.png"
          />
        </div>

        <div className="flex gap-2 pt-2">
          <Button asChild variant="outline">
            <Link to="/applications">Cancel</Link>
          </Button>
          <OutlineButton type="submit" className="w-auto" loading={submitting} disabled={!name}>
            Create application
          </OutlineButton>
        </div>
      </form>
    </PageContainer>
  )
}
