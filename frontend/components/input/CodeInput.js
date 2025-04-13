import {useRef} from "react"

export default function CodeInput({digits, setDigits}) {
    const inputRefs = [useRef(null), useRef(null), useRef(null), useRef(null)]

    const handleCodeInputChange = (index, value) => {
        const newDigits = [...digits]
        newDigits[index] = value
        if (value.length === 1 && index < digits.length - 1) {
            inputRefs[index + 1].current.focus()
        }
        setDigits(newDigits)
    }

    const handleCodeKeyDown = (index, event) => {
        if (event.key === "Backspace" && index > 0) {
            if (digits[index] === "") {
                inputRefs[index - 1].current.focus()
            } else {
                const newDigits = [...digits]
                newDigits[index] = ""
                setDigits(newDigits)
            }
        } else if (event.key === "ArrowLeft" && index > 0) {
            inputRefs[index - 1].current.focus()
        } else if (event.key === "ArrowRight" && index < digits.length - 1) {
            inputRefs[index + 1].current.focus()
        }
    }

    const handlePaste = (event) => {
        event.preventDefault()
        const clipboardData = event.clipboardData || window.clipboardData
        const pastedText = clipboardData.getData("text")
        if (/^\d{4}$/.test(pastedText)) {
            const newDigits = pastedText.split("")
            newDigits.forEach((digit, i) => {
                if (i < digits.length) {
                    inputRefs[i].current.value = digit
                }
            })
            setDigits(newDigits)
        }
    }

    return (
        digits.map((digit, index) => (
            <input
                key={index}
                ref={inputRefs[index]}
                className="aspect-square w-1/4 appearance-none bg-zinc-800/80 ring-1 ring-zinc-700 py-2 px-3 text-center text-colors-rev text-6xl font-medium placeholder:text-zinc-400 placeholder:font-medium leading-tight focus:outline-none focus:shadow-outline"
                type="text"
                value={digit}
                maxLength={1}
                onChange={(e) => handleCodeInputChange(index, e.target.value)}
                onKeyDown={(e) => handleCodeKeyDown(index, e)}
                onPaste={(e) => handlePaste(e)}
            />
        ))
    )
}