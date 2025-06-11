"use client"
import {useAuth} from "../../../providers/AuthProvider"
import {useEffect, useState} from "react"
import {toast} from "sonner"
import CroppedScreen from "../../../components/subscreen/CroppedScreen"
import New from "./New"
import {Add, Delete, Search} from "@mui/icons-material"
import Edit from "./Edit"
import ModalBase from "../../../components/modals/ModalBase"
import Pagination from "../../../components/pagination/Pagination"
import {useRouter, useSearchParams} from "next/navigation"

export default function Page() {
    const router = useRouter()
    const sp = useSearchParams()
    const {adminFetch, isAdmin} = useAuth()

    const [q, setQ] = useState("")
    const [roles, setRoles] = useState(null)
    const [areUSure, setAreUSure] = useState(false)

    const [openNewScreen, setOpenNewScreen] = useState(false)
    const [openedRoleId, setOpenedRoleId] = useState(null)
    const [roleToDelete, setRoleToDelete] = useState(null)


    const handleSearchChange = async (e) => {
        const { value } = e.target
        setQ(value)

        try {
            const params = new URLSearchParams({
                ...(value.length >= 3 && { search: value })
            })

            const response = await adminFetch(`/api/roles?${params}`)
            if (!response.ok) {
                const data = await response.json()
                toast.error(data.errors)
                return
            }

            setRoles(await response.json())
        } catch (error) {
            console.error(error)
            toast.error("Failed to search users")
        }
    }

    const removeRole = async (id) => {
        const response = await adminFetch(`/api/roles/${id}`, {
            method: "DELETE",
        })

        if (!response.ok) {
            const data = await response.json()
            toast.error(data.errors)
            return
        }

        setRoles(prev => ({
            ...prev,
            data: prev.data.filter(r => r.id !== id)
        }))
        setRoleToDelete(null)
        setAreUSure(false)
        toast.success("Role deleted")
    }

    const successCreateCallback = (rl) => {
        setRoles((prev) => ({
            ...prev,
            data: [rl, ...prev.data]
        }))
        setOpenNewScreen(false)
    }

    const successEditCallback = (rl) => {
        setRoles(prev => ({
            ...prev,
            data: prev.data.map(r => r.id === rl.id ? rl : r)
        }))
        setOpenedRoleId(false)
    }

    useEffect(() => {
        if (!isAdmin) {
            toast.error("You are not an admin")
            return router.push("/")
        }
        const fetchData = async () => {
            const params = new URLSearchParams(sp)
            const [roles] = await Promise.all([
                adminFetch(`/api/roles?${params}`, {
                    method: "GET",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    cache: "no-store",
                }),
            ])

            if (!roles.ok) {
                const data = await roles.json()
                console.log(data.errors)
                return null
            }
            const rls = await roles.json()
            setRoles(rls)
        }
        fetchData()
    }, [])

    if (!roles) return null
    return (
        <div className={`flex justify-center items-center min-h-screen min-w-screen gap-10`}>
            <div className={`animate-fadeIn mt-50 mb-20 flex flex-col gap-3 w-full max-w-2xl`}>
                <div className={`max-w-2xl flex flex-col gap-5 justify-center items-center w-full`}>

                    <CroppedScreen isOpen={openNewScreen} close={() => setOpenNewScreen(false)}>
                        <div className={`w-full flex py-20 flex-col justify-center items-center`}>
                            <div className={`max-w-2xl w-full flex flex-col gap-5`}>
                                <New
                                    close={() => setOpenNewScreen(false)}
                                    successCallback={successCreateCallback}
                                />
                            </div>
                        </div>
                    </CroppedScreen>

                    <div className={`w-full flex flex-col gap-1`}>
                        <span className={`flex text-xs tracking-widest text-zinc-400 flex justify-start w-full uppercase`}>{`// ADMIN`}</span>
                        <h1 className={`text-6xl tracking-widest text-zinc-800 dark:text-zinc-200 flex justify-start w-full uppercase`}>
                            Roles
                        </h1>
                    </div>

                    <div className={`flex flex-row w-full h-full items-center gap-5`}>
                        <div className={`icon-input-wrapper`}>
                            <div className={`icon-container`}>
                                <Search fontSize={"medium"} />
                            </div>
                            <input
                                type="text"
                                name={"search"}
                                value={q}
                                placeholder={"Developer"}
                                onChange={handleSearchChange}
                                className={`icon-input`}
                            />
                        </div>

                        <button className={`primary-b`} onClick={() => setOpenNewScreen(true)}>
                            <Add />
                        </button>
                    </div>

                    {roles.data.map((r) => (
                        <div key={r.id} className={`w-full flex flex-row gap-5`}>
                            <CroppedScreen
                                isOpen={openedRoleId === r.id}
                                close={() => setOpenedRoleId(null)}
                            >
                                <div className={`w-full flex py-20 flex-col justify-center items-center`}>
                                    <div className={`max-w-2xl w-full flex flex-col gap-5`}>
                                        <Edit
                                            rl={r}
                                            close={() => setOpenedRoleId(false)}
                                            successCallback={successEditCallback}
                                        />
                                    </div>
                                </div>
                            </CroppedScreen>

                            <ModalBase isOpen={areUSure} setIsOpen={setAreUSure}>
                                <div className={`flex flex-col gap-3 bg-zinc-950 p-5`}>
                                    <div className={`flex gap-3 w-full justify-between items-center`}>
                                        <p className={`flex flex-col`}>
                                            Are you sure?
                                            <span className={`text-zinc-500 text-xs`}>
                                        this action cannot be undone
                                    </span>
                                        </p>
                                    </div>

                                    <div className={`flex gap-3 w-full`}>
                                        <button onClick={() => removeRole(roleToDelete)} className={`w-full primary-b flex justify-center items-center`}>
                                            Yes
                                        </button>
                                        <button onClick={() => setAreUSure(false)}
                                                className={`w-full primary-b flex justify-center items-center`}>
                                            No
                                        </button>
                                    </div>
                                </div>
                            </ModalBase>

                            <div onClick={() => setOpenedRoleId(r.id)} className={`p-3 gap-3 bg-zinc-900/90 hover:bg-zinc-800 hover:scale-99 ring-1 ring-zinc-700 text-zinc-100 w-full flex flex-col justify-between duration-200 cursor-pointer`}>

                                <div className={`flex flex-row gap-3`}>
                                    <div className={`flex flex-col gap-1`}>
                                        <p className={`text-zinc-500 text-xs`}>{r.id}</p>
                                        <p className={`text-md`}>{r.name}</p>
                                        <p className={`text-sm`}>{r.description}</p>
                                    </div>
                                </div>

                                <div className={`w-full grid grid-cols-2 gap-3`}>
                                    {r?.permissions.length > 0 && (
                                        r.permissions.slice(0, 4).map((p) => (
                                            <div key={p.id}
                                                 className={`ring-2 ring-zinc-500`}>
                                                <div className="px-4 py-1">
                                                    <p className="text-xs tracking-wider text-zinc-200 capitalize">{p.name}</p>
                                                    <p className="text-xs tracking-wider text-zinc-400 capitalize">{p.description}</p>
                                                </div>
                                            </div>
                                        ))
                                    )}
                                </div>
                            </div>

                            <div className={`flex flex-col gap-5`}>
                                <button onClick={() => {
                                    setAreUSure(true)
                                    setRoleToDelete(r.id)
                                }} className={`sec-b h-full hover:scale-99`}>
                                    <Delete />
                                </button>
                            </div>
                        </div>
                    ))}

                    <Pagination
                        currentPage={roles.current_page}
                        totalPages={roles.total_pages}
                        hasNextPage={roles.has_next_page}
                    />
                </div>
            </div>
        </div>
    )
}