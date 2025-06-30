
export async function GetMeCli() {
    try {
        const r = await fetch(`/api/users/me`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
            },
            cache: "no-store",
        })

        if (!r.ok) {
            const data = await r.json()
            return null
        }

        return await r.json()
    } catch (e) {
        console.error(e)
    }
}