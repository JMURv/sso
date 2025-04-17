"use client";
import {useState} from "react";
import AlternateEmailIcon from '@mui/icons-material/AlternateEmail';
import {Key, KeyboardArrowRight, Person} from "@mui/icons-material"
import {toast} from "sonner"
import {useRouter} from "next/navigation"

export default function Page() {
    const router = useRouter()
    const [fd, setFD] = useState({
        name: "",
        email: "",
        password: "",
    })

    const handleFDChange = (event) => {
        const { name, value } = event.target
        setFD((prev) => ({ ...prev, [name]: value }))
    }

    const handleReg = async () => {
        const r = await fetch("/api/users", {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            cache: "no-store",
            body: JSON.stringify({
                name: fd.name,
                email: fd.email,
                password: fd.password,
            })

        })

        if (!r.ok) {
            const data = await r.json()
            return toast.error(data.error)
        }

        await router.push("/")
    }

    return (
        <div className={`flex justify-center items-center h-screen w-screen`}>
            <div className={`w-full max-w-sm flex flex-col gap-3`}>
                <h1 className={`text-4xl font-bold tracking-widest uppercase`}>
                    Sign Up
                </h1>

                <div className={`w-full flex gap-3`}>
                    <div className={`w-full flex flex-col gap-3`}>

                        <div className={`icon-input-wrapper`}>
                            <div className={`icon-container`}>
                                <Person fontSize={"medium"} />
                            </div>
                            <input
                                type="text"
                                name={"name"}
                                value={name}
                                placeholder={"johndoe"}
                                onChange={handleFDChange}
                                className={`icon-input`}
                            />
                        </div>

                        <div className={`icon-input-wrapper`}>
                            <div className={`icon-container`}>
                                <AlternateEmailIcon fontSize={"medium"} />
                            </div>
                            <input
                                type="text"
                                name={"email"}
                                value={email}
                                placeholder={"johndoe@gmail.com"}
                                onChange={handleFDChange}
                                className={`icon-input`}
                            />
                        </div>

                        <div className={`icon-input-wrapper`}>
                            <div className={`icon-container`}>
                                <Key fontSize={"medium"} />
                            </div>
                            <input
                                type="password"
                                name={"password"}
                                value={password}
                                placeholder={"*********"}
                                onChange={handleFDChange}
                                className={`icon-input`}
                            />
                        </div>
                    </div>

                    <button onClick={handleReg} className={`primary-b`}>
                        <KeyboardArrowRight />
                    </button>
                </div>

                <div>
                    <p className={`text-xs`}>
                        Have an account?
                    </p>
                    <a href="/auth" className={`text-xs hover:underline`}>Sign in</a>
                </div>
            </div>
        </div>
    )
}