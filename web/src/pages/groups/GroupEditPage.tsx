import { useQuery, useQueryClient } from "@tanstack/react-query"
import { ArrowLeft, Trash2 } from "lucide-react"
import { useEffect, useState } from "react"
import { Link, useNavigate, useParams } from "react-router-dom"
import { toast } from "sonner"

import { PageContainer } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import type { Group } from "@/lib/groups"

import { GroupForm, type GroupFormValues } from "./GroupForm"

export default function GroupEditPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const qc = useQueryClient()

  const query = useQuery({
    queryKey: ["group", id],
    queryFn: async () => {
      const res = await api.get<Group>(`/groups/${id}`)
      return res.data
    },
    enabled: !!id,
  })

  const [values, setValues] = useState<GroupFormValues | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [deleting, setDeleting] = useState(false)

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
    const confirmed = window.confirm(
      `Delete "${query.data?.name ?? "this group"}"? Members and join requests will be removed too.`,
    )
    if (!confirmed) return
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
    }
  }

  if (query.isLoading || !values) {
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

        <Card>
          <CardHeader>
            <CardTitle>Danger zone</CardTitle>
            <CardDescription>
              Deleting a group removes all member, owner, and join-request rows linked to it. Linked application bindings stay (the apps just stop granting access via this group).
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button variant="destructive" disabled={deleting} onClick={handleDelete}>
              <Trash2 className="mr-1 size-3.5" />
              Delete group
            </Button>
          </CardContent>
        </Card>
      </div>
    </PageContainer>
  )
}
