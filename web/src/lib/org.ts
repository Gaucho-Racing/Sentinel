import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import { api } from "@/lib/api"

// Mirror of core/model/org_position.go::OrgPosition. The positions form a tree
// through parent_id; a position with a group_id has that group's live
// membership rendered beneath it.
export type OrgPosition = {
  id: string
  title: string
  parent_id: string
  entity_id: string
  group_id: string
  rank: number
  created_at: string
  updated_at: string
}

// Mirror of core/model/identity_summary.go::IdentitySummary.
export type IdentitySummary = {
  id: string
  type: "USER" | "SERVICE_ACCOUNT"
  name: string
  username?: string
  avatar_url?: string
}

export type OrgChartGroup = {
  id: string
  name: string
  member_count: number
  // The position still points at a group that has been deleted. Worth drawing
  // as broken — an empty roster would otherwise read as a team with no members.
  missing: boolean
}

// Mirror of core/api/org_position.go::orgChartNode. Flat list; the tree is
// assembled client-side from parent_id.
export type OrgChartNode = {
  position: OrgPosition
  holder?: IdentitySummary
  group?: OrgChartGroup
  members: IdentitySummary[]
}

export type OrgPositionInput = {
  title: string
  parent_id: string
  entity_id: string
  group_id: string
  rank: number
}

export function useOrgChart(enabled = true) {
  return useQuery({
    queryKey: ["org", "chart"],
    queryFn: async () => {
      const res = await api.get<OrgChartNode[]>("/org/chart")
      return res.data
    },
    enabled,
  })
}

export function useOrgPositions(enabled = true) {
  return useQuery({
    queryKey: ["org", "positions"],
    queryFn: async () => {
      const res = await api.get<OrgPosition[]>("/org/positions")
      return res.data
    },
    enabled,
  })
}

// Every mutation invalidates the whole org namespace: a reparent or a delete
// (which lifts children up a level) changes rows the caller never named.
function useOrgMutation<TArgs>(mutationFn: (args: TArgs) => Promise<unknown>) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["org"] }),
  })
}

export function useCreateOrgPosition() {
  return useOrgMutation(async (input: OrgPositionInput) => {
    const res = await api.post<OrgPosition>("/org/positions", input)
    return res.data
  })
}

export function useUpdateOrgPosition() {
  return useOrgMutation(async ({ id, ...input }: OrgPositionInput & { id: string }) => {
    const res = await api.patch<OrgPosition>(`/org/positions/${id}`, input)
    return res.data
  })
}

export function useDeleteOrgPosition() {
  return useOrgMutation(async (id: string) => {
    await api.delete(`/org/positions/${id}`)
  })
}

export type OrgTreeNode = OrgChartNode & { children: OrgTreeNode[]; depth: number }

// Assembles the flat list into a forest. Multiple roots are expected — a
// faculty advisor usually doesn't sit under the President. Nodes whose parent
// is missing are promoted to roots rather than dropped, and the seen set stops
// a cycle (which the API rejects on write) from recursing forever here.
export function buildOrgTree(nodes: OrgChartNode[]): OrgTreeNode[] {
  const byID = new Map(nodes.map((n) => [n.position.id, n]))
  const childrenOf = new Map<string, OrgChartNode[]>()
  const roots: OrgChartNode[] = []

  for (const node of nodes) {
    const parentID = node.position.parent_id
    if (!parentID || !byID.has(parentID)) {
      roots.push(node)
      continue
    }
    const siblings = childrenOf.get(parentID)
    if (siblings) siblings.push(node)
    else childrenOf.set(parentID, [node])
  }

  const seen = new Set<string>()
  const attach = (node: OrgChartNode, depth: number): OrgTreeNode => {
    seen.add(node.position.id)
    const children = (childrenOf.get(node.position.id) ?? [])
      .filter((child) => !seen.has(child.position.id))
      .map((child) => attach(child, depth + 1))
    return { ...node, children, depth }
  }

  const trees = roots.map((root) => attach(root, 0))

  // Every node in a cycle has a parent that exists, so none of them qualified as
  // a root and the whole ring would otherwise be dropped from the chart. The API
  // rejects cycles on write, but a position silently vanishing is a far worse
  // failure than one drawn at the top level, so promote whatever is unreached.
  for (const node of nodes) {
    if (!seen.has(node.position.id)) trees.push(attach(node, 0))
  }

  return trees
}
