import React from 'react'

export function Loading({
    loading,
}: {
    loading: boolean
}) {
    return (
        <div>
            {loading ? (
                <div className="flex items-center justify-center">
                    <div className="animate-spin h-4 w-4 border-b-2 rounded-lg border-gray-200"></div>
                </div>
            ) : (
                <div></div>
            )}
        </div>
    )
}