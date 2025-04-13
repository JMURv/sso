"use client"
import {Add, Check, Delete, KeyboardArrowDown} from "@mui/icons-material"
import {useState} from "react"
import {toast} from "sonner"
import PermModal from "../../../components/modals/PermModal"

const roleBlueprint = {
    name: "",
    description: "",
}


export default function New({t, close, successCallback}) {
    const [addPermModal, setAddPermModal] = useState(false)
    const [role, setRole] = useState({
        name: "",
        description: "",
        permissions: [],
    })

    const handleChange = async (e) => {
        const {name, type, value, checked} = e.target
        const newValue = type === "checkbox" ? checked : value
        const keys = name.split(".")
        setRole((prevFormData) => {
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

    const createRole = async () => {
        try {
            const permIDs = role.permissions.map(p => p.id)
            const r = await fetch(`/api/roles`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${t}`,
                },
                body: JSON.stringify({
                    name: role.name,
                    description: role.description,
                    permissions: permIDs,
                }),
            })

            if (!r.ok) {
                const data = await r.json()
                toast.error(data.errors)
                return null
            }

            role.id = await r.json()
            successCallback(role)
            toast.success("Create successful")
        } catch (e) {
            console.error(e)
            toast.error("Something went wrong")
        }
    }

    const removePerm = async (id) => {
        setRole({...role, permissions: role.permissions.filter(p => p.id !== id)})
    }

    const onChoosePerm = (perm) => {
        if (role.permissions.some(p => p.id === perm.id)) {
            toast.error("Role already has this permission")
            return
        }

        setAddPermModal(false)
        setRole({...role, permissions: [...role.permissions, perm]})
    }

    return (
        <div className={`animate-fadeIn flex flex-row gap-5`}>
            <PermModal
                isOpen={addPermModal}
                setIsOpen={setAddPermModal}
                onClick={(perm) => onChoosePerm(perm)}
            />

            <div className={`w-full flex flex-col`}>

                <div className={`flex flex-row gap-5`}>
                    <div className={`p-3 gap-3 bg-zinc-900/70 ring-1 ring-zinc-800 text-zinc-100 w-full flex flex-col justify-between`}>
                        <div className={`flex flex-row flex-wrap gap-3`}>
                            <div className={`w-full flex flex-col gap-3`}>
                                <input
                                    type="text"
                                    name={"name"}
                                    value={role.name}
                                    placeholder={"Developer"}
                                    onChange={handleChange}
                                    className={`outline-none text-sm w-full border-b p-2 border-zinc-700`}
                                />

                                <textarea
                                    name={"description"}
                                    value={role.description}
                                    placeholder={"Lorem ipsum dolor sit amet consectetur adipisicing elit. Quia, quisquam."}
                                    onChange={handleChange}
                                    className={`outline-none text-sm w-full border-b p-2 border-zinc-700`}
                                    rows={5}
                                />
                            </div>

                            <div className={`mb-2 flex justify-between items-center`}>
                                <h1 className={`text-sm tracking-widest uppercase`}>Roles</h1>
                                <button onClick={() => setAddPermModal(true)} className={`cursor-pointer p-1`}>
                                    <Add />
                                </button>
                            </div>

                            <div className={`w-full grid grid-cols-2 gap-3`}>
                                {role?.permissions.length > 0 && (
                                    role.permissions.map((p) => (
                                        <div key={p.id}
                                             className={`flex w-full justify-between bg-zinc-950`}>
                                            <div className="px-4 py-1">
                                                <p className="text-xs tracking-wider text-zinc-200 capitalize">{p.name}</p>
                                                <p className="text-xs tracking-wider text-zinc-400 capitalize">{p.description}</p>
                                            </div>
                                            <button id={`remove-role`}
                                                    onClick={() => removePerm(p.id)}
                                            >
                                                <Delete fontSize={"small"} />
                                            </button>
                                        </div>
                                    ))
                                )}
                            </div>
                        </div>
                    </div>
                </div>

                <div className={`mt-5 flex flex-row gap-5`}>
                    <button onClick={() => close()} className={`primary-b w-full h-full`}>
                        <KeyboardArrowDown />
                    </button>
                    <button onClick={createRole} className={`primary-b w-full h-full`}>
                        <Check />
                    </button>
                </div>
            </div>
        </div>
    )
}