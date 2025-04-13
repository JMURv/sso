import {useRef} from "react"
import {Transition, TransitionChild} from "@headlessui/react"


export default function CroppedScreen({isOpen, close, children}) {
    const modalRef = useRef(null)
    return (
        <Transition
            show={isOpen}
            className={`fixed inset-0 overflow-y-auto z-50`}
        >
            <div className={`flex items-center justify-center w-full h-full inset-0`}>
                <TransitionChild
                    enter="transition-all ease-in-out duration-200"
                    enterFrom="transform opacity-0"
                    enterTo="transform opacity-100"
                    leave="transition-all ease-in-out duration-200"
                    leaveFrom="transform opacity-100"
                    leaveTo="transform opacity-0"
                >
                    <div
                        className="fixed inset-0 bg-zinc-950/90"
                        onClick={close}
                    />
                </TransitionChild>

                <TransitionChild
                    enter="transition-all  duration-300"
                    enterFrom="transform translate-y-full opacity-0"
                    enterTo="transform translate-y-0  opacity-100"
                    leave="transition-all duration-300"
                    leaveFrom="transform translate-y-0  opacity-100"
                    leaveTo="transform translate-y-full opacity-0"
                >
                    <div ref={modalRef}
                         className={`fixed inset-x-0 bottom-0 overflow-auto bg-zinc-950/0 px-11 transition-height`}
                    >
                        {children}
                    </div>
                </TransitionChild>
            </div>
        </Transition>

    )
}