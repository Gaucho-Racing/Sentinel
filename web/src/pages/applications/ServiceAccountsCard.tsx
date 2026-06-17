import {
  Check,
  Copy,
  Eye,
  KeyRound,
  Plus,
  RefreshCw,
  Trash2,
} from "lucide-react"
import { useEffect, useState } from "react"
import { toast } from "sonner"

import { OutlineButton } from "@/components/OutlineButton"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
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
import { useAdmins } from "@/lib/admin"
import { loadSession } from "@/lib/auth"
import {
  isNeverExpires,
  SA_ALLOWED_SCOPES,
  SA_SCOPE_DESCRIPTIONS,
  TTL_PRESETS,
  useApplicationServiceAccounts,
  useCreateServiceAccount,
  useDeleteServiceAccount,
  useRotateServiceAccountToken,
  useViewServiceAccountToken,
  type SAScope,
  type ServiceAccount,
  type ServiceAccountWithToken,
} from "@/lib/service-accounts"

function formatDate(iso: string | null | undefined): string {
  if (!iso) return "—"
  return new Date(iso).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  })
}

function expirySummary(expiresAt: string | null | undefined): string {
  if (!expiresAt) return "No active token"
  if (isNeverExpires(expiresAt)) return "Never expires"
  const t = new Date(expiresAt).getTime()
  const ms = t - Date.now()
  if (ms <= 0) return `Expired ${formatDate(expiresAt)}`
  const days = Math.round(ms / (24 * 60 * 60 * 1000))
  if (days < 30) return `Expires in ${days} day${days === 1 ? "" : "s"}`
  return `Expires ${formatDate(expiresAt)}`
}

function extractError(e: unknown, fallback: string): string {
  const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error
  return msg ?? fallback
}

export function ServiceAccountsCard({ applicationID }: { applicationID: string }) {
  const sasQuery = useApplicationServiceAccounts(applicationID)
  const [createOpen, setCreateOpen] = useState(false)
  const [revealed, setRevealed] = useState<ServiceAccountWithToken | null>(null)
  // Viewer identity drives the "View token" affordance: only the SA's
  // creator (or any admin) can re-reveal the persisted JWT. Owners who
  // didn't create a given SA can still rotate to get a fresh token.
  const myEntityID = loadSession()?.entityId ?? ""
  const { isAdmin } = useAdmins()

  const sas = sasQuery.data ?? []

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <KeyRound className="size-4 text-muted-foreground" />
          Service accounts
        </CardTitle>
        <CardDescription>
          Non-human identities for this application. Each service account
          gets a single bearer JWT that authenticates to Sentinel; rotate
          it to swap the token without recreating the account.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {sasQuery.isLoading ? (
          <Skeleton className="h-16 w-full" />
        ) : sas.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No service accounts yet.
          </p>
        ) : (
          <ul className="space-y-3">
            {sas.map((sa) => (
              <ServiceAccountItem
                key={sa.id}
                sa={sa}
                applicationID={applicationID}
                canViewToken={isAdmin || sa.created_by === myEntityID}
                onRevealToken={(result) => setRevealed(result)}
              />
            ))}
          </ul>
        )}
        <div className="pt-1">
          <Button type="button" onClick={() => setCreateOpen(true)}>
            <Plus className="mr-1 size-3.5" />
            Add service account
          </Button>
        </div>
      </CardContent>

      <CreateServiceAccountDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        applicationID={applicationID}
        onCreated={(result) => {
          setCreateOpen(false)
          setRevealed(result)
        }}
      />

      <RevealTokenDialog
        result={revealed}
        onClose={() => setRevealed(null)}
      />
    </Card>
  )
}

