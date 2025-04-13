import {getSessionToken} from "@/lib/auth/token";
import {redirect} from "next/navigation";
import getMyDevices from "../lib/fetches/devices/me"
import GetMe from "../lib/fetches/users/me"
import Main from "./Main"

export default async function Home() {
    const t = await getSessionToken()
    if (!t) {
        redirect("/auth")
    }

    const [usr, device] = await Promise.all([
        GetMe(t),
        getMyDevices(t)
    ])

    return (
        <Main t={t} usr={usr} device={device} />
    )
}
