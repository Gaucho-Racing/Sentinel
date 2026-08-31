import { useQuery, useQueryClient } from "@tanstack/react-query"
import { ArrowLeft, CheckCircle2, RefreshCw, Trash2 } from "lucide-react"
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
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import type { Application, GroupWithLink } from "@/lib/applications"
import type { SCIMConfiguration, SCIMSyncResult } from "@/lib/scim"
import type { SAMLConfiguration } from "@/lib/saml"

function apiError(error: unknown, fallback: string) {
  return (error as { response?: { data?: { error?: string } } })?.response?.data?.error ?? fallback
}

function statusBadge(configuration: SCIMConfiguration | null | undefined) {
  if (!configuration) return <Badge variant="outline">Not configured</Badge>
  if (!configuration.enabled) return <Badge variant="outline">Disabled</Badge>
  if (configuration.token_expires_at && new Date(configuration.token_expires_at) <= new Date()) {
    return <Badge variant="destructive">Token expired</Badge>
  }
  if (configuration.last_sync_status === "FAILED") return <Badge variant="destructive">Sync failed</Badge>
  if (configuration.last_sync_status === "SUCCEEDED") return <Badge>Healthy</Badge>
  return <Badge>Enabled</Badge>
}

function ResultSummary({ result }: { result: SCIMSyncResult }) {
  return (
    <div className="space-y-3 rounded-lg border border-border/60 bg-muted/20 p-4">
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <div>
          <p className="text-xs text-muted-foreground">Users in scope</p>
          <p className="text-lg font-medium">{result.desired_users}</p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Groups in scope</p>
          <p className="text-lg font-medium">{result.desired_groups}</p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Memberships added</p>
          <p className="text-lg font-medium">{result.memberships_added}</p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Memberships removed</p>
          <p className="text-lg font-medium">{result.memberships_removed}</p>
        </div>
      </div>
      {!result.dry_run && (
        <p className="text-xs text-muted-foreground">
          Users: {result.users_created} created, {result.users_updated} updated, {result.users_deactivated} deactivated. Groups: {result.groups_created} created, {result.groups_updated} updated.
        </p>
      )}
      {result.skipped_users.length > 0 && (
        <div className="space-y-1 text-sm text-amber-400">
          <p className="font-medium">
            {result.skipped_users.length} {result.skipped_users.length === 1 ? "user was" : "users were"} skipped and excluded from AWS provisioning.
          </p>
          <ul className="list-disc space-y-1 pl-5">
            {result.skipped_users.map((user) => (
              <li key={user.entity_id}>
                {user.username || user.entity_id}: {user.reason}
                {user.groups.length > 0 ? ` (${user.groups.join(", ")})` : ""}
              </li>
            ))}
          </ul>
        </div>
      )}
      {result.validation_errors.length > 0 && (
        <ul className="space-y-1 text-sm text-destructive">
          {result.validation_errors.map((error) => <li key={error}>{error}</li>)}
        </ul>
      )}
    </div>
  )
}

