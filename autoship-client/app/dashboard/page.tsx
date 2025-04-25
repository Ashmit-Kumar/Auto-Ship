import { AddRepositoryForm } from "@/components/add-repository-form"
import { ProjectList } from "@/components/project-list"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">Host your GitHub repositories with a single click.</p>
      </div>
      <Tabs defaultValue="all" className="space-y-4">
        <TabsList>
          <TabsTrigger value="all">All Projects</TabsTrigger>
          <TabsTrigger value="active">Active</TabsTrigger>
          <TabsTrigger value="building">Building</TabsTrigger>
          <TabsTrigger value="failed">Failed</TabsTrigger>
        </TabsList>
        <TabsContent value="all" className="space-y-4">
          <AddRepositoryForm />
          <ProjectList />
        </TabsContent>
        <TabsContent value="active" className="space-y-4">
          <AddRepositoryForm />
          <ProjectList status="hosted" />
        </TabsContent>
        <TabsContent value="building" className="space-y-4">
          <AddRepositoryForm />
          <ProjectList status="building" />
        </TabsContent>
        <TabsContent value="failed" className="space-y-4">
          <AddRepositoryForm />
          <ProjectList status="failed" />
        </TabsContent>
      </Tabs>
    </div>
  )
}
