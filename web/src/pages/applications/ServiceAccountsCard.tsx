import {
  ChevronDown,
  ChevronRight,
  Copy,
  Key,
  Loader2,
  Plus,
  Trash2,
} from "lucide-react"
import { useState } from "react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
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
import {
  TTL_PRESETS,
  useApplicationServiceAccounts,
  useCreateAPIKey,
  useCreateServiceAccount,
  useDeleteServiceAccount,
  useRevokeAPIKey,
  useServiceAccountAPIKeys,
  type APIKey,
  type APIKeyWithToken,
  type ServiceAccount,
} from "@/lib/service-accounts"

function formatTime(iso: string | null | undefined): string {
  if (!iso) return "—"
  return new Date(iso).toLocaleString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  })
}

function maskedKeyID(keyID: string): string {
  // First 8 chars + ellipsis. Enough to disambiguate keys without leaking
  // anything sensitive (the secret is the other half of the token).
  return `sk_${keyID.slice(0, 8)}…`
}

export function ServiceAccountsCard({ applicationID }: { applicationID: string }) {
  const sasQuery = useApplicationServiceAccounts(applicationID)
  const createSA = useCreateServiceAccount(applicationID)
  const [createOpen, setCreateOpen] = useState(false)
  const [createName, setCreateName] = useState("")
  const [expandedID, setExpandedID] = useState<string | null>(null)

  async function handleCreate() {
    const name = createName.trim()
    if (!name) return
    try {
      const sa = await createSA.mutateAsync(name)
      setCreateOpen(false)
      setCreateName("")
      setExpandedID(sa.id)
    } catch (e) {
      toast.error(extractError(e, "Couldn't create service account."))
    }
  }

  const sas = sasQuery.data ?? []

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>Service accounts</CardTitle>
        <Button size="sm" onClick={() => setCreateOpen(true)}>
          <Plus className="mr-1 size-3.5" />
          New service account
        </Button>
      </CardHeader>
      <CardContent>
        {sasQuery.isLoading ? (
          <div className="flex justify-center py-4">
            <Loader2 className="size-4 animate-spin text-muted-foreground" />
          </div>
        ) : sas.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No service accounts yet. Create one to mint API keys.
          </p>
        ) : (
          <ul className="space-y-2">
            {sas.map((sa) => (
              <ServiceAccountRow
                key={sa.id}
                sa={sa}
                applicationID={applicationID}
                expanded={expandedID === sa.id}
                onToggle={() => setExpandedID(expandedID === sa.id ? null : sa.id)}
              />
            ))}
          </ul>
        )}
      </CardContent>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New service account</DialogTitle>
            <DialogDescription>
              Service accounts get API keys that can authenticate to Sentinel
              as a non-human identity. Pick a name that's clear about what
              the account does.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label htmlFor="sa-name">Name</Label>
            <Input
              id="sa-name"
              value={createName}
              onChange={(e) => setCreateName(e.target.value)}
              placeholder="e.g. mapache-ingest"
              autoFocus
            />
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setCreateOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={!createName.trim() || createSA.isPending}
            >
              {createSA.isPending ? "Creating…" : "Create"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  )
}

