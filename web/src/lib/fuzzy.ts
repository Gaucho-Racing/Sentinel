// Tiny fuzzy matcher. Scores `query` against `target` by walking both strings
// and rewarding (a) substring matches, (b) consecutive character matches, and
// (c) matches that land on word boundaries. Returns null when not all query
// characters can be found in order.
//
// Good enough for the small (handful → ~100 item) lists we filter through the
// UI today. If we ever need typo tolerance, swap in Fuse.js / fuzzysort.

export function fuzzyScore(query: string, target: string): number | null {
  if (!query) return 0
  if (!target) return null
  const q = query.toLowerCase()
  const t = target.toLowerCase()

  // Substring match short-circuit — strongest signal, ranked by how close to
  // the start of the target the match lands.
  const subIdx = t.indexOf(q)
  if (subIdx !== -1) {
    return 1000 - subIdx
  }

  // Subsequence match: every char in q must appear in t in order.
  let qi = 0
  let lastMatch = -1
  let consecutive = 0
  let score = 0
  for (let ti = 0; ti < t.length && qi < q.length; ti++) {
    if (t[ti] !== q[qi]) continue
    if (lastMatch === ti - 1) consecutive++
    else consecutive = 1
    const prev = ti > 0 ? t[ti - 1] : ""
    const isBoundary = ti === 0 || /[\s\-_./@]/.test(prev)
    score += 10 + consecutive * 5 + (isBoundary ? 8 : 0)
    lastMatch = ti
    qi++
  }

  if (qi < q.length) return null
  return score
}

// Filters `items` by `query`, dropping non-matches and sorting matches by
// best score across the keys returned by `getKeys`. Empty query returns the
// list unmodified so the caller can apply its own default sort.
export function fuzzyFilter<T>(
  items: T[],
  query: string,
  getKeys: (item: T) => string[],
): T[] {
  if (!query.trim()) return items
  const scored: { item: T; score: number }[] = []
  for (const item of items) {
    let best: number | null = null
    for (const key of getKeys(item)) {
      const s = fuzzyScore(query, key)
      if (s !== null && (best === null || s > best)) best = s
    }
    if (best !== null) scored.push({ item, score: best })
  }
  scored.sort((a, b) => b.score - a.score)
  return scored.map((entry) => entry.item)
}
