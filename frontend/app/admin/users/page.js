import {redirect} from "next/navigation"
import {getSessionToken} from "../../../lib/auth/token"
import GetMe from "../../../lib/fetches/users/me"
import listUsers from "../../../lib/fetches/users/list"
import UserAdm from "./UserAdm"
import listRoles from "../../../lib/fetches/roles/list"

export default async function Page({searchParams}) {
    const t = await getSessionToken()
    if (!t) {
        redirect("/auth")
    }

    const sp = await searchParams
    const [me, users, roles] = await Promise.all([
        GetMe(t),
        listUsers(t, new URLSearchParams(sp)),
        listRoles(t),
    ])

    if (!me.roles.some(role => role.name === "admin")) {
        redirect("/")
    }

    return (
        <div className={`flex justify-center items-center min-h-screen min-w-screen gap-10`}>
            <div className={`animate-fadeIn mt-50 mb-20 flex flex-col gap-3 w-full max-w-2xl`}>
                <UserAdm t={t} usrs={users} roles={roles}  />
            </div>
        </div>
    )
}