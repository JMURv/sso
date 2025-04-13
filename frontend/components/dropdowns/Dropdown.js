import {Transition} from "@headlessui/react"
import {useEffect, useRef} from "react"

export default function Dropdown({isOpen, setIsOpen, classes, children}) {
    const dropdownRef = useRef(null)

    useEffect(() => {
        const handleCloseContext = (event) => {
            if (!dropdownRef.current.contains(event.target)) {
                setIsOpen(false)
            }
        }

        document.addEventListener("mousedown", handleCloseContext)
        return () => {
            document.removeEventListener("mousedown", handleCloseContext)
        }
    }, [setIsOpen])
    return (
        <div ref={dropdownRef} className={`absolute z-10 flex justify-start items-center ${classes}`}>
            <Transition
                as={"div"}
                show={isOpen}
                enter="transition ease-out duration-100"
                enterFrom="transform opacity-0 scale-y-0"
                enterTo="transform opacity-100 scale-y-100"
                leave="transition ease-in duration-75"
                leaveFrom="transform opacity-100 scale-y-100"
                leaveTo="transform opacity-0 scale-y-0"
                className="w-full origin-top"
            >
                <div className="origin-top-right overflow-hidden bg-zinc-800 focus:outline-none">
                    {children}
                </div>
            </Transition>
        </div>
    )
}