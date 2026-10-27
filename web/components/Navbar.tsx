import Link from "next/link";
import { FolderKanban, Settings, Bot, Building2 } from "lucide-react";

export default function Navbar() {
  return (
    <nav className="bg-slate-900 border-b border-slate-800 text-white sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo / Marque */}
          <div className="flex items-center gap-3">
            <Link href="/" className="flex items-center gap-2.5 group">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-500 flex items-center justify-center shadow-lg shadow-blue-500/20 group-hover:scale-105 transition-transform">
                <Bot className="w-5 h-5 text-white" />
              </div>
              <div>
                <span className="font-extrabold text-base tracking-tight text-white block">
                  AGENTS<span className="text-blue-400">DAO</span>
                </span>
                <span className="text-[10px] text-slate-400 tracking-wider uppercase font-semibold block -mt-1">
                  Plateforme Marchés Publics
                </span>
              </div>
            </Link>
          </div>

          {/* Navigation Principale */}
          <div className="flex items-center gap-1 sm:gap-2">
            <Link
              href="/"
              className="flex items-center gap-2 px-3.5 py-2 rounded-lg text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800 transition-colors"
            >
              <FolderKanban className="w-4 h-4 text-blue-400" />
              <span>Mes Projets DAO</span>
            </Link>

            <Link
              href="/company"
              className="flex items-center gap-2 px-3.5 py-2 rounded-lg text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800 transition-colors"
            >
              <Building2 className="w-4 h-4 text-indigo-400" />
              <span>Coffre-fort RAG</span>
            </Link>

            <Link
              href="/settings"
              className="flex items-center gap-2 px-3.5 py-2 rounded-lg text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800 transition-colors"
            >
              <Settings className="w-4 h-4 text-slate-400" />
              <span className="hidden sm:inline">Configuration</span>
            </Link>
          </div>
        </div>
      </div>
    </nav>
  );
}
