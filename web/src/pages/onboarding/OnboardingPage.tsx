import { GraduationCap, Loader2 } from "lucide-react"
import { AnimatePresence, motion, type Variants } from "motion/react"
import { useEffect, useMemo, useState } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import { OutlineButton } from "@/components/OutlineButton"
import { SuccessCheck } from "@/components/SuccessCheck"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { api } from "@/lib/api"
import { validatePassword } from "@/lib/password"
import { cn } from "@/lib/utils"
import { OnboardingProgress } from "@/pages/onboarding/OnboardingProgress"
import { AcademicStep } from "@/pages/onboarding/steps/AcademicStep"
import { CredentialsStep } from "@/pages/onboarding/steps/CredentialsStep"
import { IdentityStep, type UsernameAvailability } from "@/pages/onboarding/steps/IdentityStep"
import { PersonalStep } from "@/pages/onboarding/steps/PersonalStep"
import { ReviewStep } from "@/pages/onboarding/steps/ReviewStep"
import { TeamStep } from "@/pages/onboarding/steps/TeamStep"
import { WelcomeStep } from "@/pages/onboarding/steps/WelcomeStep"
import {
  type DiscordIdentity,
  EMPTY_ONBOARDING_DATA,
  type OnboardingData,
} from "@/pages/onboarding/types"

const MOCK_LATENCY_MS = 1500
const CONVERGE_MS = 250
const CHECKMARK_DRAW_MS = 650
const HOLD_MS = 250

const STEP_SLIDE_PX = 24

const STUDENT_DOMAIN = "ucsb.edu"
const NON_STUDENT_ROLES = ["Alumni", "Mentor", "Sponsor", "Other"]

const stepVariants: Variants = {
  enter: (dir: "forward" | "back") => ({
    x: dir === "forward" ? STEP_SLIDE_PX : -STEP_SLIDE_PX,
    opacity: 0,
  }),
  center: { x: 0, opacity: 1 },
  exit: (dir: "forward" | "back") => ({
    x: dir === "forward" ? -STEP_SLIDE_PX : STEP_SLIDE_PX,
    opacity: 0,
  }),
}

type TokenState =
  | { status: "loading" }
  | { status: "ready"; identity: DiscordIdentity }
  | { status: "not_found" }
  | { status: "expired" }
  | { status: "error" }

type TokenInfoResponse = {
  discord_id: string
  discord_username: string
  discord_global_name: string
  discord_avatar_url: string
}

type StepId = "welcome" | "credentials" | "identity" | "personal" | "academic" | "team" | "review"

const STEPS: StepId[] = [
  "welcome",
  "credentials",
  "identity",
  "personal",
  "academic",
  "team",
  "review",
]

function isStepValid(step: StepId, data: OnboardingData): boolean {
  switch (step) {
    case "welcome":
      return true
    case "credentials":
      return (
        data.email.trim().length > 0 &&
        data.password.length > 0 &&
        data.password === data.passwordConfirm
      )
    case "identity":
      return (
        data.firstName.trim().length > 0 &&
        data.lastName.trim().length > 0 &&
        data.username.trim().length > 0
      )
    case "personal":
      return (
        data.gender.length > 0 && data.birthday.length > 0 && data.phoneNumber.trim().length > 0
      )
    case "academic":
      if (data.graduateLevel === "none") return true
      return (
        data.graduateLevel.length > 0 &&
        data.graduationYear.length > 0 &&
        data.major.trim().length > 0
      )
    case "team":
      return data.shirtSize.length > 0 && data.jacketSize.length > 0
    case "review":
      return true
  }
}

