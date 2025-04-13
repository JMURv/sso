"use client"

import Image from "next/image"
import {Check, Delete} from "@mui/icons-material"
import {toast} from "sonner"
import {useState} from "react"
import formatDate from "../../lib/helpers/helpers"

export default function DeviceList({t, devices}) {
    const [myDevices, setMyDevices] = useState(devices)

    const onDeviceChange = (e) => {
        const {name, value} = e.target
        const [field, deviceId] = name.split(".")
        setMyDevices((prevDevices) =>
            prevDevices.map((d) =>
                d.id.toString() === deviceId ? {...d, [field]: value} : d,
            ),
        )
    }

    const updateDevice = async (dID) => {
        const d = myDevices.find(d => d.id === dID)
        try {
            const r = await fetch(`/api/device/${dID}`, {
                method: "PUT",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${t}`,
                },
                body: JSON.stringify(d),
            })

            if (!r.ok) {
                const data = await r.json()
                toast.error(data.errors)
                return null
            }

        } catch (e) {
            console.error(e)
            return null
        }

        toast.success("Update successful")
    }

    const deleteDevice = async (e, dID) => {
        e.preventDefault()
        try {
            const r = await fetch(`/api/device/${dID}`, {
                method: "DELETE",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${t}`,
                },
            })

            if (!r.ok) {
                const data = await r.json()
                toast.error(data.errors)
                return null
            }

            setMyDevices(myDevices.filter(d => d.id !== dID))
            toast.success("Delete successful")
        } catch (e) {
            console.error(e)
        }
    }

    return (
        <div className="flex gap-3 w-full">
            {myDevices.length === 0 ? (
                <p className={`text-xs`}>No devices</p>
            ) : (
                <div className="w-full grid grid-cols-2 gap-2">
                    {myDevices.map((d, i) => {
                        const isLastAndOdd = myDevices.length % 2 !== 0 && i === myDevices.length - 1
                        return (
                            <div
                                key={d.id}
                                className={`flex w-full gap-2 ${isLastAndOdd ? "col-span-2" : ""}`}
                            >
                                <div className="bg-zinc-900/70 backdrop-blur-sm ring-inset ring-zinc-800 flex flex-col gap-3 p-4 w-full">
                                    <div className="mb-3 flex justify-between">
                                        <input
                                            type={"text"}
                                            name={`name.${d.id}`}
                                            value={d.name}
                                            onChange={onDeviceChange}
                                            className={`outline-none text-sm`}
                                        />
                                    </div>
                                    <div className="flex gap-4 items-center flex-wrap">
                                        {d.device_type === "desktop" && (
                                            <Image
                                                src={`/devices/type/desktop.svg`}
                                                width={20}
                                                height={20}
                                                alt="desktop device icon"
                                            />
                                        )}
                                        {d.os.includes("Windows") && (
                                            <Image
                                                src={`/devices/os/windows.svg`}
                                                width={20}
                                                height={20}
                                                alt="windows icon"
                                            />
                                        )}
                                        {d.browser === "Edge" && (
                                            <Image
                                                src={`/devices/browser/edge.svg`}
                                                width={20}
                                                height={20}
                                                alt="edge icon"
                                            />
                                        )}
                                        <div className="w-px h-full bg-zinc-700" />
                                        <p className="text-xs">{d.ip}</p>
                                    </div>
                                    <div className="h-px w-full bg-zinc-700" />
                                    <div className="flex flex-col">
                                        <p className="text-xs">Last active: {formatDate(d.last_active)}</p>
                                    </div>
                                </div>

                                <div className="flex flex-col gap-2 h-full">
                                    <button
                                        type="submit"
                                        onClick={() => updateDevice(d.id)}
                                        className="flex items-center cursor-pointer h-full p-2 bg-zinc-800 hover:bg-green-400 hover:text-zinc-900 duration-100"
                                    >
                                        <Check fontSize="small" />
                                    </button>
                                    <button
                                        type="submit"
                                        onClick={(e) => deleteDevice(e, d.id)}
                                        className="flex items-center cursor-pointer h-full p-2 bg-zinc-800 hover:bg-red-400 hover:text-zinc-900 duration-100"
                                    >
                                        <Delete fontSize="small" />
                                    </button>
                                </div>
                            </div>
                        )
                    })}
                </div>

            )}
        </div>
    )
}