import { useQuery, useQueryClient } from "@tanstack/react-query"
import { ArrowLeft, Bot, Plus, Sparkles, Trash2, X } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { useAdmins } from "@/lib/admin"
import { api } from "@/lib/api"
import type { Application, ApplicationWithLink } from "@/lib/applications"
import { loadSession } from "@/lib/auth"
import {
  useGroupConditionalBindings,
  type GroupConditionalBinding,
} from "@/lib/conditional"
import {
  discordRoleColorHex,
  useDiscordRoles,
  useGroupDiscordBindings,
  type GroupDiscordRoleBinding,
} from "@/lib/discord"
import type { Group, GroupMember, GroupOwner, GroupSource } from "@/lib/groups"

import { DiscordRolePickerDialog } from "./DiscordRolePickerDialog"
import { GroupForm, type GroupFormValues } from "./GroupForm"
import { GroupPickerDialog } from "./GroupPickerDialog"

function DiscordSyncCard({
  bindings,
  onAddBinding,
  onRemoveBinding,
}: {
  bindings: GroupDiscordRoleBinding[]
  onAddBinding: (roleIDs: string[]) => void
  onRemoveBinding: (bindingID: string) => void
}) {
  const [pickerOpen, setPickerOpen] = useState(false)
  const rolesQuery = useDiscordRoles()
  const roles = rolesQuery.data ?? []

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Bot className="size-4 text-muted-foreground" />
          Discord role sync
        </CardTitle>
        <CardDescription>
          Each binding is an AND-group of Discord roles — users must have every role
          in the binding to be synced through it. Group membership is the OR across
          all bindings, so add multiple bindings for either-or rules.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {bindings.length === 0 ? (
          <p className="text-sm text-muted-foreground">No Discord role bindings yet.</p>
        ) : (
          <ul className="space-y-2">
            {bindings.map((binding) => (
              <li
                key={binding.id}
                className="flex items-center justify-between gap-3 rounded-md border border-border/60 bg-muted/40 px-3 py-2"
              >
                <div className="flex min-w-0 flex-wrap items-center gap-1.5">
                  {binding.discord_role_ids.map((roleID, idx) => {
                    const role = roles.find((r) => r.id === roleID)
                    const hex = role ? discordRoleColorHex(role.color) : null
                    return (
                      <span key={roleID} className="flex items-center gap-1.5">
                        {idx > 0 && (
                          <span className="text-xs font-medium text-muted-foreground">
                            AND
                          </span>
                        )}
                        <span className="inline-flex items-center gap-1.5 rounded-md border border-border/60 bg-background/60 px-2 py-0.5">
                          <span
                            className="size-2 shrink-0 rounded-full border border-border/60"
                            style={{ backgroundColor: hex ?? "transparent" }}
                          />
                          <span className="text-sm">
                            {role ? `@${role.name}` : (
                              <code className="font-mono text-xs text-muted-foreground">
                                {roleID}
                              </code>
                            )}
                          </span>
                        </span>
                      </span>
                    )
                  })}
                </div>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  onClick={() => onRemoveBinding(binding.id)}
                >
                  <X className="size-3.5" />
                </Button>
              </li>
            ))}
          </ul>
        )}
        <div className="pt-1">
          <Button type="button" onClick={() => setPickerOpen(true)}>
            <Plus className="mr-1 size-3.5" />
            Add role binding
          </Button>
        </div>
      </CardContent>

      <DiscordRolePickerDialog
        open={pickerOpen}
        onOpenChange={setPickerOpen}
        onAddBinding={onAddBinding}
      />
    </Card>
  )
}

