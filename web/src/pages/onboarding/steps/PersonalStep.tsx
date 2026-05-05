import { PhoneInput } from "react-international-phone"
import "react-international-phone/style.css"

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

const PREFERRED_COUNTRIES = ["us", "cn", "in", "kr", "tw", "mx", "ca"]

const PHONE_INPUT_STYLE = {
  "--react-international-phone-height": "2rem",
  "--react-international-phone-font-size": "0.875rem",
  "--react-international-phone-background-color": "transparent",
  "--react-international-phone-text-color": "var(--foreground)",
  "--react-international-phone-border-color": "var(--border)",
  "--react-international-phone-border-radius": "0.5rem",
  "--react-international-phone-country-selector-background-color": "transparent",
  "--react-international-phone-country-selector-background-color-hover": "var(--muted)",
  "--react-international-phone-dropdown-item-background-color": "var(--popover)",
  "--react-international-phone-dropdown-item-background-color-hover": "var(--muted)",
  "--react-international-phone-dropdown-item-text-color": "var(--foreground)",
  "--react-international-phone-dropdown-item-dial-code-color": "var(--muted-foreground)",
  "--react-international-phone-selected-dropdown-item-background-color": "var(--muted)",
  "--react-international-phone-disabled-background-color": "transparent",
} as React.CSSProperties

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
          <PhoneInput
            inputProps={{ id: "phoneNumber", autoComplete: "tel", required: true }}
            defaultCountry="us"
            preferredCountries={PREFERRED_COUNTRIES}
            value={data.phoneNumber}
            onChange={(phone) => update({ phoneNumber: phone })}
            style={PHONE_INPUT_STYLE}
            inputClassName="!w-full !flex-1 !text-foreground placeholder:!text-muted-foreground"
            countrySelectorStyleProps={{
              buttonClassName: "!border-input",
            }}
          />
        </div>
      </div>
    </div>
  )
}
