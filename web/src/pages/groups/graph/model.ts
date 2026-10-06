// Pure graph construction. Turns the four flat API lists (groups, conditional
// bindings, Discord role bindings, application links) into the node/edge set
// the canvas renders. Kept free of React so it can be reasoned about — and
// eventually tested — on its own.

import type { Application, ApplicationGroupLink } from "@/lib/applications"
import type { GroupConditionalBinding } from "@/lib/conditional"
import type { DiscordRole, GroupDiscordRoleBinding } from "@/lib/discord"
import type { Group, GroupSource } from "@/lib/groups"

export type GraphDirection = "LR" | "TB"

export type GraphNodeKind = "group" | "and" | "role" | "app"

export type EdgeKind = "conditional" | "discord" | "application"

export type GroupNodeData = {
  direction: GraphDirection
  groupID: string
  name: string
  description: string
  sources: GroupSource[]
  memberCount: number
  ownerCount: number
  // True when a binding references a group ID that no longer exists. Surfaced
  // rather than dropped: a dangling requirement can never be satisfied, so the
  // parent group silently stops granting membership.
  missing: boolean
  dependsOn: number
  dependedOnBy: number
  appCount: number
}

// An AND junction. A binding with two or more requirements is an all-of rule,
// which plain edges can't express — fanning the requirements into a junction
// keeps "A AND B" visually distinct from two independent OR branches.
export type AndNodeData = {
  direction: GraphDirection
  kind: "conditional" | "discord"
  count: number
}

export type RoleNodeData = {
  direction: GraphDirection
  roleID: string
  name: string
  color: string | null
  missing: boolean
}

export type AppNodeData = {
  direction: GraphDirection
  appID: string
  name: string
  iconURL: string
  required: boolean
  missing: boolean
}

export type GraphNodeData = GroupNodeData | AndNodeData | RoleNodeData | AppNodeData

export type GraphNode = {
  id: string
  kind: GraphNodeKind
  data: GraphNodeData
}

export type GraphEdge = {
  id: string
  source: string
  target: string
  kind: EdgeKind
  // Set on application edges whose link is access-gating.
  required?: boolean
}

export type BuildInput = {
  groups: Group[]
  conditionalBindings: GroupConditionalBinding[]
  discordBindings: GroupDiscordRoleBinding[]
  discordRoles: DiscordRole[]
  applications: Application[]
  applicationLinks: ApplicationGroupLink[]
}

export type BuildOptions = {
  direction: GraphDirection
  showDiscord: boolean
  showApplications: boolean
  showIsolated: boolean
  focusNodeID: string | null
  // Kept even when it has no edges, so deep-linking to an unlinked group still
  // lands on something instead of an empty canvas.
  pinnedNodeID: string | null
}

export type BuiltGraph = {
  nodes: GraphNode[]
  edges: GraphEdge[]
  // Groups dropped because they had no edges under the current layers. Counted
  // so the UI can offer to bring them back instead of silently hiding them.
  isolatedCount: number
}

export const groupNodeID = (id: string) => `g:${id}`
export const roleNodeID = (id: string) => `r:${id}`
export const appNodeID = (id: string) => `a:${id}`

export function nodeKindFromID(id: string): GraphNodeKind {
  if (id.startsWith("g:")) return "group"
  if (id.startsWith("r:")) return "role"
  if (id.startsWith("a:")) return "app"
  return "and"
}

function discordRoleColorHex(color: number): string | null {
  if (!color) return null
  return `#${color.toString(16).padStart(6, "0")}`
}

