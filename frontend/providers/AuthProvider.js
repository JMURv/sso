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
    const [roles, setRoles] = useState([])
    const [isAdmin, setIsAdmin] = useState(false)

    async function authFetch(url, options = {}) {
        const opts = { ...options }
        opts.headers = {
            ...opts.headers,
            Authorization: `Bearer ${access}`
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
                Authorization: `Bearer ${token}`
            }
            response = await fetch(url, retryOpts)
            return response
        }
    }

    async function refreshSession(refresh) {
        try {
            const res = await fetch(`/api/auth/jwt/refresh`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "User-Agent": navigator.userAgent,
                    "X-Forwarded-For": "123.123.123.123"
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
            try {
                if (access) {
                    const me = await GetMeCli(access)
                    console.log(me)
                    setRoles(me?.roles)
                    setIsAdmin(me?.roles.some(role => role.name === "admin"))
                }
            } catch (error) {
                console.error("Failed to load user data:", error)
                await logout()
            }
        }
        loadUserData()
    }, [access])

    return (
        <AuthContext.Provider value={{ access, roles, isAdmin, authFetch, logout }}>
            {children}
        </AuthContext.Provider>
    )
}