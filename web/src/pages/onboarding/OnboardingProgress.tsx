import { motion } from "motion/react"

type OnboardingProgressProps = {
  step: number
  total: number
}

export function OnboardingProgress({ step, total }: OnboardingProgressProps) {
  return (
    <div className="flex items-center gap-1.5">
      {Array.from({ length: total }).map((_, i) => {
        const reached = i <= step
        const current = i === step
        return (
          <div
            key={i}
            className="relative h-1 flex-1 overflow-hidden rounded-full bg-muted"
          >
            <motion.div
              className="absolute inset-0 origin-left rounded-full bg-foreground"
              initial={false}
              animate={{ scaleX: reached ? 1 : 0 }}
              transition={{ duration: 0.35, ease: "easeOut" }}
            />
            <motion.div
              className="absolute inset-0 origin-left rounded-full bg-gradient-to-r from-gr-pink to-gr-purple"
              initial={false}
              animate={{
                scaleX: reached ? 1 : 0,
                opacity: current ? 1 : 0,
              }}
              transition={{
                scaleX: { duration: 0.35, ease: "easeOut" },
                opacity: { duration: 0.25, ease: "easeOut" },
              }}
            />
          </div>
        )
      })}
    </div>
  )
}
