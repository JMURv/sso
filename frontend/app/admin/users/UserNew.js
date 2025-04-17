"use client"
import RolesModal from "../../../components/modals/RolesModal"
import {Add, Check, Close, Delete, KeyboardArrowDown, WbSunny} from "@mui/icons-material"
import AlternateEmailIcon from "@mui/icons-material/AlternateEmail"
import Image from "next/image"
import {useState} from "react"
import {toast} from "sonner"
import {DefaultUserImage} from "../../config"
import {useAuth} from "../../../providers/AuthProvider"


export default function UserNew({close, successCallback}) {
    const {authFetch} = useAuth()
    const [addRoleModal, setAddRoleModal] = useState(false)
    const [avatarFile, setAvatarFile] = useState(null)
    const [avatarPreview, setAvatarPreview] = useState(DefaultUserImage)
    const [user, setUser] = useState({
        name: "",
        email: "",
        password: "",
        is_active: false,
        is_email_verified: false,
        roles: [],
    })

    const handleUsrChange = async (e) => {
        const {name, type, value, checked} = e.target
        const newValue = type === "checkbox" ? checked : value
        const keys = name.split(".")
        setUser((prevFormData) => {
            let updatedFormData = {...prevFormData}

            let field = updatedFormData
            keys.forEach((key, index) => {
                if (index === keys.length - 1) {
                    field[key] = newValue
                } else {
                    field = field[key]
                }
            })
            return updatedFormData
        })
    }

    const onChooseRole = (role) => {
        if (user.roles.some(r => r.id === role.id)) {
            toast.error("User already has this role")
            return
        }

        setAddRoleModal(false)
        setUser({...user, roles: [...user.roles, role]})
    }

    const removeRole = async (id) => {
        setUser({...user, roles: user.roles.filter(role => role.id !== id)})
    }

    const createUser = async () => {
        const fd = new FormData()
        fd.append("avatar", avatarFile)
        fd.append("data", JSON.stringify({
            name: user.name,
            email: user.email,
            password: user.password,
            is_active: user.is_active === "true",
            is_email: user.is_email === "true",
            roles: user.roles.map(r => r.id),
        }))

        const response = await authFetch("/api/users/", {
            method: "POST",
            body: fd,
        })

        if (!response.ok) {
            const data = await r.json()
            toast.error(data.errors)
            return
        }

        const data = await response.json()
        user.id = data.id
        successCallback(user)
        toast.success("Create successful")
    }

    const handleFileUpload = (e) => {
        const file = e.target.files[0]
        if (file) {
            setAvatarFile(file)
            setAvatarPreview(URL.createObjectURL(file))
        }
    }

    return (
        <div className={`animate-fadeIn flex flex-row gap-5`}>
            <RolesModal
                isOpen={addRoleModal}
                setIsOpen={setAddRoleModal}
                onClick={(role) => onChooseRole(role)}
            />

            <div className={`w-full flex flex-col`}>
                <div className={`flex flex-row gap-5 mb-5`}>
                    <button onClick={() => setUser({...user, is_active: !user.is_active})}
                            className={`sec-b flex gap-5`}>
                        <WbSunny />
                        {user.is_active ? (
                            <div className={`text-green-400`}>
                                <Check />
                            </div>
                        ) : (
                            <div className={`text-red-400`}>
                                <Close />
                            </div>
                        )}
                    </button>
                    <button onClick={() => setUser({...user, is_email_verified: !user.is_email_verified})}
                            className={`sec-b flex gap-5`}>
                        <AlternateEmailIcon />
                        {user.is_email_verified ? (
                            <div className={`text-green-400`}>
                                <Check />
                            </div>
                        ) : (
                            <div className={`text-red-400`}>
                                <Close />
                            </div>
                        )}
                    </button>
                </div>


                <div className={`flex flex-row gap-5`}>
                    <div className={`p-3 gap-3 bg-zinc-900/70 ring-1 ring-zinc-800 text-zinc-100 w-full flex flex-col justify-between`}>
                        <div className={`flex flex-row flex-wrap gap-3`}>
                            <label htmlFor="avatar" className={`cursor-pointer`}>
                                <input id={`avatar`} type="file" onChange={handleFileUpload} className={`hidden`} />
                                <Image
                                    src={avatarPreview || DefaultUserImage}
                                    width={150}
                                    height={150}
                                    alt="new user avatar"
                                    className={`min-w-[100px] max-w-[100px] min-h-[100px] max-h-[100px] object-cover`}
                                />
                            </label>
                            <div className={`flex flex-col gap-3`}>
                                <input
                                    type="text"
                                    name={"name"}
                                    value={user.name}
                                    placeholder={"johndoe"}
                                    onChange={handleUsrChange}
                                    className={`outline-none text-sm`}
                                />

                                <input
                                    type="text"
                                    name={"email"}
                                    value={user.email}
                                    placeholder={"johndoe@example.com"}
                                    onChange={handleUsrChange}
                                    className={`outline-none text-sm`}
                                />


                                <input
                                    type="password"
                                    name={"password"}
                                    value={user.password}
                                    placeholder={"*********"}
                                    onChange={handleUsrChange}
                                    className={`outline-none font-medium text-md`}
                                />
                            </div>
                        </div>

                        <div className={`w-full h-px bg-zinc-700 my-3`} />

                        <div className={`mb-5`}>
                            <div className={`mb-2 flex justify-between items-center`}>
                                <h1 className={`text-sm tracking-widest uppercase`}>Roles</h1>
                                <button onClick={() => setAddRoleModal(true)} className={`cursor-pointer p-1`}>
                                    <Add />
                                </button>
                            </div>
                            <div className={`w-full grid grid-cols-2 gap-3`}>
                                {user.roles.length === 0 ? (
                                    <p className={`text-xs`}>No roles</p>
                                ) : (
                                    user.roles.map((r, i) => {
                                        const isLastAndOdd = user.roles.length % 2 !== 0 && i === user.roles.length - 1
                                        return (
                                            <div key={r.id}
                                                 className={`bg-zinc-950 ring-zinc-700 ${isLastAndOdd ? "col-span-2" : ""}`}>
                                                <div className="flex gap-5 justify-between items-center">
                                                    <p className="px-4 py-1 text-xs tracking-wider text-zinc-200 capitalize">{r.name}</p>
                                                    <button id={`remove-role`}
                                                            onClick={() => removeRole(r.id)}
                                                    >
                                                        <Delete fontSize={"small"} />
                                                    </button>
                                                </div>
                                            </div>
                                        )
                                    })
                                )}

                            </div>
                        </div>
                    </div>
                </div>

                <div className={`mt-5 flex flex-row gap-5`}>
                    <button onClick={() => close()} className={`primary-b w-full h-full`}>
                        <KeyboardArrowDown />
                    </button>
                    <button onClick={createUser} className={`primary-b w-full h-full`}>
                        <Check />
                    </button>
                </div>

            </div>
        </div>
    )
}