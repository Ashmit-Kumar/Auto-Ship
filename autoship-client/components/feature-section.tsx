import { Code, Rocket, Shield, Github } from "lucide-react"

export function FeatureSection() {
  return (
    <section className="w-full py-12 md:py-24 lg:py-32 bg-muted/50">
      <div className="container px-4 md:px-6">
        <div className="flex flex-col items-center justify-center space-y-4 text-center">
          <div className="space-y-2">
            <div className="inline-block rounded-lg bg-muted px-3 py-1 text-sm">Features</div>
            <h2 className="text-3xl font-bold tracking-tighter sm:text-5xl">Everything You Need</h2>
            <p className="max-w-[900px] text-muted-foreground md:text-xl/relaxed lg:text-base/relaxed xl:text-xl/relaxed">
              GitHost provides all the tools you need to host and share your projects with the world.
            </p>
          </div>
        </div>
        <div className="mx-auto grid max-w-5xl grid-cols-1 gap-6 py-12 md:grid-cols-2 lg:grid-cols-4">
          <div className="flex flex-col items-center space-y-2 rounded-lg border bg-background p-6 shadow-sm">
            <div className="rounded-full bg-primary/10 p-3">
              <Github className="h-6 w-6 text-primary" />
            </div>
            <h3 className="text-xl font-bold">GitHub Integration</h3>
            <p className="text-center text-sm text-muted-foreground">
              Connect your GitHub account and host repositories with a single click.
            </p>
          </div>
          <div className="flex flex-col items-center space-y-2 rounded-lg border bg-background p-6 shadow-sm">
            <div className="rounded-full bg-primary/10 p-3">
              <Rocket className="h-6 w-6 text-primary" />
            </div>
            <h3 className="text-xl font-bold">Fast Deployment</h3>
            <p className="text-center text-sm text-muted-foreground">
              Automatic builds and deployments with real-time status updates.
            </p>
          </div>
          <div className="flex flex-col items-center space-y-2 rounded-lg border bg-background p-6 shadow-sm">
            <div className="rounded-full bg-primary/10 p-3">
              <Code className="h-6 w-6 text-primary" />
            </div>
            <h3 className="text-xl font-bold">Custom Domains</h3>
            <p className="text-center text-sm text-muted-foreground">
              Connect your own domain or use our free subdomain for your projects.
            </p>
          </div>
          <div className="flex flex-col items-center space-y-2 rounded-lg border bg-background p-6 shadow-sm">
            <div className="rounded-full bg-primary/10 p-3">
              <Shield className="h-6 w-6 text-primary" />
            </div>
            <h3 className="text-xl font-bold">Secure Hosting</h3>
            <p className="text-center text-sm text-muted-foreground">
              SSL certificates and secure hosting for all your projects.
            </p>
          </div>
        </div>
      </div>
    </section>
  )
}
