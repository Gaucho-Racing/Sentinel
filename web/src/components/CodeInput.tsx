import { useState, type RefObject } from "react"

import { cn } from "@/lib/utils"

export const CODE_LENGTH = 6

export function CodeInput({
  inputRef,
  value,
  onChange,
  disabled,
  invalid,
}: {
  inputRef: RefObject<HTMLInputElement | null>
  value: string
  onChange: (value: string) => void
  disabled: boolean
  invalid: boolean
}) {
  const [focused, setFocused] = useState(false)
  const active = Math.min(value.length, CODE_LENGTH - 1)

  return (
    <div className="relative" onClick={() => inputRef.current?.focus()}>
      <div className="flex justify-center gap-2 sm:gap-2.5">
        {Array.from({ length: CODE_LENGTH }, (_, index) => {
          const digit = value[index]
          const isActive = focused && !disabled && index === active
          return (
            <div
              key={index}
              className={cn(
                "flex h-16 max-w-[3.25rem] min-w-0 flex-1 items-center justify-center rounded-xl border bg-card font-mono text-3xl font-semibold transition-colors sm:h-[4.25rem] sm:text-4xl",
                invalid
                  ? "border-destructive/70"
                  : isActive
                    ? "border-gr-pink ring-2 ring-gr-pink/25"
                    : digit
                      ? "border-foreground/30"
                      : "border-border",
                disabled && "opacity-60",
              )}
            >
              {digit ?? (isActive && <span className="h-8 w-0.5 animate-pulse rounded-full bg-gr-pink" />)}
            </div>
          )
        })}
      </div>
      <input
        ref={inputRef}
        id="code"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        onFocus={() => setFocused(true)}
        onBlur={() => setFocused(false)}
        inputMode="numeric"
        autoComplete="one-time-code"
        pattern="[0-9]*"
        maxLength={CODE_LENGTH}
        disabled={disabled}
        autoFocus
        aria-invalid={invalid}
        className="absolute inset-0 h-full w-full cursor-pointer opacity-0"
      />
    </div>
  )
}
