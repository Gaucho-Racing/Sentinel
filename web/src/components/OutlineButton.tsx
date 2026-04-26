import { Loader2 } from "lucide-react"
import type { ButtonHTMLAttributes } from "react"

import { cn } from "@/lib/utils"

type OutlineButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  loading?: boolean
}

/**
 * Outline-style button with a gradient border (gr-pink -> gr-purple).
 * On hover the inner fill fades to reveal the full gradient.
 */
export function OutlineButton({
  className,
  children,
  loading = false,
  disabled,
  ...props
}: OutlineButtonProps) {
  return (
    <button
      {...props}
      disabled={disabled || loading}
      className={cn(
        "group relative inline-flex h-10 w-full items-center justify-center overflow-hidden rounded-xl bg-gradient-to-br from-gr-pink to-gr-purple p-0.5 text-sm font-medium text-white transition-opacity disabled:pointer-events-none disabled:opacity-50",
        className,
      )}
    >
      <span className="relative flex size-full items-center justify-center rounded-[10px] bg-background px-4 py-2 transition-colors group-hover:bg-transparent">
        {loading ? <Loader2 className="size-4 animate-spin" /> : children}
      </span>
    </button>
  )
}
