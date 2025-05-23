"use client"
import {
    Check,
    KeyboardArrowDown,
} from "@mui/icons-material"
import {useState} from "react"
import {toast} from "sonner"
import {useAuth} from "../../../providers/AuthProvider"

export default function Edit({srvOBJ, close, successCallback}) {
    const {authFetch} = useAuth()
    const [obj, setOBJ] = useState(srvOBJ)

    const handleChange = async (e) => {
        const {name, type, value, checked} = e.target
        const newValue = type === "checkbox" ? checked : value
        const keys = name.split(".")
        setOBJ((prev) => {
            let updatedFormData = {...prev}

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

    const updateOBJ = async () => {
        const response = await authFetch(`/api/perm/${srvOBJ.id}`, {
            method: "PUT",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                name: obj.name,
                description: obj.description,
            }),
        })

        if (!response.ok) {
            const data = await response.json()
            toast.error(data.errors)
            return
        }

        successCallback(obj)
        toast.success("Update successful")
    }

    return (
        <div className={`animate-fadeIn w-full flex flex-row gap-5`}>
            <div className={`w-full flex flex-col`}>
                <div className={`flex flex-row gap-5`}>
                    <div className={`p-3 gap-3 bg-zinc-900/70 ring-1 ring-zinc-800 text-zinc-100 w-full flex flex-col justify-between`}>
                        <div className={`flex flex-row flex-wrap gap-3`}>
                            <div className={`w-full flex flex-col gap-3`}>
                                <input
                                    type="text"
                                    name={"name"}
                                    value={obj.name}
                                    placeholder={"Developer"}
                                    onChange={handleChange}
                                    className={`outline-none text-sm w-full border-b p-2 border-zinc-700`}
                                />
                                <textarea
                                    name={"description"}
                                    value={obj.description}
                                    placeholder={"Lorem ipsum dolor sit amet consectetur adipisicing elit. Quia, quisquam."}
                                    onChange={handleChange}
                                    className={`outline-none text-sm w-full border-b p-2 border-zinc-700`}
                                    rows={5}
                                />
                            </div>
                        </div>
                    </div>
                </div>

                <div className={`mt-5 flex flex-row gap-5`}>
                    <button onClick={() => close()} className={`primary-b w-full h-full`}>
                        <KeyboardArrowDown />
                    </button>
                    <button onClick={updateOBJ} className={`primary-b w-full h-full`}>
                        <Check />
                    </button>
                </div>
            </div>
        </div>
    )
}