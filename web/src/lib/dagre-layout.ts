import Dagre from "@dagrejs/dagre"

// Shared dagre placement for the two React Flow canvases (the group dependency
// graph and the org chart). Only the positioning is common — each chart keeps
// its own node sizes, components, and edge styling.

export type LayoutDirection = "LR" | "TB"

export type SizedNode = {
  id: string
  width: number
  height: number
}

export type LayoutLink = {
  source: string
  target: string
}

export type PlacedNode = {
  id: string
  x: number
  y: number
}

export type DagreOptions = {
  direction: LayoutDirection
  ranksep?: number
  nodesep?: number
  edgesep?: number
}

export function placeNodes(
  nodes: SizedNode[],
  links: LayoutLink[],
  options: DagreOptions,
): Map<string, PlacedNode> {
  const { direction } = options
  const g = new Dagre.graphlib.Graph().setDefaultEdgeLabel(() => ({}))
  g.setGraph({
    rankdir: direction,
    ranksep: options.ranksep ?? (direction === "LR" ? 120 : 90),
    nodesep: options.nodesep ?? (direction === "LR" ? 28 : 44),
    edgesep: options.edgesep ?? 16,
    marginx: 24,
    marginy: 24,
  })

  for (const node of nodes) {
    // Fresh object per node: dagre writes the computed x/y back into the label
    // it is handed, so sharing one size object would collapse every node using
    // it onto a single position.
    g.setNode(node.id, { width: node.width, height: node.height })
  }
  for (const link of links) {
    g.setEdge(link.source, link.target)
  }

  Dagre.layout(g)

  const placed = new Map<string, PlacedNode>()
  for (const node of nodes) {
    const result = g.node(node.id)
    // dagre anchors at the node centre; React Flow anchors at the top left.
    placed.set(node.id, {
      id: node.id,
      x: result.x - node.width / 2,
      y: result.y - node.height / 2,
    })
  }
  return placed
}
