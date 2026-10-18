import { PhoneInput } from "react-international-phone"
import "react-international-phone/style.css"

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

export function PhoneNumberInput({
  id,
  value,
  onChange,
  required,
  autoFocus,
}: {
  id: string
  value: string
  onChange: (phone: string) => void
  required?: boolean
  autoFocus?: boolean
}) {
  return (
    <PhoneInput
      inputProps={{ id, autoComplete: "tel", required, autoFocus }}
      defaultCountry="us"
      preferredCountries={PREFERRED_COUNTRIES}
      value={value}
      onChange={onChange}
      style={PHONE_INPUT_STYLE}
      inputClassName="!w-full !flex-1 !text-foreground placeholder:!text-muted-foreground"
      countrySelectorStyleProps={{
        buttonClassName: "!border-input",
      }}
    />
  )
}
