"use client"
import {Check, Close, Fingerprint, Logout, Person} from "@mui/icons-material"
import Image from "next/image"
import {useEffect, useState} from "react"
import {toast} from "sonner"
import WAModal from "../components/modals/WAModal"
import ModalBase from "../components/modals/ModalBase"
import Oauth2Conns from "../components/oauth2/Oauth2Conns"
import DeviceList from "../components/ui/DeviceList"
import {useAuth} from "../providers/AuthProvider"

const DefaultUserImage = "/defaults/user.png"

const gradientDirection = [
    "bg-gradient-to-r",
    "bg-gradient-to-l",
    "bg-gradient-to-t",
    "bg-gradient-to-b",
    "bg-gradient-to-tr"
]
const gradientClasses = [
    "from-pink-500 to-yellow-500",
    "from-blue-500 to-green-400",
    "from-purple-500 to-indigo-500",
    "from-orange-400 to-amber-500",
    "from-sky-500 to-indigo-500"
]

export default function Page() {
    const { authFetch, me, setMe, logout } = useAuth()
    const [devices, setDevices] = useState(null)
    const [isWA, setIsWA] = useState(false)
    const [isProfile, setIsProfile] = useState(false)
    const [avatarPreview, setAvatarPreview] = useState(me?.avatar || DefaultUserImage)
    const [avatarFile, setAvatarFile] = useState(null)

    const updateUser = async () => {
        const fd = new FormData()
        fd.append("avatar", avatarFile)
        fd.append("data", JSON.stringify({
            name: me.name,
            email: me.email,
            is_active: me.is_active === "true",
            is_email: me.is_email === "true"
        }))

        const response = await authFetch(`/api/users/me`, {
            method: "PUT",
            body: fd,
        })

        if (!response.ok) {
            const data = await response.json()
            toast.error(data.errors)
            return
        }

        setIsProfile(false)
        toast.success("Update successful")
    }

    const onFormDataChange = async (e) => {
        setMe({...me, [e.target.name]: e.target.value})
    }

    const handleFileUpload = (e) => {
        const file = e.target.files[0]
        if (file) {
            setAvatarFile(file)
            setAvatarPreview(URL.createObjectURL(file))
        }
    }

    const successWARegisterCallback = () => {
        setMe({...me, is_wa: true})
    }

    useEffect(() => {
        if (me?.avatar) {
            setAvatarPreview("/"+me.avatar)
        }
    }, [me?.avatar])

    useEffect(() => {
        const fetch = async () => {
            const [device] = await Promise.all([
                authFetch(`/api/device`, {
                    method: "GET",
                    cache: "no-store",
                }),
            ])

            if (device) {
                const deviceData = await device.json()
                setDevices(deviceData)
            }
        }
        fetch()
    }, [])

    useEffect(() => {}, [me])
    return (
        <div className={`flex justify-center items-center min-h-screen min-w-screen gap-10`}>
            {me ? (
             <>
                 <WAModal isWA={isWA} setIsWA={setIsWA} callback={successWARegisterCallback} />
                 <ModalBase isOpen={isProfile} setIsOpen={setIsProfile}>
                     <div className={`flex flex-col bg-zinc-950 gap-5 p-5 max-w-xs`}>

                         <div className={`flex flex-col`}>
                             <div className={`mb-2 flex justify-between items-center`}>
                                 <h2 className={`text-xl`}>Profile</h2>
                                 <button className={`cursor-pointer`} onClick={() => setIsProfile(false)}>
                                     <Close fontSize={"small"} />
                                 </button>
                             </div>
                             <p className={`text-zinc-400 text-sm`}>
                                 You can update your profile information here
                             </p>
                         </div>

                         <div className={`flex flex-col gap-5`}>
                             <div className={`flex flex-col gap-3`}>
                                 <div className="flex flex-col gap-3">
                                     <label htmlFor="avatar" className={`cursor-pointer`}>
                                         <input id={`avatar`} type="file" onChange={handleFileUpload} className={`hidden`} />
                                         <Image
                                             src={avatarPreview || DefaultUserImage}
                                             width={300}
                                             height={300}
                                             alt="my avatar"
                                             className={`rounded-sm object-cover w-full aspect-square`}
                                         />
                                     </label>

                                     <div className={`flex flex-col gap-1`}>
                                         <input
                                             type={"text"}
                                             name={`name`}
                                             value={me.name}
                                             onChange={onFormDataChange}
                                             className={`outline-none font-medium text-md`}
                                         />

                                         <p className={`font-medium text-sm text-zinc-500`}>{me.email}</p>
                                     </div>
                                 </div>
                             </div>
                         </div>

                         <div className={`flex flex-row gap-3`}>
                             <button onClick={updateUser} className={`w-full primary-b flex justify-center items-center`}>
                                 <Check />
                             </button>
                             <button onClick={() => setIsProfile(false)}
                                     className={`w-full primary-b flex justify-center items-center`}>
                                 <Close />
                             </button>
                         </div>

                     </div>
                 </ModalBase>

                 <div className={`animate-fadeIn mt-50 mb-20 flex flex-col gap-3 w-full max-w-2xl`}>
                     <h1 className={`text-5xl`}>
                         Hello, {me.name}
                     </h1>

                     <div className={`flex justify-between items-center`}>
                         <p className={`text-xl`}>
                             This is your authorization hub
                         </p>
                         <div className={`flex gap-5`}>
                             {!me.is_wa && (
                                 <button onClick={() => setIsWA(true)} className={`text-zinc-300 cursor-pointer`}>
                                     <Fingerprint />
                                 </button>
                             )}

                             <button onClick={() => setIsProfile(true)} className={`text-zinc-300 cursor-pointer`}>
                                 <Person />
                             </button>

                             <button onClick={logout} className={`text-zinc-300 cursor-pointer`}>
                                 <Logout />
                             </button>
                         </div>
                     </div>

                     <div className={`h-[2px] w-full bg-zinc-800`} />

                     <div className={`mb-5 flex flex-col gap-3`}>
                         <div>
                             <h2>You have following roles</h2>
                             <span className={`font-medium text-xs text-zinc-500`}>
                            To change your roles, please contact your administrator
                        </span>
                         </div>
                         <div className="flex flex-wrap gap-3">
                             {me.roles.length === 0 ? (
                                 <p className={`text-xs`}>No roles</p>
                             ) : (
                                 me.roles.map((rr, i) => {
                                     const gradient = gradientClasses[i % gradientClasses.length]
                                     const direction = gradientDirection[i % gradientDirection.length]
                                     return (
                                         <div key={rr.id}
                                              className={`p-[2px] ${direction} ${gradient}`}>
                                             <div className="bg-zinc-950 px-4 py-1">
                                                 <p className="text-sm font-normal tracking-wider text-zinc-200 capitalize">{rr.name}</p>
                                             </div>
                                         </div>
                                     )
                                 })
                             )}
                         </div>
                     </div>

                     <div className={`mb-5 flex flex-col gap-3`}>
                         <h2>My devices</h2>
                         <DeviceList devices={devices} />
                     </div>

                     <div className={`flex flex-col gap-3`}>
                         <h2>My oauth2 connections</h2>
                         <Oauth2Conns conns={me.oauth2_connections} />
                     </div>
                 </div>
             </>
            ) : null}

        </div>
    )
}
