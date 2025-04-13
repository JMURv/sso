import {getSessionToken} from "../../../lib/auth/token"
import {redirect} from "next/navigation"
import GetMe from "../../../lib/fetches/users/me"
import List from "./List"
import listPermissions from "../../../lib/fetches/permissions/list"

export default async function Page({searchParams}) {
    const t = await getSessionToken()
    if (!t) {
        redirect("/auth")
    }

    const sp = await searchParams
    const [me, prms] = await Promise.all([
        GetMe(t),
        listPermissions(t, new URLSearchParams(sp)),
    ])

    if (!me.roles.some(role => role.name === "admin")) {
        redirect("/")
    }
    return (
        <div className={`flex justify-center items-center min-h-screen min-w-screen gap-10`}>
            <div className={`animate-fadeIn mt-50 mb-20 flex flex-col gap-3 w-full max-w-2xl`}>
                <List t={t} prms={prms}  />
            </div>
        </div>
    )
}