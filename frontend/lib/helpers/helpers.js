const formatDate = (dateString) => {
    const date = new Date(dateString)
    const year = date.getFullYear()
    const monthName = date.toLocaleString("en-US", { month: "long" }) // "April"
    const day = ("0" + date.getDate()).slice(-2)
    const hour = ("0" + date.getHours()).slice(-2)
    const minute = ("0" + date.getMinutes()).slice(-2)
    const second = ("0" + date.getSeconds()).slice(-2)
    return `${year} ${monthName} ${day} ${hour}:${minute}:${second}`
}

export default formatDate