'use client';

import Link from "next/link";
import { Settings, FolderTree, Database, Key, Server, CheckCircle2, ArrowRight } from "lucide-react";

export default function SettingsPage() {
  return (
    <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      <div>
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-xl bg-slate-900 text-white">
            <Settings size={22} />
          </div>
          <div>
            <h1 className="text-xl font-black text-slate-900 tracking-tight">Configuration de la Plateforme</h1>
            <p className="text-xs text-slate-500">Paramètres système, environnement et connecteurs d&apos;automatisation.</p>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-blue-50 text-blue-600">
                <FolderTree size={20} />
              </div>
              <div>
                <h3 className="font-bold text-sm text-slate-900">Arborescence Google Drive</h3>
                <p className="text-xs text-slate-500">AGENTS_DAO_WORKSPACE</p>
              </div>
            </div>
            <span className="inline-flex items-center gap-1 text-[11px] font-bold text-emerald-700 bg-emerald-50 px-2.5 py-1 rounded-full border border-emerald-200">
              <CheckCircle2 size={12} /> Actif
            </span>
          </div>
          <p className="text-xs text-slate-600 leading-relaxed">
            Dossier racine hébergeant le coffre-fort d&apos;entreprise RAG (<code>01_COFFRE_FORT_ENTREPRISE</code>) et l&apos;ensemble des dossiers d&apos;appels d&apos;offres (<code>02_PROJETS_DAO</code>).
          </p>
          <div className="pt-2">
            <Link
              href="/company"
              className="inline-flex items-center gap-1.5 text-xs font-bold text-blue-600 hover:text-blue-700 hover:underline"
            >
              Gérer dans le Coffre-fort <ArrowRight size={14} />
            </Link>
          </div>
        </div>

        <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-indigo-50 text-indigo-600">
                <Server size={20} />
              </div>
              <div>
                <h3 className="font-bold text-sm text-slate-900">Moteur d&apos;IA Multi-Agents</h3>
                <p className="text-xs text-slate-500">Gemini 2.5 Flash & Pro</p>
              </div>
            </div>
            <span className="inline-flex items-center gap-1 text-[11px] font-bold text-emerald-700 bg-emerald-50 px-2.5 py-1 rounded-full border border-emerald-200">
              <CheckCircle2 size={12} /> Connecté
            </span>
          </div>
          <p className="text-xs text-slate-600 leading-relaxed">
            Agents spécialisés : Juriste ARCOP, Rédacteur Technique BTP, Expert Chiffrage & Responsable Conformité Éligibilité.
          </p>
          <div className="pt-2">
            <Link
              href="/"
              className="inline-flex items-center gap-1.5 text-xs font-bold text-indigo-600 hover:text-indigo-700 hover:underline"
            >
              Consulter mes projets <ArrowRight size={14} />
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