function ConditionalSyncCard({
  groupID,
  bindings,
  groupNamesByID,
  onAddBinding,
  onRemoveBinding,
}: {
  groupID: string
  bindings: GroupConditionalBinding[]
  groupNamesByID: Record<string, string>
  onAddBinding: (groupIDs: string[]) => void
  onRemoveBinding: (bindingID: string) => void
}) {
  const [pickerOpen, setPickerOpen] = useState(false)

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Sparkles className="size-4 text-muted-foreground" />
          Conditional rule
        </CardTitle>
        <CardDescription>
          Each binding is an AND-group of other Sentinel groups — entities must be a
          member of every group in the binding to be synced through it. Group membership
          is the OR across all bindings, so add multiple bindings for either-or rules.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {bindings.length === 0 ? (
          <p className="text-sm text-muted-foreground">No conditional bindings yet.</p>
        ) : (
          <ul className="space-y-2">
            {bindings.map((binding) => (
              <li
                key={binding.id}
                className="flex items-center justify-between gap-3 rounded-md border border-border/60 bg-muted/40 px-3 py-2"
              >
                <div className="flex min-w-0 flex-wrap items-center gap-1.5">
                  {binding.required_group_ids.map((reqID, idx) => {
                    const name = groupNamesByID[reqID]
                    return (
                      <span key={reqID} className="flex items-center gap-1.5">
                        {idx > 0 && (
                          <span className="text-xs font-medium text-muted-foreground">
                            AND
                          </span>
                        )}
                        <span className="inline-flex items-center gap-1.5 rounded-md border border-border/60 bg-background/60 px-2 py-0.5">
                          <span className="text-sm">
                            {name ?? (
                              <code className="font-mono text-xs text-muted-foreground">
                                {reqID}
                              </code>
                            )}
                          </span>
                        </span>
                      </span>
                    )
                  })}
                </div>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  onClick={() => onRemoveBinding(binding.id)}
                >
                  <X className="size-3.5" />
                </Button>
              </li>
            ))}
          </ul>
        )}
        <div className="pt-1">
          <Button type="button" onClick={() => setPickerOpen(true)}>
            <Plus className="mr-1 size-3.5" />
            Add conditional binding
          </Button>
        </div>
      </CardContent>

      <GroupPickerDialog
        open={pickerOpen}
        onOpenChange={setPickerOpen}
        excludeGroupID={groupID}
        onAddBinding={onAddBinding}
      />
    </Card>
  )
}

