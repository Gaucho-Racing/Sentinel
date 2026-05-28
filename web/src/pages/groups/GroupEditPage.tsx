import { useQuery, useQueryClient } from "@tanstack/react-query"
import { ArrowLeft, Bot, Plus, Trash2, X } from "lucide-react"
import { useEffect, useState } from "react"
import { Link, useNavigate, useParams } from "react-router-dom"
import { toast } from "sonner"

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
  useRemoveGroupDiscordBinding,
} from "@/lib/discord"
import type { Group, GroupOwner } from "@/lib/groups"

import { DiscordRolePickerDialog } from "./DiscordRolePickerDialog"
import { GroupForm, type GroupFormValues } from "./GroupForm"

function DiscordSyncCard({ groupID }: { groupID: string }) {
  const [pickerOpen, setPickerOpen] = useState(false)
  const bindingsQuery = useGroupDiscordBindings(groupID)
  const rolesQuery = useDiscordRoles()
  const removeBinding = useRemoveGroupDiscordBinding(groupID)

  const bindings = bindingsQuery.data ?? []
  const roles = rolesQuery.data ?? []

  async function handleRemove(bindingID: string) {
    if (removeBinding.isPending) return
    try {
      await removeBinding.mutateAsync(bindingID)
      toast.success("Binding removed")
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't remove binding."
      toast.error(message)
    }
  }

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
                  disabled={removeBinding.isPending}
                  onClick={() => handleRemove(binding.id)}
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
        groupID={groupID}
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

  const [values, setValues] = useState<GroupFormValues | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [confirmOpen, setConfirmOpen] = useState(false)

  useEffect(() => {
    if (query.data && values === null) {
      setValues({
        name: query.data.name,
        description: query.data.description,
        allowed_sources: query.data.allowed_sources ?? [],
      })
    }
  }, [query.data, values])

  async function handleSubmit() {
    if (!values || !id) return
    setSubmitting(true)
    try {
      await api.post<Group>("/groups", {
        id,
        name: values.name.trim(),
        description: values.description,
        allowed_sources: values.allowed_sources,
      })
      qc.invalidateQueries({ queryKey: ["groups"] })
      qc.invalidateQueries({ queryKey: ["group", id] })
      toast.success("Group updated")
      navigate(`/groups/${id}`)
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't save the group."
      toast.error(message)
    } finally {
      setSubmitting(false)
    }
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

      <div className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight">Edit {group.name}</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Update the group's basics and allowed sources.
        </p>
      </div>

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
            />
          </CardContent>
        </Card>

        {values.allowed_sources.includes("DISCORD") && (
          <DiscordSyncCard groupID={id!} />
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
    </PageContainer>
  )
}
