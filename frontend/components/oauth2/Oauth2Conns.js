import {Google} from "@mui/icons-material"

export default function Oauth2Conns({conns}){
    return (
        <div className="flex gap-3">
            {conns.length === 0 ? (
                <p className={`text-xs`}>No oauth2 connections</p>
            ) : (
                conns.map((conn) => (
                    <div key={conn.provider + conn.provider_id}>
                        {conn.provider === "google" && (
                            <div className={`bg-zinc-900 p-2`}>
                                <Google />
                            </div>
                        )}
                    </div>
                ))
            )}
        </div>
    )
}