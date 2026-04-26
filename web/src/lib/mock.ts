// Mock data for design-stage UI work. Replaced with real API calls later.

export type EntityType = "USER" | "SERVICE_ACCOUNT"

export type Entity = {
  id: string
  type: EntityType
  name: string
  email: string
  avatarUrl?: string
}

export type Application = {
  id: string
  clientId: string
  name: string
  description: string
  iconUrl?: string
  url?: string
  lastAccessedAt?: string
}

export type Group = {
  id: string
  name: string
  description: string
  memberCount: number
  role: "member" | "owner"
}

export type Login = {
  id: string
  applicationId: string
  applicationName: string
  scope: string
  ipAddress: string
  at: string
}

export const mockUser: Entity = {
  id: "ent_01kpgkjbstpswced3c61rjrbkh",
  type: "USER",
  name: "Bharat Kathi",
  email: "bharat@gauchoracing.com",
  avatarUrl: undefined,
}

export const mockApplications: Application[] = [
  {
    id: "app_01kpy5f8263c4rqnhn9v2akdvf",
    clientId: "sentinel",
    name: "Sentinel",
    description: "Gaucho Racing's authentication service",
    url: "https://sso.gauchoracing.com",
    lastAccessedAt: "2026-04-26T09:11:00Z",
  },
  {
    id: "app_01kpwwxn5dwvgmhrr6kefzc89c",
    clientId: "blix",
    name: "Blix",
    description: "Telemetry visualization and dashboarding",
    url: "https://blix.gauchoracing.com",
    lastAccessedAt: "2026-04-25T16:21:00Z",
  },
  {
    id: "app_01kpwwxn83y7mqsdt30ydmg6vt",
    clientId: "wiki",
    name: "Wiki",
    description: "Internal team documentation",
    url: "https://wiki.gauchoracing.com",
    lastAccessedAt: "2026-04-25T14:02:00Z",
  },
  {
    id: "app_01kpwwxnaedwjqfbm95krbyw0e",
    clientId: "rincon",
    name: "Rincon",
    description: "Service registry and routing",
    url: "https://rincon.gauchoracing.com",
    lastAccessedAt: "2026-04-22T08:40:00Z",
  },
  {
    id: "app_01kpwwxnaedw9k4bhx9qjvgg2m",
    clientId: "mechanic",
    name: "Mechanic",
    description: "Build and deployment dashboard",
    url: "https://mechanic.gauchoracing.com",
    lastAccessedAt: "2026-04-21T11:32:00Z",
  },
  {
    id: "app_01kpwwxnaedw88f5gn8jvw9hcq",
    clientId: "mapache",
    name: "Mapache",
    description: "Driver analysis and lap timing",
    url: "https://mapache.gauchoracing.com",
    lastAccessedAt: "2026-04-19T19:05:00Z",
  },
]

export const mockGroups: Group[] = [
  { id: "grp_01kpwwy0001", name: "Members", description: "All Gaucho Racing team members", memberCount: 87, role: "member" },
  { id: "grp_01kpwwy0002", name: "Officers", description: "Team officers", memberCount: 9, role: "owner" },
  { id: "grp_01kpwwy0003", name: "Software", description: "Software subteam", memberCount: 14, role: "owner" },
  { id: "grp_01kpwwy0004", name: "BlixUsers", description: "Granted access to Blix", memberCount: 42, role: "member" },
]

export const mockRecentLogins: Login[] = [
  {
    id: "jwt_01kq07jctsm4agjbdcs6fhmy75",
    applicationId: "app_01kpwwxn5dwvgmhrr6kefzc89c",
    applicationName: "Blix",
    scope: "user:read groups:read",
    ipAddress: "192.168.1.42",
    at: "2026-04-25T16:21:00Z",
  },
  {
    id: "jwt_01kq05vpx1f8aw3qxd2hnn9pmy",
    applicationId: "app_01kpwwxn83y7mqsdt30ydmg6vt",
    applicationName: "Wiki",
    scope: "user:read",
    ipAddress: "192.168.1.42",
    at: "2026-04-25T14:02:00Z",
  },
  {
    id: "jwt_01kq04hgcq77a8dz5skk9z5x6t",
    applicationId: "app_01kpy5f8263c4rqnhn9v2akdvf",
    applicationName: "Sentinel",
    scope: "user:read user:write applications:read",
    ipAddress: "10.0.4.18",
    at: "2026-04-24T22:11:00Z",
  },
]