export function buildGraph(input: BuildInput, options: BuildOptions): BuiltGraph {
  const groupsByID = new Map(input.groups.map((g) => [g.id, g]))
  const rolesByID = new Map(input.discordRoles.map((r) => [r.id, r]))
  const appsByID = new Map(input.applications.map((a) => [a.id, a]))

  const nodes = new Map<string, GraphNode>()
  // Keyed by source|target|kind so two OR-branches that happen to name the same
  // single requirement collapse into one line instead of stacking invisibly.
  const edges = new Map<string, GraphEdge>()

  const addEdge = (source: string, target: string, kind: EdgeKind, required = false) => {
    const id = `${source}|${target}|${kind}`
    const existing = edges.get(id)
    if (existing) {
      // A group linked to the same app twice keeps the gating flag.
      if (required) existing.required = true
      return
    }
    edges.set(id, { id, source, target, kind, required })
  }

  const ensureGroupNode = (groupID: string) => {
    const id = groupNodeID(groupID)
    const existing = nodes.get(id)
    if (existing) return id
    const group = groupsByID.get(groupID)
    nodes.set(id, {
      id,
      kind: "group",
      data: {
        direction: options.direction,
        groupID,
        name: group?.name ?? groupID,
        description: group?.description ?? "",
        sources: group?.allowed_sources ?? [],
        memberCount: group?.member_count ?? 0,
        ownerCount: group?.owner_count ?? 0,
        missing: !group,
        dependsOn: 0,
        dependedOnBy: 0,
        appCount: 0,
      },
    })
    return id
  }

  for (const group of input.groups) ensureGroupNode(group.id)

  for (const binding of input.conditionalBindings) {
    const required = binding.required_group_ids ?? []
    if (required.length === 0) continue
    const target = ensureGroupNode(binding.group_id)

    if (required.length === 1) {
      addEdge(ensureGroupNode(required[0]), target, "conditional")
      continue
    }
    const andID = `and:c:${binding.id}`
    nodes.set(andID, {
      id: andID,
      kind: "and",
      data: { direction: options.direction, kind: "conditional", count: required.length },
    })
    for (const requiredID of required) {
      addEdge(ensureGroupNode(requiredID), andID, "conditional")
    }
    addEdge(andID, target, "conditional")
  }

  if (options.showDiscord) {
    for (const binding of input.discordBindings) {
      const roleIDs = binding.discord_role_ids ?? []
      if (roleIDs.length === 0) continue
      const target = ensureGroupNode(binding.group_id)

      const ensureRoleNode = (roleID: string) => {
        const id = roleNodeID(roleID)
        if (nodes.has(id)) return id
        const role = rolesByID.get(roleID)
        nodes.set(id, {
          id,
          kind: "role",
          data: {
            direction: options.direction,
            roleID,
            name: role?.name ?? roleID,
            color: role ? discordRoleColorHex(role.color) : null,
            missing: !role,
          },
        })
        return id
      }

      if (roleIDs.length === 1) {
        addEdge(ensureRoleNode(roleIDs[0]), target, "discord")
        continue
      }
      const andID = `and:d:${binding.id}`
      nodes.set(andID, {
        id: andID,
        kind: "and",
        data: { direction: options.direction, kind: "discord", count: roleIDs.length },
      })
      for (const roleID of roleIDs) addEdge(ensureRoleNode(roleID), andID, "discord")
      addEdge(andID, target, "discord")
    }
  }

  if (options.showApplications) {
    for (const link of input.applicationLinks) {
      const source = ensureGroupNode(link.group_id)
      const id = appNodeID(link.application_id)
      if (!nodes.has(id)) {
        const app = appsByID.get(link.application_id)
        nodes.set(id, {
          id,
          kind: "app",
          data: {
            direction: options.direction,
            appID: link.application_id,
            name: app?.name ?? link.application_id,
            iconURL: app?.icon_url ?? "",
            required: link.required,
            missing: !app,
          },
        })
      } else if (link.required) {
        ;(nodes.get(id)!.data as AppNodeData).required = true
      }
      addEdge(source, id, "application", link.required)
    }
  }

  let nodeList = [...nodes.values()]
  let edgeList = [...edges.values()]

  // Degree is only meaningful once the layer filters above have run — a group
  // with nothing but a hidden application link is isolated in this view.
  const degree = new Map<string, number>()
  for (const edge of edgeList) {
    degree.set(edge.source, (degree.get(edge.source) ?? 0) + 1)
    degree.set(edge.target, (degree.get(edge.target) ?? 0) + 1)
  }

  const isolated = nodeList.filter((n) => n.kind === "group" && !degree.get(n.id))
  if (!options.showIsolated) {
    const dropped = new Set(isolated.map((n) => n.id))
    dropped.delete(options.pinnedNodeID ?? "")
    nodeList = nodeList.filter((n) => !dropped.has(n.id))
  }

  if (options.focusNodeID) {
    const reachable = traverse(edgeList, options.focusNodeID)
    nodeList = nodeList.filter((n) => reachable.nodes.has(n.id))
    edgeList = edgeList.filter((e) => reachable.edges.has(e.id))
  }

  annotateGroupCounts(nodeList, edgeList)

  return { nodes: nodeList, edges: edgeList, isolatedCount: isolated.length }
}

