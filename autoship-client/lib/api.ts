// Mock API functions for demonstration purposes
// In a real app, these would make actual API calls

type Project = {
  id: string
  name: string
  repoUrl: string
  status: "cloning" | "building" | "hosted" | "failed"
  deployedUrl?: string
  createdAt: string
  updatedAt: string
}

// Mock data
let mockProjects: Project[] = [
  {
    id: "proj_1",
    name: "next-blog",
    repoUrl: "https://github.com/username/next-blog",
    status: "hosted",
    deployedUrl: "https://next-blog.githost.app",
    createdAt: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
    updatedAt: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
  },
  {
    id: "proj_2",
    name: "react-dashboard",
    repoUrl: "https://github.com/username/react-dashboard",
    status: "building",
    createdAt: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
    updatedAt: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
  },
  {
    id: "proj_3",
    name: "api-service",
    repoUrl: "https://github.com/username/api-service",
    status: "failed",
    createdAt: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
    updatedAt: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
  },
]

// Get all projects
export async function getProjects(): Promise<Project[]> {
  // Simulate API delay
  await new Promise((resolve) => setTimeout(resolve, 1000))
  return [...mockProjects]
}

// Add a new project
export async function addProject(repoUrl: string): Promise<Project> {
  // Simulate API delay
  await new Promise((resolve) => setTimeout(resolve, 1500))

  // Extract repo name from URL
  const urlParts = repoUrl.split("/")
  const name = urlParts[urlParts.length - 1]

  const newProject: Project = {
    id: `proj_${Date.now()}`,
    name,
    repoUrl,
    status: "cloning",
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  }

  mockProjects = [newProject, ...mockProjects]

  // Simulate status changes over time
  setTimeout(() => {
    newProject.status = "building"
    newProject.updatedAt = new Date().toISOString()
  }, 3000)

  setTimeout(() => {
    // 80% chance of success
    if (Math.random() > 0.2) {
      newProject.status = "hosted"
      newProject.deployedUrl = `https://${name}.githost.app`
    } else {
      newProject.status = "failed"
    }
    newProject.updatedAt = new Date().toISOString()
  }, 8000)

  return newProject
}

// Rebuild a project
export async function rebuildProject(id: string): Promise<Project> {
  // Simulate API delay
  await new Promise((resolve) => setTimeout(resolve, 1000))

  const project = mockProjects.find((p) => p.id === id)

  if (!project) {
    throw new Error("Project not found")
  }

  project.status = "building"
  project.updatedAt = new Date().toISOString()

  // Simulate build process
  setTimeout(() => {
    // 80% chance of success
    if (Math.random() > 0.2) {
      project.status = "hosted"
      if (!project.deployedUrl) {
        project.deployedUrl = `https://${project.name}.githost.app`
      }
    } else {
      project.status = "failed"
    }
    project.updatedAt = new Date().toISOString()
  }, 5000)

  return project
}

// Delete a project
export async function deleteProject(id: string): Promise<void> {
  // Simulate API delay
  await new Promise((resolve) => setTimeout(resolve, 1000))

  const projectIndex = mockProjects.findIndex((p) => p.id === id)

  if (projectIndex === -1) {
    throw new Error("Project not found")
  }

  mockProjects.splice(projectIndex, 1)
}
