'use client';

import { use, useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import AnalysisResults from "@/components/dashboard/AnalysisResults";
import TechnicalOfferEditor from "@/components/dashboard/TechnicalOfferEditor";
import { Project, ConsolidatedAnalysis, DriveDocument, ProjectTask, ComplianceCheckResult, TechnicalOfferData, CompanyDocument } from "@/components/dashboard/types";
import {
  ArrowLeft,
  ExternalLink,
  Bot,
  Sparkles,
  FolderTree,
  AlertTriangle,
  UploadCloud,
  FileText,
  RotateCw,
  CheckCircle2,
  Circle,
  Clock,
  ListTodo,
  LogIn,
  ShieldCheck,
  ShieldAlert,
  FolderCheck,
  ArrowRight,
  Copy,
  RefreshCw,
  Check,
  AlertCircle,
  Calendar,
} from "lucide-react";

const API = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

const STATUS_LABEL: Record<string, { label: string; cls: string }> = {
  BROUILLON: { label: "Brouillon", cls: "bg-slate-700 text-slate-200" },
  EN_ANALYSE: { label: "En analyse", cls: "bg-blue-500/20 text-blue-200 animate-pulse" },
  EN_MONTAGE: { label: "En montage", cls: "bg-emerald-500/20 text-emerald-200" },
  PRET_DEPOT: { label: "Prêt au dépôt", cls: "bg-indigo-500/20 text-indigo-200" },
};

const CAT_CLS: Record<string, string> = {
  ADMINISTRATIF: "bg-emerald-50 text-emerald-700 border-emerald-100",
  TECHNIQUE: "bg-indigo-50 text-indigo-700 border-indigo-100",
  FINANCIER: "bg-amber-50 text-amber-700 border-amber-100",
  VISITE_SITE: "bg-rose-50 text-rose-700 border-rose-100",
};

const NEXT_STATUS: Record<ProjectTask["statut"], ProjectTask["statut"]> = {
  A_FAIRE: "EN_COURS",
  EN_COURS: "TERMINEE",
  TERMINEE: "A_FAIRE",
};

export default function ProjectDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id: projectId } = use(params);

  const [project, setProject] = useState<Project | null>(null);
  const [loading, setLoading] = useState(true);
  const [docs, setDocs] = useState<DriveDocument[]>([]);
  const [needsAuth, setNeedsAuth] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [starting, setStarting] = useState(false);
  const fileInput = useRef<HTMLInputElement>(null);

  // Onglets & Contrôle de conformité RAG & Offre Technique
  const [activeTab, setActiveTab] = useState<"compliance" | "technical" | "analysis">("compliance");
  const [compliance, setCompliance] = useState<ComplianceCheckResult | null>(null);
  const [complianceLoading, setComplianceLoading] = useState(false);
  const [copyingDocs, setCopyingDocs] = useState(false);
  const [copyResultMsg, setCopyResultMsg] = useState<{ text: string; error?: boolean; folderUrl?: string } | null>(null);
  const [complianceFilter, setComplianceFilter] = useState<"ALL" | "CONFORME" | "A_RENOUVELER" | "MANQUANT" | "ELIMINATOIRE">("ALL");
  const [movingToWorkspace, setMovingToWorkspace] = useState(false);
  const [moveSuccess, setMoveSuccess] = useState(false);
  const [technicalOffer, setTechnicalOffer] = useState<TechnicalOfferData | null>(null);
  const [companyDocs, setCompanyDocs] = useState<CompanyDocument[]>([]);

  const analyzing = project?.statut === "EN_ANALYSE" || starting;

  const fetchProject = useCallback(async () => {
    try {
      const res = await fetch(`${API}/api/projects/${projectId}`, { credentials: "include" });
      if (res.ok) setProject(await res.json());
    } catch (err) {
      console.error("Erreur récupération projet:", err);
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  const fetchDocs = useCallback(async () => {
    try {
      const res = await fetch(`${API}/api/projects/${projectId}/documents`, { credentials: "include" });
      const data = await res.json();
      if (res.status === 401) {
        setNeedsAuth(true);
        return;
      }
      setNeedsAuth(false);
      if (res.ok) setDocs(data.documents ?? []);
      else setError(data.error ?? "Impossible de lister les documents.");
    } catch {
      setError("Impossible de joindre l'API.");
    }
  }, [projectId]);

  const fetchCompliance = useCallback(async () => {
    try {
      setComplianceLoading(true);
      const res = await fetch(`${API}/api/projects/${projectId}/compliance-check`, { credentials: "include" });
      if (res.ok) {
        setCompliance(await res.json());
      }
    } catch (err) {
      console.error("Erreur récupération conformité:", err);
    } finally {
      setComplianceLoading(false);
    }
  }, [projectId]);

  const fetchTechnicalOffer = useCallback(async () => {
    try {
      const res = await fetch(`${API}/api/projects/${projectId}/technical-offer`, { credentials: "include" });
      if (res.ok) {
        setTechnicalOffer(await res.json());
      }
    } catch (err) {
      console.error("Erreur récupération offre technique:", err);
    }
  }, [projectId]);

  const fetchCompanyDocs = useCallback(async () => {
    try {
      const res = await fetch(`${API}/api/company`, { credentials: "include" });
      if (res.ok) {
        const data = await res.json();
        setCompanyDocs(data.documents || []);
      }
    } catch {}
  }, []);

  useEffect(() => {
    fetchProject();
    fetchDocs();
    fetchCompliance();
    fetchTechnicalOffer();
    fetchCompanyDocs();
  }, [fetchProject, fetchDocs, fetchCompliance, fetchTechnicalOffer, fetchCompanyDocs]);

  // Polling pendant l'analyse
  useEffect(() => {
    if (project?.statut !== "EN_ANALYSE") return;
    const t = setInterval(() => {
      fetchProject();
      fetchCompliance();
      fetchTechnicalOffer();
    }, 3000);
    return () => clearInterval(t);
  }, [project?.statut, fetchProject, fetchCompliance, fetchTechnicalOffer]);

  const uploadFiles = async (files: FileList | File[]) => {
    const list = Array.from(files);
    if (list.length === 0) return;
    setError(null);
    setUploading(true);
    try {
      for (const file of list) {
        if (!file.name.toLowerCase().endsWith(".pdf")) {
          setError(`« ${file.name} » ignoré : seuls les PDF sont acceptés.`);
          continue;
        }
        const form = new FormData();
        form.append("file", file);
        const res = await fetch(`${API}/api/projects/${projectId}/documents`, {
          method: "POST",
          body: form,
          credentials: "include",
        });
        if (!res.ok) {
          const data = await res.json().catch(() => ({}));
          if (res.status === 401) setNeedsAuth(true);
          setError(data.error ?? `Échec du téléversement de ${file.name}`);
        }
      }
      await fetchDocs();
    } finally {
      setUploading(false);
      if (fileInput.current) fileInput.current.value = "";
    }
  };

  const launchAnalysis = async () => {
    setError(null);
    setStarting(true);
    try {
      const res = await fetch(`${API}/api/projects/${projectId}/analyze`, { method: "POST", credentials: "include" });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        if (res.status === 401) setNeedsAuth(true);
        setError(data.error ?? "Impossible de lancer l'analyse.");
      }
      await fetchProject();
    } finally {
      setStarting(false);
    }
  };

  const toggleTask = async (task: ProjectTask) => {
    const statut = NEXT_STATUS[task.statut];
    setProject((p) => p && { ...p, tasks: p.tasks?.map((t) => (t.id === task.id ? { ...t, statut } : t)) });
    const res = await fetch(`${API}/api/tasks/${task.id}`, {
      method: "PATCH",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ statut }),
    });
    if (!res.ok) {
      fetchProject();
    } else {
      fetchCompliance();
    }
  };

  const handleCopyRagDocs = async () => {
    setCopyingDocs(true);
    setCopyResultMsg(null);
    try {
      const res = await fetch(`${API}/api/projects/${projectId}/copy-rag-documents`, {
        method: "POST",
        credentials: "include",
      });
      const data = await res.json();
      if (res.ok) {
        setCopyResultMsg({
          text: data.message || `${data.copied_count} pièce(s) transférée(s) sur Drive avec succès !`,
          folderUrl: data.target_folder_url,
        });
        await fetchProject();
        await fetchCompliance();
      } else {
        setCopyResultMsg({ text: data.error || "Erreur lors de la copie des documents.", error: true });
      }
    } catch {
      setCopyResultMsg({ text: "Impossible de joindre l'API pour copier les documents.", error: true });
    } finally {
      setCopyingDocs(false);
    }
  };

  const handleMoveToWorkspace = async () => {
    setMovingToWorkspace(true);
    try {
      const res = await fetch(`${API}/api/projects/${projectId}/move-to-workspace`, {
        method: "POST",
        credentials: "include",
      });
      const data = await res.json();
      if (res.ok) {
        setMoveSuccess(true);
        setTimeout(() => setMoveSuccess(false), 6000);
      } else {
        setError(data.error || "Impossible de classer le projet.");
      }
    } catch {
      setError("Erreur de connexion lors du classement du projet.");
    } finally {
      setMovingToWorkspace(false);
    }
  };

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-16 text-center space-y-4">
        <div className="w-10 h-10 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto" />
        <p className="text-sm text-slate-500">Chargement du dossier d&apos;appel d&apos;offres...</p>
      </div>
    );
  }

  if (!project) {
    return (
      <div className="max-w-xl mx-auto px-4 py-16 text-center">
        <h2 className="text-xl font-bold text-slate-900">Projet introuvable</h2>
        <Link href="/" className="inline-flex items-center gap-2 text-sm text-blue-600 mt-4 hover:underline">
          <ArrowLeft size={16} />
          <span>Retour aux projets</span>
        </Link>
      </div>
    );
  }

  let analysis: ConsolidatedAnalysis | null = null;
  if (project.analyse_result_json) {
    try {
      const parsed = JSON.parse(project.analyse_result_json);
      if (parsed.administrative || parsed.technique) analysis = parsed;
    } catch {
      /* ancien format ignoré */
    }
  }

  const status = STATUS_LABEL[project.statut] ?? STATUS_LABEL.BROUILLON;
  const tasks = project.tasks ?? [];
  const done = tasks.filter((t) => t.statut === "TERMINEE").length;
  const canAnalyze = docs.length > 0 && !analyzing && !needsAuth;

  return (
    <div className="min-h-screen bg-slate-50 pb-16">
      {/* En-tête */}
      <div className="bg-slate-900 border-b border-slate-800 text-white px-4 sm:px-8 py-4">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="flex items-center gap-4 min-w-0">
            <Link href="/" className="p-2 rounded-xl bg-slate-800 text-slate-400 hover:text-white transition-colors">
              <ArrowLeft size={18} />
            </Link>
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <span className="text-xs text-blue-400 font-bold uppercase tracking-widest">
                  {project.numero_dao ? `DAO N° ${project.numero_dao}` : `Projet #${project.id}`}
                </span>
                <span className={`text-[10px] font-bold px-2 py-0.5 rounded-md ${status.cls}`}>{status.label}</span>
              </div>
              <h1 className="text-lg font-extrabold truncate max-w-2xl">{project.titre}</h1>
              {project.acheteur && <p className="text-xs text-slate-400 truncate">{project.acheteur}</p>}
            </div>
          </div>

          <div className="flex items-center gap-3 shrink-0">
            {project.drive_folder_url && (
              <>
                <button
                  type="button"
                  onClick={handleMoveToWorkspace}
                  disabled={movingToWorkspace}
                  title="Déplacer ce projet dans AGENTS_DAO_WORKSPACE / 02_PROJETS_DAO sur Google Drive"
                  className="inline-flex items-center gap-1.5 px-3 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white text-xs font-semibold rounded-xl border border-slate-700 transition-colors cursor-pointer"
                >
                  {movingToWorkspace ? (
                    <RefreshCw size={13} className="animate-spin text-indigo-400" />
                  ) : moveSuccess ? (
                    <Check size={13} className="text-emerald-400" />
                  ) : (
                    <FolderCheck size={13} className="text-indigo-400" />
                  )}
                  <span>{moveSuccess ? "Classé dans 02_PROJETS_DAO" : "Classer dans 02_PROJETS_DAO"}</span>
                </button>

                <a
                  href={project.drive_folder_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-2 px-3.5 py-2 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold rounded-xl border border-slate-700 transition-colors"
                >
                  <FolderTree size={14} className="text-indigo-400" />
                  <span>Drive</span>
                  <ExternalLink size={12} />
                </a>
              </>
            )}
            <button
              id="launch-analysis"
              onClick={launchAnalysis}
              disabled={!canAnalyze}
              title={docs.length === 0 ? "Déposez d'abord le DAO (PDF)" : undefined}
              className="inline-flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white text-xs font-bold rounded-xl shadow-md shadow-blue-500/20 disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer transition-all"
            >
              {analyzing ? (
                <>
                  <div className="w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  <span>Analyse multi-agents en cours…</span>
                </>
              ) : analysis ? (
                <>
                  <RotateCw size={14} />
                  <span>Relancer l&apos;analyse IA</span>
                </>
              ) : (
                <>
                  <Sparkles size={14} />
                  <span>Lancer l&apos;analyse IA</span>
                </>
              )}
            </button>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-8 py-8 grid lg:grid-cols-[360px_1fr] gap-6">
        {/* Colonne gauche : documents + tâches */}
        <aside className="space-y-6">
          {needsAuth && (
            <a
              href={`${API}/login`}
              className="flex items-center gap-3 p-4 bg-blue-50 border border-blue-200 rounded-2xl text-sm text-blue-800 hover:bg-blue-100 transition-colors"
            >
              <LogIn size={18} />
              <span>Connectez-vous à Google Drive pour gérer les documents.</span>
            </a>
          )}

          {error && (
            <div className="flex gap-3 p-4 bg-rose-50 border border-rose-200 rounded-2xl text-sm text-rose-800">
              <AlertTriangle size={18} className="shrink-0 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {!analyzing && project.analyse_error && (
            <div className="flex gap-3 p-4 bg-amber-50 border border-amber-200 rounded-2xl text-xs text-amber-800 break-words">
              <AlertTriangle size={16} className="shrink-0 mt-0.5" />
              <span>{project.analyse_error}</span>
            </div>
          )}

          <section className="bg-white border border-slate-200 rounded-2xl p-5 shadow-sm">
            <h2 className="text-sm font-bold text-slate-900 flex items-center gap-2 mb-4">
              <FileText size={16} className="text-blue-600" /> Dossier d&apos;appel d&apos;offres
            </h2>

            <div
              id="dao-dropzone"
              role="button"
              tabIndex={0}
              onClick={() => !needsAuth && fileInput.current?.click()}
              onKeyDown={(e) => e.key === "Enter" && fileInput.current?.click()}
              onDragOver={(e) => {
                e.preventDefault();
                setDragOver(true);
              }}
              onDragLeave={() => setDragOver(false)}
              onDrop={(e) => {
                e.preventDefault();
                setDragOver(false);
                if (!needsAuth) uploadFiles(e.dataTransfer.files);
              }}
              className={`border-2 border-dashed rounded-xl p-6 text-center cursor-pointer transition-all ${
                dragOver ? "border-blue-500 bg-blue-50 scale-[1.02]" : "border-slate-200 hover:border-blue-400 hover:bg-slate-50"
              } ${needsAuth ? "opacity-50 cursor-not-allowed" : ""}`}
            >
              {uploading ? (
                <div className="w-6 h-6 border-2 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto" />
              ) : (
                <UploadCloud size={28} className="mx-auto text-blue-500" />
              )}
              <p className="text-sm font-semibold text-slate-700 mt-2">
                {uploading ? "Téléversement vers Drive…" : "Déposez le DAO (PDF)"}
              </p>
              <p className="text-xs text-slate-400 mt-1">RC, CCAP, CCTP, BPU… — max 30 Mo par fichier</p>
              <input
                ref={fileInput}
                type="file"
                accept="application/pdf,.pdf"
                multiple
                hidden
                onChange={(e) => e.target.files && uploadFiles(e.target.files)}
              />
            </div>

            <ul className="mt-4 space-y-2">
              {docs.map((d) => (
                <li key={d.id} className="flex items-center gap-2 p-2.5 rounded-lg bg-slate-50 border border-slate-100 text-xs">
                  <FileText size={14} className="text-rose-500 shrink-0" />
                  <span className="truncate flex-1 text-slate-700">{d.name}</span>
                  {d.webViewLink && (
                    <a href={d.webViewLink} target="_blank" rel="noopener noreferrer" className="text-slate-400 hover:text-blue-600">
                      <ExternalLink size={12} />
                    </a>
                  )}
                </li>
              ))}
              {docs.length === 0 && !needsAuth && (
                <li className="text-xs text-slate-400 italic text-center">Aucun document déposé pour l&apos;instant.</li>
              )}
            </ul>
          </section>

          {tasks.length > 0 && (
            <section className="bg-white border border-slate-200 rounded-2xl p-5 shadow-sm">
              <div className="flex items-center justify-between mb-3">
                <h2 className="text-sm font-bold text-slate-900 flex items-center gap-2">
                  <ListTodo size={16} className="text-indigo-600" /> Tâches de montage
                </h2>
                <span className="text-xs text-slate-500">{done}/{tasks.length}</span>
              </div>
              <div className="h-1.5 bg-slate-100 rounded-full overflow-hidden mb-4">
                <div
                  className="h-full bg-gradient-to-r from-emerald-500 to-teal-400 transition-all duration-500"
                  style={{ width: `${(done / tasks.length) * 100}%` }}
                />
              </div>
              <ul className="space-y-2 max-h-[520px] overflow-y-auto pr-1">
                {tasks.map((t) => (
                  <li key={t.id}>
                    <button
                      id={`task-${t.id}`}
                      onClick={() => toggleTask(t)}
                      className="w-full text-left flex gap-2.5 p-2.5 rounded-lg hover:bg-slate-50 transition-colors cursor-pointer"
                    >
                      {t.statut === "TERMINEE" ? (
                        <CheckCircle2 size={16} className="text-emerald-500 shrink-0 mt-0.5" />
                      ) : t.statut === "EN_COURS" ? (
                        <Clock size={16} className="text-blue-500 shrink-0 mt-0.5" />
                      ) : (
                        <Circle size={16} className="text-slate-300 shrink-0 mt-0.5" />
                      )}
                      <span className="flex-1 min-w-0">
                        <span className={`block text-xs ${t.statut === "TERMINEE" ? "line-through text-slate-400" : "text-slate-700"}`}>
                          {t.titre}
                        </span>
                        <span className={`inline-block mt-1 text-[10px] px-1.5 py-0.5 rounded border ${CAT_CLS[t.categorie] ?? ""}`}>
                          {t.categorie}
                        </span>
                      </span>
                    </button>
                  </li>
                ))}
              </ul>
            </section>
          )}
        </aside>

        {/* Colonne droite : Tabs (Conformité RAG / Offre Technique / Synthèse IA) */}
        <main className="space-y-6">
          {/* Navigation des Onglets */}
          <div className="flex items-center gap-2 border-b border-slate-200 pb-2 flex-wrap">
            <button
              onClick={() => setActiveTab("compliance")}
              className={`px-4 py-2.5 rounded-xl text-xs font-bold transition-all flex items-center gap-2 cursor-pointer ${
                activeTab === "compliance"
                  ? "bg-indigo-600 text-white shadow-sm"
                  : "bg-white text-slate-600 hover:text-slate-900 border border-slate-200"
              }`}
            >
              <ShieldCheck size={16} />
              <span>Conformité & Coffre-fort RAG</span>
              {compliance && (
                <span className={`px-2 py-0.5 rounded-full text-[10px] font-mono ${
                  activeTab === "compliance" ? "bg-white/20 text-white" : "bg-slate-100 text-slate-700"
                }`}>
                  {compliance.matched_count}/{compliance.total_requirements}
                </span>
              )}
            </button>

            <button
              onClick={() => setActiveTab("technical")}
              className={`px-4 py-2.5 rounded-xl text-xs font-bold transition-all flex items-center gap-2 cursor-pointer ${
                activeTab === "technical"
                  ? "bg-indigo-600 text-white shadow-sm"
                  : "bg-white text-slate-600 hover:text-slate-900 border border-slate-200"
              }`}
            >
              <FileText size={16} />
              <span>Offre Technique & Mémoire</span>
              {technicalOffer && (technicalOffer.sections || []).length > 0 && (
                <span className={`px-2 py-0.5 rounded-full text-[10px] font-mono ${
                  activeTab === "technical" ? "bg-white/20 text-white" : "bg-slate-100 text-slate-700"
                }`}>
                  {(technicalOffer.sections || []).filter(s => s.statut === "GENERE" || s.statut === "VALIDE").length}/{(technicalOffer.sections || []).length}
                </span>
              )}
            </button>

            <button
              onClick={() => setActiveTab("analysis")}
              className={`px-4 py-2.5 rounded-xl text-xs font-bold transition-all flex items-center gap-2 cursor-pointer ${
                activeTab === "analysis"
                  ? "bg-indigo-600 text-white shadow-sm"
                  : "bg-white text-slate-600 hover:text-slate-900 border border-slate-200"
              }`}
            >
              <Bot size={16} />
              <span>Synthèse Multi-Agents</span>
              {analysis && (
                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
              )}
            </button>
          </div>

          {analyzing ? (
            <div className="bg-white border border-slate-200 rounded-2xl p-10 shadow-sm text-center space-y-5">
              <div className="relative w-20 h-20 mx-auto">
                <div className="absolute inset-0 rounded-full bg-blue-500/20 animate-ping" />
                <div className="relative w-20 h-20 rounded-full bg-gradient-to-br from-blue-600 to-indigo-600 flex items-center justify-center text-white shadow-lg">
                  <Bot size={34} />
                </div>
              </div>
              <h2 className="text-lg font-bold text-slate-900">Les agents analysent votre DAO…</h2>
              <div className="flex flex-col sm:flex-row justify-center gap-3 text-xs">
                <span className="px-3 py-1.5 rounded-lg bg-emerald-50 text-emerald-700 border border-emerald-100">⚖️ Conformité administrative</span>
                <span className="px-3 py-1.5 rounded-lg bg-indigo-50 text-indigo-700 border border-indigo-100">🏗️ Ingénierie technique</span>
              </div>
              <p className="text-xs text-slate-400">Cela prend généralement 30 secondes à 2 minutes selon la taille des documents.</p>
            </div>
          ) : activeTab === "technical" ? (
            technicalOffer ? (
              <TechnicalOfferEditor
                projectId={projectId}
                data={technicalOffer}
                companyDocs={companyDocs}
                onRefresh={async () => {
                  await fetchTechnicalOffer();
                  await fetchProject();
                }}
              />
            ) : (
              <div className="bg-white border border-slate-200 rounded-2xl p-12 text-center text-xs text-slate-400 animate-pulse">
                Chargement de l'offre technique...
              </div>
            )
          ) : activeTab === "analysis" ? (
            analysis ? (
              <AnalysisResults data={analysis} />
            ) : (
              <div className="bg-white border border-slate-200 rounded-2xl p-10 shadow-sm text-center space-y-4">
                <div className="w-16 h-16 bg-blue-50 text-blue-600 rounded-2xl flex items-center justify-center mx-auto border border-blue-100">
                  <Bot size={32} />
                </div>
                <h2 className="text-lg font-bold text-slate-900">Prêt pour l&apos;analyse multi-agents</h2>
                <ol className="text-sm text-slate-500 max-w-md mx-auto text-left space-y-2">
                  <li><b className="text-slate-700">1.</b> Déposez les pièces du DAO (PDF) à gauche.</li>
                  <li><b className="text-slate-700">2.</b> Cliquez sur <b>« Lancer l&apos;analyse IA »</b> en haut à droite.</li>
                  <li><b className="text-slate-700">3.</b> Les agents extraient critères, lots, calendrier et plan du mémoire technique.</li>
                </ol>
              </div>
            )
          ) : (
            /* Onglet Conformité & Coffre-fort RAG */
            <div className="space-y-6">
              {/* Statut Global de Conformité */}
              {compliance && (
                <div
                  className={`border rounded-2xl p-6 shadow-sm space-y-4 ${
                    compliance.global_status === "NON_CONFORME"
                      ? "bg-rose-50/70 border-rose-200 text-rose-950"
                      : compliance.global_status === "ATTENTION"
                      ? "bg-amber-50/70 border-amber-200 text-amber-950"
                      : "bg-emerald-50/70 border-emerald-200 text-emerald-950"
                  }`}
                >
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                    <div className="flex items-start gap-3">
                      <div className={`p-3 rounded-xl shrink-0 ${
                        compliance.global_status === "NON_CONFORME"
                          ? "bg-rose-100 text-rose-600"
                          : compliance.global_status === "ATTENTION"
                          ? "bg-amber-100 text-amber-600"
                          : "bg-emerald-100 text-emerald-600"
                      }`}>
                        {compliance.global_status === "NON_CONFORME" ? (
                          <ShieldAlert size={26} />
                        ) : (
                          <ShieldCheck size={26} />
                        )}
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <span className={`text-[11px] font-extrabold uppercase px-2.5 py-0.5 rounded-full border ${
                            compliance.global_status === "NON_CONFORME"
                              ? "bg-rose-200/60 text-rose-900 border-rose-300"
                              : compliance.global_status === "ATTENTION"
                              ? "bg-amber-200/60 text-amber-900 border-amber-300"
                              : "bg-emerald-200/60 text-emerald-900 border-emerald-300"
                          }`}>
                            {compliance.global_status === "NON_CONFORME"
                              ? "Dossier Non Conforme — Risque d'Élimination"
                              : compliance.global_status === "ATTENTION"
                              ? "Dossier Incomplet — Pièces à Fournir"
                              : "Dossier 100% Conforme"}
                          </span>
                        </div>
                        <h2 className="text-base font-bold mt-1 text-slate-900">
                          {compliance.global_status === "NON_CONFORME"
                            ? "Certains critères éliminatoires ne sont pas satisfaits par votre coffre-fort."
                            : compliance.global_status === "ATTENTION"
                            ? "Toutes les pièces requises ne sont pas encore présentes dans votre coffre-fort."
                            : "Toutes les pièces nécessaires sont présentes et à jour dans votre coffre-fort !"}
                        </h2>
                        <p className="text-xs text-slate-600 mt-0.5">
                          {compliance.company_documents_count} document(s) analysé(s) dans votre coffre-fort d&apos;entreprise.
                        </p>
                      </div>
                    </div>

                    <Link
                      href="/company"
                      className="inline-flex items-center gap-2 px-4 py-2 bg-white hover:bg-slate-50 text-slate-700 text-xs font-bold rounded-xl border border-slate-300 shadow-sm shrink-0 transition-colors"
                    >
                      <FolderTree size={14} className="text-indigo-600" />
                      <span>Gérer le Coffre-fort RAG</span>
                    </Link>
                  </div>

                  {/* Tuiles Statistiques */}
                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-2">
                    <div className="bg-white/90 p-3.5 rounded-xl border border-slate-200/60 shadow-xs">
                      <p className="text-[11px] text-slate-500 font-medium">Exigences DAO</p>
                      <p className="text-lg font-black text-slate-900 mt-0.5">{compliance.total_requirements}</p>
                    </div>
                    <div className="bg-white/90 p-3.5 rounded-xl border border-slate-200/60 shadow-xs">
                      <p className="text-[11px] text-emerald-600 font-medium">Conformes RAG</p>
                      <p className="text-lg font-black text-emerald-700 mt-0.5">{compliance.matched_count}</p>
                    </div>
                    <div className="bg-white/90 p-3.5 rounded-xl border border-slate-200/60 shadow-xs">
                      <p className="text-[11px] text-amber-600 font-medium">À Renouveler / Exp.</p>
                      <p className="text-lg font-black text-amber-700 mt-0.5">{compliance.expired_count}</p>
                    </div>
                    <div className="bg-white/90 p-3.5 rounded-xl border border-slate-200/60 shadow-xs">
                      <p className="text-[11px] text-rose-600 font-medium">Manquants</p>
                      <p className="text-lg font-black text-rose-700 mt-0.5">{compliance.missing_count}</p>
                    </div>
                  </div>
                </div>
              )}

              {/* Action Bar : Transfert 1-Clic vers Drive */}
              <div className="bg-white border border-slate-200 rounded-2xl p-5 shadow-sm space-y-3">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                  <div className="flex items-center gap-3">
                    <div className="p-2.5 bg-indigo-50 border border-indigo-100 rounded-xl text-indigo-600 shrink-0">
                      <FolderCheck size={22} />
                    </div>
                    <div>
                      <h3 className="text-sm font-bold text-slate-900">
                        Synchronisation Google Drive — Offre Administrative
                      </h3>
                      <p className="text-xs text-slate-500 mt-0.5">
                        Copie automatique des pièces conformes du coffre-fort vers le dossier du projet :{" "}
                        <code className="text-indigo-600 font-mono bg-indigo-50 px-1.5 py-0.5 rounded">
                          02_Offre_Administrative
                        </code>
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 shrink-0">
                    {compliance?.offre_admin_folder_url && (
                      <a
                        href={compliance.offre_admin_folder_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="p-2.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-xl text-xs font-semibold transition-colors flex items-center gap-1.5"
                        title="Ouvrir le dossier 02_Offre_Administrative sur Drive"
                      >
                        <ExternalLink size={14} />
                        <span className="hidden sm:inline">Dossier Drive</span>
                      </a>
                    )}

                    <button
                      onClick={handleCopyRagDocs}
                      disabled={copyingDocs || !compliance || compliance.matched_count === 0}
                      className="px-4 py-2.5 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white text-xs font-bold rounded-xl shadow-md shadow-indigo-500/20 transition-all flex items-center gap-2 cursor-pointer"
                    >
                      {copyingDocs ? (
                        <>
                          <RefreshCw size={14} className="animate-spin" />
                          <span>Copie sur Google Drive...</span>
                        </>
                      ) : (
                        <>
                          <Copy size={14} />
                          <span>Transférer les pièces conformes ({compliance?.matched_count || 0})</span>
                        </>
                      )}
                    </button>
                  </div>
                </div>

                {copyResultMsg && (
                  <div className={`p-3 rounded-xl text-xs flex items-center justify-between gap-2 ${
                    copyResultMsg.error ? "bg-rose-50 text-rose-800 border border-rose-200" : "bg-emerald-50 text-emerald-800 border border-emerald-200"
                  }`}>
                    <div className="flex items-center gap-2">
                      {copyResultMsg.error ? <AlertCircle size={16} /> : <CheckCircle2 size={16} />}
                      <span>{copyResultMsg.text}</span>
                    </div>
                    {copyResultMsg.folderUrl && (
                      <a
                        href={copyResultMsg.folderUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="underline font-bold hover:text-emerald-950 flex items-center gap-1 shrink-0"
                      >
                        <span>Voir sur Google Drive</span>
                        <ExternalLink size={12} />
                      </a>
                    )}
                  </div>
                )}
              </div>

              {/* Table des exigences et matching */}
              <div className="bg-white border border-slate-200 rounded-2xl p-5 shadow-sm space-y-4">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-slate-100">
                  <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2">
                    <FileText size={16} className="text-indigo-600" />
                    Détail du Contrôle d&apos;Éligibilité ({compliance?.total_requirements || 0})
                  </h3>

                  {/* Filtres */}
                  <div className="flex items-center gap-1.5 flex-wrap">
                    <button
                      onClick={() => setComplianceFilter("ALL")}
                      className={`px-2.5 py-1 rounded-lg text-[11px] font-bold transition-colors cursor-pointer ${
                        complianceFilter === "ALL" ? "bg-slate-900 text-white" : "bg-slate-100 text-slate-600 hover:bg-slate-200"
                      }`}
                    >
                      Tous ({compliance?.total_requirements || 0})
                    </button>
                    <button
                      onClick={() => setComplianceFilter("ELIMINATOIRE")}
                      className={`px-2.5 py-1 rounded-lg text-[11px] font-bold transition-colors cursor-pointer ${
                        complianceFilter === "ELIMINATOIRE" ? "bg-rose-600 text-white" : "bg-rose-50 text-rose-700 hover:bg-rose-100 border border-rose-200"
                      }`}
                    >
                      Éliminatoires ({(compliance?.requirements || []).filter((r) => r.is_eliminatory).length})
                    </button>
                    <button
                      onClick={() => setComplianceFilter("CONFORME")}
                      className={`px-2.5 py-1 rounded-lg text-[11px] font-bold transition-colors cursor-pointer ${
                        complianceFilter === "CONFORME" ? "bg-emerald-600 text-white" : "bg-emerald-50 text-emerald-700 hover:bg-emerald-100 border border-emerald-200"
                      }`}
                    >
                      Conformes ({compliance?.matched_count || 0})
                    </button>
                    <button
                      onClick={() => setComplianceFilter("MANQUANT")}
                      className={`px-2.5 py-1 rounded-lg text-[11px] font-bold transition-colors cursor-pointer ${
                        complianceFilter === "MANQUANT" ? "bg-slate-800 text-white" : "bg-slate-100 text-slate-700 hover:bg-slate-200"
                      }`}
                    >
                      Manquants ({compliance?.missing_count || 0})
                    </button>
                  </div>
                </div>

                {complianceLoading ? (
                  <div className="h-40 bg-slate-100 rounded-xl animate-pulse" />
                ) : !compliance || !compliance.requirements || compliance.requirements.length === 0 ? (
                  <div className="p-8 text-center text-xs text-slate-400">
                    Aucune exigence administrative extraite pour ce projet. Lancez l&apos;analyse IA sur le DAO.
                  </div>
                ) : (
                  <div className="space-y-3">
                    {(compliance.requirements || [])
                      .filter((r) => {
                        if (complianceFilter === "ELIMINATOIRE") return r.is_eliminatory;
                        if (complianceFilter === "CONFORME") return r.status === "CONFORME";
                        if (complianceFilter === "A_RENOUVELER") return r.status === "A_RENOUVELER";
                        if (complianceFilter === "MANQUANT") return r.status === "MANQUANT";
                        return true;
                      })
                      .map((reqItem, idx) => {
                        const isMatch = reqItem.status === "CONFORME";
                        const isRenew = reqItem.status === "A_RENOUVELER";
                        return (
                          <div
                            key={idx}
                            className={`p-4 rounded-xl border transition-all flex flex-col md:flex-row md:items-center justify-between gap-4 ${
                              reqItem.is_eliminatory && !isMatch
                                ? "bg-rose-50/40 border-rose-200 hover:border-rose-300"
                                : isMatch
                                ? "bg-emerald-50/30 border-emerald-200 hover:border-emerald-300"
                                : isRenew
                                ? "bg-amber-50/30 border-amber-200 hover:border-amber-300"
                                : "bg-slate-50/40 border-slate-200 hover:border-slate-300"
                            }`}
                          >
                            {/* Colonne gauche : exigence DAO */}
                            <div className="space-y-1.5 flex-1 min-w-0">
                              <div className="flex items-center gap-2 flex-wrap">
                                {reqItem.is_eliminatory && (
                                  <span className="px-2 py-0.5 rounded text-[10px] font-extrabold uppercase bg-rose-600 text-white tracking-wider">
                                    Critère Éliminatoire
                                  </span>
                                )}
                                <span className="text-[10px] font-bold text-slate-500 uppercase">
                                  {reqItem.category}
                                </span>
                              </div>
                              <p className="text-xs font-bold text-slate-900 leading-snug">
                                {reqItem.requirement}
                              </p>
                              <p className="text-[11px] text-slate-500 flex items-center gap-1">
                                <Clock size={11} className="text-slate-400" />
                                <span>{reqItem.comment}</span>
                              </p>
                            </div>

                            {/* Colonne droite : document du coffre-fort associé */}
                            <div className="flex flex-col md:items-end justify-between gap-2 shrink-0 md:w-72">
                              {/* Statut Badge */}
                              <div className="flex items-center gap-2">
                                <span
                                  className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-bold border ${
                                    isMatch
                                      ? "bg-emerald-100 text-emerald-800 border-emerald-200"
                                      : isRenew
                                      ? "bg-amber-100 text-amber-800 border-amber-200"
                                      : "bg-rose-100 text-rose-800 border-rose-200"
                                  }`}
                                >
                                  {isMatch ? (
                                    <>
                                      <CheckCircle2 size={12} />
                                      <span>Conforme</span>
                                    </>
                                  ) : isRenew ? (
                                    <>
                                      <AlertTriangle size={12} />
                                      <span>À renouveler</span>
                                    </>
                                  ) : (
                                    <>
                                      <AlertCircle size={12} />
                                      <span>Manquant</span>
                                    </>
                                  )}
                                </span>
                              </div>

                              {/* Infos document */}
                              {reqItem.matched_document ? (
                                <div className="text-left md:text-right space-y-0.5">
                                  <p className="text-xs font-semibold text-slate-800 truncate max-w-xs">
                                    {reqItem.matched_document.titre}
                                  </p>
                                  {reqItem.matched_document.web_view_link && (
                                    <a
                                      href={reqItem.matched_document.web_view_link}
                                      target="_blank"
                                      rel="noopener noreferrer"
                                      className="inline-flex items-center gap-1 text-[11px] font-semibold text-indigo-600 hover:underline"
                                    >
                                      <span>Voir sur Drive</span>
                                      <ExternalLink size={10} />
                                    </a>
                                  )}
                                </div>
                              ) : (
                                <Link
                                  href="/company"
                                  className="inline-flex items-center gap-1 text-[11px] font-semibold text-indigo-600 hover:underline"
                                >
                                  <span>Ajouter au Coffre-fort</span>
                                  <ArrowRight size={10} />
                                </Link>
                              )}
                            </div>
                          </div>
                        );
                      })}
                  </div>
                )}
              </div>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
