"use client"

import { useEffect, useState } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Skeleton } from "@/components/ui/skeleton"
import { useToast } from "@/components/ui/use-toast"
import { deleteProject, getProjects, rebuildProject } from "@/lib/api"
import { ExternalLink, Github, MoreVertical, RefreshCw, Trash } from "lucide-react"

type Project = {
  id: string
  name: string
  repoUrl: string
  status: "cloning" | "building" | "hosted" | "failed"
  deployedUrl?: string
  createdAt: string
  updatedAt: string
}

export function ProjectList({ status }: { status?: Project["status"] }) {
  const [projects, setProjects] = useState<Project[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const { toast } = useToast()

  useEffect(() => {
    const fetchProjects = async () => {
      try {
        const data = await getProjects()
        setProjects(data)
      } catch (error) {
        toast({
          variant: "destructive",
          title: "Failed to load projects",
          description: "There was a problem loading your projects.",
        })
      } finally {
        setIsLoading(false)
      }
    }

    fetchProjects()

    // Set up polling for status updates
    const interval = setInterval(fetchProjects, 10000) // Poll every 10 seconds

    return () => clearInterval(interval)
  }, [toast])

  const handleRebuild = async (id: string) => {
    try {
      await rebuildProject(id)
      toast({
        title: "Rebuild initiated",
        description: "Your project is being rebuilt.",
      })

      // Update the project status in the UI
      setProjects(
        projects.map((project) => (project.id === id ? { ...project, status: "building" as const } : project)),
      )
    } catch (error) {
      toast({
        variant: "destructive",
        title: "Failed to rebuild",
        description: "There was a problem rebuilding your project.",
      })
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await deleteProject(id)
      toast({
        title: "Project deleted",
        description: "Your project has been deleted.",
      })

      // Remove the project from the UI
      setProjects(projects.filter((project) => project.id !== id))
    } catch (error) {
      toast({
        variant: "destructive",
        title: "Failed to delete",
        description: "There was a problem deleting your project.",
      })
    }
  }

  const filteredProjects = status ? projects.filter((project) => project.status === status) : projects

  if (isLoading) {
    return (
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <Card key={i}>
            <CardHeader>
              <Skeleton className="h-5 w-40" />
              <Skeleton className="h-4 w-64" />
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-3/4" />
              </div>
            </CardContent>
            <CardFooter>
              <Skeleton className="h-10 w-28" />
            </CardFooter>
          </Card>
        ))}
      </div>
    )
  }

  if (filteredProjects.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>No projects found</CardTitle>
          <CardDescription>
            {status ? `You don't have any ${status} projects.` : "You haven't added any projects yet."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">Add a GitHub repository to get started.</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-4">
      {filteredProjects.map((project) => (
        <Card key={project.id}>
          <CardHeader className="flex flex-row items-start justify-between space-y-0">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Github className="h-5 w-5" />
                {project.name}
              </CardTitle>
              <CardDescription>{project.repoUrl}</CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <StatusBadge status={project.status} />
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon">
                    <MoreVertical className="h-4 w-4" />
                    <span className="sr-only">Open menu</span>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuLabel>Actions</DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={() => handleRebuild(project.id)}>
                    <RefreshCw className="mr-2 h-4 w-4" />
                    Rebuild
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => handleDelete(project.id)}>
                    <Trash className="mr-2 h-4 w-4" />
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-sm">
              <p>
                <span className="font-medium">Created:</span> {new Date(project.createdAt).toLocaleDateString()}
              </p>
              <p>
                <span className="font-medium">Last updated:</span> {new Date(project.updatedAt).toLocaleDateString()}
              </p>
            </div>
          </CardContent>
          <CardFooter className="flex justify-between">
            {project.status === "hosted" && project.deployedUrl ? (
              <Button asChild variant="outline">
                <a href={project.deployedUrl} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="mr-2 h-4 w-4" />
                  View Site
                </a>
              </Button>
            ) : (
              <Button disabled variant="outline">
                <ExternalLink className="mr-2 h-4 w-4" />
                View Site
              </Button>
            )}
            <Button variant="outline" onClick={() => handleRebuild(project.id)}>
              <RefreshCw className="mr-2 h-4 w-4" />
              Rebuild
            </Button>
          </CardFooter>
        </Card>
      ))}
    </div>
  )
}

function StatusBadge({ status }: { status: Project["status"] }) {
  switch (status) {
    case "cloning":
      return (
        <Badge variant="outline" className="bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300">
          Cloning
        </Badge>
      )
    case "building":
      return (
        <Badge variant="outline" className="bg-yellow-50 text-yellow-700 dark:bg-yellow-950 dark:text-yellow-300">
          Building
        </Badge>
      )
    case "hosted":
      return (
        <Badge variant="outline" className="bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300">
          Hosted
        </Badge>
      )
    case "failed":
      return (
        <Badge variant="outline" className="bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300">
          Failed
        </Badge>
      )
    default:
      return null
  }
}
