import { Bell, Check, ShieldAlert, UserPlus } from "lucide-react"
import { useState } from "react"
import { useNavigate } from "react-router-dom"

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { mockNotifications, type Notification } from "@/lib/mock"

const ICONS: Record<Notification["type"], typeof Bell> = {
  group_request: UserPlus,
  token_issued: Check,
  system: ShieldAlert,
}

function relativeTime(iso: string) {
  const ms = Date.now() - new Date(iso).getTime()
  const minutes = Math.floor(ms / 60_000)
  if (minutes < 1) return "just now"
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

export function HeaderNotifications() {
  const navigate = useNavigate()
  const [items, setItems] = useState(mockNotifications)
  const unreadCount = items.filter((n) => !n.read).length

  function open(notification: Notification) {
    setItems((current) =>
      current.map((n) => (n.id === notification.id ? { ...n, read: true } : n)),
    )
    if (notification.href) navigate(notification.href)
  }

  function markAllRead() {
    setItems((current) => current.map((n) => ({ ...n, read: true })))
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          aria-label={`Notifications${unreadCount ? ` (${unreadCount} unread)` : ""}`}
          className="relative flex size-9 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted/40 hover:text-foreground"
        >
          <Bell className="size-4" />
          {unreadCount > 0 && (
            <span className="absolute right-1.5 top-1.5 size-2 rounded-full bg-gr-pink" />
          )}
        </button>
      </DropdownMenuTrigger>

      <DropdownMenuContent align="end" sideOffset={10} className="w-80 p-0">
        <div className="flex items-center justify-between px-3 py-2">
          <DropdownMenuLabel className="p-0 text-sm">Notifications</DropdownMenuLabel>
          {unreadCount > 0 && (
            <button
              onClick={markAllRead}
              className="text-xs text-muted-foreground transition-colors hover:text-foreground"
            >
              Mark all read
            </button>
          )}
        </div>
        <DropdownMenuSeparator className="m-0" />

        {items.length === 0 ? (
          <div className="px-3 py-6 text-center text-sm text-muted-foreground">
            You're all caught up.
          </div>
        ) : (
          <ul className="max-h-96 overflow-y-auto">
            {items.map((notification) => {
              const Icon = ICONS[notification.type]
              return (
                <li key={notification.id}>
                  <button
                    onClick={() => open(notification)}
                    className="flex w-full items-start gap-3 px-3 py-3 text-left transition-colors hover:bg-muted/40"
                  >
                    <div className="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-md bg-muted/60 text-muted-foreground">
                      <Icon className="size-3.5" />
                    </div>
                    <div className="flex-1 space-y-1">
                      <p className="text-sm leading-tight">{notification.title}</p>
                      {notification.body && (
                        <p className="text-xs text-muted-foreground">{notification.body}</p>
                      )}
                      <p className="text-[11px] text-muted-foreground">{relativeTime(notification.at)}</p>
                    </div>
                    {!notification.read && (
                      <span className="mt-2 size-2 shrink-0 rounded-full bg-gr-pink" />
                    )}
                  </button>
                </li>
              )
            })}
          </ul>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
