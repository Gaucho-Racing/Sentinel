import { useQueryClient } from "@tanstack/react-query"
import { Copy } from "lucide-react"
import { useState, type FormEvent, type ReactNode } from "react"
import { toast } from "sonner"

import { ChangeEmailDialog } from "@/components/ChangeEmailDialog"
import { ChangePasswordDialog } from "@/components/ChangePasswordDialog"
import { EmailAddress } from "@/components/EmailAddress"
import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import { useAuth, type Entity } from "@/lib/auth"

type Profile = NonNullable<Entity["user"]>

type ProfileValues = Pick<
  Profile,
  | "first_name"
  | "last_name"
  | "gender"
  | "graduate_level"
  | "graduation_year"
  | "major"
  | "shirt_size"
  | "jacket_size"
  | "sae_registration_number"
  | "occupation_title"
  | "occupation_company"
>

const GENDERS = [
  { value: "male", label: "Male" },
  { value: "female", label: "Female" },
  { value: "non_binary", label: "Non-binary" },
  { value: "other", label: "Other" },
  { value: "prefer_not_to_say", label: "Prefer not to say" },
]

const LEVELS = [
  { value: "undergraduate", label: "Undergraduate" },
  { value: "graduate", label: "Graduate" },
  { value: "phd", label: "PhD" },
  { value: "none", label: "N/A" },
]

const SIZES = ["XS", "S", "M", "L", "XL", "XXL"]

function initialValues(profile: Profile): ProfileValues {
  return {
    first_name: profile.first_name,
    last_name: profile.last_name,
    gender: profile.gender,
    graduate_level: profile.graduate_level,
    graduation_year: profile.graduation_year,
    major: profile.major,
    shirt_size: profile.shirt_size,
    jacket_size: profile.jacket_size,
    sae_registration_number: profile.sae_registration_number,
    occupation_title: profile.occupation_title,
    occupation_company: profile.occupation_company,
  }
}

function Section({ title, description, children }: { title: string; description: string; children: ReactNode }) {
  return (
    <section className="space-y-3">
      <div>
        <h2 className="text-base font-semibold">{title}</h2>
        <p className="mt-1 text-sm text-muted-foreground">{description}</p>
      </div>
      {children}
    </section>
  )
}

function FieldGroup({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="space-y-3">
      <h3 className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{title}</h3>
      <div className="grid gap-x-4 gap-y-3 sm:grid-cols-2">{children}</div>
    </div>
  )
}

function SettingRow({ label, children, action }: { label: string; children: ReactNode; action?: ReactNode }) {
  return (
    <div className="flex flex-col gap-1.5 rounded-lg px-3 py-1.5 sm:min-h-10 sm:flex-row sm:items-center sm:gap-4">
      <span className="w-28 shrink-0 text-sm text-muted-foreground">{label}</span>
      <div className="min-w-0 flex-1 text-sm">{children}</div>
      {action && <div className="shrink-0">{action}</div>}
    </div>
  )
}

function CopyableID({ value }: { value: string }) {
  async function copy() {
    try {
      await navigator.clipboard.writeText(value)
      toast.success("Copied to clipboard")
    } catch {
      toast.error("Couldn't copy to clipboard")
    }
  }
  return (
    <button
      type="button"
      onClick={copy}
      className="group inline-flex max-w-full items-center gap-1.5 font-mono text-xs text-muted-foreground transition-colors hover:text-foreground"
    >
      <span className="truncate">{value}</span>
      <Copy className="size-3 shrink-0 opacity-60 group-hover:opacity-100" />
    </button>
  )
}

