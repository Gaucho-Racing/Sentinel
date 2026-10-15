import { BadgeCheck } from "lucide-react"

import { cn } from "@/lib/utils"

export function EmailAddress({ email, verified, className }: { email: string; verified: boolean; className?: string }) {
  return (
    <span className={cn("inline-flex min-w-0 items-center gap-1", className)}>
      <span className="break-all">{email}</span>
      {verified && (
        <span title="Verified" role="img" aria-label="Verified email" className="shrink-0">
          <BadgeCheck className="size-3.5 text-emerald-500" />
        </span>
      )}
    </span>
  )
}
