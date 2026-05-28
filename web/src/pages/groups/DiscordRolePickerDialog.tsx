import { Bot, Check, Search } from "lucide-react"
import { useEffect, useMemo, useState } from "react"

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
import { discordRoleColorHex, useDiscordRoles } from "@/lib/discord"
import { fuzzyFilter } from "@/lib/fuzzy"

export function DiscordRolePickerDialog({
  open,
  onOpenChange,
  onAddBinding,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  onAddBinding: (roleIDs: string[]) => void
}) {
  const rolesQuery = useDiscordRoles()
  const [search, setSearch] = useState("")
  const [selected, setSelected] = useState<Set<string>>(new Set())

  // Reset transient state every time the dialog reopens.
  useEffect(() => {
    if (open) {
      setSearch("")
      setSelected(new Set())
    }
  }, [open])

  const eligibleRoles = useMemo(() => {
    if (!rolesQuery.data) return []
    const filtered = rolesQuery.data
      .filter((r) => r.position > 0 && !r.managed)
      .sort((a, b) => b.position - a.position)
    return fuzzyFilter(filtered, search, (r) => [r.name])
  }, [rolesQuery.data, search])

  function toggle(roleID: string) {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(roleID)) next.delete(roleID)
      else next.add(roleID)
      return next
    })
  }

  function commit() {
    if (selected.size === 0) return
    onAddBinding([...selected])
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="gap-5 sm:max-w-md">
        <DialogHeader className="gap-3">
          <div className="flex size-10 items-center justify-center rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-white">
            <Bot className="size-5" />
          </div>
          <DialogTitle>Add Discord role binding</DialogTitle>
          <DialogDescription>
            Pick one or more roles. Users must have <strong>all</strong> selected roles to be
            synced through this binding. To express "either-or", add a second binding.
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
                const checked = selected.has(role.id)
                const hex = discordRoleColorHex(role.color)
                return (
                  <li key={role.id}>
                    <button
                      type="button"
                      onClick={() => toggle(role.id)}
                      className="flex w-full items-center justify-between gap-3 px-3 py-2 text-left transition-colors hover:bg-muted/40"
                    >
                      <span className="flex min-w-0 items-center gap-2.5">
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
                        <span
                          className="size-2.5 shrink-0 rounded-full border border-border/60"
                          style={{ backgroundColor: hex ?? "transparent" }}
                          title={hex ?? "no color"}
                        />
                        <span className="truncate text-sm">{role.name}</span>
                      </span>
                    </button>
                  </li>
                )
              })}
            </ul>
          )}
        </div>

        <div className="flex items-center justify-between gap-2 pt-1">
          <p className="text-xs text-muted-foreground">
            {selected.size === 0
              ? "Select at least one role."
              : selected.size === 1
                ? "1 role selected."
                : `${selected.size} roles selected — all required.`}
          </p>
          <div className="flex gap-2">
            <Button
              type="button"
              variant="ghost"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button
              type="button"
              disabled={selected.size === 0}
              onClick={commit}
            >
              Add binding
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
