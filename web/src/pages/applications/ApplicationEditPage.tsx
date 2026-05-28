import { useQuery, useQueryClient } from "@tanstack/react-query"
import { ArrowLeft, Plus, ShieldAlert, Trash2, X } from "lucide-react"
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

function BasicInfoCard({
  name,
  description,
  iconURL,
  launchURL,
  onChangeName,
  onChangeDescription,
  onChangeIconURL,
  onChangeLaunchURL,
}: {
  name: string
  description: string
  iconURL: string
  launchURL: string
  onChangeName: (v: string) => void
  onChangeDescription: (v: string) => void
  onChangeIconURL: (v: string) => void
  onChangeLaunchURL: (v: string) => void
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Basic info</CardTitle>
        <CardDescription>Name, description, branding, and where the launch button sends users.</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-5">
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input id="name" value={name} onChange={(e) => onChangeName(e.target.value)} required />
          </div>
          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={description}
              onChange={(e) => onChangeDescription(e.target.value)}
              rows={2}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="launch_url">Launch URL</Label>
            <Input
              id="launch_url"
              type="url"
              value={launchURL}
              onChange={(e) => onChangeLaunchURL(e.target.value)}
              placeholder="https://app.gauchoracing.com"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="icon_url">Icon URL</Label>
            <Input
              id="icon_url"
              type="url"
              value={iconURL}
              onChange={(e) => onChangeIconURL(e.target.value)}
            />
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function RedirectURIsCard({
  uris,
  onAddURI,
  onRemoveURI,
}: {
  uris: string[]
  onAddURI: (uri: string) => void
  onRemoveURI: (uri: string) => void
}) {
  const [draft, setDraft] = useState("")
  const [wildcardConfirm, setWildcardConfirm] = useState<string | null>(null)

  const wildcardExamples = useMemo(
    () => (wildcardConfirm ? redirectURIWildcardExamples(wildcardConfirm) : []),
    [wildcardConfirm],
  )

  function stageURI(uri: string) {
    if (uris.includes(uri)) {
      toast.error("That URI is already in the list.")
      return
    }
    onAddURI(uri)
    setDraft("")
    setWildcardConfirm(null)
  }

  function handleSubmit() {
    const uri = draft.trim()
    if (!uri) return
    if (uri.includes("*")) {
      setWildcardConfirm(uri)
      return
    }
    stageURI(uri)
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
        {uris.length === 0 ? (
          <p className="text-sm text-muted-foreground">No redirect URIs registered yet.</p>
        ) : (
          <ul className="space-y-2">
            {uris.map((uri) => (
              <li
                key={uri}
                className="flex items-center gap-1 rounded-md border border-border/60 bg-muted/40 px-2.5 py-1.5"
              >
                <code className="flex-1 truncate font-mono text-xs">{uri}</code>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  onClick={() => onRemoveURI(uri)}
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
          <Button type="submit" disabled={!draft.trim()}>
            <Plus className="mr-1 size-3.5" />
            Add
          </Button>
        </form>
      </CardContent>

      <Dialog
        open={wildcardConfirm !== null}
        onOpenChange={(open) => {
          if (!open) setWildcardConfirm(null)
        }}
      >
        <DialogContent className="gap-5 sm:max-w-md">
          <DialogHeader className="gap-3">
            <div className="flex size-10 items-center justify-center rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-white">
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
              onClick={() => setWildcardConfirm(null)}
            >
              Cancel
            </Button>
            <Button
              type="button"
              onClick={() => wildcardConfirm && stageURI(wildcardConfirm)}
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
  const navigate = useNavigate()
  const qc = useQueryClient()

  const query = useQuery({
    queryKey: ["application", "id", id],
    queryFn: async () => {
      const res = await api.get<Application>(`/applications/${id}`)
      return res.data
    },
    enabled: !!id,
  })

  // Basics form state.
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [iconURL, setIconURL] = useState("")
  const [launchURL, setLaunchURL] = useState("")
  const [initialized, setInitialized] = useState(false)

  // Staged redirect URI changes — applied on Save.
  const [pendingURIAdds, setPendingURIAdds] = useState<string[]>([])
  const [pendingURIRemoves, setPendingURIRemoves] = useState<Set<string>>(new Set())

  // Dialog / in-flight state.
  const [submitting, setSubmitting] = useState(false)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    if (query.data && !initialized) {
      setName(query.data.name)
      setDescription(query.data.description)
      setIconURL(query.data.icon_url)
      setLaunchURL(query.data.launch_url)
      setInitialized(true)
    }
  }, [query.data, initialized])

  const serverURIs = query.data?.redirect_uris ?? []
  const effectiveURIs = [
    ...serverURIs.filter((u) => !pendingURIRemoves.has(u)),
    ...pendingURIAdds,
  ]

  function handleAddURI(uri: string) {
    setPendingURIAdds((prev) => (prev.includes(uri) ? prev : [...prev, uri]))
  }

  function handleRemoveURI(uri: string) {
    if (pendingURIAdds.includes(uri)) {
      setPendingURIAdds((prev) => prev.filter((u) => u !== uri))
      return
    }
    setPendingURIRemoves((prev) => {
      const next = new Set(prev)
      next.add(uri)
      return next
    })
  }

  async function commitSave() {
    if (!id) return
    setSubmitting(true)
    try {
      for (const uri of pendingURIRemoves) {
        await api.delete(`/applications/${id}/redirect-uris`, { params: { uri } })
      }
      for (const uri of pendingURIAdds) {
        await api.post(`/applications/${id}/redirect-uris`, { redirect_uri: uri })
      }
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

  async function handleDelete() {
    if (!id || deleting) return
    setDeleting(true)
    try {
      await api.delete(`/applications/${id}`)
      qc.invalidateQueries({ queryKey: ["applications"] })
      toast.success("Application deleted")
      navigate("/applications")
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't delete the application."
      toast.error(message)
      setDeleting(false)
      setDeleteConfirmOpen(false)
    }
  }

  if (query.isLoading || !initialized) {
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

  if (query.isError || !query.data) {
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

  const app = query.data

  return (
    <PageContainer>
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to={`/applications/${id}`}>
          <ArrowLeft className="mr-1 size-3.5" />
          Back to {app.name}
        </Link>
      </Button>

      <header className="mb-8 flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0">
          <h1 className="text-2xl font-semibold tracking-tight">Edit {app.name}</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Update metadata and redirect URIs. Nothing is saved until you click Save changes.
          </p>
        </div>
        <OutlineButton
          type="button"
          className="w-auto"
          loading={submitting}
          disabled={!name.trim()}
          onClick={commitSave}
        >
          Save changes
        </OutlineButton>
      </header>

      <div className="space-y-4">
        <BasicInfoCard
          name={name}
          description={description}
          iconURL={iconURL}
          launchURL={launchURL}
          onChangeName={setName}
          onChangeDescription={setDescription}
          onChangeIconURL={setIconURL}
          onChangeLaunchURL={setLaunchURL}
        />
        <RedirectURIsCard
          uris={effectiveURIs}
          onAddURI={handleAddURI}
          onRemoveURI={handleRemoveURI}
        />
        <Card>
          <CardHeader>
            <CardTitle>Danger zone</CardTitle>
            <CardDescription>
              Deleting an application removes its OAuth client credentials and all redirect
              URIs. Existing access tokens and refresh tokens issued under this client will
              stop working immediately.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button
              variant="destructive"
              disabled={deleting}
              onClick={() => setDeleteConfirmOpen(true)}
            >
              <Trash2 className="mr-1 size-3.5" />
              Delete application
            </Button>
          </CardContent>
        </Card>
      </div>

      <Dialog
        open={deleteConfirmOpen}
        onOpenChange={(open) => {
          if (!deleting) setDeleteConfirmOpen(open)
        }}
      >
        <DialogContent className="gap-5 sm:max-w-md">
          <DialogHeader className="gap-3">
            <div className="flex size-10 items-center justify-center rounded-xl bg-destructive/10 text-destructive">
              <Trash2 className="size-5" />
            </div>
            <DialogTitle>Delete {app.name}?</DialogTitle>
            <DialogDescription>
              This permanently removes the application along with its OAuth credentials and
              redirect URIs. Any tokens already issued under this client stop validating
              immediately. This can't be undone.
            </DialogDescription>
          </DialogHeader>

          <div className="flex justify-end gap-2 pt-1">
            <Button
              type="button"
              variant="ghost"
              disabled={deleting}
              onClick={() => setDeleteConfirmOpen(false)}
            >
              Cancel
            </Button>
            <Button
              type="button"
              variant="destructive"
              disabled={deleting}
              onClick={handleDelete}
            >
              {deleting ? "Deleting…" : "Delete application"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </PageContainer>
  )
}
