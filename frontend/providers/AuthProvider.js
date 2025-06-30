"use client"
import React, {createContext, useContext, useEffect, useState} from "react"
import {GetMeCli} from "../lib/fetches/users/me"
import {toast} from "sonner"
import {useRouter} from "next/navigation"

const AuthContext = createContext()

export const useAuth = () => {
    return useContext(AuthContext)
}

export const AuthProvider = ({children}) => {
    const router = useRouter()
    const [me, setMe] = useState(null)
    const [roles, setRoles] = useState([])
    const [isAdmin, setIsAdmin] = useState(false)

    async function authFetch(url, options = {}) {
        const opts = { ...options }
        opts.headers = {
            ...opts.headers,
            headers: {"User-Agent": navigator.userAgent},
        }

        let response = await fetch(url, opts)
        if (response.status !== 401) {
            return response
        }

        await refreshSession()
        response = await fetch(url, opts)
        if (response.status === 401) {
            return router.push("/auth")
        }
        return response
    }

    async function adminFetch(url, options = {}) {
        if (!isAdmin) {
            return toast.error("You are not an admin")
        }
        return authFetch(url, options)
    }

    async function refreshSession() {
        try {
            await fetch(`/api/auth/jwt/refresh`, {
                method: "POST",
                headers: {
                    "User-Agent": navigator.userAgent,
                },
            })
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
                },
            })

            if (!r.ok) {
                const data = await r.json()
                toast.error(data.errors)
            }

            toast.success("Logout successful")
            setRoles([])
            setIsAdmin(false)
            await router.push("/auth")
        } catch (e) {
            console.error(e)
        }
    }

    const loadUserData = async () => {
        const response = authFetch(`/api/users/me`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
            },
            cache: "no-store",
        })

        const r = await response
        if (r && r.ok) {
            const me = await r.json()
            setMe(me)
            setRoles(me?.roles)
            setIsAdmin(me?.roles.some(role => role.name === "admin"))
        }
    }

    useEffect(() => {
        loadUserData()
    }, [])


    return (
        <AuthContext.Provider value={{ me, setMe, roles, isAdmin, authFetch, adminFetch, logout }}>
            {children}
        </AuthContext.Provider>
    )
}