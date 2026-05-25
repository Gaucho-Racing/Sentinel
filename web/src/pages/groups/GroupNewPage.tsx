import { useQueryClient } from "@tanstack/react-query"
import { ArrowLeft } from "lucide-react"
import { useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { toast } from "sonner"

import { PageContainer } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { api } from "@/lib/api"
import type { Group } from "@/lib/groups"

import { GroupForm, type GroupFormValues } from "./GroupForm"

const INITIAL: GroupFormValues = {
  name: "",
  description: "",
  allowed_sources: ["DIRECT"],
}

export default function GroupNewPage() {
  const navigate = useNavigate()
  const qc = useQueryClient()
  const [values, setValues] = useState<GroupFormValues>(INITIAL)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit() {
    setSubmitting(true)
    try {
      const res = await api.post<Group>("/groups", {
        name: values.name.trim(),
        description: values.description,
        allowed_sources: values.allowed_sources,
      })
      qc.invalidateQueries({ queryKey: ["groups"] })
      toast.success("Group created")
      navigate(`/groups/${res.data.id}`)
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't create group."
      toast.error(message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <PageContainer>
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to="/groups">
          <ArrowLeft className="mr-1 size-3.5" />
          All groups
        </Link>
      </Button>

      <div className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight">New group</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Set the name, description, and which sources are allowed to populate it.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Basics</CardTitle>
          <CardDescription>You can change all of these later.</CardDescription>
        </CardHeader>
        <CardContent>
          <GroupForm
            values={values}
            onChange={setValues}
            onSubmit={handleSubmit}
            submitting={submitting}
            submitLabel="Create group"
          />
        </CardContent>
      </Card>
    </PageContainer>
  )
}
