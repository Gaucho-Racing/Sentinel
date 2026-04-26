import { useId } from "react"

/**
 * Animated success checkmark — circle draws first, then the check stroke.
 * Stroke uses the gr-pink -> gr-purple brand gradient.
 * Sized via the className passed in.
 */
export function SuccessCheck({ className }: { className?: string }) {
  const gradientId = `success-check-${useId().replace(/:/g, "")}`
  return (
    <svg viewBox="0 0 52 52" className={className} aria-hidden>
      <defs>
        <linearGradient id={gradientId} x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stopColor="var(--color-gr-pink)" />
          <stop offset="100%" stopColor="var(--color-gr-purple)" />
        </linearGradient>
      </defs>
      <circle
        cx="26"
        cy="26"
        r="24"
        fill="none"
        stroke={`url(#${gradientId})`}
        strokeWidth="2"
        style={{
          strokeDasharray: 151,
          strokeDashoffset: 151,
          animation: "draw-stroke 400ms ease-out forwards",
        }}
      />
      <path
        d="M14 27 L22 35 L38 18"
        fill="none"
        stroke={`url(#${gradientId})`}
        strokeWidth="3"
        strokeLinecap="round"
        strokeLinejoin="round"
        style={{
          strokeDasharray: 40,
          strokeDashoffset: 40,
          animation: "draw-stroke 250ms ease-out 350ms forwards",
        }}
      />
    </svg>
  )
}
