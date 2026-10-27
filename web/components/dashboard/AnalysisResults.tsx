'use client';

import { ConsolidatedAnalysis } from "./types";
import {
  Building,
  ShieldAlert,
  FileCheck,
  CalendarDays,
  Layers,
  HardHat,
  ListOrdered,
  Wrench,
  Users,
  Lightbulb,
  AlertTriangle,
} from "lucide-react";
import { ReactNode } from "react";

function Section({ icon, title, children, accent = "blue" }: { icon: ReactNode; title: string; children: ReactNode; accent?: string }) {
  const accents: Record<string, string> = {
    blue: "bg-blue-50 text-blue-600 border-blue-100",
    red: "bg-rose-50 text-rose-600 border-rose-100",
    green: "bg-emerald-50 text-emerald-600 border-emerald-100",
    amber: "bg-amber-50 text-amber-600 border-amber-100",
    indigo: "bg-indigo-50 text-indigo-600 border-indigo-100",
  };
  return (
    <section className="bg-white border border-slate-200 rounded-2xl p-5 shadow-sm hover:shadow-md transition-shadow">
      <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2.5 mb-3">
        <span className={`w-8 h-8 rounded-lg border flex items-center justify-center ${accents[accent]}`}>{icon}</span>
        {title}
      </h3>
      {children}
    </section>
  );
}

function Bullets({ items, empty = "Non précisé dans le DAO." }: { items?: string[]; empty?: string }) {
  if (!items || items.length === 0) return <p className="text-xs text-slate-400 italic">{empty}</p>;
  return (
    <ul className="space-y-1.5">
      {items.map((it, i) => (
        <li key={i} className="text-sm text-slate-700 flex gap-2">
          <span className="text-slate-300 mt-0.5">•</span>
          <span>{it}</span>
        </li>
      ))}
    </ul>
  );
}

export default function AnalysisResults({ data }: { data: ConsolidatedAnalysis }) {
  const a = data.administrative;
  const t = data.technique;

  return (
    <div className="space-y-6">
      {data.erreurs && data.erreurs.length > 0 && (
        <div className="flex gap-3 p-4 bg-amber-50 border border-amber-200 rounded-2xl text-sm text-amber-800">
          <AlertTriangle size={18} className="shrink-0 mt-0.5" />
          <div>
            <p className="font-semibold">Analyse partielle</p>
            {data.erreurs.map((e, i) => (
              <p key={i} className="text-xs mt-1">{e}</p>
            ))}
          </div>
        </div>
      )}

      {a && (
        <div className="space-y-4">
          <h2 className="text-xs font-bold uppercase tracking-widest text-slate-500">Agent Conformité Administrative</h2>
          <div className="bg-gradient-to-br from-slate-900 to-indigo-950 text-white rounded-2xl p-6 shadow-lg">
            <div className="flex items-center gap-2 text-indigo-300 text-xs font-semibold uppercase tracking-wider">
              <Building size={14} /> {a.acheteur || "Acheteur non identifié"}
            </div>
            <p className="text-lg font-bold mt-2 leading-snug">{a.objet_marche}</p>
            <div className="flex flex-wrap gap-2 mt-4 text-xs">
              {a.numero_appel_offres && (
                <span className="px-2.5 py-1 rounded-lg bg-white/10 border border-white/10">N° {a.numero_appel_offres}</span>
              )}
              {a.date_limite_depot && (
                <span className="px-2.5 py-1 rounded-lg bg-rose-500/20 border border-rose-400/30 text-rose-100">
                  Dépôt : {a.date_limite_depot}
                </span>
              )}
            </div>
          </div>

          <div className="grid md:grid-cols-2 gap-4">
            <Section icon={<ShieldAlert size={16} />} title="Critères éliminatoires" accent="red">
              <Bullets items={a.criteres_eliminatoires} />
            </Section>
            <Section icon={<FileCheck size={16} />} title="Pièces administratives" accent="green">
              <Bullets items={a.pieces_administratives} />
            </Section>
            <Section icon={<CalendarDays size={16} />} title="Calendrier" accent="amber">
              {a.calendrier && a.calendrier.length > 0 ? (
                <ol className="space-y-2">
                  {a.calendrier.map((e, i) => (
                    <li key={i} className="flex justify-between gap-3 text-sm">
                      <span className="text-slate-700">{e.etape}</span>
                      <span className="text-xs font-mono text-slate-500 shrink-0">{e.date || "—"}</span>
                    </li>
                  ))}
                </ol>
              ) : (
                <Bullets items={[]} />
              )}
            </Section>
            <Section icon={<Layers size={16} />} title="Lots" accent="indigo">
              {a.lots && a.lots.length > 0 ? (
                <div className="space-y-3">
                  {a.lots.map((l) => (
                    <div key={l.numero_lot} className="p-3 rounded-xl bg-slate-50 border border-slate-100">
                      <p className="text-sm font-semibold text-slate-800">Lot {l.numero_lot} — {l.intitule}</p>
                      {l.cautionnement && <p className="text-xs text-slate-500 mt-1">Caution : {l.cautionnement}</p>}
                    </div>
                  ))}
                </div>
              ) : (
                <Bullets items={[]} empty="Lot unique / non précisé." />
              )}
            </Section>
          </div>
          {a.recommandations && (
            <Section icon={<Lightbulb size={16} />} title="Recommandations" accent="amber">
              <p className="text-sm text-slate-700 whitespace-pre-line">{a.recommandations}</p>
            </Section>
          )}
        </div>
      )}

      {t && (
        <div className="space-y-4">
          <h2 className="text-xs font-bold uppercase tracking-widest text-slate-500">Agent Ingénierie Technique</h2>
          <Section icon={<HardHat size={16} />} title="Consistance des travaux / prestations">
            <p className="text-sm text-slate-700 whitespace-pre-line">{t.consistance_travaux}</p>
          </Section>
          <div className="grid md:grid-cols-2 gap-4">
            <Section icon={<ListOrdered size={16} />} title="Plan du mémoire technique proposé" accent="indigo">
              {t.plan_memoire_technique && t.plan_memoire_technique.length > 0 ? (
                <ol className="space-y-1.5 list-decimal list-inside text-sm text-slate-700">
                  {t.plan_memoire_technique.map((p, i) => <li key={i}>{p}</li>)}
                </ol>
              ) : (
                <Bullets items={[]} />
              )}
            </Section>
            <Section icon={<AlertTriangle size={16} />} title="Contraintes d'exécution" accent="red">
              <Bullets items={t.contraintes_execution} />
            </Section>
            <Section icon={<Users size={16} />} title="Personnel clé" accent="green">
              <Bullets items={t.personnel_cle} />
            </Section>
            <Section icon={<Wrench size={16} />} title="Matériel & normes" accent="amber">
              <Bullets items={[...(t.materiel_requis ?? []), ...(t.normes_exigees ?? [])]} />
            </Section>
          </div>
          {t.points_vigilance && (
            <Section icon={<Lightbulb size={16} />} title="Points de vigilance" accent="amber">
              <p className="text-sm text-slate-700 whitespace-pre-line">{t.points_vigilance}</p>
            </Section>
          )}
        </div>
      )}
    </div>
  );
}
