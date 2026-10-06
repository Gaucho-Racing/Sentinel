import {
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  Panel,
  ReactFlow,
  ReactFlowProvider,
  useEdgesState,
  useNodesState,
  useReactFlow,
  type Edge,
  type Node,
  type NodeTypes,
} from "@xyflow/react"
import { Maximize2, Network } from "lucide-react"
import { useCallback, useEffect, useMemo, useState } from "react"
import { useNavigate } from "react-router-dom"

import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { placeNodes } from "@/lib/dagre-layout"
import { buildOrgTree, useOrgChart, type OrgChartNode } from "@/lib/org"

import { MemberNode, PositionNode, type MemberNodeData, type PositionNodeData } from "./nodes"

import "@xyflow/react/dist/style.css"
import "@/pages/groups/graph/graph.css"

const nodeTypes: NodeTypes = {
  position: PositionNode,
  member: MemberNode,
}

const colorMode = document.documentElement.classList.contains("dark") ? "dark" : "light"

const SIZE = {
  position: { width: 224, height: 86 },
  member: { width: 184, height: 36 },
}

const EDGE_COLOR = "#8412fc"
const ROSTER_EDGE_COLOR = "#8412fc80"

function positionNodeID(id: string) {
  return `p:${id}`
}

function memberNodeID(positionID: string, entityID: string) {
  // Scoped to the position: someone in two attached groups gets a box under
  // each, and React Flow needs those to be distinct nodes.
  return `m:${positionID}:${entityID}`
}