function adjacency(edgeList: GraphEdge[]) {
  const incoming = new Map<string, GraphEdge[]>()
  const outgoing = new Map<string, GraphEdge[]>()
  for (const edge of edgeList) {
    const into = incoming.get(edge.target)
    if (into) into.push(edge)
    else incoming.set(edge.target, [edge])

    const outOf = outgoing.get(edge.source)
    if (outOf) outOf.push(edge)
    else outgoing.set(edge.source, [edge])
  }
  return { incoming, outgoing }
}

// Counts shown on the group node. Requirements and dependents are measured
// through AND junctions so a two-of rule reads as two requirements, not one.
function annotateGroupCounts(nodeList: GraphNode[], edgeList: GraphEdge[]) {
  const { incoming, outgoing } = adjacency(edgeList)

  const expand = (edges: GraphEdge[], side: "source" | "target"): string[] => {
    const out: string[] = []
    for (const edge of edges) {
      const other = edge[side]
      if (nodeKindFromID(other) !== "and") {
        out.push(other)
        continue
      }
      const through = side === "source" ? (incoming.get(other) ?? []) : (outgoing.get(other) ?? [])
      for (const hop of through) out.push(hop[side])
    }
    return out
  }

  for (const node of nodeList) {
    if (node.kind !== "group") continue
    const data = node.data as GroupNodeData
    const inEdges = incoming.get(node.id) ?? []
    const outEdges = outgoing.get(node.id) ?? []
    data.dependsOn = new Set(expand(inEdges, "source")).size
    data.dependedOnBy = new Set(
      expand(
        outEdges.filter((e) => e.kind !== "application"),
        "target",
      ),
    ).size
    data.appCount = outEdges.filter((e) => e.kind === "application").length
  }
}

export type Traversal = {
  nodes: Set<string>
  edges: Set<string>
  upstream: Set<string>
  downstream: Set<string>
}

// Everything the start node depends on, plus everything that depends on it.
// Collecting the edges walked (rather than every edge between reachable nodes)
// keeps sibling links out of the highlight.
export function traverse(edgeList: GraphEdge[], startID: string): Traversal {
  const { incoming, outgoing } = adjacency(edgeList)

  const nodes = new Set<string>([startID])
  const edges = new Set<string>()

  const walk = (adjacency: Map<string, GraphEdge[]>, step: (e: GraphEdge) => string) => {
    const seen = new Set<string>()
    const queue = [startID]
    while (queue.length) {
      const current = queue.shift()!
      for (const edge of adjacency.get(current) ?? []) {
        edges.add(edge.id)
        const next = step(edge)
        nodes.add(next)
        if (seen.has(next)) continue
        seen.add(next)
        queue.push(next)
      }
    }
    return seen
  }

  const upstream = walk(incoming, (e) => e.source)
  const downstream = walk(outgoing, (e) => e.target)

  // An AND junction reached from one of its inputs would otherwise show up with
  // its siblings missing, reading as a weaker rule than it is. Pull them in.
  for (const edge of edgeList) {
    if (nodes.has(edge.target) && nodeKindFromID(edge.target) === "and") {
      nodes.add(edge.source)
      edges.add(edge.id)
    }
  }

  return { nodes, edges, upstream, downstream }
}
