import { geoEqualEarth, geoPath } from "d3-geo"
import type { Feature, FeatureCollection, Geometry } from "geojson"
import { alpha2ToNumeric } from "i18n-iso-countries"
import { useEffect, useMemo, useState } from "react"
import { feature } from "topojson-client"
import type { GeometryCollection, Topology } from "topojson-specification"
import usAtlasURL from "us-atlas/states-albers-10m.json?url"
import worldAtlasURL from "world-atlas/countries-110m.json?url"

import { Skeleton } from "@/components/ui/skeleton"

export type MapScope = "us" | "world"

export type MapValue = { key: string; label: string; count: number; verified?: number }

type Shape = { id: string; name: string; path: string }

const VIEWBOX = { us: [975, 610], world: [960, 500] } as const
const atlasCache = new Map<MapScope, Promise<Topology>>()

function loadAtlas(scope: MapScope) {
  let atlas = atlasCache.get(scope)
  if (!atlas) {
    atlas = fetch(scope === "us" ? usAtlasURL : worldAtlasURL).then((response) => response.json() as Promise<Topology>)
    atlasCache.set(scope, atlas)
  }
  return atlas
}

function shapesFor(scope: MapScope, topology: Topology): Shape[] {
  const object = topology.objects[scope === "us" ? "states" : "countries"] as GeometryCollection<{ name: string }>
  const collection = feature(topology, object) as FeatureCollection<Geometry, { name: string }>
  const [width, height] = VIEWBOX[scope]
  const path =
    scope === "us"
      ? geoPath()
      : geoPath(geoEqualEarth().fitExtent([[8, 8], [width - 8, height - 8]], collection))
  return collection.features
    .filter((shape: Feature<Geometry, { name: string }>) => shape.properties.name !== "Antarctica")
    .map((shape) => ({
      id: String(shape.id ?? shape.properties.name),
      name: shape.properties.name,
      path: path(shape) ?? "",
    }))
}

// Keys are state names for the US map and ISO alpha-2 codes for the world map.
function keyForShape(scope: MapScope, shape: Shape) {
  return scope === "us" ? shape.name : shape.id
}

function normalizeKey(scope: MapScope, key: string) {
  return scope === "us" ? key : (alpha2ToNumeric(key) ?? key)
}

export function ChoroplethMap({ scope, values, height = 340 }: { scope: MapScope; values: MapValue[]; height?: number }) {
  const [loaded, setLoaded] = useState<{ scope: MapScope; shapes: Shape[] } | null>(null)
  const shapes = loaded?.scope === scope ? loaded.shapes : null
  const [hovered, setHovered] = useState<{ shape: Shape; x: number; y: number } | null>(null)

  useEffect(() => {
    let cancelled = false
    loadAtlas(scope).then((topology) => {
      if (!cancelled) setLoaded({ scope, shapes: shapesFor(scope, topology) })
    })
    return () => {
      cancelled = true
    }
  }, [scope])

  const byKey = useMemo(() => {
    const map = new Map<string, MapValue>()
    for (const value of values) map.set(normalizeKey(scope, value.key), value)
    return map
  }, [scope, values])
  const max = Math.max(1, ...values.map((value) => value.count))

  if (!shapes) return <Skeleton className="w-full rounded-md" style={{ height }} />

  const [width, viewHeight] = VIEWBOX[scope]
  const hoveredValue = hovered ? byKey.get(keyForShape(scope, hovered.shape)) : undefined

  return (
    <div className="relative" style={{ height }} onMouseLeave={() => setHovered(null)}>
      <svg viewBox={`0 0 ${width} ${viewHeight}`} className="size-full" role="img" aria-label={`${scope === "us" ? "United States" : "World"} map of member phone numbers`}>
        {shapes.map((shape) => {
          const value = byKey.get(keyForShape(scope, shape))
          const intensity = value ? 0.25 + 0.75 * Math.sqrt(value.count / max) : 0
          return (
            <path
              key={shape.id}
              d={shape.path}
              fill={value ? "var(--color-gr-pink)" : "var(--color-muted)"}
              fillOpacity={value ? intensity : 0.5}
              stroke="var(--color-background)"
              strokeWidth={scope === "us" ? 1 : 0.5}
              className="transition-[fill-opacity] hover:fill-opacity-100"
              onMouseMove={(event) => {
                const bounds = event.currentTarget.ownerSVGElement?.parentElement?.getBoundingClientRect()
                if (!bounds) return
                setHovered({ shape, x: event.clientX - bounds.left, y: event.clientY - bounds.top })
              }}
            />
          )
        })}
      </svg>

      {hovered && (
        <div
          className="pointer-events-none absolute z-10 rounded-lg border border-border bg-card px-3 py-2 text-xs shadow-md"
          style={{ left: hovered.x + 12, top: hovered.y + 12 }}
        >
          <p className="font-medium text-foreground">{hoveredValue?.label ?? hovered.shape.name}</p>
          <p className="text-muted-foreground">
            {hoveredValue
              ? `${hoveredValue.count} member${hoveredValue.count === 1 ? "" : "s"}${
                  hoveredValue.verified !== undefined ? ` · ${hoveredValue.verified} verified` : ""
                }`
              : "No members"}
          </p>
        </div>
      )}

      <div className="absolute bottom-0 left-0 flex items-center gap-2 text-[11px] text-muted-foreground">
        <span>1</span>
        <span className="h-2 w-24 rounded-full bg-gradient-to-r from-gr-pink/25 to-gr-pink" />
        <span>{max}</span>
      </div>
    </div>
  )
}