export default function SCIMSettingsPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [draft, setDraft] = useState<{
    endpoint: string
    accessToken: string
    tokenExpiresAt: string
    enabled: boolean
  } | null>(null)
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [previewing, setPreviewing] = useState(false)
  const [syncing, setSyncing] = useState(false)
  const [removing, setRemoving] = useState(false)
  const [removeOpen, setRemoveOpen] = useState(false)
  const [result, setResult] = useState<SCIMSyncResult | null>(null)

  const appQuery = useQuery({
    queryKey: ["application", "id", id],
    queryFn: async () => (await api.get<Application>(`/applications/${id}`)).data,
    enabled: !!id,
  })
  const samlQuery = useQuery({
    queryKey: ["application", "id", id, "saml"],
    queryFn: async () => {
      try {
        return (await api.get<SAMLConfiguration>(`/saml/applications/${id}/config`)).data
      } catch (error) {
        if ((error as { response?: { status?: number } })?.response?.status === 404) return null
        throw error
      }
    },
    enabled: !!id,
    retry: false,
  })
  const configQuery = useQuery({
    queryKey: ["application", "id", id, "scim"],
    queryFn: async () => {
      try {
        return (await api.get<SCIMConfiguration>(`/saml/applications/${id}/scim/config`)).data
      } catch (error) {
        if ((error as { response?: { status?: number } })?.response?.status === 404) return null
        throw error
      }
    },
    enabled: !!id,
    retry: false,
  })
  const groupsQuery = useQuery({
    queryKey: ["application", "id", id, "groups"],
    queryFn: async () => (await api.get<GroupWithLink[]>(`/applications/${id}/groups`)).data,
    enabled: !!id,
  })

  const form = draft ?? {
    endpoint: configQuery.data?.endpoint ?? "",
    accessToken: "",
    tokenExpiresAt: configQuery.data?.token_expires_at?.slice(0, 10) ?? "",
    enabled: configQuery.data?.enabled ?? true,
  }

  function updateDraft(changes: Partial<typeof form>) {
    setDraft({ ...form, ...changes })
  }

  async function save() {
    if (!id || saving) return
    setSaving(true)
    try {
      const stored = (await api.put<SCIMConfiguration>(`/saml/applications/${id}/scim/config`, {
        endpoint: form.endpoint,
        access_token: form.accessToken,
        token_expires_at: form.tokenExpiresAt ? new Date(`${form.tokenExpiresAt}T23:59:59`).toISOString() : null,
        enabled: form.enabled,
      })).data
      queryClient.setQueryData(["application", "id", id, "scim"], stored)
      setDraft({
        endpoint: stored.endpoint,
        accessToken: "",
        tokenExpiresAt: stored.token_expires_at?.slice(0, 10) ?? "",
        enabled: stored.enabled,
      })
      toast.success("SCIM configuration saved")
    } catch (error) {
      toast.error(apiError(error, "Couldn't save the SCIM configuration."))
    } finally {
      setSaving(false)
    }
  }

  async function testConnection() {
    if (!id || testing) return
    setTesting(true)
    try {
      await api.post(`/saml/applications/${id}/scim/test`)
      toast.success("Connected to the AWS SCIM endpoint")
    } catch (error) {
      toast.error(apiError(error, "Couldn't connect to the SCIM endpoint."))
    } finally {
      setTesting(false)
    }
  }

  async function previewScope() {
    if (!id || previewing) return
    setPreviewing(true)
    try {
      const preview = (await api.post<SCIMSyncResult>(`/saml/applications/${id}/scim/preview`)).data
      setResult(preview)
      if (preview.validation_errors.length === 0) {
        if (preview.skipped_users.length > 0) toast.warning("Provisioning scope is valid with skipped users")
        else toast.success("Provisioning scope is valid")
      }
    } catch (error) {
      toast.error(apiError(error, "Couldn't validate the provisioning scope."))
    } finally {
      setPreviewing(false)
    }
  }

  async function synchronize() {
    if (!id || syncing) return
    setSyncing(true)
    try {
      const syncResult = (await api.post<SCIMSyncResult>(`/saml/applications/${id}/scim/sync`)).data
      setResult(syncResult)
      await queryClient.invalidateQueries({ queryKey: ["application", "id", id, "scim"] })
      if (syncResult.skipped_users.length > 0) toast.warning("AWS provisioning synchronized with skipped users")
      else toast.success("AWS provisioning synchronized")
    } catch (error) {
      const response = (error as { response?: { data?: { result?: SCIMSyncResult } } }).response?.data
      if (response?.result) setResult(response.result)
      toast.error(apiError(error, "Couldn't synchronize AWS provisioning."))
    } finally {
      setSyncing(false)
    }
  }

  async function removeConfiguration() {
    if (!id || removing) return
    setRemoving(true)
    try {
      await api.delete(`/saml/applications/${id}/scim/config`)
      queryClient.setQueryData(["application", "id", id, "scim"], null)
      toast.success("SCIM configuration removed")
      navigate(`/applications/${id}/saml`)
    } catch (error) {
      toast.error(apiError(error, "Couldn't remove the SCIM configuration."))
    } finally {
      setRemoving(false)
    }
  }

  if (appQuery.isLoading || samlQuery.isLoading || configQuery.isLoading) {
    return (
      <PageContainer>
        <Skeleton className="mb-4 h-4 w-32" />
        <Skeleton className="mb-8 h-8 w-64" />
        <Skeleton className="h-96" />
      </PageContainer>
    )
  }

  if (appQuery.isError || samlQuery.isError || configQuery.isError || !appQuery.data) {
    return (
      <PageContainer>
        <p className="text-sm text-muted-foreground">
          You cannot manage provisioning for this application.
        </p>
      </PageContainer>
    )
  }

  const configuration = configQuery.data
  const awsProfile = samlQuery.data?.profile === "AWS_IDENTITY_CENTER"
  const tokenExpired = !!configuration?.token_expires_at && new Date(configuration.token_expires_at) <= new Date()

  return (
    <PageContainer>
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to={`/applications/${id}`}>
          <ArrowLeft className="mr-1 size-3.5" />
          Back to {appQuery.data.name}
        </Link>
      </Button>

      <header className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-semibold tracking-tight">AWS provisioning</h1>
            {statusBadge(configuration)}
          </div>
          <p className="mt-1 text-sm text-muted-foreground">
            Provision linked Sentinel groups and their members into AWS IAM Identity Center.
          </p>
        </div>
        {awsProfile && (
          <OutlineButton type="button" className="w-auto" loading={saving} onClick={save}>
            Save configuration
          </OutlineButton>
        )}
      </header>

      <div className="mb-6 flex gap-2">
        <Button asChild variant="outline" size="sm">
          <Link to={`/applications/${id}/saml`}>Single sign-on</Link>
        </Button>
        <Button size="sm">Provisioning</Button>
      </div>

      {!awsProfile ? (
        <Card><CardHeader><CardTitle>AWS SAML profile required</CardTitle><CardDescription>Save this application with the AWS IAM Identity Center SAML profile before configuring provisioning.</CardDescription></CardHeader><CardContent><Button asChild><Link to={`/applications/${id}/saml`}>Configure SAML</Link></Button></CardContent></Card>
      ) : (
        <div className="space-y-4">
          <Card>
            <CardHeader><CardTitle>Connection</CardTitle><CardDescription>Use the SCIM endpoint and access token generated in IAM Identity Center. The token is write-only after saving.</CardDescription></CardHeader>
            <CardContent className="space-y-5">
              <div className="space-y-2"><Label htmlFor="scim_endpoint">SCIM endpoint</Label><Input id="scim_endpoint" type="url" value={form.endpoint} onChange={(event) => updateDraft({ endpoint: event.target.value })} placeholder="https://scim.us-east-1.amazonaws.com/..." className="font-mono text-xs" /></div>
              <div className="space-y-2"><Label htmlFor="scim_token">Access token</Label><Input id="scim_token" type="password" value={form.accessToken} onChange={(event) => updateDraft({ accessToken: event.target.value })} placeholder={configuration?.token_configured ? "Leave blank to keep the current token" : "Paste the AWS access token"} autoComplete="new-password" /></div>
              <div className="space-y-2"><Label htmlFor="scim_expiry">Token expiration</Label><Input id="scim_expiry" type="date" value={form.tokenExpiresAt} onChange={(event) => updateDraft({ tokenExpiresAt: event.target.value })} /><p className="text-xs text-muted-foreground">AWS IAM Identity Center SCIM tokens expire after one year.</p></div>
              <label className="flex cursor-pointer items-center gap-2 text-sm"><input type="checkbox" checked={form.enabled} onChange={(event) => updateDraft({ enabled: event.target.checked })} className="size-4 accent-gr-pink" />Enable manual synchronization</label>
              {configuration && <Button type="button" variant="outline" disabled={testing} onClick={testConnection}><CheckCircle2 className="mr-1 size-3.5" />{testing ? "Testing…" : "Test connection"}</Button>}
            </CardContent>
          </Card>

          <Card>
            <CardHeader><CardTitle>Provisioning scope</CardTitle><CardDescription>Every active human member of a linked group is provisioned. AWS account and permission-set assignments remain managed in AWS.</CardDescription></CardHeader>
            <CardContent className="space-y-4">
              {groupsQuery.isLoading ? <Skeleton className="h-12" /> : (groupsQuery.data?.length ?? 0) === 0 ? <p className="text-sm text-muted-foreground">No groups are linked to this application.</p> : <div className="flex flex-wrap gap-2">{groupsQuery.data?.map((group) => <Badge key={group.id} variant="outline">{group.name} · {group.member_count} members</Badge>)}</div>}
              <Button type="button" variant="outline" disabled={previewing} onClick={previewScope}>{previewing ? "Validating…" : "Validate scope"}</Button>
              {result?.dry_run && <ResultSummary result={result} />}
            </CardContent>
          </Card>

          <Card>
            <CardHeader><CardTitle>Synchronization</CardTitle><CardDescription>Runs an idempotent reconciliation now. Users leaving all linked groups are deactivated; Sentinel never deletes AWS users or groups.</CardDescription></CardHeader>
            <CardContent className="space-y-4">
              {configuration?.last_sync_at && <p className="text-sm text-muted-foreground">Last attempt {new Date(configuration.last_sync_at).toLocaleString()} · {configuration.last_sync_status.toLowerCase()}</p>}
              {configuration?.last_sync_error && <p className="rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">{configuration.last_sync_error}</p>}
              <Button type="button" disabled={!configuration?.enabled || tokenExpired || syncing} onClick={synchronize}><RefreshCw className={`mr-1 size-3.5 ${syncing ? "animate-spin" : ""}`} />{syncing ? "Synchronizing…" : "Sync now"}</Button>
              {result && !result.dry_run && <ResultSummary result={result} />}
            </CardContent>
          </Card>

          {configuration && (
            <Card>
              <CardHeader>
                <CardTitle>Remove provisioning</CardTitle>
                <CardDescription>
                  Deletes Sentinel's SCIM configuration and resource mappings. Existing AWS users and groups are left untouched.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Button variant="destructive" onClick={() => setRemoveOpen(true)}>
                  <Trash2 className="mr-1 size-3.5" />
                  Remove SCIM configuration
                </Button>
              </CardContent>
            </Card>
          )}
        </div>
      )}

      <Dialog open={removeOpen} onOpenChange={setRemoveOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Remove SCIM provisioning?</DialogTitle>
            <DialogDescription>
              Sentinel will forget the AWS connection and managed resource mappings. It will not delete or deactivate anything in AWS.
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => setRemoveOpen(false)}>Cancel</Button>
            <Button variant="destructive" disabled={removing} onClick={removeConfiguration}>
              {removing ? "Removing…" : "Remove provisioning"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </PageContainer>
  )
}
