'use client';

import { useEffect, useState, useRef } from "react";
import { CompanyProfile, CompanyDocument, TypeDocumentEntreprise } from "@/components/dashboard/types";
import {
  Building2,
  FileText,
  Upload,
  Calendar,
  Sparkles,
  UserCheck,
  Award,
  Truck,
  CheckCircle2,
  Plus,
  ExternalLink,
  FolderTree,
  AlertCircle,
  Clock,
  ShieldCheck,
  FolderCheck,
  RefreshCw,
} from "lucide-react";

export default function CompanySettingsPage() {
  const [profile, setProfile] = useState<CompanyProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [showAddDocModal, setShowAddDocModal] = useState(false);

  // Initialisation & Synchronisation Drive Workspace
  const [initializingWorkspace, setInitializingWorkspace] = useState(false);
  const [syncingDrive, setSyncingDrive] = useState(false);
  const [analyzingAllCVs, setAnalyzingAllCVs] = useState(false);
  const [analyzingDocId, setAnalyzingDocId] = useState<string | null>(null);
  const [workspaceMsg, setWorkspaceMsg] = useState<{ text: string; error?: boolean } | null>(null);
  const [selectedCategory, setSelectedCategory] = useState<string>("ALL");

  // Formulaire d'ajout / téléversement de document RAG
  const [docFile, setDocFile] = useState<File | null>(null);
  const [docTitre, setDocTitre] = useState("");
  const [docType, setDocType] = useState<TypeDocumentEntreprise>("ATTESTATION_FISCALE_SOCIALE");
  const [docDesc, setDocDesc] = useState("");
  const [docExp, setDocExp] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  const fetchProfile = async () => {
    try {
      setLoading(true);
      const res = await fetch(`${apiUrl}/api/company`, { credentials: "include" });
      if (res.ok) {
        const data = await res.json();
        setProfile(data);
      }
    } catch (err) {
      console.error("Erreur récupération profil:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProfile();
  }, []);

  const handleInitWorkspace = async () => {
    setInitializingWorkspace(true);
    setWorkspaceMsg(null);
    try {
      const res = await fetch(`${apiUrl}/api/workspace/init`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
      });
      const data = await res.json();
      if (res.ok) {
        setProfile(data);
        setWorkspaceMsg({ text: "Espace AGENTS_DAO_WORKSPACE créé avec succès sur Google Drive !" });
      } else {
        setWorkspaceMsg({ text: data.error || "Impossible d'initialiser l'espace Drive.", error: true });
      }
    } catch {
      setWorkspaceMsg({ text: "Erreur de connexion à l'API.", error: true });
    } finally {
      setInitializingWorkspace(false);
    }
  };

  const handleSyncDrive = async () => {
    setSyncingDrive(true);
    setWorkspaceMsg(null);
    try {
      const res = await fetch(`${apiUrl}/api/company/sync-drive`, {
        method: "POST",
        credentials: "include",
      });
      const data = await res.json();
      if (res.ok) {
        setWorkspaceMsg({ text: data.message });
        await fetchProfile();
      } else {
        setWorkspaceMsg({ text: data.error || "Erreur de synchronisation Drive.", error: true });
      }
    } catch {
      setWorkspaceMsg({ text: "Erreur de communication avec l'API.", error: true });
    } finally {
      setSyncingDrive(false);
    }
  };

  const handleAnalyzeAllCVs = async () => {
    setAnalyzingAllCVs(true);
    setWorkspaceMsg(null);
    try {
      const res = await fetch(`${apiUrl}/api/company/analyze-all-cvs`, {
        method: "POST",
        credentials: "include",
      });
      const data = await res.json();
      if (res.ok) {
        setWorkspaceMsg({ text: data.message });
        await fetchProfile();
      } else {
        setWorkspaceMsg({ text: data.error || "Erreur lors de l'analyse des CVs.", error: true });
      }
    } catch {
      setWorkspaceMsg({ text: "Erreur de communication avec l'API.", error: true });
    } finally {
      setAnalyzingAllCVs(false);
    }
  };

  const handleAnalyzeSingleCV = async (docId: string) => {
    setAnalyzingDocId(docId);
    setWorkspaceMsg(null);
    try {
      const res = await fetch(`${apiUrl}/api/company/documents/${docId}/analyze`, {
        method: "POST",
        credentials: "include",
      });
      const data = await res.json();
      if (res.ok) {
        setWorkspaceMsg({ text: data.message });
        await fetchProfile();
      } else {
        setWorkspaceMsg({ text: data.error || "Erreur lors de l'analyse du CV.", error: true });
      }
    } catch {
      setWorkspaceMsg({ text: "Erreur de communication avec l'API.", error: true });
    } finally {
      setAnalyzingDocId(null);
    }
  };

  const handleUploadDocument = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!docFile && !docTitre.trim()) return;

    setSubmitting(true);
    setUploadError(null);
    try {
      if (docFile) {
        const formData = new FormData();
        formData.append("file", docFile);
        formData.append("titre", docTitre.trim() || docFile.name);
        formData.append("type_document", docType);
        formData.append("description", docDesc.trim());
        if (docExp) formData.append("date_expiration", docExp);

        const res = await fetch(`${apiUrl}/api/company/documents/upload`, {
          method: "POST",
          body: formData,
          credentials: "include",
        });

        if (!res.ok) {
          const errData = await res.json().catch(() => ({}));
          throw new Error(errData.error || "Erreur lors du téléversement du document");
        }
      } else {
        // Ajout sans fichier Drive
        const res = await fetch(`${apiUrl}/api/company/documents`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({
            titre: docTitre.trim(),
            type_doc: docType,
            description: docDesc.trim(),
            date_expiration: docExp ? new Date(docExp).toISOString() : undefined,
          }),
        });

        if (!res.ok) throw new Error("Erreur enregistrement métadonnées");
      }

      await fetchProfile();
      setShowAddDocModal(false);
      setDocFile(null);
      setDocTitre("");
      setDocDesc("");
      setDocExp("");
    } catch (err: any) {
      setUploadError(err.message || "Erreur inattendue");
    } finally {
      setSubmitting(false);
    }
  };

  const getDocTypeBadge = (type: TypeDocumentEntreprise) => {
    switch (type) {
      case "CV_EXPERT":
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-50 text-blue-700 border border-blue-200">
            <UserCheck size={12} />
            CV & Personnel
          </span>
        );
      case "REFERENCE_SERVICE_FAIT":
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
            <Award size={12} />
            Attestation de Service Fait
          </span>
        );
      case "JUSTIFICATIF_MATERIEL":
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-purple-50 text-purple-700 border border-purple-200">
            <Truck size={12} />
            Matériel & Engin
          </span>
        );
      case "ATTESTATION_FISCALE_SOCIALE":
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200">
            <ShieldCheck size={12} />
            Fiscale & Sociale (Quitus, CSS...)
          </span>
        );
      case "BILAN_FINANCIER":
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-teal-50 text-teal-700 border border-teal-200">
            <FileText size={12} />
            Bilan Financier
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-slate-100 text-slate-700 border border-slate-200">
            <FileText size={12} />
            Administratif (NINEA, RCCM...)
          </span>
        );
    }
  };

  const isExpired = (dateStr?: string) => {
    if (!dateStr) return false;
    return new Date(dateStr) < new Date();
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* En-tête */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6 border-b border-slate-200">
        <div>
          <h1 className="text-2xl font-black text-slate-900 tracking-tight flex items-center gap-3">
            <Building2 className="w-7 h-7 text-indigo-600" />
            Coffre-fort RAG & Référentiel Entreprise
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            Gérez vos pièces administratives permanentes et synchronisez-les avec votre Google Drive d'entreprise.
          </p>
        </div>

        <button
          onClick={() => setShowAddDocModal(true)}
          className="inline-flex items-center justify-center gap-2 px-5 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-semibold rounded-xl shadow-md shadow-indigo-500/20 transition-all cursor-pointer shrink-0"
        >
          <Plus size={18} />
          <span>Ajouter une pièce RAG</span>
        </button>
      </div>

      {/* Bannière Espace Google Drive */}
      <div className="bg-gradient-to-r from-slate-900 via-indigo-950 to-slate-900 border border-slate-800 rounded-2xl p-6 text-white shadow-lg space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="flex items-start gap-3">
            <div className="p-3 bg-indigo-600/30 border border-indigo-500/30 rounded-xl text-indigo-300 shrink-0">
              <FolderTree size={24} />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="text-xs font-mono font-bold uppercase tracking-wider text-indigo-400">
                  Arborescence Google Drive
                </span>
                {profile?.workspace_folder_id ? (
                  <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
                    Connecté & Opérationnel
                  </span>
                ) : (
                  <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-500/20 text-amber-300 border border-amber-500/30">
                    À initialiser
                  </span>
                )}
              </div>
              <h2 className="text-lg font-bold text-white mt-1">
                Porte d&apos;entrée principale : <code className="text-indigo-300 bg-slate-800/80 px-2 py-0.5 rounded font-mono text-sm">AGENTS_DAO_WORKSPACE</code>
              </h2>
              <p className="text-xs text-slate-300 mt-1 max-w-2xl leading-relaxed">
                Ce dossier racine sur votre Drive regroupe vos pièces RAG permanentes (<code className="text-slate-200">01_COFFRE_FORT_ENTREPRISE</code>) et accueille automatiquement tous vos dossiers de projets (<code className="text-slate-200">02_PROJETS_DAO</code>).
              </p>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-3 shrink-0">
            {profile?.workspace_folder_url && (
              <>
                <button
                  type="button"
                  onClick={handleSyncDrive}
                  disabled={syncingDrive || analyzingAllCVs}
                  title="Scanner le dossier 01_COFFRE_FORT_ENTREPRISE sur Google Drive et importer automatiquement les CVs et documents"
                  className="inline-flex items-center gap-2 px-4 py-2.5 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold rounded-xl border border-slate-700 transition-all cursor-pointer shadow-sm disabled:opacity-50"
                >
                  <RefreshCw size={14} className={syncingDrive ? "animate-spin text-indigo-400" : "text-slate-400"} />
                  <span>{syncingDrive ? "Synchronisation..." : "Synchroniser depuis Google Drive"}</span>
                </button>

                <button
                  type="button"
                  onClick={handleAnalyzeAllCVs}
                  disabled={analyzingAllCVs || syncingDrive}
                  title="Lire le contenu de tous les CVs déposés sur Drive et indexer leurs compétences pour les DAO"
                  className="inline-flex items-center gap-2 px-4 py-2.5 bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 text-white text-xs font-semibold rounded-xl transition-all cursor-pointer shadow-md shadow-indigo-600/25 disabled:opacity-50"
                >
                  <Sparkles size={14} className={analyzingAllCVs ? "animate-spin text-amber-300" : "text-amber-300"} />
                  <span>{analyzingAllCVs ? "Lecture IA des CVs..." : "Analyser tous les CVs par l'IA"}</span>
                </button>

                <a
                  href={profile.workspace_folder_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-2 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold rounded-xl transition-all shadow-md shadow-indigo-600/30"
                >
                  <FolderCheck size={16} />
                  <span>Ouvrir sur Drive</span>
                  <ExternalLink size={12} />
                </a>
              </>
            )}

            {!profile?.workspace_folder_url && (
              <button
                onClick={handleInitWorkspace}
                disabled={initializingWorkspace}
                className="inline-flex items-center gap-2 px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white text-xs font-bold rounded-xl transition-all shadow-md shadow-indigo-600/30 cursor-pointer"
              >
                {initializingWorkspace ? (
                  <>
                    <RefreshCw size={14} className="animate-spin" />
                    <span>Création de l'arborescence...</span>
                  </>
                ) : (
                  <>
                    <Sparkles size={14} />
                    <span>Initialiser l&apos;arborescence Drive (1 clic)</span>
                  </>
                )}
              </button>
            )}
          </div>
        </div>

        {workspaceMsg && (
          <div className={`p-3 rounded-xl text-xs flex items-center gap-2 ${workspaceMsg.error ? "bg-red-500/20 text-red-200 border border-red-500/30" : "bg-emerald-500/20 text-emerald-200 border border-emerald-500/30"}`}>
            {workspaceMsg.error ? <AlertCircle size={14} /> : <CheckCircle2 size={14} />}
            <span>{workspaceMsg.text}</span>
          </div>
        )}
      </div>

      {/* Fiche Entreprise */}
      {profile && (
        <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm">
          <h2 className="text-sm font-bold uppercase tracking-wider text-slate-500 mb-4 flex items-center gap-2">
            <CheckCircle2 size={16} className="text-indigo-600" />
            Fiche Signalétique de l&apos;Entreprise
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="p-4 bg-slate-50 rounded-xl border border-slate-100">
              <p className="text-xs text-slate-400 font-medium">Raison Sociale</p>
              <p className="text-sm font-bold text-slate-900 mt-1">{profile.nom}</p>
            </div>
            <div className="p-4 bg-slate-50 rounded-xl border border-slate-100">
              <p className="text-xs text-slate-400 font-medium">Numéro NINEA</p>
              <p className="text-sm font-mono font-bold text-slate-900 mt-1">{profile.ninea || "Non renseigné"}</p>
            </div>
            <div className="p-4 bg-slate-50 rounded-xl border border-slate-100">
              <p className="text-xs text-slate-400 font-medium">Registre de Commerce (RCCM)</p>
              <p className="text-sm font-mono font-bold text-slate-900 mt-1">{profile.rccm || "Non renseigné"}</p>
            </div>
            <div className="p-4 bg-slate-50 rounded-xl border border-slate-100">
              <p className="text-xs text-slate-400 font-medium">Contact Principal</p>
              <p className="text-sm font-bold text-slate-900 mt-1 truncate">{profile.email || "Non renseigné"}</p>
            </div>
          </div>
        </div>
      )}

      {/* Documents RAG indexés */}
      <div className="space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <h2 className="text-lg font-bold text-slate-900 flex items-center gap-2">
            <Sparkles size={18} className="text-amber-500" />
            Pièces Justificatives Indexées ({profile?.documents?.length || 0})
          </h2>

          <div className="flex flex-wrap items-center gap-1.5">
            <button
              onClick={() => setSelectedCategory("ALL")}
              className={`px-2.5 py-1 rounded-lg text-xs font-bold transition-colors cursor-pointer ${
                selectedCategory === "ALL" ? "bg-slate-900 text-white" : "bg-slate-100 text-slate-600 hover:bg-slate-200"
              }`}
            >
              Tous ({profile?.documents?.length || 0})
            </button>
            <button
              onClick={() => setSelectedCategory("CV_EXPERT")}
              className={`px-2.5 py-1 rounded-lg text-xs font-bold transition-colors cursor-pointer ${
                selectedCategory === "CV_EXPERT" ? "bg-blue-600 text-white" : "bg-blue-50 text-blue-700 hover:bg-blue-100 border border-blue-200"
              }`}
            >
              CV & Personnel ({profile?.documents?.filter(d => d.type_doc === "CV_EXPERT").length || 0})
            </button>
            <button
              onClick={() => setSelectedCategory("ATTESTATION_FISCALE_SOCIALE")}
              className={`px-2.5 py-1 rounded-lg text-xs font-bold transition-colors cursor-pointer ${
                selectedCategory === "ATTESTATION_FISCALE_SOCIALE" ? "bg-amber-600 text-white" : "bg-amber-50 text-amber-700 hover:bg-amber-100 border border-amber-200"
              }`}
            >
              Fiscale & Sociale ({profile?.documents?.filter(d => d.type_doc === "ATTESTATION_FISCALE_SOCIALE").length || 0})
            </button>
            <button
              onClick={() => setSelectedCategory("REFERENCE_SERVICE_FAIT")}
              className={`px-2.5 py-1 rounded-lg text-xs font-bold transition-colors cursor-pointer ${
                selectedCategory === "REFERENCE_SERVICE_FAIT" ? "bg-emerald-600 text-white" : "bg-emerald-50 text-emerald-700 hover:bg-emerald-100 border border-emerald-200"
              }`}
            >
              Références Chantiers ({profile?.documents?.filter(d => d.type_doc === "REFERENCE_SERVICE_FAIT").length || 0})
            </button>
            <button
              onClick={() => setSelectedCategory("JUSTIFICATIF_MATERIEL")}
              className={`px-2.5 py-1 rounded-lg text-xs font-bold transition-colors cursor-pointer ${
                selectedCategory === "JUSTIFICATIF_MATERIEL" ? "bg-purple-600 text-white" : "bg-purple-50 text-purple-700 hover:bg-purple-100 border border-purple-200"
              }`}
            >
              Matériel & Engins ({profile?.documents?.filter(d => d.type_doc === "JUSTIFICATIF_MATERIEL").length || 0})
            </button>
          </div>
        </div>

        {loading ? (
          <div className="h-40 bg-slate-200 rounded-2xl animate-pulse" />
        ) : !profile?.documents || profile.documents.length === 0 ? (
          <div className="bg-white border border-slate-200 rounded-2xl p-8 text-center space-y-3">
            <p className="text-sm text-slate-600 font-medium">
              Aucun document d&apos;entreprise indexé pour le moment dans le coffre-fort.
            </p>
            <p className="text-xs text-slate-400 max-w-md mx-auto">
              Téléversez vos CVs d&apos;experts, attestations fiscales, NINEA, RCCM ou références de chantiers pour que la plateforme pré-remplisse automatiquement vos offres administratives et techniques.
            </p>
            <button
              onClick={() => setShowAddDocModal(true)}
              className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-50 text-indigo-700 hover:bg-indigo-100 border border-indigo-200 rounded-xl text-xs font-bold transition-colors cursor-pointer"
            >
              <Upload size={14} />
              <span>Téléverser une première pièce</span>
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {profile.documents
              .filter((doc) => selectedCategory === "ALL" || doc.type_doc === selectedCategory)
              .map((doc) => {
              const expired = isExpired(doc.date_expiration);
              let meta: any = null;
              if (doc.metadonnees_json) {
                try { meta = JSON.parse(doc.metadonnees_json); } catch {}
              }
              return (
                <div
                  key={doc.id}
                  className="bg-white border border-slate-200 rounded-2xl p-5 shadow-sm space-y-3 hover:border-slate-300 transition-all flex flex-col justify-between"
                >
                  <div className="space-y-3">
                    <div className="flex items-center justify-between gap-2">
                      {getDocTypeBadge(doc.type_doc)}
                      {doc.date_expiration && (
                        <span className={`inline-flex items-center gap-1 text-[11px] font-bold px-2 py-0.5 rounded-md ${expired ? "bg-rose-50 text-rose-700 border border-rose-200" : "bg-emerald-50 text-emerald-700 border border-emerald-200"}`}>
                          <Calendar size={11} />
                          {expired ? "Expiré : " : "Exp : "}
                          {new Date(doc.date_expiration).toLocaleDateString("fr-FR")}
                        </span>
                      )}
                    </div>
                    <div>
                      <h3 className="text-sm font-bold text-slate-900 leading-snug">{doc.titre}</h3>

                      {meta?.specialite_principale && (
                        <div className="flex flex-wrap items-center gap-1.5 mt-2">
                          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-bold bg-indigo-50 text-indigo-800 border border-indigo-200">
                            <Sparkles size={10} className="text-indigo-600" />
                            {meta.specialite_principale}
                          </span>
                          {meta.experience_annees ? (
                            <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-semibold bg-slate-100 text-slate-700 border border-slate-200">
                              <Clock size={10} className="text-slate-500" />
                              {meta.experience_annees} ans d&apos;exp.
                            </span>
                          ) : null}
                        </div>
                      )}

                      {doc.description && (
                        <p className="text-xs text-slate-600 mt-2 line-clamp-3 leading-relaxed bg-slate-50 p-2 rounded-lg border border-slate-100">
                          {doc.description}
                        </p>
                      )}
                      {doc.file_name && (
                        <p className="text-[11px] text-slate-400 mt-1 truncate font-mono">
                          Fichier : {doc.file_name}
                        </p>
                      )}
                    </div>
                  </div>

                  <div className="pt-2 border-t border-slate-100 flex items-center justify-between gap-2 text-xs">
                    {doc.type_doc === "CV_EXPERT" && doc.google_drive_id ? (
                      <button
                        type="button"
                        onClick={() => handleAnalyzeSingleCV(doc.id)}
                        disabled={analyzingDocId === doc.id || analyzingAllCVs}
                        className="inline-flex items-center gap-1 text-[11px] font-bold text-indigo-700 hover:text-indigo-900 bg-indigo-50 hover:bg-indigo-100 px-2.5 py-1 rounded-lg border border-indigo-200 transition-colors disabled:opacity-50 cursor-pointer"
                        title="Lire le fichier PDF complet avec Gemini et extraire les diplômes et compétences"
                      >
                        {analyzingDocId === doc.id ? (
                          <>
                            <RefreshCw size={11} className="animate-spin text-indigo-600" />
                            <span>Lecture IA...</span>
                          </>
                        ) : (
                          <>
                            <Sparkles size={11} className="text-amber-500" />
                            <span>{meta?.specialite_principale ? "Ré-analyser" : "Analyser par l'IA"}</span>
                          </>
                        )}
                      </button>
                    ) : (
                      <span className="text-[11px] text-slate-400 font-mono">
                        {new Date(doc.created_at).toLocaleDateString("fr-FR")}
                      </span>
                    )}

                    {doc.web_view_link ? (
                      <a
                        href={doc.web_view_link}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex items-center gap-1 text-xs font-semibold text-indigo-600 hover:text-indigo-800"
                      >
                        <span>Voir sur Drive</span>
                        <ExternalLink size={12} />
                      </a>
                    ) : (
                      <span className="text-[11px] text-slate-400">Métadonnée locale</span>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Modal Téléversement Document RAG */}
      {showAddDocModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-200">
          <div className="bg-white rounded-2xl shadow-2xl border border-slate-100 max-w-lg w-full overflow-hidden">
            <div className="bg-gradient-to-r from-slate-900 to-indigo-950 text-white p-6">
              <h3 className="text-lg font-bold">Ajouter une Pièce au Coffre-fort RAG</h3>
              <p className="text-xs text-slate-400 mt-0.5">
                Le document sera sauvegardé sur Drive dans <code className="text-indigo-300">01_COFFRE_FORT_ENTREPRISE</code> et utilisable pour vos DAO.
              </p>
            </div>

            <form onSubmit={handleUploadDocument} className="p-6 space-y-4">
              {uploadError && (
                <div className="p-3 bg-red-50 border border-red-200 text-red-700 text-xs rounded-xl flex items-center gap-2">
                  <AlertCircle size={16} className="shrink-0" />
                  <span>{uploadError}</span>
                </div>
              )}

              {/* Téléversement de fichier */}
              <div>
                <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                  Fichier à téléverser (PDF, Word .docx/.doc, Images) <span className="text-slate-400">(Recommandé)</span>
                </label>
                <div
                  onClick={() => fileInputRef.current?.click()}
                  className="border-2 border-dashed border-slate-200 hover:border-indigo-400 hover:bg-indigo-50/50 rounded-xl p-4 text-center cursor-pointer transition-all"
                >
                  <Upload size={22} className="mx-auto text-indigo-500 mb-1" />
                  <p className="text-xs font-semibold text-slate-700">
                    {docFile ? docFile.name : "Cliquez pour choisir un fichier (PDF, Word ou Image)"}
                  </p>
                  <p className="text-[10px] text-slate-400 mt-0.5">Max 30 Mo — Sauvegarde automatique dans 01_COFFRE_FORT_ENTREPRISE</p>
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept=".pdf,.png,.jpg,.jpeg,.doc,.docx"
                    className="hidden"
                    onChange={(e) => {
                      if (e.target.files && e.target.files[0]) {
                        const f = e.target.files[0];
                        setDocFile(f);
                        if (!docTitre) {
                          setDocTitre(f.name.replace(/\.[^/.]+$/, ""));
                        }
                      }
                    }}
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                  Type de Document
                </label>
                <select
                  value={docType}
                  onChange={(e) => setDocType(e.target.value as TypeDocumentEntreprise)}
                  className="w-full px-4 py-2.5 rounded-xl border border-slate-300 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 bg-white"
                >
                  <option value="ATTESTATION_FISCALE_SOCIALE">Attestation Fiscale & Sociale (Quitus, CSS, IPRES...)</option>
                  <option value="AUTRE">Document Administratif Général (NINEA, RCCM, Déclaration...)</option>
                  <option value="CV_EXPERT">CV Expert / Personnel Clé</option>
                  <option value="REFERENCE_SERVICE_FAIT">Attestation de Service Fait / Référence Chantier</option>
                  <option value="JUSTIFICATIF_MATERIEL">Justificatif de Matériel / Engin</option>
                  <option value="BILAN_FINANCIER">Bilan Financier Certifié</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                  Intitulé du document <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  required
                  value={docTitre}
                  onChange={(e) => setDocTitre(e.target.value)}
                  placeholder="ex: Quitus Fiscal DGID 2026 ou NINEA Entreprise"
                  className="w-full px-4 py-2.5 rounded-xl border border-slate-300 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
              </div>

              <div>
                <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                  Date d&apos;expiration (Optionnel pour pièces renouvelables)
                </label>
                <input
                  type="date"
                  value={docExp}
                  onChange={(e) => setDocExp(e.target.value)}
                  className="w-full px-4 py-2.5 rounded-xl border border-slate-300 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
              </div>

              <div>
                <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                  Description / Points Clés
                </label>
                <textarea
                  rows={2}
                  value={docDesc}
                  onChange={(e) => setDocDesc(e.target.value)}
                  placeholder="ex: Délivré le 15/01/2026. Valable pour l'ensemble des marchés publics."
                  className="w-full px-4 py-2.5 rounded-xl border border-slate-300 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
              </div>

              <div className="pt-3 flex items-center justify-end gap-3 border-t border-slate-100">
                <button
                  type="button"
                  onClick={() => setShowAddDocModal(false)}
                  disabled={submitting}
                  className="px-4 py-2 text-sm font-semibold text-slate-600 hover:bg-slate-100 rounded-xl"
                >
                  Annuler
                </button>
                <button
                  type="submit"
                  disabled={submitting || (!docFile && !docTitre.trim())}
                  className="px-5 py-2.5 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white text-sm font-semibold rounded-xl cursor-pointer shadow-md shadow-indigo-500/20 flex items-center gap-2"
                >
                  {submitting && <RefreshCw size={14} className="animate-spin" />}
                  <span>{submitting ? "Téléversement..." : "Enregistrer dans le RAG"}</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
