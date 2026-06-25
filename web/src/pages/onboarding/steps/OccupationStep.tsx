import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import type { StepProps } from "@/pages/onboarding/types"

type Props = StepProps & {
  isAlumni?: boolean
}

export function OccupationStep({ data, update, isAlumni }: Props) {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-lg font-semibold tracking-tight">Current occupation</h2>
        <p className="text-sm text-muted-foreground">
          {isAlumni
            ? "Tell us what you're up to now — we love seeing where alumni land."
            : "Helps the team understand who's in our network."}
        </p>
      </div>

      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="occupationTitle">Job title</Label>
          <Input
            id="occupationTitle"
            placeholder="Software Engineer"
            value={data.occupationTitle}
            onChange={(e) => update({ occupationTitle: e.target.value })}
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="occupationCompany">Company</Label>
          <Input
            id="occupationCompany"
            placeholder="Anduril"
            value={data.occupationCompany}
            onChange={(e) => update({ occupationCompany: e.target.value })}
            required
          />
        </div>
      </div>
    </div>
  )
}
