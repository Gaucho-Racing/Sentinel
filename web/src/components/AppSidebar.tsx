import { BarChart3, Boxes, LayoutDashboard, Settings, Users } from "lucide-react"
import { Link, useLocation } from "react-router-dom"

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

const NAV_ITEMS = [
  { to: "/", label: "Dashboard", icon: LayoutDashboard },
  { to: "/applications", label: "Applications", icon: Boxes },
  { to: "/groups", label: "Groups", icon: Users },
  { to: "/analytics", label: "Analytics", icon: BarChart3 },
  { to: "/settings", label: "Settings", icon: Settings },
]

function isActive(currentPath: string, target: string) {
  if (target === "/") return currentPath === "/"
  return currentPath === target || currentPath.startsWith(`${target}/`)
}

export function AppSidebar() {
  const { pathname } = useLocation()

  return (
    <Sidebar collapsible="none">
      <SidebarHeader>
        <Link to="/" className="flex items-center gap-2 px-2 py-1.5">
          <div className="size-7 shrink-0 rounded-md bg-gradient-to-br from-gr-pink to-gr-purple" />
          <span className="text-base font-semibold tracking-tight">Sentinel</span>
        </Link>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {NAV_ITEMS.map((item) => (
                <SidebarMenuItem key={item.to}>
                  <SidebarMenuButton asChild isActive={isActive(pathname, item.to)}>
                    <Link to={item.to}>
                      <item.icon className="size-4" />
                      <span>{item.label}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <div className="px-2 py-1.5 text-xs text-muted-foreground">
          <span>Gaucho Racing · </span>
          <span className="font-mono">v0.1.0</span>
        </div>
      </SidebarFooter>
    </Sidebar>
  )
}