function LinkedApplicationsCard({
  links,
  allApps,
  onAdd,
  onRemove,
  onToggleRequired,
}: {
  links: { id: string; name: string; required: boolean }[]
  allApps: Application[]
  onAdd: (appID: string, required: boolean) => void
  onRemove: (appID: string) => void
  onToggleRequired: (appID: string) => void
}) {
  const [draftAppID, setDraftAppID] = useState("")
  const [draftRequired, setDraftRequired] = useState(false)

  const linkedSet = useMemo(() => new Set(links.map((l) => l.id)), [links])
  const availableApps = useMemo(
    () => allApps.filter((a) => !linkedSet.has(a.id)),
    [allApps, linkedSet],
  )

  function handleAdd() {
    if (!draftAppID) return
    onAdd(draftAppID, draftRequired)
    setDraftAppID("")
    setDraftRequired(false)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Linked applications</CardTitle>
        <CardDescription>
          Applications linked to this group see members in their token's{" "}
          <code className="font-mono text-xs">groups</code> claim. Marking the link as{" "}
          <span className="font-medium">required</span> gates the app — only users in this group
          can obtain a token for it.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {links.length === 0 ? (
          <p className="text-sm text-muted-foreground">No applications linked yet.</p>
        ) : (
          <ul className="space-y-2">
            {links.map((link) => (
              <li
                key={link.id}
                className="flex items-center gap-2 rounded-md border border-border/60 bg-muted/40 px-2.5 py-1.5"
              >
                <span className="flex-1 truncate text-sm">{link.name}</span>
                <Badge
                  variant={link.required ? "default" : "outline"}
                  className="cursor-pointer select-none"
                  onClick={() => onToggleRequired(link.id)}
                  title={
                    link.required
                      ? "Required for access — click to make optional"
                      : "Optional — click to require for access"
                  }
                >
                  {link.required ? "Required" : "Optional"}
                </Badge>
                <Button variant="ghost" size="icon-sm" onClick={() => onRemove(link.id)}>
                  <X className="size-3.5" />
                </Button>
              </li>
            ))}
          </ul>
        )}
        <div className="flex flex-wrap items-center gap-2 pt-2">
          <Select value={draftAppID} onValueChange={setDraftAppID}>
            <SelectTrigger className="flex-1 min-w-[180px]">
              <SelectValue placeholder="Select an application to link…" />
            </SelectTrigger>
            <SelectContent>
              {availableApps.length === 0 ? (
                <div className="px-2 py-1.5 text-xs text-muted-foreground">
                  No more applications to link.
                </div>
              ) : (
                availableApps.map((a) => (
                  <SelectItem key={a.id} value={a.id}>
                    {a.name}
                  </SelectItem>
                ))
              )}
            </SelectContent>
          </Select>
          <Badge
            variant={draftRequired ? "default" : "outline"}
            className="cursor-pointer select-none"
            onClick={() => setDraftRequired((v) => !v)}
          >
            {draftRequired ? "Required" : "Optional"}
          </Badge>
          <Button type="button" disabled={!draftAppID} onClick={handleAdd}>
            <Plus className="mr-1 size-3.5" />
            Link
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

export default function GroupEditPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const myEntityID = loadSession()?.entityId ?? ""
  const { isAdmin, isLoading: adminsLoading } = useAdmins()

  const query = useQuery({
    queryKey: ["group", id],
    queryFn: async () => {
      const res = await api.get<Group>(`/groups/${id}`)
      return res.data
    },
    enabled: !!id,
  })

  const ownersQuery = useQuery({
    queryKey: ["group", id, "owners"],
    queryFn: async () => {
      const res = await api.get<GroupOwner[]>(`/groups/${id}/owners`)
      return res.data
    },
    enabled: !!id,
  })

  const membersQuery = useQuery({
    queryKey: ["group", id, "members"],
    queryFn: async () => {
      const res = await api.get<GroupMember[]>(`/groups/${id}/members`)
      return res.data
    },
    enabled: !!id,
  })

  const bindingsQuery = useGroupDiscordBindings(id ?? "")
  const conditionalBindingsQuery = useGroupConditionalBindings(id ?? "")

  // All groups, used by the conditional editor to resolve required_group_ids
  // → names for the chips and to feed the picker dialog. Cheap query for
  // typical org scale.
  const allGroupsQuery = useQuery({
    queryKey: ["groups"],
    queryFn: async () => {
      const res = await api.get<Group[]>("/groups")
      return res.data
    },
  })
  const groupNamesByID = useMemo(() => {
    const m: Record<string, string> = {}
    for (const g of allGroupsQuery.data ?? []) m[g.id] = g.name
    return m
  }, [allGroupsQuery.data])

  const linkedAppsQuery = useQuery({
    queryKey: ["group", id, "applications"],
    queryFn: async () => {
      const res = await api.get<ApplicationWithLink[]>(`/groups/${id}/applications`)
      return res.data
    },
    enabled: !!id,
  })

  const allAppsQuery = useQuery({
    queryKey: ["applications"],
    queryFn: async () => {
      const res = await api.get<Application[]>(`/applications`)
      return res.data
    },
    staleTime: 5 * 60 * 1000,
  })

  // Staged application-link state. Map<application_id, required>. Mirrors
  // the LinkedGroupsCard pattern on the application edit page.
  const [appLinkState, setAppLinkState] = useState<Map<string, boolean>>(new Map())
  const [appLinksInitialized, setAppLinksInitialized] = useState(false)

  const [values, setValues] = useState<GroupFormValues | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [cascadeConfirmOpen, setCascadeConfirmOpen] = useState(false)
  // Pending binding state — staged changes are applied to the server in
  // commitSave alongside the basics, so the Save button is the single commit
  // point for the entire page.
  const [pendingBindingAdds, setPendingBindingAdds] = useState<
    { tempID: string; discord_role_ids: string[] }[]
  >([])
  const [pendingBindingRemoves, setPendingBindingRemoves] = useState<Set<string>>(
    new Set(),
  )
  const [pendingConditionalAdds, setPendingConditionalAdds] = useState<
    { tempID: string; required_group_ids: string[] }[]
  >([])
  const [pendingConditionalRemoves, setPendingConditionalRemoves] = useState<Set<string>>(
    new Set(),
  )

  const serverBindings = bindingsQuery.data ?? []
  const effectiveBindings: GroupDiscordRoleBinding[] = [
    ...serverBindings.filter((b) => !pendingBindingRemoves.has(b.id)),
    ...pendingBindingAdds.map((p) => ({
      id: p.tempID,
      group_id: id ?? "",
      discord_role_ids: p.discord_role_ids,
      created_at: "",
    })),
  ]

  function handleAddBinding(roleIDs: string[]) {
    if (roleIDs.length === 0) return
    setPendingBindingAdds((prev) => [
      ...prev,
      {
        tempID: `pending_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 6)}`,
        discord_role_ids: roleIDs,
      },
    ])
  }

  function handleRemoveBinding(bindingID: string) {
    // A pending add gets dropped from the pendingAdds list. A server-side
    // binding goes into pendingRemoves so commitSave can issue the DELETE.
    if (bindingID.startsWith("pending_")) {
      setPendingBindingAdds((prev) => prev.filter((p) => p.tempID !== bindingID))
      return
    }
    setPendingBindingRemoves((prev) => {
      const next = new Set(prev)
      next.add(bindingID)
      return next
    })
  }

  const serverConditionalBindings = conditionalBindingsQuery.data ?? []
  const effectiveConditionalBindings: GroupConditionalBinding[] = [
    ...serverConditionalBindings.filter((b) => !pendingConditionalRemoves.has(b.id)),
    ...pendingConditionalAdds.map((p) => ({
      id: p.tempID,
      group_id: id ?? "",
      required_group_ids: p.required_group_ids,
      created_at: "",
    })),
  ]

  function handleAddConditionalBinding(groupIDs: string[]) {
    if (groupIDs.length === 0) return
    setPendingConditionalAdds((prev) => [
      ...prev,
      {
        tempID: `pending_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 6)}`,
        required_group_ids: groupIDs,
      },
    ])
  }

  function handleRemoveConditionalBinding(bindingID: string) {
    if (bindingID.startsWith("pending_")) {
      setPendingConditionalAdds((prev) => prev.filter((p) => p.tempID !== bindingID))
      return
    }
    setPendingConditionalRemoves((prev) => {
      const next = new Set(prev)
      next.add(bindingID)
      return next
    })
  }

  useEffect(() => {
    if (query.data && values === null) {
      setValues({
        name: query.data.name,
        description: query.data.description,
        allowed_sources: query.data.allowed_sources ?? [],
      })
    }
  }, [query.data, values])

  useEffect(() => {
    if (linkedAppsQuery.data && !appLinksInitialized) {
      const m = new Map<string, boolean>()
      for (const a of linkedAppsQuery.data) m.set(a.id, a.required)
      setAppLinkState(m)
      setAppLinksInitialized(true)
    }
  }, [linkedAppsQuery.data, appLinksInitialized])

  const appLinkList = useMemo(() => {
    const byID = new Map((allAppsQuery.data ?? []).map((a) => [a.id, a.name]))
    return Array.from(appLinkState, ([id, required]) => ({
      id,
      name: byID.get(id) ?? id,
      required,
    }))
  }, [appLinkState, allAppsQuery.data])

  function handleAddAppLink(appID: string, required: boolean) {
    setAppLinkState((prev) => {
      const next = new Map(prev)
      next.set(appID, required)
      return next
    })
  }

  function handleRemoveAppLink(appID: string) {
    setAppLinkState((prev) => {
      const next = new Map(prev)
      next.delete(appID)
      return next
    })
  }

  function handleToggleAppRequired(appID: string) {
    setAppLinkState((prev) => {
      const next = new Map(prev)
      const cur = next.get(appID)
      if (cur === undefined) return prev
      next.set(appID, !cur)
      return next
    })
  }

  async function commitSave() {
    if (!values || !id) return
    setSubmitting(true)
    try {
      // Only apply staged binding changes if Discord is staying enabled.
      // If DISCORD is being unchecked the group will stop honoring bindings
      // regardless, so any pending edits would just create orphans.
      const keepingDiscord = values.allowed_sources.includes("DISCORD")
      if (keepingDiscord) {
        for (const bindingID of pendingBindingRemoves) {
          await api.delete(`/discord/role-bindings/${bindingID}`, {
            params: { group_id: id },
          })
        }
        for (const add of pendingBindingAdds) {
          await api.post(`/discord/role-bindings`, {
            group_id: id,
            discord_role_ids: add.discord_role_ids,
          })
        }
      }

      // Same gate for conditional bindings — apply staged edits only if
      // CONDITIONAL is staying enabled. cascadeRemovedSources on the
      // backend will already strip CONDITIONAL members when the source is
      // disabled, so adds/removes here would be wasted writes.
      const keepingConditional = values.allowed_sources.includes("CONDITIONAL")
      if (keepingConditional) {
        for (const bindingID of pendingConditionalRemoves) {
          await api.delete(`/groups/${id}/conditional-bindings/${bindingID}`)
        }
        for (const add of pendingConditionalAdds) {
          await api.post(`/groups/${id}/conditional-bindings`, {
            required_group_ids: add.required_group_ids,
          })
        }
      }
      // Diff application links against the server state. POST is upsert,
      // so we send any link whose required flag differs (or doesn't exist
      // yet); DELETE anything the server has that's no longer in our state.
      const serverAppLinks = linkedAppsQuery.data ?? []
      const serverAppByID = new Map(serverAppLinks.map((l) => [l.id, l.required]))
      for (const [appID, required] of appLinkState) {
        const serverReq = serverAppByID.get(appID)
        if (serverReq === undefined || serverReq !== required) {
          await api.post(`/applications/${appID}/groups`, {
            group_id: id,
            required,
          })
        }
      }
      for (const a of serverAppLinks) {
        if (!appLinkState.has(a.id)) {
          await api.delete(`/applications/${a.id}/groups/${id}`)
        }
      }

      await api.post<Group>("/groups", {
        id,
        name: values.name.trim(),
        description: values.description,
        allowed_sources: values.allowed_sources,
      })
      qc.invalidateQueries({ queryKey: ["groups"] })
      qc.invalidateQueries({ queryKey: ["group", id] })
      qc.invalidateQueries({ queryKey: ["group", id, "members"] })
      qc.invalidateQueries({ queryKey: ["group", id, "discord-bindings"] })
      qc.invalidateQueries({ queryKey: ["group", id, "applications"] })
      toast.success("Group updated")
      navigate(`/groups/${id}`)
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't save the group."
      toast.error(message)
    } finally {
      setSubmitting(false)
      setCascadeConfirmOpen(false)
    }
  }

  function handleSubmit() {
    if (!values || !id || !query.data) return
    const savedSources = new Set(query.data.allowed_sources ?? [])
    const draftSources = new Set(values.allowed_sources)
    const removed: GroupSource[] = []
    for (const s of savedSources) {
      if (!draftSources.has(s) && s !== "DIRECT") removed.push(s as GroupSource)
    }
    const removingDiscord = removed.includes("DISCORD")
    const removingConditional = removed.includes("CONDITIONAL")
    // Effective binding count = what the user sees in the UI right now
    // (server bindings minus staged removes plus staged adds).
    const effectiveBindingCount = effectiveBindings.length
    const members = membersQuery.data ?? []
    const discordMemberCount = members.filter((m) => m.source === "DISCORD").length
    const conditionalMemberCount = members.filter((m) => m.source === "CONDITIONAL").length
    const hasDestructiveImpact =
      (removingDiscord && (effectiveBindingCount > 0 || discordMemberCount > 0)) ||
      (removingConditional && conditionalMemberCount > 0)
    if (hasDestructiveImpact) {
      setCascadeConfirmOpen(true)
      return
    }
    void commitSave()
  }

  async function handleDelete() {
    if (!id || deleting) return
    setDeleting(true)
    try {
      await api.delete(`/groups/${id}`)
      qc.invalidateQueries({ queryKey: ["groups"] })
      toast.success("Group deleted")
      navigate("/groups")
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't delete the group."
      toast.error(message)
      setDeleting(false)
      setConfirmOpen(false)
    }
  }

  if (
    query.isLoading ||
    !values ||
    ownersQuery.isLoading ||
    membersQuery.isLoading ||
    bindingsQuery.isLoading ||
    conditionalBindingsQuery.isLoading ||
    adminsLoading
  ) {
    return (
      <PageContainer>
        <Skeleton className="mb-4 h-4 w-24" />
        <Skeleton className="mb-8 h-8 w-48" />
        <Skeleton className="h-96" />
      </PageContainer>
    )
  }

  if (query.isError || !query.data) {
    return (
      <PageContainer>
        <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
          <Link to="/groups">
            <ArrowLeft className="mr-1 size-3.5" />
            All groups
          </Link>
        </Button>
        <p className="text-sm text-muted-foreground">Group not found.</p>
      </PageContainer>
    )
  }

  const group = query.data
  const owners = ownersQuery.data ?? []
  const isOwner = !!myEntityID && owners.some((o) => o.entity_id === myEntityID)
  const canEdit = isOwner || isAdmin

  if (!canEdit) {
    return (
      <PageContainer>
        <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
          <Link to={`/groups/${id}`}>
            <ArrowLeft className="mr-1 size-3.5" />
            Back to {group.name}
          </Link>
        </Button>
        <p className="text-sm text-muted-foreground">
          Only group owners and admins can edit this group.
        </p>
      </PageContainer>
    )
  }

  return (
    <PageContainer>
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to={`/groups/${id}`}>
          <ArrowLeft className="mr-1 size-3.5" />
          Back to {group.name}
        </Link>
      </Button>

      <header className="mb-8 flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0">
          <h1 className="text-2xl font-semibold tracking-tight">Edit {group.name}</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Update the group's basics, sync sources, and rules. Nothing is saved until
            you click Save changes.
          </p>
        </div>
        <OutlineButton
          type="button"
          className="w-auto"
          loading={submitting}
          disabled={!values.name.trim()}
          onClick={handleSubmit}
        >
          Save changes
        </OutlineButton>
      </header>

      <div className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>Basics</CardTitle>
            <CardDescription>Name, description, and source configuration.</CardDescription>
          </CardHeader>
          <CardContent>
            <GroupForm
              values={values}
              onChange={setValues}
              onSubmit={handleSubmit}
              submitting={submitting}
              submitLabel="Save changes"
              hideSubmit
            />
          </CardContent>
        </Card>

        {values.allowed_sources.includes("DISCORD") && (
          <DiscordSyncCard
            bindings={effectiveBindings}
            onAddBinding={handleAddBinding}
            onRemoveBinding={handleRemoveBinding}
          />
        )}

        {values.allowed_sources.includes("CONDITIONAL") && id && (
          <ConditionalSyncCard
            groupID={id}
            bindings={effectiveConditionalBindings}
            groupNamesByID={groupNamesByID}
            onAddBinding={handleAddConditionalBinding}
            onRemoveBinding={handleRemoveConditionalBinding}
          />
        )}

        <LinkedApplicationsCard
          links={appLinkList}
          allApps={allAppsQuery.data ?? []}
          onAdd={handleAddAppLink}
          onRemove={handleRemoveAppLink}
          onToggleRequired={handleToggleAppRequired}
        />

        <Card>
          <CardHeader>
            <CardTitle>Danger zone</CardTitle>
            <CardDescription>
              Deleting a group removes all member, owner, and join-request rows linked to it. Linked application bindings stay (the apps just stop granting access via this group).
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button
              variant="destructive"
              disabled={deleting}
              onClick={() => setConfirmOpen(true)}
            >
              <Trash2 className="mr-1 size-3.5" />
              Delete group
            </Button>
          </CardContent>
        </Card>
      </div>

      <Dialog
        open={confirmOpen}
        onOpenChange={(open) => {
          if (!deleting) setConfirmOpen(open)
        }}
      >
        <DialogContent className="gap-5 sm:max-w-md">
          <DialogHeader className="gap-3">
            <div className="flex size-10 items-center justify-center rounded-xl bg-destructive/10 text-destructive">
              <Trash2 className="size-5" />
            </div>
            <DialogTitle>Delete {group.name}?</DialogTitle>
            <DialogDescription>
              This permanently removes the group along with all member, owner, and
              join-request rows linked to it. Linked application bindings stay (the apps
              just stop granting access via this group). This can't be undone.
            </DialogDescription>
          </DialogHeader>

          <div className="flex justify-end gap-2 pt-1">
            <Button
              type="button"
              variant="ghost"
              disabled={deleting}
              onClick={() => setConfirmOpen(false)}
            >
              Cancel
            </Button>
            <Button
              type="button"
              variant="destructive"
              disabled={deleting}
              onClick={handleDelete}
            >
              {deleting ? "Deleting…" : "Delete group"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog
        open={cascadeConfirmOpen}
        onOpenChange={(open) => {
          if (!submitting) setCascadeConfirmOpen(open)
        }}
      >
        <DialogContent className="gap-5 sm:max-w-md">
          <DialogHeader className="gap-3">
            <div className="flex size-10 items-center justify-center rounded-xl bg-destructive/10 text-destructive">
              <Trash2 className="size-5" />
            </div>
            <DialogTitle>Save changes?</DialogTitle>
            <DialogDescription>
              You're removing sources that currently have configuration or synced members.
              Saving will also wipe the items below. Direct members are not affected.
            </DialogDescription>
          </DialogHeader>

          {(() => {
            const savedSources = new Set(query.data?.allowed_sources ?? [])
            const draftSources = new Set(values.allowed_sources)
            const removingDiscord =
              savedSources.has("DISCORD") && !draftSources.has("DISCORD")
            const removingConditional =
              savedSources.has("CONDITIONAL") && !draftSources.has("CONDITIONAL")
            const bindingCount = effectiveBindings.length
            const members = membersQuery.data ?? []
            const discordMembers = members.filter((m) => m.source === "DISCORD").length
            const conditionalMembers = members.filter(
              (m) => m.source === "CONDITIONAL",
            ).length
            return (
              <ul className="space-y-2 text-sm">
                {removingDiscord && (
                  <li className="rounded-md border border-border/60 bg-muted/30 p-3">
                    <p className="font-medium">Discord sync</p>
                    <p className="mt-0.5 text-xs text-muted-foreground">
                      {bindingCount} role binding{bindingCount === 1 ? "" : "s"} will be
                      deleted ·{" "}
                      {discordMembers} synced member{discordMembers === 1 ? "" : "s"} will be
                      removed from the group.
                    </p>
                  </li>
                )}
                {removingConditional && (
                  <li className="rounded-md border border-border/60 bg-muted/30 p-3">
                    <p className="font-medium">Conditional rule</p>
                    <p className="mt-0.5 text-xs text-muted-foreground">
                      The rule will be deleted ·{" "}
                      {conditionalMembers} synced member
                      {conditionalMembers === 1 ? "" : "s"} will be removed from the group.
                    </p>
                  </li>
                )}
              </ul>
            )
          })()}

          <div className="flex justify-end gap-2 pt-1">
            <Button
              type="button"
              variant="ghost"
              disabled={submitting}
              onClick={() => setCascadeConfirmOpen(false)}
            >
              Keep editing
            </Button>
            <Button
              type="button"
              variant="destructive"
              disabled={submitting}
              onClick={commitSave}
            >
              {submitting ? "Saving…" : "Save anyway"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </PageContainer>
  )
}
