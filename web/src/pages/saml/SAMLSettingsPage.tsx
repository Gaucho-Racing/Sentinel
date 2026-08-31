import { useQuery, useQueryClient } from "@tanstack/react-query"
import { ArrowLeft, Copy, Plus, Trash2, X } from "lucide-react"
import { useState } from "react"
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
import type { Application } from "@/lib/applications"
import {
  GENERIC_SAML_ATTRIBUTE_MAPPINGS,
  SAML_ATTRIBUTE_FORMAT_BASIC,
  SAML_ATTRIBUTE_FORMAT_UNSPECIFIED,
  SAML_ATTRIBUTE_FORMAT_URI,
  SAML_NAME_ID_FORMAT_EMAIL,
  SAML_NAME_ID_FORMAT_PERSISTENT,
  SAML_NAME_ID_FORMAT_UNSPECIFIED,
  newSAMLConfiguration,
  type SAMLAssertionPreview,
  type SAMLAttributeMapping,
  type SAMLAttributeSource,
  type SAMLConfiguration,
  type SAMLNameIDSource,
  type SAMLProfile,
} from "@/lib/saml"
import { useUsers, userName } from "@/lib/users"

const attributeSources: Array<{ value: SAMLAttributeSource; label: string; multi: boolean }> = [
  { value: "EMAIL", label: "Email", multi: false },
  { value: "USERNAME", label: "Username", multi: false },
  { value: "FIRST_NAME", label: "First name", multi: false },
  { value: "LAST_NAME", label: "Last name", multi: false },
  { value: "DISPLAY_NAME", label: "Display name", multi: false },
  { value: "ENTITY_ID", label: "Entity ID", multi: false },
  { value: "GROUP_NAMES", label: "Group names", multi: true },
  { value: "GROUP_IDS", label: "Group IDs", multi: true },
  { value: "CONSTANT", label: "Constant", multi: false },
]

function apiError(error: unknown, fallback: string) {
  return (error as { response?: { data?: { error?: string } } })?.response?.data?.error ?? fallback
}

