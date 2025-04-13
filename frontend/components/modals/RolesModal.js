import {useEffect, useState} from "react"
import ModalBase from "./ModalBase"
import {toast} from "sonner"
import {Search} from "@mui/icons-material"

export default function RolesModal({isOpen, setIsOpen, onClick}) {
    const [q, setQ] = useState("")
    const [roles, setRoles] = useState()
    const [initialRoles, setInitialRoles] = useState()

    const handleSearchChange = async (e) => {
        e.preventDefault()
        const {value} = e.target

        setQ(value)
        if (value.length < 3) {
            setRoles(initialRoles)
            return
        }

        try {
            const r = await fetch(`/api/roles?search=${value}`, {
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

            setRoles(await r.json())
        } catch (e) {
            console.log(e)
            toast.error("Unexpected error")
        }
    }

    useEffect(() => {
        const fetchRoles = async () => {
            const r = await fetch("/api/roles", {
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
            setRoles(r)
            setInitialRoles(r)
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
                    {roles?.data?.length > 0 && (
                        roles.data.map((role) => {
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