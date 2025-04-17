"use client"
import Link from "next/link"
import {usePathname} from "next/navigation"
import {useAuth} from "../../providers/AuthProvider"

export default function Header() {
    const pathname = usePathname()
    const {isAdmin} = useAuth()
    return (
        <header>
            <nav className={`${pathname === "/auth" ? "hidden": "hidden md:block"}`}>
                <div className="fixed flex w-full p-5 items-center justify-between">
                    <div className={`flex gap-5 items-center`}>
                        <Link href="/" className='text-xl sm:text-3xl font-semibold text-zinc-800 dark:text-zinc-100 uppercase'>
                            {`// SSO`}
                        </Link>

                        {isAdmin && (
                            <>
                                <div className={`w-px h-9 bg-zinc-700`}/>

                                <Link href="/admin/users" className={`text-sm font-medium text-zinc-100 uppercase ${pathname === "/admin/users" ? "border-b-2 border-zinc-400" : ""}`}>
                                    {`users`}
                                </Link>

                                <Link href="/admin/roles" className={`text-sm font-medium text-zinc-100 uppercase ${pathname === "/admin/roles" ? "border-b-2 border-zinc-400" : ""}`}>
                                    {`roles`}
                                </Link>

                                <Link href="/admin/permissions" className={`text-sm font-medium text-zinc-100 uppercase ${pathname === "/admin/permissions" ? "border-b-2 border-zinc-400" : ""}`}>
                                    {`permissions`}
                                </Link>
                            </>
                        )}

                    </div>
                </div>
            </nav>
        </header>
    )
}