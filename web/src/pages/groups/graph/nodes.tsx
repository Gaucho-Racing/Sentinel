import { Handle, Position, type Node, type NodeProps } from "@xyflow/react"
import { Boxes, Users } from "lucide-react"

import { ApplicationIcon } from "@/components/ApplicationIcon"
import { cn } from "@/lib/utils"
import type {
  AndNodeData,
  AppNodeData,
  GraphDirection,
  GroupNodeData,
  RoleNodeData,
} from "./model"

type GroupFlowNode = Node<GroupNodeData, "group">
type AndFlowNode = Node<AndNodeData, "and">
type RoleFlowNode = Node<RoleNodeData, "role">
type AppFlowNode = Node<AppNodeData, "app">

const SOURCE_DOT: Record<string, string> = {
  DIRECT: "bg-muted-foreground/60",
  DISCORD: "bg-discord-blurple",
  CONDITIONAL: "bg-gr-purple",
}

function handlePositions(direction: GraphDirection) {
  return direction === "LR"
    ? { target: Position.Left, source: Position.Right }
    : { target: Position.Top, source: Position.Bottom }
}

// Handles exist only as edge anchors — the graph is read-only, so they're
// sized down to nothing and never hit-tested.
function Anchors({ direction }: { direction: GraphDirection }) {
  const { target, source } = handlePositions(direction)
  return (
    <>
      <Handle type="target" position={target} isConnectable={false} className="!opacity-0" />
      <Handle type="source" position={source} isConnectable={false} className="!opacity-0" />
    </>
  )
}

export function GroupNode({ data, selected }: NodeProps<GroupFlowNode>) {
  return (
    <div
      className={cn(
        "flex size-full items-center gap-2.5 rounded-lg border bg-card px-3 shadow-sm transition-colors",
        data.missing
          ? "border-dashed border-destructive/50"
          : "border-border/70 hover:border-gr-purple/60",
        selected && "border-gr-purple ring-2 ring-gr-purple/40",
      )}
    >
      <Anchors direction={data.direction} />
      <div
        className={cn(
          "flex size-8 shrink-0 items-center justify-center rounded-md text-sm font-semibold text-white",
          data.missing
            ? "bg-destructive/70"
            : "bg-gradient-to-br from-gr-pink to-gr-purple",
        )}
      >
        {data.missing ? "?" : data.name.slice(0, 1).toUpperCase()}
      </div>

      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-1.5">
          <p
            className={cn(
              "truncate text-[13px] font-medium leading-tight",
              data.missing && "font-mono text-[11px] text-muted-foreground",
            )}
          >
            {data.name}
          </p>
          <span className="flex shrink-0 items-center gap-0.5">
            {data.sources.map((source) => (
              <span
                key={source}
                title={source.toLowerCase()}
                className={cn("size-1.5 rounded-full", SOURCE_DOT[source])}
              />
            ))}
          </span>
        </div>

        {data.missing ? (
          <p className="mt-1 text-[10px] text-destructive">deleted — requirement unsatisfiable</p>
        ) : (
          <div className="mt-1 flex items-center gap-2 text-[10px] text-muted-foreground">
            <span className="flex items-center gap-1">
              <Users className="size-2.5" />
              {data.memberCount}
            </span>
            {data.dependsOn > 0 && <span>needs {data.dependsOn}</span>}
            {data.dependedOnBy > 0 && <span>feeds {data.dependedOnBy}</span>}
            {data.appCount > 0 && (
              <span className="flex items-center gap-1">
                <Boxes className="size-2.5" />
                {data.appCount}
              </span>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

export function AndNode({ data }: NodeProps<AndFlowNode>) {
  return (
    <div
      className={cn(
        "flex size-full items-center justify-center rounded-full border bg-background text-[10px] font-semibold uppercase tracking-wide",
        data.kind === "conditional"
          ? "border-gr-purple/60 text-gr-purple"
          : "border-discord-blurple/60 text-discord-blurple",
      )}
      title={`all ${data.count} required`}
    >
      <Anchors direction={data.direction} />
      all {data.count}
    </div>
  )
}

export function RoleNode({ data, selected }: NodeProps<RoleFlowNode>) {
  return (
    <div
      className={cn(
        "flex size-full items-center gap-2 rounded-lg border bg-card px-3 shadow-sm",
        data.missing
          ? "border-dashed border-destructive/50"
          : "border-discord-blurple/40",
        selected && "border-discord-blurple ring-2 ring-discord-blurple/40",
      )}
    >
      <Anchors direction={data.direction} />
      <span
        className="size-2.5 shrink-0 rounded-full border border-border/60"
        style={{ backgroundColor: data.color ?? "transparent" }}
      />
      <div className="min-w-0">
        <p className="truncate text-[13px] leading-tight">
          {data.missing ? (
            <span className="font-mono text-[11px] text-muted-foreground">{data.roleID}</span>
          ) : (
            `@${data.name}`
          )}
        </p>
        <p className="text-[10px] text-muted-foreground">
          {data.missing ? "unknown discord role" : "discord role"}
        </p>
      </div>
    </div>
  )
}

export function AppNode({ data, selected }: NodeProps<AppFlowNode>) {
  return (
    <div
      className={cn(
        "flex size-full items-center gap-2.5 rounded-lg border bg-card px-3 shadow-sm",
        data.missing ? "border-dashed border-destructive/50" : "border-gr-pink/40",
        selected && "border-gr-pink ring-2 ring-gr-pink/40",
      )}
    >
      <Anchors direction={data.direction} />
      <ApplicationIcon
        name={data.name}
        iconUrl={data.iconURL}
        className="size-7 rounded-md text-xs"
      />
      <div className="min-w-0">
        <p className="truncate text-[13px] font-medium leading-tight">{data.name}</p>
        <p className="text-[10px] text-muted-foreground">
          {data.required ? "access gated" : "application"}
        </p>
      </div>
    </div>
  )
}
