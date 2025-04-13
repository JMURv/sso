import {Transition, TransitionChild} from '@headlessui/react'

export default function SidebarBase({show, toggle, children, classes, direction}) {
    return(
        <Transition
            as={"div"}
            show={show}
            className={`fixed top-0 bottom-0 overflow-y-auto z-50`}
        >
            <TransitionChild
                enter="transition-all ease-in-out duration-300 ease-in-out"
                enterFrom="transform opacity-0"
                enterTo="transform opacity-100"
                leave="transition-all ease-in-out duration-300 ease-in-out"
                leaveFrom="transform opacity-100"
                leaveTo="transform opacity-0"
            >
                <div
                    className="fixed inset-0 bg-zinc-900/40 z-40"
                    onClick={toggle}
                />
            </TransitionChild>

            <TransitionChild
                enter="transition-all ease-in-out duration-300 ease-in-out"
                enterFrom={`transform opacity-0 ${direction ? "-translate-x-full":"translate-x-full"}`}
                enterTo={`transform opacity-100 ${direction ? "translate-x-0":"-translate-x-0"}`}
                leave="transition-all ease-in-out duration-300 ease-in-out"
                leaveFrom={`transform opacity-100 ${direction ? "translate-x-0":"-translate-x-0"}`}
                leaveTo={`transform opacity-0 ${direction ? "-translate-x-full":"translate-x-full"}`}
                className={`fixed h-full ${classes} z-50`}
            >
                {children}
            </TransitionChild>
        </Transition>
    )
}