import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import type { StepProps } from "@/pages/onboarding/types"

const GRADUATE_LEVELS = [
  { value: "undergraduate", label: "Undergraduate" },
  { value: "graduate", label: "Graduate" },
  { value: "phd", label: "PhD" },
]

export function AcademicStep({ data, update }: StepProps) {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-lg font-semibold tracking-tight">Academic info</h2>
        <p className="text-sm text-muted-foreground">
          Helps us track team eligibility and graduation timelines.
        </p>
      </div>

      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="graduateLevel">Level</Label>
          <Select
            value={data.graduateLevel}
            onValueChange={(v) => update({ graduateLevel: v })}
          >
            <SelectTrigger id="graduateLevel" className="w-full">
              <SelectValue placeholder="Select" />
            </SelectTrigger>
            <SelectContent>
              {GRADUATE_LEVELS.map((g) => (
                <SelectItem key={g.value} value={g.value}>
                  {g.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <Label htmlFor="graduationYear">Graduation year</Label>
          <Input
            id="graduationYear"
            type="number"
            inputMode="numeric"
            placeholder="2027"
            min={2020}
            max={2035}
            value={data.graduationYear}
            onChange={(e) => update({ graduationYear: e.target.value })}
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="major">Major</Label>
          <Input
            id="major"
            placeholder="Mechanical Engineering"
            value={data.major}
            onChange={(e) => update({ major: e.target.value })}
            required
          />
        </div>
      </div>
    </div>
  )
}
