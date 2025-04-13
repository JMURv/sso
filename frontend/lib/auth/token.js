import {cookies, headers} from "next/headers"

export async function getSessionToken() {
    const c = await cookies()
    const access = c.get("access")?.value || ""
    if (access) return access

    const refresh = c.get("refresh")?.value || ""
    if (refresh) {
        const headersList = await headers()
        const userAgent = headersList.get("user-agent") || "unknown"
        const ip = headersList.get("x-forwarded-for") || "unknown"

        // TODO: set cookies on the client
        try {
            const r = await fetch(`${process.env.BACKEND_URL}/api/auth/jwt/refresh`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "User-Agent": userAgent,
                    "X-Forwarded-For": "123.123.123.123"
                },
                cache: "no-store",
                body: JSON.stringify({refresh: refresh}),
            })

            if (!r.ok) {
                const data = await r.json()
                return null
            }

            const data = await r.json()
            return data.access
        } catch (e) {
            console.error(e)
            return null
        }
    }

    return null
}