function ProfileForm({ entity, profile }: { entity: Entity; profile: Profile }) {
  const queryClient = useQueryClient()
  const [values, setValues] = useState(() => initialValues(profile))
  const [saving, setSaving] = useState(false)
  const [changingEmail, setChangingEmail] = useState(false)
  const [changingPassword, setChangingPassword] = useState(false)
  const changed = JSON.stringify(values) !== JSON.stringify(initialValues(profile))
  const name = `${profile.first_name} ${profile.last_name}`.trim() || profile.username
  const email = entity.email_auth?.email ?? profile.email ?? ""
  const joined = new Date(profile.created_at).toLocaleDateString(undefined, { month: "long", year: "numeric" })

  function update(patch: Partial<ProfileValues>) {
    setValues((current) => ({ ...current, ...patch }))
  }

  async function handleSave(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (saving || !changed) return
    setSaving(true)
    try {
      const response = await api.patch<Profile>("/users/@me/profile", {
        ...values,
        first_name: values.first_name.trim(),
        last_name: values.last_name.trim(),
        major: values.major.trim(),
        sae_registration_number: values.sae_registration_number.trim(),
        occupation_title: values.occupation_title.trim(),
        occupation_company: values.occupation_company.trim(),
      })
      setValues(initialValues(response.data))
      queryClient.setQueryData<Entity>(["currentEntity", profile.entity_id], (current) =>
        current ? { ...current, user: response.data } : current,
      )
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["currentEntity"] }),
        queryClient.invalidateQueries({ queryKey: ["users"] }),
      ])
      toast.success("Profile updated")
    } catch (error: unknown) {
      const message =
        (error as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Couldn't save your profile. Try again."
      toast.error(message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-10">
      <Card>
        <CardContent className="flex flex-col gap-4 sm:flex-row sm:items-center">
          <Avatar className="size-16">
            <AvatarImage src={profile.avatar_url} alt={name} />
            <AvatarFallback className="text-lg">
              {name.split(/\s+/).map((part) => part[0]).slice(0, 2).join("").toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <div className="min-w-0 flex-1 space-y-1">
            <div className="flex flex-wrap items-center gap-2">
              <h2 className="truncate text-xl font-semibold tracking-tight">{name}</h2>
              {profile.initial_role && <Badge variant="secondary" className="capitalize">{profile.initial_role}</Badge>}
            </div>
            <p className="text-sm text-muted-foreground">
              @{profile.username} · Joined {joined}
            </p>
          </div>
        </CardContent>
      </Card>

      <Section title="Sign-in & security" description="How you sign in to Sentinel and every app that uses it.">
        <Card>
          <CardContent className="px-3">
            <SettingRow
              label="Email"
              action={
                <Button type="button" variant="outline" size="sm" onClick={() => setChangingEmail(true)}>
                  Change
                </Button>
              }
            >
              <EmailAddress email={email} verified={!!entity.email_auth?.verified_at} />
            </SettingRow>
            <SettingRow
              label="Password"
              action={
                <Button type="button" variant="outline" size="sm" onClick={() => setChangingPassword(true)}>
                  Change
                </Button>
              }
            >
              <span className="tracking-widest text-muted-foreground">••••••••••</span>
            </SettingRow>
            <SettingRow label="Username">@{profile.username}</SettingRow>
            <SettingRow label="User ID">
              <CopyableID value={profile.id} />
            </SettingRow>
            <SettingRow label="Entity ID">
              <CopyableID value={entity.id} />
            </SettingRow>
          </CardContent>
        </Card>
      </Section>

      <form onSubmit={handleSave}>
        <Section title="Profile" description="What teammates see on your team profile and in the member directory.">
          <Card>
            <CardContent className="space-y-6">
              <FieldGroup title="Personal">
                <div className="space-y-1.5">
                  <Label htmlFor="profile-first-name">First name</Label>
                  <Input id="profile-first-name" autoComplete="given-name" required value={values.first_name} onChange={(e) => update({ first_name: e.target.value })} />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="profile-last-name">Last name</Label>
                  <Input id="profile-last-name" autoComplete="family-name" required value={values.last_name} onChange={(e) => update({ last_name: e.target.value })} />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="profile-gender">Gender</Label>
                  <Select value={values.gender} onValueChange={(value) => update({ gender: value })}>
                    <SelectTrigger id="profile-gender" className="w-full"><SelectValue placeholder="Select" /></SelectTrigger>
                    <SelectContent>
                      {values.gender && !GENDERS.some((option) => option.value === values.gender) && <SelectItem value={values.gender}>{values.gender}</SelectItem>}
                      {GENDERS.map((option) => <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>)}
                    </SelectContent>
                  </Select>
                </div>
              </FieldGroup>

              <FieldGroup title="Academic">
                <div className="space-y-1.5">
                  <Label htmlFor="profile-level">Level</Label>
                  <Select value={values.graduate_level} onValueChange={(value) => update({ graduate_level: value })}>
                    <SelectTrigger id="profile-level" className="w-full"><SelectValue placeholder="Select" /></SelectTrigger>
                    <SelectContent>
                      {values.graduate_level && !LEVELS.some((option) => option.value === values.graduate_level) && <SelectItem value={values.graduate_level}>{values.graduate_level}</SelectItem>}
                      {LEVELS.map((option) => <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>)}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="profile-year">Graduation year</Label>
                  <Input id="profile-year" type="number" min={0} max={2100} value={values.graduation_year || ""} onChange={(e) => update({ graduation_year: e.target.value === "" ? 0 : Number(e.target.value) })} />
                </div>
                <div className="space-y-1.5 sm:col-span-2">
                  <Label htmlFor="profile-major">Major</Label>
                  <Input id="profile-major" value={values.major} onChange={(e) => update({ major: e.target.value })} />
                </div>
              </FieldGroup>

              <FieldGroup title="Team">
                {(["shirt_size", "jacket_size"] as const).map((field) => (
                  <div className="space-y-1.5" key={field}>
                    <Label htmlFor={`profile-${field}`}>{field === "shirt_size" ? "Shirt size" : "Jacket size"}</Label>
                    <Select value={values[field]} onValueChange={(value) => update({ [field]: value })}>
                      <SelectTrigger id={`profile-${field}`} className="w-full"><SelectValue placeholder="Select" /></SelectTrigger>
                      <SelectContent>
                        {values[field] && !SIZES.includes(values[field]) && <SelectItem value={values[field]}>{values[field]}</SelectItem>}
                        {SIZES.map((size) => <SelectItem key={size} value={size}>{size}</SelectItem>)}
                      </SelectContent>
                    </Select>
                  </div>
                ))}
                <div className="space-y-1.5 sm:col-span-2">
                  <Label htmlFor="profile-sae">SAE registration number</Label>
                  <Input id="profile-sae" value={values.sae_registration_number} onChange={(e) => update({ sae_registration_number: e.target.value })} />
                </div>
              </FieldGroup>

              <FieldGroup title="Occupation">
                <div className="space-y-1.5">
                  <Label htmlFor="profile-title">Job title</Label>
                  <Input id="profile-title" value={values.occupation_title} onChange={(e) => update({ occupation_title: e.target.value })} />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="profile-company">Company</Label>
                  <Input id="profile-company" value={values.occupation_company} onChange={(e) => update({ occupation_company: e.target.value })} />
                </div>
              </FieldGroup>
            </CardContent>
          </Card>
        </Section>

        {changed && (
          <div className="sticky bottom-4 z-10 mt-6 flex justify-end animate-in fade-in-0 slide-in-from-bottom-4 duration-200">
            <div className="flex items-center gap-3 rounded-xl border border-border bg-popover/95 py-2 pr-2 pl-4 shadow-lg backdrop-blur">
              <span className="text-sm text-muted-foreground">You have unsaved changes</span>
              <Button type="button" variant="ghost" size="sm" disabled={saving} onClick={() => setValues(initialValues(profile))}>
                Discard
              </Button>
              <Button type="submit" size="sm" disabled={saving}>
                {saving ? "Saving…" : "Save changes"}
              </Button>
            </div>
          </div>
        )}
      </form>

      <ChangeEmailDialog
        open={changingEmail}
        currentEmail={email}
        onOpenChange={setChangingEmail}
        onChanged={async (newEmail) => {
          toast.success(`Check ${newEmail} for a verification code`)
          await queryClient.invalidateQueries({ queryKey: ["currentEntity"] })
        }}
      />
      <ChangePasswordDialog open={changingPassword} onOpenChange={setChangingPassword} />
    </div>
  )
}

export default function ProfilePage() {
  const { user, isLoading } = useAuth()

  return (
    <PageContainer>
      <PageHeader title="Profile" description="Your account, sign-in details, and team profile." />
      {isLoading ? (
        <Skeleton className="h-96" />
      ) : user?.user ? (
        <ProfileForm key={user.user.id} entity={user} profile={user.user} />
      ) : (
        <p className="text-sm text-muted-foreground">Couldn't load your profile. Refresh the page to try again.</p>
      )}
    </PageContainer>
  )
}
