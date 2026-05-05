export type OnboardingData = {
  email: string
  password: string
  passwordConfirm: string
  firstName: string
  lastName: string
  username: string
  gender: string
  birthday: string
  phoneNumber: string
  graduateLevel: string
  graduationYear: string
  major: string
  shirtSize: string
  jacketSize: string
  saeRegistrationNumber: string
}

export type DiscordIdentity = {
  username: string
  globalName: string
  avatarUrl?: string
}

export const EMPTY_ONBOARDING_DATA: OnboardingData = {
  email: "",
  password: "",
  passwordConfirm: "",
  firstName: "",
  lastName: "",
  username: "",
  gender: "",
  birthday: "",
  phoneNumber: "",
  graduateLevel: "",
  graduationYear: "",
  major: "",
  shirtSize: "",
  jacketSize: "",
  saeRegistrationNumber: "",
}

export type StepProps = {
  data: OnboardingData
  update: (patch: Partial<OnboardingData>) => void
}
