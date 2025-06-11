"use client"
import {useAuth} from "../../../providers/AuthProvider"
import {useEffect, useState} from "react"
import {toast} from "sonner"
import {useRouter, useSearchParams} from "next/navigation"
import CroppedScreen from "../../../components/subscreen/CroppedScreen"
import New from "./New"
import {Add, Delete, Search} from "@mui/icons-material"
import Edit from "./Edit"
import ModalBase from "../../../components/modals/ModalBase"
import Pagination from "../../../components/pagination/Pagination"

export default function Page() {
    const router = useRouter()
    const sp = useSearchParams()
    const {adminFetch, isAdmin} = useAuth()

    const [q, setQ] = useState("")
    const [perms, setPerms] = useState(null)
    const [areUSure, setAreUSure] = useState(false)

    const [openNewScreen, setOpenNewScreen] = useState(false)
    const [openedObjId, setOpenedObjId] = useState(null)
    const [objToDelete, setObjToDelete] = useState(null)

    const handleSearchChange = async (e) => {
        const { value } = e.target
        setQ(value)

        try {
            const params = new URLSearchParams({
                ...(value.length >= 3 && { search: value })
            })

            const response = await adminFetch(`/api/perm?${params}`)
            if (!response.ok) {
                const data = await response.json()
                toast.error(data.errors)
                return
            }

            setPerms(await response.json())
        } catch (error) {
            console.error(error)
            toast.error("Failed to search users")
        }
    }

    const removeObj = async (id) => {
        const response = await adminFetch(`/api/perm/${id}`, {
            method: "DELETE",
        })

        if (!response.ok) {
            const data = await response.json()
            toast.error(data.errors)
            return
        }

        setPerms(prev => ({
            ...prev,
            data: prev.data.filter(p => p.id !== id)
        }))
        setObjToDelete(null)
        setAreUSure(false)
        toast.success("Permission deleted")
    }

    const successCreateCallback = (obj) => {
        setPerms((prev) => ({
            ...prev,
            data: [obj, ...prev.data]
        }))
        setOpenNewScreen(false)
    }

    const successEditCallback = (obj) => {
        setPerms(prev => ({
            ...prev,
            data: prev.data.map(p => p.id === obj.id ? obj : p)
        }))
        setOpenedObjId(false)
    }

    useEffect(() => {
        if (!isAdmin) {
            toast.error("You are not an admin")
            return router.push("/")
        }
        const fetchData = async () => {
            const params = new URLSearchParams(sp)
            const [perms] = await Promise.all([
                adminFetch(`/api/perm?${params}`, {
                    method: "GET",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    cache: "no-store",
                }),
            ])

            if (!perms.ok) {
                const data = await perms.json()
                console.log(data.errors)
                return null
            }
            const prms = await perms.json()
            setPerms(prms)
        }
        fetchData()
    }, [])

    if (!perms) return null
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
                            Permissions
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

                    {perms.data.map((obj) => (
                        <div key={obj.id} className={`w-full flex flex-row gap-5`}>
                            <CroppedScreen
                                isOpen={openedObjId === obj.id}
                                close={() => setOpenedObjId(null)}
                            >
                                <div className={`w-full flex py-20 flex-col justify-center items-center`}>
                                    <div className={`max-w-2xl w-full flex flex-col gap-5`}>
                                        <Edit
                                            srvOBJ={obj}
                                            close={() => setOpenedObjId(false)}
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
                                        <button onClick={() => removeObj(objToDelete)} className={`w-full primary-b flex justify-center items-center`}>
                                            Yes
                                        </button>
                                        <button onClick={() => setAreUSure(false)}
                                                className={`w-full primary-b flex justify-center items-center`}>
                                            No
                                        </button>
                                    </div>
                                </div>
                            </ModalBase>

                            <div onClick={() => setOpenedObjId(obj.id)} className={`p-3 gap-3 bg-zinc-900/90 hover:bg-zinc-800 hover:scale-99 ring-1 ring-zinc-700 text-zinc-100 w-full flex flex-col justify-between duration-200 cursor-pointer`}>
                                <div className={`flex flex-row gap-3`}>
                                    <div className={`flex flex-col gap-1`}>
                                        <p className={`text-zinc-500 text-xs`}>{obj.id}</p>
                                        <p className={`text-md`}>{obj.name}</p>
                                        <p className={`text-xs text-zinc-500`}>{obj.description}</p>
                                    </div>
                                </div>
                            </div>

                            <div className={`flex flex-col gap-5`}>
                                <button onClick={() => {
                                    setAreUSure(true)
                                    setObjToDelete(obj.id)
                                }} className={`sec-b h-full hover:scale-99`}>
                                    <Delete />
                                </button>
                            </div>
                        </div>
                    ))}

                    <Pagination
                        currentPage={perms.current_page}
                        totalPages={perms.total_pages}
                        hasNextPage={perms.has_next_page}
                    />
                </div>
            </div>
        </div>
    )
}