import { useQuery, useQueryClient } from "@tanstack/react-query"
import { ArrowLeft } from "lucide-react"
import { useEffect, useState } from "react"
import { Link, useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import { DiscordIcon, GithubIcon } from "@/components/icons/socials"
import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
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
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4 text-muted-foreground">
        <Link to="/settings">
          <ArrowLeft className="mr-1 size-3.5" />
          Back to settings
        </Link>
      </Button>
      <PageHeader title="Connected accounts" description="Manage the external accounts connected to Sentinel." />
      {isLoading ? <Skeleton className="h-48" /> : !user ? (
        <p className="text-sm text-muted-foreground">Couldn't load your accounts. Refresh the page to try again.</p>
      ) : (
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <DiscordIcon className="size-5 text-discord-blurple" />
                Discord
              </CardTitle>
            </CardHeader>
            <CardContent className="text-sm">
              {discord ? (
                <div className="space-y-1">
                  <p>Username: <span className="font-medium">{discord.metadata?.username || "Unknown"}</span></p>
                  <p>Discord ID: <code className="font-mono text-xs">{discord.external_id}</code></p>
                </div>
              ) : (
                <p className="text-muted-foreground">No Discord account connected.</p>
              )}
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <span className="flex size-7 items-center justify-center rounded-full bg-black">
                  <GithubIcon className="size-5 text-white" />
                </span>
                GitHub
              </CardTitle>
            </CardHeader>
            <CardContent>
              {github ? <div className="space-y-3">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div className="space-y-1 text-sm">
                    <p>Username: <span className="font-medium">{githubStatus.data?.username ?? github.metadata?.username ?? "Unknown"}</span></p>
                    <p>GitHub ID: <code className="font-mono text-xs">{github.external_id}</code></p>
                  </div>
                  <Button type="button" variant="outline" onClick={() => setUnlinkOpen(true)}>Unlink GitHub</Button>
                </div>
                {githubStatus.data?.status === "pending" && (
                  <p className="text-sm text-amber-700 dark:text-amber-400">Your Gaucho Racing GitHub invitation is pending. <a className="font-medium underline underline-offset-2 hover:text-amber-900 dark:hover:text-amber-300" href="https://github.com/orgs/gaucho-racing/invitation" target="_blank" rel="noreferrer">Accept your invitation</a>.</p>
                )}
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
