import {useEffect, useState} from "react"
import ModalBase from "./ModalBase"
import {toast} from "sonner"
import {Search} from "@mui/icons-material"

export default function PermModal({isOpen, setIsOpen, onClick}) {
    const [q, setQ] = useState("")
    const [perms, setPerms] = useState()
    const [initialPerms, setInitialPerms] = useState()

    const handleSearchChange = async (e) => {
        e.preventDefault()
        const {value} = e.target

        setQ(value)
        if (value.length < 3) {
            setPerms(initialPerms)
            return
        }

        try {
            const r = await fetch(`/api/perm?search=${value}`, {
                method: "GET",
                headers: {
                    "Content-Type": "application/json",
                },
                cache: "no-store",
            })

            if (!r.ok) {
                const data = await r.json()
                toast.error(data.errors)
                return
            }

            setPerms(await r.json())
        } catch (e) {
            console.log(e)
            toast.error("Unexpected error")
        }
    }

    useEffect(() => {
        const fetchRoles = async () => {
            const r = await fetch("/api/perm", {
                method: "GET",
                headers: {
                    "Content-Type": "application/json",
                },
                cache: "no-store",
            })

            if (!r.ok) {
                const data = await r.json()
                toast.error(data.errors)
                return []
            }
            return await r.json()
        }
        fetchRoles().then((r) => {
            setPerms(r)
            setInitialPerms(r)
        })
    }, [])
    return (
        <ModalBase isOpen={isOpen} setIsOpen={setIsOpen}>
            <div className={`w-full md:min-w-md max-w-xs bg-zinc-950 flex flex-col gap-3 p-5`}>
                <div className={`flex flex-row w-full h-full items-center gap-5`}>
                    <div className={`icon-input-wrapper`}>
                        <div className={`icon-container`}>
                            <Search fontSize={"medium"} />
                        </div>
                        <input
                            type="text"
                            name={"search"}
                            value={q}
                            placeholder={"developer"}
                            onChange={handleSearchChange}
                            className={`icon-input`}
                        />
                    </div>
                </div>

                <div className={`grid grid-cols-1 gap-2`}>
                    {perms?.data?.length > 0 && (
                        perms.data.map((role) => {
                            return (
                                <div onClick={() => onClick(role)}
                                     className={`cursor-pointer flex flex-row bg-zinc-900 hover:bg-zinc-800 items-center gap-2 px-4 py-2 h-full`}
                                     key={role.id}>
                                    <p className={`text-sm text-zinc-300 capitalize`}>
                                        {role.name}
                                    </p>
                                </div>
                            )
                        })
                    )}

                </div>
            </div>
        </ModalBase>
    )
}