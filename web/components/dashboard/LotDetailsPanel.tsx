import { Lot } from "./types";
import { CheckSquare, ShieldCheck } from "lucide-react";

interface LotDetailsPanelProps {
  lot: Lot;
}

export default function LotDetailsPanel({ lot }: LotDetailsPanelProps) {
  return (
    <div className="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden">
      <div className="bg-slate-900 text-white p-6 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <span className="text-blue-400 text-xs font-bold uppercase tracking-widest">
            Focus Dépouillement
          </span>
          <h3 className="text-lg font-bold mt-1">
            Lot n°{lot.numero_lot} — {lot.intitule}
          </h3>
        </div>
        <div className="bg-slate-800 border border-slate-700 rounded-xl px-4 py-2.5 shrink-0">
          <p className="text-[10px] text-slate-400 font-bold uppercase tracking-wider">
            Caution de Soumission
          </p>
          <p className="text-amber-400 font-mono font-bold text-lg">
            {lot.cautionnement}
          </p>
        </div>
      </div>

      <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Critères Techniques */}
        <div className="space-y-4">
          <h4 className="text-sm font-bold text-slate-800 uppercase tracking-wider flex items-center gap-2 border-b border-slate-100 pb-2">
            <ShieldCheck size={18} className="text-red-500" />
            Critères Techniques & Capacité
          </h4>
          <ul className="space-y-3">
            {lot.criteres_techniques.map((critere, idx) => (
              <li
                key={idx}
                className="bg-slate-50 border border-slate-200/60 rounded-xl p-3.5 text-sm text-slate-700 leading-relaxed"
              >
                {critere}
              </li>
            ))}
          </ul>
        </div>

        {/* Check-list des pièces */}
        <div className="space-y-4">
          <h4 className="text-sm font-bold text-slate-800 uppercase tracking-wider flex items-center gap-2 border-b border-slate-100 pb-2">
            <CheckSquare size={18} className="text-green-600" />
            Pièces Exigées (Check-list)
          </h4>
          <div className="space-y-2">
            {lot.pieces_requises.map((piece, idx) => (
              <label
                key={idx}
                className="flex items-start gap-3 p-3 bg-white hover:bg-slate-50 border border-slate-200 rounded-xl cursor-pointer transition-colors text-sm text-slate-700 group"
              >
                <input
                  type="checkbox"
                  className="mt-0.5 h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500/30 accent-blue-600 cursor-pointer"
                />
                <span className="group-hover:text-slate-900 transition-colors">
                  {piece}
                </span>
              </label>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
