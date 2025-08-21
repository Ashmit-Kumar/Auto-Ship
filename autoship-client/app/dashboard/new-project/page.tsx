import { AddRepositoryForm } from "@/components/add-repository-form"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ArrowLeft } from "lucide-react"
import Link from "next/link"

export default function NewProjectPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Link href="/dashboard">
          <Button variant="ghost" size="icon">
            <ArrowLeft className="h-4 w-4" />
            <span className="sr-only">Back</span>
          </Button>
        </Link>
        <h1 className="text-3xl font-bold tracking-tight">New Project</h1>
      </div>
      <div className="grid gap-6 md:grid-cols-2">
        <AddRepositoryForm />
        <Card>
          <CardHeader>
            <CardTitle>How it works</CardTitle>
            <CardDescription>Learn how GitHost builds and deploys your projects.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <h3 className="font-medium">1. Connect your repository</h3>
              <p className="text-sm text-muted-foreground">Enter the URL of your GitHub repository to get started.</p>
            </div>
            <div className="space-y-2">
              <h3 className="font-medium">2. Build process</h3>
              <p className="text-sm text-muted-foreground">
                We'll clone your repository and build it according to the detected framework.
              </p>
            </div>
            <div className="space-y-2">
              <h3 className="font-medium">3. Deployment</h3>
              <p className="text-sm text-muted-foreground">
                Once built, your project will be deployed to a unique URL.
              </p>
            </div>
            <div className="space-y-2">
              <h3 className="font-medium">4. Continuous updates</h3>
              <p className="text-sm text-muted-foreground">
                You can rebuild your project at any time to update it with the latest changes.
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
