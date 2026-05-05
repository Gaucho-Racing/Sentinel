import type { OnboardingData } from "@/pages/onboarding/types"

const GENDER_LABELS: Record<string, string> = {
  male: "Male",
  female: "Female",
  non_binary: "Non-binary",
  other: "Other",
  prefer_not_to_say: "Prefer not to say",
}

const LEVEL_LABELS: Record<string, string> = {
  undergraduate: "Undergraduate",
  graduate: "Graduate",
  phd: "PhD",
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-baseline justify-between gap-4 py-2">
      <span className="text-xs uppercase tracking-wider text-muted-foreground">{label}</span>
      <span className="truncate text-sm">{value || "—"}</span>
    </div>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <p className="mb-1 text-xs uppercase tracking-wider text-muted-foreground">{title}</p>
      <div className="divide-y divide-border/60 rounded-lg border border-border/60">
        <div className="px-4 py-1">{children}</div>
      </div>
    </div>
  )
}

export function ReviewStep({ data }: { data: OnboardingData }) {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-lg font-semibold tracking-tight">Review</h2>
        <p className="text-sm text-muted-foreground">
          Take one last look — you can update most of this later from settings.
        </p>
      </div>

      <div className="space-y-4">
        <Section title="Account">
          <Row label="Email" value={data.email} />
          <Row label="Username" value={data.username} />
        </Section>

        <Section title="Identity">
          <Row label="Name" value={`${data.firstName} ${data.lastName}`.trim()} />
          <Row label="Birthday" value={data.birthday} />
          <Row label="Gender" value={GENDER_LABELS[data.gender] ?? ""} />
          <Row label="Phone" value={data.phoneNumber} />
        </Section>

        <Section title="Academic">
          <Row label="Level" value={LEVEL_LABELS[data.graduateLevel] ?? ""} />
          <Row label="Grad year" value={data.graduationYear} />
          <Row label="Major" value={data.major} />
        </Section>

        <Section title="Team">
          <Row label="Shirt size" value={data.shirtSize} />
          <Row label="Jacket size" value={data.jacketSize} />
          <Row label="SAE #" value={data.saeRegistrationNumber} />
        </Section>
      </div>
    </div>
  )
}
