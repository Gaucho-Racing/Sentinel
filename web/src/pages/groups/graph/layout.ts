import type { Edge, Node } from "@xyflow/react"
import { MarkerType, Position } from "@xyflow/react"

import { placeNodes } from "@/lib/dagre-layout"

import type { BuiltGraph, EdgeKind, GraphDirection, GraphNodeKind } from "./model"

// Fixed sizes rather than measured ones: dagre needs dimensions before the
// nodes mount, and a measure-then-relayout pass makes the canvas visibly jump.
// The node components are sized to match.
export const NODE_SIZE: Record<GraphNodeKind, { width: number; height: number }> = {
  group: { width: 236, height: 80 },
  and: { width: 68, height: 34 },
  role: { width: 196, height: 44 },
  app: { width: 208, height: 52 },
}

export const EDGE_COLOR = {
  conditional: "#8412fc",
  discord: "#7289da",
  application: "#e105a3",
} as const

export function layoutGraph(graph: BuiltGraph, direction: GraphDirection) {
  const placed = placeNodes(
    graph.nodes.map((node) => ({ id: node.id, ...NODE_SIZE[node.kind] })),
    graph.edges,
    { direction },
  )

  const nodes: Node[] = graph.nodes.map((node) => {
    const size = NODE_SIZE[node.kind]
    const at = placed.get(node.id)!
    return {
      id: node.id,
      type: node.kind,
      data: node.data,
      position: { x: at.x, y: at.y },
      width: size.width,
      height: size.height,
      sourcePosition: direction === "LR" ? Position.Right : Position.Bottom,
      targetPosition: direction === "LR" ? Position.Left : Position.Top,
      draggable: true,
      connectable: false,
    }
  })

  const edges: Edge[] = graph.edges.map((edge) => ({
    id: edge.id,
    source: edge.source,
    target: edge.target,
    type: "smoothstep",
    pathOptions: { borderRadius: 16 },
    data: { kind: edge.kind, required: edge.required ?? false },
    ...edgeAppearance(edge.kind, edge.required ?? false, "normal"),
  }))

  return { nodes, edges }
}

export type EdgeState = "normal" | "active" | "dimmed"

// Edge appearance is recomputed rather than patched when the selection changes,
// so the dim/highlight pass never has to unwind a previous one.
export function edgeAppearance(kind: EdgeKind, required: boolean, state: EdgeState) {
  const color = EDGE_COLOR[kind]
  const dimmed = state === "dimmed"
  return {
    animated: state === "active",
    zIndex: state === "active" ? 10 : 0,
    style: {
      stroke: color,
      strokeWidth: state === "active" ? 2.5 : 1.5,
      // Non-gating application links are advisory — they add the group to the
      // token claim but don't block access — so they get a lighter dashed line.
      strokeDasharray: kind === "application" && !required ? "4 4" : undefined,
      opacity: dimmed ? 0.1 : 1,
    },
    // No opacity on the marker: SVG applies the path's opacity to its markers
    // too, so the arrowhead dims along with the line.
    markerEnd: { type: MarkerType.ArrowClosed, width: 16, height: 16, color },
  }
}
