export default async function listPermissions(t, params) {
    try {
        const r = await fetch(`${process.env.BACKEND_URL}/api/perm?${params}`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${t}`,
            },
            cache: "no-store",
        })

        if (!r.ok) {
            const data = await r.json()
            console.log(data.errors)
            return null
        }

        return await r.json()
    } catch (e) {
        console.error(e)
    }
}