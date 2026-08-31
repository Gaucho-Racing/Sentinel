import { useState } from "react"

import { cn } from "@/lib/utils"

export function ApplicationIcon({
  name,
  iconUrl,
  className,
  fallbackClassName,
}: {
  name: string
  iconUrl?: string
  className?: string
  fallbackClassName?: string
}) {
  const [failedUrl, setFailedUrl] = useState<string | null>(null)
  const imageUrl = iconUrl && failedUrl !== iconUrl ? iconUrl : null

  return (
    <div
      className={cn(
        "flex shrink-0 items-center justify-center overflow-hidden",
        className,
        !imageUrl &&
          "bg-gradient-to-br from-gr-pink to-gr-purple font-semibold text-white",
        !imageUrl && fallbackClassName,
      )}
    >
      {imageUrl ? (
        <img
          src={imageUrl}
          alt={name}
          className="size-full object-contain"
          onError={() => setFailedUrl(imageUrl)}
        />
      ) : (
        (name.slice(0, 1) || "?").toUpperCase()
      )}
    </div>
  )
}
