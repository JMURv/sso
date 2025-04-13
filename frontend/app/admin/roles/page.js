import {getSessionToken} from "../../../lib/auth/token"
import {redirect} from "next/navigation"
import GetMe from "../../../lib/fetches/users/me"
import listRoles from "../../../lib/fetches/roles/list"
import List from "./List"

export default async function Page({searchParams}) {
    const t = await getSessionToken()
    if (!t) {
        redirect("/auth")
    }

    const sp = await searchParams
    const [me, rls] = await Promise.all([
        GetMe(t),
        listRoles(t, new URLSearchParams(sp)),
    ])

    if (!me.roles.some(role => role.name === "admin")) {
        redirect("/")
    }
    return (
        <div className={`flex justify-center items-center min-h-screen min-w-screen gap-10`}>
            <div className={`animate-fadeIn mt-50 mb-20 flex flex-col gap-3 w-full max-w-2xl`}>
                <List t={t} rls={rls}  />
            </div>
        </div>
    )
}