import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Github } from "lucide-react"
import { HeroSection } from "@/components/hero-section"
import { FeatureSection } from "@/components/feature-section"

export default function Home() {
  return (
    <div className="flex min-h-screen flex-col">
      <header className="container flex h-16 items-center justify-between py-4">
        <div className="flex items-center gap-2">
          <Github className="h-6 w-6" />
          <span className="text-xl font-bold">GitHost</span>
        </div>
        <div className="flex items-center gap-4">
          <Link href="/login">
            <Button variant="ghost">Login</Button>
          </Link>
          <Link href="/signup">
            <Button>Sign Up</Button>
          </Link>
        </div>
      </header>
      <main className="flex-1">
        <HeroSection />
        <FeatureSection />
      </main>
      <footer className="border-t py-6">
        <div className="container flex flex-col items-center justify-between gap-4 md:flex-row">
          <div className="flex items-center gap-2">
            <Github className="h-5 w-5" />
            <span className="text-sm font-medium">GitHost</span>
          </div>
          <p className="text-sm text-muted-foreground">
            &copy; {new Date().getFullYear()} GitHost. All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  )
}
