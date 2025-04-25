"use client"

import type React from "react"

import { createContext, useEffect, useState } from "react"
import { useRouter, usePathname } from "next/navigation"

type User = {
  id: string
  name: string
  email: string
}

type AuthContextType = {
  user: User | null
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  signup: (name: string, email: string, password: string) => Promise<void>
  logout: () => void
}

export const AuthContext = createContext<AuthContextType>({
  user: null,
  isLoading: true,
  login: async () => {},
  signup: async () => {},
  logout: () => {},
})

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const router = useRouter()
  const pathname = usePathname()

  // Check if user is logged in on mount
  useEffect(() => {
    const checkAuth = async () => {
      try {
        // In a real app, this would verify the token with your API
        const storedUser = localStorage.getItem("user")
        if (storedUser) {
          setUser(JSON.parse(storedUser))
        }
      } catch (error) {
        console.error("Auth check failed:", error)
      } finally {
        setIsLoading(false)
      }
    }

    checkAuth()
  }, [])

  // Protect routes
  useEffect(() => {
    if (!isLoading) {
      // If user is not logged in and trying to access protected routes
      if (!user && pathname?.startsWith("/dashboard")) {
        router.push("/login")
      }

      // If user is logged in and trying to access auth routes
      if (user && (pathname === "/login" || pathname === "/signup")) {
        router.push("/dashboard")
      }
    }
  }, [user, isLoading, pathname, router])

  const login = async (email: string, password: string) => {
    // In a real app, this would call your API
    // Simulating API call
    return new Promise<void>((resolve, reject) => {
      setTimeout(() => {
        // Mock successful login
        if (email && password) {
          const mockUser = {
            id: "user_" + Math.random().toString(36).substr(2, 9),
            name: email.split("@")[0],
            email,
          }
          setUser(mockUser)
          localStorage.setItem("user", JSON.stringify(mockUser))
          resolve()
        } else {
          reject(new Error("Invalid credentials"))
        }
      }, 1000)
    })
  }

  const signup = async (name: string, email: string, password: string) => {
    // In a real app, this would call your API
    // Simulating API call
    return new Promise<void>((resolve, reject) => {
      setTimeout(() => {
        // Mock successful signup
        if (name && email && password) {
          const mockUser = {
            id: "user_" + Math.random().toString(36).substr(2, 9),
            name,
            email,
          }
          setUser(mockUser)
          localStorage.setItem("user", JSON.stringify(mockUser))
          resolve()
        } else {
          reject(new Error("Invalid user data"))
        }
      }, 1000)
    })
  }

  const logout = () => {
    setUser(null)
    localStorage.removeItem("user")
    router.push("/login")
  }

  return <AuthContext.Provider value={{ user, isLoading, login, signup, logout }}>{children}</AuthContext.Provider>
}
