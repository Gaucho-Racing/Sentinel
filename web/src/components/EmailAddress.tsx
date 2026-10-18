import { BadgeCheck } from "lucide-react"

import { cn } from "@/lib/utils"

export function VerifiedValue({
  value,
  verified,
  label,
  className,
}: {
  value: string
  verified: boolean
  label: string
  className?: string
}) {
  return (
    <span className={cn("inline-flex min-w-0 items-center gap-1", className)}>
      <span className="break-all">{value}</span>
      {verified && (
        <span title="Verified" role="img" aria-label={`Verified ${label}`} className="shrink-0">
          <BadgeCheck className="size-3.5 text-emerald-500" />
        </span>
      )}
    </span>
  )
}

export function EmailAddress({ email, verified, className }: { email: string; verified: boolean; className?: string }) {
  return <VerifiedValue value={email} verified={verified} label="email" className={className} />
}