function ServiceAccountItem({
  sa,
  applicationID,
  canViewToken,
  onRevealToken,
}: {
  sa: ServiceAccount
  applicationID: string
  canViewToken: boolean
  onRevealToken: (result: ServiceAccountWithToken) => void
}) {
  const viewToken = useViewServiceAccountToken()
  const [confirmRotate, setConfirmRotate] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)

  const scopes = sa.scope.trim() ? sa.scope.split(/\s+/) : []
  const tokenExp = sa.active_token?.expires_at
  const hasActiveToken = !!sa.active_token

  async function handleView() {
    try {
      const token = await viewToken.mutateAsync(sa.id)
      onRevealToken({ service_account: sa, token })
    } catch (e) {
      toast.error(extractError(e, "Couldn't load token."))
    }
  }

  return (
    <li className="rounded-md border border-border/60 bg-muted/40 p-3">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="truncate font-mono text-sm">{sa.name}</p>
          <p className="mt-0.5 text-xs text-muted-foreground">
            <code className="font-mono">{sa.id}</code>
            <span className="mx-1.5">·</span>
            Created {formatDate(sa.created_at)}
          </p>
        </div>
        <div className="flex shrink-0 items-center gap-1">
          {canViewToken && hasActiveToken && (
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={handleView}
              disabled={viewToken.isPending}
              title="View token"
            >
              <Eye className="size-3.5" />
            </Button>
          )}
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={() => setConfirmRotate(true)}
            title="Rotate token"
          >
            <RefreshCw className="size-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={() => setConfirmDelete(true)}
            title="Delete service account"
          >
            <Trash2 className="size-3.5" />
          </Button>
        </div>
      </div>

      <div className="mt-3 grid gap-2 text-xs sm:grid-cols-2">
        <div>
          <p className="uppercase tracking-wider text-muted-foreground">Scope</p>
          <div className="mt-1 flex flex-wrap items-center gap-1.5">
            {scopes.length === 0 ? (
              <span className="text-muted-foreground">No scope</span>
            ) : (
              scopes.map((s) => (
                <span
                  key={s}
                  className="inline-flex items-center rounded-md border border-border/60 bg-background/60 px-2 py-0.5 font-mono text-xs"
                >
                  {s}
                </span>
              ))
            )}
          </div>
        </div>
        <div>
          <p className="uppercase tracking-wider text-muted-foreground">Token</p>
          <p className="mt-1">{expirySummary(tokenExp)}</p>
        </div>
      </div>

      <RotateTokenDialog
        open={confirmRotate}
        onOpenChange={setConfirmRotate}
        sa={sa}
        applicationID={applicationID}
        onRotated={(result) => {
          setConfirmRotate(false)
          onRevealToken(result)
        }}
      />

      <DeleteServiceAccountDialog
        open={confirmDelete}
        onOpenChange={setConfirmDelete}
        sa={sa}
        applicationID={applicationID}
      />
    </li>
  )
}

