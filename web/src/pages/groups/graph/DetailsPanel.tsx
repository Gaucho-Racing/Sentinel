import { ArrowUpRight, Crown, Focus, Users, X } from "lucide-react"
import type { ReactNode } from "react"
import { Link } from "react-router-dom"

import { ApplicationIcon } from "@/components/ApplicationIcon"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { SOURCE_LABEL } from "@/lib/groups"
import { cn } from "@/lib/utils"
import {
  appNodeID,
  groupNodeID,
  nodeKindFromID,
  roleNodeID,
  type BuildInput,
} from "./model"

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="border-t border-border/60 px-4 py-3">
      <p className="mb-2 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
        {title}
      </p>
      {children}
    </div>
  )
}

function Empty({ children }: { children: ReactNode }) {
  return <p className="text-xs italic text-muted-foreground">{children}</p>
}

function Chip({
  onClick,
  className,
  children,
}: {
  onClick?: () => void
  className?: string
  children: ReactNode
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={!onClick}
      className={cn(
        "inline-flex max-w-full items-center gap-1.5 rounded-md border border-border/60 bg-muted/40 px-2 py-0.5 text-left text-xs transition-colors",
        onClick && "hover:border-gr-purple/60 hover:bg-muted",
        className,
      )}
    >
      {children}
    </button>
  )
}

// Each binding is an AND set; the sets themselves are OR'd. Rendering one row
// per binding with "AND" between the chips and "or" between rows keeps the
// evaluation order readable without a nested list.
function RuleList({
  rows,
  emptyLabel,
}: {
  rows: { id: string; items: { key: string; node: ReactNode }[] }[]
  emptyLabel: string
}) {
  if (rows.length === 0) return <Empty>{emptyLabel}</Empty>
  return (
    <div className="space-y-1.5">
      {rows.map((row, rowIndex) => (
        <div key={row.id} className="flex flex-wrap items-center gap-1.5">
          {rowIndex > 0 && (
            <span className="text-[10px] font-medium uppercase text-muted-foreground">or</span>
          )}
          {row.items.map((item, index) => (
            <span key={item.key} className="flex items-center gap-1.5">
              {index > 0 && (
                <span className="text-[10px] font-medium uppercase text-muted-foreground">
                  and
                </span>
              )}
              {item.node}
            </span>
          ))}
        </div>
      ))}
    </div>
  )
}

