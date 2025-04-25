import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { FormDescription, FormLabel } from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { Switch } from "@/components/ui/switch"

export default function SettingsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
        <p className="text-muted-foreground">Manage your account settings and preferences.</p>
      </div>
      <div className="grid gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Profile</CardTitle>
            <CardDescription>Update your personal information.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <FormLabel htmlFor="name">Name</FormLabel>
                <Input id="name" placeholder="Your name" defaultValue="John Doe" />
              </div>
              <div className="space-y-2">
                <FormLabel htmlFor="email">Email</FormLabel>
                <Input id="email" type="email" placeholder="Your email" defaultValue="john@example.com" />
              </div>
            </div>
          </CardContent>
          <CardFooter>
            <Button>Save Changes</Button>
          </CardFooter>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>GitHub Integration</CardTitle>
            <CardDescription>Manage your GitHub connection.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <FormLabel>GitHub Account</FormLabel>
                <FormDescription>
                  Connected as <span className="font-medium">username</span>
                </FormDescription>
              </div>
              <Button variant="outline">Disconnect</Button>
            </div>
            <div className="flex items-center space-x-2">
              <Switch id="auto-sync" defaultChecked />
              <FormLabel htmlFor="auto-sync">Automatically sync with GitHub</FormLabel>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Notifications</CardTitle>
            <CardDescription>Configure how you receive notifications.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center space-x-2">
              <Switch id="build-notifications" defaultChecked />
              <FormLabel htmlFor="build-notifications">Build notifications</FormLabel>
            </div>
            <div className="flex items-center space-x-2">
              <Switch id="deployment-notifications" defaultChecked />
              <FormLabel htmlFor="deployment-notifications">Deployment notifications</FormLabel>
            </div>
            <div className="flex items-center space-x-2">
              <Switch id="error-notifications" defaultChecked />
              <FormLabel htmlFor="error-notifications">Error notifications</FormLabel>
            </div>
          </CardContent>
          <CardFooter>
            <Button>Save Preferences</Button>
          </CardFooter>
        </Card>
      </div>
    </div>
  )
}
