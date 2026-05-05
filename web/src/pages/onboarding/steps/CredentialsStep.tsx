import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import type { StepProps } from "@/pages/onboarding/types"

export function CredentialsStep({ data, update }: StepProps) {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-lg font-semibold tracking-tight">Sign-in details</h2>
        <p className="text-sm text-muted-foreground">
          You'll use these to sign in to Sentinel and any team app that uses it.
        </p>
      </div>

      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="email">Email</Label>
          <Input
            id="email"
            type="email"
            autoComplete="email"
            placeholder="you@ucsb.edu"
            value={data.email}
            onChange={(e) => update({ email: e.target.value })}
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="password">Password</Label>
          <Input
            id="password"
            type="password"
            autoComplete="new-password"
            value={data.password}
            onChange={(e) => update({ password: e.target.value })}
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="passwordConfirm">Confirm password</Label>
          <Input
            id="passwordConfirm"
            type="password"
            autoComplete="new-password"
            value={data.passwordConfirm}
            onChange={(e) => update({ passwordConfirm: e.target.value })}
            required
          />
        </div>
      </div>
    </div>
  )
}
