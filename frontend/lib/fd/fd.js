export function simpleChange(e, setFormData) {
    e.preventDefault()
    const { name, type, value, checked } = e.target
    const newValue = type === 'checkbox' ? checked : value

    setFormData((prev) => ({
        ...prev,
        [name]: newValue
    }))
}