function CreateServiceAccountDialog({
  open,
  onOpenChange,
  applicationID,
  onCreated,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  applicationID: string
  onCreated: (result: ServiceAccountWithToken) => void
}) {
  const createSA = useCreateServiceAccount(applicationID)
  const [name, setName] = useState("")
  const [scopeSet, setScopeSet] = useState<Set<SAScope>>(new Set())
  // 365-day default — picked over the PR #78 "90 days" since the user
  // chose 1 year as the default for SA tokens.
  const [ttlDays, setTtlDays] = useState("365")

  useEffect(() => {
    if (open) {
      setName("")
      setScopeSet(new Set())
      setTtlDays("365")
    }
  }, [open])

  function toggleScope(s: SAScope) {
    setScopeSet((prev) => {
      const next = new Set(prev)
      if (next.has(s)) next.delete(s)
      else next.add(s)
      return next
    })
  }

  async function handleCreate() {
    const trimmed = name.trim()
    if (!trimmed) return
    try {
      const result = await createSA.mutateAsync({
        name: trimmed,
        scope: [...scopeSet].join(" "),
        ttl_days: parseInt(ttlDays, 10),
      })
      onCreated(result)
    } catch (e) {
      toast.error(extractError(e, "Couldn't create service account."))
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        if (!createSA.isPending) onOpenChange(o)
      }}
    >
      <DialogContent className="gap-5 sm:max-w-md">
        <DialogHeader className="gap-3">
          <div className="flex size-10 items-center justify-center rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-white">
            <KeyRound className="size-5" />
          </div>
          <DialogTitle>New service account</DialogTitle>
          <DialogDescription>
            A service account gets one bearer JWT, shown <strong>once</strong>{" "}
            after creation. Pick a name that's clear about what this account
            does and the scopes it needs.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="sa-name">Name</Label>
            <Input
              id="sa-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. mapache-ingest"
              autoFocus
              disabled={createSA.isPending}
            />
          </div>
          <div className="space-y-2">
            <Label>Scope</Label>
            <ul className="space-y-1.5">
              {SA_ALLOWED_SCOPES.map((s) => {
                const checked = scopeSet.has(s)
                return (
                  <li key={s}>
                    <button
                      type="button"
                      onClick={() => toggleScope(s)}
                      disabled={createSA.isPending}
                      className="flex w-full items-center gap-2.5 rounded-md border border-border/60 bg-background/60 px-3 py-2 text-left transition-colors hover:bg-muted/40 disabled:opacity-50"
                    >
                      <span
                        className={
                          "flex size-4 shrink-0 items-center justify-center rounded border " +
                          (checked
                            ? "border-gr-pink bg-gr-pink text-white"
                            : "border-border bg-background")
                        }
                      >
                        {checked && <Check className="size-3" />}
                      </span>
                      <span className="flex flex-col">
                        <code className="font-mono text-sm">{s}</code>
                        <span className="text-xs text-muted-foreground">
                          {SA_SCOPE_DESCRIPTIONS[s]}
                        </span>
                      </span>
                    </button>
                  </li>
                )
              })}
            </ul>
          </div>
          <div className="space-y-2">
            <Label htmlFor="sa-ttl">Token expires after</Label>
            <Select
              value={ttlDays}
              onValueChange={setTtlDays}
              disabled={createSA.isPending}
            >
              <SelectTrigger id="sa-ttl">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {TTL_PRESETS.map((preset) => (
                  <SelectItem key={preset.days} value={String(preset.days)}>
                    {preset.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        <div className="flex justify-end gap-2 pt-1">
          <Button
            type="button"
            variant="ghost"
            disabled={createSA.isPending}
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <OutlineButton
            type="button"
            size="sm"
            className="w-auto"
            loading={createSA.isPending}
            disabled={!name.trim() || createSA.isPending}
            onClick={handleCreate}
          >
            Create service account
          </OutlineButton>
        </div>
      </DialogContent>
    </Dialog>
  )
}

function RotateTokenDialog({
  open,
  onOpenChange,
  sa,
  applicationID,
  onRotated,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  sa: ServiceAccount
  applicationID: string
  onRotated: (result: ServiceAccountWithToken) => void
}) {
  const rotate = useRotateServiceAccountToken(applicationID)

  async function handleRotate() {
    try {
      const result = await rotate.mutateAsync(sa.id)
      onRotated(result)
    } catch (e) {
      toast.error(extractError(e, "Couldn't rotate token."))
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        if (!rotate.isPending) onOpenChange(o)
      }}
    >
      <DialogContent className="gap-5 sm:max-w-md">
        <DialogHeader className="gap-3">
          <div className="flex size-10 items-center justify-center rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-white">
            <RefreshCw className="size-5" />
          </div>
          <DialogTitle>Rotate token for {sa.name}?</DialogTitle>
          <DialogDescription>
            Mints a fresh JWT with the same scope and TTL. Any client still
            using the previous token will start getting 401s immediately —
            redeploy them with the new token after you copy it.
          </DialogDescription>
        </DialogHeader>

        <div className="flex justify-end gap-2 pt-1">
          <Button
            type="button"
            variant="ghost"
            disabled={rotate.isPending}
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <OutlineButton
            type="button"
            size="sm"
            className="w-auto"
            loading={rotate.isPending}
            disabled={rotate.isPending}
            onClick={handleRotate}
          >
            Rotate token
          </OutlineButton>
        </div>
      </DialogContent>
    </Dialog>
  )
}

function RevealTokenDialog({
  result,
  onClose,
}: {
  result: ServiceAccountWithToken | null
  onClose: () => void
}) {
  function copyToken() {
    if (!result) return
    void navigator.clipboard.writeText(result.token)
    toast.success("Token copied")
  }

  return (
    <Dialog
      open={result !== null}
      onOpenChange={(o) => {
        if (!o) onClose()
      }}
    >
      <DialogContent className="gap-5 sm:max-w-md">
        <DialogHeader className="gap-3">
          <div className="flex size-10 items-center justify-center rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-white">
            <KeyRound className="size-5" />
          </div>
          <DialogTitle>Copy your bearer token</DialogTitle>
          <DialogDescription>
            This is the only time the full token for{" "}
            <strong>{result?.service_account.name}</strong> will be shown.
            Save it somewhere safe — if you lose it you'll have to rotate.
          </DialogDescription>
        </DialogHeader>

        <div className="flex items-center gap-1 rounded-md border border-border/60 bg-muted/40 px-2.5 py-1.5">
          <code className="flex-1 break-all font-mono text-xs">{result?.token}</code>
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={copyToken}
            title="Copy token"
          >
            <Copy className="size-3.5" />
          </Button>
        </div>

        <div className="flex justify-end gap-2 pt-1">
          <OutlineButton
            type="button"
            size="sm"
            className="w-auto"
            onClick={onClose}
          >
            Done
          </OutlineButton>
        </div>
      </DialogContent>
    </Dialog>
  )
}

function DeleteServiceAccountDialog({
  open,
  onOpenChange,
  sa,
  applicationID,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  sa: ServiceAccount
  applicationID: string
}) {
  const deleteSA = useDeleteServiceAccount(applicationID)

  async function handleDelete() {
    try {
      await deleteSA.mutateAsync(sa.id)
      toast.success(`Deleted ${sa.name}`)
      onOpenChange(false)
    } catch (e) {
      toast.error(extractError(e, "Couldn't delete service account."))
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        if (!deleteSA.isPending) onOpenChange(o)
      }}
    >
      <DialogContent className="gap-5 sm:max-w-md">
        <DialogHeader className="gap-3">
          <div className="flex size-10 items-center justify-center rounded-xl bg-destructive/10 text-destructive">
            <Trash2 className="size-5" />
          </div>
          <DialogTitle>Delete {sa.name}?</DialogTitle>
          <DialogDescription>
            The service account and its active token are removed
            immediately. Any client still using the token will start
            getting 401s. This can't be undone.
          </DialogDescription>
        </DialogHeader>

        <div className="flex justify-end gap-2 pt-1">
          <Button
            type="button"
            variant="ghost"
            disabled={deleteSA.isPending}
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="destructive"
            disabled={deleteSA.isPending}
            onClick={handleDelete}
          >
            {deleteSA.isPending ? "Deleting…" : "Delete service account"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