function AttributeMappingEditor({
  mapping,
  awsProfile,
  onChange,
  onRemove,
}: {
  mapping: SAMLAttributeMapping
  awsProfile: boolean
  onChange: (mapping: SAMLAttributeMapping) => void
  onRemove: () => void
}) {
  return (
    <div className="space-y-4 rounded-lg border border-border/60 bg-muted/20 p-4">
      <div className="flex items-start gap-3">
        <div className="grid flex-1 gap-4 md:grid-cols-2">
          <div className="space-y-2">
            <Label>Attribute name</Label>
            <Input
              value={mapping.name}
              onChange={(event) => onChange({ ...mapping, name: event.target.value })}
              placeholder="groups"
              className="font-mono text-xs"
            />
          </div>
          <div className="space-y-2">
            <Label>Friendly name</Label>
            <Input
              value={mapping.friendly_name}
              onChange={(event) => onChange({ ...mapping, friendly_name: event.target.value })}
              placeholder="groups"
            />
          </div>
          <div className="space-y-2">
            <Label>Value source</Label>
            <Select
              value={mapping.source}
              onValueChange={(value) =>
                onChange({ ...mapping, source: value as SAMLAttributeSource })
              }
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {attributeSources.map((source) => (
                  <SelectItem
                    key={source.value}
                    value={source.value}
                    disabled={awsProfile && source.multi}
                  >
                    {source.label}
                    {source.multi ? " (multi-valued)" : ""}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>Name format</Label>
            <Select
              value={mapping.name_format}
              onValueChange={(value) => onChange({ ...mapping, name_format: value })}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={SAML_ATTRIBUTE_FORMAT_BASIC}>Basic</SelectItem>
                <SelectItem value={SAML_ATTRIBUTE_FORMAT_URI}>URI</SelectItem>
                <SelectItem value={SAML_ATTRIBUTE_FORMAT_UNSPECIFIED}>Unspecified</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <Button type="button" variant="ghost" size="icon-sm" onClick={onRemove}>
          <X className="size-3.5" />
        </Button>
      </div>
      {mapping.source === "CONSTANT" && (
        <div className="space-y-2">
          <Label>Constant value</Label>
          <Input
            value={mapping.constant}
            onChange={(event) => onChange({ ...mapping, constant: event.target.value })}
          />
        </div>
      )}
      <label className="flex cursor-pointer items-center gap-2 text-sm">
        <input
          type="checkbox"
          checked={mapping.omit_if_empty}
          onChange={(event) => onChange({ ...mapping, omit_if_empty: event.target.checked })}
          className="size-4 accent-gr-pink"
        />
        Omit this attribute when it has no value
      </label>
    </div>
  )
}

export default function SAMLSettingsPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const usersQuery = useUsers({ enabled: !!id })
  const [draft, setDraft] = useState<SAMLConfiguration | null>(null)
  const [existsOverride, setExistsOverride] = useState<boolean | null>(null)
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [previewEntityID, setPreviewEntityID] = useState("")
  const [preview, setPreview] = useState<SAMLAssertionPreview | null>(null)
  const [previewing, setPreviewing] = useState(false)

  const appQuery = useQuery({
    queryKey: ["application", "id", id],
    queryFn: async () => (await api.get<Application>(`/applications/${id}`)).data,
    enabled: !!id,
  })

  const configQuery = useQuery({
    queryKey: ["application", "id", id, "saml"],
    queryFn: async () => {
      try {
        return (await api.get<SAMLConfiguration>(`/saml/applications/${id}/config`)).data
      } catch (error) {
        if ((error as { response?: { status?: number } })?.response?.status === 404) {
          return null
        }
        throw error
      }
    },
    enabled: !!id,
    retry: false,
  })

  const configuration =
    id && configQuery.data !== undefined
      ? (draft ?? configQuery.data ?? newSAMLConfiguration(id))
      : null
  const exists = existsOverride ?? configQuery.data != null

  function updateConfiguration(changes: Partial<SAMLConfiguration>) {
    if (!configuration) return
    setDraft({ ...configuration, ...changes })
    setPreview(null)
  }

  function changeProfile(profile: SAMLProfile) {
    if (!configuration) return
    if (profile === "AWS_IDENTITY_CENTER") {
      updateConfiguration({
        profile,
        name_id_source: "EMAIL",
        name_id_format: SAML_NAME_ID_FORMAT_EMAIL,
        attribute_mappings: configuration.attribute_mappings.filter(
          (mapping) => mapping.source !== "GROUP_NAMES" && mapping.source !== "GROUP_IDS",
        ),
      })
      return
    }
    updateConfiguration({ profile })
  }

  function changeMapping(index: number, mapping: SAMLAttributeMapping) {
    if (!configuration) return
    const mappings = [...configuration.attribute_mappings]
    mappings[index] = mapping
    updateConfiguration({ attribute_mappings: mappings })
  }

  function addMapping() {
    if (!configuration) return
    updateConfiguration({
      attribute_mappings: [
        ...configuration.attribute_mappings,
        {
          name: "",
          friendly_name: "",
          name_format: SAML_ATTRIBUTE_FORMAT_BASIC,
          source: "EMAIL",
          constant: "",
          omit_if_empty: true,
        },
      ],
    })
  }

  async function save() {
    if (!id || !configuration || saving) return
    setSaving(true)
    try {
      const stored = (
        await api.put<SAMLConfiguration>(`/saml/applications/${id}/config`, configuration)
      ).data
      setDraft(stored)
      setExistsOverride(true)
      queryClient.setQueryData(["application", "id", id, "saml"], stored)
      toast.success("SAML configuration saved")
    } catch (error) {
      toast.error(apiError(error, "Couldn't save the SAML configuration."))
    } finally {
      setSaving(false)
    }
  }

  async function removeConfiguration() {
    if (!id || deleting) return
    setDeleting(true)
    try {
      await api.delete(`/saml/applications/${id}/config`)
      queryClient.setQueryData(["application", "id", id, "saml"], null)
      toast.success("SAML disabled")
      navigate(`/applications/${id}/edit`)
    } catch (error) {
      toast.error(apiError(error, "Couldn't disable SAML."))
      setDeleting(false)
      setDeleteOpen(false)
    }
  }

  async function loadPreview() {
    if (!id || !configuration || !previewEntityID || previewing) return
    setPreviewing(true)
    try {
      const result = (
        await api.post<SAMLAssertionPreview>(`/saml/applications/${id}/preview`, {
          entity_id: previewEntityID,
          configuration,
        })
      ).data
      setPreview(result)
    } catch (error) {
      setPreview(null)
      toast.error(apiError(error, "Couldn't resolve the assertion preview."))
    } finally {
      setPreviewing(false)
    }
  }

  if (appQuery.isLoading || configQuery.isLoading || !configuration) {
    return (
      <PageContainer>
        <Skeleton className="mb-4 h-4 w-32" />
        <Skeleton className="mb-8 h-8 w-64" />
        <div className="space-y-4">
          <Skeleton className="h-72" />
          <Skeleton className="h-56" />
        </div>
      </PageContainer>
    )
  }

  if (appQuery.isError || configQuery.isError || !appQuery.data) {
    return (
      <PageContainer>
        <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
          <Link to={`/applications/${id}/edit`}>
            <ArrowLeft className="mr-1 size-3.5" />
            Application settings
          </Link>
        </Button>
        <p className="text-sm text-muted-foreground">
          {configQuery.isError ? "You cannot manage SAML for this application." : "Application not found."}
        </p>
      </PageContainer>
    )
  }

  const application = appQuery.data
  const awsProfile = configuration.profile === "AWS_IDENTITY_CENTER"
  const metadataProvided = configuration.metadata_xml.trim() !== ""
  const idpMetadataURL = new URL(
    "/saml/metadata",
    import.meta.env.VITE_API_URL || window.location.origin,
  ).toString()

  return (
    <PageContainer>
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to={`/applications/${id}/edit`}>
          <ArrowLeft className="mr-1 size-3.5" />
          Back to {application.name} settings
        </Link>
      </Button>

      <header className="mb-8 flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-semibold tracking-tight">SAML configuration</h1>
            <Badge variant={exists ? "default" : "outline"}>{exists ? "Enabled" : "Not configured"}</Badge>
          </div>
          <p className="mt-1 text-sm text-muted-foreground">
            Configure {application.name} as a SAML service provider.
          </p>
        </div>
        <OutlineButton type="button" className="w-auto" loading={saving} onClick={save}>
          Save configuration
        </OutlineButton>
      </header>

      <div className="mb-6 flex gap-2">
        <Button size="sm">Single sign-on</Button>
        <Button asChild variant="outline" size="sm">
          <Link to={`/applications/${id}/saml/scim`}>Provisioning</Link>
        </Button>
      </div>

      <div className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>Integration</CardTitle>
            <CardDescription>
              Choose the relying-party profile and exchange metadata with the service provider.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-5">
            <div className="space-y-2">
              <Label>Profile</Label>
              <Select value={configuration.profile} onValueChange={(value) => changeProfile(value as SAMLProfile)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="GENERIC">Generic SAML 2.0</SelectItem>
                  <SelectItem value="AWS_IDENTITY_CENTER">AWS IAM Identity Center</SelectItem>
                </SelectContent>
              </Select>
              {awsProfile && (
                <p className="text-xs text-muted-foreground">
                  Uses an email-format NameID and prevents multi-valued group claims. Linked groups and users can be provisioned from the Provisioning tab.
                </p>
              )}
            </div>
            <div className="space-y-2">
              <Label>Sentinel IdP metadata URL</Label>
              <div className="flex items-center gap-1 rounded-md border border-border/60 bg-muted/40 px-2.5 py-1.5">
                <code className="flex-1 break-all font-mono text-xs">{idpMetadataURL}</code>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  onClick={async () => {
                    await navigator.clipboard.writeText(idpMetadataURL)
                    toast.success("IdP metadata URL copied")
                  }}
                >
                  <Copy className="size-3.5" />
                </Button>
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="metadata_xml">Service provider metadata XML</Label>
              <Textarea
                id="metadata_xml"
                value={configuration.metadata_xml}
                onChange={(event) => {
                  const metadataXML = event.target.value
                  updateConfiguration(
                    metadataXML.trim()
                      ? { metadata_xml: metadataXML, entity_id: "", acs_url: "" }
                      : { metadata_xml: metadataXML },
                  )
                }}
                rows={7}
                placeholder="<EntityDescriptor …>…</EntityDescriptor>"
                className="font-mono text-xs"
              />
              <p className="text-xs text-muted-foreground">
                When present, Sentinel validates this metadata and derives the entity ID and all accepted ACS endpoints from it.
              </p>
            </div>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="entity_id">SP entity ID</Label>
                <Input
                  id="entity_id"
                  value={configuration.entity_id}
                  onChange={(event) => updateConfiguration({ entity_id: event.target.value })}
                  disabled={metadataProvided}
                  placeholder="https://service.example.com/saml/metadata"
                  className="font-mono text-xs"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="acs_url">ACS URL</Label>
                <Input
                  id="acs_url"
                  type="url"
                  value={configuration.acs_url}
                  onChange={(event) => updateConfiguration({ acs_url: event.target.value })}
                  disabled={metadataProvided}
                  placeholder="https://service.example.com/saml/acs"
                  className="font-mono text-xs"
                />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Subject</CardTitle>
            <CardDescription>Choose the Sentinel identity value emitted as the assertion NameID.</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label>NameID source</Label>
              <Select
                value={configuration.name_id_source}
                disabled={awsProfile}
                onValueChange={(value) => updateConfiguration({ name_id_source: value as SAMLNameIDSource })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="EMAIL">Email</SelectItem>
                  <SelectItem value="USERNAME">Username</SelectItem>
                  <SelectItem value="ENTITY_ID">Entity ID</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>NameID format</Label>
              <Select
                value={configuration.name_id_format}
                disabled={awsProfile}
                onValueChange={(value) => updateConfiguration({ name_id_format: value })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={SAML_NAME_ID_FORMAT_EMAIL}>Email address</SelectItem>
                  <SelectItem value={SAML_NAME_ID_FORMAT_PERSISTENT}>Persistent</SelectItem>
                  <SelectItem value={SAML_NAME_ID_FORMAT_UNSPECIFIED}>Unspecified</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex-row items-start justify-between gap-4">
            <div>
              <CardTitle>Attribute mappings</CardTitle>
              <CardDescription className="mt-1.5">
                Emit only the attributes this service provider consumes. Attribute names must be unique.
              </CardDescription>
            </div>
            <Button type="button" variant="outline" size="sm" onClick={addMapping}>
              <Plus className="mr-1 size-3.5" />
              Add attribute
            </Button>
          </CardHeader>
          <CardContent className="space-y-3">
            {configuration.attribute_mappings.length === 0 ? (
              <div className="rounded-lg border border-dashed p-6 text-center">
                <p className="text-sm text-muted-foreground">No attributes will be emitted.</p>
                {!awsProfile && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="mt-2"
                    onClick={() =>
                      updateConfiguration({
                        attribute_mappings: GENERIC_SAML_ATTRIBUTE_MAPPINGS.map((mapping) => ({ ...mapping })),
                      })
                    }
                  >
                    Load generic defaults
                  </Button>
                )}
              </div>
            ) : (
              configuration.attribute_mappings.map((mapping, index) => (
                <AttributeMappingEditor
                  key={index}
                  mapping={mapping}
                  awsProfile={awsProfile}
                  onChange={(next) => changeMapping(index, next)}
                  onRemove={() =>
                    updateConfiguration({
                      attribute_mappings: configuration.attribute_mappings.filter((_, current) => current !== index),
                    })
                  }
                />
              ))
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Assertion preview</CardTitle>
            <CardDescription>
              Resolve the current unsaved configuration for a Sentinel user. This does not issue or sign an assertion.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-wrap gap-2">
              <Select value={previewEntityID} onValueChange={(value) => { setPreviewEntityID(value); setPreview(null) }}>
                <SelectTrigger className="min-w-[240px] flex-1">
                  <SelectValue placeholder="Select a user" />
                </SelectTrigger>
                <SelectContent>
                  {(usersQuery.data ?? []).map((user) => (
                    <SelectItem key={user.entity_id} value={user.entity_id}>
                      {userName(user)}{user.email ? ` · ${user.email}` : ""}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Button type="button" disabled={!previewEntityID || previewing} onClick={loadPreview}>
                {previewing ? "Resolving…" : "Preview"}
              </Button>
            </div>
            {preview && (
              <div className="space-y-4 rounded-lg border border-border/60 bg-muted/20 p-4">
                <div>
                  <p className="text-xs uppercase tracking-wider text-muted-foreground">NameID</p>
                  <code className="mt-1 block break-all font-mono text-xs">{preview.name_id}</code>
                  <code className="mt-1 block break-all font-mono text-[11px] text-muted-foreground">
                    {preview.name_id_format}
                  </code>
                </div>
                <div className="space-y-2">
                  <p className="text-xs uppercase tracking-wider text-muted-foreground">Attributes</p>
                  {preview.attributes.length === 0 ? (
                    <p className="text-sm text-muted-foreground">No attributes.</p>
                  ) : (
                    preview.attributes.map((attribute) => (
                      <div key={attribute.name} className="rounded-md border border-border/60 bg-background px-3 py-2">
                        <code className="block break-all font-mono text-xs">{attribute.name}</code>
                        <div className="mt-1 flex flex-wrap gap-1">
                          {attribute.values.map((value, index) => (
                            <Badge key={`${value}-${index}`} variant="outline" className="font-mono text-[11px]">
                              {value || "(empty)"}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    ))
                  )}
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {exists && (
          <Card>
            <CardHeader>
              <CardTitle>Disable SAML</CardTitle>
              <CardDescription>Remove this application's SAML service-provider registration.</CardDescription>
            </CardHeader>
            <CardContent>
              <Button variant="destructive" onClick={() => setDeleteOpen(true)}>
                <Trash2 className="mr-1 size-3.5" />
                Disable SAML
              </Button>
            </CardContent>
          </Card>
        )}
      </div>

      <Dialog open={deleteOpen} onOpenChange={(open) => !deleting && setDeleteOpen(open)}>
        <DialogContent className="gap-5 sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Disable SAML for {application.name}?</DialogTitle>
            <DialogDescription>
              Sentinel will stop accepting authentication requests from this service provider. This does not delete the application.
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-2">
            <Button variant="ghost" disabled={deleting} onClick={() => setDeleteOpen(false)}>
              Cancel
            </Button>
            <Button variant="destructive" disabled={deleting} onClick={removeConfiguration}>
              {deleting ? "Disabling…" : "Disable SAML"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </PageContainer>
  )
}
