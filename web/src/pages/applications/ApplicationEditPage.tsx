import { useQuery, useQueryClient } from "@tanstack/react-query"
import { ArrowLeft, Plus, ShieldAlert, X } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { Link, useNavigate, useParams } from "react-router-dom"
import { toast } from "sonner"

import { OutlineButton } from "@/components/OutlineButton"
import { PageContainer } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Skeleton } from "@/components/ui/skeleton"
import { Textarea } from "@/components/ui/textarea"
import { api } from "@/lib/api"
import { redirectURIWildcardExamples, type Application } from "@/lib/applications"

function BasicInfoCard({ id, app }: { id: string; app: Application }) {
  const qc = useQueryClient()
  const navigate = useNavigate()
  const [name, setName] = useState(app.name)
  const [description, setDescription] = useState(app.description)
  const [iconURL, setIconURL] = useState(app.icon_url)
  const [launchURL, setLaunchURL] = useState(app.launch_url)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (submitting) return
    setSubmitting(true)
    try {
      await api.put(`/applications/${id}`, {
        name,
        description,
        icon_url: iconURL,
        launch_url: launchURL,
      })
      qc.invalidateQueries({ queryKey: ["application", "id", id] })
      qc.invalidateQueries({ queryKey: ["applications"] })
      toast.success("Application updated")
      navigate(`/applications/${id}`)
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't save the application."
      toast.error(message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Basic info</CardTitle>
        <CardDescription>Name, description, branding, and where the launch button sends users.</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-5">
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input id="name" value={name} onChange={(e) => setName(e.target.value)} required />
          </div>
          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
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
              placeholder="https://app.gauchoracing.com"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="icon_url">Icon URL</Label>
            <Input
              id="icon_url"
              type="url"
              value={iconURL}
              onChange={(e) => setIconURL(e.target.value)}
            />
          </div>
          <div className="flex justify-end pt-2">
            <OutlineButton type="submit" className="w-auto" loading={submitting} disabled={!name}>
              Save changes
            </OutlineButton>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}

function RedirectURIsCard({ id, app }: { id: string; app: Application }) {
  const qc = useQueryClient()
  const [draft, setDraft] = useState("")
  const [adding, setAdding] = useState(false)
  const [removing, setRemoving] = useState<string | null>(null)
  const [wildcardConfirm, setWildcardConfirm] = useState<string | null>(null)

  const wildcardExamples = useMemo(
    () => (wildcardConfirm ? redirectURIWildcardExamples(wildcardConfirm) : []),
    [wildcardConfirm],
  )

  async function addURI(uri: string) {
    setAdding(true)
    try {
      await api.post(`/applications/${id}/redirect-uris`, { redirect_uri: uri })
      qc.invalidateQueries({ queryKey: ["application", "id", id] })
      setDraft("")
      setWildcardConfirm(null)
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't add redirect URI."
      toast.error(message)
    } finally {
      setAdding(false)
    }
  }

  function handleSubmit() {
    const uri = draft.trim()
    if (!uri || adding) return
    if (uri.includes("*")) {
      setWildcardConfirm(uri)
      return
    }
    addURI(uri)
  }

  async function remove(uri: string) {
    if (removing) return
    setRemoving(uri)
    try {
      await api.delete(`/applications/${id}/redirect-uris`, { params: { uri } })
      qc.invalidateQueries({ queryKey: ["application", "id", id] })
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't remove redirect URI."
      toast.error(message)
    } finally {
      setRemoving(null)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Redirect URIs</CardTitle>
        <CardDescription>
          Where OAuth callbacks are allowed to land. Add the exact URIs your app will redirect
          users to after the consent screen.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {app.redirect_uris.length === 0 ? (
          <p className="text-sm text-muted-foreground">No redirect URIs registered yet.</p>
        ) : (
          <ul className="space-y-2">
            {app.redirect_uris.map((uri) => (
              <li
                key={uri}
                className="flex items-center gap-1 rounded-md border border-border/60 bg-muted/40 px-2.5 py-1.5"
              >
                <code className="flex-1 truncate font-mono text-xs">{uri}</code>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  disabled={removing === uri}
                  onClick={() => remove(uri)}
                >
                  <X className="size-3.5" />
                </Button>
              </li>
            ))}
          </ul>
        )}
        <form
          onSubmit={(e) => {
            e.preventDefault()
            handleSubmit()
          }}
          className="flex gap-2 pt-2"
        >
          <Input
            placeholder="https://app.gauchoracing.com/auth/callback"
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
          />
          <Button type="submit" disabled={!draft.trim() || adding}>
            <Plus className="mr-1 size-3.5" />
            Add
          </Button>
        </form>
      </CardContent>

      <Dialog
        open={wildcardConfirm !== null}
        onOpenChange={(open) => {
          if (!open && !adding) setWildcardConfirm(null)
        }}
      >
        <DialogContent className="gap-5 sm:max-w-md">
          <DialogHeader className="gap-3">
            <div className="flex size-10 items-center justify-center rounded-xl bg-amber-500/15 text-amber-500">
              <ShieldAlert className="size-5" />
            </div>
            <DialogTitle>This URI contains a wildcard</DialogTitle>
            <DialogDescription>
              <code className="font-mono text-foreground">*</code> matches any sequence of
              characters — including dots and slashes — anywhere in the URI. Wildcards widen
              the attack surface (open redirect, subdomain takeover); prefer an exact URI
              when possible.
            </DialogDescription>
          </DialogHeader>

          {wildcardConfirm && (
            <div className="space-y-2">
              <p className="text-xs font-medium text-muted-foreground">Pattern</p>
              <code className="block break-all rounded-md border border-border/60 bg-muted/40 px-2.5 py-1.5 font-mono text-xs">
                {wildcardConfirm}
              </code>
            </div>
          )}

          {wildcardExamples.length > 0 && (
            <div className="space-y-2">
              <p className="text-xs font-medium text-muted-foreground">Examples that would be accepted</p>
              <ul className="space-y-1.5">
                {wildcardExamples.map((example) => (
                  <li
                    key={example}
                    className="break-all rounded-md border border-border/60 bg-muted/40 px-2.5 py-1.5 font-mono text-xs"
                  >
                    {example}
                  </li>
                ))}
              </ul>
            </div>
          )}

          <div className="flex justify-end gap-2 pt-1">
            <Button
              type="button"
              variant="ghost"
              disabled={adding}
              onClick={() => setWildcardConfirm(null)}
            >
              Cancel
            </Button>
            <Button
              type="button"
              disabled={adding}
              onClick={() => wildcardConfirm && addURI(wildcardConfirm)}
            >
              Add anyway
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </Card>
  )
}

export default function ApplicationEditPage() {
  const { id } = useParams<{ id: string }>()

  const query = useQuery({
    queryKey: ["application", "id", id],
    queryFn: async () => {
      const res = await api.get<Application>(`/applications/${id}`)
      return res.data
    },
    enabled: !!id,
  })

  // We hold a stable copy of the loaded app so child forms can prefill from
  // it without re-reading from the query on every render.
  const [appData, setAppData] = useState<Application | null>(null)
  useEffect(() => {
    if (query.data) setAppData(query.data)
  }, [query.data])

  if (query.isLoading || !appData) {
    return (
      <PageContainer>
        <Skeleton className="mb-4 h-4 w-24" />
        <Skeleton className="mb-8 h-8 w-48" />
        <div className="space-y-4">
          <Skeleton className="h-64" />
          <Skeleton className="h-48" />
        </div>
      </PageContainer>
    )
  }

  if (query.isError) {
    return (
      <PageContainer>
        <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
          <Link to="/applications">
            <ArrowLeft className="mr-1 size-3.5" />
            All applications
          </Link>
        </Button>
        <p className="text-sm text-muted-foreground">Application not found.</p>
      </PageContainer>
    )
  }

  return (
    <PageContainer>
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to={`/applications/${id}`}>
          <ArrowLeft className="mr-1 size-3.5" />
          Back to {appData.name}
        </Link>
      </Button>

      <div className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight">Edit {appData.name}</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Application metadata and OAuth redirect URIs.
        </p>
      </div>

      <div className="space-y-4">
        <BasicInfoCard id={id!} app={appData} />
        <RedirectURIsCard id={id!} app={query.data ?? appData} />
      </div>
    </PageContainer>
  )
}
