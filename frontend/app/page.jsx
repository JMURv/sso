import {getRefreshToken, getSessionToken} from "../lib/auth/token"
import {redirect} from "next/navigation";
import getMyDevices from "../lib/fetches/devices/me"
import GetMe from "../lib/fetches/users/me"
import Main from "./Main"

export default async function Home() {
    const access = await getSessionToken()
    const refresh = await getRefreshToken()
    if (!access || !refresh) {
        redirect("/auth")
    }

    const [usr, device] = await Promise.all([
        GetMe(access),
        getMyDevices(access)
    ])

    return (
        <Main usr={usr} device={device} />
    )
}
