import {
  Copy,
  KeyRound,
  Plus,
  Trash2,
  X,
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
  return `sk_${keyID.slice(0, 8)}…`
}

function extractError(e: unknown, fallback: string): string {
  const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error
  return msg ?? fallback
}

export function ServiceAccountsCard({ applicationID }: { applicationID: string }) {
  const sasQuery = useApplicationServiceAccounts(applicationID)
  const [createOpen, setCreateOpen] = useState(false)

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
          can have one or more API keys that authenticate to Sentinel as
          the service account.
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
      />
    </Card>
  )
}

function ServiceAccountItem({
  sa,
  applicationID,
}: {
  sa: ServiceAccount
  applicationID: string
}) {
  const keysQuery = useServiceAccountAPIKeys(sa.id)
  const [createKeyOpen, setCreateKeyOpen] = useState(false)
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false)
  const [revealed, setRevealed] = useState<APIKeyWithToken | null>(null)
  const [revokeTarget, setRevokeTarget] = useState<APIKey | null>(null)

  const keys = keysQuery.data ?? []

  return (
    <li className="rounded-md border border-border/60 bg-muted/40 p-3">
      <div className="flex items-center justify-between gap-3">
        <div className="min-w-0">
          <p className="truncate font-mono text-sm">{sa.name}</p>
          <p className="mt-0.5 text-xs text-muted-foreground">
            Created {formatTime(sa.created_at)}
            <span className="mx-1.5">·</span>
            <code className="font-mono">{sa.id}</code>
          </p>
        </div>
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={() => setConfirmDeleteOpen(true)}
          title="Delete service account"
        >
          <Trash2 className="size-3.5" />
        </Button>
      </div>

      <div className="mt-3 space-y-2">
        <p className="text-xs uppercase tracking-wider text-muted-foreground">
          API keys
        </p>
        {keysQuery.isLoading ? (
          <Skeleton className="h-10 w-full" />
        ) : keys.length === 0 ? (
          <p className="text-sm text-muted-foreground">No API keys yet.</p>
        ) : (
          <ul className="space-y-2">
            {keys.map((k) => (
              <APIKeyRow
                key={k.id}
                k={k}
                onRevoke={() => setRevokeTarget(k)}
              />
            ))}
          </ul>
        )}
        <div className="pt-1">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => setCreateKeyOpen(true)}
          >
            <Plus className="mr-1 size-3.5" />
            New API key
          </Button>
        </div>
      </div>

      <CreateAPIKeyDialog
        open={createKeyOpen}
        onOpenChange={setCreateKeyOpen}
        saID={sa.id}
        onCreated={(result) => {
          setCreateKeyOpen(false)
          setRevealed(result)
        }}
      />

      <RevealAPIKeyDialog
        token={revealed?.token ?? null}
        keyName={revealed?.key.name ?? ""}
        onClose={() => setRevealed(null)}
      />

      <RevokeAPIKeyDialog
        target={revokeTarget}
        saID={sa.id}
        onClose={() => setRevokeTarget(null)}
      />

      <DeleteServiceAccountDialog
        open={confirmDeleteOpen}
        onOpenChange={setConfirmDeleteOpen}
        sa={sa}
        applicationID={applicationID}
        keyCount={keys.length}
      />
    </li>
  )
}

function APIKeyRow({ k, onRevoke }: { k: APIKey; onRevoke: () => void }) {
  return (
    <li className="flex items-center justify-between gap-3 rounded-md border border-border/60 bg-background/60 px-3 py-2">
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-sm font-medium">{k.name}</span>
          <code className="font-mono text-xs text-muted-foreground">
            {maskedKeyID(k.key_id)}
          </code>
        </div>
        <div className="flex flex-wrap items-center gap-x-3 gap-y-0.5 text-xs text-muted-foreground">
          <span>Last used {formatTime(k.last_used_at)}</span>
          <span>
            {k.expires_at ? `Expires ${formatTime(k.expires_at)}` : "Never expires"}
          </span>
          {k.scope && (
            <span>
              Scope: <code className="font-mono">{k.scope}</code>
            </span>
          )}
        </div>
      </div>
      <Button
        variant="ghost"
        size="icon-sm"
        onClick={onRevoke}
        title="Revoke key"
      >
        <X className="size-3.5" />
      </Button>
    </li>
  )
}

