import { Loader2 } from "lucide-react"
import type { ComponentType, SVGProps } from "react"
import { useState } from "react"
import { useNavigate } from "react-router-dom"

import { OutlineButton } from "@/components/OutlineButton"
import { SuccessCheck } from "@/components/SuccessCheck"
import { DiscordIcon, GoogleIcon } from "@/components/icons/socials"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { DISCORD_INVITE_URL } from "@/lib/links"
import { cn } from "@/lib/utils"

type ProviderId = "google" | "discord"
type LoadingTarget = "email" | ProviderId | null

const PROVIDERS: Array<{
  id: ProviderId
  label: string
  Icon: ComponentType<SVGProps<SVGSVGElement>>
  iconClassName?: string
}> = [
  { id: "google", label: "Continue with Google", Icon: GoogleIcon },
  {
    id: "discord",
    label: "Continue with Discord",
    Icon: DiscordIcon,
    iconClassName: "text-discord-blurple",
  },
]

// Mock latency for the design-stage flows. Replace with real API calls.
const MOCK_LATENCY_MS = 1500

// Convergence (login content collapses) -> checkmark draws -> hold -> navigate.
const CONVERGE_MS = 250
const CHECKMARK_DRAW_MS = 650 // circle (400) + check (250) starting at +350ms
const HOLD_MS = 250

export default function LoginPage() {
  const navigate = useNavigate()
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [loading, setLoading] = useState<LoadingTarget>(null)
  const [transitioning, setTransitioning] = useState(false)
  const isBusy = loading !== null || transitioning

  async function handleSuccess() {
    setLoading(null)
    setTransitioning(true)
    // Wait for convergence + checkmark draw + hold before navigating.
    await new Promise((resolve) =>
      setTimeout(resolve, CONVERGE_MS + CHECKMARK_DRAW_MS + HOLD_MS),
    )
    if (document.startViewTransition) {
      document.startViewTransition(() => navigate("/"))
    } else {
      navigate("/")
    }
  }

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault()
    if (isBusy) return
    setLoading("email")
    await new Promise((resolve) => setTimeout(resolve, MOCK_LATENCY_MS))
    handleSuccess()
  }

  async function handleProvider(id: ProviderId) {
    if (isBusy) return
    setLoading(id)
    await new Promise((resolve) => setTimeout(resolve, MOCK_LATENCY_MS))
    handleSuccess()
  }

  return (
    <main className="relative flex min-h-svh items-center justify-center px-4 py-12">
      <div
        className={cn(
          "w-full max-w-sm space-y-8 transition-all ease-in",
          transitioning
            ? "scale-0 opacity-0 duration-[250ms]"
            : "scale-100 opacity-100 duration-200",
        )}
      >
        <div className="flex flex-col items-center gap-3 text-center">
          <img src="/logo/gr-logo-blank.png" alt="Gaucho Racing" className="size-12" />
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Sign in to Sentinel</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              Choose how you'd like to sign in.
            </p>
          </div>
        </div>

        <div className="space-y-8">
          <form onSubmit={handleSubmit} noValidate className="space-y-2">
            <Label htmlFor="email" className="sr-only">Email</Label>
            <Input
              id="email"
              type="email"
              autoComplete="email"
              placeholder="Email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={isBusy}
              required
            />

            <Label htmlFor="password" className="sr-only">Password</Label>
            <Input
              id="password"
              type="password"
              autoComplete="current-password"
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={isBusy}
              required
            />

            <OutlineButton
              type="submit"
              className="mt-4"
              loading={loading === "email"}
              disabled={isBusy}
            >
              Sign in
            </OutlineButton>
          </form>

          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t border-border/60" />
            </div>
            <div className="relative flex justify-center">
              <span className="bg-background px-2 text-xs uppercase tracking-wider text-muted-foreground">
                or
              </span>
            </div>
          </div>

          <div className="space-y-2">
            {PROVIDERS.map(({ id, label, Icon, iconClassName }) => {
              const isLoading = loading === id
              return (
                <Button
                  key={id}
                  variant="outline"
                  className="w-full justify-center gap-2"
                  disabled={isBusy}
                  onClick={() => handleProvider(id)}
                >
                  {isLoading ? (
                    <Loader2 className="size-4 animate-spin" />
                  ) : (
                    <Icon className={`size-4 ${iconClassName ?? ""}`} />
                  )}
                  {label}
                </Button>
              )
            })}
          </div>

          <p className="text-center text-xs text-muted-foreground">
            Don't have an account? Reach out on the{" "}
            <a
              href={DISCORD_INVITE_URL}
              target="_blank"
              rel="noreferrer"
              className="text-foreground transition-colors hover:text-gr-pink"
            >
              Discord
            </a>
            .
          </p>
        </div>
      </div>

      {transitioning && (
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 flex items-center justify-center"
          style={{ animationDelay: `${CONVERGE_MS}ms` }}
        >
          <SuccessCheck className="size-20" />
        </div>
      )}
    </main>
  )
}
