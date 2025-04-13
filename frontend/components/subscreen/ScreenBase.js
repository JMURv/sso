import {useRef, useState} from "react"
import {Transition, TransitionChild} from "@headlessui/react"

const defaultHeight = 85

export default function SubScreenBase({isOpen, close, handleScroll, height, children}) {
    const [modalHeight, setModalHeight] = useState(height || defaultHeight)
    const modalRef = useRef(null)

    const handleDrag = (startEvent) => {
        const startY = startEvent.touches ? startEvent.touches[0].clientY : startEvent.clientY

        const onMove = (moveEvent) => {
            const moveY = moveEvent.touches ? moveEvent.touches[0].clientY : moveEvent.clientY
            const deltaY = startY - moveY
            const newHeight = Math.max(0, Math.min(100, modalHeight + (deltaY / window.innerHeight) * 100))

            setModalHeight(newHeight)
        }

        const onEnd = (moveEvent) => {
            const moveY = moveEvent.touches ? moveEvent.touches[0].clientY : moveEvent.clientY
            const deltaY = startY - moveY
            const newHeight = Math.max(0, Math.min(100, modalHeight + (deltaY / window.innerHeight) * 100))

            if (newHeight < 20) {
                close()
                setTimeout(() => {
                    setModalHeight(defaultHeight)
                }, 300)
            }
            if (newHeight > 95) {
                setModalHeight(100)
            }

            document.removeEventListener("mousemove", onMove)
            document.removeEventListener("mouseup", onEnd)
            document.removeEventListener("touchmove", onMove)
            document.removeEventListener("touchend", onEnd)
        }

        document.addEventListener("mousemove", onMove)
        document.addEventListener("mouseup", onEnd)
        document.addEventListener("touchmove", onMove)
        document.addEventListener("touchend", onEnd)
    }

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
                        className="fixed inset-0 bg-zinc-900/40"
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
                         style={{height: `${modalHeight}vh`}}
                         className={`fixed inset-x-0 bottom-0 overflow-auto bg-zinc-950 px-11 transition-height`}
                         onScroll={handleScroll}
                    >

                        <div
                            onMouseDown={handleDrag}
                            onTouchStart={handleDrag}
                            className={`sticky p-5 top-0 flex justify-center items-center gap-5 cursor-grab bg-zinc-900/40`}>
                            <span />

                            <div
                                id={`search-thumb`}
                                className={`h-2 w-full max-w-xl w-full bg-zinc-700 justify-center items-center`}
                            />

                        </div>
                        {children}
                    </div>
                </TransitionChild>
            </div>
        </Transition>

    )
}