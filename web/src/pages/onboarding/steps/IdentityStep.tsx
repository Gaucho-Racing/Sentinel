import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import type { StepProps } from "@/pages/onboarding/types"

export function IdentityStep({ data, update }: StepProps) {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-lg font-semibold tracking-tight">Who are you?</h2>
        <p className="text-sm text-muted-foreground">
          Your name and a username for the team to know you by.
        </p>
      </div>

      <div className="space-y-4">
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-2">
            <Label htmlFor="firstName">First name</Label>
            <Input
              id="firstName"
              autoComplete="given-name"
              value={data.firstName}
              onChange={(e) => update({ firstName: e.target.value })}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="lastName">Last name</Label>
            <Input
              id="lastName"
              autoComplete="family-name"
              value={data.lastName}
              onChange={(e) => update({ lastName: e.target.value })}
              required
            />
          </div>
        </div>

        <div className="space-y-2">
          <Label htmlFor="username">Username</Label>
          <Input
            id="username"
            autoComplete="username"
            placeholder="bharat"
            value={data.username}
            onChange={(e) => update({ username: e.target.value })}
            required
          />
          <p className="text-xs text-muted-foreground">
            Lowercase letters, numbers, dots, and underscores. Used in URLs and mentions.
          </p>
        </div>
      </div>
    </div>
  )
}
