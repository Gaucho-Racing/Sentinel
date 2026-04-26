import { Boxes, Search, User, Users } from "lucide-react"
import { useEffect, useState } from "react"
import { useNavigate } from "react-router-dom"

import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from "@/components/ui/command"
import { mockApplications, mockGroups, mockMembers } from "@/lib/mock"

export function HeaderSearch() {
  const [open, setOpen] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "k" && (event.metaKey || event.ctrlKey)) {
        event.preventDefault()
        setOpen((current) => !current)
      }
    }
    window.addEventListener("keydown", onKeyDown)
    return () => window.removeEventListener("keydown", onKeyDown)
  }, [])

  function go(href: string) {
    setOpen(false)
    navigate(href)
  }

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="flex h-9 w-full max-w-sm items-center gap-2 rounded-md border border-border/60 bg-muted/40 px-3 text-sm text-muted-foreground transition-colors hover:bg-muted/60"
      >
        <Search className="size-4" />
        <span className="flex-1 text-left">Search Sentinel…</span>
        <kbd className="rounded bg-background px-1.5 py-0.5 font-mono text-[10px]">⌘K</kbd>
      </button>

      <CommandDialog
        open={open}
        onOpenChange={setOpen}
        title="Search Sentinel"
        description="Find applications, groups, and members."
      >
        <CommandInput placeholder="Search applications, groups, members…" />
        <CommandList>
          <CommandEmpty>No results.</CommandEmpty>

          <CommandGroup heading="Applications">
            {mockApplications.map((app) => (
              <CommandItem
                key={app.id}
                value={`${app.name} ${app.clientId}`}
                onSelect={() => go(`/applications`)}
              >
                <Boxes className="size-4" />
                <span>{app.name}</span>
                <CommandShortcut className="font-mono">{app.clientId}</CommandShortcut>
              </CommandItem>
            ))}
          </CommandGroup>

          <CommandSeparator />

          <CommandGroup heading="Groups">
            {mockGroups.map((group) => (
              <CommandItem
                key={group.id}
                value={`${group.name} ${group.description}`}
                onSelect={() => go(`/groups`)}
              >
                <Users className="size-4" />
                <span>{group.name}</span>
                <CommandShortcut>{group.memberCount} members</CommandShortcut>
              </CommandItem>
            ))}
          </CommandGroup>

          <CommandSeparator />

          <CommandGroup heading="Members">
            {mockMembers.map((member) => (
              <CommandItem
                key={member.id}
                value={`${member.name} ${member.email}`}
                onSelect={() => go(`/settings`)}
              >
                <User className="size-4" />
                <span>{member.name}</span>
                <CommandShortcut>{member.email}</CommandShortcut>
              </CommandItem>
            ))}
          </CommandGroup>
        </CommandList>
      </CommandDialog>
    </>
  )
}
