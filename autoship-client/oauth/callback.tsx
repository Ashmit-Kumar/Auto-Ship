"use client"

import { useEffect } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { useAuth } from "@/hooks/use-auth"

export default function OAuthCallback() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const { token } = useAuth()

  useEffect(() => {
    const token = searchParams.get("token")
    if (token) {
      localStorage.setItem("token", token)
      // If your backend also provides user info, parse and save it here
      router.push("/dashboard")
    }
  }, [searchParams, router])

  return <p>{"Logging you in..."}</p>
}
