import { useQuery } from "@tanstack/react-query"
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
import { Boxes, Bot, EyeOff, LayoutGrid, Maximize2, Sparkles } from "lucide-react"
import {
  useCallback,
  useEffect,
  useImperativeHandle,
  useMemo,
  useRef,
  useState,
  type Ref,
} from "react"
import { useSearchParams } from "react-router-dom"

import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { api } from "@/lib/api"
import { useAllApplicationGroupLinks, useApplications } from "@/lib/applications"
import { useAllConditionalBindings } from "@/lib/conditional"
import { useAllDiscordBindings, useDiscordRoles } from "@/lib/discord"
import type { Group } from "@/lib/groups"
import { cn } from "@/lib/utils"

import { DetailsPanel } from "./DetailsPanel"
import { EDGE_COLOR, edgeAppearance, layoutGraph, type EdgeState } from "./layout"
import {
  buildGraph,
  groupNodeID,
  nodeKindFromID,
  traverse,
  type BuildInput,
  type EdgeKind,
  type GraphDirection,
} from "./model"
import { AndNode, AppNode, GroupNode, RoleNode } from "./nodes"

import "@xyflow/react/dist/style.css"
import "./graph.css"

const nodeTypes: NodeTypes = {
  group: GroupNode,
  and: AndNode,
  role: RoleNode,
  app: AppNode,
}

// The app currently hard-codes dark on <html>; reading the class keeps the
// canvas in step if that ever becomes a user setting.
const colorMode = document.documentElement.classList.contains("dark") ? "dark" : "light"

const MINIMAP_COLOR: Record<string, string> = {
  group: "#8412fc",
  and: "#8412fc",
  role: "#7289da",
  app: "#e105a3",
}

// The search box lives on the page toolbar, outside the canvas. Moving the
// viewport is an imperative command against React Flow rather than a piece of
// derived state, so the page drives it through a handle.
export type GraphHandle = { centerOnGroup: (groupID: string) => void }

function ToggleButton({
  active,
  onClick,
  icon: Icon,
  children,
}: {
  active: boolean
  onClick: () => void
  icon: typeof Boxes
  children: React.ReactNode
}) {
  return (
    <Button
      variant={active ? "secondary" : "ghost"}
      size="sm"
      onClick={onClick}
      aria-pressed={active}
      className={cn(!active && "text-muted-foreground")}
    >
      <Icon className="size-3.5" />
      {children}
    </Button>
  )
}

function Legend() {
  const items: { label: string; color: string; dashed?: boolean }[] = [
    { label: "requires group", color: EDGE_COLOR.conditional },
    { label: "discord role sync", color: EDGE_COLOR.discord },
    { label: "grants app access", color: EDGE_COLOR.application, dashed: true },
  ]
  return (
    <div className="rounded-lg border border-border/60 bg-card/90 px-3 py-2 backdrop-blur">
      <div className="flex flex-col gap-1.5">
        {items.map((item) => (
          <div
            key={item.label}
            className="flex items-center gap-2 text-[11px] text-muted-foreground"
          >
            <svg width="22" height="2" aria-hidden>
              <line
                x1="0"
                y1="1"
                x2="22"
                y2="1"
                stroke={item.color}
                strokeWidth="2"
                strokeDasharray={item.dashed ? "4 3" : undefined}
              />
            </svg>
            {item.label}
          </div>
        ))}
      </div>
    </div>
  )
}

