'use client';

import { useState } from "react";
import { DataAnalyseDAO, Lot } from "./types";
import LotDetailsPanel from "./LotDetailsPanel";
import {
  Building2,
  FileText,
  Layers,
  AlertTriangle,
  Lightbulb,
  ChevronRight,
} from "lucide-react";

interface DashboardDAOProps {
  data: DataAnalyseDAO;
}

export default function DashboardDAO({ data }: DashboardDAOProps) {
  const [selectedLot, setSelectedLot] = useState<Lot | null>(
    data.lots.length > 0 ? data.lots[0] : null
  );

  return (
    <div className="min-h-screen bg-slate-50 font-sans">
      {/* En-tête */}
      <header className="bg-slate-900 text-white px-6 py-5 shadow-md">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row md:items-center justify-between gap-3">
          <div>
            <p className="text-blue-400 text-xs font-bold uppercase tracking-widest mb-1">
              Plateforme Multi-Agents DAO
            </p>
            <h1 className="text-xl font-extrabold leading-tight">
              {data.objet_marche}
            </h1>
          </div>
          <div className="flex items-center gap-2 bg-slate-800 border border-slate-700 rounded-xl px-4 py-2.5 shrink-0">
            <Building2 size={16} className="text-slate-400" />
            <span className="text-sm text-slate-300 font-medium">
              {data.acheteur}
            </span>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-8 space-y-8">
        {/* Informations générales */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <InfoCard
            icon={<FileText size={20} className="text-blue-500" />}
            label="N° Appel d'Offres"
            value={data.numero_appel_offres ?? "Non renseigné"}
          />
          <InfoCard
            icon={<Layers size={20} className="text-indigo-500" />}
            label="Nombre de Lots"
            value={`${data.lots.length} lot${data.lots.length > 1 ? "s" : ""}`}
          />
          <InfoCard
            icon={<Building2 size={20} className="text-emerald-500" />}
            label="Acheteur"
            value={data.acheteur}
          />
        </div>

        {/* Critères éliminatoires */}
        {data.criteres_eliminatoires && data.criteres_eliminatoires.length > 0 && (
          <section className="bg-red-50 border border-red-200 rounded-2xl p-6">
            <h2 className="text-sm font-bold text-red-700 uppercase tracking-wider flex items-center gap-2 mb-4">
              <AlertTriangle size={18} />
              Critères Éliminatoires
            </h2>
            <ul className="space-y-2">
              {data.criteres_eliminatoires.map((c, i) => (
                <li
                  key={i}
                  className="flex items-start gap-2 text-sm text-red-800"
                >
                  <span className="mt-1.5 h-1.5 w-1.5 rounded-full bg-red-500 shrink-0" />
                  {c}
                </li>
              ))}
            </ul>
          </section>
        )}

        {/* Recommandations */}
        {data.recommandations && (
          <section className="bg-amber-50 border border-amber-200 rounded-2xl p-6">
            <h2 className="text-sm font-bold text-amber-700 uppercase tracking-wider flex items-center gap-2 mb-3">
              <Lightbulb size={18} />
              Recommandations de l&apos;Agent
            </h2>
            <p className="text-sm text-amber-900 leading-relaxed">
              {data.recommandations}
            </p>
          </section>
        )}

        {/* Sélecteur de lots + Détails */}
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* Liste des lots */}
          <div className="lg:col-span-1 space-y-2">
            <h2 className="text-xs font-bold text-slate-500 uppercase tracking-wider px-1 mb-3">
              Lots du marché
            </h2>
            {data.lots.map((lot) => (
              <button
                key={lot.numero_lot}
                onClick={() => setSelectedLot(lot)}
                className={`w-full text-left px-4 py-3.5 rounded-xl border transition-all flex items-center justify-between gap-2 ${
                  selectedLot?.numero_lot === lot.numero_lot
                    ? "bg-slate-900 text-white border-slate-900 shadow-md"
                    : "bg-white text-slate-700 border-slate-200 hover:border-slate-400 hover:bg-slate-50"
                }`}
              >
                <div>
                  <span
                    className={`text-[10px] font-bold uppercase tracking-widest ${
                      selectedLot?.numero_lot === lot.numero_lot
                        ? "text-blue-400"
                        : "text-slate-400"
                    }`}
                  >
                    Lot {lot.numero_lot}
                  </span>
                  <p className="text-sm font-semibold mt-0.5 leading-snug">
                    {lot.intitule}
                  </p>
                </div>
                <ChevronRight
                  size={16}
                  className={
                    selectedLot?.numero_lot === lot.numero_lot
                      ? "text-blue-400"
                      : "text-slate-300"
                  }
                />
              </button>
            ))}
          </div>

          {/* Panneau de détails */}
          <div className="lg:col-span-3">
            {selectedLot ? (
              <LotDetailsPanel lot={selectedLot} />
            ) : (
              <div className="flex items-center justify-center h-48 rounded-2xl border border-dashed border-slate-300 text-slate-400 text-sm">
                Sélectionnez un lot pour voir les détails
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}

// Composant utilitaire carte d'info
function InfoCard({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
}) {
  return (
    <div className="bg-white border border-slate-200 rounded-2xl p-5 flex items-start gap-4 shadow-sm">
      <div className="p-2.5 bg-slate-50 rounded-xl border border-slate-100">
        {icon}
      </div>
      <div className="min-w-0">
        <p className="text-[11px] text-slate-500 font-bold uppercase tracking-wider">
          {label}
        </p>
        <p className="text-sm font-semibold text-slate-800 mt-0.5 truncate">
          {value}
        </p>
      </div>
    </div>
  );
}
