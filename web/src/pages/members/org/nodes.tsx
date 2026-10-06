import { Handle, Position, type Node, type NodeProps } from "@xyflow/react"
import { ChevronDown, ChevronRight, Users } from "lucide-react"

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { entityInitials } from "@/components/EntityChip"
import { cn } from "@/lib/utils"

export type PositionNodeData = {
  title: string
  holderName: string | null
  holderAvatar?: string
  groupName: string | null
  groupMissing: boolean
  memberCount: number
  expanded: boolean
  onToggleRoster: () => void
  onOpenHolder: () => void
}

export type MemberNodeData = {
  name: string
  username?: string
  avatarUrl?: string
  onOpen: () => void
}

type PositionFlowNode = Node<PositionNodeData, "position">
type MemberFlowNode = Node<MemberNodeData, "member">

// The chart is read-only; handles exist only as edge anchors.
function Anchors() {
  return (
    <>
      <Handle type="target" position={Position.Top} isConnectable={false} className="!opacity-0" />
      <Handle
        type="source"
        position={Position.Bottom}
        isConnectable={false}
        className="!opacity-0"
      />
    </>
  )
}

export function PositionNode({ data, selected }: NodeProps<PositionFlowNode>) {
  const vacant = !data.holderName
  return (
    <div
      className={cn(
        "flex size-full flex-col justify-center gap-1 rounded-lg border bg-card px-3 py-2 shadow-sm transition-colors",
        vacant ? "border-dashed border-border" : "border-border/70",
        selected && "border-gr-purple ring-2 ring-gr-purple/40",
      )}
    >
      <Anchors />

      <p className="truncate text-[13px] font-semibold leading-tight">{data.title}</p>

      <button
        type="button"
        onClick={(e) => {
          e.stopPropagation()
          if (!vacant) data.onOpenHolder()
        }}
        disabled={vacant}
        className={cn(
          "flex items-center gap-1.5 text-left",
          !vacant && "hover:underline",
        )}
      >
        {vacant ? (
          <span className="text-xs italic text-muted-foreground">Vacant</span>
        ) : (
          <>
            <Avatar className="size-4">
              <AvatarImage src={data.holderAvatar || undefined} alt={data.holderName!} />
              <AvatarFallback className="bg-gradient-to-br from-gr-pink to-gr-purple text-[8px] font-semibold text-white">
                {entityInitials(data.holderName!)}
              </AvatarFallback>
            </Avatar>
            <span className="truncate text-xs text-muted-foreground">{data.holderName}</span>
          </>
        )}
      </button>

      {data.groupName !== null && (
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation()
            if (data.memberCount > 0) data.onToggleRoster()
          }}
          disabled={data.memberCount === 0}
          className={cn(
            "flex items-center gap-1 self-start rounded border px-1.5 py-0.5 text-[10px] transition-colors",
            data.groupMissing
              ? "border-destructive/50 text-destructive"
              : "border-border/60 text-muted-foreground",
            data.memberCount > 0 && "hover:bg-muted",
          )}
        >
          {data.memberCount > 0 &&
            (data.expanded ? (
              <ChevronDown className="size-2.5" />
            ) : (
              <ChevronRight className="size-2.5" />
            ))}
          <Users className="size-2.5" />
          {data.groupMissing ? "group deleted" : `${data.groupName} · ${data.memberCount}`}
        </button>
      )}
    </div>
  )
}

export function MemberNode({ data }: NodeProps<MemberFlowNode>) {
  return (
    <button
      type="button"
      onClick={data.onOpen}
      className="flex size-full items-center gap-2 rounded-lg border border-border/60 bg-card px-2.5 text-left shadow-sm transition-colors hover:border-gr-purple/60 hover:bg-muted/40"
    >
      <Anchors />
      <Avatar className="size-5">
        <AvatarImage src={data.avatarUrl || undefined} alt={data.name} />
        <AvatarFallback className="bg-gradient-to-br from-gr-pink to-gr-purple text-[8px] font-semibold text-white">
          {entityInitials(data.name)}
        </AvatarFallback>
      </Avatar>
      <span className="min-w-0 flex-1 truncate text-xs">{data.name}</span>
    </button>
  )
}
