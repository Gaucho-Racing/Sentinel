import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Crown, Search, UserPlus } from "lucide-react"
import { useMemo, useState } from "react"
import { toast } from "sonner"

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
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
import { api } from "@/lib/api"
import {
  addCustom,
  addPreset,
  DURATION_PRESETS,
  formatAbsoluteDate,
  MAX_BY_UNIT,
  type DurationPreset,
  type DurationUnit,
} from "@/lib/duration"
import { fuzzyFilter } from "@/lib/fuzzy"

type AddGroupPersonMode = "member" | "owner"

const ENTITY_ID_PATTERN = /^ent_[0-9a-hjkmnp-tv-z]{26}$/i

type UserOption = {
  id: string
  entity_id: string
  username: string
  first_name: string
  last_name: string
  email: string
  avatar_url: string
}

function userName(user: UserOption) {
  const fullName = `${user.first_name} ${user.last_name}`.trim()
  return fullName || user.username || user.entity_id
}

function userInitials(user: UserOption) {
  return userName(user)
    .split(/\s+/)
    .map((part) => part[0])
    .filter(Boolean)
    .slice(0, 2)
    .join("")
    .toUpperCase()
}

function selectableText(user: UserOption) {
  return [
    user.first_name,
    user.last_name,
    user.username,
    user.email,
    user.entity_id,
  ].filter(Boolean)
}

function durationEnd(
  mode: DurationPreset | "custom",
  customAmount: number,
  customUnit: DurationUnit,
) {
  if (mode === "custom") {
    return addCustom(new Date(), customAmount, customUnit)
  }
  return addPreset(new Date(), mode)
}

