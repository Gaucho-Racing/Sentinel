import { useQuery, useQueryClient } from "@tanstack/react-query"
import { ArrowLeft, Bot, Plus, Trash2, X } from "lucide-react"
import { useEffect, useState } from "react"
import { Link, useNavigate, useParams } from "react-router-dom"
import { toast } from "sonner"

import { OutlineButton } from "@/components/OutlineButton"
import { PageContainer } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Skeleton } from "@/components/ui/skeleton"
import { useAdmins } from "@/lib/admin"
import { api } from "@/lib/api"
import { loadSession } from "@/lib/auth"
import {
  discordRoleColorHex,
  useDiscordRoles,
  useGroupDiscordBindings,
  type GroupDiscordRoleBinding,
} from "@/lib/discord"
import type { Group, GroupMember, GroupOwner, GroupSource } from "@/lib/groups"

import { DiscordRolePickerDialog } from "./DiscordRolePickerDialog"
import { GroupForm, type GroupFormValues } from "./GroupForm"

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

  useEffect(() => {
    if (query.data && values === null) {
      setValues({
        name: query.data.name,
        description: query.data.description,
        allowed_sources: query.data.allowed_sources ?? [],
      })
    }
  }, [query.data, values])

  async function commitSave() {
    if (!values || !id) return
    setSubmitting(true)
    try {
      // Apply staged binding changes only if Discord is staying enabled.
      // When DISCORD is being unchecked the backend cascade in
      // CreateOrUpdateGroup wipes every binding for the group, so any pending
      // adds/removes here are wasted work.
      const keepingDiscord = values.allowed_sources.includes("DISCORD")
      if (keepingDiscord) {
        for (const bindingID of pendingBindingRemoves) {
          await api.delete(`/groups/${id}/discord-bindings/${bindingID}`)
        }
        for (const add of pendingBindingAdds) {
          await api.post(`/groups/${id}/discord-bindings`, {
            discord_role_ids: add.discord_role_ids,
          })
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
