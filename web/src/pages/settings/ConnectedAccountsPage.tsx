import { useQuery, useQueryClient } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { Link, useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import { useAuth } from "@/lib/auth"

type GitHubLinkStatus = {
  status: "active" | "pending" | "not_invited"
  username: string
}

export default function ConnectedAccountsPage() {
  const { user, isLoading } = useAuth()
  const queryClient = useQueryClient()
  const [searchParams, setSearchParams] = useSearchParams()
  const linkedCallback = searchParams.get("github") === "linked"
  const [linking, setLinking] = useState(false)
  const [unlinking, setUnlinking] = useState(false)
  const [unlinkOpen, setUnlinkOpen] = useState(false)
  const github = user?.external_auths?.find((auth) => auth.provider === "GITHUB")
  const discord = user?.external_auths?.find((auth) => auth.provider === "DISCORD")
  const githubStatus = useQuery({
    queryKey: ["githubLinkStatus", user?.id, github?.external_id],
    queryFn: async () => (await api.get<GitHubLinkStatus>("/github/link/status")).data,
    enabled: !!github,
    refetchInterval: 30_000,
  })

  useEffect(() => {
    if (linkedCallback) {
      void queryClient.invalidateQueries({ queryKey: ["currentEntity"] })
      setSearchParams((current) => {
        const next = new URLSearchParams(current)
        next.delete("github")
        return next
      }, { replace: true })
    }
  }, [linkedCallback, queryClient, setSearchParams])

  async function linkGithub() {
    setLinking(true)
    try {
      const response = await api.post<{ url: string }>("/github/link")
      window.location.assign(response.data.url)
    } catch (error: unknown) {
      const message = (error as { response?: { data?: { error?: string } } })?.response?.data?.error ?? "Couldn't start GitHub linking."
      toast.error(message)
      setLinking(false)
    }
  }

  async function unlinkGithub() {
    if (!github || unlinking) return
    setUnlinking(true)
    try {
      await api.delete("/github/link")
      await queryClient.invalidateQueries({ queryKey: ["currentEntity"] })
      queryClient.removeQueries({ queryKey: ["githubLinkStatus", user?.id] })
      setUnlinkOpen(false)
      toast.success("GitHub account unlinked")
    } catch (error: unknown) {
      const message = (error as { response?: { data?: { error?: string } } })?.response?.data?.error ?? "Couldn't unlink your GitHub account."
      toast.error(message)
    } finally {
      setUnlinking(false)
    }
  }

  return (
    <PageContainer>
      <PageHeader title="Connected accounts" description="Manage the external accounts connected to Sentinel." />
      <Button asChild variant="ghost" className="mb-6"><Link to="/settings">Back to settings</Link></Button>
      {isLoading ? <Skeleton className="h-48" /> : !user ? (
        <p className="text-sm text-muted-foreground">Couldn't load your accounts. Refresh the page to try again.</p>
      ) : (
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Discord</CardTitle>
              <CardDescription>Your team Discord connection.</CardDescription>
            </CardHeader>
            <CardContent className="text-sm">
              {discord ? <p>Connected as <strong>{discord.metadata?.username ?? discord.external_id}</strong></p> : (
                <p className="text-muted-foreground">No Discord account connected.</p>
              )}
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>GitHub</CardTitle>
              <CardDescription>Link your GitHub account to manage Gaucho Racing organization access.</CardDescription>
            </CardHeader>
            <CardContent>
              {github ? <div className="space-y-3">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <p className="text-sm">Connected as <strong>{githubStatus.data?.username ?? github.metadata?.username ?? github.external_id}</strong></p>
                  <Button type="button" variant="outline" onClick={() => setUnlinkOpen(true)}>Unlink GitHub</Button>
                </div>
                {githubStatus.data?.status === "pending" && (
                  <p className="text-sm text-muted-foreground">Your Gaucho Racing GitHub invitation is pending. <a className="text-foreground underline" href="https://github.com/orgs/gaucho-racing/invitation" target="_blank" rel="noreferrer">Accept your invitation</a>.</p>
                )}
                {githubStatus.data?.status === "active" && <p className="text-sm text-muted-foreground">You’re a member of the Gaucho Racing GitHub organization.</p>}
                {githubStatus.data?.status === "not_invited" && <p className="text-sm text-muted-foreground">No active GitHub invitation. Organization access requires a GithubMembers or GithubAdmins group.</p>}
                {githubStatus.isError && <p className="text-sm text-destructive">Couldn’t load GitHub invitation status.</p>}
              </div> : (
                <Button type="button" disabled={linking} onClick={linkGithub}>{linking ? "Connecting…" : "Connect GitHub"}</Button>
              )}
            </CardContent>
          </Card>
        </div>
      )}
      <Dialog open={unlinkOpen} onOpenChange={(open) => { if (!unlinking) setUnlinkOpen(open) }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Unlink GitHub account?</DialogTitle>
            <DialogDescription>
              This removes the account link and may revoke access to the Gaucho Racing GitHub organization. You can link a different account afterward.
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" disabled={unlinking} onClick={() => setUnlinkOpen(false)}>Cancel</Button>
            <Button type="button" variant="destructive" disabled={unlinking} onClick={unlinkGithub}>{unlinking ? "Unlinking…" : "Unlink GitHub"}</Button>
          </div>
        </DialogContent>
      </Dialog>
    </PageContainer>
  )
}
