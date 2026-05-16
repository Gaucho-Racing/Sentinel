import { ArrowLeft, Copy } from "lucide-react"
import { useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { toast } from "sonner"

import { OutlineButton } from "@/components/OutlineButton"
import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { api } from "@/lib/api"
import type { Application } from "@/lib/applications"

type CreatedApplication = Application & { client_secret: string }

export default function ApplicationNewPage() {
  const navigate = useNavigate()
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [iconURL, setIconURL] = useState("")
  const [launchURL, setLaunchURL] = useState("")
  const [submitting, setSubmitting] = useState(false)
  const [created, setCreated] = useState<CreatedApplication | null>(null)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (submitting) return
    setSubmitting(true)
    try {
      const res = await api.post<CreatedApplication>("/applications", {
        name,
        description,
        icon_url: iconURL,
        launch_url: launchURL,
      })
      setCreated(res.data)
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't create the application."
      toast.error(message)
    } finally {
      setSubmitting(false)
    }
  }

  function copy(value: string, label: string) {
    navigator.clipboard.writeText(value)
    toast.success(`${label} copied`)
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

      <Dialog open={!!created} onOpenChange={(open) => !open && navigate(`/applications/${created!.id}`)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Application created</DialogTitle>
            <DialogDescription>
              Save the client secret somewhere safe — it won't be shown again.
            </DialogDescription>
          </DialogHeader>
          {created && (
            <div className="space-y-3">
              <div>
                <p className="text-xs uppercase tracking-wider text-muted-foreground">Client ID</p>
                <div className="mt-1 flex items-center gap-2">
                  <code className="flex-1 break-all rounded bg-muted px-2 py-1 font-mono text-xs">
                    {created.client_id}
                  </code>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    onClick={() => copy(created.client_id, "Client ID")}
                  >
                    <Copy className="size-3.5" />
                  </Button>
                </div>
              </div>
              <div>
                <p className="text-xs uppercase tracking-wider text-muted-foreground">
                  Client secret
                </p>
                <div className="mt-1 flex items-center gap-2">
                  <code className="flex-1 break-all rounded bg-muted px-2 py-1 font-mono text-xs">
                    {created.client_secret}
                  </code>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    onClick={() => copy(created.client_secret, "Client secret")}
                  >
                    <Copy className="size-3.5" />
                  </Button>
                </div>
              </div>
              <OutlineButton
                type="button"
                onClick={() => navigate(`/applications/${created.id}`)}
              >
                Done
              </OutlineButton>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </PageContainer>
  )
}
