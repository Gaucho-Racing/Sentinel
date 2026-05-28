import { Bot, Search } from "lucide-react"
import { useMemo, useState } from "react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import {
  discordRoleColorHex,
  useAddGroupDiscordBinding,
  useDiscordRoles,
} from "@/lib/discord"

export function DiscordRolePickerDialog({
  open,
  onOpenChange,
  groupID,
  alreadyBoundRoleIDs,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  groupID: string
  alreadyBoundRoleIDs: Set<string>
}) {
  const rolesQuery = useDiscordRoles()
  const addBinding = useAddGroupDiscordBinding(groupID)
  const [search, setSearch] = useState("")

  const eligibleRoles = useMemo(() => {
    if (!rolesQuery.data) return []
    // Drop @everyone (position 0) and managed roles (integration-owned —
    // assigning them via our bot would silently no-op). Sort highest
    // position first to match Discord's own listing.
    const filtered = rolesQuery.data
      .filter((r) => r.position > 0 && !r.managed)
      .sort((a, b) => b.position - a.position)
    const needle = search.trim().toLowerCase()
    if (!needle) return filtered
    return filtered.filter((r) => r.name.toLowerCase().includes(needle))
  }, [rolesQuery.data, search])

  async function pick(roleID: string) {
    if (addBinding.isPending) return
    try {
      await addBinding.mutateAsync(roleID)
      toast.success("Discord role added")
      onOpenChange(false)
      setSearch("")
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't add role."
      toast.error(message)
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        if (!addBinding.isPending) {
          onOpenChange(o)
          if (!o) setSearch("")
        }
      }}
    >
      <DialogContent className="gap-5 sm:max-w-md">
        <DialogHeader className="gap-3">
          <div className="flex size-10 items-center justify-center rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-white">
            <Bot className="size-5" />
          </div>
          <DialogTitle>Add Discord role</DialogTitle>
          <DialogDescription>
            Members of the selected role will be synced into this group automatically.
            @everyone and integration-managed roles are hidden.
          </DialogDescription>
        </DialogHeader>

        <div className="relative">
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            type="search"
            placeholder="Search roles…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
            autoFocus
          />
        </div>

        <div className="max-h-72 overflow-y-auto rounded-md border border-border/60">
          {rolesQuery.isLoading ? (
            <div className="space-y-2 p-2">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-8 w-full" />
              ))}
            </div>
          ) : rolesQuery.isError ? (
            <p className="p-4 text-center text-sm text-destructive">
              Couldn't load Discord roles.
            </p>
          ) : eligibleRoles.length === 0 ? (
            <p className="p-4 text-center text-sm text-muted-foreground">
              {search.trim() ? `No roles match "${search}".` : "No roles available."}
            </p>
          ) : (
            <ul className="divide-y divide-border/60">
              {eligibleRoles.map((role) => {
                const bound = alreadyBoundRoleIDs.has(role.id)
                const hex = discordRoleColorHex(role.color)
                return (
                  <li key={role.id}>
                    <button
                      type="button"
                      onClick={() => pick(role.id)}
                      disabled={bound || addBinding.isPending}
                      className="flex w-full items-center justify-between gap-3 px-3 py-2 text-left transition-colors hover:bg-muted/40 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      <span className="flex min-w-0 items-center gap-2.5">
                        <span
                          className="size-2.5 shrink-0 rounded-full border border-border/60"
                          style={{ backgroundColor: hex ?? "transparent" }}
                          title={hex ?? "no color"}
                        />
                        <span className="truncate text-sm">{role.name}</span>
                      </span>
                      {bound && (
                        <span className="text-xs text-muted-foreground">already bound</span>
                      )}
                    </button>
                  </li>
                )
              })}
            </ul>
          )}
        </div>

        <div className="flex justify-end gap-2 pt-1">
          <Button
            type="button"
            variant="ghost"
            disabled={addBinding.isPending}
            onClick={() => onOpenChange(false)}
          >
            Close
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
