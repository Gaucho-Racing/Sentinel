import { useQuery } from "@tanstack/react-query"
import { Check, Search, Sparkles } from "lucide-react"
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
import { api } from "@/lib/api"
import { fuzzyFilter } from "@/lib/fuzzy"
import type { Group } from "@/lib/groups"

// GroupPickerDialog selects one-or-more existing groups to compose into a
// conditional binding's AND-group. Mirrors DiscordRolePickerDialog but
// keyed on Sentinel groups instead of Discord roles. `excludeGroupID` is
// the parent group of the binding — excluded from the picker since a
// binding can't require its own group (the backend would reject anyway).
export function GroupPickerDialog({
  open,
  onOpenChange,
  excludeGroupID,
  onAddBinding,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  excludeGroupID: string
  onAddBinding: (groupIDs: string[]) => void
}) {
  const groupsQuery = useQuery({
    queryKey: ["groups"],
    queryFn: async () => {
      const res = await api.get<Group[]>("/groups")
      return res.data
    },
    enabled: open,
  })
  const [search, setSearch] = useState("")
  const [selected, setSelected] = useState<Set<string>>(new Set())

  // Reset transient state every time the dialog reopens.
  useEffect(() => {
    if (open) {
      setSearch("")
      setSelected(new Set())
    }
  }, [open])

  const eligibleGroups = useMemo(() => {
    if (!groupsQuery.data) return []
    const filtered = groupsQuery.data
      .filter((g) => g.id !== excludeGroupID)
      .sort((a, b) => a.name.localeCompare(b.name))
    return fuzzyFilter(filtered, search, (g) => [g.name, g.description])
  }, [groupsQuery.data, excludeGroupID, search])

  function toggle(groupID: string) {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(groupID)) next.delete(groupID)
      else next.add(groupID)
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
            <Sparkles className="size-5" />
          </div>
          <DialogTitle>Add conditional binding</DialogTitle>
          <DialogDescription>
            Pick one or more groups. Entities must be a member of <strong>all</strong> selected
            groups to be synced through this binding. To express "either-or", add a second binding.
          </DialogDescription>
        </DialogHeader>

        <div className="relative">
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            type="search"
            placeholder="Search groups…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
            autoFocus
          />
        </div>

        <div className="max-h-72 overflow-y-auto rounded-md border border-border/60">
          {groupsQuery.isLoading ? (
            <div className="space-y-2 p-2">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-8 w-full" />
              ))}
            </div>
          ) : groupsQuery.isError ? (
            <p className="p-4 text-center text-sm text-destructive">
              Couldn't load groups.
            </p>
          ) : eligibleGroups.length === 0 ? (
            <p className="p-4 text-center text-sm text-muted-foreground">
              {search.trim() ? `No groups match "${search}".` : "No other groups available."}
            </p>
          ) : (
            <ul className="divide-y divide-border/60">
              {eligibleGroups.map((group) => {
                const checked = selected.has(group.id)
                return (
                  <li key={group.id}>
                    <button
                      type="button"
                      onClick={() => toggle(group.id)}
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
                        <span className="truncate text-sm">{group.name}</span>
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {group.member_count} member{group.member_count === 1 ? "" : "s"}
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
              ? "Select at least one group."
              : selected.size === 1
                ? "1 group selected."
                : `${selected.size} groups selected — all required.`}
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