function OrgChartCanvas() {
  const navigate = useNavigate()
  const chartQuery = useOrgChart()
  const [expanded, setExpanded] = useState<Set<string>>(new Set())

  const chart = useMemo(() => chartQuery.data ?? [], [chartQuery.data])
  // Built for the side effect of validating the shape — a parent pointing at a
  // missing row is promoted to a root rather than vanishing from the chart.
  const tree = useMemo(() => buildOrgTree(chart), [chart])

  const toggleRoster = useCallback((positionID: string) => {
    setExpanded((current) => {
      const next = new Set(current)
      if (next.has(positionID)) next.delete(positionID)
      else next.add(positionID)
      return next
    })
  }, [])

  const openEntity = useCallback(
    (entityID: string) => navigate(`/members/${entityID}`),
    [navigate],
  )

  const layout = useMemo(() => {
    const sized: { id: string; width: number; height: number }[] = []
    const links: { source: string; target: string }[] = []
    const build: { id: string; type: "position" | "member"; data: object }[] = []

    const visible = new Set(chart.map((n) => n.position.id))

    for (const node of chart) {
      const id = positionNodeID(node.position.id)
      sized.push({ id, ...SIZE.position })
      build.push({
        id,
        type: "position",
        data: {
          title: node.position.title,
          holderName: node.holder?.name ?? null,
          holderAvatar: node.holder?.avatar_url,
          groupName: node.group ? node.group.name : null,
          groupMissing: node.group?.missing ?? false,
          memberCount: node.members.length,
          expanded: expanded.has(node.position.id),
          onToggleRoster: () => toggleRoster(node.position.id),
          onOpenHolder: () => node.position.entity_id && openEntity(node.position.entity_id),
        } satisfies PositionNodeData,
      })

      const parentID = node.position.parent_id
      if (parentID && visible.has(parentID)) {
        links.push({ source: positionNodeID(parentID), target: id })
      }

      if (!expanded.has(node.position.id)) continue
      for (const member of node.members) {
        const memberID = memberNodeID(node.position.id, member.id)
        sized.push({ id: memberID, ...SIZE.member })
        links.push({ source: id, target: memberID })
        build.push({
          id: memberID,
          type: "member",
          data: {
            name: member.name,
            username: member.username,
            avatarUrl: member.avatar_url,
            onOpen: () => openEntity(member.id),
          } satisfies MemberNodeData,
        })
      }
    }

    const placed = placeNodes(sized, links, { direction: "TB", ranksep: 70, nodesep: 24 })

    const nodes: Node[] = build.map((node) => {
      const at = placed.get(node.id)!
      const size = node.type === "position" ? SIZE.position : SIZE.member
      return {
        id: node.id,
        type: node.type,
        data: node.data as Record<string, unknown>,
        position: { x: at.x, y: at.y },
        width: size.width,
        height: size.height,
        connectable: false,
      }
    })

    const edges: Edge[] = links.map((link) => {
      const toMember = link.target.startsWith("m:")
      return {
        id: `${link.source}->${link.target}`,
        source: link.source,
        target: link.target,
        type: "smoothstep",
        pathOptions: { borderRadius: 12 },
        style: {
          stroke: toMember ? ROSTER_EDGE_COLOR : EDGE_COLOR,
          strokeWidth: toMember ? 1 : 1.5,
          strokeDasharray: toMember ? "4 4" : undefined,
        },
      }
    })

    return { nodes, edges }
  }, [chart, expanded, toggleRoster, openEntity])

  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([])
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([])
  const { fitView } = useReactFlow()

  useEffect(() => {
    setNodes(layout.nodes)
    setEdges(layout.edges)
  }, [layout, setNodes, setEdges])

  // Only refit when the set of drawn nodes changes — expanding a roster should
  // reframe, but a re-render that keeps the same shape should not yank the
  // viewport out from under someone who has panned.
  const shape = layout.nodes.map((n) => n.id).join("|")
  useEffect(() => {
    if (layout.nodes.length === 0) return
    const frame = requestAnimationFrame(() => {
      void fitView({ padding: 0.15, duration: 400, maxZoom: 1 })
    })
    return () => cancelAnimationFrame(frame)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [shape, fitView])

  const positionCount = chart.length
  const filled = chart.filter((n) => n.holder).length

  if (chartQuery.isLoading) {
    return (
      <div className="flex min-h-0 flex-1 overflow-hidden rounded-xl border border-border/60">
        <Skeleton className="size-full rounded-none" />
      </div>
    )
  }

  if (tree.length === 0) {
    return (
      <div className="grid min-h-0 flex-1 place-items-center rounded-xl border border-border/60 px-6 text-center">
        <div>
          <Network className="mx-auto mb-3 size-6 text-muted-foreground" />
          <p className="text-sm font-medium">No org chart yet</p>
          <p className="mt-1 max-w-sm text-xs text-muted-foreground">
            An admin can build the structure in Settings → Admin → Org chart. Positions
            attached to a group render that group's members automatically.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="sentinel-graph relative min-h-0 flex-1 overflow-hidden rounded-xl border border-border/60">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        nodeTypes={nodeTypes}
        nodesConnectable={false}
        edgesFocusable={false}
        minZoom={0.1}
        maxZoom={2}
        colorMode={colorMode}
        fitView
      >
        <Background variant={BackgroundVariant.Dots} gap={18} size={1} />
        <Controls showInteractive={false} position="bottom-right" />
        <MiniMap
          pannable
          zoomable
          position="top-right"
          className="!bg-card"
          maskColor="color-mix(in oklab, var(--background) 70%, transparent)"
          nodeColor={() => EDGE_COLOR}
        />
        <Panel position="top-left">
          <div className="flex items-center gap-2 rounded-lg border border-border/60 bg-card/90 px-3 py-2 backdrop-blur">
            <p className="text-[11px] text-muted-foreground">
              {positionCount} position{positionCount === 1 ? "" : "s"} · {filled} filled
            </p>
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label="Fit view"
              onClick={() => void fitView({ padding: 0.15, duration: 400 })}
            >
              <Maximize2 className="size-3.5" />
            </Button>
          </div>
        </Panel>
      </ReactFlow>
    </div>
  )
}

export default function OrgChartView() {
  return (
    <ReactFlowProvider>
      <OrgChartCanvas />
    </ReactFlowProvider>
  )
}

export type { OrgChartNode }
