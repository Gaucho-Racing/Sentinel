import { AxiosError } from "axios"
import { ArrowLeft, ChevronRight, Network, Pencil, Plus, Trash2, X } from "lucide-react"
import { useMemo, useState } from "react"
import { Link } from "react-router-dom"
import { toast } from "sonner"

import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
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
import { Skeleton } from "@/components/ui/skeleton"
import { useAdmins } from "@/lib/admin"
import { useGroups } from "@/lib/groups"
import {
  buildOrgTree,
  useCreateOrgPosition,
  useDeleteOrgPosition,
  useOrgChart,
  useUpdateOrgPosition,
  type OrgPositionInput,
  type OrgTreeNode,
} from "@/lib/org"
import { userName, useUsers } from "@/lib/users"
import { cn } from "@/lib/utils"

type FormState = OrgPositionInput & { id?: string }

const EMPTY_FORM: FormState = {
  title: "",
  parent_id: "",
  entity_id: "",
  group_id: "",
  rank: 0,
}

function apiError(error: unknown, fallback: string) {
  if (error instanceof AxiosError) {
    const message = (error.response?.data as { error?: string } | undefined)?.error
    if (message) return message
  }
  return fallback
}

// A searchable single-select. The team is large enough that a plain Select over
// every member is unusable, and cmdk is already in the bundle for the header
// search.
function Picker({
  label,
  placeholder,
  value,
  options,
  onChange,
}: {
  label: string
  placeholder: string
  value: string
  options: { value: string; label: string; hint?: string }[]
  onChange: (next: string) => void
}) {
  const [open, setOpen] = useState(false)
  const selected = options.find((o) => o.value === value)

  return (
    <div className="space-y-1.5">
      <Label>{label}</Label>
      <div className="flex items-center gap-1.5">
        <Button
          type="button"
          variant="outline"
          className="min-w-0 flex-1 justify-start font-normal"
          onClick={() => setOpen(true)}
        >
          <span className={cn("truncate", !selected && "text-muted-foreground")}>
            {selected?.label ?? placeholder}
          </span>
        </Button>
        {value && (
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            aria-label={`Clear ${label}`}
            onClick={() => onChange("")}
          >
            <X className="size-3.5" />
          </Button>
        )}
      </div>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="p-0">
          <DialogHeader className="sr-only">
            <DialogTitle>{label}</DialogTitle>
          </DialogHeader>
          <Command>
            <CommandInput placeholder={`Search ${label.toLowerCase()}…`} />
            <CommandList>
              <CommandEmpty>No match.</CommandEmpty>
              <CommandGroup>
                {options.map((option) => (
                  <CommandItem
                    key={option.value}
                    value={`${option.label} ${option.hint ?? ""}`}
                    onSelect={() => {
                      onChange(option.value)
                      setOpen(false)
                    }}
                  >
                    <span className="truncate">{option.label}</span>
                    {option.hint && (
                      <span className="ml-auto truncate text-xs text-muted-foreground">
                        {option.hint}
                      </span>
                    )}
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
          </Command>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function PositionRow({
  node,
  onAddChild,
  onEdit,
  onDelete,
}: {
  node: OrgTreeNode
  onAddChild: (parentID: string) => void
  onEdit: (node: OrgTreeNode) => void
  onDelete: (node: OrgTreeNode) => void
}) {
  return (
    <>
      <li
        className="flex items-center gap-2 px-4 py-2.5"
        style={{ paddingLeft: `${node.depth * 20 + 16}px` }}
      >
        {node.depth > 0 && <ChevronRight className="size-3 shrink-0 text-muted-foreground" />}
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium leading-none">{node.position.title}</p>
          <p className="mt-1 truncate text-xs text-muted-foreground">
            {node.holder?.name ?? "Vacant"}
          </p>
        </div>
        {node.group && (
          <Badge
            variant="outline"
            className={cn(
              "h-5 text-[10px]",
              node.group.missing && "border-destructive/50 text-destructive",
            )}
          >
            {node.group.missing ? "group deleted" : `${node.group.name} · ${node.group.member_count}`}
          </Badge>
        )}
        <div className="flex shrink-0 items-center gap-0.5">
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label="Add position under this one"
            onClick={() => onAddChild(node.position.id)}
          >
            <Plus className="size-3.5" />
          </Button>
          <Button variant="ghost" size="icon-sm" aria-label="Edit" onClick={() => onEdit(node)}>
            <Pencil className="size-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label="Delete"
            className="text-destructive"
            onClick={() => onDelete(node)}
          >
            <Trash2 className="size-3.5" />
          </Button>
        </div>
      </li>
      {node.children.map((child) => (
        <PositionRow
          key={child.position.id}
          node={child}
          onAddChild={onAddChild}
          onEdit={onEdit}
          onDelete={onDelete}
        />
      ))}
    </>
  )
}

function Editor() {
  const chartQuery = useOrgChart()
  const usersQuery = useUsers()
  const groupsQuery = useGroups()

  const create = useCreateOrgPosition()
  const update = useUpdateOrgPosition()
  const remove = useDeleteOrgPosition()

  const [form, setForm] = useState<FormState | null>(null)
  const [pendingDelete, setPendingDelete] = useState<OrgTreeNode | null>(null)

  const chart = useMemo(() => chartQuery.data ?? [], [chartQuery.data])
  const tree = useMemo(() => buildOrgTree(chart), [chart])

  const peopleOptions = useMemo(
    () =>
      [...(usersQuery.data ?? [])]
        .sort((a, b) => userName(a).localeCompare(userName(b)))
        .map((member) => ({
          value: member.entity_id,
          label: userName(member),
          hint: member.username ? `@${member.username}` : undefined,
        })),
    [usersQuery.data],
  )

  const groupOptions = useMemo(
    () =>
      [...(groupsQuery.data ?? [])]
        .sort((a, b) => a.name.localeCompare(b.name))
        .map((group) => ({
          value: group.id,
          label: group.name,
          hint: `${group.member_count} members`,
        })),
    [groupsQuery.data],
  )

  // A position can't be reparented under itself or its own descendants; the API
  // rejects it anyway, but the option shouldn't be offered in the first place.
  const parentOptions = useMemo(() => {
    const banned = new Set<string>()
    if (form?.id) {
      const collect = (node: OrgTreeNode) => {
        banned.add(node.position.id)
        node.children.forEach(collect)
      }
      const find = (nodes: OrgTreeNode[]): OrgTreeNode | undefined => {
        for (const node of nodes) {
          if (node.position.id === form.id) return node
          const hit = find(node.children)
          if (hit) return hit
        }
        return undefined
      }
      const self = find(tree)
      if (self) collect(self)
    }
    return chart
      .filter((node) => !banned.has(node.position.id))
      .map((node) => ({ value: node.position.id, label: node.position.title }))
      .sort((a, b) => a.label.localeCompare(b.label))
  }, [chart, tree, form?.id])

  const submit = () => {
    if (!form) return
    if (!form.title.trim()) {
      toast.error("Title is required")
      return
    }
    const payload: OrgPositionInput = {
      title: form.title.trim(),
      parent_id: form.parent_id,
      entity_id: form.entity_id,
      group_id: form.group_id,
      rank: Number.isFinite(form.rank) ? form.rank : 0,
    }
    const action = form.id
      ? update.mutateAsync({ id: form.id, ...payload })
      : create.mutateAsync(payload)
    action
      .then(() => {
        toast.success(form.id ? "Position updated" : "Position created")
        setForm(null)
      })
      .catch((error) => toast.error(apiError(error, "Failed to save position")))
  }

  const confirmDelete = () => {
    if (!pendingDelete) return
    remove
      .mutateAsync(pendingDelete.position.id)
      .then(() => {
        toast.success("Position deleted")
        setPendingDelete(null)
      })
      .catch((error) => toast.error(apiError(error, "Failed to delete position")))
  }

  return (
    <>
      <div className="mb-4 flex items-center justify-between gap-3">
        <p className="text-sm text-muted-foreground">
          {chart.length} position{chart.length === 1 ? "" : "s"}
        </p>
        <Button onClick={() => setForm({ ...EMPTY_FORM })}>
          <Plus className="size-3.5" />
          New position
        </Button>
      </div>

      <Card className="overflow-hidden p-0">
        <CardContent className="p-0">
          {chartQuery.isLoading ? (
            <div className="space-y-2 p-6">
              <Skeleton className="h-8 w-full" />
              <Skeleton className="h-8 w-3/4" />
            </div>
          ) : tree.length === 0 ? (
            <div className="px-6 py-12 text-center">
              <Network className="mx-auto mb-3 size-6 text-muted-foreground" />
              <p className="text-sm font-medium">No positions yet</p>
              <p className="mx-auto mt-1 max-w-sm text-xs text-muted-foreground">
                Start with a root position like President, then add the directors beneath it.
                Attach a group to a position and its members render in the chart automatically.
              </p>
            </div>
          ) : (
            <ul className="divide-y divide-border">
              {tree.map((node) => (
                <PositionRow
                  key={node.position.id}
                  node={node}
                  onAddChild={(parentID) => setForm({ ...EMPTY_FORM, parent_id: parentID })}
                  onEdit={(target) =>
                    setForm({
                      id: target.position.id,
                      title: target.position.title,
                      parent_id: target.position.parent_id,
                      entity_id: target.position.entity_id,
                      group_id: target.position.group_id,
                      rank: target.position.rank,
                    })
                  }
                  onDelete={setPendingDelete}
                />
              ))}
            </ul>
          )}
        </CardContent>
      </Card>

      <Dialog open={!!form} onOpenChange={(open) => !open && setForm(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{form?.id ? "Edit position" : "New position"}</DialogTitle>
            <DialogDescription>
              A position with a group attached renders that group's live membership beneath it.
            </DialogDescription>
          </DialogHeader>

          {form && (
            <div className="space-y-4">
              <div className="space-y-1.5">
                <Label htmlFor="org-title">Title</Label>
                <Input
                  id="org-title"
                  value={form.title}
                  onChange={(e) => setForm({ ...form, title: e.target.value })}
                  placeholder="Electrical Technical Director"
                />
              </div>

              <Picker
                label="Reports to"
                placeholder="No parent (root position)"
                value={form.parent_id}
                options={parentOptions}
                onChange={(next) => setForm({ ...form, parent_id: next })}
              />

              <Picker
                label="Held by"
                placeholder="Vacant"
                value={form.entity_id}
                options={peopleOptions}
                onChange={(next) => setForm({ ...form, entity_id: next })}
              />

              <Picker
                label="Team"
                placeholder="No group attached"
                value={form.group_id}
                options={groupOptions}
                onChange={(next) => setForm({ ...form, group_id: next })}
              />

              <div className="space-y-1.5">
                <Label htmlFor="org-rank">Order among siblings</Label>
                <Input
                  id="org-rank"
                  type="number"
                  value={form.rank}
                  onChange={(e) => setForm({ ...form, rank: Number(e.target.value) })}
                />
              </div>
            </div>
          )}

          <DialogFooter>
            <Button variant="ghost" onClick={() => setForm(null)}>
              Cancel
            </Button>
            <Button onClick={submit} disabled={create.isPending || update.isPending}>
              {form?.id ? "Save" : "Create"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!pendingDelete} onOpenChange={(open) => !open && setPendingDelete(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete {pendingDelete?.position.title}?</DialogTitle>
            <DialogDescription>
              {pendingDelete && pendingDelete.children.length > 0
                ? `Its ${pendingDelete.children.length} direct report${
                    pendingDelete.children.length === 1 ? "" : "s"
                  } will move up to report to its parent instead.`
                : "This position has no reports beneath it."}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setPendingDelete(null)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={confirmDelete} disabled={remove.isPending}>
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

export default function OrgChartEditorPage() {
  const { isAdmin, isLoading } = useAdmins()

  return (
    <PageContainer>
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to="/settings">
          <ArrowLeft className="size-4" />
          Settings
        </Link>
      </Button>
      <PageHeader
        title="Org chart"
        description="The team's reporting structure. Everyone can view it from the Members page."
      />

      {isLoading ? (
        <div className="space-y-4">
          <Skeleton className="h-10 rounded-lg" />
          <Skeleton className="h-64 rounded-lg" />
        </div>
      ) : isAdmin ? (
        <Editor />
      ) : (
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Network className="size-4 text-muted-foreground" />
              <CardTitle>Restricted</CardTitle>
            </div>
            <CardDescription>Editing the org chart is available to admins.</CardDescription>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            You can still view the chart from the{" "}
            <Link to="/members?view=org" className="underline">
              Members page
            </Link>
            .
          </CardContent>
        </Card>
      )}
    </PageContainer>
  )
}
