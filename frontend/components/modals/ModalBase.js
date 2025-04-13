import {Transition, TransitionChild} from "@headlessui/react";

export default function ModalBase({isOpen, setIsOpen, children}) {
    const closeModal = () => {
        setIsOpen(false)
    }

    return (
        <Transition
            show={isOpen}
            className={`fixed top-0 bottom-0 overflow-y-auto z-50`}
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
                        className="fixed inset-0 bg-zinc-900/40 z-50"
                        onClick={closeModal}
                    />
                </TransitionChild>

                <TransitionChild
                    enter="transition-all ease-in-out duration-200"
                    enterFrom="transform scale-90 opacity-0"
                    enterTo="transform scale-100 opacity-100"
                    leave="transition-all ease-in-out duration-200"
                    leaveFrom="transform scale-100 opacity-100"
                    leaveTo="transform scale-90 opacity-0"
                >
                    <div className={`overflow-auto min-w-[350px] max-h-[80vh] max-w-3xl card p-6 card-outline z-50`}>
                        {children}
                    </div>
                </TransitionChild>
            </div>
        </Transition>
    )
}