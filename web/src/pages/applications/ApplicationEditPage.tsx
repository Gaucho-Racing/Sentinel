import { useQuery, useQueryClient } from "@tanstack/react-query"
import { ArrowLeft } from "lucide-react"
import { useEffect, useState } from "react"
import { Link, useNavigate, useParams } from "react-router-dom"
import { toast } from "sonner"

import { OutlineButton } from "@/components/OutlineButton"
import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Skeleton } from "@/components/ui/skeleton"
import { Textarea } from "@/components/ui/textarea"
import { api } from "@/lib/api"
import type { Application } from "@/lib/applications"

export default function ApplicationEditPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const qc = useQueryClient()

  const query = useQuery({
    queryKey: ["application", "id", id],
    queryFn: async () => {
      const res = await api.get<Application>(`/applications/${id}`)
      return res.data
    },
    enabled: !!id,
  })

  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [iconURL, setIconURL] = useState("")
  const [launchURL, setLaunchURL] = useState("")
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!query.data) return
    setName(query.data.name)
    setDescription(query.data.description)
    setIconURL(query.data.icon_url)
    setLaunchURL(query.data.launch_url)
  }, [query.data])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (submitting) return
    setSubmitting(true)
    try {
      await api.put(`/applications/${id}`, {
        name,
        description,
        icon_url: iconURL,
        launch_url: launchURL,
      })
      qc.invalidateQueries({ queryKey: ["application", "id", id] })
      qc.invalidateQueries({ queryKey: ["applications"] })
      toast.success("Application updated")
      navigate(`/applications/${id}`)
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't save the application."
      toast.error(message)
    } finally {
      setSubmitting(false)
    }
  }

  if (query.isLoading) {
    return (
      <PageContainer>
        <Skeleton className="mb-4 h-4 w-24" />
        <Skeleton className="mb-8 h-10 w-48" />
        <div className="max-w-xl space-y-5">
          <Skeleton className="h-9" />
          <Skeleton className="h-16" />
          <Skeleton className="h-9" />
          <Skeleton className="h-9" />
        </div>
      </PageContainer>
    )
  }

  if (query.isError || !query.data) {
    return (
      <PageContainer>
        <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
          <Link to="/applications">
            <ArrowLeft className="mr-1 size-3.5" />
            All applications
          </Link>
        </Button>
        <p className="text-sm text-muted-foreground">Application not found.</p>
      </PageContainer>
    )
  }

  return (
    <PageContainer>
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to={`/applications/${id}`}>
          <ArrowLeft className="mr-1 size-3.5" />
          Back to {query.data.name}
        </Link>
      </Button>

      <PageHeader title={`Edit ${query.data.name}`} description="Update application metadata." />

      <form onSubmit={handleSubmit} className="max-w-xl space-y-5">
        <div className="space-y-2">
          <Label htmlFor="name">Name</Label>
          <Input
            id="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="description">Description</Label>
          <Textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={2}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="launch_url">Launch URL</Label>
          <Input
            id="launch_url"
            type="url"
            value={launchURL}
            onChange={(e) => setLaunchURL(e.target.value)}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="icon_url">Icon URL</Label>
          <Input
            id="icon_url"
            type="url"
            value={iconURL}
            onChange={(e) => setIconURL(e.target.value)}
          />
        </div>

        <div className="flex gap-2 pt-2">
          <Button asChild variant="outline">
            <Link to={`/applications/${id}`}>Cancel</Link>
          </Button>
          <OutlineButton type="submit" className="w-auto" loading={submitting} disabled={!name}>
            Save changes
          </OutlineButton>
        </div>
      </form>
    </PageContainer>
  )
}
