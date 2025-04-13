"use client"
import {useRouter, useSearchParams} from "next/navigation"
import {KeyboardArrowLeft, KeyboardArrowRight} from "@mui/icons-material"

export default function Pagination({currentPage, totalPages, hasNextPage}) {
    const r = useRouter()
    const searchParams = useSearchParams()

    const onClick = async (e, page) => {
        e.preventDefault()
        const newSearchParams = new URLSearchParams(searchParams)
        if (page >= 1) {
            newSearchParams.set("page", page)
            await r.push(`?${newSearchParams.toString()}`)
        }
    }

    const page = Math.max(1, Math.min(currentPage, totalPages))
    const startPage = Math.max(1, page - 1)
    const endPage = Math.min(totalPages, page + 1)

    const pageNumbers = []
    for (let i = startPage; i <= endPage; i++) {
        pageNumbers.push(i)
    }

    return (
        <div id="pagination" className="flex flex-row flex-wrap gap-3 justify-center mt-10 mb-20">
            {/* Previous Page Button */}
            <button
                onClick={(e) => onClick(e, currentPage - 1)}
                disabled={currentPage === 1}
                className={`pagi-arrow`}
            >
                <KeyboardArrowLeft />
            </button>

            {/* First Page Link */}
            {startPage > 1 && (
                <>
                    <button onClick={(e) => onClick(e, 1)} className={`pagi-arrow`}>
                        1
                    </button>
                    {startPage > 2 &&
                        <span className="text-colors ring-zinc-900 justify-center items-center flex">...</span>}
                </>
            )}

            {/* Page Numbers */}
            {pageNumbers.map((pageNum) => (
                <button
                    onClick={(e) => onClick(e, pageNum)}
                    key={pageNum}
                    className={`pagi-arrow ${page === pageNum && "bg-zinc-900"} duration-200`}
                >
                    {pageNum}
                </button>
            ))}

            {/* Last Page Link */}
            {endPage < totalPages && (
                <>
                    {endPage < totalPages - 1 && <span className="px-4 py-2 text-colors">...</span>}
                    <button
                        onClick={(e) => onClick(e, totalPages)}
                        className={`pagi-arrow ${currentPage === totalPages && "bg-zinc-900 text-colors-rev"} duration-200`}
                    >
                        {totalPages}
                    </button>
                </>
            )}

            {/* Next Page Button */}
            <button
                onClick={(e) => onClick(e, currentPage + 1)}
                disabled={!hasNextPage}
                className={`pagi-arrow`}
            >
                <KeyboardArrowRight />
            </button>
        </div>
    )
}
