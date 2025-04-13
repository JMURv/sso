export default async function DeleteDevice(t) {
    try {
        const r = await fetch(`${process.env.BACKEND_URL}/api/device`, {
            method: "DELETE",
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