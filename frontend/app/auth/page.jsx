"use client";
import {useEffect, useState} from "react"
import { toast } from "sonner";
import AlternateEmailIcon from '@mui/icons-material/AlternateEmail';
import {
    Fingerprint,
    GitHub,
    Google,
    Key,
    KeyboardArrowLeft,
    KeyboardArrowRight,
    Lock,
} from "@mui/icons-material"
import {useRouter, useSearchParams} from "next/navigation"
import Image from "next/image"
import CodeInput from "../../components/input/CodeInput"
import {base64UrlToArrayBuffer} from "../../lib/auth/wa"
import {useReCaptcha} from "next-recaptcha-v3"
import {useAuth} from "../../providers/AuthProvider"

export default function Page() {
    const router = useRouter()
    const params = useSearchParams()
    const {login} = useAuth()

    const [email, setEmail] = useState("")
    const [password, setPassword] = useState("")
    const [isCode, setIsCode] = useState(false)
    const [digits, setDigits] = useState(['', '', '', ''])
    const { executeRecaptcha } = useReCaptcha()

    const successAuth = async (access, refresh) => {
        await login(access, refresh)
        if (params.has("redirect") && params.get("redirect") !== "") {
            await router.push(params.get("redirect"))
            return
        }
        await router.push("/")
        await router.refresh()
    }

    const handleEmailChange = async (event) => {
        setEmail(event.target.value)
    }

    const handlePasswordChange = async (event) => {
        setPassword(event.target.value)
    }

    const handleEmailAuth = async () => {
        const token = await executeRecaptcha("email_auth")

        const r = await fetch("/api/auth/email/send", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "User-Agent": navigator.userAgent,
            },
            cache: "no-store",
            body: JSON.stringify({
                email: email,
                password: password,
                token: token,
            })
        })

        if (!r.ok) {
            const data = await r.json()
            return toast.error(data.errors)
        }
        const cLength = r.headers.get("content-length");
        if (cLength > 300) {
            const data = await r.json()
            return await successAuth(data.access, data.refresh)
        } else {
            setIsCode(true)
        }
    }

    const handlePasswordAuth = async () => {
        const token = await executeRecaptcha("pass_auth")
        const r = await fetch("/api/auth/jwt", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "User-Agent": navigator.userAgent,
            },
            cache: "no-store",
            body: JSON.stringify({
                email: email,
                password: password,
                token: token,
            })
        })

        if (!r.ok) {
            const data = await r.json()
            return toast.error(data.errors)
        }

        const data = await r.json()
        return await successAuth(data.access, data.refresh)
    }

    const handleProvider = async (provider) => {
        await router.push(`/api/auth/oauth2/${provider}/start`)
    }

    const handleForgotPassword  = async () => {
        const token = await executeRecaptcha("forgot_pass")
        const r = await fetch("/api/auth/recovery/send", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "User-Agent": navigator.userAgent,
            },
            body: JSON.stringify({
                email: email,
                token: token,
            })
        })

        if (!r.ok) {
            const data = await r.json()
            return toast.error(data.errors)
        }
        return toast.success("Email sent")
    }

    const handleWebAuthn = async () => {
        const token = await executeRecaptcha("wa_login")
        const r = await fetch("/api/auth/webauthn/login/start", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "User-Agent": navigator.userAgent,
            },
            body: JSON.stringify({
                email: email,
                token: token,
            })
        })

        if (!r.ok) {
            const data = await r.json()
            return toast.error(data.errors)
        }

        const options = await r.json()
        const publicKeyOptions = {
            ...options.publicKey,
            challenge: base64UrlToArrayBuffer(options.publicKey.challenge),
            allowCredentials: options.publicKey.allowCredentials?.map(cred => ({
                ...cred,
                id: base64UrlToArrayBuffer(cred.id),
            })),
        }

        let assertion;
        try {
            assertion = await navigator.credentials.get({
                publicKey: publicKeyOptions,
            })
        } catch (err) {
            console.error(err)
            toast.error("Authentication failed or was cancelled")
            return
        }

        const finR = await fetch("/api/auth/webauthn/login/finish", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "User-Agent": navigator.userAgent,
                "X-User-Email": email
            },
            body: JSON.stringify(assertion),
        })

        if (!finR.ok) {
            const data = await finR.json()
            toast.error(data.errors)
            return
        }

        toast.success("Login successful!")
        const data = await finR.json()
        return await successAuth(data.access, data.refresh)
    }

    useEffect(() => {
        let code = digits.join("")
        if (code.length === 4) {
            const CheckLoginCode = async () => {
                try {
                    const r = await fetch("/api/auth/email/check", {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            "User-Agent": navigator.userAgent,
                        },
                        body: JSON.stringify({
                            email: email,
                            code: parseInt(code)
                        }),
                    })

                    if (r.ok) {
                        setIsCode(false)
                        const data = await r.json()
                        return await successAuth(data.access, data.refresh)
                    } else {
                        toast.error("Invalid code")
                    }
                } catch (e) {
                    toast.error("Failed to get server")
                    console.log("Failed to get server", e)
                }
            }
            CheckLoginCode()
        }
    }, [digits])

    return (
        <div className={`flex justify-center items-center h-screen w-screen`}>
            <div className={`-z-10 fixed w-screen h-screen`}>
                <div className={`absolute bg-zinc-950/70 w-screen h-screen`}/>
                <video loop muted autoPlay src={`/bg/vids/main2.mp4`} className={`object-cover w-screen h-screen`} />
            </div>
            <div className={`w-full max-w-sm flex flex-col gap-3`}>
                {isCode ? (
                    <>
                        <h1 className={`text-4xl font-bold tracking-widest uppercase`}>
                            Enter code
                        </h1>
                        <div className={`w-full flex gap-3`}>
                            <CodeInput
                                digits={digits}
                                setDigits={setDigits}
                            />
                        </div>
                        <button onClick={() => setIsCode(false)} className={`primary-b`}>
                            <KeyboardArrowLeft />
                        </button>
                    </>
                ):(
                    <>
                        <h1 className={`text-4xl font-bold tracking-widest uppercase`}>
                            Sign In
                        </h1>

                        <button onClick={handleForgotPassword} className={`w-fit text-xs cursor-pointer hover:underline`}>
                            Forgot password?
                        </button>

                        <div className={`w-full flex gap-3`}>
                            <div className={`w-full flex flex-col gap-3`}>

                                <div className={`icon-input-wrapper`}>
                                    <div className={`icon-container`}>
                                        <AlternateEmailIcon fontSize={"medium"} />
                                    </div>
                                    <input
                                        type="text"
                                        name={"email"}
                                        value={email}
                                        placeholder={"johndoe@gmail.com"}
                                        onChange={handleEmailChange}
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
                                        onChange={handlePasswordChange}
                                        className={`icon-input`}
                                    />
                                </div>
                            </div>

                            <button onClick={handleEmailAuth} className={`primary-b`}>
                                <KeyboardArrowRight />
                            </button>
                        </div>


                        <div className={`grid grid-cols-2 gap-3 mb-3`}>
                            <button onClick={handlePasswordAuth} className={`primary-b`}>
                                <Lock />
                            </button>
                            <button onClick={() => handleProvider("google")} className={`primary-b`}>
                                <Google />
                            </button>
                            {/* START TODO: Implement additional providers */}

                            {/*<button onClick={() => handleProvider("github")} className={`primary-b`}>*/}
                            {/*    <GitHub />*/}
                            {/*</button>*/}
                            {/*<button onClick={() => handleProvider("vk")} className={`primary-b flex justify-center items-center`}>*/}
                            {/*    <Image src={`/vk.svg`} width={23} height={23} alt={``} />*/}
                            {/*</button>*/}
                            {/*<button onClick={() => handleProvider("gosuslugi")} className={`primary-b flex justify-center items-center`}>*/}
                            {/*    <Image src={`/gosuslugi.svg`} width={23} height={23} alt={``} />*/}
                            {/*</button>*/}

                            {/* END */}
                            <button onClick={handleWebAuthn} className={`primary-b w-full col-span-2`}>
                                <Fingerprint />
                            </button>
                        </div>

                        <div>
                            <p className={`text-xs`}>
                                Doesn't have an account?
                            </p>
                            <a href="/reg" className={`text-xs hover:underline`}>
                                Sign up
                            </a>
                        </div>
                    </>
                )}
            </div>
        </div>
    );
}