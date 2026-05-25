import { Check } from "lucide-react"

import { OutlineButton } from "@/components/OutlineButton"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { ALL_GROUP_SOURCES, SOURCE_LABEL, type GroupSource } from "@/lib/groups"

export type GroupFormValues = {
  name: string
  description: string
  allowed_sources: GroupSource[]
}

const SOURCE_HINT: Record<GroupSource, string> = {
  DIRECT: "Owners add members manually or approve join requests.",
  DISCORD: "Members with a linked Discord role are added automatically.",
  CONDITIONAL: "Auto-populated by a rule against entity profiles.",
}

export function GroupForm({
  values,
  onChange,
  onSubmit,
  submitting,
  submitLabel,
  innerSurface = "card",
}: {
  values: GroupFormValues
  onChange: (next: GroupFormValues) => void
  onSubmit: () => void
  submitting: boolean
  submitLabel: string
  innerSurface?: "card" | "background"
}) {
  function toggleSource(source: GroupSource) {
    const has = values.allowed_sources.includes(source)
    const next = has
      ? values.allowed_sources.filter((s) => s !== source)
      : [...values.allowed_sources, source]
    onChange({ ...values, allowed_sources: next })
  }

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        if (!submitting) onSubmit()
      }}
      className="space-y-5"
    >
      <div className="space-y-2">
        <Label htmlFor="name">Name</Label>
        <Input
          id="name"
          value={values.name}
          onChange={(e) => onChange({ ...values, name: e.target.value })}
          required
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="description">Description</Label>
        <Textarea
          id="description"
          value={values.description}
          onChange={(e) => onChange({ ...values, description: e.target.value })}
          rows={2}
        />
      </div>
      <div className="space-y-2">
        <Label>Allowed sources</Label>
        <p className="text-xs text-muted-foreground">
          Which mechanisms can add members to this group.
        </p>
        <div className="space-y-2 pt-1">
          {ALL_GROUP_SOURCES.map((source) => {
            const selected = values.allowed_sources.includes(source)
            return (
              <button
                key={source}
                type="button"
                onClick={() => toggleSource(source)}
                className={
                  "flex w-full items-start gap-3 rounded-md border p-3 text-left transition-colors " +
                  (selected
                    ? "border-gr-pink/40 bg-gr-pink/5 hover:bg-gr-pink/10"
                    : "border-border/60 bg-muted/30 hover:bg-muted/50")
                }
              >
                <div
                  className={
                    "mt-0.5 flex size-4 shrink-0 items-center justify-center rounded border " +
                    (selected
                      ? "border-gr-pink bg-gr-pink text-white"
                      : "border-border bg-background")
                  }
                >
                  {selected && <Check className="size-3" />}
                </div>
                <div className="min-w-0 flex-1 leading-tight">
                  <p className="font-mono text-sm">{SOURCE_LABEL[source]}</p>
                  <p className="mt-0.5 text-xs text-muted-foreground">
                    {SOURCE_HINT[source]}
                  </p>
                </div>
              </button>
            )
          })}
        </div>
      </div>
      <div className="flex justify-end pt-2">
        <OutlineButton
          type="submit"
          className="w-auto"
          innerClassName={innerSurface === "card" ? "bg-card" : undefined}
          loading={submitting}
          disabled={!values.name.trim()}
        >
          {submitLabel}
        </OutlineButton>
      </div>
    </form>
  )
}
