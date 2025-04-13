export async function uploadImage(e) {
    const file = e.target.files[0]
    if (file) {
        const fd = new FormData()
        fd.append('file', file)

    }
}