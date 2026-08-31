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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import type { Application, GroupWithLink } from "@/lib/applications"
import type { SCIMConfiguration, SCIMSyncInterval, SCIMSyncResult, SCIMSyncRun } from "@/lib/scim"
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
  if (configuration.last_sync_status === "QUEUED") return <Badge variant="outline">Sync queued</Badge>
  if (configuration.last_sync_status === "RUNNING") return <Badge variant="outline">Syncing</Badge>
  if (configuration.last_sync_status === "FAILED") return <Badge variant="destructive">Sync failed</Badge>
  if (configuration.last_sync_status === "SUCCEEDED") return <Badge>Healthy</Badge>
  return <Badge>Enabled</Badge>
}

const syncIntervals: { value: SCIMSyncInterval; label: string }[] = [
  { value: "5m", label: "Every 5 minutes" },
  { value: "15m", label: "Every 15 minutes" },
  { value: "30m", label: "Every 30 minutes" },
  { value: "1h", label: "Every hour" },
  { value: "6h", label: "Every 6 hours" },
  { value: "24h", label: "Daily" },
]

function runStatusBadge(run: SCIMSyncRun) {
  if (run.status === "FAILED") return <Badge variant="destructive">Failed</Badge>
  if (run.status === "SUCCEEDED") return <Badge>Succeeded</Badge>
  if (run.status === "RUNNING") return <Badge variant="outline">Running</Badge>
  return <Badge variant="outline">Queued</Badge>
}