function CreateServiceAccountDialog({
  open,
  onOpenChange,
  applicationID,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  applicationID: string
}) {
  const createSA = useCreateServiceAccount(applicationID)
  const [name, setName] = useState("")

  useEffect(() => {
    if (open) setName("")
  }, [open])

  async function handleCreate() {
    const trimmed = name.trim()
    if (!trimmed) return
    try {
      await createSA.mutateAsync(trimmed)
      onOpenChange(false)
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
            Pick a name that's clear about what this account does. Service
            accounts get API keys that authenticate to Sentinel as a
            non-human identity.
          </DialogDescription>
        </DialogHeader>

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
  // 90 days is the agreed default. The dropdown also offers 30 / 365 / never.
  const [ttlDays, setTtlDays] = useState("90")
  const [scope, setScope] = useState("")

  useEffect(() => {
    if (open) {
      setName("")
      setTtlDays("90")
      setScope("")
    }
  }, [open])

  async function handleCreate() {
    const trimmed = name.trim()
    if (!trimmed) return
    try {
      const result = await createKey.mutateAsync({
        name: trimmed,
        ttl_days: parseInt(ttlDays, 10),
        scope: scope.trim(),
      })
      onCreated(result)
    } catch (e) {
      toast.error(extractError(e, "Couldn't create API key."))
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        if (!createKey.isPending) onOpenChange(o)
      }}
    >
      <DialogContent className="gap-5 sm:max-w-md">
        <DialogHeader className="gap-3">
          <div className="flex size-10 items-center justify-center rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-white">
            <KeyRound className="size-5" />
          </div>
          <DialogTitle>New API key</DialogTitle>
          <DialogDescription>
            The full key is shown <strong>once</strong> after creation and
            never again — copy it then.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="key-name">Name</Label>
            <Input
              id="key-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. production-ingest"
              autoFocus
              disabled={createKey.isPending}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="key-ttl">Expires after</Label>
            <Select
              value={ttlDays}
              onValueChange={setTtlDays}
              disabled={createKey.isPending}
            >
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
              placeholder="space-separated, leave blank for none"
              disabled={createKey.isPending}
            />
          </div>
        </div>

        <div className="flex justify-end gap-2 pt-1">
          <Button
            type="button"
            variant="ghost"
            disabled={createKey.isPending}
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <OutlineButton
            type="button"
            size="sm"
            className="w-auto"
            loading={createKey.isPending}
            disabled={!name.trim() || createKey.isPending}
            onClick={handleCreate}
          >
            Create key
          </OutlineButton>
        </div>
      </DialogContent>
    </Dialog>
  )
}

function RevealAPIKeyDialog({
  token,
  keyName,
  onClose,
}: {
  token: string | null
  keyName: string
  onClose: () => void
}) {
  function copyToken() {
    if (!token) return
    void navigator.clipboard.writeText(token)
    toast.success("API key copied")
  }

  return (
    <Dialog
      open={token !== null}
      onOpenChange={(o) => {
        if (!o) onClose()
      }}
    >
      <DialogContent className="gap-5 sm:max-w-md">
        <DialogHeader className="gap-3">
          <div className="flex size-10 items-center justify-center rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-white">
            <KeyRound className="size-5" />
          </div>
          <DialogTitle>Copy your API key</DialogTitle>
          <DialogDescription>
            This is the only time the full key for <strong>{keyName}</strong>{" "}
            will be shown. Save it somewhere safe — if you lose it you'll
            have to revoke and mint a new one.
          </DialogDescription>
        </DialogHeader>

        <div className="flex items-center gap-1 rounded-md border border-border/60 bg-muted/40 px-2.5 py-1.5">
          <code className="flex-1 break-all font-mono text-xs">{token}</code>
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={copyToken}
            title="Copy key"
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

function RevokeAPIKeyDialog({
  target,
  saID,
  onClose,
}: {
  target: APIKey | null
  saID: string
  onClose: () => void
}) {
  const revoke = useRevokeAPIKey(saID)

  async function handleRevoke() {
    if (!target) return
    try {
      await revoke.mutateAsync(target.id)
      toast.success(`Revoked ${target.name}`)
      onClose()
    } catch (e) {
      toast.error(extractError(e, "Couldn't revoke key."))
    }
  }

  return (
    <Dialog
      open={target !== null}
      onOpenChange={(o) => {
        if (!o && !revoke.isPending) onClose()
      }}
    >
      <DialogContent className="gap-5 sm:max-w-md">
        <DialogHeader className="gap-3">
          <div className="flex size-10 items-center justify-center rounded-xl bg-destructive/10 text-destructive">
            <Trash2 className="size-5" />
          </div>
          <DialogTitle>Revoke {target?.name}?</DialogTitle>
          <DialogDescription>
            Any client still using this key will start getting 401s
            immediately. This can't be undone — you'd need to mint a fresh
            key and redeploy.
          </DialogDescription>
        </DialogHeader>

        <div className="flex justify-end gap-2 pt-1">
          <Button
            type="button"
            variant="ghost"
            disabled={revoke.isPending}
            onClick={onClose}
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="destructive"
            disabled={revoke.isPending}
            onClick={handleRevoke}
          >
            {revoke.isPending ? "Revoking…" : "Revoke key"}
          </Button>
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
  keyCount,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  sa: ServiceAccount
  applicationID: string
  keyCount: number
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
            {keyCount === 0
              ? "This service account has no keys. Deleting it removes the account row and frees the name."
              : `Every API key on this service account will be revoked (${keyCount} key${keyCount === 1 ? "" : "s"} total). Any clients still using them will start getting 401s.`}{" "}
            This can't be undone.
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
