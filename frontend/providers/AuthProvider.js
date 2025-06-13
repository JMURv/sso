"use client"
import React, {createContext, useContext, useEffect, useState} from "react"
import {GetMeCli} from "../lib/fetches/users/me"
import {toast} from "sonner"
import {useRouter} from "next/navigation"

const AuthContext = createContext()

export const useAuth = () => {
    return useContext(AuthContext)
}

export const AuthProvider = ({accessSrv, refreshSrv, children}) => {
    const router = useRouter()
    const [access, setAccess] = useState(accessSrv)
    const [refresh, setRefresh] = useState(refreshSrv)
    const [me, setMe] = useState(null)
    const [roles, setRoles] = useState([])
    const [isAdmin, setIsAdmin] = useState(false)
    const [isLoading, setIsLoading] = useState(true)

    async function authFetch(url, options = {}) {
        if (access) {
            const opts = { ...options }
            opts.headers = {
                ...opts.headers,
                Authorization: `Bearer ${access}`,
                headers: {"User-Agent": navigator.userAgent},
            }

            let response = await fetch(url, opts)
            if (response.status !== 403) {
                return response
            }

            const token = await refreshSession(refresh)
            if (token) {
                const retryOpts = { ...options }
                retryOpts.headers = {
                    ...retryOpts.headers,
                    Authorization: `Bearer ${token}`,
                    headers: {"User-Agent": navigator.userAgent},
                }
                response = await fetch(url, retryOpts)
                if (response.status === 401 || response.status === 400) {
                    return router.push("/auth")
                }
                return response
            }
        }
        return router.push("/auth")
    }

    async function adminFetch(url, options = {}) {
        if (!isAdmin) {
            return toast.error("You are not an admin")
        }
        return authFetch(url, options)
    }

    async function refreshSession(refresh) {
        try {
            const res = await fetch(`/api/auth/jwt/refresh`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "User-Agent": navigator.userAgent,
                },
                body: JSON.stringify({ refresh: refresh }),
            })
            if (!res.ok) {
                const error = await res.json()
                console.error(error)
                return
            }

            const data = await res.json()
            setAccess(data.access)
            setRefresh(data.refresh)
            return data.access
        } catch (e) {
            console.error("Error refreshing:", e)
            await logout()
        }
    }

    async function login(access, refresh) {
        setAccess(access)
        setRefresh(refresh)
    }

    async function logout() {
        try {
            const r = await fetch(`/api/auth/logout`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${access}`,
                },
            })

            if (!r.ok) {
                const data = await r.json()
                toast.error(data.errors)
                return null
            }

            toast.success("Logout successful")
            setAccess(null)
            setRefresh(null)
            setRoles([])
            setIsAdmin(false)
            await router.push("/auth")
        } catch (e) {
            console.error(e)
        }
    }

    useEffect(() => {
        const loadUserData = async () => {
            setIsLoading(true)
            try {
                if (access) {
                    const me = await GetMeCli(access)
                    setMe(me)
                    setRoles(me?.roles)
                    setIsAdmin(me?.roles.some(role => role.name === "admin"))
                }
            } catch (error) {
                console.error("Failed to load user data:", error)
                await logout()
            } finally {
                setIsLoading(false)
            }
        }
        loadUserData()
    }, [access])

    if (isLoading) {
        return <div suppressHydrationWarning className="flex justify-center items-center h-screen"/>
    }
    return (
        <AuthContext.Provider value={{ access, me, setMe, roles, isAdmin, authFetch, adminFetch, login, logout }}>
            {children}
        </AuthContext.Provider>
    )
}