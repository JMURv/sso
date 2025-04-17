import {cookies} from "next/headers"

export async function getSessionToken() {
    const c = await cookies()
    const access = c.get("access")?.value || ""
    if (access) {
        return access
    }

    return null
}

export async function getRefreshToken() {
    const c = await cookies()
    const refresh = c.get("refresh")?.value || ""
    if (refresh) {
        return refresh
    }

    return null
}