function RunHistoryItem({ run }: { run: SCIMSyncRun }) {
  const changes = run.result.users_created + run.result.users_updated + run.result.users_deactivated
    + run.result.groups_created + run.result.groups_updated + run.result.memberships_added + run.result.memberships_removed
  return (
    <div className="space-y-2 rounded-lg border border-border/60 p-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          {runStatusBadge(run)}
          <span className="text-sm font-medium">{run.trigger === "MANUAL" ? "Manual" : "Scheduled"} sync</span>
        </div>
        <span className="text-xs text-muted-foreground">{new Date(run.requested_at).toLocaleString()}</span>
      </div>
      {run.status === "RUNNING" && <p className="text-xs text-muted-foreground">Reconciliation is running in the background.</p>}
      {run.status === "QUEUED" && <p className="text-xs text-muted-foreground">Waiting for a worker.</p>}
      {run.status === "SUCCEEDED" && (
        <p className="text-xs text-muted-foreground">
          {run.result.desired_users} users and {run.result.desired_groups} groups reconciled · {changes} changes · {run.result.skipped_users.length} skipped
        </p>
      )}
      {run.error && <p className="text-sm text-destructive">{run.error}</p>}
      {run.result.skipped_users.length > 0 && (
        <p className="text-xs text-amber-400">
          Skipped: {run.result.skipped_users.map((user) => user.username || user.entity_id).join(", ")}
        </p>
      )}
    </div>
  )
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
    syncInterval: SCIMSyncInterval
  } | null>(null)
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [previewing, setPreviewing] = useState(false)
  const [syncing, setSyncing] = useState(false)
  const [removing, setRemoving] = useState(false)
  const [removeOpen, setRemoveOpen] = useState(false)
  const [result, setResult] = useState<SCIMSyncResult | null>(null)
  const [visibleHistoryCount, setVisibleHistoryCount] = useState(3)

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
    refetchInterval: 30000,
  })
  const groupsQuery = useQuery({
    queryKey: ["application", "id", id, "groups"],
    queryFn: async () => (await api.get<GroupWithLink[]>(`/applications/${id}/groups`)).data,
    enabled: !!id,
  })
  const runsQuery = useQuery({
    queryKey: ["application", "id", id, "scim", "syncs"],
    queryFn: async () => (await api.get<SCIMSyncRun[]>(`/saml/applications/${id}/scim/syncs`, { params: { limit: 20 } })).data,
    enabled: !!id && !!configQuery.data,
    refetchInterval: 3000,
  })

  const form = draft ?? {
    endpoint: configQuery.data?.endpoint ?? "",
    accessToken: "",
    tokenExpiresAt: configQuery.data?.token_expires_at?.slice(0, 10) ?? "",
    enabled: configQuery.data?.enabled ?? true,
    syncInterval: configQuery.data?.sync_interval ?? "1h",
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
        sync_interval: form.syncInterval,
      })).data
      queryClient.setQueryData(["application", "id", id, "scim"], stored)
      setDraft({
        endpoint: stored.endpoint,
        accessToken: "",
        tokenExpiresAt: stored.token_expires_at?.slice(0, 10) ?? "",
        enabled: stored.enabled,
        syncInterval: stored.sync_interval,
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
      const run = (await api.post<SCIMSyncRun>(`/saml/applications/${id}/scim/sync`)).data
      queryClient.setQueryData<SCIMSyncRun[]>(["application", "id", id, "scim", "syncs"], (current = []) => [run, ...current.filter((item) => item.id !== run.id)])
      await queryClient.invalidateQueries({ queryKey: ["application", "id", id, "scim"] })
      toast.success(run.status === "RUNNING" ? "AWS provisioning is running" : "AWS provisioning queued")
    } catch (error) {
      toast.error(apiError(error, "Couldn't queue AWS provisioning."))
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
  const activeRun = runsQuery.data?.find((run) => run.status === "QUEUED" || run.status === "RUNNING")
  const latestCompletedRun = runsQuery.data?.find((run) => run.status === "SUCCEEDED" || run.status === "FAILED")
  const historyRuns = runsQuery.data?.filter((run) => run.id !== activeRun?.id) ?? []
  const visibleHistoryRuns = historyRuns.slice(0, visibleHistoryCount)

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
              <div className="space-y-2">
                <Label>Sync interval</Label>
                <Select value={form.syncInterval} onValueChange={(value) => updateDraft({ syncInterval: value as SCIMSyncInterval })}>
                  <SelectTrigger className="w-56"><SelectValue /></SelectTrigger>
                  <SelectContent>{syncIntervals.map((interval) => <SelectItem key={interval.value} value={interval.value}>{interval.label}</SelectItem>)}</SelectContent>
                </Select>
              </div>
              <label className="flex cursor-pointer items-center gap-2 text-sm"><input type="checkbox" checked={form.enabled} onChange={(event) => updateDraft({ enabled: event.target.checked })} className="size-4 accent-gr-pink" />Enable scheduled and manual synchronization</label>
              {configuration?.enabled && configuration.next_sync_at && <p className="text-xs text-muted-foreground">Next scheduled sync {new Date(configuration.next_sync_at).toLocaleString()}</p>}
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
            <CardHeader><CardTitle>Synchronization</CardTitle><CardDescription>Reconciliation runs in the background and continues if this page closes. Users leaving all linked groups are deactivated; Sentinel never deletes AWS users or groups.</CardDescription></CardHeader>
            <CardContent className="space-y-4">
              {configuration?.last_sync_at && <p className="text-sm text-muted-foreground">Last attempt {new Date(configuration.last_sync_at).toLocaleString()} · {configuration.last_sync_status.toLowerCase()}</p>}
              {configuration?.last_sync_error && <p className="rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">{configuration.last_sync_error}</p>}
              <Button type="button" disabled={!configuration?.enabled || tokenExpired || syncing || !!activeRun} onClick={synchronize}><RefreshCw className={`mr-1 size-3.5 ${syncing || activeRun?.status === "RUNNING" ? "animate-spin" : ""}`} />{syncing ? "Queuing…" : activeRun?.status === "RUNNING" ? "Synchronizing…" : activeRun ? "Sync queued" : "Sync now"}</Button>
              {activeRun && <RunHistoryItem run={activeRun} />}
              {!activeRun && latestCompletedRun?.status === "SUCCEEDED" && <ResultSummary result={latestCompletedRun.result} />}
              <div className="space-y-2 border-t border-border/60 pt-4">
                <h3 className="text-sm font-medium">Sync history</h3>
                {runsQuery.isLoading ? <Skeleton className="h-20" /> : runsQuery.isError ? <p className="text-sm text-destructive">Could not load synchronization history.</p> : historyRuns.length === 0 ? <p className="text-sm text-muted-foreground">No completed synchronization runs yet.</p> : <>{visibleHistoryRuns.map((run) => <RunHistoryItem key={run.id} run={run} />)}{visibleHistoryCount < historyRuns.length && <Button type="button" variant="outline" size="sm" onClick={() => setVisibleHistoryCount((count) => count + 3)}>Load more</Button>}</>}
              </div>
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
