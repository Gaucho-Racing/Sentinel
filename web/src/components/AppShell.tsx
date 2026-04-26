import { Outlet } from "react-router-dom"

import { AppFooter } from "@/components/AppFooter"
import { AppHeader } from "@/components/AppHeader"
import { AppSidebar } from "@/components/AppSidebar"
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar"
import { TooltipProvider } from "@/components/ui/tooltip"

export function AppShell() {
  return (
    <TooltipProvider>
      <SidebarProvider className="h-svh">
        <AppSidebar />
        <SidebarInset className="overflow-y-auto">
          <div className="flex min-h-full flex-col">
            <AppHeader />
            <div className="flex-1">
              <Outlet />
            </div>
            <AppFooter />
          </div>
        </SidebarInset>
      </SidebarProvider>
    </TooltipProvider>
  )
}
