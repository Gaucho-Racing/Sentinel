import { Loader2 } from "lucide-react"
import type { ButtonHTMLAttributes } from "react"

import { cn } from "@/lib/utils"

type OutlineButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  loading?: boolean
  innerClassName?: string
  size?: "default" | "sm"
}

/**
 * Outline-style button with a gradient border (gr-pink -> gr-purple).
 * On hover the inner fill fades to reveal the full gradient.
 */
export function OutlineButton({
  className,
  innerClassName,
  children,
  loading = false,
  disabled,
  size = "default",
  ...props
}: OutlineButtonProps) {
  const outer =
    size === "sm"
      ? "h-7 rounded-lg p-px text-[0.8rem]"
      : "h-10 rounded-xl p-0.5 text-sm"
  const inner =
    size === "sm"
      ? "gap-1 rounded-[7px] px-2.5"
      : "rounded-[10px] px-4 py-2"
  return (
    <button
      {...props}
      disabled={disabled || loading}
      className={cn(
        "group relative inline-flex w-full items-center justify-center overflow-hidden bg-gradient-to-br from-gr-pink to-gr-purple font-medium text-white transition-opacity disabled:pointer-events-none disabled:opacity-50",
        outer,
        className,
      )}
    >
      <span
        className={cn(
          "relative flex size-full items-center justify-center bg-background transition-colors group-hover:bg-transparent",
          inner,
          innerClassName,
        )}
      >
        {loading ? <Loader2 className="size-4 animate-spin" /> : children}
      </span>
    </button>
  )
}
