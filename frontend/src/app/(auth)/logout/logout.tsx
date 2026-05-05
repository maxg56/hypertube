"use client"
import React from "react";
import { useTranslation } from "react-i18next";

const LogoutButton: React.FC = () => {
	const { t } = useTranslation();
	
	return (
		<button
			// onClick={logout}
			className="bg-red-500 text-white text-xl w-40 px-4 py-2 rounded-2xl hover:bg-blue-600"
		>
			{t("Log Out")}
		</button>
	);
};

export default LogoutButton;