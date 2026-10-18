
import { PhoneNumberInput } from "@/components/PhoneNumberInput"
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

const GENDERS = [
  { value: "male", label: "Male" },
  { value: "female", label: "Female" },
  { value: "non_binary", label: "Non-binary" },
  { value: "other", label: "Other" },
  { value: "prefer_not_to_say", label: "Prefer not to say" },
]


export function PersonalStep({ data, update }: StepProps) {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-lg font-semibold tracking-tight">A bit about you</h2>
        <p className="text-sm text-muted-foreground">
          Used for team rosters, FSAE registration, and emergency contact.
        </p>
      </div>

      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="gender">Gender</Label>
          <Select value={data.gender} onValueChange={(v) => update({ gender: v })}>
            <SelectTrigger id="gender" className="w-full">
              <SelectValue placeholder="Select" />
            </SelectTrigger>
            <SelectContent>
              {GENDERS.map((g) => (
                <SelectItem key={g.value} value={g.value}>
                  {g.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <Label htmlFor="birthday">Birthday</Label>
          <Input
            id="birthday"
            type="date"
            value={data.birthday}
            onChange={(e) => update({ birthday: e.target.value })}
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="phoneNumber">Phone number</Label>
          <PhoneNumberInput
            id="phoneNumber"
            required
            value={data.phoneNumber}
            onChange={(phone) => update({ phoneNumber: phone })}
          />
        </div>
      </div>
    </div>
  )
}
