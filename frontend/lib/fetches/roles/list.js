export default async function listRoles(t, params) {
    try {
        const r = await fetch(`${process.env.BACKEND_URL}/api/roles?${params}`, {
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