import React from "react";

export default function Page({ children, title, subtitle, icon, cols = 1 }: { children: React.ReactNode, title: string, subtitle?: React.ReactNode, icon: React.ReactNode, cols?: number }) {

  return (
    <div className="p-12 md:p-16 xl:p-32 flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <div className="flex flex-row gap-4 items-center">
          {icon}
          <h1 className="text-2xl font-bold">{title}</h1>
        </div>
        {subtitle && <p className="text-sm text-gray-500">{subtitle}</p>}

      </div>
      <div className={``}>
        {children}
      </div>
    </div>
  )
}