function ServiceAccountRow({
  sa,
  applicationID,
  expanded,
  onToggle,
}: {
  sa: ServiceAccount
  applicationID: string
  expanded: boolean
  onToggle: () => void
}) {
  const deleteSA = useDeleteServiceAccount(applicationID)
  const [confirmDelete, setConfirmDelete] = useState(false)

  async function handleDelete() {
    try {
      await deleteSA.mutateAsync(sa.id)
      toast.success(`Deleted ${sa.name}.`)
      setConfirmDelete(false)
    } catch (e) {
      toast.error(extractError(e, "Couldn't delete service account."))
    }
  }

  return (
    <li className="rounded-md border border-border/60 bg-muted/30">
      <div className="flex items-center justify-between gap-3 px-3 py-2">
        <button
          type="button"
          onClick={onToggle}
          className="flex flex-1 items-center gap-2 text-left"
        >
          {expanded ? (
            <ChevronDown className="size-4 shrink-0 text-muted-foreground" />
          ) : (
            <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
          )}
          <span className="font-mono text-sm">{sa.name}</span>
          <span className="text-xs text-muted-foreground">
            {formatTime(sa.created_at)}
          </span>
        </button>
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={(e) => {
            e.stopPropagation()
            setConfirmDelete(true)
          }}
          title="Delete service account"
        >
          <Trash2 className="size-3.5" />
        </Button>
      </div>
      {expanded && (
        <div className="border-t border-border/60 px-3 py-3">
          <APIKeyList saID={sa.id} />
        </div>
      )}

      <Dialog open={confirmDelete} onOpenChange={setConfirmDelete}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete {sa.name}?</DialogTitle>
            <DialogDescription>
              This revokes every API key on this service account immediately.
              Any clients still using those keys will start getting 401s.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setConfirmDelete(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={deleteSA.isPending}
            >
              {deleteSA.isPending ? "Deleting…" : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </li>
  )
}

function APIKeyList({ saID }: { saID: string }) {
  const keysQuery = useServiceAccountAPIKeys(saID)
  const revokeKey = useRevokeAPIKey(saID)
  const [createOpen, setCreateOpen] = useState(false)
  const [revealed, setRevealed] = useState<APIKeyWithToken | null>(null)

  async function handleRevoke(keyID: string, name: string) {
    if (!confirm(`Revoke API key "${name}"? Clients using it will stop working.`)) return
    try {
      await revokeKey.mutateAsync(keyID)
      toast.success(`Revoked ${name}.`)
    } catch (e) {
      toast.error(extractError(e, "Couldn't revoke key."))
    }
  }

  const keys = keysQuery.data ?? []

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          {keys.length === 0 ? "No API keys yet." : `${keys.length} key${keys.length === 1 ? "" : "s"}`}
        </p>
        <Button size="sm" variant="outline" onClick={() => setCreateOpen(true)}>
          <Key className="mr-1 size-3.5" />
          New API key
        </Button>
      </div>
      {keys.length > 0 && (
        <ul className="space-y-1.5">
          {keys.map((k) => (
            <APIKeyRow key={k.id} k={k} onRevoke={() => handleRevoke(k.id, k.name)} />
          ))}
        </ul>
      )}

      <CreateAPIKeyDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        saID={saID}
        onCreated={(result) => {
          setCreateOpen(false)
          setRevealed(result)
        }}
      />

      <RevealAPIKeyDialog
        token={revealed?.token ?? null}
        onClose={() => setRevealed(null)}
      />
    </div>
  )
}

function APIKeyRow({ k, onRevoke }: { k: APIKey; onRevoke: () => void }) {
  return (
    <li className="flex items-center justify-between gap-3 rounded-md border border-border/60 bg-background/60 px-3 py-2 text-sm">
      <div className="flex min-w-0 flex-1 flex-col">
        <div className="flex items-center gap-2">
          <span className="font-medium">{k.name}</span>
          <code className="font-mono text-xs text-muted-foreground">
            {maskedKeyID(k.key_id)}
          </code>
        </div>
        <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
          <span>Last used {formatTime(k.last_used_at)}</span>
          <span>
            {k.expires_at ? `Expires ${formatTime(k.expires_at)}` : "Never expires"}
          </span>
          {k.scope && <span className="font-mono">scope: {k.scope}</span>}
        </div>
      </div>
      <Button variant="ghost" size="icon-sm" onClick={onRevoke} title="Revoke key">
        <Trash2 className="size-3.5" />
      </Button>
    </li>
  )
}

function CreateAPIKeyDialog({
  open,
  onOpenChange,
  saID,
  onCreated,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  saID: string
  onCreated: (result: APIKeyWithToken) => void
}) {
  const createKey = useCreateAPIKey(saID)
  const [name, setName] = useState("")
  // Default to 90 days as agreed; the user can pick 30/365/never from the
  // dropdown if they want different.
  const [ttlDays, setTtlDays] = useState("90")
  const [scope, setScope] = useState("")

  async function handleCreate() {
    const trimmed = name.trim()
    if (!trimmed) return
    try {
      const result = await createKey.mutateAsync({
        name: trimmed,
        ttl_days: parseInt(ttlDays, 10),
        scope: scope.trim(),
      })
      // Reset for next open before passing off to the reveal dialog.
      setName("")
      setScope("")
      setTtlDays("90")
      onCreated(result)
    } catch (e) {
      toast.error(extractError(e, "Couldn't create API key."))
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>New API key</DialogTitle>
          <DialogDescription>
            The full key is shown once after creation and never again — copy
            it then.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3">
          <div className="space-y-2">
            <Label htmlFor="key-name">Name</Label>
            <Input
              id="key-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. production-ingest"
              autoFocus
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="key-ttl">Expires after</Label>
            <Select value={ttlDays} onValueChange={setTtlDays}>
              <SelectTrigger id="key-ttl">
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
          <div className="space-y-2">
            <Label htmlFor="key-scope">Scope</Label>
            <Input
              id="key-scope"
              value={scope}
              onChange={(e) => setScope(e.target.value)}
              placeholder="e.g. read write (space-separated)"
            />
            <p className="text-xs text-muted-foreground">
              Space-separated. Leave blank for no scope.
            </p>
          </div>
        </div>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            onClick={handleCreate}
            disabled={!name.trim() || createKey.isPending}
          >
            {createKey.isPending ? "Creating…" : "Create"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function RevealAPIKeyDialog({
  token,
  onClose,
}: {
  token: string | null
  onClose: () => void
}) {
  function copyToken() {
    if (!token) return
    void navigator.clipboard.writeText(token)
    toast.success("Copied to clipboard.")
  }

  return (
    <Dialog open={token !== null} onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Copy your API key now</DialogTitle>
          <DialogDescription>
            This is the only time you'll see the full key. Save it somewhere
            safe — if you lose it you'll have to revoke and mint a new one.
          </DialogDescription>
        </DialogHeader>
        <div className="rounded-md border border-border/60 bg-muted/40 p-3">
          <code className="block break-all font-mono text-xs">{token}</code>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={copyToken}>
            <Copy className="mr-1 size-3.5" />
            Copy
          </Button>
          <Button onClick={onClose}>Done</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// extractError pulls the {error: "..."} message from an axios error, with
// a sensible fallback so toasts always say something useful.
function extractError(e: unknown, fallback: string): string {
  const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error
  return msg ?? fallback
}
