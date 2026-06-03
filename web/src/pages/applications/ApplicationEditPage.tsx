import { useQuery, useQueryClient } from "@tanstack/react-query"
import { ArrowLeft, Plus, ShieldAlert, Trash2, X } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { Link, useNavigate, useParams } from "react-router-dom"
import { toast } from "sonner"

import { OutlineButton } from "@/components/OutlineButton"
import { PageContainer } from "@/components/PageContainer"
import { Badge } from "@/components/ui/badge"
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { Textarea } from "@/components/ui/textarea"
import { api } from "@/lib/api"
import {
  redirectURIWildcardExamples,
  type Application,
  type GroupWithLink,
  type SAMLConfig,
} from "@/lib/applications"
import type { Group } from "@/lib/groups"

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

function LinkedGroupsCard({
  links,
  allGroups,
  onAdd,
  onRemove,
  onToggleRequired,
}: {
  links: { id: string; name: string; required: boolean }[]
  allGroups: Group[]
  onAdd: (groupID: string, required: boolean) => void
  onRemove: (groupID: string) => void
  onToggleRequired: (groupID: string) => void
}) {
  const [draftGroupID, setDraftGroupID] = useState("")
  const [draftRequired, setDraftRequired] = useState(false)

  const linkedSet = useMemo(() => new Set(links.map((l) => l.id)), [links])
  const availableGroups = useMemo(
    () => allGroups.filter((g) => !linkedSet.has(g.id)),
    [allGroups, linkedSet],
  )

  function handleAdd() {
    if (!draftGroupID) return
    onAdd(draftGroupID, draftRequired)
    setDraftGroupID("")
    setDraftRequired(false)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Linked groups</CardTitle>
        <CardDescription>
          Groups linked here flow into the token's <code className="font-mono text-xs">groups</code>{" "}
          claim during OAuth. Mark a link as <span className="font-medium">required</span> to gate
          access — users must be in at least one required group to obtain a token.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {links.length === 0 ? (
          <p className="text-sm text-muted-foreground">No groups linked yet.</p>
        ) : (
          <ul className="space-y-2">
            {links.map((link) => (
              <li
                key={link.id}
                className="flex items-center gap-2 rounded-md border border-border/60 bg-muted/40 px-2.5 py-1.5"
              >
                <span className="flex-1 truncate text-sm">{link.name}</span>
                <Badge
                  variant={link.required ? "default" : "outline"}
                  className="cursor-pointer select-none"
                  onClick={() => onToggleRequired(link.id)}
                  title={
                    link.required
                      ? "Required for access — click to make optional"
                      : "Optional — click to require for access"
                  }
                >
                  {link.required ? "Required" : "Optional"}
                </Badge>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  onClick={() => onRemove(link.id)}
                >
                  <X className="size-3.5" />
                </Button>
              </li>
            ))}
          </ul>
        )}
        <div className="flex flex-wrap items-center gap-2 pt-2">
          <Select value={draftGroupID} onValueChange={setDraftGroupID}>
            <SelectTrigger className="flex-1 min-w-[180px]">
              <SelectValue placeholder="Select a group to link…" />
            </SelectTrigger>
            <SelectContent>
              {availableGroups.length === 0 ? (
                <div className="px-2 py-1.5 text-xs text-muted-foreground">
                  No more groups to link.
                </div>
              ) : (
                availableGroups.map((g) => (
                  <SelectItem key={g.id} value={g.id}>
                    {g.name}
                  </SelectItem>
                ))
              )}
            </SelectContent>
          </Select>
          <Badge
            variant={draftRequired ? "default" : "outline"}
            className="cursor-pointer select-none"
            onClick={() => setDraftRequired((v) => !v)}
          >
            {draftRequired ? "Required" : "Optional"}
          </Badge>
          <Button type="button" disabled={!draftGroupID} onClick={handleAdd}>
            <Plus className="mr-1 size-3.5" />
            Link
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

function SamlConfigCard({
  entityID,
  acsURL,
  metadataXML,
  onChangeEntityID,
  onChangeACSURL,
  onChangeMetadataXML,
}: {
  entityID: string
  acsURL: string
  metadataXML: string
  onChangeEntityID: (v: string) => void
  onChangeACSURL: (v: string) => void
  onChangeMetadataXML: (v: string) => void
}) {
  // The IdP metadata lives at the issuer root (no /api prefix) — admins hand
  // this URL to the SP to establish trust.
  const idpMetadataURL = `${import.meta.env.VITE_API_URL}/saml/metadata`

  return (
    <Card>
      <CardHeader>
        <CardTitle>SAML single sign-on</CardTitle>
        <CardDescription>
          Register this app as a SAML service provider. Leave the entity ID blank to disable
          SAML. Give the SP your IdP metadata URL below, then enter the SP's entity ID and ACS
          URL — or paste its metadata XML to have those derived automatically.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-5">
        <div className="space-y-2">
          <Label>IdP metadata URL</Label>
          <code className="block break-all rounded-md border border-border/60 bg-muted/40 px-2.5 py-1.5 font-mono text-xs">
            {idpMetadataURL}
          </code>
        </div>
        <div className="space-y-2">
          <Label htmlFor="saml_entity_id">SP entity ID</Label>
          <Input
            id="saml_entity_id"
            value={entityID}
            onChange={(e) => onChangeEntityID(e.target.value)}
            placeholder="https://app.gauchoracing.com/saml/metadata"
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="saml_acs_url">Assertion Consumer Service (ACS) URL</Label>
          <Input
            id="saml_acs_url"
            type="url"
            value={acsURL}
            onChange={(e) => onChangeACSURL(e.target.value)}
            placeholder="https://app.gauchoracing.com/saml/acs"
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="saml_metadata_xml">SP metadata XML (optional)</Label>
          <Textarea
            id="saml_metadata_xml"
            value={metadataXML}
            onChange={(e) => onChangeMetadataXML(e.target.value)}
            rows={4}
            placeholder="<EntityDescriptor …>…</EntityDescriptor>"
            className="font-mono text-xs"
          />
          <p className="text-xs text-muted-foreground">
            When provided, the ACS URL and signing certificate are read from the metadata and
            take precedence over the fields above.
          </p>
        </div>
      </CardContent>
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

  const linksQuery = useQuery({
    queryKey: ["application", "id", id, "groups"],
    queryFn: async () => {
      const res = await api.get<GroupWithLink[]>(`/applications/${id}/groups`)
      return res.data
    },
    enabled: !!id,
  })

  const groupsQuery = useQuery({
    queryKey: ["groups"],
    queryFn: async () => {
      const res = await api.get<Group[]>(`/groups`)
      return res.data
    },
    staleTime: 5 * 60 * 1000,
  })

  // SAML SP registration. 404 = no SAML config yet, which is the common case,
  // so a not-found is a successful empty result rather than an error.
  const samlQuery = useQuery({
    queryKey: ["application", "id", id, "saml"],
    queryFn: async () => {
      try {
        const res = await api.get<SAMLConfig>(`/applications/${id}/saml`)
        return res.data
      } catch (err) {
        if ((err as { response?: { status?: number } })?.response?.status === 404) {
          return null
        }
        throw err
      }
    },
    enabled: !!id,
    retry: false,
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

  // Staged group-link state. Map<group_id, required>. Initialized from the
  // server response on first load; mutations only touch this map until Save.
  const [linkState, setLinkState] = useState<Map<string, boolean>>(new Map())
  const [linksInitialized, setLinksInitialized] = useState(false)

  // Staged SAML config — applied on Save. samlExisted tracks whether the server
  // already had a config so Save can DELETE it when the entity ID is cleared.
  const [samlEntityID, setSamlEntityID] = useState("")
  const [samlACSURL, setSamlACSURL] = useState("")
  const [samlMetadataXML, setSamlMetadataXML] = useState("")
  const [samlExisted, setSamlExisted] = useState(false)
  const [samlInitialized, setSamlInitialized] = useState(false)

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

  useEffect(() => {
    if (linksQuery.data && !linksInitialized) {
      const m = new Map<string, boolean>()
      for (const l of linksQuery.data) m.set(l.id, l.required)
      setLinkState(m)
      setLinksInitialized(true)
    }
  }, [linksQuery.data, linksInitialized])

  useEffect(() => {
    if (!samlInitialized && (samlQuery.data !== undefined || samlQuery.isError)) {
      const cfg = samlQuery.data ?? null
      setSamlEntityID(cfg?.entity_id ?? "")
      setSamlACSURL(cfg?.acs_url ?? "")
      setSamlMetadataXML(cfg?.metadata_xml ?? "")
      setSamlExisted(cfg !== null)
      setSamlInitialized(true)
    }
  }, [samlQuery.data, samlQuery.isError, samlInitialized])

  // Render rows for the card: each link gets its current desired-required
  // state from linkState, name from groupsQuery (canonical source). Falls
  // back to the group_id when names haven't loaded.
  const linkList = useMemo(() => {
    const byID = new Map((groupsQuery.data ?? []).map((g) => [g.id, g.name]))
    return Array.from(linkState, ([id, required]) => ({
      id,
      name: byID.get(id) ?? id,
      required,
    }))
  }, [linkState, groupsQuery.data])

  function handleAddGroupLink(groupID: string, required: boolean) {
    setLinkState((prev) => {
      const next = new Map(prev)
      next.set(groupID, required)
      return next
    })
  }

  function handleRemoveGroupLink(groupID: string) {
    setLinkState((prev) => {
      const next = new Map(prev)
      next.delete(groupID)
      return next
    })
  }

  function handleToggleGroupRequired(groupID: string) {
    setLinkState((prev) => {
      const next = new Map(prev)
      const cur = next.get(groupID)
      if (cur === undefined) return prev
      next.set(groupID, !cur)
      return next
    })
  }

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

      // Diff group links against the server state. POST is upsert, so we
      // send any link whose required flag differs (or doesn't exist yet);
      // DELETE anything the server has that's no longer in our state.
      const serverLinks = linksQuery.data ?? []
      const serverByID = new Map(serverLinks.map((l) => [l.id, l.required]))
      for (const [groupID, required] of linkState) {
        const serverReq = serverByID.get(groupID)
        if (serverReq === undefined || serverReq !== required) {
          await api.post(`/applications/${id}/groups`, {
            group_id: groupID,
            required,
          })
        }
      }
      for (const l of serverLinks) {
        if (!linkState.has(l.id)) {
          await api.delete(`/applications/${id}/groups/${l.id}`)
        }
      }

      // SAML config: upsert when an entity ID is set, delete when it's been
      // cleared on an app that previously had one.
      const samlEntity = samlEntityID.trim()
      if (samlEntity) {
        await api.post(`/applications/${id}/saml`, {
          entity_id: samlEntity,
          acs_url: samlACSURL.trim(),
          metadata_xml: samlMetadataXML.trim(),
        })
      } else if (samlExisted) {
        await api.delete(`/applications/${id}/saml`)
      }

      await api.put(`/applications/${id}`, {
        name,
        description,
        icon_url: iconURL,
        launch_url: launchURL,
      })
      qc.invalidateQueries({ queryKey: ["application", "id", id] })
      qc.invalidateQueries({ queryKey: ["application", "id", id, "groups"] })
      qc.invalidateQueries({ queryKey: ["application", "id", id, "saml"] })
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

  if (query.isLoading || !initialized || !linksInitialized || !samlInitialized) {
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
        <LinkedGroupsCard
          links={linkList}
          allGroups={groupsQuery.data ?? []}
          onAdd={handleAddGroupLink}
          onRemove={handleRemoveGroupLink}
          onToggleRequired={handleToggleGroupRequired}
        />
        <SamlConfigCard
          entityID={samlEntityID}
          acsURL={samlACSURL}
          metadataXML={samlMetadataXML}
          onChangeEntityID={setSamlEntityID}
          onChangeACSURL={setSamlACSURL}
          onChangeMetadataXML={setSamlMetadataXML}
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