export default function OnboardingPage() {
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const token = params.get("token")

  const [stepIndex, setStepIndex] = useState(0)
  const [direction, setDirection] = useState<"forward" | "back">("forward")
  const [data, setData] = useState<OnboardingData>(EMPTY_ONBOARDING_DATA)
  const [submitting, setSubmitting] = useState(false)
  const [transitioning, setTransitioning] = useState(false)
  const [confirmedNonStudentDomain, setConfirmedNonStudentDomain] = useState<
    string | null
  >(null)
  const [nonStudentRole, setNonStudentRole] = useState<string | null>(null)
  const [studentDialogOpen, setStudentDialogOpen] = useState(false)
  const [tokenState, setTokenState] = useState<TokenState>({ status: "loading" })
  const [usernameAvailability, setUsernameAvailability] = useState<UsernameAvailability>("idle")

  useEffect(() => {
    if (!token) return
    let cancelled = false
    api
      .get<TokenInfoResponse>(`/discord/onboarding-tokens/${token}`)
      .then((res) => {
        if (cancelled) return
        setTokenState({
          status: "ready",
          identity: {
            username: res.data.discord_username,
            globalName: res.data.discord_global_name,
            avatarUrl: res.data.discord_avatar_url || undefined,
          },
        })
      })
      .catch((err) => {
        if (cancelled) return
        const status = err.response?.status
        if (status === 404) setTokenState({ status: "not_found" })
        else if (status === 410) setTokenState({ status: "expired" })
        else setTokenState({ status: "error" })
      })
    return () => {
      cancelled = true
    }
  }, [token])

  const currentStep = STEPS[stepIndex]
  const isLast = stepIndex === STEPS.length - 1
  const isFirst = stepIndex === 0
  const canProceed = useMemo(() => isStepValid(currentStep, data), [currentStep, data])

  const emailDomain = data.email.split("@")[1]?.toLowerCase() ?? ""
  const needsStudentConfirm =
    currentStep === "credentials" &&
    emailDomain.includes(".") &&
    emailDomain !== STUDENT_DOMAIN &&
    confirmedNonStudentDomain !== emailDomain

  function update(patch: Partial<OnboardingData>) {
    setData((prev) => ({ ...prev, ...patch }))
  }

  function advance() {
    setDirection("forward")
    setStepIndex((i) => i + 1)
  }

  function handleConfirmNonStudent(role: string) {
    setNonStudentRole(role)
    setConfirmedNonStudentDomain(emailDomain)
    setStudentDialogOpen(false)
    advance()
  }

  async function handleNext() {
    if (!canProceed || submitting) return

    if (currentStep === "credentials") {
      const passwordError = validatePassword(data.password)
      if (passwordError) {
        toast.error(passwordError)
        return
      }
    }

    if (currentStep === "identity" && usernameAvailability === "taken") {
      toast.error("Username is already taken")
      return
    }

    if (needsStudentConfirm) {
      setStudentDialogOpen(true)
      return
    }

    if (!isLast) {
      advance()
      return
    }

    setSubmitting(true)
    await new Promise((resolve) => setTimeout(resolve, MOCK_LATENCY_MS))
    setSubmitting(false)
    setTransitioning(true)
    await new Promise((resolve) =>
      setTimeout(resolve, CONVERGE_MS + CHECKMARK_DRAW_MS + HOLD_MS),
    )
    if (document.startViewTransition) {
      document.startViewTransition(() => navigate("/"))
    } else {
      navigate("/")
    }
  }

  function handleBack() {
    if (isFirst || submitting) return
    setDirection("back")
    setStepIndex((i) => i - 1)
  }

  if (!token) return <InviteError kind="not_found" />
  if (tokenState.status === "loading") {
    return (
      <main className="flex min-h-svh items-center justify-center px-4 py-12">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </main>
    )
  }
  if (tokenState.status !== "ready") return <InviteError kind={tokenState.status} />

  return (
    <main className="relative flex min-h-svh items-center justify-center px-4 py-12">
      <div
        className={cn(
          "w-full max-w-md space-y-8 transition-all ease-in",
          transitioning
            ? "scale-0 opacity-0 duration-[250ms]"
            : "scale-100 opacity-100 duration-200",
        )}
      >
        <div className="flex flex-col items-center gap-3 text-center">
          <img src="/logo/gr-logo-blank.png" alt="Gaucho Racing" className="size-12" />
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Set up your account</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              Step {stepIndex + 1} of {STEPS.length}
            </p>
          </div>
        </div>

        <OnboardingProgress step={stepIndex} total={STEPS.length} />

        <AnimatePresence mode="wait" custom={direction} initial={false}>
          <motion.div
            key={currentStep}
            custom={direction}
            variants={stepVariants}
            initial="enter"
            animate="center"
            exit="exit"
            transition={{ duration: 0.2, ease: "easeOut" }}
          >
            {currentStep === "welcome" && <WelcomeStep identity={tokenState.identity} />}
            {currentStep === "credentials" && (
              <CredentialsStep data={data} update={update} />
            )}
            {currentStep === "identity" && (
              <IdentityStep
                data={data}
                update={update}
                onAvailabilityChange={setUsernameAvailability}
              />
            )}
            {currentStep === "personal" && <PersonalStep data={data} update={update} />}
            {currentStep === "academic" && (
              <AcademicStep
                data={data}
                update={update}
                nonStudentRole={
                  confirmedNonStudentDomain === emailDomain &&
                  nonStudentRole &&
                  nonStudentRole !== "Alumni"
                    ? nonStudentRole
                    : null
                }
              />
            )}
            {currentStep === "team" && <TeamStep data={data} update={update} />}
            {currentStep === "review" && <ReviewStep data={data} />}
          </motion.div>
        </AnimatePresence>

        <div className="flex gap-2">
          <Button
            type="button"
            variant="outline"
            className="h-10 flex-1 rounded-xl"
            disabled={isFirst || submitting}
            onClick={handleBack}
          >
            Back
          </Button>
          <OutlineButton
            type="button"
            className="flex-1"
            loading={submitting}
            disabled={!canProceed}
            onClick={handleNext}
          >
            {isLast ? "Create account" : "Continue"}
          </OutlineButton>
        </div>
      </div>

      {transitioning && (
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 flex items-center justify-center"
        >
          <SuccessCheck className="size-20" />
        </div>
      )}

      <Dialog open={studentDialogOpen} onOpenChange={setStudentDialogOpen}>
        <DialogContent showCloseButton={false} className="gap-5 sm:max-w-md">
          <DialogHeader className="gap-3">
            <div className="flex size-10 items-center justify-center rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple text-white">
              <GraduationCap className="size-5" />
            </div>
            <DialogTitle>Are you a current student?</DialogTitle>
            <DialogDescription>
              You're using{" "}
              <span className="font-mono text-foreground">@{emailDomain}</span>{" "}
              instead of @ucsb.edu. Current students should use their UCSB email so we
              can verify enrollment.
            </DialogDescription>
          </DialogHeader>

          <OutlineButton
            type="button"
            innerClassName="bg-popover"
            onClick={() => setStudentDialogOpen(false)}
          >
            I'll use my UCSB email
          </OutlineButton>

          <div className="space-y-2.5">
            <p className="text-sm text-foreground/80">
              Or, if you're not a student, I'm a/an:
            </p>
            <div className="flex flex-wrap gap-2">
              {NON_STUDENT_ROLES.map((role) => (
                <button
                  key={role}
                  type="button"
                  onClick={() => handleConfirmNonStudent(role)}
                  className="rounded-full border border-border bg-background px-3 py-1 text-sm text-foreground transition-colors hover:border-foreground/60 hover:bg-muted/40"
                >
                  {role}
                </button>
              ))}
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </main>
  )
}

const ERROR_COPY: Record<"not_found" | "expired" | "error", { title: string; body: React.ReactNode }> = {
  not_found: {
    title: "Invalid invite link",
    body: (
      <>
        We couldn't find that link. Run{" "}
        <code className="font-mono text-xs">!verify</code> in the team Discord to get a fresh
        one.
      </>
    ),
  },
  expired: {
    title: "Link expired",
    body: (
      <>
        This onboarding link has expired or already been used. Run{" "}
        <code className="font-mono text-xs">!verify</code> in the team Discord for a new one.
      </>
    ),
  },
  error: {
    title: "Something went wrong",
    body: "We couldn't load your invite. Try refreshing the page in a minute.",
  },
}

function InviteError({ kind }: { kind: "not_found" | "expired" | "error" }) {
  const { title, body } = ERROR_COPY[kind]
  return (
    <main className="flex min-h-svh items-center justify-center px-4 py-12">
      <div className="w-full max-w-sm space-y-2 text-center">
        <h1 className="text-xl font-semibold tracking-tight">{title}</h1>
        <p className="text-sm text-muted-foreground">{body}</p>
      </div>
    </main>
  )
}
