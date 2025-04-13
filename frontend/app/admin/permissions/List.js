"use client"
import {useEffect, useState} from "react"
import {toast} from "sonner"
import CroppedScreen from "../../../components/subscreen/CroppedScreen"
import {Add, Delete, Search} from "@mui/icons-material"
import ModalBase from "../../../components/modals/ModalBase"
import Pagination from "../../../components/pagination/Pagination"
import New from "./New"
import Edit from "./Edit"

export default function List({t, prms}) {
    const [q, setQ] = useState("")
    const [perms, setPerms] = useState(prms)
    const [areUSure, setAreUSure] = useState(false)

    const [openNewScreen, setOpenNewScreen] = useState(false)
    const [openedObjId, setOpenedObjId] = useState(null)
    const [objToDelete, setObjToDelete] = useState(null)

    useEffect(() => {
        setPerms(prms)
    }, [prms])

    const handleSearchChange = async (e) => {
        e.preventDefault()
        const {value} = e.target

        setQ(value)
        if (value.length < 3) {
            setPerms(prms)
            return
        }

        try {
            const r = await fetch(`/api/perm?search=${value}`, {
                method: "GET",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${t}`,
                },
                cache: "no-store",
            })

            if (!r.ok) {
                const data = await r.json()
                toast.error(data.errors)
                return []
            }

            setPerms(await r.json())
        } catch (e) {
            console.log(e)
            toast.error("Unexpected error")
        }
    }

    const removeObj = async (id) => {
        try {
            const r = await fetch(`/api/perm/${id}`, {
                method: "DELETE",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${t}`,
                }
            })

            if (!r.ok) {
                const data = await r.json()
                toast.error(data.errors)
                return []
            }

            setPerms(prev => ({
                ...prev,
                data: prev.data.filter(p => p.id !== id)
            }))
            setObjToDelete(null)
            setAreUSure(false)
            toast.success("Permission deleted")

        } catch (e) {
            console.log(e)
            toast.error("Unexpected error")
        }
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

    return (
        <div className={`max-w-2xl flex flex-col gap-5 justify-center items-center w-full`}>

            <CroppedScreen isOpen={openNewScreen} close={() => setOpenNewScreen(false)}>
                <div className={`w-full flex py-20 flex-col justify-center items-center`}>
                    <div className={`max-w-2xl w-full flex flex-col gap-5`}>
                        <New
                            t={t}
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
                                    t={t}
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
    )
}