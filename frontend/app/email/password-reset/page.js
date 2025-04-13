"use client"
import {useState} from "react";
import {toast} from "sonner";
import {KeyboardArrowRight, LockOutlined} from "@mui/icons-material"
import {simpleChange} from "../../../lib/fd/fd"
import {useSearchParams} from "next/navigation"

const SuccessMessage = "Пароль изменен"

export default function Page(){
    const searchParams = useSearchParams()
    const uidb64 = searchParams.get("uidb64")
    const token = searchParams.get("token")
    const [formData, setFormData] = useState({
        password: "",
        uidb64: uidb64,
        token: token,
    })

    const onSubmit = async (e) => {
        e.preventDefault()

        formData.token = parseInt(formData.token)
        const r = await fetch("/api/auth/recovery/check", {
            method: "POST",
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(formData),
        })

        if (!r.ok) {
            const data = await r.json()
            return toast.error(data.errors)
        }

        toast.success(SuccessMessage)
        window.location.href = "/"
    }

    return (
        <div>
            <div className={`w-full h-screen flex flex-col justify-center items-center pt-20`}>
                <form onSubmit={onSubmit}>
                    <h1 className={`text-4xl font-bold tracking-widest uppercase mb-3`}>
                        Password reset
                    </h1>
                    <div className={`w-full flex gap-3`}>
                        <div className={`icon-input-wrapper`}>
                            <div className={`icon-container flex justify-center items-center`}>
                                <LockOutlined fontSize={"medium"} />
                            </div>
                            <input
                                type="password"
                                name={`password`}
                                placeholder={"*********"}
                                onChange={(e) => simpleChange(e, setFormData)}
                                className={`icon-input`}
                            />
                        </div>
                        <button type={"submit"} className={`primary-b`}>
                            <KeyboardArrowRight />
                        </button>
                    </div>
                </form>
            </div>
        </div>
    )
}