export function AddGroupPersonDialog({
  open,
  onOpenChange,
  groupID,
  groupName,
  initialMode,
  existingMemberEntityIDs,
  existingOwnerEntityIDs,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  groupID: string
  groupName: string
  initialMode: AddGroupPersonMode
  existingMemberEntityIDs: string[]
  existingOwnerEntityIDs: string[]
}) {
  const qc = useQueryClient()
  const [mode, setMode] = useState<AddGroupPersonMode>(initialMode)
  const [search, setSearch] = useState("")
  const [selectedEntityID, setSelectedEntityID] = useState("")
  const [durationMode, setDurationMode] = useState<DurationPreset | "custom">("1mo")
  const [customAmount, setCustomAmount] = useState(7)
  const [customUnit, setCustomUnit] = useState<DurationUnit>("days")

  const usersQuery = useQuery({
    queryKey: ["users"],
    queryFn: async () => {
      const res = await api.get<UserOption[]>("/users")
      return res.data
    },
    staleTime: 5 * 60 * 1000,
    enabled: open,
  })

  const existing = useMemo(
    () => new Set(mode === "member" ? existingMemberEntityIDs : existingOwnerEntityIDs),
    [mode, existingMemberEntityIDs, existingOwnerEntityIDs],
  )

  const eligibleUsers = useMemo(
    () =>
      (usersQuery.data ?? []).filter(
        (user) => user.entity_id && !existing.has(user.entity_id),
      ),
    [usersQuery.data, existing],
  )

  const filteredUsers = useMemo(() => {
    const needle = search.trim()
    if (!needle) return eligibleUsers.slice(0, 8)
    return fuzzyFilter(eligibleUsers, needle, selectableText).slice(0, 8)
  }, [eligibleUsers, search])

  const typedEntityID = search.trim()
  const canUseTypedEntityID =
    ENTITY_ID_PATTERN.test(typedEntityID) && !existing.has(typedEntityID)
  const targetEntityID = selectedEntityID || (canUseTypedEntityID ? typedEntityID : "")
  const selectedUser = (usersQuery.data ?? []).find((user) => user.entity_id === targetEntityID)
  const memberExpiresAt = durationEnd(durationMode, customAmount, customUnit)

  const mutation = useMutation({
    mutationFn: async () => {
      if (!targetEntityID) return
      if (mode === "owner") {
        await api.post(`/groups/${groupID}/owners`, { entity_id: targetEntityID })
        return
      }
      if (durationMode === "custom") {
        if (
          !Number.isFinite(customAmount) ||
          customAmount <= 0 ||
          customAmount > MAX_BY_UNIT[customUnit]
        ) {
          throw new Error(`Pick between 1 and ${MAX_BY_UNIT[customUnit]} ${customUnit}.`)
        }
      }
      await api.post(`/groups/${groupID}/members`, {
        entity_id: targetEntityID,
        source: "DIRECT",
        has_expiration: true,
        expires_at: memberExpiresAt.toISOString(),
      })
    },
    onSuccess: async () => {
      await Promise.all([
        qc.invalidateQueries({ queryKey: ["group", groupID] }),
        qc.invalidateQueries({ queryKey: ["group", groupID, "members"] }),
        qc.invalidateQueries({ queryKey: ["group", groupID, "owners"] }),
      ])
      toast.success(mode === "member" ? "Member added" : "Owner added")
      onOpenChange(false)
    },
    onError: (err: unknown) => {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        (err as Error).message ??
        (mode === "member" ? "Couldn't add member." : "Couldn't add owner.")
      toast.error(message)
    },
  })

  function selectMode(nextMode: AddGroupPersonMode) {
    if (mutation.isPending) return
    setMode(nextMode)
    setSelectedEntityID("")
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        if (!mutation.isPending) onOpenChange(nextOpen)
      }}
    >
      <DialogContent className="gap-5 sm:max-w-lg">
        <DialogHeader className="gap-3">
          <div className="flex size-10 items-center justify-center rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-white">
            {mode === "member" ? (
              <UserPlus className="size-5" />
            ) : (
              <Crown className="size-5" />
            )}
          </div>
          <DialogTitle>Add to {groupName}</DialogTitle>
          <DialogDescription>
            Choose a person and assign their group access.
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-2 gap-2 rounded-lg bg-muted/40 p-1">
          <Button
            type="button"
            variant={mode === "member" ? "secondary" : "ghost"}
            onClick={() => selectMode("member")}
            className="gap-1.5"
          >
            <UserPlus className="size-3.5" />
            Member
          </Button>
          <Button
            type="button"
            variant={mode === "owner" ? "secondary" : "ghost"}
            onClick={() => selectMode("owner")}
            className="gap-1.5"
          >
            <Crown className="size-3.5" />
            Owner
          </Button>
        </div>

        <div className="space-y-2">
          <Label htmlFor="group-person-search">Person</Label>
          <div className="relative">
            <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              id="group-person-search"
              type="search"
              placeholder="Search by name, username, email, or entity ID"
              value={search}
              onChange={(event) => {
                setSearch(event.target.value)
                setSelectedEntityID("")
              }}
              className="pl-9"
            />
          </div>
          <div className="max-h-64 overflow-y-auto rounded-lg border border-border/60 bg-muted/20">
            {usersQuery.isLoading ? (
              <div className="space-y-2 p-3">
                <Skeleton className="h-10 w-full" />
                <Skeleton className="h-10 w-full" />
                <Skeleton className="h-10 w-full" />
              </div>
            ) : filteredUsers.length === 0 && !canUseTypedEntityID ? (
              <p className="p-4 text-center text-sm text-muted-foreground">
                No available users found.
              </p>
            ) : (
              <div className="divide-y divide-border/60">
                {canUseTypedEntityID && (
                  <button
                    type="button"
                    onClick={() => setSelectedEntityID(typedEntityID)}
                    className={
                      "flex w-full items-center gap-3 px-3 py-2.5 text-left transition-colors hover:bg-muted/60 " +
                      (selectedEntityID === typedEntityID ? "bg-muted/70" : "")
                    }
                  >
                    <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-muted font-mono text-[10px] text-muted-foreground">
                      ID
                    </div>
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-mono text-sm">{typedEntityID}</p>
                      <p className="truncate text-xs text-muted-foreground">Entity ID</p>
                    </div>
                  </button>
                )}
                {filteredUsers.map((user) => (
                  <button
                    key={user.id}
                    type="button"
                    onClick={() => setSelectedEntityID(user.entity_id)}
                    className={
                      "flex w-full items-center gap-3 px-3 py-2.5 text-left transition-colors hover:bg-muted/60 " +
                      (selectedEntityID === user.entity_id ? "bg-muted/70" : "")
                    }
                  >
                    <Avatar className="size-8">
                      {user.avatar_url && (
                        <AvatarImage src={user.avatar_url} alt={userName(user)} />
                      )}
                      <AvatarFallback className="text-xs">
                        {userInitials(user) || userName(user).slice(0, 1).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm">{userName(user)}</p>
                      <p className="truncate text-xs text-muted-foreground">
                        {user.username ? `@${user.username}` : user.entity_id}
                        {user.email ? ` - ${user.email}` : ""}
                      </p>
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>

        {mode === "member" && (
          <div className="space-y-2">
            <Label>Membership duration</Label>
            <div className="flex flex-wrap gap-2">
              {DURATION_PRESETS.map((preset) => {
                const active = durationMode === preset.value
                return (
                  <button
                    key={preset.value}
                    type="button"
                    onClick={() => setDurationMode(preset.value)}
                    className={
                      "rounded-md border px-3 py-1.5 text-sm transition-colors " +
                      (active
                        ? "border-gr-pink/40 bg-gr-pink/10 text-foreground"
                        : "border-border/60 bg-muted/30 text-muted-foreground hover:bg-muted/50")
                    }
                  >
                    {preset.label}
                  </button>
                )
              })}
              <button
                type="button"
                onClick={() => setDurationMode("custom")}
                className={
                  "rounded-md border px-3 py-1.5 text-sm transition-colors " +
                  (durationMode === "custom"
                    ? "border-gr-pink/40 bg-gr-pink/10 text-foreground"
                    : "border-border/60 bg-muted/30 text-muted-foreground hover:bg-muted/50")
                }
              >
                Custom
              </button>
            </div>
            {durationMode === "custom" && (
              <div className="flex flex-wrap items-center gap-2 pt-1">
                <Input
                  type="number"
                  min={1}
                  max={MAX_BY_UNIT[customUnit]}
                  value={customAmount}
                  onChange={(event) =>
                    setCustomAmount(parseInt(event.target.value, 10) || 0)
                  }
                  className="w-24"
                />
                <Select
                  value={customUnit}
                  onValueChange={(value) => setCustomUnit(value as DurationUnit)}
                >
                  <SelectTrigger className="w-32">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="hours">Hours</SelectItem>
                    <SelectItem value="days">Days</SelectItem>
                    <SelectItem value="weeks">Weeks</SelectItem>
                    <SelectItem value="months">Months</SelectItem>
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  Max {MAX_BY_UNIT[customUnit]}
                </p>
              </div>
            )}
            <p className="text-xs text-muted-foreground">
              Ends {formatAbsoluteDate(memberExpiresAt.toISOString())}
            </p>
          </div>
        )}

        {targetEntityID && (
          <div className="rounded-lg border border-border/60 bg-muted/30 p-3">
            <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Selected
            </p>
            <p className="mt-1 text-sm">
              {selectedUser ? userName(selectedUser) : targetEntityID}
            </p>
            {selectedUser?.username && (
              <p className="text-xs text-muted-foreground">@{selectedUser.username}</p>
            )}
          </div>
        )}

        <div className="flex justify-end gap-2 pt-1">
          <Button
            type="button"
            variant="ghost"
            disabled={mutation.isPending}
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button
            type="button"
            disabled={!targetEntityID || mutation.isPending}
            onClick={() => mutation.mutate()}
          >
            {mutation.isPending
              ? mode === "member"
                ? "Adding member..."
                : "Adding owner..."
              : mode === "member"
                ? "Add member"
                : "Add owner"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