export function DetailsPanel({
  nodeID,
  input,
  isolated,
  onIsolateChange,
  onSelect,
  onClose,
}: {
  nodeID: string
  input: BuildInput
  isolated: boolean
  onIsolateChange: (next: boolean) => void
  onSelect: (nodeID: string) => void
  onClose: () => void
}) {
  const kind = nodeKindFromID(nodeID)
  const rawID = nodeID.slice(2)

  const groupName = (id: string) => input.groups.find((g) => g.id === id)?.name

  const groupChip = (id: string) => {
    const name = groupName(id)
    return (
      <Chip onClick={() => onSelect(groupNodeID(id))}>
        <span className={cn("truncate", !name && "font-mono text-[11px] text-destructive")}>
          {name ?? id}
        </span>
      </Chip>
    )
  }

  const roleChip = (id: string) => {
    const role = input.discordRoles.find((r) => r.id === id)
    const hex = role?.color ? `#${role.color.toString(16).padStart(6, "0")}` : null
    return (
      <Chip onClick={() => onSelect(roleNodeID(id))}>
        <span
          className="size-2 shrink-0 rounded-full border border-border/60"
          style={{ backgroundColor: hex ?? "transparent" }}
        />
        <span className={cn("truncate", !role && "font-mono text-[11px] text-destructive")}>
          {role ? `@${role.name}` : id}
        </span>
      </Chip>
    )
  }

  let header: ReactNode
  let body: ReactNode = null

  if (kind === "group") {
    const group = input.groups.find((g) => g.id === rawID)
    const conditional = input.conditionalBindings.filter((b) => b.group_id === rawID)
    const discord = input.discordBindings.filter((b) => b.group_id === rawID)
    const dependents = [
      ...new Set(
        input.conditionalBindings
          .filter((b) => b.required_group_ids?.includes(rawID))
          .map((b) => b.group_id),
      ),
    ]
    const appLinks = input.applicationLinks.filter((l) => l.group_id === rawID)

    header = (
      <>
        <div className="flex items-start gap-3">
          <div className="flex size-9 shrink-0 items-center justify-center rounded-md bg-gradient-to-br from-gr-pink to-gr-purple text-base font-semibold text-white">
            {(group?.name ?? "?").slice(0, 1).toUpperCase()}
          </div>
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium leading-tight">
              {group?.name ?? rawID}
            </p>
            <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">
              {group?.description || "No description."}
            </p>
          </div>
        </div>
        {group ? (
          <div className="mt-3 flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
            <span className="flex items-center gap-1">
              <Users className="size-3" />
              {group.member_count}
            </span>
            <span className="flex items-center gap-1">
              <Crown className="size-3" />
              {group.owner_count}
            </span>
            {group.allowed_sources?.map((source) => (
              <Badge key={source} variant="outline" className="h-5 font-mono text-[10px]">
                {SOURCE_LABEL[source]}
              </Badge>
            ))}
          </div>
        ) : (
          <p className="mt-3 text-xs text-destructive">
            This group no longer exists. Any rule that requires it can never be satisfied.
          </p>
        )}
      </>
    )

    body = (
      <>
        <Section title="Requires">
          <RuleList
            emptyLabel="No conditional rule — membership isn't derived from other groups."
            rows={conditional.map((binding) => ({
              id: binding.id,
              items: (binding.required_group_ids ?? []).map((id) => ({
                key: id,
                node: groupChip(id),
              })),
            }))}
          />
        </Section>

        <Section title="Discord roles">
          <RuleList
            emptyLabel="No Discord role binding."
            rows={discord.map((binding) => ({
              id: binding.id,
              items: (binding.discord_role_ids ?? []).map((id) => ({
                key: id,
                node: roleChip(id),
              })),
            }))}
          />
        </Section>

        <Section title={`Feeds ${dependents.length} group${dependents.length === 1 ? "" : "s"}`}>
          {dependents.length === 0 ? (
            <Empty>Nothing depends on this group.</Empty>
          ) : (
            <div className="flex flex-wrap gap-1.5">
              {dependents.map((id) => (
                <span key={id}>{groupChip(id)}</span>
              ))}
            </div>
          )}
        </Section>

        <Section title="Grants access to">
          {appLinks.length === 0 ? (
            <Empty>Not linked to any application.</Empty>
          ) : (
            <div className="space-y-1.5">
              {appLinks.map((link) => {
                const app = input.applications.find((a) => a.id === link.application_id)
                return (
                  <div key={link.application_id} className="flex items-center gap-2">
                    <Chip onClick={() => onSelect(appNodeID(link.application_id))}>
                      <ApplicationIcon
                        name={app?.name ?? "?"}
                        iconUrl={app?.icon_url}
                        className="size-4 rounded-sm text-[9px]"
                      />
                      <span className="truncate">{app?.name ?? link.application_id}</span>
                    </Chip>
                    {link.required && (
                      <Badge
                        variant="outline"
                        className="h-5 border-gr-pink/40 bg-gr-pink/10 text-[10px] text-gr-pink"
                      >
                        gates access
                      </Badge>
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </Section>

        {group && (
          <div className="border-t border-border/60 px-4 py-3">
            <Button asChild variant="outline" size="sm" className="w-full">
              <Link to={`/groups/${group.id}`}>
                Open group
                <ArrowUpRight className="size-3.5" />
              </Link>
            </Button>
          </div>
        )}
      </>
    )
  } else if (kind === "role") {
    const role = input.discordRoles.find((r) => r.id === rawID)
    const boundGroups = [
      ...new Set(
        input.discordBindings
          .filter((b) => b.discord_role_ids?.includes(rawID))
          .map((b) => b.group_id),
      ),
    ]

    header = (
      <>
        <p className="text-sm font-medium leading-tight">
          {role ? `@${role.name}` : <span className="font-mono text-xs">{rawID}</span>}
        </p>
        <p className="mt-0.5 text-xs text-muted-foreground">
          {role
            ? "Discord role. Membership flows into Sentinel on every sync."
            : "This role is no longer in the guild, so the bindings below never match."}
        </p>
      </>
    )
    body = (
      <Section title="Bound to">
        {boundGroups.length === 0 ? (
          <Empty>Not bound to any group.</Empty>
        ) : (
          <div className="flex flex-wrap gap-1.5">
            {boundGroups.map((id) => (
              <span key={id}>{groupChip(id)}</span>
            ))}
          </div>
        )}
      </Section>
    )
  } else if (kind === "app") {
    const app = input.applications.find((a) => a.id === rawID)
    const links = input.applicationLinks.filter((l) => l.application_id === rawID)
    const gating = links.filter((l) => l.required)

    header = (
      <>
        <div className="flex items-start gap-3">
          <ApplicationIcon
            name={app?.name ?? "?"}
            iconUrl={app?.icon_url}
            className="size-9 rounded-md text-base"
          />
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium leading-tight">{app?.name ?? rawID}</p>
            <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">
              {app?.description || "No description."}
            </p>
          </div>
        </div>
        <p className="mt-3 text-xs text-muted-foreground">
          {gating.length > 0
            ? `Access requires membership in at least one of ${gating.length} gating group${gating.length === 1 ? "" : "s"}.`
            : "No group gates access — every linked group only adds a token claim."}
        </p>
      </>
    )
    body = (
      <>
        <Section title="Linked groups">
          {links.length === 0 ? (
            <Empty>No linked groups.</Empty>
          ) : (
            <div className="space-y-1.5">
              {links.map((link) => (
                <div key={link.group_id} className="flex items-center gap-2">
                  {groupChip(link.group_id)}
                  {link.required && (
                    <Badge
                      variant="outline"
                      className="h-5 border-gr-pink/40 bg-gr-pink/10 text-[10px] text-gr-pink"
                    >
                      gates access
                    </Badge>
                  )}
                </div>
              ))}
            </div>
          )}
        </Section>
        {app && (
          <div className="border-t border-border/60 px-4 py-3">
            <Button asChild variant="outline" size="sm" className="w-full">
              <Link to={`/applications/${app.id}`}>
                Open application
                <ArrowUpRight className="size-3.5" />
              </Link>
            </Button>
          </div>
        )}
      </>
    )
  } else {
    header = (
      <>
        <p className="text-sm font-medium leading-tight">All-of junction</p>
        <p className="mt-0.5 text-xs text-muted-foreground">
          Everything feeding this junction has to be held at once. Separate junctions
          into the same group are alternatives — any one of them grants membership.
        </p>
      </>
    )
  }

  return (
    <aside className="flex w-80 shrink-0 flex-col overflow-y-auto border-l border-border/60 bg-card">
      <div className="px-4 py-4">
        <div className="mb-3 flex items-center justify-between gap-2">
          <Button
            variant={isolated ? "secondary" : "ghost"}
            size="sm"
            onClick={() => onIsolateChange(!isolated)}
          >
            <Focus className="size-3.5" />
            {isolated ? "Showing subgraph" : "Isolate"}
          </Button>
          <Button variant="ghost" size="icon-sm" onClick={onClose} aria-label="Close panel">
            <X className="size-4" />
          </Button>
        </div>
        {header}
      </div>
      {body}
    </aside>
  )
}
