"use client"
import {useRouter, useSearchParams} from "next/navigation"
import {useAuth} from "../../../providers/AuthProvider"
import {useEffect, useState} from "react"
import {toast} from "sonner"
import CroppedScreen from "../../../components/subscreen/CroppedScreen"
import UserNew from "./UserNew"
import SidebarBase from "../../../components/sidebars/SidebarBase"
import {Add, Close, CropSquare, Delete, FilterList, KeyboardArrowRight, Search, Sort, Square} from "@mui/icons-material"
import Dropdown from "../../../components/dropdowns/Dropdown"
import UserEdit from "./UserEdit"
import ModalBase from "../../../components/modals/ModalBase"
import Image from "next/image"
import Pagination from "../../../components/pagination/Pagination"
import {DefaultUserImage} from "../../config"

export default function Page() {
    const router = useRouter()
    const sp = useSearchParams()
    const {adminFetch, isAdmin} = useAuth()

    const [q, setQ] = useState("")
    const [filterURL, setFilterURL] = useState("")
    const [users, setUsers] = useState(null)
    const [roles, setRoles] = useState(null)
    const [areUSure, setAreUSure] = useState(false)
    const [openFilters, setOpenFilters] = useState(false)
    const [openSort, setOpenSort] = useState(false)
    const [openRoles, setOpenRoles] = useState(false)

    const [openNewUsrScreen, setOpenNewUsrScreen] = useState(false)

    const [openedUserId, setOpenedUserId] = useState(null)
    const [userToDelete, setUserToDelete] = useState(null)

    const handleSearchChange = async (e) => {
        const { value } = e.target
        setQ(value)

        try {
            const params = new URLSearchParams({
                ...(value.length >= 3 && { search: value })
            })

            const response = await adminFetch(`/api/users?${params}`)
            if (!response.ok) {
                const data = await response.json()
                toast.error(data.errors)
                return
            }

            setUsers(await response.json())
        } catch (error) {
            toast.error("Failed to search users")
        }
    }

    const handleApplyFilters = async (e) => {
        e.preventDefault()
        window.location.href = filterURL
    }

    const handleClearFilters = async() => {
        setFilterURL("")
        await router.push("?")
        window.location.href = "?"
    }

    const handleSortChange = async (e, val) => {
        const url = new URLSearchParams(sp.toString())
        if (url.has("sort")) {
            if (url.get("sort") === val) {
                url.delete("sort")
                setFilterURL(`?${url.toString()}`)
                await router.push(`?${url.toString()}`)
                return
            }
        }

        url.set("sort", val)
        url.delete("page")
        setFilterURL(`?${url.toString()}`)
        await router.push(`?${url.toString()}`)
    }

    const handleFilterChange = async (e, val) => {
        const url = new URLSearchParams(sp.toString())
        if (val.startsWith("is_")) {
            if (url.has(val)) {
                url.delete(val)
            } else {
                url.set(val, "true")
            }
        }
        else {
            const currentRoles = url.has("roles") ? url.get("roles").split(",") : []

            const roleIndex = currentRoles.indexOf(val);
            if (roleIndex > -1) {
                currentRoles.splice(roleIndex, 1)
            } else {
                currentRoles.push(val)
            }

            if (currentRoles.length > 0) {
                url.set("roles", currentRoles.join(","))
            } else {
                url.delete("roles")
            }
        }

        url.delete("page")
        setFilterURL(`?${url.toString()}`)
        await router.push(`?${url.toString()}`)
    }

    const removeUser = async (id) => {
        const response = await adminFetch(`/api/users/${id}`, {
            method: "DELETE",
        })

        if (!response.ok) {
            const data = await response.json()
            toast.error(data.errors)
            return
        }

        setUsers(prev => ({
            ...prev,
            data: prev.data.filter(u => u.id !== id)
        }))
        setUserToDelete(null)
        setAreUSure(false)
        toast.success("User deleted")
    }

    const successCreateCallback = (usr) => {
        setUsers((prev) => ({
            ...prev,
            data: [usr, ...prev.data]
        }))
        setOpenNewUsrScreen(false)
    }

    const successEditCallback = (usr) => {
        setUsers(prev => ({
            ...prev,
            data: prev.data.map(u => u.id === usr.id ? usr : u)
        }))
        setOpenedUserId(null)
    }

    useEffect(() => {
        if (!isAdmin) {
            toast.error("You are not an admin")
            return router.push("/")
        }
        const fetchData = async () => {
            const params = new URLSearchParams(sp)
            const [users, roles] = await Promise.all([
                adminFetch(`/api/users?${params}`, {
                    method: "GET",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    cache: "no-store",
                }),
                adminFetch(`/api/roles`, {
                    method: "GET",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    cache: "no-store",
                }),
            ])

            if (!users.ok) {
                const data = await users.json()
                console.log(data.errors)
                return null
            }
            const usrs = await users.json()
            setUsers(usrs)

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

    if (!users || !roles) return null

    return (
        <div className={`flex justify-center items-center min-h-screen min-w-screen gap-10`}>
            <div className={`animate-fadeIn mt-50 mb-20 flex flex-col gap-3 w-full max-w-2xl`}>
                <div className={`max-w-2xl flex flex-col gap-5 justify-center items-center w-full`}>

                    <CroppedScreen isOpen={openNewUsrScreen} close={() => setOpenNewUsrScreen(false)}>
                        <div className={`w-full flex py-20 flex-col justify-center items-center`}>
                            <div className={`max-w-2xl w-full flex flex-col gap-5`}>
                                <UserNew
                                    close={() => setOpenNewUsrScreen(false)}
                                    successCallback={successCreateCallback}
                                />
                            </div>
                        </div>
                    </CroppedScreen>

                    <SidebarBase
                        show={openFilters}
                        toggle={() => setOpenFilters(false)}
                        classes={"right-0 bg-zinc-950 w-full border-1 border-zinc-800 border-l max-w-sm z-50"}
                    >
                        <div className={`flex flex-col items-start p-4`}>
                            <div className={`flex w-full justify-between items-center`}>
                                <p className={`text-md tracking-widest`}>Filters</p>
                                <Close />
                            </div>

                            <div className={`w-full h-px my-5 bg-zinc-800`}/>

                            <div className={`mb-5 w-full flex flex-col gap-5`}>
                                <button
                                    onClick={(e) => handleFilterChange(e, "is_active")}
                                    className={`w-full uppercase text-xs flex flex-row gap-3 items-center cursor-pointer bg-zinc-950/80 border-1 border-zinc-200 dark:border-zinc-700 px-4 py-3 text-zinc-50 hover:bg-zinc-950 duration-200`}
                                >
                                    {sp.has("is_active") && sp.get("is_active") ? (
                                        <Square />
                                    ) : (
                                        <CropSquare />
                                    )}
                                    active
                                </button>

                                <button
                                    onClick={(e) => handleFilterChange(e, "is_email_verified")}
                                    className={`w-full uppercase text-xs flex flex-row gap-3 items-center cursor-pointer bg-zinc-950/80 border-1 border-zinc-200 dark:border-zinc-700 px-4 py-3 text-zinc-50 hover:bg-zinc-950 duration-200`}
                                >
                                    {sp.has("is_email_verified") && sp.get("is_email_verified") ? (
                                        <Square />
                                    ) : (
                                        <CropSquare />
                                    )}
                                    verified email
                                </button>

                                <button
                                    onClick={(e) => handleFilterChange(e, "is_wa")}
                                    className={`w-full uppercase text-xs flex flex-row gap-3 items-center cursor-pointer bg-zinc-950/80 border-1 border-zinc-200 dark:border-zinc-700 px-4 py-3 text-zinc-50 hover:bg-zinc-950 duration-200`}
                                >
                                    {sp.has("is_wa") && sp.get("is_wa") ? (
                                        <Square />
                                    ) : (
                                        <CropSquare />
                                    )}
                                    WebAuthN
                                </button>
                            </div>

                            <div className={`flex flex-col gap-3 relative w-full`}>

                                <Dropdown isOpen={openRoles} setIsOpen={setOpenRoles} classes={`top-full w-full origin-top-right`}>
                                    <div className={`flex flex-col w-full`}>
                                        {roles.data.map((r) => (
                                            <button
                                                key={r.name}
                                                onClick={(e) => handleFilterChange(e, r.name)}
                                                className={`uppercase text-xs flex flex-row justify-between items-center cursor-pointer bg-zinc-950/80 border-b border-x border-zinc-200 dark:border-zinc-700 px-4 py-3 text-zinc-50 hover:bg-zinc-950 duration-200`}
                                            >
                                                {sp.has("roles") && sp.get("roles").includes(r.name) ? (
                                                    <Square />
                                                ) : (
                                                    <CropSquare />
                                                )}
                                                {r.name}
                                                <KeyboardArrowRight />
                                            </button>
                                        ))}
                                    </div>
                                </Dropdown>

                                <button onClick={() => setOpenRoles(true)} className={`sec-b w-full`}>
                                    Roles
                                </button>
                            </div>

                            <div className={`w-full flex gap-5 mt-auto`}>
                                <button onClick={handleClearFilters} className={`primary-b w-full`}>
                                    Reset
                                </button>
                                <button onClick={handleApplyFilters} className={`primary-b w-full`}>
                                    Apply
                                </button>
                            </div>
                        </div>
                    </SidebarBase>

                    <SidebarBase
                        show={openSort}
                        toggle={() => setOpenSort(false)}
                        classes={"right-0 bg-zinc-950 w-full border-1 border-zinc-800 border-l max-w-sm z-50"}>
                        <div className={`flex flex-col items-start p-4`}>
                            <div className={`flex w-full justify-between items-center`}>
                                <p className={`text-md tracking-widest`}>Sort</p>
                                <Close />
                            </div>

                            <div className={`w-full h-px my-5 bg-zinc-800`}/>

                            <div className={`flex flex-col gap-3 relative w-full`}>

                                <div className={`flex flex-col gap-5 w-full`}>
                                    <button
                                        onClick={(e) => handleSortChange(e, "created_at")}
                                        className={`uppercase text-xs flex flex-row justify-between items-center cursor-pointer bg-zinc-950/80 border border-zinc-200 dark:border-zinc-700 px-4 py-3 text-zinc-50 hover:bg-zinc-950 duration-200`}
                                    >
                                        <div className={`flex gap-3 items-center`}>
                                            {sp.has("sort") && sp.get("sort") === "created_at" ? (
                                                <Square />
                                            ) : (
                                                <CropSquare />
                                            )}
                                            {`Дата создания по возрастанию`}
                                        </div>
                                        <KeyboardArrowRight />
                                    </button>
                                    <button
                                        onClick={(e) => handleSortChange(e, "-created_at")}
                                        className={`uppercase text-xs flex flex-row justify-between items-center cursor-pointer bg-zinc-950/80 border border-zinc-200 dark:border-zinc-700 px-4 py-3 text-zinc-50 hover:bg-zinc-950 duration-200`}
                                    >
                                        <div className={`flex gap-3 items-center`}>
                                            {sp.has("sort") && sp.get("sort") === "-created_at" ? (
                                                <Square />
                                            ) : (
                                                <CropSquare />
                                            )}
                                            {`Дата создания по убыванию`}
                                        </div>
                                        <KeyboardArrowRight />
                                    </button>
                                </div>
                            </div>


                            <div className={`w-full flex gap-5 mt-auto`}>
                                <button onClick={handleClearFilters} className={`primary-b w-full`}>
                                    Reset
                                </button>
                                <button onClick={handleApplyFilters} className={`primary-b w-full`}>
                                    Apply
                                </button>
                            </div>
                        </div>
                    </SidebarBase>

                    <div className={`w-full flex flex-col gap-1`}>
                        <span className={`flex text-xs tracking-widest text-zinc-400 flex justify-start w-full uppercase`}>{`// ADMIN`}</span>
                        <h1 className={`text-6xl tracking-widest text-zinc-800 dark:text-zinc-200 flex justify-start w-full uppercase`}>
                            Users
                        </h1>
                    </div>

                    <div className={`flex flex-row w-full h-full items-center gap-5`}>
                        <button onClick={() => setOpenFilters(true)} className={`sec-b`}>
                            <FilterList />
                        </button>

                        <button onClick={() => setOpenSort(true)} className={`sec-b`}>
                            <Sort />
                        </button>

                        <div className={`icon-input-wrapper`}>
                            <div className={`icon-container`}>
                                <Search fontSize={"medium"} />
                            </div>
                            <input
                                type="text"
                                name={"search"}
                                value={q}
                                placeholder={"johndoe"}
                                onChange={handleSearchChange}
                                className={`icon-input`}
                            />
                        </div>

                        <button className={`primary-b`} onClick={() => setOpenNewUsrScreen(true)}>
                            <Add />
                        </button>
                    </div>

                    {users.data.map((u) => (
                        <div key={u.id} className={`w-full flex flex-row gap-5`}>
                            <CroppedScreen
                                isOpen={openedUserId === u.id}
                                close={() => setOpenedUserId(null)}
                            >
                                <div className={`w-full flex py-20 flex-col justify-center items-center`}>
                                    <div className={`max-w-2xl w-full flex flex-col gap-5`}>
                                        <UserEdit
                                            usr={u}
                                            close={() => setOpenedUserId(null)}
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
                                        <button onClick={() => removeUser(userToDelete)} className={`w-full primary-b flex justify-center items-center`}>
                                            Yes
                                        </button>
                                        <button onClick={() => setAreUSure(false)}
                                                className={`w-full primary-b flex justify-center items-center`}>
                                            No
                                        </button>
                                    </div>
                                </div>
                            </ModalBase>

                            <div onClick={() => setOpenedUserId(u.id)} className={`p-3 gap-3 bg-zinc-900/90 hover:bg-zinc-800 hover:scale-99 ring-1 ring-zinc-700 text-zinc-100 w-full flex flex-col justify-between duration-200 cursor-pointer`}>

                                <div className={`flex flex-row gap-3`}>
                                    <Image src={DefaultUserImage} alt={`${u.name}`} width={75} height={75} className={`object-cover`} />
                                    <div className={`flex flex-col gap-1`}>
                                        <p className={`text-md`}>{u.name}</p>
                                        <p className={`text-sm`}>{u.email}</p>
                                        <p className={`text-zinc-500 text-xs`}>{u.id}</p>
                                    </div>
                                </div>

                                <div className={`w-full flex flex-wrap gap-3`}>
                                    {u.roles.slice(0, 4).map((r) => (
                                        <div key={r.id}
                                             className={`ring-2 ring-zinc-500`}>
                                            <div className="px-4 py-1">
                                                <p className="text-xs tracking-wider text-zinc-200 capitalize">{r.name}</p>
                                            </div>
                                        </div>
                                    ))}
                                </div>

                            </div>

                            <div className={`flex flex-col gap-5`}>
                                <button onClick={() => {
                                    setAreUSure(true)
                                    setUserToDelete(u.id)
                                }} className={`sec-b h-full hover:scale-99`}>
                                    <Delete />
                                </button>
                            </div>
                        </div>
                    ))}

                    <Pagination
                        currentPage={users.current_page}
                        totalPages={users.total_pages}
                        hasNextPage={users.has_next_page}
                    />
                </div>
            </div>
        </div>
    )
}