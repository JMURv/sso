import ModalBase from "./ModalBase"
import {Fingerprint} from "@mui/icons-material"
import {toast} from "sonner"
import {base64UrlToArrayBuffer} from "../../lib/auth/wa"
import {useRouter} from "next/navigation"
import {useAuth} from "../../providers/AuthProvider"

export default function WAModal({isWA, setIsWA, callback}) {
    const router = useRouter()
    const {authFetch} = useAuth()

    const handleWebAuthnStartReg = async () => {
        const r = await authFetch("/api/auth/webauthn/register/start", {
            method: "POST",
        })

        if (!r.ok) {
            const data = await r.json()
            return toast.error(data.errors)
        }
        const options = await r.json()

        let credential
        try {
            credential = await navigator.credentials.create({
                publicKey: {
                    ...options.publicKey,
                    challenge: base64UrlToArrayBuffer(options.publicKey.challenge),
                    user: {
                        ...options.publicKey.user,
                        id: base64UrlToArrayBuffer(options.publicKey.user.id),
                    },
                },
            })
        } catch (err) {
            console.error(err)
            toast.error("Authentication failed")
            return
        }

        const fin = await authFetch("/api/auth/webauthn/register/finish", {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(credential),
        })

        if (!fin.ok) {
            const data = await fin.json()
            toast.error(data.errors)
            return
        }

        await callback()
        setIsWA(false)
        toast.success("success")
        await router.refresh()
    }

    return (
        <ModalBase title={"WebAuthn"} isOpen={isWA} setIsOpen={setIsWA} >
            <div className={`flex flex-col gap-3 bg-zinc-950 p-5`}>
                <div className={`flex gap-3 w-full justify-between items-center`}>
                    <p>Use WebAuthn?</p>
                    <Fingerprint style={{fontSize: 30}} />
                </div>
                <div className={`flex gap-3 w-full`}>
                    <button onClick={handleWebAuthnStartReg} className={`w-full primary-b flex justify-center items-center`}>
                        Yes
                    </button>
                    <button onClick={() => setIsWA(false)} className={`w-full primary-b flex justify-center items-center`}>
                        No
                    </button>
                </div>
            </div>
        </ModalBase>
    )
}