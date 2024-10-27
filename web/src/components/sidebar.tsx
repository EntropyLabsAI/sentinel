import * as React from "react"
import { BookIcon, Check, ChevronsUpDown, GalleryVerticalEnd, Search, GithubIcon, InspectIcon, FileIcon, RailSymbol, Building2Icon, LucideBuilding, CogIcon, HistoryIcon, BarChartIcon } from "lucide-react"
import { Link, useLocation } from 'react-router-dom'

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Separator } from "@/components/ui/separator"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarRail,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { Card, Button } from "@radix-ui/themes"
import { CardHeader, CardTitle, CardDescription, CardContent } from "./ui/card"
import { useEffect, useState } from "react"
import { PickaxeIcon } from "lucide-react"
const data = {
  versions: ["0.1.2"],
  navMain: [
    {
      title: "Getting started",
      url: "#",
      items: [
        {
          title: "Supervisors",
          url: "/supervisor",
          isActive: false,
          disabled: false,
          icon: <InspectIcon />
        },
        {
          title: "Tools",
          url: "/tools",
          isActive: false,
          icon: <PickaxeIcon />
        },
        {
          title: "Datasets",
          url: "/datasets",
          isActive: false,
          icon: <FileIcon />
        },
        {
          title: "Agent Runs",
          url: "/agents",
          isActive: false,
          disabled: true,
          icon: <RailSymbol />
        },
        {
          title: "Execution History",
          url: "/execution_history",
          isActive: false,
          disabled: true,
          icon: <HistoryIcon />
        },
        {
          title: "Stats",
          url: "/stats",
          isActive: false,
          disabled: true,
          icon: <BarChartIcon />
        },
        {
          title: "API Spec",
          url: "/api",
          isActive: false,
          disabled: false,
          icon: <CogIcon />
        },
        {
          title: "GitHub",
          url: "https://github.com/EntropyLabsAI/sentinel",
          isActive: false,
          disabled: false,
          icon: <GithubIcon />
        },
        {
          title: "Documentation",
          url: "https://docs.entropy-labs.ai",
          isActive: false,
          disabled: false,
          icon: <BookIcon />
        },
      ],
    }
  ],
}

interface SidebarProps {
  isSocketConnected: boolean;
  children: React.ReactNode;
}

export default function SidebarComponent({ isSocketConnected, children }: SidebarProps) {
  const [selectedVersion, setSelectedVersion] = React.useState(data.versions[0])
  // Access environment variables
  // @ts-ignore
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  // @ts-ignore
  const WEBSOCKET_BASE_URL = import.meta.env.VITE_WEBSOCKET_BASE_URL;


  const location = useLocation();
  const [currentPath, setCurrentPath] = useState<string[]>([])

  useEffect(() => {
    if (location.pathname === '/') {
      setCurrentPath([]);
      return;
    }

    const splitPath = location.pathname.split('/');
    // Remove the first element
    splitPath.shift();

    setCurrentPath(splitPath);
  }, [location.pathname]);


  return (
    <SidebarProvider>
      <Sidebar>
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <SidebarMenuButton
                    size="lg"
                    className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                  >
                    <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                      <LucideBuilding className="size-4" />
                    </div>
                    <div className="flex flex-col gap-0.5 leading-none">
                      <span className="font-semibold">Sentinel</span>
                      <span className="">v{selectedVersion}</span>
                    </div>
                    <ChevronsUpDown className="ml-auto" />
                  </SidebarMenuButton>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  className="w-[--radix-dropdown-menu-trigger-width]"
                  align="start"
                >
                  {data.versions.map((version) => (
                    <DropdownMenuItem
                      key={version}
                      onSelect={() => setSelectedVersion(version)}
                    >
                      v{version}{" "}
                      {version === selectedVersion && (
                        <Check className="ml-auto" />
                      )}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </SidebarMenuItem>
          </SidebarMenu>
          <form>
            {/* <SidebarGroup className="py-0">
              <SidebarGroupContent className="relative">
                <Label htmlFor="search" className="sr-only">
                  Search
                </Label>
                <SidebarInput
                  id="search"
                  placeholder="Search the docs..."
                  className="pl-8"
                />
                <Search className="pointer-events-none absolute left-2 top-1/2 size-4 -translate-y-1/2 select-none opacity-50" />
              </SidebarGroupContent>
            </SidebarGroup> */}
            <SidebarGroup className="py-0">
              <SidebarGroupContent className="relative">
                <div className="flex flex-col gap-0.5 leading-none">
                  <span className="text-xs tracking-wide">Agent Supervision & Evaluation</span>
                </div>
              </SidebarGroupContent>
            </SidebarGroup>
          </form>
        </SidebarHeader>
        <SidebarContent>
          {data.navMain.map((item) => (
            <SidebarGroup key={item.title}>
              <SidebarGroupLabel>{item.title}</SidebarGroupLabel>
              <SidebarGroupContent>
                <SidebarMenu>
                  {item.items.map((subItem) => (
                    <SidebarMenuItem key={subItem.title}>
                      <SidebarMenuButton
                        asChild
                        isActive={currentPath[currentPath.length - 1] === subItem.url}
                        disabled={subItem.disabled}
                        className={subItem.disabled ? "opacity-50 cursor-not-allowed" : ""}
                      >
                        <Link to={subItem.url} onClick={e => subItem.disabled && e.preventDefault()}>
                          {subItem.icon}
                          {subItem.title}
                        </Link>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  ))}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          ))}
        </SidebarContent>

        <SidebarFooter>
          <div className="p-1">
            <Card className="shadow-none">
              <form>
                <CardHeader className="p-4 pb-0">
                  <CardTitle className="text-sm">
                    Config
                  </CardTitle>
                  <CardDescription>
                  </CardDescription>
                </CardHeader>
                <CardContent className="grid gap-2.5 p-4">
                  <p className="text-xs font-mono">[API] {API_BASE_URL}</p>
                  <div className="flex items-center gap-1">
                    <p className="text-xs font-mono">[WS] {WEBSOCKET_BASE_URL}</p>
                    <span
                      className={`ml-2 h-3 w-3 rounded-full ${isSocketConnected ? 'bg-green-500' : 'bg-red-500'
                        }`}
                    ></span>
                  </div>
                </CardContent>
              </form>
            </Card>
          </div>
        </SidebarFooter>
        <SidebarRail />
      </Sidebar>
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink href="/supervisor">
                  <Link to="/">
                    home
                  </Link>
                </BreadcrumbLink>
              </BreadcrumbItem>
              {currentPath.length > 0 && (

                currentPath.map((path) => (
                  <>
                    <BreadcrumbSeparator className="hidden md:block" />
                    <BreadcrumbItem key={path}>
                      <BreadcrumbLink>
                        <Link to={`/${path}`}>
                          {path}
                        </Link>
                      </BreadcrumbLink>
                    </BreadcrumbItem>
                  </>
                ))
              )}
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex-grow mt-24">
          {children}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
