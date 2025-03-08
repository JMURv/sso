"use client";
import {useState} from "react";
import AlternateEmailIcon from '@mui/icons-material/AlternateEmail';
import {Key, Login} from "@mui/icons-material";

export default function Page() {
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const handleEmailChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setEmail(event.target.value);
    };

    const handlePasswordChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setPassword(event.target.value);
    };

    return (
        <div className={`flex justify-center items-center h-screen w-screen`}>
            <div className={`w-full max-w-sm flex flex-col gap-3`}>
                <h1 className={`text-4xl font-bold`}>
                    Sign In
                </h1>

                <div className={`w-full flex gap-3`}>
                    <div className={`w-full flex flex-col gap-3`}>
                        <div className={`flex w-full icon-input-wrapper`}>
                            <AlternateEmailIcon />
                            <input
                                type="text"
                                name={"email"}
                                value={email}
                                placeholder={""}
                                onChange={handleEmailChange}
                                className={`icon-input w-full`}
                            />
                        </div>

                        <div className={`flex w-full icon-input-wrapper`}>
                            <Key />
                            <input
                                type="text"
                                name={"password"}
                                value={password}
                                placeholder={""}
                                onChange={handlePasswordChange}
                                className={`icon-input w-full`}
                            />
                        </div>
                    </div>

                    <button className={`primary-b`}>
                        <Login />
                    </button>
                </div>
            </div>
        </div>
    );
}