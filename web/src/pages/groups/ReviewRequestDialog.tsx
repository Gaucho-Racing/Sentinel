import { useQueryClient } from "@tanstack/react-query"
import { CheckCircle2, XCircle } from "lucide-react"
import { useEffect, useState } from "react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { api } from "@/lib/api"
import {
  addCustom,
  addPreset,
  DURATION_PRESETS,
  formatAbsoluteDate,
  formatDurationBetween,
  MAX_BY_UNIT,
  type DurationPreset,
  type DurationUnit,
} from "@/lib/duration"
import type { GroupJoinRequest } from "@/lib/groups"

type Action = "approve" | "reject"

export function ReviewRequestDialog({
  open,
  onOpenChange,
  groupID,
  request,
  action,
  reviewerEntityID,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  groupID: string
  request: GroupJoinRequest | null
  action: Action
  reviewerEntityID: string
}) {
  const qc = useQueryClient()
  const [comment, setComment] = useState("")
  const [submitting, setSubmitting] = useState(false)
  const [durationChoice, setDurationChoice] = useState<"keep" | DurationPreset | "custom">("keep")
  const [customAmount, setCustomAmount] = useState(7)
  const [customUnit, setCustomUnit] = useState<DurationUnit>("days")

  // Reset transient state when the dialog opens against a fresh request.
  useEffect(() => {
    if (open) {
      setComment("")
      setDurationChoice("keep")
      setCustomAmount(7)
      setCustomUnit("days")
    }
  }, [open, request?.id])

  if (!request) return null

  const isApprove = action === "approve"
  const requestedDuration = request.has_expiration
    ? formatDurationBetween(request.created_at, request.expires_at)
    : null

  async function submit() {
    if (!request || submitting) return

    let overrideExpiresAt: Date | null = null
    if (isApprove && durationChoice !== "keep") {
      if (durationChoice === "custom") {
        if (
          !Number.isFinite(customAmount) ||
          customAmount <= 0 ||
          customAmount > MAX_BY_UNIT[customUnit]
        ) {
          toast.error(`Pick between 1 and ${MAX_BY_UNIT[customUnit]} ${customUnit}.`)
          return
        }
        overrideExpiresAt = addCustom(new Date(), customAmount, customUnit)
      } else {
        overrideExpiresAt = addPreset(new Date(), durationChoice)
      }
    }

    setSubmitting(true)
    try {
      const trimmed = comment.trim()
      if (trimmed) {
        await api.post(`/groups/${groupID}/requests/${request.id}/comments`, {
          entity_id: reviewerEntityID,
          comment: trimmed,
        })
      }

      const payload: Record<string, unknown> = { reviewed_by: reviewerEntityID }
      if (overrideExpiresAt) {
        payload.has_expiration = true
        payload.expires_at = overrideExpiresAt.toISOString()
      }
      await api.post(`/groups/${groupID}/requests/${request.id}/${action}`, payload)

      qc.invalidateQueries({ queryKey: ["group", groupID, "requests"] })
      qc.invalidateQueries({ queryKey: ["group", groupID, "requests", request.id] })
      qc.invalidateQueries({ queryKey: ["group", groupID, "members"] })
      qc.invalidateQueries({ queryKey: ["group", groupID] })
      toast.success(isApprove ? "Request approved" : "Request rejected")
      onOpenChange(false)
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        `Couldn't ${action} request.`
      toast.error(message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        if (!submitting) onOpenChange(o)
      }}
    >
      <DialogContent className="gap-5 sm:max-w-md">
        <DialogHeader className="gap-3">
          <div
            className={
              "flex size-10 items-center justify-center rounded-xl " +
              (isApprove
                ? "bg-emerald-500/10 text-emerald-500"
                : "bg-destructive/10 text-destructive")
            }
          >
            {isApprove ? (
              <CheckCircle2 className="size-5" />
            ) : (
              <XCircle className="size-5" />
            )}
          </div>
          <DialogTitle>{isApprove ? "Approve request" : "Reject request"}</DialogTitle>
          <DialogDescription>
            {isApprove
              ? "Adds the requester as a direct member of the group."
              : "Lets the requester know they were denied. They can submit a new request later."}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-2">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
            Comment {isApprove ? "(optional)" : "(optional, explanation)"}
          </p>
          <Textarea
            placeholder={isApprove ? "Welcome them in…" : "Why is this being rejected?"}
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            rows={3}
          />
        </div>

        {isApprove && (
          <div className="space-y-2">
            <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Membership duration
            </p>
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                onClick={() => setDurationChoice("keep")}
                className={
                  "rounded-md border px-3 py-1.5 text-sm transition-colors " +
                  (durationChoice === "keep"
                    ? "border-gr-pink/40 bg-gr-pink/10 text-foreground"
                    : "border-border/60 bg-muted/30 text-muted-foreground hover:bg-muted/50")
                }
              >
                Keep as requested
                {requestedDuration && (
                  <span className="ml-1.5 text-xs text-muted-foreground">
                    · {requestedDuration}
                  </span>
                )}
              </button>
              {DURATION_PRESETS.map((p) => {
                const active = durationChoice === p.value
                return (
                  <button
                    key={p.value}
                    type="button"
                    onClick={() => setDurationChoice(p.value)}
                    className={
                      "rounded-md border px-3 py-1.5 text-sm transition-colors " +
                      (active
                        ? "border-gr-pink/40 bg-gr-pink/10 text-foreground"
                        : "border-border/60 bg-muted/30 text-muted-foreground hover:bg-muted/50")
                    }
                  >
                    {p.label}
                  </button>
                )
              })}
              <button
                type="button"
                onClick={() => setDurationChoice("custom")}
                className={
                  "rounded-md border px-3 py-1.5 text-sm transition-colors " +
                  (durationChoice === "custom"
                    ? "border-gr-pink/40 bg-gr-pink/10 text-foreground"
                    : "border-border/60 bg-muted/30 text-muted-foreground hover:bg-muted/50")
                }
              >
                Custom
              </button>
            </div>
            {durationChoice === "custom" && (
              <div className="flex items-center gap-2 pt-1">
                <Input
                  type="number"
                  min={1}
                  max={MAX_BY_UNIT[customUnit]}
                  value={customAmount}
                  onChange={(e) =>
                    setCustomAmount(parseInt(e.target.value, 10) || 0)
                  }
                  className="w-24"
                />
                <Select
                  value={customUnit}
                  onValueChange={(v) => setCustomUnit(v as DurationUnit)}
                >
                  <SelectTrigger className="w-32">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="hours">Hours</SelectItem>
                    <SelectItem value="days">Days</SelectItem>
                    <SelectItem value="weeks">Weeks</SelectItem>
                    <SelectItem value="months">Months</SelectItem>
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  Max {MAX_BY_UNIT[customUnit]}
                </p>
              </div>
            )}
            <p className="text-xs text-muted-foreground">
              Ends{" "}
              {formatAbsoluteDate(
                durationChoice === "keep"
                  ? request.expires_at
                  : (durationChoice === "custom"
                      ? addCustom(new Date(), customAmount, customUnit)
                      : addPreset(new Date(), durationChoice)
                    ).toISOString(),
              )}
            </p>
          </div>
        )}

        <div className="flex justify-end gap-2 pt-1">
          <Button
            type="button"
            variant="ghost"
            disabled={submitting}
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant={isApprove ? "default" : "destructive"}
            disabled={submitting}
            onClick={submit}
          >
            {submitting
              ? isApprove
                ? "Approving…"
                : "Rejecting…"
              : isApprove
                ? "Approve"
                : "Reject"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
