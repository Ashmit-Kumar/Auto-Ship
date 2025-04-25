import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Github } from "lucide-react"

export function HeroSection() {
  return (
    <section className="w-full py-12 md:py-24 lg:py-32">
      <div className="container px-4 md:px-6">
        <div className="grid gap-6 lg:grid-cols-[1fr_400px] lg:gap-12 xl:grid-cols-[1fr_600px]">
          <div className="flex flex-col justify-center space-y-4">
            <div className="space-y-2">
              <h1 className="text-3xl font-bold tracking-tighter sm:text-5xl xl:text-6xl/none">
                Host Your GitHub Repositories with a Single Click
              </h1>
              <p className="max-w-[600px] text-muted-foreground md:text-xl">
                GitHost makes it easy to deploy and share your projects. Connect your GitHub account, select a
                repository, and we'll handle the rest.
              </p>
            </div>
            <div className="flex flex-col gap-2 min-[400px]:flex-row">
              <Link href="/signup">
                <Button size="lg" className="w-full min-[400px]:w-auto">
                  Get Started
                </Button>
              </Link>
              <Link href="/login">
                <Button size="lg" variant="outline" className="w-full min-[400px]:w-auto">
                  <Github className="mr-2 h-4 w-4" />
                  Login with GitHub
                </Button>
              </Link>
            </div>
          </div>
          <div className="flex items-center justify-center">
            <div className="relative h-[350px] w-full overflow-hidden rounded-xl border bg-background p-4 shadow-xl sm:h-[400px] lg:h-[500px]">
              <div className="absolute inset-0 bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-950/30 dark:to-indigo-950/30">
                <div className="absolute inset-0 bg-grid-black/[0.02] dark:bg-grid-white/[0.02]" />
              </div>
              <div className="relative z-10 flex h-full flex-col rounded-lg border bg-background p-6 shadow-lg">
                <div className="flex items-center gap-2 border-b pb-4">
                  <div className="h-3 w-3 rounded-full bg-red-500" />
                  <div className="h-3 w-3 rounded-full bg-yellow-500" />
                  <div className="h-3 w-3 rounded-full bg-green-500" />
                  <div className="ml-2 text-sm font-medium">Dashboard</div>
                </div>
                <div className="mt-4 space-y-4">
                  <div className="h-8 w-full rounded-md bg-muted" />
                  <div className="grid gap-4">
                    <div className="h-24 rounded-md border bg-card p-4 shadow-sm">
                      <div className="flex justify-between">
                        <div className="h-4 w-32 rounded bg-muted" />
                        <div className="h-4 w-16 rounded bg-green-100 dark:bg-green-900" />
                      </div>
                      <div className="mt-2 h-4 w-48 rounded bg-muted" />
                      <div className="mt-2 h-4 w-24 rounded bg-muted" />
                    </div>
                    <div className="h-24 rounded-md border bg-card p-4 shadow-sm">
                      <div className="flex justify-between">
                        <div className="h-4 w-32 rounded bg-muted" />
                        <div className="h-4 w-16 rounded bg-blue-100 dark:bg-blue-900" />
                      </div>
                      <div className="mt-2 h-4 w-48 rounded bg-muted" />
                      <div className="mt-2 h-4 w-24 rounded bg-muted" />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
