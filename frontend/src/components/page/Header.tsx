import React from "react";
import { Avatar, AvatarImage, AvatarFallback } from "../ui/avatar";
import {Button} from "../ui/button";
import LogoutButton from "../../app/(auth)/logout/logout";
import LanguageSwitcher from "./LanguageSwitcher";
import Link from 'next/link'

export default function Header() {
    return (
        <header className="bg-blue-400 text-white p-4">
            <div className="flex items-center justify-between">
                    <Link href="/profile">
                        <Avatar className="size-20 ml-8 bg-gray-300 cursor-pointer hover:opacity-80 transition-opacity">
                            <AvatarImage
                                src={`https://robohash.org/1.png?set=set1`}
                                alt="Avatar"
                            />
                            <AvatarFallback>HT</AvatarFallback>
                        </Avatar>
                    </Link>
                <h1 className="text-2xl text-gray-800 font-bold text-center flex-1">Hypertube</h1>
                <div className="flex items-center gap-4">
                    <LanguageSwitcher />
                    <LogoutButton />
                </div>
            </div>            
        </header>
    );
}