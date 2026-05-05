import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { cn } from "@/lib/utils"
import type { OnboardingData, StepProps } from "@/pages/onboarding/types"

const SIZES = ["XS", "S", "M", "L", "XL", "XXL"]

type SizeField = "shirtSize" | "jacketSize"

function SizePicker({
  label,
  field,
  value,
  onChange,
}: {
  label: string
  field: SizeField
  value: string
  onChange: (patch: Partial<OnboardingData>) => void
}) {
  return (
    <div className="space-y-2">
      <Label>{label}</Label>
      <div className="grid grid-cols-6 gap-1.5" role="radiogroup" aria-label={label}>
        {SIZES.map((size) => {
          const selected = value === size
          return (
            <button
              key={size}
              type="button"
              role="radio"
              aria-checked={selected}
              onClick={() => onChange({ [field]: size })}
              className={cn(
                "h-9 rounded-md border text-xs font-medium transition-colors",
                selected
                  ? "border-foreground bg-foreground text-background"
                  : "border-border bg-background text-muted-foreground hover:border-foreground/40 hover:text-foreground",
              )}
            >
              {size}
            </button>
          )
        })}
      </div>
    </div>
  )
}

export function TeamStep({ data, update }: StepProps) {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-lg font-semibold tracking-tight">Team gear</h2>
        <p className="text-sm text-muted-foreground">
          Sizes for team apparel. SAE registration is optional and can be added later.
        </p>
      </div>

      <div className="space-y-4">
        <SizePicker
          label="Shirt size"
          field="shirtSize"
          value={data.shirtSize}
          onChange={update}
        />

        <SizePicker
          label="Jacket size"
          field="jacketSize"
          value={data.jacketSize}
          onChange={update}
        />

        <div className="space-y-2">
          <Label htmlFor="saeRegistrationNumber">SAE registration number</Label>
          <Input
            id="saeRegistrationNumber"
            placeholder="Optional"
            value={data.saeRegistrationNumber}
            onChange={(e) => update({ saeRegistrationNumber: e.target.value })}
          />
        </div>
      </div>
    </div>
  )
}