function GroupGraphCanvas({ handleRef }: { handleRef: Ref<GraphHandle> }) {
  const [searchParams, setSearchParams] = useSearchParams()

  const [direction, setDirection] = useState<GraphDirection>("LR")
  const [showDiscord, setShowDiscord] = useState(true)
  const [showApplications, setShowApplications] = useState(true)
  const [showIsolated, setShowIsolated] = useState(false)
  const [isolate, setIsolate] = useState(false)
  const [selectedID, setSelectedID] = useState<string | null>(() => {
    const focus = searchParams.get("focus")
    return focus ? groupNodeID(focus) : null
  })

  const groupsQuery = useQuery({
    queryKey: ["groups"],
    queryFn: async () => {
      const res = await api.get<Group[]>("/groups")
      return res.data
    },
  })
  const conditionalQuery = useAllConditionalBindings()
  const discordBindingsQuery = useAllDiscordBindings()
  const discordRolesQuery = useDiscordRoles()
  const applicationsQuery = useApplications()
  const applicationLinksQuery = useAllApplicationGroupLinks()

  const input: BuildInput = useMemo(
    () => ({
      groups: groupsQuery.data ?? [],
      conditionalBindings: conditionalQuery.data ?? [],
      discordBindings: discordBindingsQuery.data ?? [],
      discordRoles: discordRolesQuery.data ?? [],
      applications: applicationsQuery.data ?? [],
      applicationLinks: applicationLinksQuery.data ?? [],
    }),
    [
      groupsQuery.data,
      conditionalQuery.data,
      discordBindingsQuery.data,
      discordRolesQuery.data,
      applicationsQuery.data,
      applicationLinksQuery.data,
    ],
  )

  const graph = useMemo(
    () =>
      buildGraph(input, {
        direction,
        showDiscord,
        showApplications,
        showIsolated,
        focusNodeID: isolate ? selectedID : null,
        pinnedNodeID: selectedID,
      }),
    [input, direction, showDiscord, showApplications, showIsolated, isolate, selectedID],
  )

  const layout = useMemo(() => layoutGraph(graph, direction), [graph, direction])

  // Null unless something is selected — a null highlight means "draw everything
  // at full strength" rather than "nothing matched".
  const highlight = useMemo(() => {
    if (!selectedID) return null
    if (!graph.nodes.some((n) => n.id === selectedID)) return null
    return traverse(graph.edges, selectedID)
  }, [graph, selectedID])

  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([])
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([])
  const { fitView } = useReactFlow()

  useEffect(() => {
    setNodes(layout.nodes)
    setEdges(layout.edges)
  }, [layout, setNodes, setEdges])

  // Decoration runs as its own pass over whatever is in state, so dragging a
  // node and then changing the selection doesn't snap it back to the layout.
  useEffect(() => {
    setNodes((current) =>
      current.map((node) => ({
        ...node,
        selected: node.id === selectedID,
        className: highlight && !highlight.nodes.has(node.id) ? "is-dimmed" : undefined,
      })),
    )
    setEdges((current) =>
      current.map((edge) => {
        const kind = (edge.data?.kind ?? "conditional") as EdgeKind
        const required = Boolean(edge.data?.required)
        const state: EdgeState = !highlight
          ? "normal"
          : highlight.edges.has(edge.id)
            ? "active"
            : "dimmed"
        return { ...edge, ...edgeAppearance(kind, required, state) }
      }),
    )
  }, [layout, highlight, selectedID, setNodes, setEdges])

  // A deep link into the graph should land on that group rather than on the
  // whole graph, so the first fit after the nodes exist targets it instead.
  const pendingDeepLink = useRef(selectedID)
  useEffect(() => {
    if (layout.nodes.length === 0) return
    const target = pendingDeepLink.current
    const frame = requestAnimationFrame(() => {
      pendingDeepLink.current = null
      if (target && layout.nodes.some((n) => n.id === target)) {
        void fitView({ nodes: [{ id: target }], duration: 500, maxZoom: 1.2, padding: 2 })
        return
      }
      void fitView({ padding: 0.15, duration: 400, maxZoom: 1 })
    })
    return () => cancelAnimationFrame(frame)
  }, [layout, fitView])

  const select = useCallback(
    (nodeID: string | null) => {
      setSelectedID(nodeID)
      setSearchParams(
        (params) => {
          const next = new URLSearchParams(params)
          if (nodeID && nodeKindFromID(nodeID) === "group") {
            next.set("focus", nodeID.slice(2))
          } else {
            next.delete("focus")
          }
          return next
        },
        { replace: true },
      )
    },
    [setSearchParams],
  )

  const centerOn = useCallback(
    (nodeID: string) => {
      select(nodeID)
      requestAnimationFrame(() => {
        void fitView({ nodes: [{ id: nodeID }], duration: 600, maxZoom: 1.2, padding: 2 })
      })
    },
    [fitView, select],
  )

  useImperativeHandle(
    handleRef,
    () => ({ centerOnGroup: (groupID: string) => centerOn(groupNodeID(groupID)) }),
    [centerOn],
  )

  const isLoading = groupsQuery.isLoading || conditionalQuery.isLoading
  const selectionVisible = selectedID && graph.nodes.some((n) => n.id === selectedID)

  const groupCount = graph.nodes.filter((n) => n.kind === "group").length
  const dependencyCount = graph.edges.filter((e) => e.kind !== "application").length

  return (
    <div className="flex min-h-0 flex-1 overflow-hidden rounded-xl border border-border/60 bg-background">
      <div className="sentinel-graph relative min-w-0 flex-1">
        {isLoading ? (
          <div className="grid size-full place-items-center">
            <Skeleton className="size-full rounded-none" />
          </div>
        ) : graph.nodes.length === 0 ? (
          <div className="grid size-full place-items-center px-6 text-center">
            <div>
              <Sparkles className="mx-auto mb-3 size-6 text-muted-foreground" />
              <p className="text-sm font-medium">Nothing to draw yet</p>
              <p className="mt-1 max-w-sm text-xs text-muted-foreground">
                No group derives its membership from another group, a Discord role, or
                gates an application. Add a conditional rule to a group and it will show up
                here.
              </p>
            </div>
          </div>
        ) : (
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            nodeTypes={nodeTypes}
            onNodeClick={(_, node) => select(node.id)}
            onPaneClick={() => {
              select(null)
              setIsolate(false)
            }}
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
              nodeColor={(node) => MINIMAP_COLOR[node.type ?? "group"] ?? "#8412fc"}
            />
            <Panel position="top-left">
              <div className="rounded-lg border border-border/60 bg-card/90 p-1.5 backdrop-blur">
                <div className="flex flex-wrap items-center gap-1">
                  <ToggleButton
                    active={showDiscord}
                    onClick={() => setShowDiscord((v) => !v)}
                    icon={Bot}
                  >
                    Discord
                  </ToggleButton>
                  <ToggleButton
                    active={showApplications}
                    onClick={() => setShowApplications((v) => !v)}
                    icon={Boxes}
                  >
                    Apps
                  </ToggleButton>
                  <ToggleButton
                    active={showIsolated}
                    onClick={() => setShowIsolated((v) => !v)}
                    icon={EyeOff}
                  >
                    Unlinked{graph.isolatedCount > 0 ? ` (${graph.isolatedCount})` : ""}
                  </ToggleButton>
                  <span className="mx-0.5 h-5 w-px bg-border" />
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setDirection((d) => (d === "LR" ? "TB" : "LR"))}
                  >
                    <LayoutGrid className="size-3.5" />
                    {direction === "LR" ? "Horizontal" : "Vertical"}
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label="Fit view"
                    onClick={() => void fitView({ padding: 0.15, duration: 400 })}
                  >
                    <Maximize2 className="size-3.5" />
                  </Button>
                </div>
                <p className="px-2 pb-0.5 pt-1 text-[11px] text-muted-foreground">
                  {groupCount} groups · {dependencyCount} dependencies
                </p>
              </div>
            </Panel>
            <Panel position="bottom-left">
              <Legend />
            </Panel>
          </ReactFlow>
        )}
      </div>

      {selectionVisible && selectedID && (
        <DetailsPanel
          nodeID={selectedID}
          input={input}
          isolated={isolate}
          onIsolateChange={setIsolate}
          onSelect={centerOn}
          onClose={() => {
            select(null)
            setIsolate(false)
          }}
        />
      )}
    </div>
  )
}

export default function GroupGraphView({ ref }: { ref: Ref<GraphHandle> }) {
  return (
    <ReactFlowProvider>
      <GroupGraphCanvas handleRef={ref} />
    </ReactFlowProvider>
  )
}
