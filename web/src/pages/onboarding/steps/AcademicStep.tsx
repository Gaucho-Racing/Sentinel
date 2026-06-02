import { Info } from "lucide-react"

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

const STUDENT_LEVELS = [
  { value: "undergraduate", label: "Undergraduate" },
  { value: "graduate", label: "Graduate" },
  { value: "phd", label: "PhD" },
]

const NONE_OPTION = { value: "none", label: "N/A" }

type Props = StepProps & {
  nonStudentRole?: string | null
  isAlumni?: boolean
}

export function AcademicStep({ data, update, nonStudentRole, isAlumni }: Props) {
  const isNonStudent = data.graduateLevel === "none"
  const levels = nonStudentRole ? [...STUDENT_LEVELS, NONE_OPTION] : STUDENT_LEVELS

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-lg font-semibold tracking-tight">
          {isAlumni ? "Graduation info" : "Academic info"}
        </h2>
        <p className="text-sm text-muted-foreground">
          {isAlumni
            ? "Tell us about when you graduated from UCSB."
            : "Helps us track team eligibility and graduation timelines."}
        </p>
      </div>

      {nonStudentRole && (
        <div className="flex gap-3 rounded-md border border-gr-pink/40 bg-gr-pink/5 px-3 py-2.5 text-xs duration-300 animate-in fade-in slide-in-from-top-1">
          <Info className="mt-0.5 size-4 shrink-0 text-gr-pink" />
          <p className="leading-snug">
            <span className="text-foreground">
              Since you marked yourself as {nonStudentRole.toLowerCase()},
            </span>{" "}
            <span className="text-muted-foreground">
              feel free to set Level to N/A and leave the other fields blank.
            </span>
          </p>
        </div>
      )}

      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="graduateLevel">{isAlumni ? "Degree" : "Level"}</Label>
          <Select
            value={data.graduateLevel}
            onValueChange={(v) => update({ graduateLevel: v })}
          >
            <SelectTrigger id="graduateLevel" className="w-full">
              <SelectValue placeholder="Select" />
            </SelectTrigger>
            <SelectContent>
              {levels.map((g) => (
                <SelectItem key={g.value} value={g.value}>
                  {g.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <Label htmlFor="graduationYear">
            {isAlumni ? "Year graduated" : "Graduation year"}
          </Label>
          <Input
            id="graduationYear"
            type="number"
            inputMode="numeric"
            placeholder={isNonStudent ? "Optional" : isAlumni ? "2023" : "2027"}
            min={2020}
            max={2035}
            value={data.graduationYear}
            onChange={(e) => update({ graduationYear: e.target.value })}
            required={!isNonStudent}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="major">{isAlumni ? "Major studied" : "Major"}</Label>
          <Input
            id="major"
            placeholder={isNonStudent ? "Optional" : "Mechanical Engineering"}
            value={data.major}
            onChange={(e) => update({ major: e.target.value })}
            required={!isNonStudent}
          />
        </div>
      </div>
    </div>
  )
}
