import React from "react";
import { Avatar, AvatarImage, AvatarFallback } from "../ui/avatar";
import {Button} from "../ui/button";

export default function Header() {
    return (
        <header className="bg-blue-400 text-white p-4">
            <div className="flex items-center justify-between">
                <Avatar className="size-20 ml-8 bg-gray-300">
                    <AvatarImage
                        src={`https://robohash.org/1.png?set=set1`}
                        alt="Avatar"
                    />
                    <AvatarFallback>HT</AvatarFallback>
                </Avatar>
                <h1 className="text-2xl text-gray-800 font-bold text-center flex-1">Hypertube</h1>
                <Button className="mr-8 bg-red-500 hover:bg-red-600 text-white">
                    Se déconnecter
                </Button>
            </div>            
        </header>
    );
}