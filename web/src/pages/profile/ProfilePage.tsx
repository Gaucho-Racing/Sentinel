import { useQueryClient } from "@tanstack/react-query"
import { useEffect, useState, type FormEvent } from "react"
import { useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import { PageContainer, PageHeader } from "@/components/PageContainer"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
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

function ProfileForm({ profile, email }: { profile: Profile; email: string }) {
  const queryClient = useQueryClient()
  const [values, setValues] = useState(() => initialValues(profile))
  const [saving, setSaving] = useState(false)
  const changed = JSON.stringify(values) !== JSON.stringify(initialValues(profile))
  const name = `${profile.first_name} ${profile.last_name}`.trim() || profile.username

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
    <form onSubmit={handleSave} className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Account</CardTitle>
          <CardDescription>Your account identifiers are managed separately from your profile.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-3">
            <Avatar className="size-12">
              <AvatarImage src={profile.avatar_url} alt={name} />
              <AvatarFallback>{name.split(/\s+/).map((part) => part[0]).slice(0, 2).join("").toUpperCase()}</AvatarFallback>
            </Avatar>
            <span className="text-sm font-medium">{name}</span>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="profile-username">Username</Label>
              <Input id="profile-username" value={profile.username} readOnly />
            </div>
            <div className="space-y-2">
              <Label htmlFor="profile-email">Email</Label>
              <Input id="profile-email" value={email} readOnly />
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Personal details</CardTitle>
          <CardDescription>The name and details shown in your team profile.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <Label htmlFor="profile-first-name">First name</Label>
            <Input id="profile-first-name" autoComplete="given-name" required value={values.first_name} onChange={(e) => update({ first_name: e.target.value })} />
          </div>
          <div className="space-y-2">
            <Label htmlFor="profile-last-name">Last name</Label>
            <Input id="profile-last-name" autoComplete="family-name" required value={values.last_name} onChange={(e) => update({ last_name: e.target.value })} />
          </div>
          <div className="space-y-2">
            <Label htmlFor="profile-gender">Gender</Label>
            <Select value={values.gender} onValueChange={(value) => update({ gender: value })}>
              <SelectTrigger id="profile-gender" className="w-full"><SelectValue placeholder="Select" /></SelectTrigger>
              <SelectContent>
                {values.gender && !GENDERS.some((option) => option.value === values.gender) && <SelectItem value={values.gender}>{values.gender}</SelectItem>}
                {GENDERS.map((option) => <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>)}
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Academic details</CardTitle>
          <CardDescription>Your degree and graduation information.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <Label htmlFor="profile-level">Level</Label>
            <Select value={values.graduate_level} onValueChange={(value) => update({ graduate_level: value })}>
              <SelectTrigger id="profile-level" className="w-full"><SelectValue placeholder="Select" /></SelectTrigger>
              <SelectContent>
                {values.graduate_level && !LEVELS.some((option) => option.value === values.graduate_level) && <SelectItem value={values.graduate_level}>{values.graduate_level}</SelectItem>}
                {LEVELS.map((option) => <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>)}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="profile-year">Graduation year</Label>
            <Input id="profile-year" type="number" min={0} max={2100} value={values.graduation_year || ""} onChange={(e) => update({ graduation_year: e.target.value === "" ? 0 : Number(e.target.value) })} />
          </div>
          <div className="space-y-2 sm:col-span-2">
            <Label htmlFor="profile-major">Major</Label>
            <Input id="profile-major" value={values.major} onChange={(e) => update({ major: e.target.value })} />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Team details</CardTitle>
          <CardDescription>Apparel sizes and SAE registration.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          {(["shirt_size", "jacket_size"] as const).map((field) => (
            <div className="space-y-2" key={field}>
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
          <div className="space-y-2 sm:col-span-2">
            <Label htmlFor="profile-sae">SAE registration number</Label>
            <Input id="profile-sae" value={values.sae_registration_number} onChange={(e) => update({ sae_registration_number: e.target.value })} />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Occupation</CardTitle>
          <CardDescription>What you do and where you work.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <Label htmlFor="profile-title">Job title</Label>
            <Input id="profile-title" value={values.occupation_title} onChange={(e) => update({ occupation_title: e.target.value })} />
          </div>
          <div className="space-y-2">
            <Label htmlFor="profile-company">Company</Label>
            <Input id="profile-company" value={values.occupation_company} onChange={(e) => update({ occupation_company: e.target.value })} />
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" disabled={!changed || saving} onClick={() => setValues(initialValues(profile))}>Discard changes</Button>
        <Button type="submit" disabled={!changed || saving}>{saving ? "Saving…" : "Save changes"}</Button>
      </div>
    </form>
  )
}

export default function ProfilePage() {
  const { user, isLoading } = useAuth()
  const [searchParams] = useSearchParams()
  const queryClient = useQueryClient()
  const linkedCallback = searchParams.get("github") === "linked"
  const [linking, setLinking] = useState(false)
  const github = user?.external_auths?.find((auth) => auth.provider === "GITHUB")

  async function linkGithub() {
    setLinking(true)
    try {
      const response = await api.post<{ url: string }>("/github/link")
      window.location.assign(response.data.url)
    } catch (error: unknown) {
      const message = (error as { response?: { data?: { error?: string } } })?.response?.data?.error ?? "Couldn't start GitHub linking."
      toast.error(message)
      setLinking(false)
    }
  }

  useEffect(() => {
    if (linkedCallback) {
      void queryClient.invalidateQueries({ queryKey: ["currentEntity"] })
    }
  }, [linkedCallback, queryClient])

  return (
    <PageContainer className="max-w-3xl">
      <PageHeader title="Profile" description="Manage the details on your team profile." />
      <Card className="mb-4">
        <CardHeader>
          <CardTitle>GitHub</CardTitle>
          <CardDescription>Link your GitHub account to manage Gaucho Racing organization access.</CardDescription>
        </CardHeader>
        <CardContent>
          {github ? <p className="text-sm">Connected as <strong>{github.metadata?.username ?? github.external_id}</strong></p> : (
            <Button type="button" disabled={linking || !user} onClick={linkGithub}>{linking ? "Connecting…" : "Connect GitHub"}</Button>
          )}
        </CardContent>
      </Card>
      {isLoading ? (
        <Skeleton className="h-96" />
      ) : user?.user ? (
        <ProfileForm key={user.user.id} profile={user.user} email={user.email_auth?.email ?? user.user.email ?? ""} />
      ) : (
        <p className="text-sm text-muted-foreground">Couldn't load your profile. Refresh the page to try again.</p>
      )}
    </PageContainer>
  )
}
