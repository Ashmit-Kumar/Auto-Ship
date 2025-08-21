"use client"

import { useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"
import * as z from "zod"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { useToast } from "@/components/ui/use-toast"
import { addProject } from "@/lib/api"

const formSchema = z.object({
  repoUrl: z
    .string()
    .url({ message: "Please enter a valid URL" })
    .refine((url) => url.includes("github.com"), "Please enter a valid GitHub repository URL"),
})

export function AddRepositoryForm() {
  const { toast } = useToast()
  const [isLoading, setIsLoading] = useState(false)

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      repoUrl: "",
    },
  })

  async function onSubmit(values: z.infer<typeof formSchema>) {
    setIsLoading(true)
    try {
      await addProject(values.repoUrl)
      toast({
        title: "Repository added",
        description: "Your repository is being processed.",
      })
      form.reset()
    } catch (error) {
      toast({
        variant: "destructive",
        title: "Failed to add repository",
        description: "There was a problem adding your repository.",
      })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Add Repository</CardTitle>
        <CardDescription>Enter a GitHub repository URL to host it on GitHost.</CardDescription>
      </CardHeader>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)}>
          <CardContent>
            <FormField
              control={form.control}
              name="repoUrl"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>GitHub Repository URL</FormLabel>
                  <FormControl>
                    <Input placeholder="https://github.com/username/repository" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </CardContent>
          <CardFooter>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? "Adding..." : "Add Repository"}
            </Button>
          </CardFooter>
        </form>
      </Form>
    </Card>
  )
}
